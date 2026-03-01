package client

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// OrdersClient handles order management operations
type OrdersClient struct {
	client      *Client
	logger      *slog.Logger
	baseURL     string // For testing purposes
	tokenGetter TokenGetter
}

// NewOrdersClient creates a new orders client
func NewOrdersClient(httpClient *Client, logger *slog.Logger, tokenGetter TokenGetter) *OrdersClient {
	return &OrdersClient{
		client:      httpClient,
		logger:      logger,
		baseURL:     baseAPIURL,
		tokenGetter: tokenGetter,
	}
}

// SetBaseURL sets the base URL for testing purposes
func (o *OrdersClient) SetBaseURL(url string) {
	o.baseURL = url
}

// PlaceOrder places an order for a specific account
// POST /v1/accounts/{accountHash}/orders
func (o *OrdersClient) PlaceOrder(ctx context.Context, accountHash string, order any) (*http.Response, error) {
	// Add deadline to prevent indefinite blocking
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	url := fmt.Sprintf("%s/v1/accounts/%s/orders", o.baseURL, accountHash)
	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", o.tokenGetter.GetAccessToken()),
		"Accept":        "application/json",
		"Content-Type":  "application/json",
	}

	o.logger.Info("placing order",
		"accountHash", accountHash,
		"url", url,
	)

	resp, err := o.client.Post(ctx, url, headers, order)
	if err != nil {
		o.logger.Error("failed to place order",
			"accountHash", accountHash,
			"error", err,
		)
		return nil, fmt.Errorf("failed to place order: %w", err)
	}

	o.logger.Info("order placed successfully",
		"accountHash", accountHash,
		"status", resp.StatusCode,
	)

	return resp, nil
}

// PreviewOrder validates an order before placing it
// POST /v1/accounts/{accountHash}/orders/validate
func (o *OrdersClient) PreviewOrder(ctx context.Context, accountHash string, order any) (*http.Response, error) {
	// Add deadline to prevent indefinite blocking
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	url := fmt.Sprintf("%s/v1/accounts/%s/orders/validate", o.baseURL, accountHash)
	headers := map[string]string{
		"Accept":       "application/json",
		"Content-Type": "application/json",
	}

	o.logger.Info("previewing order",
		"accountHash", accountHash,
		"url", url,
	)

	resp, err := o.client.Post(ctx, url, headers, order)
	if err != nil {
		o.logger.Error("failed to preview order",
			"accountHash", accountHash,
			"error", err,
		)
		return nil, fmt.Errorf("failed to preview order: %w", err)
	}

	o.logger.Info("order previewed successfully",
		"accountHash", accountHash,
		"status", resp.StatusCode,
	)

	return resp, nil
}

// CancelOrder cancels a specific order by its ID
// DELETE /v1/accounts/{accountHash}/orders/{orderId}
func (o *OrdersClient) CancelOrder(ctx context.Context, accountHash string, orderId string) (*http.Response, error) {
	// Add deadline to prevent indefinite blocking
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	url := fmt.Sprintf("%s/v1/accounts/%s/orders/%s", o.baseURL, accountHash, orderId)
	headers := map[string]string{}

	o.logger.Info("cancelling order",
		"accountHash", accountHash,
		"orderId", orderId,
		"url", url,
	)

	resp, err := o.client.Delete(ctx, url, headers)
	if err != nil {
		o.logger.Error("failed to cancel order",
			"accountHash", accountHash,
			"orderId", orderId,
			"error", err,
		)
		return nil, fmt.Errorf("failed to cancel order: %w", err)
	}

	o.logger.Info("order cancelled successfully",
		"accountHash", accountHash,
		"orderId", orderId,
		"status", resp.StatusCode,
	)

	return resp, nil
}

// ReplaceOrder replaces an existing order for an account
// PUT /v1/accounts/{accountHash}/orders/{orderId}
func (o *OrdersClient) ReplaceOrder(ctx context.Context, accountHash string, orderId string, order any) (*http.Response, error) {
	// Add deadline to prevent indefinite blocking
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	url := fmt.Sprintf("%s/v1/accounts/%s/orders/%s", o.baseURL, accountHash, orderId)
	headers := map[string]string{
		"Accept":       "application/json",
		"Content-Type": "application/json",
	}

	o.logger.Info("replacing order",
		"accountHash", accountHash,
		"orderId", orderId,
		"url", url,
	)

	resp, err := o.client.Put(ctx, url, headers, order)
	if err != nil {
		o.logger.Error("failed to replace order",
			"accountHash", accountHash,
			"orderId", orderId,
			"error", err,
		)
		return nil, fmt.Errorf("failed to replace order: %w", err)
	}

	o.logger.Info("order replaced successfully",
		"accountHash", accountHash,
		"orderId", orderId,
		"status", resp.StatusCode,
	)

	return resp, nil
}

// OrderDetails gets a specific order by its ID for a specific account
// GET /v1/accounts/{accountHash}/orders/{orderId}
func (o *OrdersClient) OrderDetails(ctx context.Context, accountHash string, orderId string) (*http.Response, error) {
	// Add deadline to prevent indefinite blocking
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	url := fmt.Sprintf("%s/v1/accounts/%s/orders/%s", o.baseURL, accountHash, orderId)
	headers := map[string]string{}

	o.logger.Info("getting order details",
		"accountHash", accountHash,
		"orderId", orderId,
		"url", url,
	)

	resp, err := o.client.Get(ctx, url, headers)
	if err != nil {
		o.logger.Error("failed to get order details",
			"accountHash", accountHash,
			"orderId", orderId,
			"error", err,
		)
		return nil, fmt.Errorf("failed to get order details: %w", err)
	}

	o.logger.Info("order details retrieved successfully",
		"accountHash", accountHash,
		"orderId", orderId,
		"status", resp.StatusCode,
	)

	return resp, nil
}
