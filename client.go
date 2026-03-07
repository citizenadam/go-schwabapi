// Package schwabdev provides a Go client for the Charles Schwab API.
// It supports OAuth token management, REST API access for accounts,
// orders, and market data, and real-time WebSocket streaming.
//
// This package is not affiliated with or endorsed by Schwab.
// Licensed under the MIT license and acts in accordance with Schwab's API terms and conditions.
//
// Basic usage:
//
//	client, err := schwabdev.NewClient(appKey, appSecret, callbackURL, "", "", 0, nil)
//	streamer := stream.NewStreamer(logger, client.TokenManager(), infoSrc)
package schwabdev

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client is the main client for interacting with the Schwab API.
// It manages authentication, HTTP requests, and token lifecycle.
type Client struct {
	tokenManager *TokenManager
	httpClient   *http.Client
	baseURL      string
	logger       *slog.Logger
	timeout      time.Duration
}

// NewClient creates a new Client instance for accessing the Schwab API.
//
// Parameters:
//   - appKey: Schwab API application key (32 or 48 characters)
//   - appSecret: Schwab API application secret (16 or 64 characters)
//   - callbackURL: OAuth callback URL (must be HTTPS, cannot end with /)
//   - storagePath: Path to token JSON file (default: ~/.schwabdev/tokens.json)
//   - encryption: Optional Fernet encryption key for token storage
//   - timeout: HTTP request timeout (use 0 for DefaultHTTPRequestTimeout)
//   - callOnAuth: Optional callback — receives auth URL, returns callback URL after
//     the user completes the OAuth flow. Pass nil to fall back to stdin prompt.
//
// Returns *Client and error if validation or initialization fails.
func NewClient(appKey, appSecret, callbackURL, storagePath, encryption string, timeout time.Duration, callOnAuth func(authURL string) (string, error)) (*Client, error) {
	// Validate timeout
	if timeout <= 0 {
		timeout = DefaultHTTPRequestTimeout
	}

	// Create logger
	logger := slog.Default()

	// Create TokenManager backed by file storage.
	tokenManager, err := NewTokenManagerWithFilePath(appKey, appSecret, callbackURL, storagePath, encryption, logger, callOnAuth)
	if err != nil {
		return nil, err
	}

	// Create HTTP client with timeout
	httpClient := &http.Client{
		Timeout: timeout,
	}

	// Create Client instance
	client := &Client{
		tokenManager: tokenManager,
		httpClient:   httpClient,
		baseURL:      "https://api.schwabapi.com",
		logger:       logger,
		timeout:      timeout,
	}

	// Ensure tokens are up to date on init
	if _, err := tokenManager.UpdateTokens(false, false); err != nil {
		// Log warning but don't fail - tokens might not exist yet for first-time setup
		logger.Debug("Could not update tokens during initialization", "error", err)
	}

	return client, nil
}

