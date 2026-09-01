package schwabdev

import (
	"context"
	"encoding/json"
	"net/http"
	"sync/atomic"
	"testing"
	"time"
)

// lockedMemoryStorage is a memoryStorage that also implements TokenStoreLocker,
// so we can exercise the synchronized refresh lifecycle without filesystem I/O.
type lockedMemoryStorage struct {
	memoryStorage
	lockCount atomic.Int64
}

func (m *lockedMemoryStorage) AcquireRefreshLock(_ context.Context) (func(), error) {
	m.lockCount.Add(1)
	// No-op lock for in-process test; the lifecycle logic is what we exercise.
	return func() {}, nil
}

// TestUpdateAccessToken_AdoptsPeerRefreshedToken verifies that when the store
// holds a token different from (and fresher than) the in-memory one, the
// synchronized lifecycle adopts it and performs ZERO HTTP refresh calls.
func TestUpdateAccessToken_AdoptsPeerRefreshedToken(t *testing.T) {
	var grantCalls atomic.Int64
	handler := func(w http.ResponseWriter, r *http.Request) {
		grantCalls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "never-used",
			"refresh_token": "never-used",
			"expires_in":    1800,
		})
	}

	storage := &lockedMemoryStorage{}
	tm, _ := newTestTokenManager(t, handler, &storage.memoryStorage)
	// Override the storage with the TokenStoreLocker-capable wrapper so the
	// synchronized lifecycle exercises the lock path.
	tm.storage = storage

	// Seed in-memory with an OLD access token.
	old := time.Now().UTC().Add(-29 * time.Minute)
	tm.mu.Lock()
	tm.accessToken = "old-access-token"
	tm.accessTokenIssued = old
	tm.refreshToken = "old-refresh-token"
	tm.refreshTokenIssued = time.Now().UTC().Add(-1 * time.Hour)
	tm.mu.Unlock()

	// Seed the STORE with a NEWER token (as if a peer refreshed it).
	peerIssued := time.Now().UTC().Add(-2 * time.Minute)
	storage.rec = &TokenRecord{
		AccessTokenIssued:  peerIssued,
		RefreshTokenIssued: time.Now().UTC().Add(-1 * time.Hour),
		AccessToken:        "peer-access-token",
		RefreshToken:       "peer-refresh-token",
		ExpiresIn:          1800,
	}

	if err := tm.updateAccessToken(); err != nil {
		t.Fatalf("updateAccessToken: %v", err)
	}

	if got := grantCalls.Load(); got != 0 {
		t.Fatalf("expected 0 HTTP refresh calls (peer token adopted), got %d", got)
	}
	if storage.lockCount.Load() != 1 {
		t.Fatalf("expected lock acquired once, got %d", storage.lockCount.Load())
	}

	tm.mu.RLock()
	adopted := tm.accessToken
	tm.mu.RUnlock()
	if adopted != "peer-access-token" {
		t.Fatalf("accessToken = %q, want adopted %q", adopted, "peer-access-token")
	}
}

// TestUpdateAccessToken_RefreshesWhenStoreStale verifies that when the store
// holds the SAME token as in-memory (no peer refresh), the lifecycle performs
// the HTTP refresh.
func TestUpdateAccessToken_RefreshesWhenStoreStale(t *testing.T) {
	var grantCalls atomic.Int64
	handler := func(w http.ResponseWriter, r *http.Request) {
		grantCalls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "fresh-access-token",
			"refresh_token": "fresh-rotated-refresh-token",
			"expires_in":    1800,
		})
	}

	storage := &lockedMemoryStorage{}
	tm, _ := newTestTokenManager(t, handler, &storage.memoryStorage)
	// Override the storage with the TokenStoreLocker-capable wrapper so the
	// synchronized lifecycle exercises the lock path.
	tm.storage = storage

	now := time.Now().UTC()
	tm.mu.Lock()
	tm.accessToken = "stale-access-token"
	tm.accessTokenIssued = now.Add(-29 * time.Minute)
	tm.refreshToken = "stale-refresh-token"
	tm.refreshTokenIssued = now.Add(-1 * time.Hour)
	tm.mu.Unlock()

	// Store matches in-memory (same access token) => no peer refresh to adopt.
	storage.rec = &TokenRecord{
		AccessTokenIssued:  now.Add(-29 * time.Minute),
		RefreshTokenIssued: now.Add(-1 * time.Hour),
		AccessToken:        "stale-access-token",
		RefreshToken:       "stale-refresh-token",
		ExpiresIn:          1800,
	}

	if err := tm.updateAccessToken(); err != nil {
		t.Fatalf("updateAccessToken: %v", err)
	}
	if got := grantCalls.Load(); got != 1 {
		t.Fatalf("expected 1 HTTP refresh call, got %d", got)
	}
}

