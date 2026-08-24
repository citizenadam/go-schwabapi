package schwabdev

import (
	"context"
	"time"
)

// TokenSource defines the interface for supplying and storing access tokens.
// Implement this to use external token storage (database, Vault, environment, file, etc.)
//
// The RoundTripper calls GetToken before each request and SaveToken
// after a successful OAuth refresh.
//
// Note: This is separate from the Streamer's TokenProvider interface
// which only returns AccessToken() string.
type TokenSource interface {
	// GetToken retrieves the current access token, refresh token, and expiry.
	// Return (nil, nil) if no token exists yet.
	// Return error only on storage failures (not token expiration).
	GetToken(ctx context.Context) (*Token, error)

	// SaveToken persists refreshed tokens back to storage.
	// Called by RoundTripper after successful OAuth refresh.
	SaveToken(ctx context.Context, token *Token) error
}

// Token represents OAuth tokens with expiry metadata.
type Token struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

// Expired returns true if the token has passed its expiry time.
func (t *Token) Expired() bool {
	return time.Now().After(t.ExpiresAt)
}

// NeedsRefresh returns true if the token expires within the given buffer.
// Use this to trigger refresh before actual expiration.
func (t *Token) NeedsRefresh(buffer time.Duration) bool {
	return time.Until(t.ExpiresAt) < buffer
}
