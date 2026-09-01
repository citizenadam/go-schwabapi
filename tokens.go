package schwabdev

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/fernet/fernet-go"
)

// TokenManager manages OAuth tokens for the Schwab API.
// All exported methods are safe for concurrent use.
type TokenManager struct {
	appKey      string
	appSecret   string
	callbackURL string

	encryptionKey *fernet.Key
	logger        *slog.Logger

	// callOnAuth receives the authorization URL and must return the full
	// callback URL after the user completes the OAuth flow. When nil the
	// manager falls back to an interactive stdin prompt.
	callOnAuth func(authURL string) (callbackURL string, err error)

	// storage is the pluggable persistence backend. TokenManager never
	// accesses the underlying medium directly — all I/O goes through here.
	storage TokenStorage

	// mu guards the in-memory token fields below.
	mu sync.RWMutex

	accessToken  string
	refreshToken string
	idToken      string

	accessTokenIssued   time.Time
	refreshTokenIssued  time.Time
	accessTokenTimeout  time.Duration
	refreshTokenTimeout time.Duration

	// oauthTokenURL is the endpoint for OAuth token grants. Defaults to
	// OAuthTokenEndpoint; overridable in tests via an httptest.Server.
	oauthTokenURL string
}

// NewTokenManager creates a TokenManager using a caller-supplied TokenStorage.
// Use NewTokenManagerWithFilePath for the common case of file-based persistence.
func NewTokenManager(
	appKey, appSecret, callbackURL string,
	storage TokenStorage,
	encryption string,
	logger *slog.Logger,
	callOnAuth func(authURL string) (string, error),
) (*TokenManager, error) {
	if err := validateParams(appKey, appSecret, callbackURL); err != nil {
		return nil, err
	}

	tm := &TokenManager{
		appKey:              appKey,
		appSecret:           appSecret,
		callbackURL:         callbackURL,
		storage:             storage,
		logger:              logger,
		callOnAuth:          callOnAuth,
		accessTokenTimeout:  AccessTokenValidity,
		refreshTokenTimeout: RefreshTokenValidity,
		oauthTokenURL:       OAuthTokenEndpoint,
	}

	if encryption != "" {
		key, err := ValidateKey(encryption)
		if err == nil {
			tm.encryptionKey = key
		} else if logger != nil {
			logger.Warn("[Schwabdev] Encryption key invalid, proceeding without encryption", "error", err)
		}
	}

	return tm, nil
}

// NewTokenManagerWithFilePath is the convenience constructor for the common
// case: file-based token persistence with no external dependencies.
// storagePath may be empty (defaults to ~/.schwabdev/tokens.json) or start with ~.
func NewTokenManagerWithFilePath(
	appKey, appSecret, callbackURL, storagePath, encryption string,
	logger *slog.Logger,
	callOnAuth func(authURL string) (string, error),
) (*TokenManager, error) {
	if strings.HasSuffix(storagePath, "/") {
		return nil, ErrTokensDBEndsWithSlash
	}
	storage, err := NewFileTokenStorage(storagePath)
	if err != nil {
		return nil, err
	}
	return NewTokenManager(appKey, appSecret, callbackURL, storage, encryption, logger, callOnAuth)
}

// Close releases resources held by the storage backend.
func (tm *TokenManager) Close() error {
	return tm.storage.Close()
}

// Reload forces the TokenManager to reload its state from storage.
// Use after manually saving new tokens to the DB (e.g. after re-auth)
// to synchronize the in-memory state without triggering a token refresh.
func (tm *TokenManager) Reload() error {
	return tm.loadFromStorage()
}

// AccessToken returns a guaranteed-fresh access token, refreshing it first if
// it is within the expiry threshold. This is the correct entry point for any
// component (e.g. the streamer) that needs a token without holding a stale copy.
func (tm *TokenManager) AccessToken() (string, error) {
	if _, err := tm.UpdateTokens(false, false); err != nil {
		return "", err
	}
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.accessToken, nil
}

// TokenInfo holds a snapshot of the current token state for display purposes.
type TokenInfo struct {
	AccessToken        string    // raw token value (truncate before displaying)
	AccessTokenIssued  time.Time // when the current access token was issued
	RefreshTokenIssued time.Time // when the current refresh token was issued
	AccessTokenExpiry  time.Time // when the access token expires
	RefreshTokenExpiry time.Time // when the refresh token expires
}

