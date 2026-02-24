package stream

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"time"
)

// ReconnectManager handles auto-reconnect logic with exponential backoff.
// Based on Python's StreamBase._run_streamer() method.
type ReconnectManager struct {
	logger      *slog.Logger
	backoffTime time.Duration // Current backoff time (starts at 2s)
	maxBackoff  time.Duration // Maximum backoff (120s)
	minUptime   time.Duration // Minimum uptime before stable (90s)
}

// NewReconnectManager creates a new reconnect manager with exponential backoff.
func NewReconnectManager(logger *slog.Logger) *ReconnectManager {
	return &ReconnectManager{
		logger:      logger,
		backoffTime: 2 * time.Second,   // Initial backoff: 2 seconds
		maxBackoff:  120 * time.Second, // Maximum backoff: 120 seconds
		minUptime:   90 * time.Second,  // Minimum uptime: 90 seconds
	}
}

// ShouldReconnect determines if a reconnection attempt should be made based on uptime.
// Returns false if connection crashed within minimum uptime (likely invalid login or no subscriptions).
func (r *ReconnectManager) ShouldReconnect(uptime time.Duration, err error) bool {
	if uptime < r.minUptime {
		r.logger.Error("Stream crashed within minimum uptime, not restarting",
			"uptime", uptime,
			"minUptime", r.minUptime,
			"error", err)
		return false
	}
	return true
}

// WaitForBackoff waits for the current backoff time with context cancellation support.
// Doubles backoff time for next attempt (capped at maxBackoff).
func (r *ReconnectManager) WaitForBackoff(ctx context.Context) error {
	r.logger.Info("Waiting before reconnect attempt",
		"backoffTime", r.backoffTime)

	select {
	case <-time.After(r.backoffTime):
		// Double backoff time for next attempt (capped at maxBackoff)
		r.backoffTime = time.Duration(math.Min(
			float64(r.backoffTime*2),
			float64(r.maxBackoff),
		))
		return nil
	case <-ctx.Done():
		r.logger.Debug("Backoff wait cancelled by context")
		return ctx.Err()
	}
}

// ResetBackoff resets the backoff time to initial value (2 seconds).
// Should be called on successful connection.
func (r *ReconnectManager) ResetBackoff() {
	r.backoffTime = 2 * time.Second
	r.logger.Debug("Backoff time reset to initial value")
}

// GetBackoffTime returns the current backoff time.
func (r *ReconnectManager) GetBackoffTime() time.Duration {
	return r.backoffTime
}

// GetMinUptime returns the minimum uptime required before considering connection stable.
func (r *ReconnectManager) GetMinUptime() time.Duration {
	return r.minUptime
}

// LogReconnectAttempt logs a reconnection attempt with backoff time.
func (r *ReconnectManager) LogReconnectAttempt(attempt int) {
	r.logger.Info("Stream connection lost, reconnecting...",
		"attempt", attempt,
		"backoffTime", r.backoffTime)
}

// LogConnectionSuccess logs a successful connection.
func (r *ReconnectManager) LogConnectionSuccess() {
	r.logger.Info("Stream connection established successfully")
}

// LogConnectionError logs a connection error.
func (r *ReconnectManager) LogConnectionError(err error) {
	r.logger.Error("Stream connection error",
		"error", err)
}

// CalculateUptime calculates uptime from start time to now.
func CalculateUptime(startTime time.Time) time.Duration {
	return time.Since(startTime)
}

// WithBackoffContext creates a child context with a deadline based on backoff time.
// Useful for preventing indefinite blocking during reconnection attempts.
func (r *ReconnectManager) WithBackoffContext(ctx context.Context) (context.Context, context.CancelFunc) {
	// Use a reasonable timeout for reconnection attempts (e.g., 5 minutes)
	timeout := 5 * time.Minute
	return context.WithTimeout(ctx, timeout)
}

// ReconnectWithBackoff executes a reconnection attempt with exponential backoff.
// Returns error if reconnection fails or context is cancelled.
func (r *ReconnectManager) ReconnectWithBackoff(
	ctx context.Context,
	connectFunc func(context.Context) error,
) error {
	attempt := 0

	for {
		attempt++
		r.LogReconnectAttempt(attempt)

		// Check if context is cancelled before attempting connection
		if ctx.Err() != nil {
			return fmt.Errorf("reconnection cancelled: %w", ctx.Err())
		}

		// Attempt connection
		startTime := time.Now()
		err := connectFunc(ctx)
		if err != nil {
			uptime := CalculateUptime(startTime)

			// Check if we should continue reconnecting
			if !r.ShouldReconnect(uptime, err) {
				return fmt.Errorf("connection failed too quickly: %w", err)
			}

			// Wait for backoff before next attempt
			if waitErr := r.WaitForBackoff(ctx); waitErr != nil {
				return fmt.Errorf("backoff wait failed: %w", waitErr)
			}

			continue
		}

		// Connection successful
		r.ResetBackoff()
		r.LogConnectionSuccess()
		return nil
	}
}
