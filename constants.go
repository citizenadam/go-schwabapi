package schwabdev

import "time"

// HTTP Client Constants
const (
	// DefaultHTTPRequestTimeout is the default timeout for HTTP requests to the Schwab API
	DefaultHTTPRequestTimeout = 10 * time.Second

	// OAuthTokenRequestTimeout is the timeout for OAuth token request operations
	OAuthTokenRequestTimeout = 30 * time.Second
)

// Token Management Constants
const (
	// AccessTokenValidity is the validity period for access tokens (30 minutes)
	AccessTokenValidity = 1800 * time.Second

	// RefreshTokenValidity is the validity period for refresh tokens (7 days)
	RefreshTokenValidity = 604800 * time.Second

	// AccessTokenRefreshThreshold is the time before expiry to refresh access token (61 seconds)
	AccessTokenRefreshThreshold = 61 * time.Second

	// RefreshTokenRefreshThreshold is the time before expiry to refresh refresh token (60.5 minutes)
	RefreshTokenRefreshThreshold = 3630 * time.Second
)

// WebSocket Streaming Constants
const (
	// WSPingInterval is the interval between WebSocket ping messages
	WSPingInterval = 20 * time.Second

	// WSPingTimeout is how long to wait for pong responses from the server
	WSPingTimeout = 30 * time.Second

	// WSReconnectBackoffInitial is the initial backoff time for reconnection attempts
	WSReconnectBackoffInitial = 2 * time.Second

	// WSReconnectBackoffMax is the maximum backoff time for reconnection attempts
	WSReconnectBackoffMax = 120 * time.Second

	// WSCrashThreshold is the threshold for detecting stream crashes
	WSCrashThreshold = 90 * time.Second

	// WSLoopReadyWait is the timeout for waiting for the event loop to be ready
	WSLoopReadyWait = 4 * time.Second

	// WSCloseTimeout is the timeout for closing the WebSocket connection
	WSCloseTimeout = 5 * time.Second

	// WSThreadJoinTimeout is the timeout for joining the streaming thread
	WSThreadJoinTimeout = 5 * time.Second
)

// Background Task Constants
const (
	// TokenCheckerSleep is the sleep interval for the token checker background task
	TokenCheckerSleep = 30 * time.Second

	// AutoCheckerSleep is the sleep interval for the auto checker background task
	AutoCheckerSleep = 30 * time.Second
)

// Validation Constants
const (
	// AppKeyLength1 is the first valid length for app keys
	AppKeyLength1 = 32

	// AppKeyLength2 is the second valid length for app keys
	AppKeyLength2 = 48

	// AppSecretLength1 is the first valid length for app secrets
	AppSecretLength1 = 16

	// AppSecretLength2 is the second valid length for app secrets
	AppSecretLength2 = 64
)

// Encryption Constants
const (
	// EncryptionPrefix is the prefix added to encrypted token values
	EncryptionPrefix = "enc:"
)
