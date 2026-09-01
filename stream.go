package schwabdev

import (
	"context"
	"fmt"
	"log/slog"
	"maps"
	"math/rand"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

const (
	pingInterval = 20 * time.Second
	pingTimeout  = 10 * time.Second

	// DefaultReadyTimeout is the maximum time WaitReady waits for the
	// WebSocket connection to be established and logged in before returning
	// an error. Configurable via SetReadyTimeout.
	DefaultReadyTimeout = 60 * time.Second

	// loginResponseTimeout bounds how long Start waits for the server's
	// LOGIN response before replaying subscriptions. The server closes the
	// connection if an ADD arrives before LOGIN is processed ("STREAM
	// CONNECTION NOT FOUND - Please login again."), so subscriptions must
	// not be written until the LOGIN round-trip completes.
	loginResponseTimeout = 10 * time.Second

	// readyRecheckInterval is how often WaitReady re-fetches the current
	// ready channel. The channel is replaced on every disconnect, so a
	// waiter that captured a pre-reconnect channel must re-check to notice
	// a new connection's readiness without waiting a full timeout.
	readyRecheckInterval = 250 * time.Millisecond
)

// TokenProvider is any type that can return a fresh, valid access token on
// demand. *schwabdev.TokenManager satisfies this interface directly via its
// AccessToken() method, so no adapter is needed.
type TokenProvider interface {
	AccessToken() (string, error)
}

// InfoSource returns the streamer connection metadata from the Schwab
// userPreference endpoint. The map must contain at minimum:
//
//	"streamerSocketUrl"      string
//	"schwabClientChannel"    string
//	"schwabClientFunctionId" string
//	"schwabClientCustomerId" string
//	"schwabClientCorrelId"   string
type InfoSource func() (map[string]any, error)

// Streamer handles the full WebSocket lifecycle for the Schwab Streamer API.
type Streamer struct {
	tokens    TokenProvider
	infoSrc   InfoSource
	logger    *slog.Logger
	reconnect *ReconnectManager

	mu            sync.RWMutex
	conn          *websocket.Conn
	subscriptions map[string]map[string][]string // service → key → fields
	requestID     atomic.Int64

	// ready is closed once after the connection is established, logged in,
	// and subscriptions replayed. It is re-created (open) on every
	// disconnect so WaitReady callers block during reconnection cycles.
	ready       chan struct{}
	readyOnce   sync.Once
	readyTimeout time.Duration

	// cachedInfo caches the streamer connection metadata for the current
	// connection so send does not trigger an HTTP roundtrip per request.
	// It is invalidated at the start of each reconnect cycle so every new
	// connection fetches fresh session keys.
	cachedInfo map[string]any
	infoMu     sync.RWMutex
}

// NewStreamer initialises the streamer.
//
//   - tokens: provides a fresh access token whenever one is needed (always
//     current, never stale).
//   - infoSrc: fetches streamer connection info from the Schwab API.
func NewStreamer(logger *slog.Logger, tokens TokenProvider, infoSrc InfoSource) *Streamer {
	return &Streamer{
		tokens:        tokens,
		infoSrc:       infoSrc,
		logger:        logger,
		reconnect:     NewReconnectManager(logger),
		subscriptions: make(map[string]map[string][]string),
		ready:         make(chan struct{}),
		readyTimeout:  DefaultReadyTimeout,
	}
}

// SetReadyTimeout sets the maximum time WaitReady waits for the streamer to
// be ready. Must be called before Start, as the value is read without
// synchronization during send.
func (s *Streamer) SetReadyTimeout(timeout time.Duration) {
	if timeout <= 0 {
		timeout = DefaultReadyTimeout
	}
	s.readyTimeout = timeout
}

// Start connects, logs in, replays subscriptions, and then reads messages into
// dataChan until the context is cancelled or an unrecoverable error occurs.
// Transient disconnects are handled automatically with exponential backoff.
func (s *Streamer) Start(ctx context.Context, dataChan chan<- []byte) error {
	return s.reconnect.ReconnectWithBackoff(ctx, func(innerCtx context.Context) error {
		// Invalidate cached streamer info so the new connection fetches
		// fresh session keys (they rotate on reconnect).
		s.invalidateCachedInfo()

		info, err := s.infoSrc()
		if err != nil {
			return fmt.Errorf("get streamer info: %w", err)
		}
		s.storeCachedInfo(info)

		wsURL, ok := info["streamerSocketUrl"].(string)
		if !ok || wsURL == "" {
			return fmt.Errorf("streamerSocketUrl missing or empty")
		}

		c, _, err := websocket.Dial(innerCtx, wsURL, nil)
		if err != nil {
			return fmt.Errorf("websocket dial: %w", err)
		}

		s.mu.Lock()
		s.conn = c
		s.mu.Unlock()

		defer func() {
			s.mu.Lock()
			s.conn = nil
			// Re-arm readiness: WaitReady callers block until the next
			// connection completes login.
			s.ready = make(chan struct{})
			s.readyOnce = sync.Once{}
			s.mu.Unlock()
		}()

		if err := s.login(innerCtx, info); err != nil {
			c.Close(websocket.StatusInternalError, "login failed") //nolint:errcheck // teardown; error not actionable
			return fmt.Errorf("login: %w", err)
		}

		// Wait for the server's LOGIN response before replaying
		// subscriptions. Schwab closes the connection if a subscription
		// request arrives before LOGIN is processed ("STREAM CONNECTION NOT
		// FOUND - Please login again."), which would otherwise put the
		// reconnect loop into a permanent connect-close cycle.
		if err := s.awaitLoginResponse(innerCtx, c); err != nil {
			c.Close(websocket.StatusInternalError, "login rejected") //nolint:errcheck // teardown; error not actionable
			return fmt.Errorf("login: %w", err)
		}

		if err := s.resubscribe(innerCtx, info); err != nil {
			// Non-fatal: log and continue — the read loop may still work.
			s.logger.Error("resubscribe after reconnect failed", "error", err)
		}

		s.reconnect.ResetBackoff()

		// Signal readiness so concurrent senders can start writing requests.
		s.mu.Lock()
		s.readyOnce.Do(func() { close(s.ready) })
		s.mu.Unlock()

		// Run ping loop and read loop concurrently; whichever returns first
		// tears down the connection for the other.
		pingCtx, cancelPing := context.WithCancel(innerCtx)
		defer cancelPing()

		go s.pingLoop(pingCtx, c)

		return s.readLoop(innerCtx, c, dataChan)
	})
}

// Stop gracefully closes the WebSocket connection.
func (s *Streamer) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.conn != nil {
		s.conn.Close(websocket.StatusNormalClosure, "user requested stop") //nolint:errcheck // teardown; error not actionable
		s.conn = nil
	}
}