// Close closes the client and releases resources.
// It closes the TokenManager's database connection and cleans up HTTP client idle connections.
// Implements the io.Closer interface.
func (c *Client) Close() error {
	var errs []error

	if c.tokenManager != nil {
		if err := c.tokenManager.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if c.httpClient != nil {
		if transport, ok := c.httpClient.Transport.(*http.Transport); ok {
			transport.CloseIdleConnections()
		}
	}

	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}

// UpdateTokens updates the access and refresh tokens if needed.
// Set forceAccessToken or forceRefreshToken to true to force an update.
// Returns true if tokens were updated, false otherwise.
func (c *Client) UpdateTokens(forceAccessToken, forceRefreshToken bool) (bool, error) {
	return c.tokenManager.UpdateTokens(forceAccessToken, forceRefreshToken)
}

// TokenManager returns the underlying TokenManager, which satisfies the
// stream.TokenProvider interface. Use this to wire the streamer:
//
//	streamer := stream.NewStreamer(logger, client.TokenManager(), infoSrc)
func (c *Client) TokenManager() *TokenManager {
	return c.tokenManager
}

// ensureValidToken is kept for compatibility but authHeader now calls
// tokenManager.AccessToken() directly, which handles refresh internally.
func (c *Client) ensureValidToken(_ context.Context) error {
	_, err := c.tokenManager.UpdateTokens(false, false)
	return err
}

// authHeader returns the Authorization header value with Bearer token.
// Calls ensureValidToken to refresh if needed before returning the header.
// Returns the header string in format "Bearer {access_token}" or an error.
func (c *Client) authHeader(ctx context.Context) (string, error) {
	if err := c.ensureValidToken(ctx); err != nil {
		return "", fmt.Errorf("failed to ensure valid token: %w", err)
	}

	accessToken, err := c.tokenManager.AccessToken()
	if err != nil {
		return "", fmt.Errorf("get access token: %w", err)
	}
	if accessToken == "" {
		return "", fmt.Errorf("access token is empty")
	}

	return fmt.Sprintf("Bearer %s", accessToken), nil
}

// request is a private method that makes HTTP requests to the Schwab API.
// It handles token updates, authorization headers, request body marshaling,
// and response parsing.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - method: HTTP method (GET, POST, PUT, DELETE, etc.)
//   - path: API path (e.g., "/trader/v1/accounts")
//   - body: Request body (will be marshaled to JSON, can be nil)
//   - result: Response body destination (will be unmarshaled from JSON, can be nil)
//
// Returns the HTTP response and any error that occurred.
func (c *Client) request(ctx context.Context, method, path string, body, result any) (*http.Response, error) {
	return c.doRequest(ctx, method, path, body, result, false)
}

// doRequest executes the HTTP request with optional retry on 401 Unauthorized.
func (c *Client) doRequest(ctx context.Context, method, path string, body, result any, isRetry bool) (*http.Response, error) {
	authHeader, err := c.authHeader(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get auth header: %w", err)
	}

	fullURL := c.baseURL + path

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", authHeader)

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized && !isRetry {
		resp.Body.Close()

		if c.logger != nil {
			c.logger.Debug("Received 401 Unauthorized, forcing token refresh and retrying")
		}

		if _, err := c.tokenManager.UpdateTokens(true, false); err != nil {
			return nil, fmt.Errorf("failed to refresh token after 401: %w", err)
		}

		return c.doRequest(ctx, method, path, body, result, true)
	}

	if result != nil && resp.Body != nil {
		defer resp.Body.Close()

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}

		resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))

		if len(bodyBytes) > 0 {
			if err := json.Unmarshal(bodyBytes, result); err != nil {
				c.logger.Debug("Failed to unmarshal response body", "error", err, "status", resp.StatusCode)
			}
		}
	}

	return resp, nil
}

// parseParams filters out nil values from params and converts to url.Values.
// This matches Python's _parse_params() behavior which removes None values.
//
// Parameters:
//   - params: Map of parameter names to values (values can be nil)
//
// Returns url.Values containing only non-nil parameters converted to strings.
func (c *Client) parseParams(params map[string]any) url.Values {
	result := url.Values{}
	for key, value := range params {
		if value != nil {
			result.Set(key, fmt.Sprintf("%v", value))
		}
	}
	return result
}

// formatList converts a list to a comma-separated string.
// This matches Python's _format_list() behavior exactly:
//   - Returns empty string if list is nil
//   - Returns input as-is if it's already a string
//   - Joins with commas if it's a []string
//
// Parameters:
//   - list: Can be nil, string, or []string
//
// Returns formatted string or empty string for nil.
func (c *Client) formatList(list any) string {
	if list == nil {
		return ""
	}

	switch v := list.(type) {
	case string:
		return v
	case []string:
		if len(v) == 0 {
			return ""
		}
		var result strings.Builder
		for i, s := range v {
			if i > 0 {
				result.WriteString(",")
			}
			result.WriteString(s)
		}
		return result.String()
	default:
		// For other types, convert to string using fmt.Sprintf
		return fmt.Sprintf("%v", v)
	}
}

