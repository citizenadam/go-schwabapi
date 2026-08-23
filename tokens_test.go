package schwabdev

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

// memoryStorage is an in-process TokenStorage stub for unit tests.
type memoryStorage struct {
	mu  sync.Mutex
	rec *TokenRecord
}

func (m *memoryStorage) Load(_ context.Context) (*TokenRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.rec, nil
}

func (m *memoryStorage) Save(_ context.Context, rec TokenRecord) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rec = &rec
	return nil
}

func (m *memoryStorage) Close() error { return nil }

// newTestTokenManager builds a TokenManager pointed at an httptest.Server that
// answers token grants. The server handler must return a JSON token response.
func newTestTokenManager(t *testing.T, handler http.HandlerFunc, storage *memoryStorage) (*TokenManager, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	tm, err := NewTokenManager(
		strings.Repeat("a", 32), // AppKeyLength1
		strings.Repeat("b", 16), // AppSecretLength1
		"https://127.0.0.1/callback",
		storage,
		"", // no encryption — plaintext passthrough
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("NewTokenManager: %v", err)
	}
	tm.oauthTokenURL = srv.URL
	return tm, srv
}

// tokenGrantHandler returns a fixed OAuth token response (including a ROTATED
// refresh token, matching Schwab's refresh grant behavior).
func tokenGrantHandler(t *testing.T) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "new-access-token",
			"refresh_token": "new-rotated-refresh-token",
			"id_token":      "new-id-token",
			"expires_in":    1800,
			"token_type":    "Bearer",
			"scope":         "api",
		}); err != nil {
			t.Errorf("encode token response: %v", err)
		}
	}
}

// TestUpdateAccessToken_PreservesRefreshExpiry is the regression test for the
// refresh-token countdown bug: an access-token refresh MUST NOT re-anchor
// refreshTokenIssued (Schwab refresh tokens expire 7 days from acquisition and
// are NOT extended by access-token refreshes).
func TestUpdateAccessToken_PreservesRefreshExpiry(t *testing.T) {
	storage := &memoryStorage{}
	acquisition := time.Now().UTC().Add(-3 * 24 * time.Hour) // refresh token acquired 3 days ago
	storage.rec = &TokenRecord{
		AccessTokenIssued:  time.Now().UTC().Add(-29 * time.Minute),
		RefreshTokenIssued: acquisition,
		AccessToken:        "old-access-token",
		RefreshToken:       "old-refresh-token",
		IDToken:            "old-id-token",
		ExpiresIn:          1800,
		TokenType:          "Bearer",
		Scope:              "api",
	}

	tm, _ := newTestTokenManager(t, tokenGrantHandler(t), storage)

	// Trigger the lazy load from storage before snapshotting state.
	if err := tm.loadFromStorage(); err != nil {
		t.Fatalf("loadFromStorage: %v", err)
	}

	before := tm.TokenInfo()
	updated, err := tm.UpdateTokens(true, false) // force access-token refresh
	if err != nil {
		t.Fatalf("UpdateTokens(true,false): %v", err)
	}
	if !updated {
		t.Fatal("expected an access-token refresh to be performed")
	}

	after := tm.TokenInfo()

	// The refresh-token expiry must remain anchored to acquisition + 7 days.
	wantRT := acquisition.Add(RefreshTokenValidity)
	if diff := after.RefreshTokenExpiry.Sub(wantRT); diff < -5*time.Second || diff > 5*time.Second {
		t.Errorf("RefreshTokenExpiry moved: got %v, want ~%v (diff %v)",
			after.RefreshTokenExpiry, wantRT, diff)
	}
	if !after.RefreshTokenExpiry.Equal(before.RefreshTokenExpiry) {
		t.Errorf("RefreshTokenExpiry changed across access-token refresh: before=%v after=%v",
			before.RefreshTokenExpiry, after.RefreshTokenExpiry)
	}

	// The access-token window must advance to ~now + expires_in.
	if d := time.Until(after.AccessTokenExpiry); d < (AccessTokenValidity-5*time.Second) || d > (AccessTokenValidity+5*time.Second) {
		t.Errorf("AccessTokenExpiry not re-issued: remaining=%v, want ~%v", d, AccessTokenValidity)
	}

	// A rotated refresh-token VALUE is still stored...
	if storage.rec == nil {
		t.Fatal("no token record persisted")
	}
	if storage.rec.RefreshToken != "new-rotated-refresh-token" {
		t.Errorf("rotated refresh token not persisted: got %q", storage.rec.RefreshToken)
	}
	if storage.rec.AccessToken != "new-access-token" {
		t.Errorf("access token not persisted: got %q", storage.rec.AccessToken)
	}
	// ...but the persisted acquisition anchor is untouched.
	if !storage.rec.RefreshTokenIssued.Equal(acquisition) {
		t.Errorf("persisted refresh_token_issued moved: got %v, want %v",
			storage.rec.RefreshTokenIssued, acquisition)
	}
}

// TestUpdateRefreshToken_ReanchorsExpiry verifies that a full re-authorization
// (authorization_code grant) is the one event that resets the 7-day countdown.
func TestUpdateRefreshToken_ReanchorsExpiry(t *testing.T) {
	storage := &memoryStorage{}
	storage.rec = &TokenRecord{
		AccessTokenIssued:  time.Now().UTC().Add(-29 * time.Minute),
		RefreshTokenIssued: time.Now().UTC().Add(-6 * 24 * time.Hour),
		AccessToken:        "old-access-token",
		RefreshToken:       "old-refresh-token",
		IDToken:            "old-id-token",
		ExpiresIn:          1800,
		TokenType:          "Bearer",
		Scope:              "api",
	}

	tm, _ := newTestTokenManager(t, tokenGrantHandler(t), storage)
	// getNewTokens() uses callOnAuth when set; supply a fake callback URL with a code.
	tm.callOnAuth = func(authURL string) (string, error) {
		return "https://127.0.0.1/callback?code=test-auth-code", nil
	}

	updated, err := tm.UpdateTokens(true, true) // force full re-auth
	if err != nil {
		t.Fatalf("UpdateTokens(true,true): %v", err)
	}
	if !updated {
		t.Fatal("expected a refresh-token re-authorization to be performed")
	}

	after := tm.TokenInfo()
	// Both issued times re-anchor to now; refresh expiry = now + 7 days.
	if d := time.Since(after.RefreshTokenIssued); d < 0 || d > 5*time.Second {
		t.Errorf("RefreshTokenIssued not re-anchored: age=%v, want ~0s", d)
	}
	if d := time.Until(after.RefreshTokenExpiry); d < (RefreshTokenValidity-5*time.Second) || d > (RefreshTokenValidity+5*time.Second) {
		t.Errorf("RefreshTokenExpiry not reset to ~7d: remaining=%v", d)
	}
	if d := time.Until(after.AccessTokenExpiry); d < (AccessTokenValidity-5*time.Second) || d > (AccessTokenValidity+5*time.Second) {
		t.Errorf("AccessTokenExpiry not reset to ~30m: remaining=%v", d)
	}
	if storage.rec == nil || storage.rec.RefreshToken != "new-rotated-refresh-token" {
		t.Errorf("re-auth refresh token not persisted: %+v", storage.rec)
	}
}