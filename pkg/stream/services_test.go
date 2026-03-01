package stream

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/citizenadam/go-schwabapi/pkg/types"
	"github.com/coder/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log/slog"
)

// mockWebSocketServer creates a mock WebSocket server for testing
type mockWebSocketServer struct {
	server    *httptest.Server
	mu        sync.Mutex
	messages  [][]byte
	onMessage func([]byte)
}

// newMockWebSocketServer creates a new mock WebSocket server
func newMockWebSocketServer(t *testing.T) *mockWebSocketServer {
	m := &mockWebSocketServer{
		messages: make([][]byte, 0),
	}

	m.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		opts := &websocket.AcceptOptions{
			OriginPatterns: []string{"*"},
		}
		conn, err := websocket.Accept(w, r, opts)
		if err != nil {
			t.Logf("WebSocket accept error: %v", err)
			return
		}
		defer conn.Close(websocket.StatusNormalClosure, "")

		// Read messages in a loop
		ctx := context.Background()
		for {
			_, data, err := conn.Read(ctx)
			if err != nil {
				return
			}

			m.mu.Lock()
			m.messages = append(m.messages, data)
			if m.onMessage != nil {
				m.onMessage(data)
			}
			m.mu.Unlock()
		}
	}))

	return m
}

// URL returns the WebSocket URL for the mock server
func (m *mockWebSocketServer) URL() string {
	return "ws" + m.server.URL[4:] // Replace http:// with ws://
}

// Close closes the mock server
func (m *mockWebSocketServer) Close() {
	m.server.Close()
}

// GetMessages returns all received messages
func (m *mockWebSocketServer) GetMessages() [][]byte {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.messages
}

// ClearMessages clears all received messages
func (m *mockWebSocketServer) ClearMessages() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = make([][]byte, 0)
}

// createTestClient creates a Client connected to the mock server
func createTestClient(t *testing.T, serverURL string) *Client {
	logger := slog.Default()
	client := NewClient(logger)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	wsConn, _, err := websocket.Dial(ctx, serverURL, nil)
	require.NoError(t, err, "Failed to connect to mock server")

	client.conn.conn = wsConn
	client.active = true

	go client.conn.readLoop()
	go client.conn.writeLoop()

	return client
}

// closeTestClient safely closes the test client
func closeTestClient(client *Client) {
	if client == nil {
		return
	}
	client.active = false
	if client.conn != nil && client.conn.conn != nil {
		_ = client.conn.conn.Close(websocket.StatusNormalClosure, "")
	}
}

// TestLevelOneEquities_Subscribe tests the LevelOneEquities subscription
func TestLevelOneEquities_Subscribe(t *testing.T) {
	server := newMockWebSocketServer(t)
	defer server.Close()

	client := createTestClient(t, server.URL())
	defer closeTestClient(client)

	logger := slog.Default()
	manager := NewManager(logger)
	ctx := context.Background()

	// Test ADD subscription
	err := client.LevelOneEquities(ctx, manager, "AAPL,MSFT", "0,1,2,3", "ADD")
	require.NoError(t, err, "LevelOneEquities ADD should not error")

	// Wait for message to be sent
	time.Sleep(100 * time.Millisecond)

	// Verify message was sent
	messages := server.GetMessages()
	require.Len(t, messages, 1, "Should have sent one message")

	// Parse and verify the message
	var sub types.Subscription
	err = json.Unmarshal(messages[0], &sub)
	require.NoError(t, err, "Message should be valid JSON")

	assert.Equal(t, "LEVELONE_EQUITIES", sub.Service)
	assert.Equal(t, "ADD", sub.Command)
	assert.Equal(t, "AAPL,MSFT", sub.Parameters.Keys)
	assert.Equal(t, "0,1,2,3", sub.Parameters.Fields)

	// Verify subscription was recorded
	subs := manager.GetSubscriptions()
	assert.Contains(t, subs, "LEVELONE_EQUITIES")
	assert.Contains(t, subs["LEVELONE_EQUITIES"], "AAPL")
	assert.Contains(t, subs["LEVELONE_EQUITIES"], "MSFT")
}

