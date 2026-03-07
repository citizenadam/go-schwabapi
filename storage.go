package schwabdev

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ── Public interface ──────────────────────────────────────────────────────────

// TokenRecord is the single value that TokenStorage reads and writes.
// It contains everything needed to restore a TokenManager's state.
type TokenRecord struct {
	AccessTokenIssued  time.Time `json:"access_token_issued"`
	RefreshTokenIssued time.Time `json:"refresh_token_issued"`
	AccessToken        string    `json:"access_token"`
	RefreshToken       string    `json:"refresh_token"`
	IDToken            string    `json:"id_token"`
	ExpiresIn          int       `json:"expires_in"`
	TokenType          string    `json:"token_type"`
	Scope              string    `json:"scope"`
}

// TokenStorage is the persistence interface for token data.
// Implement this to use any backend: file, Redis, Vault, a test stub, etc.
// The two methods are called from TokenManager under its own concurrency
// controls, so implementations do not need to be independently goroutine-safe
// unless they are shared across multiple TokenManager instances.
type TokenStorage interface {
	// Load retrieves the stored token record.
	// Returns (nil, nil) when no record exists yet (first run).
	Load(ctx context.Context) (*TokenRecord, error)

	// Save persists a token record, replacing any previous value atomically.
	Save(ctx context.Context, rec TokenRecord) error

	// Close releases any resources held by the storage (connections, file
	// handles, etc.). Called by TokenManager.Close().
	Close() error
}

// ── File-based implementation (default, stdlib only) ─────────────────────────

// FileTokenStorage stores tokens as a JSON file on disk.
// It is the default implementation used when no custom storage is provided.
// A per-instance mutex provides in-process safety; cross-process safety relies
// on the atomic write pattern (write to temp file, then rename).
type FileTokenStorage struct {
	path string
	mu   sync.Mutex
}

// NewFileTokenStorage creates a FileTokenStorage at the given path.
// Path may be empty (defaults to ~/.schwabdev/tokens.json) or start with ~.
func NewFileTokenStorage(path string) (*FileTokenStorage, error) {
	path = resolvedStoragePath(path)
	if dir := filepath.Dir(path); dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("create storage directory: %w", err)
		}
	}
	return &FileTokenStorage{path: path}, nil
}

// Load reads and JSON-decodes the token file.
// Returns (nil, nil) when the file does not exist yet.
func (f *FileTokenStorage) Load(_ context.Context) (*TokenRecord, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	data, err := os.ReadFile(f.path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read token file: %w", err)
	}

	var rec TokenRecord
	if err := json.Unmarshal(data, &rec); err != nil {
		return nil, fmt.Errorf("parse token file: %w", err)
	}
	return &rec, nil
}

// Save atomically writes the token record to disk via a temp-file + rename.
// This prevents a partial write from corrupting the stored tokens.
func (f *FileTokenStorage) Save(_ context.Context, rec TokenRecord) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	data, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal token record: %w", err)
	}

	// Write to a sibling temp file, then rename into place atomically.
	tmp := f.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return fmt.Errorf("write temp token file: %w", err)
	}
	if err := os.Rename(tmp, f.path); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("commit token file: %w", err)
	}
	return nil
}

// Close is a no-op for file storage — no persistent connections to release.
func (f *FileTokenStorage) Close() error { return nil }

// ── Path helper ───────────────────────────────────────────────────────────────

// resolvedStoragePath expands ~ and supplies a default when path is empty.
func resolvedStoragePath(path string) string {
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ".schwabdev/tokens.json"
		}
		return filepath.Join(home, ".schwabdev", "tokens.json")
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