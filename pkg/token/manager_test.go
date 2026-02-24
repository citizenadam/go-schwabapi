package token

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log/slog"
)

func TestNewManager(t *testing.T) {
	// Create a temporary database file
	tmpFile, err := os.CreateTemp("", "testdb*.db")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	logger := slog.Default()
	manager, err := NewManager(tmpFile.Name(), logger)

	assert.NoError(t, err)
	assert.NotNil(t, manager)
	assert.NotNil(t, manager.db)
	assert.Equal(t, logger, manager.logger)
	assert.NoError(t, manager.Close())
}

func TestNewManager_DatabaseError(t *testing.T) {
	// Try to create manager with invalid path
	logger := slog.Default()
	manager, err := NewManager("/invalid/path/to/db.db", logger)

	assert.Error(t, err)
	assert.Nil(t, manager)
	assert.Contains(t, err.Error(), "failed to create database")
}

func TestManager_GetAccessToken(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "testdb*.db")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	logger := slog.Default()
	manager, err := NewManager(tmpFile.Name(), logger)
	require.NoError(t, err)
	defer manager.Close()

	// Initially empty
	token := manager.GetAccessToken()
	assert.Empty(t, token)
}

func TestManager_GetRefreshToken(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "testdb*.db")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	logger := slog.Default()
	manager, err := NewManager(tmpFile.Name(), logger)
	require.NoError(t, err)
	defer manager.Close()

	// Initially empty
	token := manager.GetRefreshToken()
	assert.Empty(t, token)
}

func TestManager_GetIDToken(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "testdb*.db")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	logger := slog.Default()
	manager, err := NewManager(tmpFile.Name(), logger)
	require.NoError(t, err)
	defer manager.Close()

	// Initially empty
	token := manager.GetIDToken()
	assert.Empty(t, token)
}

func TestManager_GetAccessTokenIssued(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "testdb*.db")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	logger := slog.Default()
	manager, err := NewManager(tmpFile.Name(), logger)
	require.NoError(t, err)
	defer manager.Close()

	// Initially zero time
	issued := manager.GetAccessTokenIssued()
	assert.True(t, issued.IsZero())
}

func TestManager_GetRefreshTokenIssued(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "testdb*.db")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	logger := slog.Default()
	manager, err := NewManager(tmpFile.Name(), logger)
	require.NoError(t, err)
	defer manager.Close()

	// Initially zero time
	issued := manager.GetRefreshTokenIssued()
	assert.True(t, issued.IsZero())
}

func TestManager_UpdateTokens_NoUpdateNeeded(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "testdb*.db")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	logger := slog.Default()
	manager, err := NewManager(tmpFile.Name(), logger)
	require.NoError(t, err)
	defer manager.Close()

	ctx := context.Background()
	updated, err := manager.UpdateTokens(ctx, false, false)

	assert.NoError(t, err)
	assert.False(t, updated)
}

func TestManager_UpdateTokens_ForceAccessToken(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "testdb*.db")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	logger := slog.Default()
	manager, err := NewManager(tmpFile.Name(), logger)
	require.NoError(t, err)
	defer manager.Close()

	ctx := context.Background()
	updated, err := manager.UpdateTokens(ctx, true, false)

	assert.NoError(t, err)
	assert.True(t, updated)

	// Verify access token issued time was updated
	issued := manager.GetAccessTokenIssued()
	assert.False(t, issued.IsZero())
}

func TestManager_UpdateTokens_ForceRefreshToken(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "testdb*.db")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	logger := slog.Default()
	manager, err := NewManager(tmpFile.Name(), logger)
	require.NoError(t, err)
	defer manager.Close()

	ctx := context.Background()
	updated, err := manager.UpdateTokens(ctx, false, true)

	assert.NoError(t, err)
	assert.True(t, updated)

	// Verify both token issued times were updated
	atIssued := manager.GetAccessTokenIssued()
	rtIssued := manager.GetRefreshTokenIssued()
	assert.False(t, atIssued.IsZero())
	assert.False(t, rtIssued.IsZero())
}

func TestManager_UpdateTokens_AutoRefreshAccessToken(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "testdb*.db")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	logger := slog.Default()
	manager, err := NewManager(tmpFile.Name(), logger)
	require.NoError(t, err)
	defer manager.Close()

	// Set up tokens that are about to expire
	now := time.Now().UTC()
	manager.accessTokenIssued = now.Add(-30 * time.Minute) // 30 minutes ago
	manager.expiresIn = 1800                               // 30 minutes
	manager.refreshTokenIssued = now.Add(-1 * time.Hour)   // 1 hour ago

	ctx := context.Background()
	updated, err := manager.UpdateTokens(ctx, false, false)

	assert.NoError(t, err)
	assert.True(t, updated)
}

func TestManager_UpdateTokens_AutoRefreshRefreshToken(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "testdb*.db")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	logger := slog.Default()
	manager, err := NewManager(tmpFile.Name(), logger)
	require.NoError(t, err)
	defer manager.Close()

	// Set up refresh token that is about to expire
	now := time.Now().UTC()
	manager.accessTokenIssued = now.Add(-10 * time.Minute)
	manager.expiresIn = 1800
	manager.refreshTokenIssued = now.Add(-(7*24*time.Hour + 1*time.Hour)) // Just over 7 days ago

	ctx := context.Background()
	updated, err := manager.UpdateTokens(ctx, false, false)

	assert.NoError(t, err)
	assert.True(t, updated)
}

func TestManager_Close(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "testdb*.db")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	logger := slog.Default()
	manager, err := NewManager(tmpFile.Name(), logger)
	require.NoError(t, err)

	// Close should not error
	err = manager.Close()
	assert.NoError(t, err)

	// Double close should not error
	err = manager.Close()
	assert.NoError(t, err)
}

func TestManager_ConcurrentAccess(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "testdb*.db")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	logger := slog.Default()
	manager, err := NewManager(tmpFile.Name(), logger)
	require.NoError(t, err)
	defer manager.Close()

	// Test concurrent reads
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			_ = manager.GetAccessToken()
			_ = manager.GetRefreshToken()
			_ = manager.GetIDToken()
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestManager_UpdateTokens_Concurrent(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "testdb*.db")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	logger := slog.Default()
	manager, err := NewManager(tmpFile.Name(), logger)
	require.NoError(t, err)
	defer manager.Close()

	ctx := context.Background()
	done := make(chan bool)

	// Test concurrent updates
	for i := 0; i < 5; i++ {
		go func() {
			_, _ = manager.UpdateTokens(ctx, true, false)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 5; i++ {
		<-done
	}
}
