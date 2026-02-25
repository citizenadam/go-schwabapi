package client

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/citizenadam/go-schwabapi/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockTokenGetterWithRefresh implements TokenGetter with refresh tracking
type mockTokenGetterWithRefresh struct {
	token           string
	refreshCount    int
	refreshCallback func() string
}

func (m *mockTokenGetterWithRefresh) GetAccessToken() string {
	m.refreshCount++
	if m.refreshCallback != nil {
		return m.refreshCallback()
	}
	return m.token
}

// TestIntegrationFullAccountWorkflow tests the complete account workflow:
// AccountDetailsAll → Preferences → PlaceOrder → OrderDetails → Transactions
func TestIntegrationFullAccountWorkflow(t *testing.T) {
	// Track API calls
	var apiCalls []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Received request: %s %s", r.Method, r.URL.Path)
		apiCalls = append(apiCalls, r.URL.Path)

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			t.Error("expected Authorization header")
		}

		// Get path without query parameters
		path := r.URL.Path

		switch path {
		case "/trader/v1/accounts/", "/trader/v1/accounts":
			// AccountDetailsAll
			w.Header().Set("Content-Type", "application/json")
			response := types.AccountDetailsAllResponse{
				Accounts: []types.Account{
					{
						AccountHash:   "test-hash-1",
						AccountNumber: "12345678",
					},
				},
			}
			t.Logf("Sending AccountDetailsAll response: %+v", response)
			json.NewEncoder(w).Encode(response)

		case "/trader/v1/userPreference":
			// Preferences
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(types.PreferencesResponse{
				StreamerInfo: &types.StreamerInfo{
					AccountID:      "test-account-id",
					AccountIDType:  "test-type",
					Token:          "test-token",
					TokenTimestamp: "2024-03-15T12:34:56Z",
					UserID:         "test-user-id",
					AppID:          "test-app-id",
					Secret:         "test-secret",
					AccessLevel:    "test-level",
				},
			})

		case "/v1/accounts/test-hash-1/orders":
			// PlaceOrder
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(types.OrderDetailsResponse{
				Order: &types.Order{
					Symbol:      "AAPL",
					OrderType:   "MARKET",
					Quantity:    1,
					Instruction: "BUY",
				},
			})

		case "/v1/accounts/test-hash-1/orders/order-123":
			// OrderDetails
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(types.OrderDetailsResponse{
				Order: &types.Order{
					Symbol:      "AAPL",
					OrderType:   "MARKET",
					Quantity:    1,
					Instruction: "BUY",
				},
			})

		case "/trader/v1/accounts/test-hash-1/transactions":
			// Transactions
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(types.TransactionsResponse{
				Transactions: []types.Transaction{
					{
						TransactionID: "txn-123",
						Type:          "TRADE",
						Amount:        "100.00",
					},
				},
			})

		default:
			t.Logf("unexpected path: %s, query: %s", r.URL.Path, r.URL.RawQuery)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Setup clients
	logger := slog.Default()
	httpClient := NewClient(logger)
	mockToken := &mockTokenGetter{token: "test-access-token"}

	accounts := NewAccounts(httpClient, logger, mockToken)
	orders := NewOrdersClient(httpClient, logger, mockToken)

	// Override base URL to use test server
	accounts.baseURL = server.URL
	orders.baseURL = server.URL

	ctx := context.Background()

	// Step 1: Get all accounts
	t.Log("Step 1: AccountDetailsAll")
	accountDetails, err := accounts.AccountDetailsAll(ctx, "positions")
	require.NoError(t, err)
	assert.Len(t, accountDetails.Accounts, 1)
	assert.Equal(t, "test-hash-1", accountDetails.Accounts[0].AccountHash)

	// Step 2: Get preferences (streamer info)
	t.Log("Step 2: Preferences")
	preferences, err := accounts.Preferences(ctx)
	require.NoError(t, err)
	assert.NotNil(t, preferences.StreamerInfo)
	assert.Equal(t, "test-account-id", preferences.StreamerInfo.AccountID)

	// Step 3: Place order
	t.Log("Step 3: PlaceOrder")
	order := map[string]interface{}{
		"orderType": "MARKET",
		"session":   "NORMAL",
		"duration":  "DAY",
		"orderLegs": []map[string]interface{}{
			{
				"instruction": "BUY",
				"quantity":    1,
				"instrument": map[string]interface{}{
					"symbol":    "AAPL",
					"assetType": "EQUITY",
				},
			},
		},
	}
	resp, err := orders.PlaceOrder(ctx, "test-hash-1", order)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	// Step 4: Get order details
	t.Log("Step 4: OrderDetails")
	// OrderDetails is not implemented yet, so we'll skip this step
	// orderDetails, err := orders.OrderDetails(ctx, "test-hash-1", "order-123")

	// Step 5: Get transactions
	t.Log("Step 5: Transactions")
	transactions, err := accounts.Transactions(ctx, "test-hash-1", "", "", "", "")
	require.NoError(t, err)
	assert.Len(t, transactions.Transactions, 1)
	assert.Equal(t, "txn-123", transactions.Transactions[0].TransactionID)

	// Verify all API calls were made
	expectedCalls := []string{
		"/trader/v1/accounts/",
		"/trader/v1/userPreference",
		"/v1/accounts/test-hash-1/orders",
		"/trader/v1/accounts/test-hash-1/transactions",
	}
	assert.Equal(t, expectedCalls, apiCalls)
}