// Valid reports whether the access token is currently unexpired.
func (t TokenInfo) Valid() bool {
	return time.Now().Before(t.AccessTokenExpiry)
}

// TokenInfo returns a snapshot of the current token state.
// It does not trigger a refresh — call UpdateTokens first if needed.
func (tm *TokenManager) TokenInfo() TokenInfo {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return TokenInfo{
		AccessToken:        tm.accessToken,
		AccessTokenIssued:  tm.accessTokenIssued,
		RefreshTokenIssued: tm.refreshTokenIssued,
		AccessTokenExpiry:  tm.accessTokenIssued.Add(tm.accessTokenTimeout),
		RefreshTokenExpiry: tm.refreshTokenIssued.Add(tm.refreshTokenTimeout),
	}
}

// UpdateTokens checks expiry and refreshes tokens as needed.
// Returns true if any refresh was performed.
func (tm *TokenManager) UpdateTokens(forceAccessToken, forceRefreshToken bool) (bool, error) {
	// Lazy-load from storage on first call.
	tm.mu.RLock()
	empty := tm.accessToken == ""
	tm.mu.RUnlock()

	if empty {
		if err := tm.loadFromStorage(); err != nil {
			return false, err
		}
	}

	now := time.Now().UTC()

	tm.mu.RLock()
	rtDelta := tm.refreshTokenTimeout - now.Sub(tm.refreshTokenIssued)
	atDelta := tm.accessTokenTimeout - now.Sub(tm.accessTokenIssued)
	tm.mu.RUnlock()

	if rtDelta < RefreshTokenRefreshThreshold || forceRefreshToken {
		if tm.logger != nil {
			if rtDelta < 0 {
				tm.logger.Warn("[Schwabdev] Refresh token has expired, re-authorisation required")
			} else {
				tm.logger.Warn("[Schwabdev] Refresh token expiring soon (<60 min), re-authorising")
			}
		}
		return true, tm.updateRefreshToken()
	}

	if atDelta < AccessTokenRefreshThreshold || forceAccessToken {
		if tm.logger != nil {
			tm.logger.Debug("[Schwabdev] Access token expiring, refreshing")
		}
		return true, tm.updateAccessToken()
	}

	return false, nil
}

// ── Storage read ──────────────────────────────────────────────────────────────

func (tm *TokenManager) loadFromStorage() error {
	rec, err := tm.storage.Load(context.Background())
	if err != nil {
		return fmt.Errorf("load tokens: %w", err)
	}
	if rec == nil {
		// No tokens stored yet — first run, nothing to load.
		return nil
	}
	return tm.applyRecord(rec)
}

// applyRecord decrypts a TokenRecord and adopts it into in-memory state.
func (tm *TokenManager) applyRecord(rec *TokenRecord) error {
	decryptedAT, err := Decrypt(rec.AccessToken, tm.encryptionKey)
	if err != nil {
		return fmt.Errorf("decrypt access token: %w", err)
	}
	decryptedRT, err := Decrypt(rec.RefreshToken, tm.encryptionKey)
	if err != nil {
		return fmt.Errorf("decrypt refresh token: %w", err)
	}

	timeout := AccessTokenValidity
	if rec.ExpiresIn > 0 {
		timeout = time.Duration(rec.ExpiresIn) * time.Second
	}

	tm.mu.Lock()
	tm.accessTokenIssued = rec.AccessTokenIssued.UTC()
	tm.refreshTokenIssued = rec.RefreshTokenIssued.UTC()
	tm.accessToken = decryptedAT
	tm.refreshToken = decryptedRT
	tm.idToken = rec.IDToken
	tm.accessTokenTimeout = timeout
	tm.mu.Unlock()

	return nil
}

// ── Storage write ─────────────────────────────────────────────────────────────

