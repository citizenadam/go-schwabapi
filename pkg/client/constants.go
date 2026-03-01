package client

const (
	// BaseAPIURL is the root URL for all Schwab API endpoints
	BaseAPIURL = "https://api.schwabapi.com"

	// OAuth endpoints
	authorizeURL = BaseAPIURL + "/v1/oauth/authorize"
	tokenURL     = BaseAPIURL + "/v1/oauth/token"
	revokeURL    = BaseAPIURL + "/v1/oauth/revoke"
)