// LinkedAccounts retrieves all linked account numbers and their hash values.
// Account numbers in plain text cannot be used outside of headers or request/response bodies.
// Use the encrypted hash values from this response for all subsequent account-specific API calls.
//
// Returns a slice of LinkedAccount containing accountNumber and hashValue pairs.
// Returns error if the request fails.
func (c *Client) LinkedAccounts(ctx context.Context) (*LinkedAccountsResponse, error) {
	var result LinkedAccountsResponse
	_, err := c.request(ctx, "GET", "/trader/v1/accounts/accountNumbers", nil, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get linked accounts: %w", err)
	}
	return &result, nil
}

// AccountDetailsAll fetches all linked account information for the authenticated user.
// By default, balances are returned. Use fields="positions" to include positions.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - fields: Optional fields to return (can be nil, e.g., "positions")
//
// Returns a pointer to AccountDetailsAllResponse containing account details and aggregated balances.
func (c *Client) AccountDetailsAll(ctx context.Context, fields *string) ([]AccountDetailsAllResponse, error) {
	path := "/trader/v1/accounts/"

	if fields != nil {
		params := c.parseParams(map[string]any{"fields": *fields})
		path = path + "?" + params.Encode()
	}

	// Change result to a slice to match the JSON array returned by Schwab
	var result []AccountDetailsAllResponse
	_, err := c.request(ctx, "GET", path, nil, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get account details for all accounts: %w", err)
	}

	return result, nil
}

// AccountDetails retrieves specific account information with balances and positions.
// The balance information on these accounts is displayed by default but positions
// will be returned based on the "positions" flag.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - accountHash: Account hash from LinkedAccounts()
//   - fields: Optional fields to return (can be nil)
//
// Returns a pointer to AccountDetailsResponse and any error that occurred.
func (c *Client) AccountDetails(ctx context.Context, accountHash string, fields *string) (*AccountDetailsResponse, error) {
	path := fmt.Sprintf("/trader/v1/accounts/%s", accountHash)

	if fields != nil {
		params := c.parseParams(map[string]any{"fields": *fields})
		path = path + "?" + params.Encode()
	}

	var result AccountDetailsResponse
	_, err := c.request(ctx, "GET", path, nil, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get account details: %w", err)
	}

	return &result, nil
}

// GetStreamerInfo fetches WebSocket streaming connection details.
// It retrieves the streamer URL and credentials needed to establish a WebSocket connection.
//
// Returns a pointer to StreamerInfo struct containing:
//   - StreamerURL: WebSocket URL for streaming data
//   - SchwabClientCorrelID: Client correlation ID for authentication
//   - SchwabClientChannel: Channel identifier
//   - SchwabClientFunctionID: Function identifier
//
// Returns error if the request fails or no streamer info is available.
func (c *Client) GetStreamerInfo(ctx context.Context) (*StreamerInfo, error) {
	var prefs PreferencesResponse

	_, err := c.request(ctx, "GET", "/trader/v1/userPreference", nil, &prefs)
	if err != nil {
		return nil, fmt.Errorf("failed to get user preferences: %w", err)
	}

	if len(prefs.StreamerInfo) == 0 {
		return nil, fmt.Errorf("no streamer info available")
	}

	return prefs.StreamerInfo[0], nil
}

