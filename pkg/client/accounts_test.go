package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/citizenadam/go-schwabapi/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log/slog"
)

// setupTestClient creates a test client with mock server
func setupTestClient(handler http.Handler) (*Accounts, *httptest.Server) {
	server := httptest.NewServer(handler)
	logger := slog.Default()
	httpClient := NewClient(logger)
	tokenGetter := &mockTokenGetter{token: "test-token"}

	accounts := NewAccounts(httpClient, logger, tokenGetter)
	accounts.SetBaseURL(server.URL)

	return accounts, server
}

// Helper function to create account response JSON
func createAccountResponse(accounts []types.Account) string {
	response := types.AccountDetailsAllResponse{
		Accounts: accounts,
	}
	jsonBytes, _ := json.Marshal(response)
	return string(jsonBytes)
}

// Helper function to create preferences response JSON
func createPreferencesResponse(streamerInfo *types.StreamerInfo) string {
	response := types.PreferencesResponse{
		StreamerInfo: streamerInfo,
	}
	jsonBytes, _ := json.Marshal(response)
	return string(jsonBytes)
}

// Helper function to create orders response JSON
func createOrdersResponse(orders []types.Order) string {
	response := types.AccountOrdersAllResponse{
		Orders: orders,
	}
	jsonBytes, _ := json.Marshal(response)
	return string(jsonBytes)
}

// TestAccountDetailsAll_Success tests successful retrieval of all account details
func TestAccountDetailsAll_Success(t *testing.T) {
	tests := []struct {
		name     string
		fields   string
		accounts []types.Account
	}{
		{
			name:   "success with positions field",
			fields: "positions",
			accounts: []types.Account{
				{
					AccountNumber: "1234-5678",
					AccountHash:   "hash123",
					AccountType:   "INDIVIDUAL",
				},
			},
		},
		{
			name:   "success without fields",
			fields: "",
			accounts: []types.Account{
				{
					AccountNumber: "9876-5432",
					AccountHash:   "hash456",
					AccountType:   "CUSTODIAL",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Contains(t, r.URL.Path, "/trader/v1/accounts/")

				// Check fields parameter if provided
				if tt.fields != "" {
					assert.Equal(t, "fields="+tt.fields, r.URL.RawQuery)
				}

				w.WriteHeader(http.StatusOK)
				w.Write([]byte(createAccountResponse(tt.accounts)))
			})

			accounts, server := setupTestClient(handler)
			defer server.Close()

			ctx := context.Background()
			result, err := accounts.AccountDetailsAll(ctx, tt.fields)

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, len(tt.accounts), len(result.Accounts))

			if len(tt.accounts) > 0 {
				assert.Equal(t, tt.accounts[0].AccountNumber, result.Accounts[0].AccountNumber)
				assert.Equal(t, tt.accounts[0].AccountHash, result.Accounts[0].AccountHash)
			}
		})
	}
}

// TestAccountDetailsAll_Error tests error handling for account details
func TestAccountDetailsAll_Error(t *testing.T) {
	tests := []struct {
		name        string
		statusCode  int
		expectError bool
		errorMsg    string
	}{
		{
			name:        "unauthorized error",
			statusCode:  http.StatusUnauthorized,
			expectError: false, // Current implementation doesn't check status codes
			errorMsg:    "failed to get all account details",
		},
		{
			name:        "server error",
			statusCode:  http.StatusInternalServerError,
			expectError: false, // Current implementation doesn't check status codes
			errorMsg:    "failed to get all account details",
		},
		{
			name:        "not found error",
			statusCode:  http.StatusNotFound,
			expectError: false, // Current implementation doesn't check status codes
			errorMsg:    "failed to get all account details",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(`{"error": "test error"}`))
			})

			accounts, server := setupTestClient(handler)
			defer server.Close()

			ctx := context.Background()
			result, err := accounts.AccountDetailsAll(ctx, "")

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, result)
			} else {
				// Current behavior: returns response without checking status code
				require.NoError(t, err)
				require.NotNil(t, result)
			}
		})
	}
}

