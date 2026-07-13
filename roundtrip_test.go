package schwabdev

import (
	"context"
	"testing"
	"time"
)

// mockTokenSource implements TokenSource for testing.
type mockTokenSource struct {
	token  *Token
	getErr error
	saveErr error
}

func (m *mockTokenSource) GetToken(ctx context.Context) (*Token, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.token, nil
}

func (m *mockTokenSource) SaveToken(ctx context.Context, t *Token) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.token = t
	return nil
}

func TestToken_NeedsRefresh(t *testing.T) {
	tests := []struct {
		name      string
		expiresIn time.Duration
		buffer   time.Duration
		want     bool
	}{
		{
			name:      "not expiring soon",
			expiresIn: 5 * time.Minute,
			buffer:   60 * time.Second,
			want:     false,
		},
		{
			name:      "needs refresh",
			expiresIn: 30 * time.Second,
			buffer:   60 * time.Second,
			want:     true,
		},
		{
			name:      "already expired",
			expiresIn: -10 * time.Second,
			buffer:   60 * time.Second,
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tok := &Token{
				AccessToken:  "test",
				RefreshToken: "refresh",
				ExpiresAt:    time.Now().Add(tt.expiresIn),
			}
			if got := tok.NeedsRefresh(tt.buffer); got != tt.want {
				t.Errorf("NeedsRefresh() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToken_Expired(t *testing.T) {
	tests := []struct {
		name      string
		expiresIn time.Duration
		want     bool
	}{
		{"not expired", 5 * time.Minute, false},
		{"expired", -10 * time.Second, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tok := &Token{
				AccessToken: "test",
				ExpiresAt:  time.Now().Add(tt.expiresIn),
			}
			if got := tok.Expired(); got != tt.want {
				t.Errorf("Expired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewSchwabRoundTripper_ConfigValidation(t *testing.T) {
	// Empty config should panic.
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic with empty config")
		}
	}()

	source := &mockTokenSource{token: &Token{AccessToken: "test", RefreshToken: "refresh"}}
	_ = NewSchwabRoundTripper(source, OAuthConfig{}, nil)
}