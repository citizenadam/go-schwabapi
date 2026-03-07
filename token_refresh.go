package schwabdev

import (
	"context"
	"log/slog"
	"time"
)

// StartTokenChecker launches a background goroutine that proactively refreshes
// tokens before they expire. It returns a cancel function — call it to stop
// the checker cleanly (e.g. in main's defer).
//
// The checker wakes up just before each token's refresh threshold rather than
// polling on a fixed interval, so it is efficient even with a 7-day refresh
// token window.
//
// Usage:
//
//	stopChecker := schwabdev.StartTokenChecker(ctx, tm, logger)
//	defer stopChecker()
func StartTokenChecker(ctx context.Context, tm *TokenManager, logger *slog.Logger) context.CancelFunc {
	ctx, cancel := context.WithCancel(ctx)
	go runTokenChecker(ctx, tm, logger)
	return cancel
}

func runTokenChecker(ctx context.Context, tm *TokenManager, logger *slog.Logger) {
	for {
		sleep := nextWakeup(tm)

		if logger != nil {
			logger.Debug("[Schwabdev] token checker sleeping",
				"wake_in", sleep.Round(time.Second))
		}

		select {
		case <-ctx.Done():
			if logger != nil {
				logger.Debug("[Schwabdev] token checker stopped")
			}
			return
		case <-time.After(sleep):
		}

		updated, err := tm.UpdateTokens(false, false)
		if err != nil {
			if logger != nil {
				logger.Error("[Schwabdev] token checker refresh failed", "error", err)
			}
			// Back off briefly before retrying so we don't spin on a
			// persistent error (e.g. network down).
			select {
			case <-ctx.Done():
				return
			case <-time.After(TokenCheckerSleep):
			}
			continue
		}

		if updated && logger != nil {
			logger.Info("[Schwabdev] token checker refreshed tokens proactively")
		}
	}
}

// nextWakeup returns how long to sleep before the next token refresh is due.
// It targets just before the refresh threshold of whichever token expires
// soonest, with a minimum of TokenCheckerSleep to avoid spinning.
func nextWakeup(tm *TokenManager) time.Duration {
	info := tm.TokenInfo()
	now := time.Now()

	// Wake up when we enter each token's refresh threshold window.
	atWakeup := info.AccessTokenExpiry.Add(-AccessTokenRefreshThreshold)
	rtWakeup := info.RefreshTokenExpiry.Add(-RefreshTokenRefreshThreshold)

	// Pick the sooner of the two.
	wakeAt := atWakeup
	if rtWakeup.Before(wakeAt) {
		wakeAt = rtWakeup
	}

	sleep := wakeAt.Sub(now)
	if sleep < TokenCheckerSleep {
		sleep = TokenCheckerSleep
	}
	return sleep
}
