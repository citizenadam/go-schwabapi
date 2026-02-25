package client

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"github.com/citizenadam/go-schwabapi/pkg/types"
)

const (
	baseAPIURL = "https://api.schwabapi.com"
)

// Accounts handles account-related API endpoints
type Accounts struct {
	httpClient  *Client
	logger      *slog.Logger
	tokenGetter TokenGetter
}

// TokenGetter interface for getting access tokens
type TokenGetter interface {
	GetAccessToken() string
}

// NewAccounts creates a new Accounts client
func NewAccounts(httpClient *Client, logger *slog.Logger, tokenGetter TokenGetter) *Accounts {
	return &Accounts{
		httpClient:  httpClient,
		logger:      logger,
		tokenGetter: tokenGetter,
	}
}

// LinkedAccounts retrieves all linked account numbers and hashes
// Endpoint: GET /trader/v1/accounts/accountNumbers
func (a *Accounts) LinkedAccounts(ctx context.Context) (*types.LinkedAccountsResponse, error) {
	// Create context with deadline to prevent indefinite blocking
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	apiURL := fmt.Sprintf("%s/trader/v1/accounts/accountNumbers", baseAPIURL)

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", a.tokenGetter.GetAccessToken()),
		"Accept":        "application/json",
	}

	resp, err := a.httpClient.Get(ctx, apiURL, headers)
	if err != nil {
		a.logger.Error("failed to get linked accounts",
			"url", apiURL,
			"error", err,
		)
		return nil, fmt.Errorf("failed to get linked accounts: %w", err)
	}

	var result types.LinkedAccountsResponse
	if err := a.httpClient.DecodeJSON(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode linked accounts response: %w", err)
	}

	a.logger.Info("successfully retrieved linked accounts",
		"count", len(result.AccountNumbers),
	)

	return &result, nil
}

// AccountDetails retrieves specific account information with balances and positions
// Endpoint: GET /trader/v1/accounts/{accountHash}
func (a *Accounts) AccountDetails(ctx context.Context, accountHash string, fields string) (*types.AccountDetailsResponse, error) {
	// Create context with deadline to prevent indefinite blocking
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	apiURL := fmt.Sprintf("%s/trader/v1/accounts/%s", baseAPIURL, url.PathEscape(accountHash))

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", a.tokenGetter.GetAccessToken()),
		"Accept":        "application/json",
	}

	resp, err := a.httpClient.Get(ctx, apiURL, headers)
	if err != nil {
		a.logger.Error("failed to get account details",
			"url", apiURL,
			"accountHash", accountHash,
			"error", err,
		)
		return nil, fmt.Errorf("failed to get account details: %w", err)
	}

	var result types.AccountDetailsResponse
	if err := a.httpClient.DecodeJSON(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode account details response: %w", err)
	}

	a.logger.Info("successfully retrieved account details",
		"accountHash", accountHash,
		"accountsCount", len(result.Accounts),
	)

	return &result, nil
}

// AccountDetailsAll retrieves all linked accounts with balances and positions
// Endpoint: GET /trader/v1/accounts/
func (a *Accounts) AccountDetailsAll(ctx context.Context, fields string) (*types.AccountDetailsAllResponse, error) {
	// Create context with deadline to prevent indefinite blocking
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	apiURL := fmt.Sprintf("%s/trader/v1/accounts/", baseAPIURL)

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", a.tokenGetter.GetAccessToken()),
		"Accept":        "application/json",
	}

	// Build query parameters
	params := url.Values{}
	if fields != "" {
		params.Add("fields", fields)
	}

	// Append query string to URL if we have parameters
	if len(params) > 0 {
		apiURL = fmt.Sprintf("%s?%s", apiURL, params.Encode())
	}

	resp, err := a.httpClient.Get(ctx, apiURL, headers)
	if err != nil {
		a.logger.Error("failed to get all account details",
			"url", apiURL,
			"error", err,
		)
		return nil, fmt.Errorf("failed to get all account details: %w", err)
	}

	var result types.AccountDetailsAllResponse
	if err := a.httpClient.DecodeJSON(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode account details all response: %w", err)
	}

	a.logger.Info("successfully retrieved all account details",
		"count", len(result.Accounts),
	)

	return &result, nil
}