// timeConvert converts a time value to the specified format.
// It handles time.Time, string (ISO 8601 date/datetime), and nil inputs.
// Returns the converted value as string or int64, or nil if input is nil.
//
// Parameters:
//   - dt: The time value to convert (time.Time, string, or nil)
//   - format: The output format (TimeFormatISO8601, TimeFormatEPOCH, TimeFormatEPOCHMS, TimeFormatYYYYMMDD)
//
// Returns the converted value and any error that occurred.
func (c *Client) timeConvert(dt any, format TimeFormat) (any, error) {
	// Handle nil input - return nil (passthrough)
	if dt == nil {
		return nil, nil
	}

	var t time.Time
	var err error

	// Parse input based on type
	switch v := dt.(type) {
	case time.Time:
		t = v
	case string:
		// Try parsing as datetime first (ISO 8601 with time)
		t, err = time.Parse(time.RFC3339, v)
		if err != nil {
			// Try parsing as date only (YYYY-MM-DD)
			t, err = time.Parse("2006-01-02", v)
			if err != nil {
				return nil, fmt.Errorf("failed to parse time string: %w", err)
			}
		}
	default:
		// Passthrough for non-datetime types (matches Python behavior)
		return dt, nil
	}

	// Convert to specified format
	switch format {
	case TimeFormatISO8601:
		// Format: YYYY-MM-DDTHH:MM:SS.sssZ (matches Python's isoformat with Z suffix)
		// Python: dt.isoformat().split('+')[0][:-3] + "Z"
		// This removes timezone info and truncates to milliseconds
		return t.UTC().Format("2006-01-02T15:04:05.000") + "Z", nil
	case TimeFormatEPOCH:
		// Unix timestamp in seconds
		return t.Unix(), nil
	case TimeFormatEPOCHMS:
		// Unix timestamp in milliseconds
		return t.UnixMilli(), nil
	case TimeFormatYYYYMMDD:
		// Date only: YYYY-MM-DD
		return t.Format("2006-01-02"), nil
	default:
		return nil, fmt.Errorf("unsupported time format: %s", format)
	}
}

// AccountOrders retrieves all orders for a specific account.
// Orders can be filtered based on input parameters. Maximum date range is 1 year.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - accountHash: Account hash from LinkedAccounts()
//   - fromEnteredTime: Start date (time.Time, string in ISO 8601 format, or nil)
//   - toEnteredTime: End date (time.Time, string in ISO 8601 format, or nil)
//   - maxResults: Maximum number of results (nil for default 3000)
//   - status: Order status filter (nil for all statuses)
//
// Returns AccountOrdersResponse containing orders for the account.
func (c *Client) AccountOrders(ctx context.Context, accountHash string, fromEnteredTime, toEnteredTime any, maxResults *int, status *string) (*AccountOrdersResponse, error) {
	fromTime, err := c.timeConvert(fromEnteredTime, TimeFormatISO8601)
	if err != nil {
		return nil, fmt.Errorf("failed to convert fromEnteredTime: %w", err)
	}

	toTime, err := c.timeConvert(toEnteredTime, TimeFormatISO8601)
	if err != nil {
		return nil, fmt.Errorf("failed to convert toEnteredTime: %w", err)
	}

	params := c.parseParams(map[string]any{
		"fromEnteredTime": fromTime,
		"toEnteredTime":   toTime,
		"maxResults":      maxResults,
		"status":          status,
	})

	path := fmt.Sprintf("/trader/v1/accounts/%s/orders", accountHash)
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	var result AccountOrdersResponse
	_, err = c.request(ctx, "GET", path, nil, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get account orders: %w", err)
	}

	return &result, nil
}

// PlaceOrder places an order for a specific account.
// The order ID is returned in the Location header of the response.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - accountHash: Account hash from LinkedAccounts()
//   - order: Order request details
//
// Returns PlaceOrderResponse containing the order ID and any error that occurred.
func (c *Client) PlaceOrder(ctx context.Context, accountHash string, order *OrderRequest) (*PlaceOrderResponse, error) {
	path := fmt.Sprintf("/trader/v1/accounts/%s/orders", accountHash)

	resp, err := c.request(ctx, "POST", path, order, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to place order: %w", err)
	}

	location := resp.Header.Get("Location")
	if location == "" {
		return nil, fmt.Errorf("no Location header in response")
	}

	parts := strings.Split(location, "/")
	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid Location header format")
	}

	orderID := parts[len(parts)-1]

	return &PlaceOrderResponse{
		OrderID: orderID,
	}, nil
}

// OrderDetails retrieves a specific order by its ID for a specific account.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - accountHash: Account hash from LinkedAccounts()
//   - orderID: Order ID to retrieve
//
// Returns OrderDetailsResponse containing order details.
// Returns error if the request fails.
func (c *Client) OrderDetails(ctx context.Context, accountHash string, orderID any) (*OrderDetailsResponse, error) {
	var result OrderDetailsResponse
	_, err := c.request(ctx, "GET", fmt.Sprintf("/trader/v1/accounts/%s/orders/%v", accountHash, orderID), nil, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get order details: %w", err)
	}
	return &result, nil
}

