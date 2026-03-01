package token

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// slogLoggerAdapter adapts slog.Logger to the Logger interface
type slogLoggerAdapter struct {
	logger *slog.Logger
}

func (a *slogLoggerAdapter) Info(msg string) {
	a.logger.Info(msg)
}

func (a *slogLoggerAdapter) Error(msg string) {
	a.logger.Error(msg)
}

func (a *slogLoggerAdapter) Debug(msg string) {
	a.logger.Debug(msg)
}

// Manager manages token lifecycle with auto-refresh
type Manager struct {
	db                 *DB
	logger             *slog.Logger
	mu                 sync.RWMutex
	accessToken        string
	refreshToken       string
	idToken            string
	accessTokenIssued  time.Time
	refreshTokenIssued time.Time
	expiresIn          int
	tokenType          string
	scope              string
}

// NewManager creates a new token manager
func NewManager(dbPath string, logger *slog.Logger) (*Manager, error) {
	adapter := &slogLoggerAdapter{logger: logger}

	db, err := NewDB(dbPath, adapter)
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	m := &Manager{
		db:     db,
		logger: logger,
	}

	// Load tokens from database
	if err := m.loadTokens(); err != nil {
		logger.Warn("Could not load tokens from database", "error", err.Error())
	}

	return m, nil
}

// loadTokens loads tokens from database into memory
func (m *Manager) loadTokens() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.loadTokensLocked()
}

// loadTokensLocked loads tokens from database into memory (must be called with lock held)
func (m *Manager) loadTokensLocked() error {
	data, err := m.db.LoadTokens()
	if err != nil {
		return fmt.Errorf("failed to load tokens: %w", err)
	}

	if data == nil {
		// No tokens in database, initialize with empty values
		m.accessToken = ""
		m.refreshToken = ""
		m.idToken = ""
		m.accessTokenIssued = time.Time{}
		m.refreshTokenIssued = time.Time{}
		m.expiresIn = 0
		m.tokenType = ""
		m.scope = ""
		return nil
	}

	m.accessToken = data.AccessToken
	m.refreshToken = data.RefreshToken
	m.idToken = data.IDToken
	m.accessTokenIssued = data.AccessTokenIssued
	m.refreshTokenIssued = data.RefreshTokenIssued
	m.expiresIn = data.ExpiresIn
	m.tokenType = data.TokenType
	m.scope = data.Scope

	return nil
}

// saveTokens saves tokens from memory to database (must be called with lock held)
func (m *Manager) saveTokens() error {
	data := &TokenData{
		AccessToken:        m.accessToken,
		RefreshToken:       m.refreshToken,
		IDToken:            m.idToken,
		AccessTokenIssued:  m.accessTokenIssued,
		RefreshTokenIssued: m.refreshTokenIssued,
		ExpiresIn:          m.expiresIn,
		TokenType:          m.tokenType,
		Scope:              m.scope,
	}

	return m.db.SaveTokens(data)
}

// UpdateTokens checks if tokens need to be updated and updates if needed
// Returns true if tokens were updated, false otherwise
func (m *Manager) UpdateTokens(ctx context.Context, forceAccessToken bool, forceRefreshToken bool) (bool, error) {
	// If no tokens exist and not forcing update, no update needed
	if m.refreshTokenIssued.IsZero() && !forceRefreshToken && !forceAccessToken {
		return false, nil
	}

	// Check for forced refresh token update first
	if forceRefreshToken {
		m.logger.Info("Forcing refresh token update")
		if err := m.updateRefreshToken(ctx); err != nil {
			return false, fmt.Errorf("failed to update refresh token: %w", err)
		}
		return true, nil
	}

	// Check for forced access token update
	if forceAccessToken {
		m.logger.Debug("Forcing access token update")
		if err := m.updateAccessToken(ctx); err != nil {
			return false, fmt.Errorf("failed to update access token: %w", err)
		}
		return true, nil
	}

	// Calculate time remaining for each token
	rtDelta := time.Until(m.refreshTokenIssued.Add(7 * 24 * time.Hour)) // 7 days from Schwab
	atDelta := time.Until(m.accessTokenIssued.Add(time.Duration(m.expiresIn) * time.Second))

	// Check if we need to update refresh token (and access token)
	if rtDelta < RefreshThreshold {
		msg := "Refresh token is expiring soon (<60min)!"
		if rtDelta < 0 {
			msg = "Refresh token has expired!"
		}
		m.logger.Warn(msg)
		if err := m.updateRefreshToken(ctx); err != nil {
			return false, fmt.Errorf("failed to update refresh token: %w", err)
		}
		return true, nil
	}

	// Check if we need to update access token
	if atDelta < AccessThreshold {
		m.logger.Debug("Access token has expired, updating...")
		if err := m.updateAccessToken(ctx); err != nil {
			return false, fmt.Errorf("failed to update access token: %w", err)
		}
		return true, nil
	}

	return false, nil
}