// WaitReady blocks until the streamer's WebSocket connection is active and
// logged in, or until the timeout/context elapses. It is safe to call
// concurrently; the ready channel is re-created on every disconnect, and
// WaitReady re-fetches the current channel so a waiter never waits a full
// timeout on a channel that was replaced by a reconnect.
func (s *Streamer) WaitReady(ctx context.Context, timeout time.Duration) error {
	waitCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	recheck := time.NewTicker(readyRecheckInterval)
	defer recheck.Stop()

	for {
		s.mu.RLock()
		ready := s.ready
		s.mu.RUnlock()

		select {
		case <-ready:
			return nil
		case <-recheck.C:
			continue // re-fetch: the channel may have been replaced by a reconnect
		case <-waitCtx.Done():
			return fmt.Errorf("streamer not ready: %w", waitCtx.Err())
		}
	}
}

// ── Keepalive ────────────────────────────────────────────────────────────────

// pingLoop sends a Ping frame every pingInterval. If the Pong is not received
// within pingTimeout the connection is forcibly closed so the read loop detects
// the error and triggers a reconnect.
func (s *Streamer) pingLoop(ctx context.Context, c *websocket.Conn) {
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pingCtx, cancel := context.WithTimeout(ctx, pingTimeout)
			if err := c.Ping(pingCtx); err != nil {
				s.logger.Warn("ping failed, closing connection", "error", err)
				c.Close(websocket.StatusGoingAway, "ping timeout") //nolint:errcheck // teardown; error not actionable
				cancel()
				return
			}
			cancel()
		}
	}
}

// ── Read loop ────────────────────────────────────────────────────────────────