// CancelOrder cancels a specific order by its ID for a specific account.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - accountHash: Account hash from LinkedAccounts()
//   - orderID: Order ID to cancel
//
// Returns CancelOrderResponse on success.
// Returns error if the request fails.
func (c *Client) CancelOrder(ctx context.Context, accountHash string, orderID any) (*CancelOrderResponse, error) {
	var result CancelOrderResponse
	_, err := c.request(ctx, "DELETE", fmt.Sprintf("/trader/v1/accounts/%s/orders/%v", accountHash, orderID), nil, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to cancel order: %w", err)
	}
	return &result, nil
}

// ReplaceOrder replaces an existing order for an account.
// The existing order will be replaced by the new order. Once replaced, the old order will be canceled and a new order will be created.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - accountHash: Account hash from LinkedAccounts()
//   - orderID: Order ID to replace
//   - order: OrderRequest object containing the new order details
//
// Returns ReplaceOrderResponse on success.
// Returns error if the request fails.
func (c *Client) ReplaceOrder(ctx context.Context, accountHash string, orderID any, order *OrderRequest) (*ReplaceOrderResponse, error) {
	var result ReplaceOrderResponse
	_, err := c.request(ctx, "PUT", fmt.Sprintf("/trader/v1/accounts/%s/orders/%v", accountHash, orderID), order, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to replace order: %w", err)
	}
	return &result, nil
}

// AccountOrdersAll retrieves all orders for all accounts.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - fromEnteredTime: Start date (time.Time, string in ISO 8601 format, or nil)
//   - toEnteredTime: End date (time.Time, string in ISO 8601 format, or nil)
//   - maxResults: Maximum number of results (optional, can be nil for default 3000)
//   - status: Order status filter (optional, can be nil)
//
// Returns AccountOrdersAllResponse containing all orders.
// Returns error if the request fails.
func (c *Client) AccountOrdersAll(ctx context.Context, fromEnteredTime, toEnteredTime any, maxResults *int, status *string) (*AccountOrdersAllResponse, error) {
	// Convert time parameters
	fromTime, err := c.timeConvert(fromEnteredTime, TimeFormatISO8601)
	if err != nil {
		return nil, fmt.Errorf("failed to convert fromEnteredTime: %w", err)
	}

	toTime, err := c.timeConvert(toEnteredTime, TimeFormatISO8601)
	if err != nil {
		return nil, fmt.Errorf("failed to convert toEnteredTime: %w", err)
	}

	// Build query parameters
	params := c.parseParams(map[string]any{
		"fromEnteredTime": fromTime,
		"toEnteredTime":   toTime,
		"maxResults":      maxResults,
		"status":          status,
	})

	path := "/trader/v1/orders"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	var result AccountOrdersAllResponse
	_, err = c.request(ctx, "GET", path, nil, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get all orders: %w", err)
	}
	return &result, nil
}

// PreviewOrder previews an order for a specific account.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - accountHash: Account hash from LinkedAccounts()
//   - order: PreviewOrderRequest object containing order details to preview
//
// Returns PreviewOrderResponse containing preview results.
// Returns error if the request fails.
func (c *Client) PreviewOrder(ctx context.Context, accountHash string, order *PreviewOrderRequest) (*PreviewOrderResponse, error) {
	var result PreviewOrderResponse
	_, err := c.request(ctx, "POST", fmt.Sprintf("/trader/v1/accounts/%s/previewOrder", accountHash), order, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to preview order: %w", err)
	}
	return &result, nil
}

