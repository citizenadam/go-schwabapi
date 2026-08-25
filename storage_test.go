package schwabdev

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestStorageFileSaveLoadRoundtrip verifies that a record written via Save can
// be read back unchanged via Load.
func TestStorageFileSaveLoadRoundtrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "tokens.json")

	storage, err := NewFileTokenStorage(path)
	if err != nil {
		t.Fatalf("NewFileTokenStorage: %v", err)
	}
	t.Cleanup(func() { _ = storage.Close() })

	want := TokenRecord{
		AccessTokenIssued:  time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		RefreshTokenIssued: time.Date(2024, 1, 14, 8, 0, 0, 0, time.UTC),
		AccessToken:        "access-abc-123",
		RefreshToken:       "refresh-xyz-789",
		IDToken:            "id-token-456",
		ExpiresIn:          1800,
		TokenType:          "Bearer",
		Scope:              "api",
	}

	ctx := context.Background()
	if err := storage.Save(ctx, want); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := storage.Load(ctx)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got == nil {
		t.Fatal("Load returned nil record")
	}
	if *got != want {
		t.Errorf("roundtrip mismatch:\n got  %+v\n want %+v", *got, want)
	}
}

// TestStorageFileLoadMissingFile verifies that Load returns (nil, nil) when the
// storage file does not exist yet (first-run scenario).
func TestStorageFileLoadMissingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nonexistent", "tokens.json")

	storage, err := NewFileTokenStorage(path)
	if err != nil {
		t.Fatalf("NewFileTokenStorage: %v", err)
	}
	t.Cleanup(func() { _ = storage.Close() })

	got, err := storage.Load(context.Background())
	if err != nil {
		t.Fatalf("Load on missing file returned error: %v", err)
	}
	if got != nil {
		t.Errorf("Load on missing file should return nil record, got %+v", got)
	}
}

// TestStorageFileSavePermissions verifies that the persisted file is created
// with 0600 permissions (read/write owner only).
func TestStorageFileSavePermissions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tokens.json")

	storage, err := NewFileTokenStorage(path)
	if err != nil {
		t.Fatalf("NewFileTokenStorage: %v", err)
	}
	t.Cleanup(func() { _ = storage.Close() })

	rec := TokenRecord{
		AccessToken:  "a",
		RefreshToken: "r",
		ExpiresIn:    1800,
		TokenType:    "Bearer",
		Scope:        "api",
	}
	if err := storage.Save(context.Background(), rec); err != nil {
		t.Fatalf("Save: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	const wantPerm os.FileMode = 0600
	if perm := info.Mode().Perm(); perm != wantPerm {
		t.Errorf("file permission: got %o, want %o", perm, wantPerm)
	}
}

// TestStorageFileOverwrite verifies that a second Save replaces the previous
// record on disk.
func TestStorageFileOverwrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tokens.json")

	storage, err := NewFileTokenStorage(path)
	if err != nil {
		t.Fatalf("NewFileTokenStorage: %v", err)
	}
	t.Cleanup(func() { _ = storage.Close() })

	ctx := context.Background()

	first := TokenRecord{AccessToken: "first", ExpiresIn: 100}
	if err := storage.Save(ctx, first); err != nil {
		t.Fatalf("Save first: %v", err)
	}

	second := TokenRecord{AccessToken: "second", ExpiresIn: 200}
	if err := storage.Save(ctx, second); err != nil {
		t.Fatalf("Save second: %v", err)
	}

	got, err := storage.Load(ctx)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got == nil {
		t.Fatal("Load returned nil record")
	}
	if got.AccessToken != "second" {
		t.Errorf("expected overwritten record (second), got %q", got.AccessToken)
	}
	if got.ExpiresIn != 200 {
		t.Errorf("expected ExpiresIn 200, got %d", got.ExpiresIn)
	}
}

// TestStorageFileClose verifies that Close is a no-op and returns nil.
func TestStorageFileClose(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tokens.json")

	storage, err := NewFileTokenStorage(path)
	if err != nil {
		t.Fatalf("NewFileTokenStorage: %v", err)
	}
	if err := storage.Close(); err != nil {
		t.Errorf("Close returned error: %v", err)
	}
}

