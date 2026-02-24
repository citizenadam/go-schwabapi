package stream

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/coder/websocket"
	"log/slog"
)

// Client represents a streaming client with reconnection logic
type Client struct {
	conn        *Conn
	logger      *slog.Logger
	backoffTime time.Duration
	maxBackoff  time.Duration
	minUptime   time.Duration // Minimum uptime before considering connection stable
	active      bool
	shouldStop  bool
}

// NewClient creates a new streaming client
func NewClient(logger *slog.Logger) *Client {
	return &Client{
		conn:        NewConn(),
		logger:      logger,
		backoffTime: 2 * time.Second,
		maxBackoff:  120 * time.Second,
		minUptime:   90 * time.Second,
	}
}

// Connect establishes a WebSocket connection with reconnection logic
func (c *Client) Connect(ctx context.Context, url string) error {
	c.shouldStop = false

	for !c.shouldStop {
		// Check if context is cancelled
		if ctx.Err() != nil {
			return ctx.Err()
		}

		startTime := time.Now()

		// Attempt connection
		err := c.connectAttempt(ctx, url)
		if err != nil {
			// Check if we should stop
			if c.shouldStop {
				return fmt.Errorf("stream stopped: %w", err)
			}

			// Calculate uptime
			uptime := time.Since(startTime)

			// If connection crashed within minUptime, don't restart
			if uptime < c.minUptime {
				c.logger.Error("Stream crashed within minimum uptime, not restarting",
					"uptime", uptime,
					"error", err)
				return fmt.Errorf("stream crashed too quickly: %w", err)
			}

			// Log reconnection attempt
			c.logger.Warn("Stream connection lost, reconnecting",
				"backoff", c.backoffTime,
				"error", err)

			// Wait for backoff period
			if err := c.waitForBackoff(ctx); err != nil {
				return err
			}

			// Double backoff time for next attempt (capped at maxBackoff)
			c.backoffTime = time.Duration(math.Min(
				float64(c.backoffTime*2),
				float64(c.maxBackoff),
			))

			continue
		}

		// Connection successful, reset backoff time
		c.backoffTime = 2 * time.Second
		c.logger.Info("Stream connected successfully")
	}

	return nil
}

// connectAttempt performs a single connection attempt
func (c *Client) connectAttempt(ctx context.Context, url string) error {
	c.logger.Debug("Attempting to connect to streaming server", "url", url)

	// Create WebSocket connection
	wsConn, _, err := websocket.Dial(ctx, url, nil)
	if err != nil {
		return fmt.Errorf("dial failed: %w", err)
	}

	// Set the underlying connection
	c.conn.conn = wsConn
	c.active = true

	// Start read/write loops
	go c.conn.readLoop()
	go c.conn.writeLoop()

	return nil
}

// waitForBackoff waits for the backoff period with context cancellation support
func (c *Client) waitForBackoff(ctx context.Context) error {
	select {
	case <-time.After(c.backoffTime):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Stop stops the streaming client
func (c *Client) Stop() {
	c.shouldStop = true
	c.active = false

	if c.conn != nil {
		_ = c.conn.Close()
	}
}

// IsActive returns whether the client is currently active
func (c *Client) IsActive() bool {
	return c.active
}

// Read returns a channel for incoming messages
func (c *Client) Read() <-chan []byte {
	return c.conn.Read()
}

// Write sends a message through the WebSocket
func (c *Client) Write(data []byte) error {
	if !c.active {
		return errors.New("client is not active")
	}
	return c.conn.Write(data)
}

// Close closes the WebSocket connection
func (c *Client) Close() error {
	c.Stop()
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