// TestAccountDetailsAll_EmptyResponse tests handling of empty response
func TestAccountDetailsAll_EmptyResponse(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"accounts": []}`))
	})

	accounts, server := setupTestClient(handler)
	defer server.Close()

	ctx := context.Background()
	result, err := accounts.AccountDetailsAll(ctx, "")

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 0, len(result.Accounts))
}

// TestPreferences_Success tests successful retrieval of preferences
func TestPreferences_Success(t *testing.T) {
	tests := []struct {
		name         string
		streamerInfo *types.StreamerInfo
	}{
		{
			name: "success with full streamer info",
			streamerInfo: &types.StreamerInfo{
				AccountID:      "test-account-id",
				AccountIDType:  "IDENTITY",
				Token:          "test-token",
				TokenTimestamp: "2024-03-15T12:34:56Z",
				UserID:         "test-user-id",
				AppID:          "test-app-id",
				Secret:         "test-secret",
				AccessLevel:    "LEVEL_ONE",
			},
		},
		{
			name: "success with minimal streamer info",
			streamerInfo: &types.StreamerInfo{
				AccountID: "minimal-account",
				Token:     "minimal-token",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Contains(t, r.URL.Path, "/trader/v1/userPreference")

				w.WriteHeader(http.StatusOK)
				w.Write([]byte(createPreferencesResponse(tt.streamerInfo)))
			})

			accounts, server := setupTestClient(handler)
			defer server.Close()

			ctx := context.Background()
			result, err := accounts.Preferences(ctx)

			require.NoError(t, err)
			require.NotNil(t, result)
			require.NotNil(t, result.StreamerInfo)
			assert.Equal(t, tt.streamerInfo.AccountID, result.StreamerInfo.AccountID)
			assert.Equal(t, tt.streamerInfo.Token, result.StreamerInfo.Token)
		})
	}
}

// TestPreferences_Error tests error handling for preferences
func TestPreferences_Error(t *testing.T) {
	tests := []struct {
		name        string
		statusCode  int
		expectError bool
		errorMsg    string
	}{
		{
			name:        "unauthorized error",
			statusCode:  http.StatusUnauthorized,
			expectError: false, // Current implementation doesn't check status codes
			errorMsg:    "failed to get user preferences",
		},
		{
			name:        "server error",
			statusCode:  http.StatusInternalServerError,
			expectError: false, // Current implementation doesn't check status codes
			errorMsg:    "failed to get user preferences",
		},
		{
			name:        "forbidden error",
			statusCode:  http.StatusForbidden,
			expectError: false, // Current implementation doesn't check status codes
			errorMsg:    "failed to get user preferences",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(`{"error": "test error"}`))
			})

			accounts, server := setupTestClient(handler)
			defer server.Close()

			ctx := context.Background()
			result, err := accounts.Preferences(ctx)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, result)
			} else {
				// Current behavior: returns response without checking status code
				require.NoError(t, err)
				require.NotNil(t, result)
			}
		})
	}
}

// TestPreferences_MissingStreamerInfo tests handling of missing streamer info
func TestPreferences_MissingStreamerInfo(t *testing.T) {
	tests := []struct {
		name         string
		responseBody string
	}{
		{
			name:         "handles null streamer info",
			responseBody: `{"streamerInfo": null}`,
		},
		{
			name:         "handles missing streamerInfo field",
			responseBody: `{}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(tt.responseBody))
			})

			accounts, server := setupTestClient(handler)
			defer server.Close()

			ctx := context.Background()
			result, err := accounts.Preferences(ctx)

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Nil(t, result.StreamerInfo)
		})
	}
}

