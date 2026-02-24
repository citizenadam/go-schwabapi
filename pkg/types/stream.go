package types

// Subscription represents a stream subscription request
type Subscription struct {
	Service    string              `json:"service,omitempty"`
	Command    string              `json:"command,omitempty"`
	RequestID  int                 `json:"requestid,omitempty"`
	Parameters *SubscriptionParams `json:"parameters,omitempty"`
}

// SubscriptionParams represents subscription parameters
type SubscriptionParams struct {
	Keys   string `json:"keys,omitempty"`
	Fields string `json:"fields,omitempty"`
}

// Message represents a stream message from the Schwab API
type Message struct {
	Service   string                 `json:"service,omitempty"`
	Command   string                 `json:"command,omitempty"`
	RequestID int                    `json:"requestid,omitempty"`
	Content   map[string]interface{} `json:"content,omitempty"`
}

// StreamRequest represents a stream request wrapper
type StreamRequest struct {
	Requests []Subscription `json:"requests,omitempty"`
}
