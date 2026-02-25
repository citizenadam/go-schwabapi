package pkg

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/citizenadam/go-schwabapi/pkg/client"
	"github.com/citizenadam/go-schwabapi/pkg/stream"
	"github.com/citizenadam/go-schwabapi/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockTokenGetter struct {
	token string
}

func (m *mockTokenGetter) GetAccessToken() string {
	return m.token
}

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

func TestIntegrationFullAccountWorkflow(t *testing.T) {
	var apiCalls []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Received request: %s %s", r.Method, r.URL.Path)
		apiCalls = append(apiCalls, r.URL.Path)

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			t.Error("expected Authorization header")
		}

		path := r.URL.Path

		switch path {
		case "/trader/v1/accounts/", "/trader/v1/accounts":
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

	logger := slog.Default()
	httpClient := client.NewClient(logger)
	mockToken := &mockTokenGetter{token: "test-access-token"}

	accounts := client.NewAccounts(httpClient, logger, mockToken)
	orders := client.NewOrdersClient(httpClient, logger, mockToken)

	accounts.SetBaseURL(server.URL)
	orders.SetBaseURL(server.URL)

	ctx := context.Background()

	t.Log("Step 1: AccountDetailsAll")
	accountDetails, err := accounts.AccountDetailsAll(ctx, "positions")
	require.NoError(t, err)
	assert.Len(t, accountDetails.Accounts, 1)
	assert.Equal(t, "test-hash-1", accountDetails.Accounts[0].AccountHash)

	t.Log("Step 2: Preferences")
	preferences, err := accounts.Preferences(ctx)
	require.NoError(t, err)
	assert.NotNil(t, preferences.StreamerInfo)
	assert.Equal(t, "test-account-id", preferences.StreamerInfo.AccountID)

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

	t.Log("Step 4: OrderDetails")

	t.Log("Step 5: Transactions")
	transactions, err := accounts.Transactions(ctx, "test-hash-1", "", "", "", "")
	require.NoError(t, err)
	assert.Len(t, transactions.Transactions, 1)
	assert.Equal(t, "txn-123", transactions.Transactions[0].TransactionID)

	expectedCalls := []string{
		"/trader/v1/accounts/",
		"/trader/v1/userPreference",
		"/v1/accounts/test-hash-1/orders",
		"/trader/v1/accounts/test-hash-1/transactions",
	}
	assert.Equal(t, expectedCalls, apiCalls)
}

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
	httpClient := client.NewClient(logger)
	accounts := client.NewAccounts(httpClient, logger, mockToken)
	accounts.SetBaseURL(server.URL)

	ctx := context.Background()

	_, err := accounts.AccountDetailsAll(ctx, "positions")
	require.NoError(t, err)
	assert.Equal(t, 1, refreshCount)

	_, err = accounts.AccountDetailsAll(ctx, "positions")
	require.NoError(t, err)
	assert.Equal(t, 2, refreshCount)
}

func TestIntegrationStreamingWorkflow(t *testing.T) {
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
	httpClient := client.NewClient(logger)
	accounts := client.NewAccounts(httpClient, logger, mockToken)
	accounts.SetBaseURL(server.URL)

	ctx := context.Background()

	streamerInfo, err := accounts.Preferences(ctx)
	require.NoError(t, err)
	assert.NotNil(t, streamerInfo.StreamerInfo)
	assert.Equal(t, "test-account-id", streamerInfo.StreamerInfo.AccountID)
	assert.Equal(t, "test-token", streamerInfo.StreamerInfo.Token)

	manager := stream.NewManager(logger)
	assert.NotNil(t, manager)

	req := &types.Subscription{
		Command: "ADD",
		Service: "LEVELONE_EQUITIES",
		Parameters: &types.SubscriptionParams{
			Keys:   "AAPL,MSFT",
			Fields: "0,1,2",
		},
	}
	err = manager.RecordRequest(ctx, req)
	require.NoError(t, err)

	subs := manager.GetSubscriptions()
	assert.Len(t, subs, 1)
	assert.Contains(t, subs, "LEVELONE_EQUITIES")
	assert.Len(t, subs["LEVELONE_EQUITIES"], 2)
	assert.Contains(t, subs["LEVELONE_EQUITIES"], "AAPL")
	assert.Contains(t, subs["LEVELONE_EQUITIES"], "MSFT")
}

func TestIntegrationStreamingReconnect(t *testing.T) {
	logger := slog.Default()
	manager := stream.NewManager(logger)

	ctx := context.Background()

	req1 := &types.Subscription{
		Command: "ADD",
		Service: "LEVELONE_EQUITIES",
		Parameters: &types.SubscriptionParams{
			Keys:   "AAPL,MSFT",
			Fields: "0,1,2",
		},
	}
	err := manager.RecordRequest(ctx, req1)
	require.NoError(t, err)

	subs1 := manager.GetSubscriptions()
	assert.Len(t, subs1, 1)
	assert.Len(t, subs1["LEVELONE_EQUITIES"], 2)

	manager.Clear()

	req2 := &types.Subscription{
		Command: "ADD",
		Service: "LEVELONE_EQUITIES",
		Parameters: &types.SubscriptionParams{
			Keys:   "AAPL,MSFT",
			Fields: "0,1,2",
		},
	}
	err = manager.RecordRequest(ctx, req2)
	require.NoError(t, err)

	subs2 := manager.GetSubscriptions()
	assert.Len(t, subs2, 1)
	assert.Len(t, subs2["LEVELONE_EQUITIES"], 2)
	assert.Contains(t, subs2["LEVELONE_EQUITIES"], "AAPL")
	assert.Contains(t, subs2["LEVELONE_EQUITIES"], "MSFT")
}