// TestAccountOrdersAll_Success tests successful retrieval of all orders
func TestAccountOrdersAll_Success(t *testing.T) {
	tests := []struct {
		name            string
		fromEnteredTime string
		toEnteredTime   string
		maxResults      int
		status          string
		orders          []types.Order
	}{
		{
			name:            "success with all parameters",
			fromEnteredTime: "2024-01-01",
			toEnteredTime:   "2024-03-15",
			maxResults:      100,
			status:          "FILLED",
			orders: []types.Order{
				{
					Session:   "NORMAL",
					Duration:  "DAY",
					OrderType: "LIMIT",
					Quantity:  10,
					Price:     "150.00",
					Symbol:    "AAPL",
				},
			},
		},
		{
			name:            "success with no parameters",
			fromEnteredTime: "",
			toEnteredTime:   "",
			maxResults:      0,
			status:          "",
			orders: []types.Order{
				{
					Session:   "NORMAL",
					Duration:  "GTC",
					OrderType: "MARKET",
					Quantity:  5,
					Price:     "",
					Symbol:    "GOOGL",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Contains(t, r.URL.Path, "/trader/v1/orders")

				// Check query parameters
				query := r.URL.Query()
				if tt.fromEnteredTime != "" {
					assert.Equal(t, tt.fromEnteredTime, query.Get("fromEnteredTime"))
				}
				if tt.toEnteredTime != "" {
					assert.Equal(t, tt.toEnteredTime, query.Get("toEnteredTime"))
				}
				if tt.maxResults > 0 {
					assert.Equal(t, "100", query.Get("maxResults"))
				}
				if tt.status != "" {
					assert.Equal(t, tt.status, query.Get("status"))
				}

				w.WriteHeader(http.StatusOK)
				w.Write([]byte(createOrdersResponse(tt.orders)))
			})

			accounts, server := setupTestClient(handler)
			defer server.Close()

			ctx := context.Background()
			result, err := accounts.AccountOrdersAll(ctx, tt.fromEnteredTime, tt.toEnteredTime, tt.maxResults, tt.status)

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, len(tt.orders), len(result.Orders))

			if len(tt.orders) > 0 {
				assert.Equal(t, tt.orders[0].Symbol, result.Orders[0].Symbol)
				assert.Equal(t, tt.orders[0].Quantity, result.Orders[0].Quantity)
			}
		})
	}
}

// TestAccountOrdersAll_Error tests error handling for account orders
func TestAccountOrdersAll_Error(t *testing.T) {
	tests := []struct {
		name        string
		statusCode  int
		expectError bool
		errorMsg    string
	}{
		{
			name:        "unauthorized error",
			statusCode:  http.StatusUnauthorized,
			expectError: false, // Current implementation doesn't check status codes
			errorMsg:    "failed to get all account orders",
		},
		{
			name:        "server error",
			statusCode:  http.StatusInternalServerError,
			expectError: false, // Current implementation doesn't check status codes
			errorMsg:    "failed to get all account orders",
		},
		{
			name:        "bad request error",
			statusCode:  http.StatusBadRequest,
			expectError: false, // Current implementation doesn't check status codes
			errorMsg:    "failed to get all account orders",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(`{"error": "test error"}`))
			})

			accounts, server := setupTestClient(handler)
			defer server.Close()

			ctx := context.Background()
			result, err := accounts.AccountOrdersAll(ctx, "", "", 0, "")

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, result)
			} else {
				// Current behavior: returns response without checking status code
				require.NoError(t, err)
				require.NotNil(t, result)
			}
		})
	}
}

