package schwabdev

import (
	"context"
	"testing"
	"time"
)

// ── Compile-time interface conformance ──────────────────────────────────────

// mockTokenProvider is a compile-time stub that verifies a concrete type
// satisfies the TokenSource interface. Named distinctly from the
// mockTokenSource already defined in roundtrip_test.go to avoid a duplicate
// declaration in the same package.
type mockTokenProvider struct {
	token   *Token
	getErr  error
	saveErr error
}

func (m *mockTokenProvider) GetToken(_ context.Context) (*Token, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.token, nil
}

func (m *mockTokenProvider) SaveToken(_ context.Context, t *Token) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.token = t
	return nil
}

// Compile-time check: *mockTokenProvider must satisfy TokenSource.
var _ TokenSource = (*mockTokenProvider)(nil)

// ── Token.Expired ────────────────────────────────────────────────────────────

func TestTokenInternal_Expired(t *testing.T) {
	t.Run("past expiry returns true", func(t *testing.T) {
		tok := &Token{
			AccessToken:  "abc123",
			RefreshToken: "refresh123",
			ExpiresAt:    time.Now().Add(-1 * time.Hour),
		}
		if !tok.Expired() {
			t.Error("Expired() = false, want true for a token whose ExpiresAt is in the past")
		}
	})

	t.Run("future expiry returns false", func(t *testing.T) {
		tok := &Token{
			AccessToken:  "abc123",
			RefreshToken: "refresh123",
			ExpiresAt:    time.Now().Add(1 * time.Hour),
		}
		if tok.Expired() {
			t.Error("Expired() = true, want false for a token whose ExpiresAt is in the future")
		}
	})
}

// ── Token.NeedsRefresh ────────────────────────────────────────────────────────

func TestTokenInternal_NeedsRefresh(t *testing.T) {
	t.Run("within buffer returns true", func(t *testing.T) {
		// Token expires in 30s; with a 60s buffer it should need a refresh.
		tok := &Token{
			AccessToken:  "abc123",
			RefreshToken: "refresh123",
			ExpiresAt:    time.Now().Add(30 * time.Second),
		}
		if !tok.NeedsRefresh(60 * time.Second) {
			t.Error("NeedsRefresh() = false, want true when expiry is within the buffer")
		}
	})

	t.Run("outside buffer returns false", func(t *testing.T) {
		// Token expires in 5m; with a 60s buffer it should NOT need a refresh yet.
		tok := &Token{
			AccessToken:  "abc123",
			RefreshToken: "refresh123",
			ExpiresAt:    time.Now().Add(5 * time.Minute),
		}
		if tok.NeedsRefresh(60 * time.Second) {
			t.Error("NeedsRefresh() = true, want false when expiry is outside the buffer")
		}
	})
}

// ── Token struct fields ───────────────────────────────────────────────────────

func TestTokenInternal_Fields(t *testing.T) {
	expiry := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	tok := &Token{
		AccessToken:  "access-value",
		RefreshToken: "refresh-value",
		ExpiresAt:    expiry,
	}

	if tok.AccessToken != "access-value" {
		t.Errorf("AccessToken = %q, want %q", tok.AccessToken, "access-value")
	}
	if tok.RefreshToken != "refresh-value" {
		t.Errorf("RefreshToken = %q, want %q", tok.RefreshToken, "refresh-value")
	}
	if !tok.ExpiresAt.Equal(expiry) {
		t.Errorf("ExpiresAt = %v, want %v", tok.ExpiresAt, expiry)
	}
}