// TestUpdateAccessToken_PreservesOmittedRefreshToken verifies that when a
// refresh response omits refresh_token, the existing refresh token and its
// issuance anchor are preserved (not blanked).
func TestUpdateAccessToken_PreservesOmittedRefreshToken(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// NOTE: no "refresh_token" key in the response — Schwab sometimes omits it.
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": "new-access-no-rt",
			"expires_in":   1800,
		})
	}

	storage := &lockedMemoryStorage{}
	tm, _ := newTestTokenManager(t, handler, &storage.memoryStorage)
	// Override the storage with the TokenStoreLocker-capable wrapper so the
	// synchronized lifecycle exercises the lock path.
	tm.storage = storage

	// Seed in-memory with a refresh token and its original acquisition anchor.
	acquisition := time.Now().UTC().Add(-2 * time.Hour)
	tm.mu.Lock()
	tm.accessToken = "old-access"
	tm.accessTokenIssued = time.Now().UTC().Add(-29 * time.Minute)
	tm.refreshToken = "keep-me"
	tm.refreshTokenIssued = acquisition
	tm.mu.Unlock()

	// Store matches in-memory so the lifecycle proceeds to refresh.
	storage.rec = &TokenRecord{
		AccessTokenIssued:  time.Now().UTC().Add(-29 * time.Minute),
		RefreshTokenIssued: acquisition,
		AccessToken:        "old-access",
		RefreshToken:       "keep-me",
		ExpiresIn:          1800,
	}

	if err := tm.updateAccessToken(); err != nil {
		t.Fatalf("updateAccessToken: %v", err)
	}

	tm.mu.RLock()
	rt := tm.refreshToken
	rtIssued := tm.refreshTokenIssued
	at := tm.accessToken
	tm.mu.RUnlock()

	if rt != "keep-me" {
		t.Fatalf("refreshToken = %q, want preserved %q", rt, "keep-me")
	}
	if !rtIssued.Equal(acquisition) {
		t.Fatalf("refreshTokenIssued = %v, want preserved anchor %v", rtIssued, acquisition)
	}
	if at != "new-access-no-rt" {
		t.Fatalf("accessToken = %q, want %q", at, "new-access-no-rt")
	}
}

// TestUpdateAccessToken_ReleaseCalledEvenOnRefreshFailure verifies the lock is
// released when the refresh path errors (deferred release).
func TestUpdateAccessToken_ReleaseCalledEvenOnRefreshFailure(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway) // refresh fails
	}

	storage := &lockedMemoryStorage{}
	tm, _ := newTestTokenManager(t, handler, &storage.memoryStorage)
	// Override the storage with the TokenStoreLocker-capable wrapper so the
	// synchronized lifecycle exercises the lock path.
	tm.storage = storage

	now := time.Now().UTC()
	tm.mu.Lock()
	tm.accessToken = "old-access"
	tm.accessTokenIssued = now.Add(-29 * time.Minute)
	tm.refreshToken = "old-refresh"
	tm.refreshTokenIssued = now.Add(-1 * time.Hour)
	tm.mu.Unlock()

	storage.rec = &TokenRecord{
		AccessTokenIssued:  now.Add(-29 * time.Minute),
		RefreshTokenIssued: now.Add(-1 * time.Hour),
		AccessToken:        "old-access",
		RefreshToken:       "old-refresh",
		ExpiresIn:          1800,
	}

	if err := tm.updateAccessToken(); err == nil {
		t.Fatal("expected error from failed refresh")
	}
	// The deferred release must have run without panicking (no deadlock).
	if storage.lockCount.Load() != 1 {
		t.Fatalf("expected lock acquired once, got %d", storage.lockCount.Load())
	}
}

// TestFileTokenStorage_ImplementsTokenStoreLocker verifies FileTokenStorage
// satisfies the TokenStoreLocker interface.
func TestFileTokenStorage_ImplementsTokenStoreLocker(t *testing.T) {
	var _ TokenStoreLocker = (*FileTokenStorage)(nil)
}

// TestFlockRefreshLock_SerializesAndIsIdempotent verifies the flock-based lock
// blocks a second acquirer until the first releases, and that release is
// idempotent.
func TestFlockRefreshLock_SerializesAndIsIdempotent(t *testing.T) {
	path := t.TempDir() + "/tokens.json"
	lock := newFlockRefreshLock(path)

	ctx := context.Background()
	release1, err := lock.AcquireRefreshLock(ctx)
	if err != nil {
		t.Fatalf("first acquire: %v", err)
	}

	// Second acquire should block; try it in a goroutine with a timeout.
	acquired2 := make(chan struct{})
	release2Ch := make(chan func(), 1)
	go func() {
		r, err := lock.AcquireRefreshLock(ctx)
		if err != nil {
			t.Errorf("second acquire: %v", err)
			close(acquired2)
			return
		}
		release2Ch <- r
		close(acquired2)
	}()

	select {
	case <-acquired2:
		t.Fatal("second acquire succeeded while lock was held")
	case <-time.After(100 * time.Millisecond):
		// Expected: still blocked.
	}

	// Release the first; the second should now acquire.
	release1()
	select {
	case <-acquired2:
	case <-time.After(2 * time.Second):
		t.Fatal("second acquire did not proceed after release")
	}
	release2 := <-release2Ch

	// Idempotent release: calling the same release twice must be safe.
	release1()
	release1()
	release2()
	release2()

	// And the lock is reusable after release.
	release3, err := lock.AcquireRefreshLock(ctx)
	if err != nil {
		t.Fatalf("re-acquire after release: %v", err)
	}
	release3()
}

// TestFlockRefreshLock_ContextCancel verifies a pending acquire returns the
// context error when cancelled.
func TestFlockRefreshLock_ContextCancel(t *testing.T) {
	path := t.TempDir() + "/tokens.json"
	lock := newFlockRefreshLock(path)

	release1, err := lock.AcquireRefreshLock(context.Background())
	if err != nil {
		t.Fatalf("first acquire: %v", err)
	}
	defer release1()

	ctx2, cancel2 := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		_, err := lock.AcquireRefreshLock(ctx2)
		errCh <- err
	}()

	// Wait for the second acquirer to be pending, then cancel its context.
	time.Sleep(50 * time.Millisecond)
	cancel2()

	select {
	case err := <-errCh:
		if err == nil {
			t.Fatal("expected context error from cancelled acquire")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("cancelled acquire did not return")
	}
}