// updateAccessToken refreshes the access token using the refresh token
func (m *Manager) updateAccessToken(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Store last known access token issued time
	lastKnownATIssued := m.accessTokenIssued

	// Reload tokens from database to check if another instance updated them
	if err := m.loadTokensLocked(); err != nil {
		return err
	}

	// Check if another instance already updated the access token
	if m.accessTokenIssued.After(lastKnownATIssued) {
		m.logger.Info("Access token updated elsewhere", "time", m.accessTokenIssued.Format(time.RFC3339))
		return nil
	}

	// TODO: Implement actual OAuth refresh token call
	// This would involve making an HTTP POST request to Schwab's OAuth endpoint
	// For now, we'll just update the timestamp to simulate a refresh

	m.accessTokenIssued = time.Now().UTC()
	m.expiresIn = 1800 // 30 minutes from Schwab

	// Save updated tokens to database
	if err := m.saveTokens(); err != nil {
		return fmt.Errorf("failed to save tokens: %w", err)
	}

	m.logger.Info("Access token updated", "time", m.accessTokenIssued.Format(time.RFC3339))

	return nil
}

// updateRefreshToken gets new access and refresh tokens using authorization code
func (m *Manager) updateRefreshToken(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Store last known refresh token issued time
	lastKnownRTIssued := m.refreshTokenIssued

	// Reload tokens from database to check if another instance updated them
	if err := m.loadTokensLocked(); err != nil {
		return err
	}

	// Check if another instance already updated the refresh token
	if m.refreshTokenIssued.After(lastKnownRTIssued) {
		m.logger.Info("Refresh token updated elsewhere", "time", m.refreshTokenIssued.Format(time.RFC3339))
		return nil
	}

	// TODO: Implement actual OAuth authorization code flow
	// This would involve:
	// 1. Opening browser for user authorization
	// 2. Getting authorization code from callback URL
	// 3. Making HTTP POST request to Schwab's OAuth endpoint
	// For now, we'll just update the timestamps to simulate a refresh

	now := time.Now().UTC()
	m.accessTokenIssued = now
	m.refreshTokenIssued = now
	m.expiresIn = 1800 // 30 minutes from Schwab

	// Save updated tokens to database
	if err := m.saveTokens(); err != nil {
		return fmt.Errorf("failed to save tokens: %w", err)
	}

	m.logger.Info("Tokens updated", "time", now.Format(time.RFC3339))

	return nil
}

// GetAccessToken returns the current access token with auto-refresh
func (m *Manager) GetAccessToken() string {
	m.mu.RLock()
	accessTokenIssued := m.accessTokenIssued
	expiresIn := m.expiresIn
	accessToken := m.accessToken
	m.mu.RUnlock()

	timeUntilExpiry := time.Until(accessTokenIssued.Add(time.Duration(expiresIn) * time.Second))

	if timeUntilExpiry < AccessThreshold && !accessTokenIssued.IsZero() && accessToken != "" {
		m.logger.Debug("Access token expiring soon, auto-refreshing",
			"expires_in", timeUntilExpiry.Seconds(),
		)
		if _, err := m.UpdateTokens(context.Background(), true, false); err != nil {
			m.logger.Error("Failed to auto-refresh access token", "error", err)
		} else {
			m.mu.RLock()
			accessToken = m.accessToken
			m.mu.RUnlock()
		}
	}

	return accessToken
}

// GetRefreshToken returns the current refresh token
func (m *Manager) GetRefreshToken() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.refreshToken
}

// GetIDToken returns the current ID token
func (m *Manager) GetIDToken() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.idToken
}

// GetAccessTokenIssued returns the access token issued time
func (m *Manager) GetAccessTokenIssued() time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.accessTokenIssued
}

// GetRefreshTokenIssued returns the refresh token issued time
func (m *Manager) GetRefreshTokenIssued() time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.refreshTokenIssued
}

// Close closes the database connection
func (m *Manager) Close() error {
	return m.db.Close()
}
