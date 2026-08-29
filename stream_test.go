package schwabdev

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

func newTestStreamer() *Streamer {
	logger := slog.New(slog.NewTextHandler(nopWriter{}, nil))
	return NewStreamer(logger, staticTokenProvider{}, testInfoSource)
}

// nopWriter discards log output so tests stay quiet.
type nopWriter struct{}

func (nopWriter) Write(p []byte) (int, error) { return len(p), nil }

type staticTokenProvider struct{}

func (staticTokenProvider) AccessToken() (string, error) { return "test-token", nil }

func testInfoSource() (map[string]any, error) {
	return map[string]any{
		"streamerSocketUrl":      "wss://example.test/ws",
		"schwabClientCustomerId": "customer-id",
		"schwabClientCorrelId":   "correl-id",
		"schwabClientChannel":    "N9",
		"schwabClientFunctionId": "APIAPP",
	}, nil
}

func TestNewStreamer_Defaults(t *testing.T) {
	s := newTestStreamer()
	if s.readyTimeout != DefaultReadyTimeout {
		t.Fatalf("readyTimeout = %v, want %v", s.readyTimeout, DefaultReadyTimeout)
	}
	s.mu.RLock()
	select {
	case <-s.ready:
		t.Fatal("ready channel must start open (not ready)")
	default:
	}
	s.mu.RUnlock()
}

func TestSetReadyTimeout(t *testing.T) {
	s := newTestStreamer()
	s.SetReadyTimeout(0)
	if s.readyTimeout != DefaultReadyTimeout {
		t.Fatalf("SetReadyTimeout(0) = %v, want default %v", s.readyTimeout, DefaultReadyTimeout)
	}
	s.SetReadyTimeout(5 * time.Second)
	if s.readyTimeout != 5*time.Second {
		t.Fatalf("SetReadyTimeout(5s) = %v, want 5s", s.readyTimeout)
	}
}

func TestWaitReady_AlreadyReady(t *testing.T) {
	s := newTestStreamer()
	s.mu.Lock()
	s.readyOnce.Do(func() { close(s.ready) })
	s.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.WaitReady(ctx, 2*time.Second); err != nil {
		t.Fatalf("WaitReady on ready streamer returned error: %v", err)
	}
}

func TestWaitReady_BlocksUntilSignaled(t *testing.T) {
	s := newTestStreamer()
	done := make(chan error, 1)
	go func() {
		done <- s.WaitReady(context.Background(), 3*time.Second)
	}()

	// Signal readiness.
	s.mu.Lock()
	s.readyOnce.Do(func() { close(s.ready) })
	s.mu.Unlock()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("WaitReady returned error after ready: %v", err)
		}
	case <-time.After(4 * time.Second):
		t.Fatal("WaitReady did not return after ready was signaled")
	}
}

func TestWaitReady_Timeout(t *testing.T) {
	s := newTestStreamer()
	err := s.WaitReady(context.Background(), 200*time.Millisecond)
	if err == nil {
		t.Fatal("WaitReady must error when the streamer never becomes ready")
	}
	if !strings.Contains(err.Error(), "streamer not ready") {
		t.Fatalf("error %q must contain %q", err, "streamer not ready")
	}
	if !strings.Contains(err.Error(), "context deadline exceeded") {
		t.Fatalf("error %q must report the timeout cause", err)
	}
}

// TestWaitReady_DetectsReplacedChannel covers the reconnect case: a waiter
// that captured the pre-reconnect ready channel must notice a new connection's
// readiness via the recheck loop instead of waiting a full timeout on a
// channel that will never close.
func TestWaitReady_DetectsReplacedChannel(t *testing.T) {
	s := newTestStreamer()

	done := make(chan error, 1)
	go func() {
		done <- s.WaitReady(context.Background(), 3*time.Second)
	}()

	// Simulate a reconnect completing: replace the ready channel with a new
	// one (as Start's defer does on disconnect) and close it.
	s.mu.Lock()
	s.ready = make(chan struct{})
	s.readyOnce = sync.Once{}
	close(s.ready)
	s.mu.Unlock()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("WaitReady must return nil once the replaced channel is closed, got: %v", err)
		}
	case <-time.After(4 * time.Second):
		t.Fatal("WaitReady waited a full timeout on a stale channel")
	}
}

