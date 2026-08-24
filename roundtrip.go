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
	"sync/atomic"
	"time"

	"golang.org/x/sync/singleflight"
)

// SchwabRoundTripper is an http.RoundTripper that automatically refreshes
// tokens before they expire. It wraps another RoundTripper and injects
// the Authorization header with a fresh access token.
//
// Use this to integrate with your existing token storage without
// modifying your auth flow.
//
// Example:
//
//	provider := &MyDBProvider{db: myDB}
//	transport := schwabdev.NewSchwabRoundTripper(provider, config, nil)
//	client := &http.Client{Transport: transport}
type SchwabRoundTripper struct {
	// Source supplies tokens - your implementation.
	Source TokenSource

	// RefreshBuffer is how long before expiry to trigger refresh.
	// Default: 60 seconds.
	RefreshBuffer time.Duration

	// Base is the underlying transport. Defaults to http.DefaultTransport.
	Base http.RoundTripper

	// Logger for debugging. Optional.
	Logger *slog.Logger

	// oauthConfig for refreshing tokens.
	// Required: AppKey, AppSecret, CallbackURL.
	oauthConfig OAuthConfig

	// Internal sync.
	mu         sync.Mutex
	refreshing atomic.Bool
	sfGroup    singleflight.Group
}

// OAuthConfig contains the credentials needed to refresh tokens.
type OAuthConfig struct {
	AppKey      string
	AppSecret   string
	CallbackURL string
}

// NewSchwabRoundTripper creates a RoundTripper wrapping base transport.
// provider must implement TokenProvider (GetToken, SaveToken).
// config must have AppKey, AppSecret, CallbackURL for refresh.
func NewSchwabRoundTripper(source TokenSource, config OAuthConfig, base http.RoundTripper) *SchwabRoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}
	if config.AppKey == "" || config.AppSecret == "" || config.CallbackURL == "" {
		panic("schwabdev.NewSchwabRoundTripper: oauth config required")
	}
	return &SchwabRoundTripper{
		Source:        source,
		Base:          base,
		oauthConfig:   config,
		RefreshBuffer: 60 * time.Second,
	}
}

// RoundTrip implements http.RoundTripper.
func (s *SchwabRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()

	// 1. Get token from provider (or refresh if needed).
	token, err := s.getOrRefreshToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("token: %w", err)
	}
	if token == nil {
		return nil, fmt.Errorf("token: no token available")
	}

	// 2. Inject Authorization header.
	req = req.Clone(ctx)
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	// 3. Delegate to base transport.
	return s.Base.RoundTrip(req)
}

// getOrRefreshToken returns a valid token, refreshing if needed.
// Uses mutex + singleflight to prevent thundering herd.
func (s *SchwabRoundTripper) getOrRefreshToken(ctx context.Context) (*Token, error) {
	// Fast path: get token without refresh.
	tok, err := s.Source.GetToken(ctx)
	if err != nil {
		return nil, err
	}
	if tok != nil && !tok.NeedsRefresh(s.RefreshBuffer) {
		return tok, nil
	}

	// Slow path: need refresh. Acquire lock.
	if !s.tryStartRefresh() {
		// Another goroutine is refreshing. Wait for it.
		// Use singleflight to deduplicate.
		res, err, _ := s.sfGroup.Do("refresh", func() (any, error) {
			// Double-check after acquire.
			tok, err := s.Source.GetToken(ctx)
			if err != nil {
				return nil, err
			}
			if tok != nil && !tok.NeedsRefresh(s.RefreshBuffer) {
				return tok, nil
			}
			return s.doRefresh(ctx)
		})
		if err != nil {
			return nil, err
		}
		return res.(*Token), nil
	}

	// We won the refresh race.
	defer s.finishRefresh()
	return s.doRefresh(ctx)
}

// tryStartRefresh returns true if this goroutine should perform refresh.
func (s *SchwabRoundTripper) tryStartRefresh() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.refreshing.Load() {
		return false
	}
	s.refreshing.Store(true)
	return true
}

// finishRefresh marks refresh as complete.
func (s *SchwabRoundTripper) finishRefresh() {
	s.refreshing.Store(false)
}

// doRefresh performs the OAuth refresh flow.
func (s *SchwabRoundTripper) doRefresh(ctx context.Context) (*Token, error) {
	// Get current refresh token.
	current, err := s.Source.GetToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("get token: %w", err)
	}
	if current == nil || current.RefreshToken == "" {
		return nil, fmt.Errorf("no refresh token available")
	}

	// Perform OAuth refresh.
	newToken, err := s.refreshOAuth(ctx, current.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("oauth refresh: %w", err)
	}

	// Save to storage.
	if err := s.Source.SaveToken(ctx, newToken); err != nil {
		return nil, fmt.Errorf("save token: %w", err)
	}

	if s.Logger != nil {
		s.Logger.Debug("token refreshed", "expires_at", newToken.ExpiresAt)
	}
	return newToken, nil
}

// refreshOAuth exchanges a refresh token for new tokens.
func (s *SchwabRoundTripper) refreshOAuth(ctx context.Context, refreshToken string) (*Token, error) {
	client := &http.Client{Timeout: OAuthTokenRequestTimeout}
	auth := base64.StdEncoding.EncodeToString([]byte(s.oauthConfig.AppKey + ":" + s.oauthConfig.AppSecret))

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.schwabapi.com/v1/oauth/token",
		strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", "Basic "+auth)
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

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

	// Extract tokens.
	accessToken, _ := result["access_token"].(string)
	refreshTokenOut, _ := result["refresh_token"].(string)
	if refreshTokenOut == "" {
		refreshTokenOut = refreshToken // Reuse if not returned.
	}

	expiresIn := 1800 // 30 minutes default.
	if exp, ok := result["expires_in"].(float64); ok {
		expiresIn = int(exp)
	}

	return &Token{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenOut,
		ExpiresAt:    time.Now().Add(time.Duration(expiresIn) * time.Second),
	}, nil
}
