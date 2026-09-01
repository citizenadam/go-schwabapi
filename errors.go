package schwabdev

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Parameter validation errors
var (
	// ErrAppKeyRequired indicates that app_key parameter is missing
	ErrAppKeyRequired = errors.New("app_key cannot be empty")

	// ErrAppSecretRequired indicates that app_secret parameter is missing
	ErrAppSecretRequired = errors.New("app_secret cannot be empty")

	// ErrCallbackURLRequired indicates that callback_url parameter is missing
	ErrCallbackURLRequired = errors.New("callback_url cannot be empty")

	// ErrTokensDBRequired indicates that tokens_db parameter is missing
	ErrTokensDBRequired = errors.New("tokens_db cannot be empty")

	// ErrInvalidKeyLength indicates app_key or app_secret has invalid length
	ErrInvalidKeyLength = errors.New("app key or app secret has invalid length")

	// ErrCallbackNotHTTPS indicates callback_url is not using HTTPS protocol
	ErrCallbackNotHTTPS = errors.New("callback_url must use https")

	// ErrCallbackEndsWithSlash indicates callback_url is a path ending with /
	ErrCallbackEndsWithSlash = errors.New("callback_url cannot be a path ending with \"/\"")

	// ErrTokensDBEndsWithSlash indicates tokens_db path ends with /
	ErrTokensDBEndsWithSlash = errors.New("tokens file cannot be a path")

	// ErrAuthCallbackNotFunc indicates call_on_notify is not a callable function
	ErrAuthCallbackNotFunc = errors.New("call_on_notify must be a callable function")
)

// Encryption and token errors
var (
	// ErrEncryptionFailed indicates token encryption failed
	ErrEncryptionFailed = errors.New("failed to encrypt token data")

	// ErrDecryptionFailed indicates token cannot be decrypted without encryption key
	ErrDecryptionFailed = errors.New("cannot decrypt token: no encryption key provided")

	// ErrInvalidGrantType indicates an invalid OAuth grant type was specified
	ErrInvalidGrantType = errors.New("invalid grant type; options are 'authorization_code' or 'refresh_token'")
)

// Client configuration errors
var (
	// ErrInvalidTimeout indicates timeout value is invalid
	ErrInvalidTimeout = errors.New("timeout must be greater than 0 and is recommended to be 5 seconds or more")

	// ErrUnsupportedTimeFormat indicates an unsupported time format was specified
	ErrUnsupportedTimeFormat = errors.New("unsupported time format")
)

// Streaming errors
var (
	// ErrStreamerUnavailable indicates streamer information is not available
	ErrStreamerUnavailable = errors.New("streamer info unavailable")
)

// HTTP status-class sentinel errors. These are returned unwrapped nowhere on
// their own; instead, SchwabAPIError implements Is so that errors.Is on a
// non-2xx API error matches the sentinel for its status class. This lets
// consumers write errors.Is(err, ErrRateLimited) without type-asserting.
var (
	// ErrBadRequest matches any Schwab API error with HTTP 400.
	ErrBadRequest = errors.New("bad request")
	// ErrUnauthorized matches any Schwab API error with HTTP 401.
	ErrUnauthorized = errors.New("unauthorized")
	// ErrForbidden matches any Schwab API error with HTTP 403.
	ErrForbidden = errors.New("forbidden")
	// ErrNotFound matches any Schwab API error with HTTP 404.
	ErrNotFound = errors.New("not found")
	// ErrRateLimited matches any Schwab API error with HTTP 429.
	ErrRateLimited = errors.New("rate limited")
	// ErrServer matches any Schwab API error with a 5xx status.
	ErrServer = errors.New("server error")
)

// SchwabAPIError is returned for HTTP responses with a non-2xx status code.
// StatusCode carries the HTTP status and Body the raw response payload, so
// callers can classify failures (4xx, 429 rate-limit, 401 auth) from the
// status code alone instead of string-matching response bodies.
//
// It also carries RequestID (from the schwab-client-correl-id response header,
// for correlating with Schwab support), RetryAfter (parsed from the Retry-After
// header on 429 responses), and Message (a human-readable message parsed from
// the response body). The Is method makes errors.Is(err, ErrRateLimited) work
// against any API error for the matching status class.
type SchwabAPIError struct {
	StatusCode int
	Body       []byte
	// RequestID is the schwab-client-correl-id response header, if present.
	RequestID string
	// RetryAfter is the parsed Retry-After duration. Non-zero only for 429s.
	RetryAfter time.Duration
	// Message is a human-readable message parsed from the response body.
	Message string
}

func (e *SchwabAPIError) Error() string {
	msg := e.Message
	if msg == "" && len(e.Body) > 0 {
		msg = string(e.Body)
	}
	if msg != "" {
		return fmt.Sprintf("Schwab API error (status %d): %s", e.StatusCode, msg)
	}
	return fmt.Sprintf("Schwab API error (status %d)", e.StatusCode)
}

// Is reports whether the target matches the status class of e. It enables
// errors.Is(err, ErrRateLimited) and friends against a SchwabAPIError.
func (e *SchwabAPIError) Is(target error) bool {
	if e == nil {
		return false
	}
	switch target {
	case ErrBadRequest:
		return e.StatusCode == http.StatusBadRequest
	case ErrUnauthorized:
		return e.StatusCode == http.StatusUnauthorized
	case ErrForbidden:
		return e.StatusCode == http.StatusForbidden
	case ErrNotFound:
		return e.StatusCode == http.StatusNotFound
	case ErrRateLimited:
		return e.StatusCode == http.StatusTooManyRequests
	case ErrServer:
		return e.StatusCode >= 500 && e.StatusCode <= 599
	default:
		return false
	}
}

// parseRetryAfter parses a Retry-After header value into a duration. It accepts
// an integer number of seconds or an HTTP-date after which to retry, clamping
// to >= 0. Returns 0 when the header is absent or unparseable.
func parseRetryAfter(value string) time.Duration {
	if value == "" {
		return 0
	}
	if secs, err := strconv.Atoi(value); err == nil {
		return time.Duration(secs) * time.Second
	}
	if when, err := http.ParseTime(value); err == nil {
		return max(time.Until(when), 0)
	}
	return 0
}

// parseErrorBody extracts a human-readable message (and optional code) from a
// Schwab API error response body. Schwab returns errors in several shapes, so
// this tolerates them all and always falls back to the raw trimmed body text
// rather than failing.
func parseErrorBody(body []byte) (message, code string) {
	if len(body) == 0 {
		return "", ""
	}
	var m map[string]any
	if err := json.Unmarshal(body, &m); err != nil {
		return strings.TrimSpace(string(body)), ""
	}

	// {"errors": [{"message": "...", "code": "..."}, ...]}
	if errs, ok := m["errors"].([]any); ok && len(errs) > 0 {
		if first, ok := errs[0].(map[string]any); ok {
			msg, _ := first["message"].(string)
			cd, _ := first["code"].(string)
			if msg != "" || cd != "" {
				return msg, cd
			}
		}
	}

	// {"message": "..."} / {"error": "..."} / {"error_description": "..."}
	for _, key := range []string{"message", "error", "error_description"} {
		if v, ok := m[key].(string); ok && v != "" {
			return v, ""
		}
	}

	return strings.TrimSpace(string(body)), ""
}
