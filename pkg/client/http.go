package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"time"
)

// Client wraps http.Client with custom configuration
type Client struct {
	httpClient *http.Client
	logger     *slog.Logger
}

// NewClient creates a new HTTP client with proper timeouts and connection settings
func NewClient(logger *slog.Logger) *Client {
	transport := &http.Transport{
		// Dial timeout: 30 seconds
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,

		// TLS handshake timeout: 10 seconds
		TLSHandshakeTimeout: 10 * time.Second,

		// Response header timeout: 10 seconds
		ResponseHeaderTimeout: 10 * time.Second,

		// Connection pool settings
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		DisableCompression:  false,
		ForceAttemptHTTP2:   true,
		MaxConnsPerHost:     0, // No limit
	}

	return &Client{
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   60 * time.Second, // Overall request timeout
		},
		logger: logger,
	}
}

// Get performs an HTTP GET request with context
func (c *Client) Get(ctx context.Context, url string, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		c.logger.Error("failed to create GET request",
			"url", url,
			"error", err,
		)
		return nil, fmt.Errorf("failed to create GET request: %w", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("GET request failed",
			"url", url,
			"error", err,
		)
		return nil, fmt.Errorf("GET request failed: %w", err)
	}

	return resp, nil
}

// Post performs an HTTP POST request with context and JSON body
func (c *Client) Post(ctx context.Context, url string, headers map[string]string, body any) (*http.Response, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		c.logger.Error("failed to marshal request body",
			"url", url,
			"error", err,
		)
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		c.logger.Error("failed to create POST request",
			"url", url,
			"error", err,
		)
		return nil, fmt.Errorf("failed to create POST request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("POST request failed",
			"url", url,
			"error", err,
		)
		return nil, fmt.Errorf("POST request failed: %w", err)
	}

	return resp, nil
}

// Put performs an HTTP PUT request with context and JSON body
func (c *Client) Put(ctx context.Context, url string, headers map[string]string, body any) (*http.Response, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		c.logger.Error("failed to marshal request body",
			"url", url,
			"error", err,
		)
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(jsonBody))
	if err != nil {
		c.logger.Error("failed to create PUT request",
			"url", url,
			"error", err,
		)
		return nil, fmt.Errorf("failed to create PUT request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("PUT request failed",
			"url", url,
			"error", err,
		)
		return nil, fmt.Errorf("PUT request failed: %w", err)
	}

	return resp, nil
}

// Delete performs an HTTP DELETE request with context
func (c *Client) Delete(ctx context.Context, url string, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		c.logger.Error("failed to create DELETE request",
			"url", url,
			"error", err,
		)
		return nil, fmt.Errorf("failed to create DELETE request: %w", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("DELETE request failed",
			"url", url,
			"error", err,
		)
		return nil, fmt.Errorf("DELETE request failed: %w", err)
	}

	return resp, nil
}

// DecodeJSON decodes the response body into the provided interface
func (c *Client) DecodeJSON(resp *http.Response, v any) error {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error("failed to read response body",
			"status", resp.StatusCode,
			"error", err,
		)
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if err := json.Unmarshal(body, v); err != nil {
		c.logger.Error("failed to unmarshal response",
			"status", resp.StatusCode,
			"body", string(body),
			"error", err,
		)
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return nil
}

// CloseIdleConnections closes any idle connections
func (c *Client) CloseIdleConnections() {
	c.httpClient.CloseIdleConnections()
}