func (tm *TokenManager) saveTokens(atIssued, rtIssued time.Time, tokenDict map[string]any) error {
	// Update in-memory state under the write lock, then release before I/O.
	tm.mu.Lock()
	if val, ok := tokenDict["access_token"].(string); ok && val != "" {
		tm.accessToken = val
	}
	if val, ok := tokenDict["refresh_token"].(string); ok && val != "" {
		tm.refreshToken = val
	}
	if val, ok := tokenDict["id_token"].(string); ok && val != "" {
		tm.idToken = val
	}
	tm.accessTokenIssued = atIssued
	tm.refreshTokenIssued = rtIssued
	expiresIn := int(AccessTokenValidity.Seconds())
	if exp, ok := tokenDict["expires_in"].(float64); ok {
		tm.accessTokenTimeout = time.Duration(exp) * time.Second
		expiresIn = int(exp)
	}
	// Capture values for storage write before releasing the lock.
	at, rt, it := tm.accessToken, tm.refreshToken, tm.idToken
	tm.mu.Unlock()

	// Encrypt outside the lock (CPU-bound, not shared state).
	encAT, err := Encrypt(at, tm.encryptionKey)
	if err != nil {
		return fmt.Errorf("encrypt access token: %w", err)
	}
	encRT, err := Encrypt(rt, tm.encryptionKey)
	if err != nil {
		return fmt.Errorf("encrypt refresh token: %w", err)
	}

	return tm.storage.Save(context.Background(), TokenRecord{
		AccessTokenIssued:  atIssued.UTC(),
		RefreshTokenIssued: rtIssued.UTC(),
		AccessToken:        encAT,
		RefreshToken:       encRT,
		IDToken:            it,
		ExpiresIn:          expiresIn,
		TokenType:          "Bearer",
		Scope:              "api",
	})
}

// ── Token refresh ─────────────────────────────────────────────────────────────

func (tm *TokenManager) updateAccessToken() error {
	// Serialize cross-process refresh. When the storage implements
	// TokenStoreLocker, acquire the lock so concurrent processes sharing the
	// same store do not consume the rotating refresh_token simultaneously
	// (which would yield invalid_grant on the loser).
	if locker, ok := tm.storage.(TokenStoreLocker); ok {
		release, err := locker.AcquireRefreshLock(context.Background())
		if err != nil {
			return fmt.Errorf("acquire refresh lock: %w", err)
		}
		defer release()
	}

	// Re-load the store now that we hold the lock. A peer may have already
	// rotated the token while we were waiting; if so, adopt it rather than
	// sending a stale refresh_token to Schwab.
	rec, err := tm.storage.Load(context.Background())
	if err != nil {
		return fmt.Errorf("load tokens: %w", err)
	}
	if rec != nil && rec.AccessToken != "" {
		tm.mu.RLock()
		cur := tm.accessToken
		tm.mu.RUnlock()
		if rec.AccessToken != cur {
			if err := tm.applyRecord(rec); err != nil {
				return err
			}
			// Adopt and skip the HTTP refresh if the peer's token is still
			// comfortably fresh (outside the refresh threshold).
			info := tm.TokenInfo()
			if time.Now().Before(info.AccessTokenExpiry.Add(-AccessTokenRefreshThreshold)) {
				if tm.logger != nil {
					tm.logger.Debug("[Schwabdev] Adopted peer-refreshed token from storage; skipping HTTP refresh")
				}
				return nil
			}
		}
	}

	tm.mu.RLock()
	rt := tm.refreshToken
	// Preserve the original refresh-token acquisition anchor. Schwab refresh
	// tokens are valid for 7 days from acquisition and are NOT extended by
	// access-token refreshes. Re-stamping refreshTokenIssued here would make
	// TokenInfo().RefreshTokenExpiry perpetually reset to ~7 days, keeping the
	// UI stuck at "6d 23h" and disabling the near-expiry re-auth warning.
	rtIssued := tm.refreshTokenIssued
	tm.mu.RUnlock()

	response, err := tm.postOAuthToken(context.Background(), "refresh_token", rt)
	if err != nil {
		return err
	}
	return tm.saveTokens(time.Now().UTC(), rtIssued, response)
}

func (tm *TokenManager) updateRefreshToken() error {
	authCode, err := tm.getNewTokens()
	if err != nil {
		return err
	}
	response, err := tm.postOAuthToken(context.Background(), "authorization_code", authCode)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	return tm.saveTokens(now, now, response)
}

