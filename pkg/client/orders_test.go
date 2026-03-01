package client

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/citizenadam/go-schwabapi/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupOrdersTestClient creates a test orders client with mock server
func setupOrdersTestClient(handler http.Handler) (*OrdersClient, *httptest.Server) {
	server := httptest.NewServer(handler)
	logger := slog.Default()
	httpClient := NewClient(logger)
	tokenGetter := &mockTokenGetter{token: "test-token"}

	ordersClient := NewOrdersClient(httpClient, logger, tokenGetter)
	ordersClient.SetBaseURL(server.URL)

	return ordersClient, server
}

// createPreviewOrderResponse creates a JSON response for preview order
func createPreviewOrderResponse(order *types.Order) string {
	response := types.PreviewOrderResponse{
		Order: order,
	}
	jsonBytes, _ := json.Marshal(response)
	return string(jsonBytes)
}

// TestPreviewOrder_Success tests successful order preview
func TestPreviewOrder_Success(t *testing.T) {
	tests := []struct {
		name        string
		accountHash string
		order       any
		expected    *types.Order
	}{
		{
			name:        "success with limit order",
			accountHash: "test-hash-123",
			order: &types.Order{
				Session:     "NORMAL",
				Duration:    "DAY",
				OrderType:   "LIMIT",
				Quantity:    100,
				Price:       "150.00",
				Symbol:      "AAPL",
				Instruction: "BUY",
			},
			expected: &types.Order{
				Session:     "NORMAL",
				Duration:    "DAY",
				OrderType:   "LIMIT",
				Quantity:    100,
				Price:       "150.00",
				Symbol:      "AAPL",
				Instruction: "BUY",
			},
		},
		{
			name:        "success with market order",
			accountHash: "test-hash-456",
			order: &types.Order{
				Session:     "NORMAL",
				Duration:    "GTC",
				OrderType:   "MARKET",
				Quantity:    50,
				Symbol:      "GOOGL",
				Instruction: "SELL",
			},
			expected: &types.Order{
				Session:     "NORMAL",
				Duration:    "GTC",
				OrderType:   "MARKET",
				Quantity:    50,
				Symbol:      "GOOGL",
				Instruction: "SELL",
			},
		},
		{
			name:        "success with stop limit order",
			accountHash: "test-hash-789",
			order: &types.Order{
				Session:     "NORMAL",
				Duration:    "DAY",
				OrderType:   "STOP_LIMIT",
				Quantity:    200,
				Price:       "145.00",
				StopPrice:   "140.00",
				Symbol:      "MSFT",
				Instruction: "SELL",
			},
			expected: &types.Order{
				Session:     "NORMAL",
				Duration:    "DAY",
				OrderType:   "STOP_LIMIT",
				Quantity:    200,
				Price:       "145.00",
				StopPrice:   "140.00",
				Symbol:      "MSFT",
				Instruction: "SELL",
			},
		},
		{
			name:        "success with option order",
			accountHash: "option-hash-001",
			order: &types.Order{
				Session:         "NORMAL",
				Duration:        "DAY",
				OrderType:       "LIMIT",
				Quantity:        5,
				Price:           "2.50",
				Symbol:          "AAPL_031524C150",
				SymbolAssetType: "OPTION",
				Instruction:     "BUY",
			},
			expected: &types.Order{
				Session:         "NORMAL",
				Duration:        "DAY",
				OrderType:       "LIMIT",
				Quantity:        5,
				Price:           "2.50",
				Symbol:          "AAPL_031524C150",
				SymbolAssetType: "OPTION",
				Instruction:     "BUY",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request method
				assert.Equal(t, http.MethodPost, r.Method)

				// Verify URL path contains account hash
				assert.Contains(t, r.URL.Path, tt.accountHash)
				assert.Contains(t, r.URL.Path, "/orders/validate")

				// Verify headers
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				assert.Equal(t, "application/json", r.Header.Get("Accept"))

				// Read and verify request body
				body, err := io.ReadAll(r.Body)
				require.NoError(t, err)

				var receivedOrder types.Order
				err = json.Unmarshal(body, &receivedOrder)
				require.NoError(t, err)

				// Write response
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(createPreviewOrderResponse(tt.expected)))
			})

			ordersClient, server := setupOrdersTestClient(handler)
			defer server.Close()

			ctx := context.Background()
			resp, err := ordersClient.PreviewOrder(ctx, tt.accountHash, tt.order)

			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Equal(t, http.StatusOK, resp.StatusCode)

			// Read and verify response body
			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			resp.Body.Close()

			var previewResp types.PreviewOrderResponse
			err = json.Unmarshal(body, &previewResp)
			require.NoError(t, err)
			require.NotNil(t, previewResp.Order)
			assert.Equal(t, tt.expected.Symbol, previewResp.Order.Symbol)
			assert.Equal(t, tt.expected.Quantity, previewResp.Order.Quantity)
			assert.Equal(t, tt.expected.OrderType, previewResp.Order.OrderType)
		})
	}
}

