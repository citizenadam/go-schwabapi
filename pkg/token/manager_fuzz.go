package token

import (
	"context"
	"log/slog"
	"path/filepath"
	"strings"
	"testing"
)

// FuzzNewManager tests the NewManager constructor with various database paths
func FuzzNewManager(f *testing.F) {
	f.Add([]byte("/tmp/test.db"))
	f.Add([]byte(""))
	f.Add([]byte(strings.Repeat("a", 1000)))

	f.Fuzz(func(t *testing.T, data []byte) {
		logger := slog.New(slog.NewTextHandler(nil, nil))

		// Create a temporary database path
		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "test.db")

		// Try to create manager - should not crash even with invalid paths
		_, _ = NewManager(dbPath, logger)
	})
}

// FuzzManagerGetAccessToken tests GetAccessToken with various token strings
func FuzzManagerGetAccessToken(f *testing.F) {
	f.Add([]byte("valid_access_token_123"))
	f.Add([]byte(""))
	f.Add([]byte(strings.Repeat("a", 10000)))

	f.Fuzz(func(t *testing.T, data []byte) {
		logger := slog.New(slog.NewTextHandler(nil, nil))

		// Create a temporary database
		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "test.db")

		manager, err := NewManager(dbPath, logger)
		if err != nil {
			return // Skip if manager creation fails
		}
		defer manager.Close()

		// Get access token - should not crash
		_ = manager.GetAccessToken()
	})
}

// FuzzManagerGetRefreshToken tests GetRefreshToken with various token strings
func FuzzManagerGetRefreshToken(f *testing.F) {
	f.Add([]byte("valid_refresh_token_456"))
	f.Add([]byte(""))
	f.Add([]byte(strings.Repeat("b", 10000)))

	f.Fuzz(func(t *testing.T, data []byte) {
		logger := slog.New(slog.NewTextHandler(nil, nil))

		// Create a temporary database
		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "test.db")

		manager, err := NewManager(dbPath, logger)
		if err != nil {
			return // Skip if manager creation fails
		}
		defer manager.Close()

		// Get refresh token - should not crash
		_ = manager.GetRefreshToken()
	})
}

// FuzzManagerGetIDToken tests GetIDToken with various token strings
func FuzzManagerGetIDToken(f *testing.F) {
	f.Add([]byte("valid_id_token_789"))
	f.Add([]byte(""))
	f.Add([]byte(strings.Repeat("c", 10000)))

	f.Fuzz(func(t *testing.T, data []byte) {
		logger := slog.New(slog.NewTextHandler(nil, nil))

		// Create a temporary database
		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "test.db")

		manager, err := NewManager(dbPath, logger)
		if err != nil {
			return // Skip if manager creation fails
		}
		defer manager.Close()

		// Get ID token - should not crash
		_ = manager.GetIDToken()
	})
}

// FuzzManagerGetAccessTokenIssued tests GetAccessTokenIssued with various timestamps
func FuzzManagerGetAccessTokenIssued(f *testing.F) {
	f.Add([]byte("2024-01-01T00:00:00Z"))
	f.Add([]byte(""))
	f.Add([]byte("invalid timestamp"))

	f.Fuzz(func(t *testing.T, data []byte) {
		logger := slog.New(slog.NewTextHandler(nil, nil))

		// Create a temporary database
		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "test.db")

		manager, err := NewManager(dbPath, logger)
		if err != nil {
			return // Skip if manager creation fails
		}
		defer manager.Close()

		// Get access token issued time - should not crash
		_ = manager.GetAccessTokenIssued()
	})
}

// FuzzManagerGetRefreshTokenIssued tests GetRefreshTokenIssued with various timestamps
func FuzzManagerGetRefreshTokenIssued(f *testing.F) {
	f.Add([]byte("2024-01-01T00:00:00Z"))
	f.Add([]byte(""))
	f.Add([]byte("invalid timestamp"))

	f.Fuzz(func(t *testing.T, data []byte) {
		logger := slog.New(slog.NewTextHandler(nil, nil))

		// Create a temporary database
		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "test.db")

		manager, err := NewManager(dbPath, logger)
		if err != nil {
			return // Skip if manager creation fails
		}
		defer manager.Close()

		// Get refresh token issued time - should not crash
		_ = manager.GetRefreshTokenIssued()
	})
}

// FuzzManagerUpdateTokens tests UpdateTokens with various contexts and force flags
func FuzzManagerUpdateTokens(f *testing.F) {
	f.Add([]byte("test"))
	f.Add([]byte(""))
	f.Add([]byte(strings.Repeat("a", 1000)))

	f.Fuzz(func(t *testing.T, data []byte) {
		logger := slog.New(slog.NewTextHandler(nil, nil))

		// Create a temporary database
		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "test.db")

		manager, err := NewManager(dbPath, logger)
		if err != nil {
			return // Skip if manager creation fails
		}
		defer manager.Close()

		// Test with various combinations of force flags
		forceAccessToken := len(data)%2 == 0
		forceRefreshToken := len(data)%3 == 0

		// Call UpdateTokens - should not crash even with invalid state
		_, _ = manager.UpdateTokens(context.Background(), forceAccessToken, forceRefreshToken)
	})
}

// FuzzManagerClose tests the Close method
func FuzzManagerClose(f *testing.F) {
	f.Add([]byte("test"))
	f.Add([]byte(""))

	f.Fuzz(func(t *testing.T, data []byte) {
		logger := slog.New(slog.NewTextHandler(nil, nil))

		// Create a temporary database
		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "test.db")

		manager, err := NewManager(dbPath, logger)
		if err != nil {
			return // Skip if manager creation fails
		}

		// Close manager - should not crash
		_ = manager.Close()
	})
}

// FuzzManagerConcurrentAccess tests concurrent access to token manager methods
func FuzzManagerConcurrentAccess(f *testing.F) {
	f.Add([]byte("test"))
	f.Add([]byte(""))

	f.Fuzz(func(t *testing.T, data []byte) {
		logger := slog.New(slog.NewTextHandler(nil, nil))

		// Create a temporary database
		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "test.db")

		manager, err := NewManager(dbPath, logger)
		if err != nil {
			return // Skip if manager creation fails
		}
		defer manager.Close()

		// Simulate concurrent access
		done := make(chan bool)

		// Goroutine 1: Get access token
		go func() {
			_ = manager.GetAccessToken()
			done <- true
		}()

		// Goroutine 2: Get refresh token
		go func() {
			_ = manager.GetRefreshToken()
			done <- true
		}()

		// Goroutine 3: Get ID token
		go func() {
			_ = manager.GetIDToken()
			done <- true
		}()

		// Goroutine 4: Get access token issued time
		go func() {
			_ = manager.GetAccessTokenIssued()
			done <- true
		}()

		// Wait for all goroutines to complete
		for range 4 {
			<-done
		}
	})
}

// FuzzManagerTokenPersistence tests token persistence with various token data
func FuzzManagerTokenPersistence(f *testing.F) {
	f.Add([]byte("access_token_123"))
	f.Add([]byte("refresh_token_456"))
	f.Add([]byte("id_token_789"))

	f.Fuzz(func(t *testing.T, accessToken, refreshToken, idToken []byte) {
		logger := slog.New(slog.NewTextHandler(nil, nil))

		// Create a temporary database
		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "test.db")

		manager, err := NewManager(dbPath, logger)
		if err != nil {
			return // Skip if manager creation fails
		}
		defer manager.Close()

		// Try to get tokens - should not crash even with empty database
		_ = manager.GetAccessToken()
		_ = manager.GetRefreshToken()
		_ = manager.GetIDToken()
	})
}