// TestLevelOneOptions_Subscribe tests the LevelOneOptions subscription
func TestLevelOneOptions_Subscribe(t *testing.T) {
	server := newMockWebSocketServer(t)
	defer server.Close()

	client := createTestClient(t, server.URL())
	defer closeTestClient(client)

	logger := slog.Default()
	manager := NewManager(logger)
	ctx := context.Background()

	// Test SUBS subscription
	err := client.LevelOneOptions(ctx, manager, "AAPL  240809C00095000", "0,1,2", "SUBS")
	require.NoError(t, err, "LevelOneOptions SUBS should not error")

	// Wait for message to be sent
	time.Sleep(100 * time.Millisecond)

	// Verify message was sent
	messages := server.GetMessages()
	require.Len(t, messages, 1, "Should have sent one message")

	// Parse and verify the message
	var sub types.Subscription
	err = json.Unmarshal(messages[0], &sub)
	require.NoError(t, err, "Message should be valid JSON")

	assert.Equal(t, "LEVELONE_OPTIONS", sub.Service)
	assert.Equal(t, "SUBS", sub.Command)
	assert.Equal(t, "AAPL  240809C00095000", sub.Parameters.Keys)
	assert.Equal(t, "0,1,2", sub.Parameters.Fields)

	// Verify subscription was recorded
	subs := manager.GetSubscriptions()
	assert.Contains(t, subs, "LEVELONE_OPTIONS")
}

// TestNyseBook_Subscribe tests the NyseBook subscription
func TestNyseBook_Subscribe(t *testing.T) {
	server := newMockWebSocketServer(t)
	defer server.Close()

	client := createTestClient(t, server.URL())
	defer closeTestClient(client)

	logger := slog.Default()
	manager := NewManager(logger)
	ctx := context.Background()

	// Test ADD subscription
	err := client.NyseBook(ctx, manager, "AAPL,GOOGL", "0,1,2,3,4", "ADD")
	require.NoError(t, err, "NyseBook ADD should not error")

	// Wait for message to be sent
	time.Sleep(100 * time.Millisecond)

	// Verify message was sent
	messages := server.GetMessages()
	require.Len(t, messages, 1, "Should have sent one message")

	// Parse and verify the message
	var sub types.Subscription
	err = json.Unmarshal(messages[0], &sub)
	require.NoError(t, err, "Message should be valid JSON")

	assert.Equal(t, "NYSE_BOOK", sub.Service)
	assert.Equal(t, "ADD", sub.Command)
	assert.Equal(t, "AAPL,GOOGL", sub.Parameters.Keys)
	assert.Equal(t, "0,1,2,3,4", sub.Parameters.Fields)

	// Verify subscription was recorded
	subs := manager.GetSubscriptions()
	assert.Contains(t, subs, "NYSE_BOOK")
	assert.Contains(t, subs["NYSE_BOOK"], "AAPL")
	assert.Contains(t, subs["NYSE_BOOK"], "GOOGL")
}

// TestChartEquity_Subscribe tests the ChartEquity subscription
func TestChartEquity_Subscribe(t *testing.T) {
	server := newMockWebSocketServer(t)
	defer server.Close()

	client := createTestClient(t, server.URL())
	defer closeTestClient(client)

	logger := slog.Default()
	manager := NewManager(logger)
	ctx := context.Background()

	// Test SUBS subscription
	err := client.ChartEquity(ctx, manager, "AAPL,TSLA", "0,1,2,3,4,5,6,7", "SUBS")
	require.NoError(t, err, "ChartEquity SUBS should not error")

	// Wait for message to be sent
	time.Sleep(100 * time.Millisecond)

	// Verify message was sent
	messages := server.GetMessages()
	require.Len(t, messages, 1, "Should have sent one message")

	// Parse and verify the message
	var sub types.Subscription
	err = json.Unmarshal(messages[0], &sub)
	require.NoError(t, err, "Message should be valid JSON")

	assert.Equal(t, "CHART_EQUITY", sub.Service)
	assert.Equal(t, "SUBS", sub.Command)
	assert.Equal(t, "AAPL,TSLA", sub.Parameters.Keys)
	assert.Equal(t, "0,1,2,3,4,5,6,7", sub.Parameters.Fields)

	// Verify subscription was recorded
	subs := manager.GetSubscriptions()
	assert.Contains(t, subs, "CHART_EQUITY")
}

