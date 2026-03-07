package schwabdev

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

const createTableSQL = `
CREATE TABLE IF NOT EXISTS schwabdev (
	access_token_issued  TEXT NOT NULL,
	refresh_token_issued TEXT NOT NULL,
	access_token         TEXT NOT NULL,
	refresh_token        TEXT NOT NULL,
	id_token             TEXT NOT NULL,
	expires_in           INTEGER,
	token_type           TEXT,
	scope                TEXT
);`

// tokenDB is the internal SQLite accessor used exclusively by TokenManager.
// It owns the single *sql.DB handle and exposes only the two operations the
// token layer needs: a serialisable write transaction and a read query.
//
// In-process concurrency is handled by TokenManager's sync.RWMutex.
// Cross-process exclusivity is handled by SQLite's busy_timeout pragma together
// with the serialisable transaction, which maps to "BEGIN IMMEDIATE" on SQLite
// (the highest isolation level the driver supports without raw SQL hacks).
type tokenDB struct {
	db   *sql.DB
	path string
}

// openTokenDB opens (or creates) the SQLite database at path.
// Path may be empty (defaults to ~/.schwabdev/tokens.db) or start with ~.
func openTokenDB(path string) (*tokenDB, error) {
	path = resolvedPath(path)

	if dir := filepath.Dir(path); dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("create db directory: %w", err)
		}
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	// Single writer to avoid "database is locked" from the driver's own pool.
	db.SetMaxOpenConns(1)

	if _, err := db.Exec(fmt.Sprintf("PRAGMA busy_timeout = %d", SQLiteBusyTimeout)); err != nil {
		db.Close()
		return nil, fmt.Errorf("set busy_timeout: %w", err)
	}

	if _, err := db.Exec(createTableSQL); err != nil {
		db.Close()
		return nil, fmt.Errorf("create table: %w", err)
	}

	return &tokenDB{db: db, path: path}, nil
}

// close releases the database connection.
func (d *tokenDB) close() error {
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}

// queryRow executes a read query and returns the single result row.
// The caller is responsible for scanning the result.
func (d *tokenDB) queryRow(ctx context.Context, query string, args ...any) *sql.Row {
	return d.db.QueryRowContext(ctx, query, args...)
}

// writeTokens replaces the single token row atomically inside a serialisable
// transaction. The callback receives the transaction and does the actual
// DELETE + INSERT so that this layer stays free of business logic.
func (d *tokenDB) writeTokens(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := d.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

// resolvedPath expands ~ and supplies the default path when empty.
func resolvedPath(path string) string {
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ".schwabdev/tokens.db"
		}
		return filepath.Join(home, ".schwabdev", "tokens.db")
	}
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}