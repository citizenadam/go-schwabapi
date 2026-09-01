//go:build !unix

package schwabdev

import (
	"context"
)

// flockRefreshLock is a no-op TokenStoreLocker on non-Unix platforms. It
// exists so FileTokenStorage satisfies the TokenStoreLocker interface on every
// platform while providing no cross-process coordination where flock is
// unavailable. In-process singleflight still dedups concurrent refreshes.
type flockRefreshLock struct{}

// newFlockRefreshLock returns a no-op flockRefreshLock; the path is unused.
func newFlockRefreshLock(_ string) *flockRefreshLock { return &flockRefreshLock{} }

// AcquireRefreshLock returns immediately with a no-op release.
func (f *flockRefreshLock) AcquireRefreshLock(_ context.Context) (func(), error) {
	return func() {}, nil
}