// Preferences retrieves user preferences including streamer information
// Endpoint: GET /trader/v1/userPreference
func (a *Accounts) Preferences(ctx context.Context) (*types.PreferencesResponse, error) {
	// Create context with deadline to prevent indefinite blocking
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	apiURL := fmt.Sprintf("%s/trader/v1/userPreference", baseAPIURL)

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", a.tokenGetter.GetAccessToken()),
		"Accept":        "application/json",
	}

	resp, err := a.httpClient.Get(ctx, apiURL, headers)
	if err != nil {
		a.logger.Error("failed to get user preferences",
			"url", apiURL,
			"error", err,
		)
		return nil, fmt.Errorf("failed to get user preferences: %w", err)
	}

	var result types.PreferencesResponse
	if err := a.httpClient.DecodeJSON(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode preferences response: %w", err)
	}

	a.logger.Info("successfully retrieved user preferences")

	return &result, nil
}

// AccountOrders retrieves all orders for a specific account
// Orders retrieved can be filtered based on input parameters. Maximum date range is 1 year.
// Endpoint: GET /trader/v1/accounts/{accountHash}/orders
func (a *Accounts) AccountOrders(ctx context.Context, accountHash string, fromEnteredTime string, toEnteredTime string, maxResults int, status string) (*types.AccountOrdersResponse, error) {
	// Create context with deadline to prevent indefinite blocking
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	apiURL := fmt.Sprintf("%s/trader/v1/accounts/%s/orders", baseAPIURL, url.PathEscape(accountHash))

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", a.tokenGetter.GetAccessToken()),
		"Accept":        "application/json",
	}

	// Build query parameters
	params := url.Values{}
	if fromEnteredTime != "" {
		params.Add("fromEnteredTime", fromEnteredTime)
	}
	if toEnteredTime != "" {
		params.Add("toEnteredTime", toEnteredTime)
	}
	if maxResults > 0 {
		params.Add("maxResults", fmt.Sprintf("%d", maxResults))
	}
	if status != "" {
		params.Add("status", status)
	}

	// Append query string to URL if we have parameters
	if len(params) > 0 {
		apiURL = fmt.Sprintf("%s?%s", apiURL, params.Encode())
	}

	resp, err := a.httpClient.Get(ctx, apiURL, headers)
	if err != nil {
		a.logger.Error("failed to get account orders",
			"url", apiURL,
			"accountHash", accountHash,
			"error", err,
		)
		return nil, fmt.Errorf("failed to get account orders: %w", err)
	}

	var result types.AccountOrdersResponse
	if err := a.httpClient.DecodeJSON(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode account orders response: %w", err)
	}

	a.logger.Info("successfully retrieved account orders",
		"accountHash", accountHash,
		"ordersCount", len(result.Orders),
	)

	return &result, nil
}

// AccountOrdersAll retrieves all orders across all accounts
// Orders retrieved can be filtered based on input parameters. Maximum date range is 1 year.
// Endpoint: GET /trader/v1/orders
func (a *Accounts) AccountOrdersAll(ctx context.Context, fromEnteredTime string, toEnteredTime string, maxResults int, status string) (*types.AccountOrdersAllResponse, error) {
	// Create context with deadline to prevent indefinite blocking
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	apiURL := fmt.Sprintf("%s/trader/v1/orders", baseAPIURL)

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", a.tokenGetter.GetAccessToken()),
		"Accept":        "application/json",
	}

	// Build query parameters
	params := url.Values{}
	if fromEnteredTime != "" {
		params.Add("fromEnteredTime", fromEnteredTime)
	}
	if toEnteredTime != "" {
		params.Add("toEnteredTime", toEnteredTime)
	}
	if maxResults > 0 {
		params.Add("maxResults", fmt.Sprintf("%d", maxResults))
	}
	if status != "" {
		params.Add("status", status)
	}

	// Append query string to URL if we have parameters
	if len(params) > 0 {
		apiURL = fmt.Sprintf("%s?%s", apiURL, params.Encode())
	}

	resp, err := a.httpClient.Get(ctx, apiURL, headers)
	if err != nil {
		a.logger.Error("failed to get all account orders",
			"url", apiURL,
			"error", err,
		)
		return nil, fmt.Errorf("failed to get all account orders: %w", err)
	}

	var result types.AccountOrdersAllResponse
	if err := a.httpClient.DecodeJSON(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode account orders all response: %w", err)
	}

	a.logger.Info("successfully retrieved all account orders",
		"ordersCount", len(result.Orders),
	)

	return &result, nil
}
