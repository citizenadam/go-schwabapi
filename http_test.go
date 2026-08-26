package schwabdev

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// seededTokenManager returns a TokenManager with valid in-memory tokens so
// authHeader resolves without touching the network. The oauth URL points at a
// token-grant server for any refresh that might slip through.
func seededTokenManager(t *testing.T) *TokenManager {
	t.Helper()
	tm, _ := newTestTokenManager(t, tokenGrantHandler(t), &memoryStorage{})
	now := time.Now().UTC()
	tm.mu.Lock()
	tm.accessToken = "test-access-token"
	tm.accessTokenIssued = now
	tm.refreshToken = "test-refresh-token"
	tm.refreshTokenIssued = now
	tm.mu.Unlock()
	return tm
}

// newQuotesClient builds a Client wired to the given quotes handler with the
// given options, seeded with valid tokens.
func newQuotesClient(t *testing.T, handler http.HandlerFunc, opts ...ClientOption) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	tm := seededTokenManager(t)
	client, err := NewClientWithTokenManager(tm, append(opts, WithBaseURL(srv.URL))...)
	if err != nil {
		t.Fatalf("NewClientWithTokenManager: %v", err)
	}
	return client, srv
}

// TestDoRequest_Non2xxReturnsSchwabAPIError verifies that non-2xx responses
// surface as typed errors carrying the status code (fail-loudly contract).
func TestDoRequest_Non2xxReturnsSchwabAPIError(t *testing.T) {
	for _, status := range []int{http.StatusBadRequest, http.StatusTooManyRequests, http.StatusUnauthorized} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			client, _ := newQuotesClient(t, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(status)
				_, _ = w.Write([]byte(`{"error":"boom"}`))
			})

			_, err := client.Quotes(context.Background(), "SPX", nil, nil)
			if err == nil {
				t.Fatalf("expected error for status %d, got nil", status)
			}

			var apiErr *SchwabAPIError
			if !errors.As(err, &apiErr) {
				t.Fatalf("expected *SchwabAPIError, got %T: %v", err, err)
			}
			if apiErr.StatusCode != status {
				t.Fatalf("StatusCode = %d, want %d", apiErr.StatusCode, status)
			}
			if !strings.Contains(apiErr.Error(), "boom") {
				t.Fatalf("error body not surfaced: %v", apiErr.Error())
			}
		})
	}
}

// TestDoRequest_UnmarshalFailureReturnsError verifies that a successful HTTP
// response with a malformed payload returns an error instead of being silently
// swallowed (previous behavior logged at Debug and returned zero-value).
func TestDoRequest_UnmarshalFailureReturnsError(t *testing.T) {
	client, _ := newQuotesClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{not-json`))
	})

	_, err := client.Quotes(context.Background(), "SPX", nil, nil)
	if err == nil {
		t.Fatalf("expected unmarshal error, got nil")
	}
	if !strings.Contains(err.Error(), "unmarshal") {
		t.Fatalf("expected unmarshal error mention, got: %v", err)
	}
}

// TestWithRoundTripper verifies the injected RoundTripper wraps every request.
func TestWithRoundTripper(t *testing.T) {
	var calls atomic.Int64
	base := http.DefaultTransport
	rt := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		calls.Add(1)
		return base.RoundTrip(req)
	})

	client, _ := newQuotesClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}, WithRoundTripper(rt))

	if _, err := client.Quotes(context.Background(), "SPX", nil, nil); err != nil {
		t.Fatalf("Quotes: %v", err)
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("RoundTripper invoked %d times, want 1", got)
	}
}

// roundTripperFunc adapts a func to http.RoundTripper.
type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

// TestExchangeCode_GrantFlow verifies ExchangeCode performs the
// authorization_code grant and returns a TokenRecord with raw tokens.
func TestExchangeCode_GrantFlow(t *testing.T) {
	var gotBasicAuth, gotGrantType, gotCode, gotRedirectURI string
	tm, srv := newTestTokenManager(t, func(w http.ResponseWriter, r *http.Request) {
		gotBasicAuth = r.Header.Get("Authorization")
		if err := r.ParseForm(); err != nil {
			t.Errorf("parse form: %v", err)
		}
		gotGrantType = r.Form.Get("grant_type")
		gotCode = r.Form.Get("code")
		gotRedirectURI = r.Form.Get("redirect_uri")

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "new-access-token",
			"refresh_token": "new-rotated-refresh-token",
			"id_token":      "new-id-token",
			"expires_in":    1800,
			"token_type":    "Bearer",
			"scope":         "api",
		})
	}, &memoryStorage{})

	rec, err := tm.ExchangeCode(context.Background(), "the-auth-code")
	if err != nil {
		t.Fatalf("ExchangeCode: %v", err)
	}

	if !strings.HasPrefix(gotBasicAuth, "Basic ") {
		t.Fatalf("expected Basic auth header, got %q", gotBasicAuth)
	}
	if gotGrantType != "authorization_code" {
		t.Fatalf("grant_type = %q, want authorization_code", gotGrantType)
	}
	if gotCode != "the-auth-code" {
		t.Fatalf("code = %q, want the-auth-code", gotCode)
	}
	if gotRedirectURI != "https://127.0.0.1/callback" {
		t.Fatalf("redirect_uri = %q", gotRedirectURI)
	}

	if rec.AccessToken != "new-access-token" {
		t.Fatalf("AccessToken = %q", rec.AccessToken)
	}
	if rec.RefreshToken != "new-rotated-refresh-token" {
		t.Fatalf("RefreshToken = %q", rec.RefreshToken)
	}
	if rec.IDToken != "new-id-token" {
		t.Fatalf("IDToken = %q", rec.IDToken)
	}
	if rec.ExpiresIn != 1800 {
		t.Fatalf("ExpiresIn = %d, want 1800", rec.ExpiresIn)
	}
	if rec.TokenType != "Bearer" {
		t.Fatalf("TokenType = %q, want Bearer", rec.TokenType)
	}
	if rec.Scope != "api" {
		t.Fatalf("Scope = %q, want api", rec.Scope)
	}

	// Both anchors are set to acquisition time.
	if rec.AccessTokenIssued.IsZero() || rec.RefreshTokenIssued.IsZero() {
		t.Fatalf("issued timestamps must be set: at=%v rt=%v", rec.AccessTokenIssued, rec.RefreshTokenIssued)
	}

	// ExchangeCode must NOT persist or mutate in-memory state (caller persists).
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	if tm.accessToken != "" {
		t.Fatalf("ExchangeCode mutated in-memory access token: %q", tm.accessToken)
	}

	_ = srv
}

// TestExchangeCode_EmptyCode verifies empty codes are rejected.
func TestExchangeCode_EmptyCode(t *testing.T) {
	tm, _ := newTestTokenManager(t, tokenGrantHandler(t), &memoryStorage{})
	_, err := tm.ExchangeCode(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty code, got nil")
	}
}