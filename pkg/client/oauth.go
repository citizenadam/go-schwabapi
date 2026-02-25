package client

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/citizenadam/go-schwabapi/pkg/types"
)

const (
	// OAuth endpoints
	authorizeURL = "https://api.schwabapi.com/v1/oauth/authorize"
	tokenURL     = "https://api.schwabapi.com/v1/oauth/token"
	revokeURL    = "https://api.schwabapi.com/v1/oauth/revoke"
)

// OAuthClient handles OAuth operations for Schwab API
type OAuthClient struct {
	httpClient  *Client
	logger      *slog.Logger
	appKey      string
	appSecret   string
	callbackURL string
	tokenGetter TokenGetter
	baseURL     string
}

// NewOAuthClient creates a new OAuth client
func NewOAuthClient(httpClient *Client, logger *slog.Logger, appKey, appSecret, callbackURL string, tokenGetter TokenGetter) *OAuthClient {
	return &OAuthClient{
		httpClient:  httpClient,
		logger:      logger,
		appKey:      appKey,
		appSecret:   appSecret,
		callbackURL: callbackURL,
		tokenGetter: tokenGetter,
		baseURL:     "https://api.schwabapi.com",
	}
}

// Authorize returns the authorization URL for the user to authenticate
func (o *OAuthClient) Authorize(ctx context.Context) (string, error) {
	// Build authorization URL with query parameters
	authURL, err := url.Parse(authorizeURL)
	if err != nil {
		o.logger.Error("failed to parse authorize URL",
			"error", err,
		)
		return "", fmt.Errorf("failed to parse authorize URL: %w", err)
	}

	query := authURL.Query()
	query.Set("client_id", o.appKey)
	query.Set("redirect_uri", o.callbackURL)
	query.Set("response_type", "code")
	authURL.RawQuery = query.Encode()

	o.logger.Info("authorization URL generated",
		"url", authURL.String(),
	)

	return authURL.String(), nil
}

// RefreshToken exchanges a refresh token for a new access token
func (o *OAuthClient) RefreshToken(ctx context.Context, refreshToken string) (*types.Token, error) {
	// Add deadline to prevent blocking indefinitely
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Prepare form data for refresh token request
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)

	// Create Basic Auth header
	authHeader := o.createBasicAuthHeader()

	// Set headers
	headers := map[string]string{
		"Authorization": authHeader,
		"Content-Type":  "application/x-www-form-urlencoded",
	}

	o.logger.Debug("refreshing access token",
		"grant_type", "refresh_token",
	)

	// Make POST request with form data
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		o.logger.Error("failed to create refresh token request",
			"error", err,
		)
		return nil, fmt.Errorf("failed to create refresh token request: %w", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := o.httpClient.httpClient.Do(req)
	if err != nil {
		o.logger.Error("refresh token request failed",
			"error", err,
		)
		return nil, fmt.Errorf("refresh token request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		o.logger.Error("refresh token request failed",
			"status", resp.StatusCode,
		)
		return nil, fmt.Errorf("refresh token request failed with status: %d", resp.StatusCode)
	}

	// Decode response
	var token types.Token
	if err := o.httpClient.DecodeJSON(resp, &token); err != nil {
		o.logger.Error("failed to decode refresh token response",
			"error", err,
		)
		return nil, fmt.Errorf("failed to decode refresh token response: %w", err)
	}

	o.logger.Info("access token refreshed successfully")

	return &token, nil
}

// RevokeToken revokes an OAuth token (access or refresh token)
func (o *OAuthClient) RevokeToken(ctx context.Context, token string, tokenType string) error {
	// Add deadline to prevent blocking indefinitely
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Prepare form data for revoke request
	data := url.Values{}
	data.Set("token", token)
	data.Set("token_type_hint", tokenType)

	// Create Basic Auth header
	authHeader := o.createBasicAuthHeader()

	// Set headers
	headers := map[string]string{
		"Authorization": authHeader,
		"Content-Type":  "application/x-www-form-urlencoded",
	}

	o.logger.Debug("revoking token",
		"token_type", tokenType,
	)

	// Make POST request with form data
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, revokeURL, strings.NewReader(data.Encode()))
	if err != nil {
		o.logger.Error("failed to create revoke token request",
			"error", err,
		)
		return fmt.Errorf("failed to create revoke token request: %w", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := o.httpClient.httpClient.Do(req)
	if err != nil {
		o.logger.Error("revoke token request failed",
			"error", err,
		)
		return fmt.Errorf("revoke token request failed: %w", err)
	}
	defer resp.Body.Close()

	// Schwab returns 200 OK for successful revocation
	if resp.StatusCode != http.StatusOK {
		o.logger.Error("revoke token request failed",
			"status", resp.StatusCode,
		)
		return fmt.Errorf("revoke token request failed with status: %d", resp.StatusCode)
	}

	o.logger.Info("token revoked successfully",
		"token_type", tokenType,
	)

	return nil
}

// GetStreamerInfo retrieves streaming authentication information from user preferences
// Endpoint: GET /trader/v1/userPreference
// Returns streamerInfo containing authentication details for streaming services
func (o *OAuthClient) GetStreamerInfo(ctx context.Context) (*types.StreamerInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	apiURL := fmt.Sprintf("%s/trader/v1/userPreference", o.baseURL)

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", o.tokenGetter.GetAccessToken()),
		"Accept":        "application/json",
	}

	o.logger.Debug("getting streamer info",
		"url", apiURL,
	)

	resp, err := o.httpClient.Get(ctx, apiURL, headers)
	if err != nil {
		o.logger.Error("failed to get streamer info",
			"url", apiURL,
			"error", err,
		)
		return nil, fmt.Errorf("failed to get streamer info: %w", err)
	}

	var result types.PreferencesResponse
	if err := o.httpClient.DecodeJSON(resp, &result); err != nil {
		o.logger.Error("failed to decode streamer info response",
			"error", err,
		)
		return nil, fmt.Errorf("failed to decode streamer info response: %w", err)
	}

	if result.StreamerInfo == nil {
		o.logger.Error("streamer info not found in preferences response")
		return nil, fmt.Errorf("streamer info not found in preferences response")
	}

	o.logger.Info("successfully retrieved streamer info",
		"accountId", result.StreamerInfo.AccountID,
	)

	return result.StreamerInfo, nil
}

// createBasicAuthHeader creates a Basic Auth header from app credentials
func (o *OAuthClient) createBasicAuthHeader() string {
	credentials := fmt.Sprintf("%s:%s", o.appKey, o.appSecret)
	encoded := base64.StdEncoding.EncodeToString([]byte(credentials))
	return fmt.Sprintf("Basic %s", encoded)
}
