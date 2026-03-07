package schwabdev

import "errors"

// Parameter validation errors
var (
	// ErrAppKeyRequired indicates that app_key parameter is missing
	ErrAppKeyRequired = errors.New("[Schwabdev] app_key cannot be None.")

	// ErrAppSecretRequired indicates that app_secret parameter is missing
	ErrAppSecretRequired = errors.New("[Schwabdev] app_secret cannot be None.")

	// ErrCallbackURLRequired indicates that callback_url parameter is missing
	ErrCallbackURLRequired = errors.New("[Schwabdev] callback_url cannot be None.")

	// ErrTokensDBRequired indicates that tokens_db parameter is missing
	ErrTokensDBRequired = errors.New("[Schwabdev] tokens_db cannot be None.")

	// ErrInvalidKeyLength indicates app_key or app_secret has invalid length
	ErrInvalidKeyLength = errors.New("[Schwabdev] App key or app secret invalid length.")

	// ErrCallbackNotHTTPS indicates callback_url is not using HTTPS protocol
	ErrCallbackNotHTTPS = errors.New("[Schwabdev] callback_url must be https.")

	// ErrCallbackEndsWithSlash indicates callback_url is a path ending with /
	ErrCallbackEndsWithSlash = errors.New("[Schwabdev] callback_url cannot be path (ends with \"/\").")

	// ErrTokensDBEndsWithSlash indicates tokens_db path ends with /
	ErrTokensDBEndsWithSlash = errors.New("[Schwabdev] Tokens file cannot be path.")

	// ErrAuthCallbackNotFunc indicates call_on_notify is not a callable function
	ErrAuthCallbackNotFunc = errors.New("[Schwabdev] call_on_notify must be a callable function.")
)

// Encryption and token errors
var (
	// ErrEncryptionFailed indicates token encryption failed
	ErrEncryptionFailed = errors.New("Failed to encrypt token data.")

	// ErrDecryptionFailed indicates token cannot be decrypted without encryption key
	ErrDecryptionFailed = errors.New("Cannot decrypt token, no encryption key provided.")

	// ErrInvalidGrantType indicates an invalid OAuth grant type was specified
	ErrInvalidGrantType = errors.New("Invalid grant type; options are 'authorization_code' or 'refresh_token'")
)

// Client configuration errors
var (
	// ErrInvalidTimeout indicates timeout value is invalid
	ErrInvalidTimeout = errors.New("Timeout must be greater than 0 and is recommended to be 5 seconds or more.")

	// ErrUnsupportedTimeFormat indicates an unsupported time format was specified
	ErrUnsupportedTimeFormat = errors.New("Unsupported time format")
)

// Streaming errors
var (
	// ErrStreamerUnavailable indicates streamer information is not available
	ErrStreamerUnavailable = errors.New("Streamer info unavailable")
)