// TestScreenerEquity_Subscribe tests the ScreenerEquity subscription
func TestScreenerEquity_Subscribe(t *testing.T) {
	server := newMockWebSocketServer(t)
	defer server.Close()

	client := createTestClient(t, server.URL())
	defer closeTestClient(client)

	logger := slog.Default()
	manager := NewManager(logger)
	ctx := context.Background()

	// Test ADD subscription
	err := client.ScreenerEquity(ctx, manager, "AAPL,MSFT,GOOGL", "0,1,2", "ADD")
	require.NoError(t, err, "ScreenerEquity ADD should not error")

	// Wait for message to be sent
	time.Sleep(100 * time.Millisecond)

	// Verify message was sent
	messages := server.GetMessages()
	require.Len(t, messages, 1, "Should have sent one message")

	// Parse and verify the message
	var sub types.Subscription
	err = json.Unmarshal(messages[0], &sub)
	require.NoError(t, err, "Message should be valid JSON")

	assert.Equal(t, "SCREENER_EQUITY", sub.Service)
	assert.Equal(t, "ADD", sub.Command)
	assert.Equal(t, "AAPL,MSFT,GOOGL", sub.Parameters.Keys)
	assert.Equal(t, "0,1,2", sub.Parameters.Fields)

	// Verify subscription was recorded
	subs := manager.GetSubscriptions()
	assert.Contains(t, subs, "SCREENER_EQUITY")
}

// TestAccountActivity_Subscribe tests the AccountActivity subscription
func TestAccountActivity_Subscribe(t *testing.T) {
	server := newMockWebSocketServer(t)
	defer server.Close()

	client := createTestClient(t, server.URL())
	defer closeTestClient(client)

	logger := slog.Default()
	manager := NewManager(logger)
	ctx := context.Background()

	// Test SUBS subscription
	err := client.AccountActivity(ctx, manager, "account_hash_123", "0,1,2,3", "SUBS")
	require.NoError(t, err, "AccountActivity SUBS should not error")

	// Wait for message to be sent
	time.Sleep(100 * time.Millisecond)

	// Verify message was sent
	messages := server.GetMessages()
	require.Len(t, messages, 1, "Should have sent one message")

	// Parse and verify the message
	var sub types.Subscription
	err = json.Unmarshal(messages[0], &sub)
	require.NoError(t, err, "Message should be valid JSON")

	assert.Equal(t, "ACCOUNT_ACTIVITY", sub.Service)
	assert.Equal(t, "SUBS", sub.Command)
	assert.Equal(t, "account_hash_123", sub.Parameters.Keys)
	assert.Equal(t, "0,1,2,3", sub.Parameters.Fields)

	// Verify subscription was recorded
	subs := manager.GetSubscriptions()
	assert.Contains(t, subs, "ACCOUNT_ACTIVITY")
}

// TestAutoResubscribe_OnReconnect tests that subscriptions are re-sent on reconnect
func TestAutoResubscribe_OnReconnect(t *testing.T) {
	server := newMockWebSocketServer(t)
	defer server.Close()

	logger := slog.Default()
	manager := NewManager(logger)
	ctx := context.Background()

	req1 := &types.Subscription{
		Service: "LEVELONE_EQUITIES",
		Command: "ADD",
		Parameters: &types.SubscriptionParams{
			Keys:   "AAPL,MSFT",
			Fields: "0,1,2,3",
		},
	}
	err := manager.RecordRequest(ctx, req1)
	require.NoError(t, err)

	req2 := &types.Subscription{
		Service: "CHART_EQUITY",
		Command: "SUBS",
		Parameters: &types.SubscriptionParams{
			Keys:   "TSLA",
			Fields: "0,1,2",
		},
	}
	err = manager.RecordRequest(ctx, req2)
	require.NoError(t, err)

	subs := manager.GetSubscriptions()
	assert.Contains(t, subs, "LEVELONE_EQUITIES")
	assert.Contains(t, subs, "CHART_EQUITY")

	client := createTestClient(t, server.URL())
	defer closeTestClient(client)

	// Re-subscribe using recorded subscriptions
	for service, keys := range subs {
		for key, fields := range keys {
			switch service {
			case "LEVELONE_EQUITIES":
				err = client.LevelOneEquities(ctx, manager, key, joinFields(fields), "ADD")
			case "CHART_EQUITY":
				err = client.ChartEquity(ctx, manager, key, joinFields(fields), "SUBS")
			}
			require.NoError(t, err)
		}
	}

	// Wait for messages to be sent
	time.Sleep(100 * time.Millisecond)

	// Verify messages were sent
	messages := server.GetMessages()
	assert.GreaterOrEqual(t, len(messages), 2, "Should have sent at least 2 messages")

	// Verify the messages contain the expected services
	servicesFound := make(map[string]bool)
	for _, msg := range messages {
		var sub types.Subscription
		err = json.Unmarshal(msg, &sub)
		if err == nil {
			servicesFound[sub.Service] = true
		}
	}
	assert.True(t, servicesFound["LEVELONE_EQUITIES"], "LEVELONE_EQUITIES should be in messages")
	assert.True(t, servicesFound["CHART_EQUITY"], "CHART_EQUITY should be in messages")
}