// Transactions retrieves all transactions for a specific account.
// Maximum number of transactions in response is 3000. Maximum date range is 1 year.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - accountHash: Account hash from LinkedAccounts()
//   - startDate: Start date (time.Time, string in ISO 8601 format, or nil)
//   - endDate: End date (time.Time, string in ISO 8601 format, or nil)
//   - types: Transaction type filter (see API documentation for possible values)
//   - symbol: Symbol filter (optional, can be nil)
//
// Returns TransactionsResponse containing list of transactions.
// Returns error if the request fails.
func (c *Client) Transactions(ctx context.Context, accountHash string, startDate, endDate any, types string, symbol *string) (*TransactionsResponse, error) {
	// Convert time parameters
	start, err := c.timeConvert(startDate, TimeFormatISO8601)
	if err != nil {
		return nil, fmt.Errorf("failed to convert startDate: %w", err)
	}

	end, err := c.timeConvert(endDate, TimeFormatISO8601)
	if err != nil {
		return nil, fmt.Errorf("failed to convert endDate: %w", err)
	}

	// Build query parameters
	params := c.parseParams(map[string]any{
		"startDate": start,
		"endDate":   end,
		"types":     types,
		"symbol":    symbol,
	})

	path := fmt.Sprintf("/trader/v1/accounts/%s/transactions", accountHash)
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	var result TransactionsResponse
	_, err = c.request(ctx, "GET", path, nil, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}
	return &result, nil
}

// TransactionDetails retrieves specific transaction information for a specific account.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - accountHash: Account hash from LinkedAccounts()
//   - transactionID: Transaction ID to retrieve
//
// Returns TransactionDetailsResponse containing transaction details.
// Returns error if the request fails.
func (c *Client) TransactionDetails(ctx context.Context, accountHash string, transactionID any) (*TransactionDetailsResponse, error) {
	var result TransactionDetailsResponse
	_, err := c.request(ctx, "GET", fmt.Sprintf("/trader/v1/accounts/%s/transactions/%v", accountHash, transactionID), nil, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction details: %w", err)
	}
	return &result, nil
}

// ============================================================================
// MARKET DATA API METHODS
// ============================================================================

// Quotes retrieves quotes for a list of tickers.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - symbols: List of symbols (can be string "AMD,INTC" or []string{"AMD", "INTC"})
//   - fields: Optional fields to return ("all", "quote", "fundamental")
//   - indicative: Whether to get indicative quotes
//
// Returns QuotesResponse containing quotes for all symbols.
// Returns error if the request fails.
func (c *Client) Quotes(ctx context.Context, symbols any, fields *string, indicative *bool) (*QuotesResponse, error) {
	params := c.parseParams(map[string]any{
		"symbols":    c.formatList(symbols),
		"fields":     fields,
		"indicative": indicative,
	})

	path := "/marketdata/v1/quotes"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	var result QuotesResponse
	_, err := c.request(ctx, "GET", path, nil, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get quotes: %w", err)
	}
	return &result, nil
}

// Quote retrieves a quote for a single symbol.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - symbolID: Ticker symbol
//   - fields: Optional fields to return ("all", "quote", "fundamental")
//
// Returns QuoteResponse containing quote for the symbol.
// Returns error if the request fails.
func (c *Client) Quote(ctx context.Context, symbolID string, fields *string) (*QuoteResponse, error) {
	params := c.parseParams(map[string]any{
		"fields": fields,
	})

	path := fmt.Sprintf("/marketdata/v1/%s/quotes", url.PathEscape(symbolID))
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	var result QuoteResponse
	_, err := c.request(ctx, "GET", path, nil, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get quote: %w", err)
	}
	return &result, nil
}