// TestSend_WaitsForReadyBeforeWriting verifies the request gating order:
// record → WaitReady → info → connection check. A call issued before
// readiness must block until ready, then fail at the nil-conn check rather
// than being rejected for not being ready.
func TestSend_WaitsForReadyBeforeWriting(t *testing.T) {
	s := newTestStreamer()
	infoCalls := 0
	s.infoSrc = func() (map[string]any, error) {
		infoCalls++
		return testInfoSource()
	}

	done := make(chan error, 1)
	go func() {
		done <- s.LevelOneEquities(context.Background(), []string{"AAPL"}, []string{"0", "1"}, "ADD")
	}()

	// Give the goroutine a moment to reach WaitReady, then signal readiness.
	time.Sleep(50 * time.Millisecond)
	s.mu.Lock()
	s.readyOnce.Do(func() { close(s.ready) })
	s.mu.Unlock()

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("expected error (no connection), got nil")
		}
		if !strings.Contains(err.Error(), "streamer not connected") {
			t.Fatalf("error %q must report the nil connection, not a readiness failure", err)
		}
	case <-time.After(4 * time.Second):
		t.Fatal("send did not proceed after readiness was signaled")
	}

	if infoCalls != 1 {
		t.Fatalf("infoSrc called %d times, want 1 (cache miss on first request)", infoCalls)
	}

	// The subscription must have been recorded for replay on reconnect.
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.subscriptions["LEVELONE_EQUITIES"]["AAPL"] == nil {
		t.Fatal("subscription was not recorded")
	}
}

// TestSend_UsesCachedInfo verifies no HTTP roundtrip per request: once info
// is cached, send must not call infoSrc.
func TestSend_UsesCachedInfo(t *testing.T) {
	s := newTestStreamer()
	infoCalls := 0
	s.infoSrc = func() (map[string]any, error) {
		infoCalls++
		return testInfoSource()
	}
	info, _ := testInfoSource()
	s.storeCachedInfo(info)

	s.mu.Lock()
	s.readyOnce.Do(func() { close(s.ready) })
	s.mu.Unlock()

	err := s.LevelOneEquities(context.Background(), []string{"AAPL"}, []string{"0", "1"}, "ADD")
	if err == nil || !strings.Contains(err.Error(), "streamer not connected") {
		t.Fatalf("expected nil-conn error, got %v", err)
	}
	if infoCalls != 0 {
		t.Fatalf("infoSrc called %d times with cached info, want 0", infoCalls)
	}
}

// wsTestServer spins up a local WebSocket server that behaves like the Schwab
// streamer for the parts Start() exercises: it accepts the connection, reads
// the LOGIN request, replies with code 0, then serves subscriptions. When
// sendData is true it pushes a LEVELONE_EQUITIES data message after the first
// subscription request.
func wsTestServer(t *testing.T, sendData bool) (*httptest.Server, string, <-chan struct{}, <-chan struct{}) {
	t.Helper()
	loginReceived := make(chan struct{}, 1)
	subReceived := make(chan struct{}, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		defer c.Close(websocket.StatusNormalClosure, "test done")

		// Read the LOGIN request.
		var login map[string]any
		if err := wsjson.Read(context.Background(), c, &login); err != nil {
			return
		}
		loginReceived <- struct{}{}
		if err := wsjson.Write(context.Background(), c, map[string]any{
			"response": []any{map[string]any{
				"service": "ADMIN", "command": "LOGIN", "content": map[string]any{"code": 0},
			}},
		}); err != nil {
			return
		}

		// Serve subscription requests until the client disconnects.
		for {
			var req map[string]any
			if err := wsjson.Read(context.Background(), c, &req); err != nil {
				return
			}
			select {
			case subReceived <- struct{}{}:
			default:
			}
			if sendData {
				_ = wsjson.Write(context.Background(), c, map[string]any{
					"data": []any{map[string]any{
						"service": "LEVELONE_EQUITIES",
						"content": []any{map[string]any{"0": "AAPL", "1": 100.5}},
					}},
				})
				sendData = false
			}
		}
	}))
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	return srv, wsURL, loginReceived, subReceived
}

