//go:build unix

package schwabdev

import (
	"context"
	"os"
	"sync"
	"syscall"
	"time"
)

// flockLockInterval is how often a pending lock attempt re-polls a
// non-blocking flock while waiting for the lock to become available.
const flockLockInterval = 50 * time.Millisecond

// flockRefreshLock is a TokenStoreLocker implementation for FileTokenStorage
// that serializes refresh attempts across processes sharing the same token
// file. It locks a sibling "<path>.lock" file via POSIX flock.
type flockRefreshLock struct {
	path string

	mu   sync.Mutex
	file *os.File
	held bool
}

// newFlockRefreshLock returns a flockRefreshLock targeting the given token
// file path (the sibling ".lock" file is derived from it).
func newFlockRefreshLock(path string) *flockRefreshLock {
	return &flockRefreshLock{path: path}
}

// AcquireRefreshLock implements TokenStoreLocker. It opens (creating if
// necessary) the sibling "<path>.lock" file and acquires an exclusive,
// non-blocking flock, polling until the lock is available or ctx is done.
//
// The returned release function is idempotent: subsequent calls are no-ops,
// so it is safe to defer once and also call from a context-cancel goroutine.
func (f *flockRefreshLock) AcquireRefreshLock(ctx context.Context) (func(), error) {
	file, err := os.OpenFile(f.path+".lock", os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}

	// Fast path: try non-blocking once before allocating the ticker.
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err == nil {
		f.mu.Lock()
		f.file = file
		f.held = true
		f.mu.Unlock()
		return f.release, nil
	}

	// Slow path: poll until the lock is available or ctx is done.
	ticker := time.NewTicker(flockLockInterval)
	defer ticker.Stop()

	for {
		if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err == nil {
			f.mu.Lock()
			f.file = file
			f.held = true
			f.mu.Unlock()
			return f.release, nil
		}
		select {
		case <-ctx.Done():
			_ = file.Close()
			return nil, ctx.Err()
		case <-ticker.C:
		}
	}
}

// release unlocks and closes the lock file exactly once. It is safe to call
// multiple times from multiple goroutines.
func (f *flockRefreshLock) release() {
	f.mu.Lock()
	defer f.mu.Unlock()
	if !f.held || f.file == nil {
		return
	}
	f.held = false
	file := f.file
	f.file = nil
	_ = syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
	_ = file.Close()
}

// Compile-time assertion that FileTokenStorage satisfies TokenStoreLocker
// through its embedded flockRefreshLock.
var _ TokenStoreLocker = (*FileTokenStorage)(nil)