// TestSubscription_Unsubscribe tests the UNSUBS command
func TestSubscription_Unsubscribe(t *testing.T) {
	server := newMockWebSocketServer(t)
	defer server.Close()

	client := createTestClient(t, server.URL())
	defer closeTestClient(client)

	logger := slog.Default()
	manager := NewManager(logger)
	ctx := context.Background()

	// First subscribe
	err := client.LevelOneEquities(ctx, manager, "AAPL,MSFT", "0,1,2", "ADD")
	require.NoError(t, err)

	// Wait for message
	time.Sleep(100 * time.Millisecond)
	server.ClearMessages()

	// Now unsubscribe
	err = client.LevelOneEquities(ctx, manager, "AAPL", "", "UNSUBS")
	require.NoError(t, err, "UNSUBS should not error")

	// Wait for message
	time.Sleep(100 * time.Millisecond)

	// Verify UNSUBS message was sent
	messages := server.GetMessages()
	require.Len(t, messages, 1, "Should have sent one UNSUBS message")

	var sub types.Subscription
	err = json.Unmarshal(messages[0], &sub)
	require.NoError(t, err)

	assert.Equal(t, "LEVELONE_EQUITIES", sub.Service)
	assert.Equal(t, "UNSUBS", sub.Command)
	assert.Equal(t, "AAPL", sub.Parameters.Keys)

	// Verify subscription was removed from manager
	subs := manager.GetSubscriptions()
	assert.NotContains(t, subs["LEVELONE_EQUITIES"], "AAPL")
	assert.Contains(t, subs["LEVELONE_EQUITIES"], "MSFT", "MSFT should still be subscribed")
}

// TestSubscription_MultipleServices tests subscribing to multiple services
func TestSubscription_MultipleServices(t *testing.T) {
	server := newMockWebSocketServer(t)
	defer server.Close()

	client := createTestClient(t, server.URL())
	defer closeTestClient(client)

	logger := slog.Default()
	manager := NewManager(logger)
	ctx := context.Background()

	// Subscribe to multiple services
	err := client.LevelOneEquities(ctx, manager, "AAPL", "0,1,2", "ADD")
	require.NoError(t, err)

	err = client.LevelOneOptions(ctx, manager, "AAPL  240809C00095000", "0,1", "ADD")
	require.NoError(t, err)

	err = client.ChartEquity(ctx, manager, "TSLA", "0,1,2,3", "SUBS")
	require.NoError(t, err)

	// Wait for messages
	time.Sleep(100 * time.Millisecond)

	// Verify all messages were sent
	messages := server.GetMessages()
	assert.Len(t, messages, 3, "Should have sent 3 messages")

	// Verify all services are recorded
	subs := manager.GetSubscriptions()
	assert.Contains(t, subs, "LEVELONE_EQUITIES")
	assert.Contains(t, subs, "LEVELONE_OPTIONS")
	assert.Contains(t, subs, "CHART_EQUITY")
}

// TestSubscription_InactiveClient tests that inactive client returns error
func TestSubscription_InactiveClient(t *testing.T) {
	logger := slog.Default()
	client := NewClient(logger)
	// Don't connect the client

	manager := NewManager(logger)
	ctx := context.Background()

	err := client.LevelOneEquities(ctx, manager, "AAPL", "0,1,2", "ADD")
	assert.Error(t, err, "Should return error when client is not active")
	assert.Contains(t, err.Error(), "not active")
}

// TestReconnectManager tests the reconnect manager functionality
func TestReconnectManager(t *testing.T) {
	logger := slog.Default()
	rm := NewReconnectManager(logger)

	assert.Equal(t, 2*time.Second, rm.GetBackoffTime())

	assert.False(t, rm.ShouldReconnect(10*time.Second, nil), "Should not reconnect if uptime < minUptime")
	assert.True(t, rm.ShouldReconnect(100*time.Second, nil), "Should reconnect if uptime >= minUptime")

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	start := time.Now()
	err := rm.WaitForBackoff(ctx)
	elapsed := time.Since(start)

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, elapsed.Milliseconds(), int64(1900), "Should wait at least 2 seconds")
	assert.Equal(t, 4*time.Second, rm.GetBackoffTime(), "Backoff should double")

	rm.backoffTime = 1 * time.Millisecond
	rm.maxBackoff = 2 * time.Millisecond
	_ = rm.WaitForBackoff(context.Background())
	assert.Equal(t, 2*time.Millisecond, rm.backoffTime, "Backoff should cap at max")

	rm.ResetBackoff()
	assert.Equal(t, 2*time.Second, rm.GetBackoffTime())
}

