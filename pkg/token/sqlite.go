package token

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// DB represents the SQLite database for token persistence
type DB struct {
	conn   *sql.DB
	mu     sync.RWMutex
	logger Logger
}

// TokenData represents the token data stored in the database
type TokenData struct {
	AccessTokenIssued  time.Time
	RefreshTokenIssued time.Time
	AccessToken        string
	RefreshToken       string
	IDToken            string
	ExpiresIn          int
	TokenType          string
	Scope              string
}

// Logger interface for logging operations
type Logger interface {
	Info(msg string)
	Error(msg string)
	Debug(msg string)
}

// NewDB creates a new token database connection
func NewDB(dbPath string, logger Logger) (*DB, error) {
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db := &DB{
		conn:   conn,
		logger: logger,
	}

	// Create schema
	if err := db.createSchema(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	return db, nil
}

// createSchema creates the token table if it doesn't exist
func (db *DB) createSchema() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	query := `
	CREATE TABLE IF NOT EXISTS schwabdev (
		access_token_issued TEXT NOT NULL,
		refresh_token_issued TEXT NOT NULL,
		access_token TEXT NOT NULL,
		refresh_token TEXT NOT NULL,
		id_token TEXT NOT NULL,
		expires_in INTEGER,
		token_type TEXT,
		scope TEXT
	);
	`

	_, err := db.conn.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	// Set busy timeout to 30 seconds
	_, err = db.conn.Exec("PRAGMA busy_timeout = 30000;")
	if err != nil {
		return fmt.Errorf("failed to set busy timeout: %w", err)
	}

	return nil
}

// LoadTokens loads tokens from the database
func (db *DB) LoadTokens() (*TokenData, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	query := `
	SELECT
		access_token_issued,
		refresh_token_issued,
		access_token,
		refresh_token,
		id_token,
		expires_in,
		token_type,
		scope
	FROM schwabdev
	LIMIT 1
	`

	var (
		atIssuedStr  string
		rtIssuedStr  string
		accessToken  string
		refreshToken string
		idToken      string
		expiresIn    sql.NullInt64
		tokenType    sql.NullString
		scope        sql.NullString
	)

	err := db.conn.QueryRow(query).Scan(
		&atIssuedStr,
		&rtIssuedStr,
		&accessToken,
		&refreshToken,
		&idToken,
		&expiresIn,
		&tokenType,
		&scope,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load tokens: %w", err)
	}

	// Parse timestamps
	atIssued, err := time.Parse(time.RFC3339, atIssuedStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse access_token_issued: %w", err)
	}

	rtIssued, err := time.Parse(time.RFC3339, rtIssuedStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse refresh_token_issued: %w", err)
	}

	data := &TokenData{
		AccessTokenIssued:  atIssued,
		RefreshTokenIssued: rtIssued,
		AccessToken:        accessToken,
		RefreshToken:       refreshToken,
		IDToken:            idToken,
		ExpiresIn:          int(expiresIn.Int64),
		TokenType:          tokenType.String,
		Scope:              scope.String,
	}

	return data, nil
}

// SaveTokens saves tokens to the database
func (db *DB) SaveTokens(data *TokenData) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	tx, err := db.conn.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Ensure rollback on error
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Delete existing row
	_, err = tx.Exec("DELETE FROM schwabdev")
	if err != nil {
		return fmt.Errorf("failed to delete existing tokens: %w", err)
	}

	// Insert new tokens
	query := `
	INSERT INTO schwabdev (
		access_token_issued,
		refresh_token_issued,
		access_token,
		refresh_token,
		id_token,
		expires_in,
		token_type,
		scope
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = tx.Exec(query,
		data.AccessTokenIssued.Format(time.RFC3339),
		data.RefreshTokenIssued.Format(time.RFC3339),
		data.AccessToken,
		data.RefreshToken,
		data.IDToken,
		data.ExpiresIn,
		data.TokenType,
		data.Scope,
	)

	if err != nil {
		return fmt.Errorf("failed to insert tokens: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}
