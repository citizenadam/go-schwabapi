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

// Order Instruction Constants
const (
	// InstructionBuyToOpen is buying to open a position
	InstructionBuyToOpen = "BUY_TO_OPEN"
	// InstructionBuyToClose is buying to close a position
	InstructionBuyToClose = "BUY_TO_CLOSE"
	// InstructionSellToOpen is selling to open a position
	InstructionSellToOpen = "SELL_TO_OPEN"
	// InstructionSellToClose is selling to close a position
	InstructionSellToClose = "SELL_TO_CLOSE"
)

// Order Position Effect Constants
const (
	// PositionEffectOpen opens a new position
	PositionEffectOpen = "OPEN"
	// PositionEffectClosed closes an existing position
	PositionEffectClosed = "CLOSED"
)

// Complex Order Strategy Constants
const (
	// ComplexOrderStrategyButterfly is an Iron Butterfly spread
	ComplexOrderStrategyButterfly = "IRON_BUTTERFLY"
	// ComplexOrderStrategyVertical is a Vertical spread
	ComplexOrderStrategyVertical = "VERTICAL"
	// ComplexOrderStrategyCombo is a Combo spread
	ComplexOrderStrategyCombo = "COMBO"
	// ComplexOrderStrategyCovered is a Covered stock position
	ComplexOrderStrategyCovered = "COVERED"
)

// Order Strategy Type Constants
const (
	// OrderStrategyTypeSingle is a single-leg order
	OrderStrategyTypeSingle = "SINGLE"
	// OrderStrategyTypeMultileg is a multi-leg order
	OrderStrategyTypeMultileg = "MULTILEG"
	// OrderStrategyTypeOco is one-cancels-other
	OrderStrategyTypeOco = "OCO"
	// OrderStrategyTypeRatio is a ratio spread
	OrderStrategyTypeRatio = "RATIO"
)

// Order Duration Constants
const (
	// OrderDurationDay is good for the day only
	OrderDurationDay = "DAY"
	// OrderDurationGoodTillCancel is good until cancelled
	OrderDurationGoodTillCancel = "GOOD_TILL_CANCEL"
	// OrderDurationFillOrKill is immediate or cancel
	OrderDurationFillOrKill = "FILL_OR_KILL"
)

// Order Session Constants
const (
	// OrderSessionNormal is the normal market session
	OrderSessionNormal = "NORMAL"
	// OrderSessionAM is the morning session
	OrderSessionAM = "AM"
	// OrderSessionPM is the afternoon session
	OrderSessionPM = "PM"
	// OrderSessionSExtended is extended hours
	OrderSessionExtended = "SEXTENDED"
)

// Order Type Constants
const (
	// OrderTypeLimit is a limit order
	OrderTypeLimit = "LIMIT"
	// OrderTypeMarket is a market order
	OrderTypeMarket = "MARKET"
	// OrderTypeStop is a stop order
	OrderTypeStop = "STOP"
	// OrderTypeStopLimit is a stop-limit order
	OrderTypeStopLimit = "STOP_LIMIT"
	// OrderTypeNetDebit is a net debit (buy) order
	OrderTypeNetDebit = "NET_DEBIT"
	// OrderTypeNetCredit is a net credit (sell) order
	OrderTypeNetCredit = "NET_CREDIT"
)