func (s *Streamer) readLoop(ctx context.Context, c *websocket.Conn, dataChan chan<- []byte) error {
	for {
		_, msg, err := c.Read(ctx)
		if err != nil {
			return err
		}
		select {
		case dataChan <- msg:
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Channel full — drop the message rather than block the WebSocket
			// read loop. Blocking would stall the WebSocket, causing the server
			// to think the client is dead and disconnect us. Schwab sends
			// incremental updates, so a dropped message is quickly superseded
			// by the next update.
			s.logger.Warn("stream data channel full, dropping message",
				"size", len(msg),
			)
		}
	}
}

// ── Auth & subscription internals ───────────────────────────────────────────

func (s *Streamer) login(ctx context.Context, info map[string]any) error {
	// Always fetch a fresh token at login time so we never send a stale one.
	token, err := s.tokens.AccessToken()
	if err != nil {
		return fmt.Errorf("get access token for login: %w", err)
	}

	params := map[string]any{
		"Authorization":          token,
		"SchwabClientChannel":    info["schwabClientChannel"],
		"SchwabClientFunctionId": info["schwabClientFunctionId"],
	}
	req := s.buildRequest("ADMIN", "LOGIN", params, info)

	s.mu.RLock()
	c := s.conn
	s.mu.RUnlock()

	return wsjson.Write(ctx, c, req)
}

// awaitLoginResponse reads messages until the ADMIN/LOGIN response arrives
// and verifies it succeeded (code 0). Any other outcome — a non-zero code,
// a read failure, or a timeout — is returned as an error so the caller can
// fail loudly instead of proceeding with an unauthenticated connection.
func (s *Streamer) awaitLoginResponse(ctx context.Context, c *websocket.Conn) error {
	readCtx, cancel := context.WithTimeout(ctx, loginResponseTimeout)
	defer cancel()

	for {
		var msg map[string]any
		if err := wsjson.Read(readCtx, c, &msg); err != nil {
			return fmt.Errorf("read login response: %w", err)
		}

		responses, _ := msg["response"].([]any)
		for _, r := range responses {
			resp, ok := r.(map[string]any)
			if !ok {
				continue
			}
			service, _ := resp["service"].(string)
			command, _ := resp["command"].(string)
			if !strings.EqualFold(service, "ADMIN") || !strings.EqualFold(command, "LOGIN") {
				continue
			}

			content, _ := resp["content"].(map[string]any)
			code, _ := content["code"].(float64)
			if code != 0 {
				msg, _ := content["msg"].(string)
				return fmt.Errorf("login rejected: code %v %s", code, msg)
			}
			s.logger.Debug("stream login confirmed", "code", code)
			return nil
		}
	}
}

func (s *Streamer) resubscribe(ctx context.Context, info map[string]any) error {
	s.mu.RLock()
	// Snapshot the subscription map so we don't hold the lock during I/O.
	type subEntry struct {
		service string
		keys    []string
		fields  []string
	}
	var entries []subEntry
	for service, keysMap := range s.subscriptions {
		// Group keys that share an identical field set into a single request.
		fieldGroups := make(map[string][]string) // fieldsCSV → keys
		for key, fields := range keysMap {
			csv := strings.Join(fields, ",")
			fieldGroups[csv] = append(fieldGroups[csv], key)
		}
		for fieldsCSV, keys := range fieldGroups {
			entries = append(entries, subEntry{
				service: service,
				keys:    keys,
				fields:  strings.Split(fieldsCSV, ","),
			})
		}
	}
	c := s.conn
	s.mu.RUnlock()

	for _, e := range entries {
		params := map[string]any{
			"keys":   strings.Join(e.keys, ","),
			"fields": strings.Join(e.fields, ","),
		}
		req := s.buildRequest(e.service, "ADD", params, info)
		if err := wsjson.Write(ctx, c, req); err != nil {
			return err
		}
	}
	return nil
}

