package schwabdev

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// PostgresTokenStorage stores tokens in a PostgreSQL table.
//
// The table is created automatically on first use. A single row per
// app key is maintained via an upsert, making it safe to run multiple
// instances of the same app against the same database — the last writer
// wins, which is correct because Schwab issues one token set per app key.
//
// Usage:
//
//	db, err := sql.Open("postgres", "postgres://user:pass@host/dbname?sslmode=require")
//	storage, err := schwabdev.NewPostgresTokenStorage(db, "myapp_schwab_tokens")
//	tm, err := schwabdev.NewTokenManager(appKey, appSecret, callbackURL, storage, "", logger, nil)
type PostgresTokenStorage struct {
	db    *sql.DB
	table string
}

// NewPostgresTokenStorage creates a PostgresTokenStorage and ensures the table
// exists. table is the unqualified table name; include a schema prefix if
// needed (e.g. "myschema.schwab_tokens").
func NewPostgresTokenStorage(db *sql.DB, table string) (*PostgresTokenStorage, error) {
	s := &PostgresTokenStorage{db: db, table: table}
	if err := s.migrate(context.Background()); err != nil {
		return nil, fmt.Errorf("postgres token storage migrate: %w", err)
	}
	return s, nil
}

// NewPostgresTokenStorageWithoutMigrate creates a PostgresTokenStorage
// without attempting to create/migrate the table. Use this when the table
// is managed by external migrations (e.g., golang-migrate, db-mate).
//
// The caller is responsible for ensuring the table exists before calling Load/Save.
//
// Table must have the schema:
//
//	CREATE TABLE schwab_tokens (
//	    singleton            BOOLEAN PRIMARY KEY DEFAULT TRUE,
//	    access_token_issued  TIMESTAMPTZ NOT NULL,
//	    refresh_token_issued TIMESTAMPTZ NOT NULL,
//	    access_token         TEXT        NOT NULL,
//	    refresh_token        TEXT        NOT NULL,
//	    id_token             TEXT        NOT NULL DEFAULT '',
//	    expires_in           INTEGER,
//	    token_type           TEXT        DEFAULT 'Bearer',
//	    scope                TEXT        DEFAULT 'api',
//	    CONSTRAINT schwab_tokens_one_row CHECK (singleton)
//	);
//
// This is useful when your application uses a migration tool and the database
// user doesn't have CREATE TABLE permissions.
func NewPostgresTokenStorageWithoutMigrate(db *sql.DB, table string) *PostgresTokenStorage {
	return &PostgresTokenStorage{db: db, table: table}
}

// Load retrieves the token record from Postgres.
// Returns (nil, nil) when no row exists yet (first run).
func (s *PostgresTokenStorage) Load(ctx context.Context) (*TokenRecord, error) {
	query := fmt.Sprintf(`
		SELECT access_token_issued, refresh_token_issued,
		       access_token, refresh_token, id_token,
		       expires_in, token_type, scope
		FROM %s LIMIT 1`, s.table)

	row := s.db.QueryRowContext(ctx, query)

	var (
		rec                TokenRecord
		atIssued, rtIssued time.Time
		expiresIn          sql.NullInt64
		tokenType, scope   sql.NullString
	)

	err := row.Scan(
		&atIssued, &rtIssued,
		&rec.AccessToken, &rec.RefreshToken, &rec.IDToken,
		&expiresIn, &tokenType, &scope,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("load tokens: %w", err)
	}

	rec.AccessTokenIssued = atIssued.UTC()
	rec.RefreshTokenIssued = rtIssued.UTC()
	if expiresIn.Valid {
		rec.ExpiresIn = int(expiresIn.Int64)
	}
	if tokenType.Valid {
		rec.TokenType = tokenType.String
	}
	if scope.Valid {
		rec.Scope = scope.String
	}

	return &rec, nil
}

// Save upserts the token record. Because Schwab issues a single token set
// per app key, the table holds at most one row — older rows are replaced.
func (s *PostgresTokenStorage) Save(ctx context.Context, rec TokenRecord) error {
	// Use a transaction with SERIALIZABLE isolation to prevent a race where
	// two processes both read "no row" and then both try to INSERT.
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	upsert := fmt.Sprintf(`
		INSERT INTO %s
			(access_token_issued, refresh_token_issued,
			 access_token, refresh_token, id_token,
			 expires_in, token_type, scope)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (singleton)
		DO UPDATE SET
			access_token_issued  = EXCLUDED.access_token_issued,
			refresh_token_issued = EXCLUDED.refresh_token_issued,
			access_token         = EXCLUDED.access_token,
			refresh_token        = EXCLUDED.refresh_token,
			id_token             = EXCLUDED.id_token,
			expires_in           = EXCLUDED.expires_in,
			token_type           = EXCLUDED.token_type,
			scope                = EXCLUDED.scope`, s.table)

	_, err = tx.ExecContext(ctx, upsert,
		rec.AccessTokenIssued.UTC(),
		rec.RefreshTokenIssued.UTC(),
		rec.AccessToken,
		rec.RefreshToken,
		rec.IDToken,
		rec.ExpiresIn,
		rec.TokenType,
		rec.Scope,
	)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("upsert tokens: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tokens: %w", err)
	}
	return nil
}

// Close is a no-op — the caller owns the *sql.DB lifecycle.
// Close your *sql.DB separately when shutting down.
func (s *PostgresTokenStorage) Close() error { return nil }

// migrate creates the token table if it does not already exist.
// The singleton column + constraint enforces a single-row invariant at the
// database level, which the upsert ON CONFLICT clause targets.
func (s *PostgresTokenStorage) migrate(ctx context.Context) error {
	ddl := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			singleton            BOOLEAN PRIMARY KEY DEFAULT TRUE,
			access_token_issued  TIMESTAMPTZ NOT NULL,
			refresh_token_issued TIMESTAMPTZ NOT NULL,
			access_token         TEXT        NOT NULL,
			refresh_token        TEXT        NOT NULL,
			id_token             TEXT        NOT NULL,
			expires_in           INTEGER,
			token_type           TEXT,
			scope                TEXT,
			CONSTRAINT %s_one_row CHECK (singleton)
		)`, s.table, s.table)

	if _, err := s.db.ExecContext(ctx, ddl); err != nil {
		return fmt.Errorf("create table: %w", err)
	}
	return nil
}