// ExchangeCode exchanges an authorization code for tokens using the
// authorization_code grant and returns the resulting TokenRecord with raw
// (unencrypted) token values. It does NOT persist the record or update the
// in-memory token state — the caller persists it via its TokenStorage and
// calls Reload() to synchronize. This is the entry point for application
// auth-callback handlers that receive the authorization code from the OAuth
// redirect and manage their own storage.
func (tm *TokenManager) ExchangeCode(ctx context.Context, code string) (*TokenRecord, error) {
	if code == "" {
		return nil, fmt.Errorf("authorization code cannot be empty")
	}

	response, err := tm.postOAuthToken(ctx, "authorization_code", code)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	expiresIn := int(AccessTokenValidity.Seconds())
	if exp, ok := response["expires_in"].(float64); ok {
		expiresIn = int(exp)
	}

	accessToken, _ := response["access_token"].(string)
	refreshToken, _ := response["refresh_token"].(string)
	idToken, _ := response["id_token"].(string)
	tokenType, _ := response["token_type"].(string)
	if tokenType == "" {
		tokenType = "Bearer"
	}

	return &TokenRecord{
		AccessTokenIssued:  now,
		RefreshTokenIssued: now,
		AccessToken:        accessToken,
		RefreshToken:       refreshToken,
		IDToken:            idToken,
		ExpiresIn:          expiresIn,
		TokenType:          tokenType,
		Scope:              "api",
	}, nil
}

// ── OAuth helpers ─────────────────────────────────────────────────────────────

func (tm *TokenManager) getNewTokens() (string, error) {
	authURL := fmt.Sprintf(
		"https://api.schwabapi.com/v1/oauth/authorize?client_id=%s&redirect_uri=%s",
		tm.appKey, url.QueryEscape(tm.callbackURL),
	)

	var rawCallback string
	if tm.callOnAuth != nil {
		cb, err := tm.callOnAuth(authURL)
		if err != nil {
			return "", fmt.Errorf("callOnAuth: %w", err)
		}
		rawCallback = cb
	} else {
		fmt.Printf("[Schwabdev] Open to authenticate: %s\n", authURL)
		fmt.Print("[Schwabdev] Paste the address bar URL here: ")
		if _, err := fmt.Scanln(&rawCallback); err != nil {
			return "", fmt.Errorf("read callback URL from stdin: %w", err)
		}
	}

	rawCallback = strings.TrimSpace(rawCallback)
	parsed, err := url.Parse(rawCallback)
	if err != nil || parsed.Query().Get("code") == "" {
		return rawCallback, nil
	}
	return parsed.Query().Get("code"), nil
}

func (tm *TokenManager) postOAuthToken(ctx context.Context, grantType, code string) (map[string]any, error) {
	client := &http.Client{Timeout: OAuthTokenRequestTimeout}
	auth := base64.StdEncoding.EncodeToString([]byte(tm.appKey + ":" + tm.appSecret))

	data := url.Values{}
	data.Set("grant_type", grantType)
	switch grantType {
	case "authorization_code":
		data.Set("code", code)
		data.Set("redirect_uri", tm.callbackURL)
	case "refresh_token":
		data.Set("refresh_token", code)
	default:
		return nil, ErrInvalidGrantType
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		tm.oauthTokenURL,
		strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() //nolint:errcheck // body fully read; close error not actionable

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse OAuth response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		if desc, ok := result["error_description"].(string); ok {
			return nil, fmt.Errorf("OAuth error: %s", desc)
		}
		return nil, fmt.Errorf("OAuth request failed (%d): %s", resp.StatusCode, body)
	}

	return result, nil
}

// ── Validation ────────────────────────────────────────────────────────────────

func validateParams(appKey, appSecret, callbackURL string) error {
	if appKey == "" {
		return ErrAppKeyRequired
	}
	if appSecret == "" {
		return ErrAppSecretRequired
	}
	if callbackURL == "" {
		return ErrCallbackURLRequired
	}
	keyLen, secretLen := len(appKey), len(appSecret)
	if (keyLen != AppKeyLength1 && keyLen != AppKeyLength2) ||
		(secretLen != AppSecretLength1 && secretLen != AppSecretLength2) {
		return ErrInvalidKeyLength
	}
	if !strings.HasPrefix(callbackURL, "https") {
		return ErrCallbackNotHTTPS
	}
	if strings.HasSuffix(callbackURL, "/") {
		return ErrCallbackEndsWithSlash
	}
	return nil
}