func TestIntegrationStreamingAutoResubscribe(t *testing.T) {
	logger := slog.Default()
	manager := stream.NewManager(logger)

	ctx := context.Background()

	req1 := &types.Subscription{
		Command: "ADD",
		Service: "LEVELONE_EQUITIES",
		Parameters: &types.SubscriptionParams{
			Keys:   "AAPL,MSFT",
			Fields: "0,1,2",
		},
	}
	err := manager.RecordRequest(ctx, req1)
	require.NoError(t, err)

	req2 := &types.Subscription{
		Command: "ADD",
		Service: "LEVELONE_OPTIONS",
		Parameters: &types.SubscriptionParams{
			Keys:   "AAPL  240809C00095000",
			Fields: "0,1,2,3",
		},
	}
	err = manager.RecordRequest(ctx, req2)
	require.NoError(t, err)

	subs := manager.GetSubscriptions()
	assert.Len(t, subs, 2)
	assert.Contains(t, subs, "LEVELONE_EQUITIES")
	assert.Contains(t, subs, "LEVELONE_OPTIONS")

	allSubs := manager.GetSubscriptions()
	assert.NotNil(t, allSubs)
	assert.Len(t, allSubs, 2)

	manager.Clear()

	for service, keys := range allSubs {
		for key, fields := range keys {
			fieldsStr := ""
			for i, f := range fields {
				if i > 0 {
					fieldsStr += ","
				}
				fieldsStr += f
			}
			req := &types.Subscription{
				Command: "ADD",
				Service: service,
				Parameters: &types.SubscriptionParams{
					Keys:   key,
					Fields: fieldsStr,
				},
			}
			err = manager.RecordRequest(ctx, req)
			require.NoError(t, err)
		}
	}

	subsAfter := manager.GetSubscriptions()
	assert.Len(t, subsAfter, 2)
	assert.Contains(t, subsAfter, "LEVELONE_EQUITIES")
	assert.Contains(t, subsAfter, "LEVELONE_OPTIONS")
}

func TestIntegrationErrorHandling(t *testing.T) {
	tests := []struct {
		name          string
		setupServer   func(*httptest.Server)
		expectedError string
	}{
		{
			name:          "handles 401 unauthorized",
			setupServer:   func(server *httptest.Server) {},
			expectedError: "failed to decode",
		},
		{
			name:          "handles 500 server error",
			setupServer:   func(server *httptest.Server) {},
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
			httpClient := client.NewClient(logger)
			mockToken := &mockTokenGetter{token: "test-access-token"}
			accounts := client.NewAccounts(httpClient, logger, mockToken)
			accounts.SetBaseURL(server.URL)

			ctx := context.Background()

			_, err := accounts.AccountDetailsAll(ctx, "positions")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

func TestIntegrationConcurrentRequests(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			t.Error("expected Authorization header")
		}

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
	httpClient := client.NewClient(logger)
	mockToken := &mockTokenGetter{token: "test-access-token"}
	accounts := client.NewAccounts(httpClient, logger, mockToken)
	accounts.SetBaseURL(server.URL)

	ctx := context.Background()

	done := make(chan bool, 5)
	for i := 0; i < 5; i++ {
		go func() {
			_, err := accounts.AccountDetailsAll(ctx, "positions")
			assert.NoError(t, err)
			done <- true
		}()
	}

	for i := 0; i < 5; i++ {
		<-done
	}
}

func TestIntegrationContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	logger := slog.Default()
	httpClient := client.NewClient(logger)
	mockToken := &mockTokenGetter{token: "test-access-token"}
	accounts := client.NewAccounts(httpClient, logger, mockToken)
	accounts.SetBaseURL(server.URL)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := accounts.AccountDetailsAll(ctx, "positions")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}

func TestIntegrationQueryParameters(t *testing.T) {
	var receivedParams map[string]string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	httpClient := client.NewClient(logger)
	mockToken := &mockTokenGetter{token: "test-access-token"}
	accounts := client.NewAccounts(httpClient, logger, mockToken)
	accounts.SetBaseURL(server.URL)

	ctx := context.Background()

	fromTime := "2024-01-01T00:00:00Z"
	toTime := "2024-01-31T23:59:59Z"
	_, err := accounts.AccountOrdersAll(ctx, fromTime, toTime, 10, "FILLED")
	require.NoError(t, err)

	assert.Equal(t, fromTime, receivedParams["fromEnteredTime"])
	assert.Equal(t, toTime, receivedParams["toEnteredTime"])
	assert.Equal(t, "10", receivedParams["maxResults"])
	assert.Equal(t, "FILLED", receivedParams["status"])
}

func TestIntegrationStreamingManagerPersistence(t *testing.T) {
	logger := slog.Default()
	manager := stream.NewManager(logger)

	ctx := context.Background()

	req := &types.Subscription{
		Command: "ADD",
		Service: "LEVELONE_EQUITIES",
		Parameters: &types.SubscriptionParams{
			Keys:   "AAPL,MSFT,GOOGL",
			Fields: "0,1,2,3",
		},
	}
	err := manager.RecordRequest(ctx, req)
	require.NoError(t, err)

	subs1 := manager.GetSubscriptions()
	subs2 := manager.GetSubscriptions()

	assert.Equal(t, len(subs1["LEVELONE_EQUITIES"]), len(subs2["LEVELONE_EQUITIES"]))
	assert.Equal(t, 3, len(subs1["LEVELONE_EQUITIES"]))

	subs1["LEVELONE_EQUITIES"]["TSLA"] = []string{"0", "1"}
	subs3 := manager.GetSubscriptions()
	assert.NotContains(t, subs3["LEVELONE_EQUITIES"], "TSLA")
}