// TestStart_FullLifecycle drives Start() against a local WS server: dial,
// login, readiness, a post-ready subscription, and a data message flowing
// into dataChan.
func TestStart_FullLifecycle(t *testing.T) {
	srv, wsURL, loginReceived, subReceived := wsTestServer(t, true)
	defer srv.Close()

	s := newTestStreamer()
	s.infoSrc = func() (map[string]any, error) {
		return map[string]any{
			"streamerSocketUrl":      wsURL,
			"schwabClientCustomerId": "customer-id",
			"schwabClientCorrelId":   "correl-id",
			"schwabClientChannel":    "N9",
			"schwabClientFunctionId": "APIAPP",
		}, nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	dataChan := make(chan []byte, 16)

	started := make(chan error, 1)
	go func() { started <- s.Start(ctx, dataChan) }()

	// The streamer must become ready (dial + login complete).
	if err := s.WaitReady(ctx, 5*time.Second); err != nil {
		t.Fatalf("streamer never became ready: %v", err)
	}
	select {
	case <-loginReceived:
	case <-time.After(3 * time.Second):
		t.Fatal("server never received the LOGIN request")
	}

	// Subscribe post-ready; the server reads the ADD and pushes data back.
	if err := s.LevelOneEquities(ctx, []string{"AAPL"}, []string{"0", "1"}, "ADD"); err != nil {
		t.Fatalf("LevelOneEquities failed: %v", err)
	}
	select {
	case <-subReceived:
	case <-time.After(3 * time.Second):
		t.Fatal("server never received the ADD request")
	}

	// The data message must arrive on dataChan. Responses (LOGIN, ADD acks)
	// also flow through dataChan, so scan until the data message shows up.
	deadline := time.After(3 * time.Second)
	for {
		select {
		case msg := <-dataChan:
			if strings.Contains(string(msg), "AAPL") {
				goto dataReceived
			}
		case <-deadline:
			t.Fatal("no data message received on dataChan")
		}
	}
dataReceived:

	cancel()
	select {
	case <-started:
	case <-time.After(3 * time.Second):
		t.Fatal("Start did not return after cancel")
	}
}

// TestAwaitLoginResponse_Rejected verifies a non-zero LOGIN response code
// fails loudly instead of proceeding with an unauthenticated connection.
func TestAwaitLoginResponse_Rejected(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		defer c.Close(websocket.StatusNormalClosure, "test done")
		var login map[string]any
		if err := wsjson.Read(context.Background(), c, &login); err != nil {
			return
		}
		_ = wsjson.Write(context.Background(), c, map[string]any{
			"response": []any{map[string]any{
				"service": "ADMIN", "command": "LOGIN",
				"content": map[string]any{"code": 12, "msg": "invalid credentials"},
			}},
		})
	}))
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	c, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer c.Close(websocket.StatusNormalClosure, "done")

	s := newTestStreamer()

	// Send a LOGIN request so the server responds.
	login := map[string]any{
		"service": "ADMIN", "command": "LOGIN", "requestid": 1,
		"SchwabClientCustomerId": "customer-id",
		"SchwabClientCorrelId":   "correl-id",
		"parameters": map[string]any{
			"Authorization":          "test-token",
			"SchwabClientChannel":    "N9",
			"SchwabClientFunctionId": "APIAPP",
		},
	}
	if err := wsjson.Write(ctx, c, login); err != nil {
		t.Fatalf("write login failed: %v", err)
	}

	err = s.awaitLoginResponse(ctx, c)
	if err == nil {
		t.Fatal("awaitLoginResponse must error on a rejected login")
	}
	if !strings.Contains(err.Error(), "login rejected") {
		t.Fatalf("error %q must mention the login rejection", err)
	}
}

// TestStart_ReplaysRecordedSubscriptions verifies subscriptions recorded
// before the connection is established are replayed after login.
func TestStart_ReplaysRecordedSubscriptions(t *testing.T) {
	srv, wsURL, _, subReceived := wsTestServer(t, false)
	defer srv.Close()

	s := newTestStreamer()
	s.infoSrc = func() (map[string]any, error) {
		return map[string]any{
			"streamerSocketUrl":      wsURL,
			"schwabClientCustomerId": "customer-id",
			"schwabClientCorrelId":   "correl-id",
			"schwabClientChannel":    "N9",
			"schwabClientFunctionId": "APIAPP",
		}, nil
	}

	// Record a subscription BEFORE the connection exists, as the application
	// does when a symbol is queued before the streamer dials.
	s.record("LEVELONE_EQUITIES", "ADD", []string{"AAPL"}, []string{"0", "1"})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	dataChan := make(chan []byte, 16)
	started := make(chan error, 1)
	go func() { started <- s.Start(ctx, dataChan) }()

	if err := s.WaitReady(ctx, 5*time.Second); err != nil {
		t.Fatalf("streamer never became ready: %v", err)
	}

	select {
	case <-subReceived:
		// The recorded subscription was replayed after login.
	case <-time.After(3 * time.Second):
		t.Fatal("recorded subscription was not replayed after login")
	}

	cancel()
	select {
	case <-started:
	case <-time.After(3 * time.Second):
		t.Fatal("Start did not return after cancel")
	}
}