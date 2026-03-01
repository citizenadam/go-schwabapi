package token

import "time"

const (
	// AccessThreshold is the remaining lifetime below which an access token is refreshed (61 seconds)
	AccessThreshold = 61 * time.Second

	// RefreshThreshold is the remaining lifetime below which a refresh token is refreshed (60.5 minutes)
	RefreshThreshold = (60*60 + 30) * time.Second
)
