package schwabdev

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

// Backend names accepted by StorageConfig and NewStorageFromEnv.
const (
	BackendFile     = "file"
	BackendPostgres = "postgres"
)

// StorageConfig holds all parameters needed to construct any supported backend.
// Fields irrelevant to the chosen backend are ignored.
//
// Environment variable mapping (used by NewStorageFromEnv):
//
//	SCHWABDEV_STORAGE         → Backend       ("file", "postgres")
//	SCHWABDEV_STORAGE_PATH    → FilePath      (file backend only)
//	SCHWABDEV_DATABASE_URL    → PostgresDSN   (postgres backend only)
//	SCHWABDEV_STORAGE_TABLE   → PostgresTable (postgres backend only, default "schwab_tokens")
type StorageConfig struct {
	// Backend selects the storage implementation.
	// Valid values: "file" (default), "postgres".
	Backend string

	// FilePath is the path to the JSON token file used by the file backend.
	// Empty defaults to ~/.schwabdev/tokens.json.
	FilePath string

	// PostgresDSN is the libpq connection string for the postgres backend.
	// e.g. "postgres://user:pass@host:5432/dbname?sslmode=require"
	PostgresDSN string

	// PostgresTable is the table name for the postgres backend.
	// Defaults to "schwab_tokens". Include a schema prefix if needed
	// (e.g. "myschema.schwab_tokens").
	PostgresTable string
}

// NewStorageFromEnv reads the standard SCHWABDEV_* environment variables and
// returns a ready-to-use TokenStorage. It is the recommended entry point for
// applications that configure via environment.
//
// Supported variables:
//
//	SCHWABDEV_STORAGE         file | postgres            (default: file)
//	SCHWABDEV_STORAGE_PATH    path to token JSON file    (file backend)
//	SCHWABDEV_DATABASE_URL    postgres DSN               (postgres backend)
//	SCHWABDEV_STORAGE_TABLE   table name                 (postgres backend, default: schwab_tokens)
func NewStorageFromEnv() (TokenStorage, error) {
	cfg := StorageConfig{
		Backend:       strings.ToLower(strings.TrimSpace(os.Getenv("SCHWABDEV_STORAGE"))),
		FilePath:      os.Getenv("SCHWABDEV_STORAGE_PATH"),
		PostgresDSN:   os.Getenv("SCHWABDEV_DATABASE_URL"),
		PostgresTable: os.Getenv("SCHWABDEV_STORAGE_TABLE"),
	}
	return NewStorageFromConfig(cfg)
}

// NewStorageFromConfig constructs a TokenStorage from an explicit StorageConfig.
// Use this when configuration comes from a config file, flags, or code rather
// than environment variables.
func NewStorageFromConfig(cfg StorageConfig) (TokenStorage, error) {
	// Normalise and default the backend name.
	backend := strings.ToLower(strings.TrimSpace(cfg.Backend))
	if backend == "" {
		backend = BackendFile
	}

	switch backend {
	case BackendFile:
		return NewFileTokenStorage(cfg.FilePath)

	case BackendPostgres:
		if cfg.PostgresDSN == "" {
			return nil, fmt.Errorf(
				"storage backend %q requires a DSN: set SCHWABDEV_DATABASE_URL or StorageConfig.PostgresDSN",
				BackendPostgres,
			)
		}
		table := cfg.PostgresTable
		if table == "" {
			table = "schwab_tokens"
		}
		db, err := sql.Open("postgres", cfg.PostgresDSN)
		if err != nil {
			return nil, fmt.Errorf("open postgres: %w", err)
		}
		storage, err := NewPostgresTokenStorage(db, table)
		if err != nil {
			db.Close()
			return nil, err
		}
		// Wrap so that storage.Close() also closes the *sql.DB we opened here.
		// (When the caller supplies their own *sql.DB via NewPostgresTokenStorage
		// directly, they manage the lifecycle themselves — Close() is a no-op.)
		return &ownedPostgresStorage{PostgresTokenStorage: storage, db: db}, nil

	default:
		return nil, fmt.Errorf(
			"unknown storage backend %q: valid values are %q or %q",
			backend, BackendFile, BackendPostgres,
		)
	}
}

// ownedPostgresStorage wraps PostgresTokenStorage for the case where the
// factory opened the *sql.DB itself and therefore owns its lifecycle.
type ownedPostgresStorage struct {
	*PostgresTokenStorage
	db *sql.DB
}

func (o *ownedPostgresStorage) Close() error {
	return o.db.Close()
}