// buildRequest is the single place that assembles a Schwab streamer request
// and increments the monotonic requestID. It intentionally does NOT acquire
// s.mu — callers manage locking around the conn reference separately.
func (s *Streamer) buildRequest(service, command string, params map[string]any, info map[string]any) map[string]any {
	id := s.requestID.Add(1)
	return map[string]any{
		"service":                strings.ToUpper(service),
		"command":                strings.ToUpper(command),
		"requestid":              id,
		"SchwabClientCustomerId": info["schwabClientCustomerId"],
		"SchwabClientCorrelId":   info["schwabClientCorrelId"],
		"parameters":             params,
	}
}

// record stores a subscription so it can be replayed after a reconnect.
func (s *Streamer) record(service, command string, keys, fields []string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.subscriptions[service] == nil {
		s.subscriptions[service] = make(map[string][]string)
	}

	switch strings.ToUpper(command) {
	case "ADD", "SUBS":
		for _, k := range keys {
			s.subscriptions[service][k] = fields
		}
	case "UNSUBS":
		for _, k := range keys {
			delete(s.subscriptions[service], k)
		}
	}
}

// invalidateCachedInfo drops the cached streamer connection metadata so the
// next request fetches fresh session keys. Called at the start of each
// reconnect cycle.
func (s *Streamer) invalidateCachedInfo() {
	s.infoMu.Lock()
	s.cachedInfo = nil
	s.infoMu.Unlock()
}

// storeCachedInfo caches the streamer connection metadata for the current
// connection.
func (s *Streamer) storeCachedInfo(info map[string]any) {
	s.infoMu.Lock()
	s.cachedInfo = info
	s.infoMu.Unlock()
}

// getCachedInfo returns the cached streamer connection metadata, or nil if
// no cache exists yet.
func (s *Streamer) getCachedInfo() map[string]any {
	s.infoMu.RLock()
	defer s.infoMu.RUnlock()
	return s.cachedInfo
}

// send records the subscription, waits for the streamer to be ready, and
// writes the request to the WebSocket. It is the shared implementation used
// by every public service method.
func (s *Streamer) send(ctx context.Context, service, command string, keys, fields []string, extra map[string]any) error {
	if len(keys) == 0 {
		return fmt.Errorf("send %s/%s: keys must not be empty", service, command)
	}

	if strings.ToUpper(command) != "LOGOUT" {
		s.record(service, command, keys, fields)
	}

	// Wait for the connection to be established and logged in. This
	// eliminates the race where subscription requests arrive before the
	// connection exists (initial startup or during a reconnection cycle).
	if err := s.WaitReady(ctx, s.readyTimeout); err != nil {
		return err
	}

	// Use cached streamer info instead of an HTTP roundtrip per request.
	info := s.getCachedInfo()
	if info == nil {
		// Cache miss — fetch fresh. This only happens if send is called
		// before Start's first info fetch completes.
		var err error
		info, err = s.infoSrc()
		if err != nil {
			return fmt.Errorf("get streamer info: %w", err)
		}
		s.storeCachedInfo(info)
	}

	params := map[string]any{
		"keys":   strings.Join(keys, ","),
		"fields": strings.Join(fields, ","),
	}
	maps.Copy(params, extra)

	req := s.buildRequest(service, command, params, info)

	s.mu.RLock()
	c := s.conn
	s.mu.RUnlock()

	if c == nil {
		return fmt.Errorf("%s: streamer not connected", service)
	}
	return wsjson.Write(ctx, c, req)
}

// ── Public service methods ───────────────────────────────────────────────────
//
// command is typically "ADD", "SUBS", or "UNSUBS".
// fields are integer indices expressed as strings ("0", "1", …) matching the
// StreamFields map in translate.go.

func (s *Streamer) LevelOneEquities(ctx context.Context, keys, fields []string, command string) error {
	return s.send(ctx, "LEVELONE_EQUITIES", command, keys, fields, nil)
}

// LevelOneOptions streams option quotes.
// Key format: [Underlying(6)|Expiry(6)|C/P(1)|Strike(8)], e.g. "AAPL  230616C00185000"
func (s *Streamer) LevelOneOptions(ctx context.Context, keys, fields []string, command string) error {
	return s.send(ctx, "LEVELONE_OPTIONS", command, keys, fields, nil)
}

func (s *Streamer) LevelOneFutures(ctx context.Context, keys, fields []string, command string) error {
	return s.send(ctx, "LEVELONE_FUTURES", command, keys, fields, nil)
}