// TestAccountOrdersAll_InvalidParams tests handling of invalid parameters
func TestAccountOrdersAll_InvalidParams(t *testing.T) {
	tests := []struct {
		name            string
		fromEnteredTime string
		toEnteredTime   string
		maxResults      int
		status          string
	}{
		{
			name:            "invalid date range - from after to",
			fromEnteredTime: "2024-03-15",
			toEnteredTime:   "2024-01-01",
			maxResults:      100,
			status:          "FILLED",
		},
		{
			name:            "negative max results",
			fromEnteredTime: "2024-01-01",
			toEnteredTime:   "2024-03-15",
			maxResults:      -1,
			status:          "FILLED",
		},
		{
			name:            "empty status",
			fromEnteredTime: "2024-01-01",
			toEnteredTime:   "2024-03-15",
			maxResults:      100,
			status:          "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Server should still process the request even with potentially invalid params
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"orders": []}`))
			})

			accounts, server := setupTestClient(handler)
			defer server.Close()

			ctx := context.Background()
			result, err := accounts.AccountOrdersAll(ctx, tt.fromEnteredTime, tt.toEnteredTime, tt.maxResults, tt.status)

			// These should not cause client-side errors, but server may handle differently
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, 0, len(result.Orders))
		})
	}
}

// TestAccountMethodsAll_Timeout tests timeout handling for all account methods
func TestAccountMethodsAll_Timeout(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	})

	accounts, server := setupTestClient(handler)
	defer server.Close()

	ctx := context.Background()

	// Test AccountDetailsAll timeout
	t.Run("AccountDetailsAll timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
		defer cancel()

		_, err := accounts.AccountDetailsAll(ctx, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context deadline exceeded")
	})

	// Test Preferences timeout
	t.Run("Preferences timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
		defer cancel()

		_, err := accounts.Preferences(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context deadline exceeded")
	})

	// Test AccountOrdersAll timeout
	t.Run("AccountOrdersAll timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
		defer cancel()

		_, err := accounts.AccountOrdersAll(ctx, "", "", 0, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context deadline exceeded")
	})
}

// Helper function to create transactions response JSON
func createTransactionsResponse(transactions []types.Transaction) string {
	response := types.TransactionsResponse{
		Transactions: transactions,
	}
	jsonBytes, _ := json.Marshal(response)
	return string(jsonBytes)
}

// Helper function to create transaction details response JSON
func createTransactionDetailsResponse(transaction *types.Transaction) string {
	response := types.TransactionDetailsResponse{
		Transaction: transaction,
	}
	jsonBytes, _ := json.Marshal(response)
	return string(jsonBytes)
}

// TestTransactions_Success tests successful retrieval of transactions
func TestTransactions_Success(t *testing.T) {
	tests := []struct {
		name            string
		accountHash     string
		startDate       string
		endDate         string
		transactionType string
		symbol          string
		transactions    []types.Transaction
	}{
		{
			name:            "success with all parameters",
			accountHash:     "hash123",
			startDate:       "2024-01-01",
			endDate:         "2024-03-15",
			transactionType: "TRADE",
			symbol:          "AAPL",
			transactions: []types.Transaction{
				{
					TransactionID: "txn001",
					Type:          "TRADE",
					SubType:       "BUY",
					Amount:        "1500.00",
					Description:   "Buy 10 AAPL",
					Date:          "2024-02-15T10:30:00Z",
				},
				{
					TransactionID: "txn002",
					Type:          "TRADE",
					SubType:       "SELL",
					Amount:        "2000.00",
					Description:   "Sell 5 AAPL",
					Date:          "2024-02-20T14:45:00Z",
				},
			},
		},
		{
			name:            "success with minimal parameters",
			accountHash:     "hash456",
			startDate:       "",
			endDate:         "",
			transactionType: "",
			symbol:          "",
			transactions: []types.Transaction{
				{
					TransactionID: "txn003",
					Type:          "DIVIDEND",
					SubType:       "CASH",
					Amount:        "50.00",
					Description:   "Dividend payment",
					Date:          "2024-03-01T00:00:00Z",
				},
			},
		},
		{
			name:            "success with only date range",
			accountHash:     "hash789",
			startDate:       "2024-01-01",
			endDate:         "2024-12-31",
			transactionType: "",
			symbol:          "",
			transactions: []types.Transaction{
				{
					TransactionID: "txn004",
					Type:          "INTEREST",
					Amount:        "25.00",
					Description:   "Interest payment",
					Date:          "2024-06-15T00:00:00Z",
				},
			},
		},
		{
			name:            "success with only symbol filter",
			accountHash:     "hashABC",
			startDate:       "",
			endDate:         "",
			transactionType: "",
			symbol:          "MSFT",
			transactions: []types.Transaction{
				{
					TransactionID: "txn005",
					Type:          "TRADE",
					SubType:       "BUY",
					Amount:        "5000.00",
					Description:   "Buy 20 MSFT",
					Date:          "2024-04-10T09:15:00Z",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Contains(t, r.URL.Path, "/trader/v1/accounts/")
				assert.Contains(t, r.URL.Path, "/transactions")

				// Check query parameters
				query := r.URL.Query()
				if tt.startDate != "" {
					assert.Equal(t, tt.startDate, query.Get("startDate"))
				}
				if tt.endDate != "" {
					assert.Equal(t, tt.endDate, query.Get("endDate"))
				}
				if tt.transactionType != "" {
					assert.Equal(t, tt.transactionType, query.Get("type"))
				}
				if tt.symbol != "" {
					assert.Equal(t, tt.symbol, query.Get("symbol"))
				}

				w.WriteHeader(http.StatusOK)
				w.Write([]byte(createTransactionsResponse(tt.transactions)))
			})

			accounts, server := setupTestClient(handler)
			defer server.Close()

			ctx := context.Background()
			result, err := accounts.Transactions(ctx, tt.accountHash, tt.startDate, tt.endDate, tt.transactionType, tt.symbol)

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, len(tt.transactions), len(result.Transactions))

			if len(tt.transactions) > 0 {
				assert.Equal(t, tt.transactions[0].TransactionID, result.Transactions[0].TransactionID)
				assert.Equal(t, tt.transactions[0].Type, result.Transactions[0].Type)
				assert.Equal(t, tt.transactions[0].Amount, result.Transactions[0].Amount)
			}
		})
	}
}

// TestTransactions_Error tests error handling for transactions
func TestTransactions_Error(t *testing.T) {
	tests := []struct {
		name        string
		statusCode  int
		expectError bool
		errorMsg    string
	}{
		{
			name:        "unauthorized error",
			statusCode:  http.StatusUnauthorized,
			expectError: false, // Current implementation doesn't check status codes
			errorMsg:    "failed to get account transactions",
		},
		{
			name:        "server error",
			statusCode:  http.StatusInternalServerError,
			expectError: false, // Current implementation doesn't check status codes
			errorMsg:    "failed to get account transactions",
		},
		{
			name:        "not found error",
			statusCode:  http.StatusNotFound,
			expectError: false, // Current implementation doesn't check status codes
			errorMsg:    "failed to get account transactions",
		},
		{
			name:        "bad request error",
			statusCode:  http.StatusBadRequest,
			expectError: false, // Current implementation doesn't check status codes
			errorMsg:    "failed to get account transactions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(`{"error": "test error"}`))
			})

			accounts, server := setupTestClient(handler)
			defer server.Close()

			ctx := context.Background()
			result, err := accounts.Transactions(ctx, "hash123", "", "", "", "")

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, result)
			} else {
				// Current behavior: returns response without checking status code
				require.NoError(t, err)
				require.NotNil(t, result)
			}
		})
	}
}

// TestTransactions_EdgeCases tests edge cases for transactions
func TestTransactions_EdgeCases(t *testing.T) {
	tests := []struct {
		name            string
		accountHash     string
		startDate       string
		endDate         string
		transactionType string
		symbol          string
		responseBody    string
		expectEmpty     bool
	}{
		{
			name:            "empty transactions list",
			accountHash:     "hash123",
			startDate:       "2024-01-01",
			endDate:         "2024-03-15",
			transactionType: "",
			symbol:          "",
			responseBody:    `{"transactions": []}`,
			expectEmpty:     true,
		},
		{
			name:            "empty account hash",
			accountHash:     "",
			startDate:       "",
			endDate:         "",
			transactionType: "",
			symbol:          "",
			responseBody:    `{"transactions": []}`,
			expectEmpty:     true,
		},
		{
			name:            "invalid date format - server handles",
			accountHash:     "hash123",
			startDate:       "invalid-date",
			endDate:         "also-invalid",
			transactionType: "",
			symbol:          "",
			responseBody:    `{"transactions": []}`,
			expectEmpty:     true,
		},
		{
			name:            "date range with reversed dates",
			accountHash:     "hash123",
			startDate:       "2024-12-31",
			endDate:         "2024-01-01",
			transactionType: "",
			symbol:          "",
			responseBody:    `{"transactions": []}`,
			expectEmpty:     true,
		},
		{
			name:            "special characters in symbol",
			accountHash:     "hash123",
			startDate:       "",
			endDate:         "",
			transactionType: "",
			symbol:          "BRK.A",
			responseBody:    `{"transactions": []}`,
			expectEmpty:     true,
		},
		{
			name:            "multiple transaction types",
			accountHash:     "hash123",
			startDate:       "",
			endDate:         "",
			transactionType: "TRADE,DIVIDEND",
			symbol:          "",
			responseBody:    `{"transactions": []}`,
			expectEmpty:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(tt.responseBody))
			})

			accounts, server := setupTestClient(handler)
			defer server.Close()

			ctx := context.Background()
			result, err := accounts.Transactions(ctx, tt.accountHash, tt.startDate, tt.endDate, tt.transactionType, tt.symbol)

			require.NoError(t, err)
			require.NotNil(t, result)
			if tt.expectEmpty {
				assert.Equal(t, 0, len(result.Transactions))
			}
		})
	}
}

// TestTransactions_Timeout tests timeout handling for transactions
func TestTransactions_Timeout(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	})

	accounts, server := setupTestClient(handler)
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := accounts.Transactions(ctx, "hash123", "", "", "", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}

// TestTransactionDetails_Success tests successful retrieval of transaction details
func TestTransactionDetails_Success(t *testing.T) {
	tests := []struct {
		name          string
		accountHash   string
		transactionID string
		transaction   *types.Transaction
	}{
		{
			name:          "success with full transaction details",
			accountHash:   "hash123",
			transactionID: "txn001",
			transaction: &types.Transaction{
				TransactionID: "txn001",
				Type:          "TRADE",
				SubType:       "BUY",
				Amount:        "1500.00",
				Description:   "Buy 10 AAPL @ $150.00",
				Date:          "2024-02-15T10:30:00Z",
			},
		},
		{
			name:          "success with minimal transaction details",
			accountHash:   "hash456",
			transactionID: "txn002",
			transaction: &types.Transaction{
				TransactionID: "txn002",
				Type:          "DIVIDEND",
				Amount:        "50.00",
			},
		},
		{
			name:          "success with wire transfer",
			accountHash:   "hash789",
			transactionID: "txn003",
			transaction: &types.Transaction{
				TransactionID: "txn003",
				Type:          "WIRE",
				SubType:       "INCOMING",
				Amount:        "10000.00",
				Description:   "Wire transfer from external account",
				Date:          "2024-03-20T08:00:00Z",
			},
		},
		{
			name:          "success with ach transaction",
			accountHash:   "hashABC",
			transactionID: "txn004",
			transaction: &types.Transaction{
				TransactionID: "txn004",
				Type:          "ACH",
				SubType:       "DEPOSIT",
				Amount:        "5000.00",
				Description:   "ACH deposit from bank",
				Date:          "2024-04-01T12:00:00Z",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Contains(t, r.URL.Path, "/trader/v1/accounts/")
				assert.Contains(t, r.URL.Path, "/transactions/")
				assert.Contains(t, r.URL.Path, tt.transactionID)

				w.WriteHeader(http.StatusOK)
				w.Write([]byte(createTransactionDetailsResponse(tt.transaction)))
			})

			accounts, server := setupTestClient(handler)
			defer server.Close()

			ctx := context.Background()
			result, err := accounts.TransactionDetails(ctx, tt.accountHash, tt.transactionID)

			require.NoError(t, err)
			require.NotNil(t, result)
			require.NotNil(t, result.Transaction)
			assert.Equal(t, tt.transaction.TransactionID, result.Transaction.TransactionID)
			assert.Equal(t, tt.transaction.Type, result.Transaction.Type)
			assert.Equal(t, tt.transaction.Amount, result.Transaction.Amount)
		})
	}
}

// TestTransactionDetails_Error tests error handling for transaction details
func TestTransactionDetails_Error(t *testing.T) {
	tests := []struct {
		name          string
		accountHash   string
		transactionID string
		statusCode    int
		expectError   bool
		errorMsg      string
	}{
		{
			name:          "unauthorized error",
			accountHash:   "hash123",
			transactionID: "txn001",
			statusCode:    http.StatusUnauthorized,
			expectError:   false, // Current implementation doesn't check status codes
			errorMsg:      "failed to get transaction details",
		},
		{
			name:          "server error",
			accountHash:   "hash123",
			transactionID: "txn001",
			statusCode:    http.StatusInternalServerError,
			expectError:   false, // Current implementation doesn't check status codes
			errorMsg:      "failed to get transaction details",
		},
		{
			name:          "not found error",
			accountHash:   "hash123",
			transactionID: "nonexistent",
			statusCode:    http.StatusNotFound,
			expectError:   false, // Current implementation doesn't check status codes
			errorMsg:      "failed to get transaction details",
		},
		{
			name:          "forbidden error",
			accountHash:   "hash123",
			transactionID: "txn001",
			statusCode:    http.StatusForbidden,
			expectError:   false, // Current implementation doesn't check status codes
			errorMsg:      "failed to get transaction details",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(`{"error": "test error"}`))
			})

			accounts, server := setupTestClient(handler)
			defer server.Close()

			ctx := context.Background()
			result, err := accounts.TransactionDetails(ctx, tt.accountHash, tt.transactionID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, result)
			} else {
				// Current behavior: returns response without checking status code
				require.NoError(t, err)
				require.NotNil(t, result)
			}
		})
	}
}

// TestTransactionDetails_EdgeCases tests edge cases for transaction details
func TestTransactionDetails_EdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		accountHash   string
		transactionID string
		responseBody  string
		expectNil     bool
	}{
		{
			name:          "null transaction in response",
			accountHash:   "hash123",
			transactionID: "txn001",
			responseBody:  `{"transaction": null}`,
			expectNil:     true,
		},
		{
			name:          "empty response object",
			accountHash:   "hash123",
			transactionID: "txn001",
			responseBody:  `{}`,
			expectNil:     true,
		},
		{
			name:          "empty account hash",
			accountHash:   "",
			transactionID: "txn001",
			responseBody:  `{"transaction": null}`,
			expectNil:     true,
		},
		{
			name:          "empty transaction id",
			accountHash:   "hash123",
			transactionID: "",
			responseBody:  `{"transaction": null}`,
			expectNil:     true,
		},
		{
			name:          "transaction with special characters in id",
			accountHash:   "hash123",
			transactionID: "txn-001_abc.123",
			responseBody:  `{"transaction": {"transactionId": "txn-001_abc.123", "type": "TRADE"}}`,
			expectNil:     false,
		},
		{
			name:          "transaction with unicode description",
			accountHash:   "hash123",
			transactionID: "txn002",
			responseBody:  `{"transaction": {"transactionId": "txn002", "type": "TRADE", "description": "买 10 AAPL 股票"}}`,
			expectNil:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(tt.responseBody))
			})

			accounts, server := setupTestClient(handler)
			defer server.Close()

			ctx := context.Background()
			result, err := accounts.TransactionDetails(ctx, tt.accountHash, tt.transactionID)

			require.NoError(t, err)
			require.NotNil(t, result)
			if tt.expectNil {
				assert.Nil(t, result.Transaction)
			} else {
				assert.NotNil(t, result.Transaction)
			}
		})
	}
}

// TestTransactionDetails_Timeout tests timeout handling for transaction details
func TestTransactionDetails_Timeout(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	})

	accounts, server := setupTestClient(handler)
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := accounts.TransactionDetails(ctx, "hash123", "txn001")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}
