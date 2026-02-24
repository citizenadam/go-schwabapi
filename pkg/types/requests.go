package types

// AccountDetailsRequest represents request parameters for account details
type AccountDetailsRequest struct {
	Fields string `json:"fields,omitempty"`
}

// AccountOrdersRequest represents request parameters for account orders
type AccountOrdersRequest struct {
	FromEnteredTime string `json:"fromEnteredTime,omitempty"`
	ToEnteredTime   string `json:"toEnteredTime,omitempty"`
	MaxResults      int    `json:"maxResults,omitempty"`
	Status          string `json:"status,omitempty"`
}

// PlaceOrderRequest represents request body for placing an order
type PlaceOrderRequest struct {
	Order *Order `json:"order,omitempty"`
}

// ReplaceOrderRequest represents request body for replacing an order
type ReplaceOrderRequest struct {
	Order *Order `json:"order,omitempty"`
}

// PreviewOrderRequest represents request body for previewing an order
type PreviewOrderRequest struct {
	Order *Order `json:"order,omitempty"`
}

// AccountOrdersAllRequest represents request parameters for all account orders
type AccountOrdersAllRequest struct {
	FromEnteredTime string `json:"fromEnteredTime,omitempty"`
	ToEnteredTime   string `json:"toEnteredTime,omitempty"`
	MaxResults      int    `json:"maxResults,omitempty"`
	Status          string `json:"status,omitempty"`
}

// TransactionsRequest represents request parameters for transactions
type TransactionsRequest struct {
	StartDate string `json:"startDate,omitempty"`
	EndDate   string `json:"endDate,omitempty"`
	Types     string `json:"types,omitempty"`
	Symbol    string `json:"symbol,omitempty"`
}

// QuotesRequest represents request parameters for quotes
type QuotesRequest struct {
	Symbols    string `json:"symbols,omitempty"`
	Fields     string `json:"fields,omitempty"`
	Indicative bool   `json:"indicative,omitempty"`
}

// QuoteRequest represents request parameters for a single quote
type QuoteRequest struct {
	Fields string `json:"fields,omitempty"`
}

// OptionChainsRequest represents request parameters for option chains
type OptionChainsRequest struct {
	Symbol                 string  `json:"symbol,omitempty"`
	ContractType           string  `json:"contractType,omitempty"`
	StrikeCount            int     `json:"strikeCount,omitempty"`
	IncludeUnderlyingQuote bool    `json:"includeUnderlyingQuote,omitempty"`
	Strategy               string  `json:"strategy,omitempty"`
	Interval               string  `json:"interval,omitempty"`
	Strike                 float64 `json:"strike,omitempty"`
	Range                  string  `json:"range,omitempty"`
	FromDate               string  `json:"fromDate,omitempty"`
	ToDate                 string  `json:"toDate,omitempty"`
	Volatility             float64 `json:"volatility,omitempty"`
	UnderlyingPrice        float64 `json:"underlyingPrice,omitempty"`
	InterestRate           float64 `json:"interestRate,omitempty"`
	DaysToExpiration       int     `json:"daysToExpiration,omitempty"`
	ExpMonth               string  `json:"expMonth,omitempty"`
	OptionType             string  `json:"optionType,omitempty"`
	Entitlement            string  `json:"entitlement,omitempty"`
}

// OptionExpirationChainRequest represents request parameters for option expiration chain
type OptionExpirationChainRequest struct {
	Symbol string `json:"symbol,omitempty"`
}

// PriceHistoryRequest represents request parameters for price history
type PriceHistoryRequest struct {
	Symbol                string `json:"symbol,omitempty"`
	PeriodType            string `json:"periodType,omitempty"`
	Period                string `json:"period,omitempty"`
	FrequencyType         string `json:"frequencyType,omitempty"`
	Frequency             int    `json:"frequency,omitempty"`
	StartDate             string `json:"startDate,omitempty"`
	EndDate               string `json:"endDate,omitempty"`
	NeedExtendedHoursData bool   `json:"needExtendedHoursData,omitempty"`
	NeedPreviousClose     bool   `json:"needPreviousClose,omitempty"`
}

// MoversRequest represents request parameters for movers
type MoversRequest struct {
	Sort      string `json:"sort,omitempty"`
	Frequency int    `json:"frequency,omitempty"`
}

// MarketHoursRequest represents request parameters for market hours
type MarketHoursRequest struct {
	Markets []string `json:"markets,omitempty"`
	Date    string   `json:"date,omitempty"`
}

// MarketHourRequest represents request parameters for a single market hour
type MarketHourRequest struct {
	Date string `json:"date,omitempty"`
}

// InstrumentsRequest represents request parameters for instruments
type InstrumentsRequest struct {
	Symbol     string `json:"symbol,omitempty"`
	Projection string `json:"projection,omitempty"`
}