// OptionChains retrieves option chain information for a ticker.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - symbol: Ticker symbol
//   - contractType: Contract type ("CALL", "PUT", "ALL")
//   - strikeCount: Strike count
//   - includeUnderlyingQuote: Include underlying quote
//   - strategy: Strategy type
//   - interval: Strike interval
//   - strike: Strike price
//   - range_: Range ("ITM", "NTM", "OTM")
//   - fromDate: From date (time.Time, string, or nil)
//   - toDate: To date (time.Time, string, or nil)
//   - volatility: Volatility
//   - underlyingPrice: Underlying price
//   - interestRate: Interest rate
//   - daysToExpiration: Days to expiration
//   - expMonth: Expiration month
//   - optionType: Option type ("ALL", "CALL", "PUT")
//   - entitlement: Entitlement ("ALL", "AMERICAN", "EUROPEAN")
//
// Returns OptionChainsResponse containing option chain data.
// Returns error if the request fails.
func (c *Client) OptionChains(ctx context.Context, symbol string, contractType *string, strikeCount *int,
	includeUnderlyingQuote *bool, strategy *string, interval *string, strike *float64, range_ *string,
	fromDate, toDate any, volatility, underlyingPrice, interestRate *float64,
	daysToExpiration *int, expMonth, optionType, entitlement *string) (*OptionChainsResponse, error) {

	from, err := c.timeConvert(fromDate, TimeFormatYYYYMMDD)
	if err != nil {
		return nil, fmt.Errorf("failed to convert fromDate: %w", err)
	}

	to, err := c.timeConvert(toDate, TimeFormatYYYYMMDD)
	if err != nil {
		return nil, fmt.Errorf("failed to convert toDate: %w", err)
	}

	params := c.parseParams(map[string]any{
		"symbol":                 symbol,
		"contractType":           contractType,
		"strikeCount":            strikeCount,
		"includeUnderlyingQuote": includeUnderlyingQuote,
		"strategy":               strategy,
		"interval":               interval,
		"strike":                 strike,
		"range":                  range_,
		"fromDate":               from,
		"toDate":                 to,
		"volatility":             volatility,
		"underlyingPrice":        underlyingPrice,
		"interestRate":           interestRate,
		"daysToExpiration":       daysToExpiration,
		"expMonth":               expMonth,
		"optionType":             optionType,
		"entitlement":            entitlement,
	})

	path := "/marketdata/v1/chains"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	var result OptionChainsResponse
	_, err = c.request(ctx, "GET", path, nil, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get option chains: %w", err)
	}
	return &result, nil
}

// OptionExpirationChain retrieves an option expiration chain for a ticker.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - symbol: Ticker symbol
//
// Returns OptionExpirationChainResponse containing expiration dates.
// Returns error if the request fails.
func (c *Client) OptionExpirationChain(ctx context.Context, symbol string) (*OptionExpirationChainResponse, error) {
	params := c.parseParams(map[string]any{
		"symbol": symbol,
	})

	path := "/marketdata/v1/expirationchain"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	var result OptionExpirationChainResponse
	_, err := c.request(ctx, "GET", path, nil, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get option expiration chain: %w", err)
	}
	return &result, nil
}

// PriceHistory retrieves price history for a ticker.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - symbol: Ticker symbol
//   - periodType: Period type ("day", "month", "year", "ytd")
//   - period: Period
//   - frequencyType: Frequency type ("minute", "daily", "weekly", "monthly")
//   - frequency: Frequency
//   - startDate: Start date (time.Time, string, or nil)
//   - endDate: End date (time.Time, string, or nil)
//   - needExtendedHoursData: Need extended hours data
//   - needPreviousClose: Need previous close
//
// Returns PriceHistoryResponse containing candle history.
// Returns error if the request fails.
func (c *Client) PriceHistory(ctx context.Context, symbol string, periodType *string, period *int,
	frequencyType *string, frequency *int, startDate, endDate any,
	needExtendedHoursData, needPreviousClose *bool) (*PriceHistoryResponse, error) {

	start, err := c.timeConvert(startDate, TimeFormatEPOCHMS)
	if err != nil {
		return nil, fmt.Errorf("failed to convert startDate: %w", err)
	}

	end, err := c.timeConvert(endDate, TimeFormatEPOCHMS)
	if err != nil {
		return nil, fmt.Errorf("failed to convert endDate: %w", err)
	}

	params := c.parseParams(map[string]any{
		"symbol":                symbol,
		"periodType":            periodType,
		"period":                period,
		"frequencyType":         frequencyType,
		"frequency":             frequency,
		"startDate":             start,
		"endDate":               end,
		"needExtendedHoursData": needExtendedHoursData,
		"needPreviousClose":     needPreviousClose,
	})

	path := "/marketdata/v1/pricehistory"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	var result PriceHistoryResponse
	_, err = c.request(ctx, "GET", path, nil, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get price history: %w", err)
	}
	return &result, nil
}