// TestStorageTokenRecordJSONRoundtrip verifies that TokenRecord marshals and
// unmarshals through JSON with the expected field names.
func TestStorageTokenRecordJSONRoundtrip(t *testing.T) {
	want := TokenRecord{
		AccessTokenIssued:  time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC),
		RefreshTokenIssued: time.Date(2024, 5, 31, 9, 15, 30, 0, time.UTC),
		AccessToken:        "at-value",
		RefreshToken:       "rt-value",
		IDToken:            "id-value",
		ExpiresIn:          1800,
		TokenType:          "Bearer",
		Scope:              "api scope",
	}

	data, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	// Verify expected JSON keys are present.
	jsonStr := string(data)
	for _, key := range []string{
		`"access_token_issued"`,
		`"refresh_token_issued"`,
		`"access_token"`,
		`"refresh_token"`,
		`"id_token"`,
		`"expires_in"`,
		`"token_type"`,
		`"scope"`,
	} {
		if !strings.Contains(jsonStr, key) {
			t.Errorf("JSON output missing key %s: %s", key, jsonStr)
		}
	}

	var got TokenRecord
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got != want {
		t.Errorf("roundtrip mismatch:\n got  %+v\n want %+v", got, want)
	}
}

// TestStorageResolvedPathDefault verifies that an empty path resolves to the
// default location under the user's home directory.
func TestStorageResolvedPathDefault(t *testing.T) {
	got := resolvedStoragePath("")

	home, err := os.UserHomeDir()
	if err != nil {
		// If home dir is unavailable the function falls back to a relative path.
		want := ".schwabdev/tokens.json"
		if got != want {
			t.Errorf("expected fallback %q, got %q", want, got)
		}
		return
	}

	want := filepath.Join(home, ".schwabdev", "tokens.json")
	if got != want {
		t.Errorf("default path: got %q, want %q", got, want)
	}
}

// TestStorageResolvedPathTildeExpansion verifies that a path starting with ~/
// is expanded to an absolute path under the home directory.
func TestStorageResolvedPathTildeExpansion(t *testing.T) {
	input := "~/custom/dir/tokens.json"
	got := resolvedStoragePath(input)

	home, err := os.UserHomeDir()
	if err != nil {
		// Fallback returns the input unchanged when home dir is unavailable.
		if got != input {
			t.Errorf("expected unchanged input %q, got %q", input, got)
		}
		return
	}

	want := filepath.Join(home, "custom", "dir", "tokens.json")
	if got != want {
		t.Errorf("tilde expansion: got %q, want %q", got, want)
	}
}

// TestStorageResolvedPathAbsolute verifies that an absolute path is returned
// unchanged.
func TestStorageResolvedPathAbsolute(t *testing.T) {
	input := "/tmp/schwab/tokens.json"
	got := resolvedStoragePath(input)
	if got != input {
		t.Errorf("absolute path: got %q, want %q", got, input)
	}
}

// TestStorageResolvedPathRelative verifies that a relative path without a
// tilde prefix is returned unchanged.
func TestStorageResolvedPathRelative(t *testing.T) {
	input := "relative/path/tokens.json"
	got := resolvedStoragePath(input)
	if got != input {
		t.Errorf("relative path: got %q, want %q", got, input)
	}
}

// TestStorageFileImplementsTokenStorage is a compile-time assertion that
// *FileTokenStorage satisfies the TokenStorage interface.
func TestStorageFileImplementsTokenStorage(t *testing.T) {
	var _ TokenStorage = (*FileTokenStorage)(nil)
}

// TestStorageMemoryImplementsTokenStorage is a compile-time assertion that the
// in-memory test stub also satisfies the TokenStorage interface.
func TestStorageMemoryImplementsTokenStorage(t *testing.T) {
	var _ TokenStorage = (*memoryStorage)(nil)
}