// TestPreviewOrder_Error tests error handling for preview order
func TestPreviewOrder_Error(t *testing.T) {
	tests := []struct {
		name         string
		accountHash  string
		order        any
		statusCode   int
		responseBody string
		expectError  bool
		errorMsg     string
	}{
		{
			name:        "unauthorized - 401",
			accountHash: "test-hash",
			order: &types.Order{
				Symbol:    "AAPL",
				OrderType: "LIMIT",
				Quantity:  100,
			},
			statusCode:   http.StatusUnauthorized,
			responseBody: `{"error": "Invalid or expired token"}`,
			expectError:  false, // Current implementation returns response without checking status
		},
		{
			name:        "internal server error - 500",
			accountHash: "test-hash",
			order: &types.Order{
				Symbol:    "AAPL",
				OrderType: "LIMIT",
				Quantity:  100,
			},
			statusCode:   http.StatusInternalServerError,
			responseBody: `{"error": "Internal server error"}`,
			expectError:  false, // Current implementation returns response without checking status
		},
		{
			name:        "bad request - 400 - invalid order",
			accountHash: "test-hash",
			order: &types.Order{
				Symbol:    "INVALID_SYMBOL",
				OrderType: "INVALID_TYPE",
				Quantity:  -100,
			},
			statusCode:   http.StatusBadRequest,
			responseBody: `{"error": "Invalid order parameters"}`,
			expectError:  false, // Current implementation returns response without checking status
		},
		{
			name:        "forbidden - 403",
			accountHash: "test-hash",
			order: &types.Order{
				Symbol:    "AAPL",
				OrderType: "LIMIT",
				Quantity:  100,
			},
			statusCode:   http.StatusForbidden,
			responseBody: `{"error": "Access denied"}`,
			expectError:  false, // Current implementation returns response without checking status
		},
		{
			name:        "not found - 404",
			accountHash: "non-existent-hash",
			order: &types.Order{
				Symbol:    "AAPL",
				OrderType: "LIMIT",
				Quantity:  100,
			},
			statusCode:   http.StatusNotFound,
			responseBody: `{"error": "Account not found"}`,
			expectError:  false, // Current implementation returns response without checking status
		},
		{
			name:        "too many requests - 429",
			accountHash: "test-hash",
			order: &types.Order{
				Symbol:    "AAPL",
				OrderType: "LIMIT",
				Quantity:  100,
			},
			statusCode:   http.StatusTooManyRequests,
			responseBody: `{"error": "Rate limit exceeded"}`,
			expectError:  false, // Current implementation returns response without checking status
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.responseBody))
			})

			ordersClient, server := setupOrdersTestClient(handler)
			defer server.Close()

			ctx := context.Background()
			resp, err := ordersClient.PreviewOrder(ctx, tt.accountHash, tt.order)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, resp)
			} else {
				// Current behavior: returns response without checking status code
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Equal(t, tt.statusCode, resp.StatusCode)
				resp.Body.Close()
			}
		})
	}
}

