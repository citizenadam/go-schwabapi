package schwabdev

import (
	"context"
	"log/slog"
	"testing"
	"time"
)

func TestNextWakeup_AccessTokenExpiringSoon(t *testing.T) {
	tm := &TokenManager{
		accessTokenIssued:   time.Now().Add(-25 * time.Minute),
		accessTokenTimeout:  30 * time.Minute,
		refreshTokenIssued:  time.Now(),
		refreshTokenTimeout: 7 * 24 * time.Hour,
	}

	sleep := nextWakeup(tm)
	// Access token expires in ~5 min, refresh threshold is 61s,
	// so wakeup should be ~3.5 min from now, but clamped to TokenCheckerSleep (30s).
	if sleep < 0 {
		t.Errorf("nextWakeup should be non-negative, got %v", sleep)
	}
	if sleep > 4*time.Minute {
		t.Errorf("nextWakeup should be at most ~4 min when access token expires soon, got %v", sleep)
	}
}

func TestNextWakeup_TokensFarFromExpiry(t *testing.T) {
	tm := &TokenManager{
		accessTokenIssued:   time.Now(),
		accessTokenTimeout:  30 * time.Minute,
		refreshTokenIssued:  time.Now(),
		refreshTokenTimeout: 7 * 24 * time.Hour,
	}

	sleep := nextWakeup(tm)
	// Access token expires in 30 min, refresh threshold is 61s,
	// so wakeup should be ~28.5 min, but clamped to at least TokenCheckerSleep (30s).
	atWakeup := time.Now().Add(30 * time.Minute).Add(-AccessTokenRefreshThreshold)
	expected := time.Until(atWakeup)
	if sleep < TokenCheckerSleep {
		t.Errorf("nextWakeup should be at least TokenCheckerSleep (%v), got %v", TokenCheckerSleep, sleep)
	}
	// Allow some margin for timing
	if sleep > expected+5*time.Second {
		t.Errorf("nextWakeup should be ~%v, got %v", expected, sleep)
	}
}

func TestStartTokenChecker_CancelImmediately(t *testing.T) {
	tm := &TokenManager{
		accessTokenIssued:   time.Now(),
		accessTokenTimeout:  30 * time.Minute,
		refreshTokenIssued:  time.Now(),
		refreshTokenTimeout: 7 * 24 * time.Hour,
	}

	logger := slog.New(slog.NewTextHandler(nil, nil)) // nil writer = discard
	ctx := context.Background()
	stop := StartTokenChecker(ctx, tm, logger)

	// Cancel immediately - should not block or panic
	stop()
}

func TestStartTokenChecker_ContextAlreadyCancelled(t *testing.T) {
	tm := &TokenManager{
		accessTokenIssued:   time.Now(),
		accessTokenTimeout:  30 * time.Minute,
		refreshTokenIssued:  time.Now(),
		refreshTokenTimeout: 7 * 24 * time.Hour,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before starting

	logger := slog.New(slog.NewTextHandler(nil, nil))
	stop := StartTokenChecker(ctx, tm, logger)
	defer stop()

	// Give the goroutine a moment to see the cancelled context
	time.Sleep(100 * time.Millisecond)
	// If we get here without hanging, the test passes
}