// Movers retrieves market movers for a specific index.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - symbol: Symbol ("$DJI", "$COMPX", "$SPX", "NYSE", "NASDAQ", "OTCBB", etc.)
//   - sort: Sort type ("VOLUME", "TRADES", "PERCENT_CHANGE_UP", "PERCENT_CHANGE_DOWN")
//   - frequency: Frequency (0, 1, 5, 10, 30, 60)
//
// Returns MoversResponse containing market movers.
// Returns error if the request fails.
func (c *Client) Movers(ctx context.Context, symbol string, sort *string, frequency *int) (*MoversResponse, error) {
	params := c.parseParams(map[string]any{
		"sort":      sort,
		"frequency": frequency,
	})

	path := fmt.Sprintf("/marketdata/v1/movers/%s", symbol)
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	var result MoversResponse
	_, err := c.request(ctx, "GET", path, nil, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get movers: %w", err)
	}
	return &result, nil
}

// MarketHours retrieves market hours for dates in the future across different markets.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - symbols: List of market symbols ("equity", "option", "bond", "future", "forex")
//   - date: Date (time.Time, string, or nil)
//
// Returns MarketHoursResponse containing market hours.
// Returns error if the request fails.
func (c *Client) MarketHours(ctx context.Context, symbols any, date any) (*MarketHoursResponse, error) {
	convertedDate, err := c.timeConvert(date, TimeFormatYYYYMMDD)
	if err != nil {
		return nil, fmt.Errorf("failed to convert date: %w", err)
	}

	params := c.parseParams(map[string]any{
		"markets": c.formatList(symbols),
		"date":    convertedDate,
	})

	path := "/marketdata/v1/markets"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	var result MarketHoursResponse
	_, err = c.request(ctx, "GET", path, nil, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get market hours: %w", err)
	}
	return &result, nil
}

// MarketHour retrieves market hours for dates in the future for a single market.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - marketID: Market ID ("equity", "option", "bond", "future", "forex")
//   - date: Date (time.Time, string, or nil)
//
// Returns MarketHourResponse containing market hours.
// Returns error if the request fails.
func (c *Client) MarketHour(ctx context.Context, marketID string, date any) (*MarketHourResponse, error) {
	convertedDate, err := c.timeConvert(date, TimeFormatYYYYMMDD)
	if err != nil {
		return nil, fmt.Errorf("failed to convert date: %w", err)
	}

	params := c.parseParams(map[string]any{
		"date": convertedDate,
	})

	path := fmt.Sprintf("/marketdata/v1/markets/%s", marketID)
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	var result MarketHourResponse
	_, err = c.request(ctx, "GET", path, nil, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get market hour: %w", err)
	}
	return &result, nil
}

// Instruments retrieves instruments for a list of symbols.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - symbols: List of symbols (can be string or []string)
//   - projection: Projection type ("symbol-search", "symbol-regex", "desc-search", "desc-regex", "search", "fundamental")
//
// Returns InstrumentsResponse containing instrument search results.
// Returns error if the request fails.
func (c *Client) Instruments(ctx context.Context, symbols any, projection string) (*InstrumentsResponse, error) {
	params := c.parseParams(map[string]any{
		"symbol":     c.formatList(symbols),
		"projection": projection,
	})

	path := "/marketdata/v1/instruments"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	var result InstrumentsResponse
	_, err := c.request(ctx, "GET", path, nil, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get instruments: %w", err)
	}
	return &result, nil
}

// InstrumentCUSIP retrieves an instrument for a single CUSIP.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - cusipID: CUSIP ID
//
// Returns InstrumentCUSIPResponse containing instrument details.
// Returns error if the request fails.
func (c *Client) InstrumentCUSIP(ctx context.Context, cusipID any) (*InstrumentCUSIPResponse, error) {
	var result InstrumentCUSIPResponse
	_, err := c.request(ctx, "GET", fmt.Sprintf("/marketdata/v1/instruments/%v", cusipID), nil, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get instrument by CUSIP: %w", err)
	}
	return &result, nil
}
