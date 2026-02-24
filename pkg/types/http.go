package types

// Account represents a Schwab trading account
type Account struct {
	AccountHash      string `json:"accountHash,omitempty"`
	AccountNumber    string `json:"accountNumber,omitempty"`
	AccountType      string `json:"accountType,omitempty"`
	AccountStatus    string `json:"accountStatus,omitempty"`
	PrimaryAccountID string `json:"primaryAccountId,omitempty"`
	DisplayName      string `json:"displayName,omitempty"`
}

// Order represents a trading order
type Order struct {
	Session          string `json:"session,omitempty"`
	Duration         string `json:"duration,omitempty"`
	OrderType        string `json:"orderType,omitempty"`
	ComplexOrderType string `json:"complexOrderType,omitempty"`
	Quantity         int    `json:"quantity,omitempty"`
	Price            string `json:"price,omitempty"`
	StopPrice        string `json:"stopPrice,omitempty"`
	Symbol           string `json:"symbol,omitempty"`
	SymbolAssetType  string `json:"symbolAssetType,omitempty"`
	Instruction      string `json:"instruction,omitempty"`
}

// Token represents OAuth tokens for Schwab API authentication
type Token struct {
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
	ExpiresIn    int    `json:"expires_in,omitempty"`
	TokenType    string `json:"token_type,omitempty"`
	Scope        string `json:"scope,omitempty"`
}