// TestPreviewOrder_EdgeCases tests edge cases for preview order
func TestPreviewOrder_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		accountHash string
		order       any
		setupMock   func(w http.ResponseWriter, r *http.Request)
		wantErr     bool
		errContains string
	}{
		{
			name:        "empty account hash",
			accountHash: "",
			order: &types.Order{
				Symbol:    "AAPL",
				OrderType: "LIMIT",
				Quantity:  100,
			},
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"error": "Account hash required"}`))
			},
			wantErr: false, // Returns response without checking status
		},
		{
			name:        "nil order",
			accountHash: "test-hash",
			order:       nil,
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{}`))
			},
			wantErr: false,
		},
		{
			name:        "empty order",
			accountHash: "test-hash",
			order:       &types.Order{},
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(createPreviewOrderResponse(&types.Order{})))
			},
			wantErr: false,
		},
		{
			name:        "complex order with all fields",
			accountHash: "test-hash",
			order: &types.Order{
				Session:          "NORMAL",
				Duration:         "GTC",
				OrderType:        "LIMIT",
				ComplexOrderType: "SINGLE",
				Quantity:         1000,
				Price:            "999.99",
				StopPrice:        "950.00",
				Symbol:           "AAPL",
				SymbolAssetType:  "EQUITY",
				Instruction:      "BUY",
			},
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(createPreviewOrderResponse(&types.Order{
					Session:          "NORMAL",
					Duration:         "GTC",
					OrderType:        "LIMIT",
					ComplexOrderType: "SINGLE",
					Quantity:         1000,
					Price:            "999.99",
					StopPrice:        "950.00",
					Symbol:           "AAPL",
					SymbolAssetType:  "EQUITY",
					Instruction:      "BUY",
				})))
			},
			wantErr: false,
		},
		{
			name:        "large quantity order",
			accountHash: "test-hash",
			order: &types.Order{
				Symbol:    "AAPL",
				OrderType: "MARKET",
				Quantity:  1000000,
			},
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(createPreviewOrderResponse(&types.Order{
					Symbol:    "AAPL",
					OrderType: "MARKET",
					Quantity:  1000000,
				})))
			},
			wantErr: false,
		},
		{
			name:        "fractional price",
			accountHash: "test-hash",
			order: &types.Order{
				Symbol:    "AAPL",
				OrderType: "LIMIT",
				Quantity:  100,
				Price:     "150.123456",
			},
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(createPreviewOrderResponse(&types.Order{
					Symbol:    "AAPL",
					OrderType: "LIMIT",
					Quantity:  100,
					Price:     "150.123456",
				})))
			},
			wantErr: false,
		},
		{
			name:        "special characters in symbol",
			accountHash: "test-hash",
			order: &types.Order{
				Symbol:    "BRK.B",
				OrderType: "LIMIT",
				Quantity:  50,
				Price:     "350.00",
			},
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(createPreviewOrderResponse(&types.Order{
					Symbol:    "BRK.B",
					OrderType: "LIMIT",
					Quantity:  50,
					Price:     "350.00",
				})))
			},
			wantErr: false,
		},
		{
			name:        "after hours session",
			accountHash: "test-hash",
			order: &types.Order{
				Session:     "AM",
				Duration:    "DAY",
				OrderType:   "LIMIT",
				Quantity:    100,
				Price:       "150.00",
				Symbol:      "AAPL",
				Instruction: "BUY",
			},
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(createPreviewOrderResponse(&types.Order{
					Session:     "AM",
					Duration:    "DAY",
					OrderType:   "LIMIT",
					Quantity:    100,
					Price:       "150.00",
					Symbol:      "AAPL",
					Instruction: "BUY",
				})))
			},
			wantErr: false,
		},
		{
			name:        "malformed JSON response",
			accountHash: "test-hash",
			order: &types.Order{
				Symbol:    "AAPL",
				OrderType: "LIMIT",
				Quantity:  100,
			},
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{invalid json}`))
			},
			wantErr: false, // Response is returned, caller handles parsing
		},
		{
			name:        "empty response body",
			accountHash: "test-hash",
			order: &types.Order{
				Symbol:    "AAPL",
				OrderType: "LIMIT",
				Quantity:  100,
			},
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte{})
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(tt.setupMock)

			ordersClient, server := setupOrdersTestClient(handler)
			defer server.Close()

			ctx := context.Background()
			resp, err := ordersClient.PreviewOrder(ctx, tt.accountHash, tt.order)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				resp.Body.Close()
			}
		})
	}
}

// TestPreviewOrder_IntegrationWithOrders tests PreviewOrder integration with other order methods
func TestPreviewOrder_IntegrationWithOrders(t *testing.T) {
	t.Run("preview then place order flow", func(t *testing.T) {
		order := &types.Order{
			Session:     "NORMAL",
			Duration:    "DAY",
			OrderType:   "LIMIT",
			Quantity:    100,
			Price:       "150.00",
			Symbol:      "AAPL",
			Instruction: "BUY",
		}
		accountHash := "test-hash-123"

		callCount := 0
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++

			if r.Method == http.MethodPost {
				if r.URL.Path == "/v1/accounts/"+accountHash+"/orders/validate" {
					// Preview order request
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(createPreviewOrderResponse(order)))
				} else if r.URL.Path == "/v1/accounts/"+accountHash+"/orders" {
					// Place order request
					w.WriteHeader(http.StatusCreated)
					w.Write([]byte(`{"orderId": "order-123"}`))
				}
			}
		})

		ordersClient, server := setupOrdersTestClient(handler)
		defer server.Close()

		ctx := context.Background()

		// Step 1: Preview the order
		previewResp, err := ordersClient.PreviewOrder(ctx, accountHash, order)
		require.NoError(t, err)
		require.NotNil(t, previewResp)
		assert.Equal(t, http.StatusOK, previewResp.StatusCode)
		previewResp.Body.Close()

		// Step 2: Place the order (simulating the flow)
		placeResp, err := ordersClient.PlaceOrder(ctx, accountHash, order)
		require.NoError(t, err)
		require.NotNil(t, placeResp)
		assert.Equal(t, http.StatusCreated, placeResp.StatusCode)
		placeResp.Body.Close()

		// Verify both calls were made
		assert.Equal(t, 2, callCount)
	})

	t.Run("preview order with different order types", func(t *testing.T) {
		orderTypes := []struct {
			name      string
			orderType string
			order     *types.Order
		}{
			{
				name:      "market order",
				orderType: "MARKET",
				order: &types.Order{
					Session:     "NORMAL",
					Duration:    "DAY",
					OrderType:   "MARKET",
					Quantity:    100,
					Symbol:      "AAPL",
					Instruction: "BUY",
				},
			},
			{
				name:      "limit order",
				orderType: "LIMIT",
				order: &types.Order{
					Session:     "NORMAL",
					Duration:    "GTC",
					OrderType:   "LIMIT",
					Quantity:    50,
					Price:       "150.00",
					Symbol:      "AAPL",
					Instruction: "BUY",
				},
			},
			{
				name:      "stop order",
				orderType: "STOP",
				order: &types.Order{
					Session:     "NORMAL",
					Duration:    "DAY",
					OrderType:   "STOP",
					Quantity:    100,
					StopPrice:   "140.00",
					Symbol:      "AAPL",
					Instruction: "SELL",
				},
			},
			{
				name:      "stop limit order",
				orderType: "STOP_LIMIT",
				order: &types.Order{
					Session:     "NORMAL",
					Duration:    "GTC",
					OrderType:   "STOP_LIMIT",
					Quantity:    100,
					Price:       "145.00",
					StopPrice:   "140.00",
					Symbol:      "AAPL",
					Instruction: "SELL",
				},
			},
			{
				name:      "trailing stop order",
				orderType: "TRAILING_STOP",
				order: &types.Order{
					Session:     "NORMAL",
					Duration:    "DAY",
					OrderType:   "TRAILING_STOP",
					Quantity:    100,
					StopPrice:   "5.00", // Trail amount
					Symbol:      "AAPL",
					Instruction: "SELL",
				},
			},
		}

		for _, ot := range orderTypes {
			t.Run(ot.name, func(t *testing.T) {
				handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, http.MethodPost, r.Method)
					assert.Contains(t, r.URL.Path, "/orders/validate")

					w.WriteHeader(http.StatusOK)
					w.Write([]byte(createPreviewOrderResponse(ot.order)))
				})

				ordersClient, server := setupOrdersTestClient(handler)
				defer server.Close()

				ctx := context.Background()
				resp, err := ordersClient.PreviewOrder(ctx, "test-hash", ot.order)

				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Equal(t, http.StatusOK, resp.StatusCode)
				resp.Body.Close()
			})
		}
	})

	t.Run("preview order with different durations", func(t *testing.T) {
		durations := []struct {
			name     string
			duration string
		}{
			{name: "DAY", duration: "DAY"},
			{name: "GTC", duration: "GTC"},
			{name: "GTD", duration: "GTD"},
			{name: "OPG", duration: "OPG"},
			{name: "CLO", duration: "CLO"},
		}

		for _, d := range durations {
			t.Run(d.name, func(t *testing.T) {
				order := &types.Order{
					Session:     "NORMAL",
					Duration:    d.duration,
					OrderType:   "LIMIT",
					Quantity:    100,
					Price:       "150.00",
					Symbol:      "AAPL",
					Instruction: "BUY",
				}

				handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(createPreviewOrderResponse(order)))
				})

				ordersClient, server := setupOrdersTestClient(handler)
				defer server.Close()

				ctx := context.Background()
				resp, err := ordersClient.PreviewOrder(ctx, "test-hash", order)

				require.NoError(t, err)
				require.NotNil(t, resp)
				resp.Body.Close()
			})
		}
	})

	t.Run("preview order with different instructions", func(t *testing.T) {
		instructions := []struct {
			name        string
			instruction string
		}{
			{name: "BUY", instruction: "BUY"},
			{name: "SELL", instruction: "SELL"},
			{name: "BUY_TO_COVER", instruction: "BUY_TO_COVER"},
			{name: "SELL_SHORT", instruction: "SELL_SHORT"},
		}

		for _, i := range instructions {
			t.Run(i.name, func(t *testing.T) {
				order := &types.Order{
					Session:     "NORMAL",
					Duration:    "DAY",
					OrderType:   "LIMIT",
					Quantity:    100,
					Price:       "150.00",
					Symbol:      "AAPL",
					Instruction: i.instruction,
				}

				handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(createPreviewOrderResponse(order)))
				})

				ordersClient, server := setupOrdersTestClient(handler)
				defer server.Close()

				ctx := context.Background()
				resp, err := ordersClient.PreviewOrder(ctx, "test-hash", order)

				require.NoError(t, err)
				require.NotNil(t, resp)
				resp.Body.Close()
			})
		}
	})
}

// TestPreviewOrder_Timeout tests timeout handling for preview order
func TestPreviewOrder_Timeout(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	})

	ordersClient, server := setupOrdersTestClient(handler)
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	order := &types.Order{
		Symbol:    "AAPL",
		OrderType: "LIMIT",
		Quantity:  100,
	}

	_, err := ordersClient.PreviewOrder(ctx, "test-hash", order)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}

// TestPreviewOrder_ContextCancellation tests context cancellation handling
func TestPreviewOrder_ContextCancellation(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	})

	ordersClient, server := setupOrdersTestClient(handler)
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel context immediately
	cancel()

	order := &types.Order{
		Symbol:    "AAPL",
		OrderType: "LIMIT",
		Quantity:  100,
	}

	_, err := ordersClient.PreviewOrder(ctx, "test-hash", order)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
}

// TestPreviewOrder_InvalidRequestBody tests handling of invalid request body
func TestPreviewOrder_InvalidRequestBody(t *testing.T) {
	// Test with a type that can't be marshaled to JSON
	invalidOrder := make(chan int)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	})

	ordersClient, server := setupOrdersTestClient(handler)
	defer server.Close()

	ctx := context.Background()
	_, err := ordersClient.PreviewOrder(ctx, "test-hash", invalidOrder)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to preview order")
}