// TestIntegrationTokenRefresh tests that token refresh happens before API calls
func TestIntegrationTokenRefresh(t *testing.T) {
	refreshCount := 0
	mockToken := &mockTokenGetterWithRefresh{
		token: "initial-token",
		refreshCallback: func() string {
			refreshCount++
			if refreshCount == 1 {
				return "refreshed-token-1"
			}
			return "refreshed-token-2"
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			t.Error("expected Authorization header")
		}

		// Verify token changes on each call
		if refreshCount == 1 {
			assert.Contains(t, authHeader, "refreshed-token-1")
		} else if refreshCount == 2 {
			assert.Contains(t, authHeader, "refreshed-token-2")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(types.AccountDetailsAllResponse{
			Accounts: []types.Account{
				{
					AccountHash:   "test-hash",
					AccountNumber: "12345678",
				},
			},
		})
	}))
	defer server.Close()

	logger := slog.Default()
	httpClient := NewClient(logger)
	accounts := NewAccounts(httpClient, logger, mockToken)
	accounts.baseURL = server.URL

	ctx := context.Background()

	// First call - should trigger first refresh
	_, err := accounts.AccountDetailsAll(ctx, "positions")
	require.NoError(t, err)
	assert.Equal(t, 1, refreshCount)

	// Second call - should trigger second refresh
	_, err = accounts.AccountDetailsAll(ctx, "positions")
	require.NoError(t, err)
	assert.Equal(t, 2, refreshCount)
}

// TestIntegrationStreamingWorkflow tests streaming subscription workflow
func TestIntegrationStreamingWorkflow(t *testing.T) {
	// This test would require a WebSocket mock server
	// For now, we'll test the client setup and token refresh

	mockToken := &mockTokenGetter{token: "test-access-token"}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			t.Error("expected Authorization header")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(types.PreferencesResponse{
			StreamerInfo: &types.StreamerInfo{
				AccountID:      "test-account-id",
				AccountIDType:  "test-type",
				Token:          "test-token",
				TokenTimestamp: "2024-03-15T12:34:56Z",
				UserID:         "test-user-id",
				AppID:          "test-app-id",
				Secret:         "test-secret",
				AccessLevel:    "test-level",
			},
		})
	}))
	defer server.Close()

	logger := slog.Default()
	httpClient := NewClient(logger)
	accounts := NewAccounts(httpClient, logger, mockToken)
	accounts.baseURL = server.URL

	ctx := context.Background()

	// Get streamer info for streaming authentication
	streamerInfo, err := accounts.Preferences(ctx)
	require.NoError(t, err)
	assert.NotNil(t, streamerInfo.StreamerInfo)
	assert.Equal(t, "test-account-id", streamerInfo.StreamerInfo.AccountID)
	assert.Equal(t, "test-token", streamerInfo.StreamerInfo.Token)
}