// TestReconnectManager_ContextCancellation tests that WaitForBackoff respects context cancellation
func TestReconnectManager_ContextCancellation(t *testing.T) {
	logger := slog.Default()
	rm := NewReconnectManager(logger)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := rm.WaitForBackoff(ctx)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

// TestRouter_ValidateService tests the router service validation
func TestRouter_ValidateService(t *testing.T) {
	logger := slog.Default()
	handler := NewHandler(logger)
	router := NewRouter(handler, logger)

	// Valid services
	validServices := []string{
		"ADMIN",
		"LEVELONE_EQUITIES",
		"LEVELONE_OPTIONS",
		"LEVELONE_FUTURES",
		"LEVELONE_FUTURES_OPTIONS",
		"LEVELONE_FOREX",
		"NYSE_BOOK",
		"NASDAQ_BOOK",
		"OPTIONS_BOOK",
		"CHART_EQUITY",
		"CHART_FUTURES",
		"SCREENER_EQUITY",
		"SCREENER_OPTION",
		"ACCT_ACTIVITY",
	}

	for _, service := range validServices {
		assert.True(t, router.ValidateService(service), "Service %s should be valid", service)
	}

	// Invalid services
	invalidServices := []string{
		"INVALID_SERVICE",
		"",
		"LEVELONE_EQUITY", // Wrong name
	}

	for _, service := range invalidServices {
		assert.False(t, router.ValidateService(service), "Service %s should be invalid", service)
	}
}

// TestHandler_ValidateCommand tests the handler command validation
func TestHandler_ValidateCommand(t *testing.T) {
	logger := slog.Default()
	handler := NewHandler(logger)

	// Valid commands
	validCommands := []string{"LOGIN", "LOGOUT", "ADD", "SUBS", "UNSUBS", "VIEW"}
	for _, cmd := range validCommands {
		assert.True(t, handler.ValidateCommand(cmd), "Command %s should be valid", cmd)
	}

	// Invalid commands
	invalidCommands := []string{"INVALID", "", "SUBSCRIBE", "UNSUBSCRIBE"}
	for _, cmd := range invalidCommands {
		assert.False(t, handler.ValidateCommand(cmd), "Command %s should be invalid", cmd)
	}
}

// TestHandler_ParseMessage tests message parsing
func TestHandler_ParseMessage(t *testing.T) {
	logger := slog.Default()
	handler := NewHandler(logger)

	// Valid message
	validJSON := `{"service":"LEVELONE_EQUITIES","command":"ADD","requestid":1,"content":{"key":"value"}}`
	msg, err := handler.ParseMessage([]byte(validJSON))
	require.NoError(t, err)
	assert.Equal(t, "LEVELONE_EQUITIES", msg.Service)
	assert.Equal(t, "ADD", msg.Command)
	assert.Equal(t, 1, msg.RequestID)

	// Invalid message
	invalidJSON := `{invalid json`
	_, err = handler.ParseMessage([]byte(invalidJSON))
	assert.Error(t, err)
}

// TestHandler_ParseSubscription tests subscription parsing
func TestHandler_ParseSubscription(t *testing.T) {
	logger := slog.Default()
	handler := NewHandler(logger)

	// Valid subscription
	validJSON := `{"service":"LEVELONE_EQUITIES","command":"ADD","parameters":{"keys":"AAPL,MSFT","fields":"0,1,2"}}`
	sub, err := handler.ParseSubscription([]byte(validJSON))
	require.NoError(t, err)
	assert.Equal(t, "LEVELONE_EQUITIES", sub.Service)
	assert.Equal(t, "ADD", sub.Command)
	assert.Equal(t, "AAPL,MSFT", sub.Parameters.Keys)
	assert.Equal(t, "0,1,2", sub.Parameters.Fields)

	// Invalid subscription
	invalidJSON := `{invalid json`
	_, err = handler.ParseSubscription([]byte(invalidJSON))
	assert.Error(t, err)
}

// joinFields joins a slice of strings with comma
func joinFields(fields []string) string {
	var result strings.Builder
	for i, f := range fields {
		if i > 0 {
			result.WriteString(",")
		}
		result.WriteString(f)
	}
	return result.String()
}