func (s *Streamer) LevelOneFuturesOptions(ctx context.Context, keys, fields []string, command string) error {
	return s.send(ctx, "LEVELONE_FUTURES_OPTIONS", command, keys, fields, nil)
}

func (s *Streamer) LevelOneForex(ctx context.Context, keys, fields []string, command string) error {
	return s.send(ctx, "LEVELONE_FOREX", command, keys, fields, nil)
}

func (s *Streamer) NYSEBook(ctx context.Context, keys, fields []string, command string) error {
	return s.send(ctx, "NYSE_BOOK", command, keys, fields, nil)
}

func (s *Streamer) NASDAQBook(ctx context.Context, keys, fields []string, command string) error {
	return s.send(ctx, "NASDAQ_BOOK", command, keys, fields, nil)
}

func (s *Streamer) OptionsBook(ctx context.Context, keys, fields []string, command string) error {
	return s.send(ctx, "OPTIONS_BOOK", command, keys, fields, nil)
}

func (s *Streamer) ChartEquity(ctx context.Context, keys, fields []string, command string) error {
	return s.send(ctx, "CHART_EQUITY", command, keys, fields, nil)
}

func (s *Streamer) ChartFutures(ctx context.Context, keys, fields []string, command string) error {
	return s.send(ctx, "CHART_FUTURES", command, keys, fields, nil)
}

func (s *Streamer) ScreenerEquity(ctx context.Context, keys, fields []string, command string) error {
	return s.send(ctx, "SCREENER_EQUITY", command, keys, fields, nil)
}

func (s *Streamer) ScreenerOption(ctx context.Context, keys, fields []string, command string) error {
	return s.send(ctx, "SCREENER_OPTION", command, keys, fields, nil)
}

// AccountActivity subscribes to account-level activity events.
// No keys or fields are required; the service defines them.
func (s *Streamer) AccountActivity(ctx context.Context, command string) error {
	keys := []string{"Account Activity"}
	fields := []string{"0", "1", "2", "3"}
	return s.send(ctx, "ACCT_ACTIVITY", command, keys, fields, nil)
}

// ── Reconnect manager ────────────────────────────────────────────────────────

// ReconnectManager handles exponential backoff with jitter between reconnect
// attempts.
type ReconnectManager struct {
	mu           sync.Mutex
	logger       *slog.Logger
	baseBackoff  time.Duration
	backoffTime  time.Duration
	maxBackoff   time.Duration
	minUptime    time.Duration
	jitterFactor float64
}

// NewReconnectManager returns a ReconnectManager with sensible defaults.
func NewReconnectManager(logger *slog.Logger) *ReconnectManager {
	return &ReconnectManager{
		logger:       logger,
		baseBackoff:  2 * time.Second,
		backoffTime:  2 * time.Second,
		maxBackoff:   120 * time.Second,
		minUptime:    90 * time.Second,
		jitterFactor: 0.2,
	}
}

// ResetBackoff resets the backoff interval to the base duration.
func (r *ReconnectManager) ResetBackoff() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.backoffTime = r.baseBackoff
}

func (r *ReconnectManager) nextSleep() time.Duration {
	r.mu.Lock()
	defer r.mu.Unlock()

	jitter := float64(r.backoffTime) * r.jitterFactor * (rand.Float64()*2 - 1)
	sleep := r.backoffTime + time.Duration(jitter)

	next := min(time.Duration(float64(r.backoffTime)*2), r.maxBackoff)
	r.backoffTime = next

	return sleep
}

// ReconnectWithBackoff calls connectFunc in a loop, backing off between
// failures. It returns only when the context is cancelled or connectFunc
// returns nil (success without a disconnect).
func (r *ReconnectManager) ReconnectWithBackoff(ctx context.Context, connectFunc func(context.Context) error) error {
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		start := time.Now()
		err := connectFunc(ctx)
		uptime := time.Since(start)

		if err == nil {
			return nil
		}

		if uptime > r.minUptime {
			r.ResetBackoff()
		}

		sleep := r.nextSleep()
		r.logger.Warn("connection lost, reconnecting",
			"error", err,
			"uptime", uptime.Round(time.Second),
			"retry_in", sleep.Round(time.Millisecond),
		)

		select {
		case <-time.After(sleep):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}