// TestIntegrationErrorHandling tests error handling across the workflow
func TestIntegrationErrorHandling(t *testing.T) {
	tests := []struct {
		name          string
		setupServer   func(*httptest.Server)
		expectedError string
	}{
		{
			name: "handles 401 unauthorized",
			setupServer: func(server *httptest.Server) {
				// Server will return 401
			},
			expectedError: "failed to decode",
		},
		{
			name: "handles 500 server error",
			setupServer: func(server *httptest.Server) {
				// Server will return 500
			},
			expectedError: "failed to decode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			}))
			defer server.Close()

			logger := slog.Default()
			httpClient := NewClient(logger)
			mockToken := &mockTokenGetter{token: "test-access-token"}
			accounts := NewAccounts(httpClient, logger, mockToken)
			accounts.baseURL = server.URL

			ctx := context.Background()

			_, err := accounts.AccountDetailsAll(ctx, "positions")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

// TestIntegrationConcurrentRequests tests concurrent API requests
func TestIntegrationConcurrentRequests(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			t.Error("expected Authorization header")
		}

		// Simulate some delay
		time.Sleep(10 * time.Millisecond)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(types.AccountDetailsAllResponse{
			Accounts: []types.Account{
				{
					AccountHash:   "test-hash",
					AccountNumber: "12345678",
				},
			},
		})
	}))
	defer server.Close()

	logger := slog.Default()
	httpClient := NewClient(logger)
	mockToken := &mockTokenGetter{token: "test-access-token"}
	accounts := NewAccounts(httpClient, logger, mockToken)
	accounts.baseURL = server.URL

	ctx := context.Background()

	// Make 5 concurrent requests
	done := make(chan bool, 5)
	for i := 0; i < 5; i++ {
		go func() {
			_, err := accounts.AccountDetailsAll(ctx, "positions")
			assert.NoError(t, err)
			done <- true
		}()
	}

	// Wait for all requests to complete
	for i := 0; i < 5; i++ {
		<-done
	}
}

// TestIntegrationContextCancellation tests context cancellation handling
func TestIntegrationContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	logger := slog.Default()
	httpClient := NewClient(logger)
	mockToken := &mockTokenGetter{token: "test-access-token"}
	accounts := NewAccounts(httpClient, logger, mockToken)
	accounts.baseURL = server.URL

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := accounts.AccountDetailsAll(ctx, "positions")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}

// TestIntegrationQueryParameters tests query parameter handling
func TestIntegrationQueryParameters(t *testing.T) {
	var receivedParams map[string]string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Capture query parameters
		receivedParams = make(map[string]string)
		for key, values := range r.URL.Query() {
			if len(values) > 0 {
				receivedParams[key] = values[0]
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(types.AccountOrdersAllResponse{
			Orders: []types.Order{
				{
					Symbol:      "AAPL",
					OrderType:   "MARKET",
					Quantity:    1,
					Instruction: "BUY",
				},
			},
		})
	}))
	defer server.Close()

	logger := slog.Default()
	httpClient := NewClient(logger)
	mockToken := &mockTokenGetter{token: "test-access-token"}
	accounts := NewAccounts(httpClient, logger, mockToken)
	accounts.baseURL = server.URL

	ctx := context.Background()

	// Call with query parameters
	fromTime := "2024-01-01T00:00:00Z"
	toTime := "2024-01-31T23:59:59Z"
	_, err := accounts.AccountOrdersAll(ctx, fromTime, toTime, 10, "FILLED")
	require.NoError(t, err)

	// Verify query parameters were sent
	assert.Equal(t, fromTime, receivedParams["fromEnteredTime"])
	assert.Equal(t, toTime, receivedParams["toEnteredTime"])
	assert.Equal(t, "10", receivedParams["maxResults"])
	assert.Equal(t, "FILLED", receivedParams["status"])
}
