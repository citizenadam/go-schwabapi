package types

// Account represents a Schwab trading account with full details
type Account struct {
	AccountHash      string `json:"accountHash,omitempty"`
	AccountNumber    string `json:"accountNumber,omitempty"`
	AccountType      string `json:"accountType,omitempty"`
	AccountStatus    string `json:"accountStatus,omitempty"`
	PrimaryAccountID string `json:"primaryAccountId,omitempty"`
	DisplayName      string `json:"displayName,omitempty"`

	// Full account details from securitiesAccount
	SecuritiesAccount *SecuritiesAccount `json:"securitiesAccount,omitempty"`
}

// SecuritiesAccount contains the full account details from Schwab API
type SecuritiesAccount struct {
	Type              string             `json:"type,omitempty"`
	AccountID         string             `json:"accountId,omitempty"`
	AccountNumber     string             `json:"accountNumber,omitempty"`
	RoundTrips        int                `json:"roundTrips,omitempty"`
	IsDayTrader       bool               `json:"isDayTrader,omitempty"`
	IsClosingOnly     bool               `json:"isClosingOnlyRestricted,omitempty"`
	CurrentBalances   *CurrentBalances   `json:"currentBalances,omitempty"`
	InitialBalances   *InitialBalances   `json:"initialBalances,omitempty"`
	Positions         []Position         `json:"positions,omitempty"`
	OrderStrategies   []OrderStrategy    `json:"orderStrategies,omitempty"`
	ProjectedBalances *ProjectedBalances `json:"projectedBalances,omitempty"`
}

// CurrentBalances contains current balance information
type CurrentBalances struct {
	AccruedInterest         float64 `json:"accruedInterest,omitempty"`
	CashBalance             float64 `json:"cashBalance,omitempty"`
	CashReceipts            float64 `json:"cashReceipts,omitempty"`
	LongOptionMarketValue   float64 `json:"longOptionMarketValue,omitempty"`
	ShortOptionMarketValue  float64 `json:"shortOptionMarketValue,omitempty"`
	LiquidationValue        float64 `json:"liquidationValue,omitempty"`
	LongMarketValue         float64 `json:"longMarketValue,omitempty"`
	ShortMarketValue        float64 `json:"shortMarketValue,omitempty"`
	MarginBalance           float64 `json:"marginBalance,omitempty"`
	AvailableFunds          float64 `json:"availableFunds,omitempty"`
	AvailableFundsNonMargin float64 `json:"availableFundsNonMarginable,omitempty"`
	BuyingPower             float64 `json:"buyingPower,omitempty"`
	DayTradingBuyingPower   float64 `json:"dayTradingBuyingPower,omitempty"`
	Equity                  float64 `json:"equity,omitempty"`
	EquityPercentage        float64 `json:"equityPercentage,omitempty"`
	IsInCall                bool    `json:"isInCall,omitempty"`
	UnsettledFunds          float64 `json:"unsettledFunds,omitempty"`
	Margin                  float64 `json:"margin,omitempty"`
}

// InitialBalances contains initial balance information
type InitialBalances struct {
	AccruedInterest        float64 `json:"accruedInterest,omitempty"`
	CashBalance            float64 `json:"cashBalance,omitempty"`
	CashReceipts           float64 `json:"cashReceipts,omitempty"`
	LongOptionMarketValue  float64 `json:"longOptionMarketValue,omitempty"`
	ShortOptionMarketValue float64 `json:"shortOptionMarketValue,omitempty"`
	LiquidationValue       float64 `json:"liquidationValue,omitempty"`
	LongMarketValue        float64 `json:"longMarketValue,omitempty"`
	ShortMarketValue       float64 `json:"shortMarketValue,omitempty"`
}

// ProjectedBalances contains projected balance information
type ProjectedBalances struct {
	CashAvailable float64 `json:"cashAvailable,omitempty"`
}

// Position represents a position from Schwab API
type Position struct {
	ShortQuantity                  float64    `json:"shortQuantity,omitempty"`
	AveragePrice                   float64    `json:"averagePrice,omitempty"`
	MarketValue                    float64    `json:"marketValue,omitempty"`
	LongQuantity                   float64    `json:"longQuantity,omitempty"`
	SettledLongQuantity            float64    `json:"settledLongQuantity,omitempty"`
	SettledShortQuantity           float64    `json:"settledShortQuantity,omitempty"`
	AverageLongPrice               float64    `json:"averageLongPrice,omitempty"`
	AverageShortPrice              float64    `json:"averageShortPrice,omitempty"`
	CurrentDayProfitLoss           float64    `json:"currentDayProfitLoss,omitempty"`
	CurrentDayProfitLossPercentage float64    `json:"currentDayProfitLossPercentage,omitempty"`
	Instrument                     Instrument `json:"instrument,omitempty"`
}

// OrderStrategy represents an order from Schwab API
type OrderStrategy struct {
	OrderID            string     `json:"orderId,omitempty"`
	Status             string     `json:"status,omitempty"`
	EnteredTime        string     `json:"enteredTime,omitempty"`
	CloseTime          string     `json:"closeTime,omitempty"`
	OrderType          string     `json:"orderType,omitempty"`
	Price              float64    `json:"price,omitempty"`
	OrderLegCollection []OrderLeg `json:"orderLegCollection,omitempty"`
	Quantity           float64    `json:"quantity,omitempty"`
	FilledQuantity     float64    `json:"filledQuantity,omitempty"`
	Session            string     `json:"session,omitempty"`
	Duration           string     `json:"duration,omitempty"`
}

// OrderLeg represents an order leg
type OrderLeg struct {
	OrderLegType string     `json:"orderLegType,omitempty"`
	Instrument   Instrument `json:"instrument,omitempty"`
	Quantity     float64    `json:"quantity,omitempty"`
	Instruction  string     `json:"instruction,omitempty"`
}

// Order represents a trading order (simplified for order creation)
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
