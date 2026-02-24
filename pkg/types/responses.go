package types

// LinkedAccountsResponse represents response for linked accounts
type LinkedAccountsResponse struct {
	AccountNumbers []AccountNumber `json:"accountNumbers,omitempty"`
}

// AccountNumber represents an account number and hash pair
type AccountNumber struct {
	AccountNumber string `json:"accountNumber,omitempty"`
	AccountHash   string `json:"accountHash,omitempty"`
}

// AccountDetailsResponse represents response for account details
type AccountDetailsResponse struct {
	Accounts []Account `json:"accounts,omitempty"`
}

// AccountOrdersResponse represents response for account orders
type AccountOrdersResponse struct {
	Orders []Order `json:"orders,omitempty"`
}

// OrderDetailsResponse represents response for order details
type OrderDetailsResponse struct {
	Order *Order `json:"order,omitempty"`
}

// PreviewOrderResponse represents response for preview order
type PreviewOrderResponse struct {
	Order *Order `json:"order,omitempty"`
}

// TransactionsResponse represents response for transactions
type TransactionsResponse struct {
	Transactions []Transaction `json:"transactions,omitempty"`
}

// Transaction represents a transaction
type Transaction struct {
	TransactionID string `json:"transactionId,omitempty"`
	Type          string `json:"type,omitempty"`
	SubType       string `json:"subType,omitempty"`
	Amount        string `json:"amount,omitempty"`
	Description   string `json:"description,omitempty"`
	Date          string `json:"date,omitempty"`
}

// TransactionDetailsResponse represents response for transaction details
type TransactionDetailsResponse struct {
	Transaction *Transaction `json:"transaction,omitempty"`
}

// PreferencesResponse represents response for user preferences
type PreferencesResponse struct {
	StreamerInfo *StreamerInfo `json:"streamerInfo,omitempty"`
}

// StreamerInfo represents streaming information
type StreamerInfo struct {
	AccountID      string `json:"accountId,omitempty"`
	AccountIDType  string `json:"accountIdType,omitempty"`
	Token          string `json:"token,omitempty"`
	TokenTimestamp string `json:"tokenTimestamp,omitempty"`
	UserID         string `json:"userId,omitempty"`
	AppID          string `json:"appId,omitempty"`
	Secret         string `json:"secret,omitempty"`
	AccessLevel    string `json:"accessLevel,omitempty"`
}

// QuotesResponse represents response for quotes
type QuotesResponse struct {
	Quotes map[string]*Quote `json:"quotes,omitempty"`
}

// Quote represents a quote
type Quote struct {
	Symbol            string  `json:"symbol,omitempty"`
	Description       string  `json:"description,omitempty"`
	LastPrice         float64 `json:"lastPrice,omitempty"`
	BidPrice          float64 `json:"bidPrice,omitempty"`
	AskPrice          float64 `json:"askPrice,omitempty"`
	BidSize           int     `json:"bidSize,omitempty"`
	AskSize           int     `json:"askSize,omitempty"`
	Volume            int     `json:"volume,omitempty"`
	OpenPrice         float64 `json:"openPrice,omitempty"`
	HighPrice         float64 `json:"highPrice,omitempty"`
	LowPrice          float64 `json:"lowPrice,omitempty"`
	PreviousClose     float64 `json:"previousClose,omitempty"`
	Change            float64 `json:"change,omitempty"`
	PercentChange     float64 `json:"percentChange,omitempty"`
	TotalVolume       int     `json:"totalVolume,omitempty"`
	TradeTime         string  `json:"tradeTime,omitempty"`
	QuoteTime         string  `json:"quoteTime,omitempty"`
	Mark              float64 `json:"mark,omitempty"`
	MarkChange        float64 `json:"markChange,omitempty"`
	MarkPercentChange float64 `json:"markPercentChange,omitempty"`
}

// QuoteResponse represents response for a single quote
type QuoteResponse struct {
	Quote *Quote `json:"quote,omitempty"`
}

// OptionChainsResponse represents response for option chains
type OptionChainsResponse struct {
	Symbol            string                                  `json:"symbol,omitempty"`
	Status            string                                  `json:"status,omitempty"`
	Strategy          string                                  `json:"strategy,omitempty"`
	Interval          float64                                 `json:"interval,omitempty"`
	IsDelayed         bool                                    `json:"isDelayed,omitempty"`
	IsIndex           bool                                    `json:"isIndex,omitempty"`
	InterestRate      float64                                 `json:"interestRate,omitempty"`
	UnderlyingPrice   float64                                 `json:"underlyingPrice,omitempty"`
	Volatility        float64                                 `json:"volatility,omitempty"`
	DaysToExpiration  int                                     `json:"daysToExpiration,omitempty"`
	NumberOfContracts int                                     `json:"numberOfContracts,omitempty"`
	CallExpDateMap    map[string]map[string][]*OptionContract `json:"callExpDateMap,omitempty"`
	PutExpDateMap     map[string]map[string][]*OptionContract `json:"putExpDateMap,omitempty"`
}

// OptionContract represents an option contract
type OptionContract struct {
	Symbol            string  `json:"symbol,omitempty"`
	Description       string  `json:"description,omitempty"`
	ExchangeName      string  `json:"exchangeName,omitempty"`
	BidPrice          float64 `json:"bidPrice,omitempty"`
	AskPrice          float64 `json:"askPrice,omitempty"`
	LastPrice         float64 `json:"lastPrice,omitempty"`
	BidSize           int     `json:"bidSize,omitempty"`
	AskSize           int     `json:"askSize,omitempty"`
	LastSize          int     `json:"lastSize,omitempty"`
	OpenPrice         float64 `json:"openPrice,omitempty"`
	HighPrice         float64 `json:"highPrice,omitempty"`
	LowPrice          float64 `json:"lowPrice,omitempty"`
	Volume            int     `json:"volume,omitempty"`
	OpenInterest      int     `json:"openInterest,omitempty"`
	TotalVolume       int     `json:"totalVolume,omitempty"`
	TradeTime         string  `json:"tradeTime,omitempty"`
	TradingStatus     string  `json:"tradingStatus,omitempty"`
	StrikePrice       float64 `json:"strikePrice,omitempty"`
	ExpirationDate    string  `json:"expirationDate,omitempty"`
	DaysToExpiration  int     `json:"daysToExpiration,omitempty"`
	ExpirationType    string  `json:"expirationType,omitempty"`
	OptionType        string  `json:"optionType,omitempty"`
	ContractSize      int     `json:"contractSize,omitempty"`
	DeliverableType   string  `json:"deliverableType,omitempty"`
	Mark              float64 `json:"mark,omitempty"`
	MarkChange        float64 `json:"markChange,omitempty"`
	MarkPercentChange float64 `json:"markPercentChange,omitempty"`
	PercentChange     float64 `json:"percentChange,omitempty"`
	PercentChangeBid  float64 `json:"percentChangeBid,omitempty"`
	PercentChangeAsk  float64 `json:"percentChangeAsk,omitempty"`
	PercentChangeLast float64 `json:"percentChangeLast,omitempty"`
	ImpliedVolatility float64 `json:"impliedVolatility,omitempty"`
	InTheMoney        bool    `json:"inTheMoney,omitempty"`
	NearTheMoney      bool    `json:"nearTheMoney,omitempty"`
}

// OptionExpirationChainResponse represents response for option expiration chain
type OptionExpirationChainResponse struct {
	ExpirationList []Expiration `json:"expirationList,omitempty"`
}

// Expiration represents an expiration date
type Expiration struct {
	ExpirationDate   string `json:"expirationDate,omitempty"`
	ExpirationType   string `json:"expirationType,omitempty"`
	DaysToExpiration int    `json:"daysToExpiration,omitempty"`
}

// PriceHistoryResponse represents response for price history
type PriceHistoryResponse struct {
	Symbol  string         `json:"symbol,omitempty"`
	Status  string         `json:"status,omitempty"`
	Error   *ErrorResponse `json:"error,omitempty"`
	Candles []Candle       `json:"candles,omitempty"`
}

// Candle represents a price history candle
type Candle struct {
	Open     float64 `json:"open,omitempty"`
	High     float64 `json:"high,omitempty"`
	Low      float64 `json:"low,omitempty"`
	Close    float64 `json:"close,omitempty"`
	Volume   int     `json:"volume,omitempty"`
	Datetime int64   `json:"datetime,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

// MoversResponse represents response for movers
type MoversResponse struct {
	Symbol string  `json:"symbol,omitempty"`
	Movers []Mover `json:"movers,omitempty"`
}

// Mover represents a mover
type Mover struct {
	Symbol        string  `json:"symbol,omitempty"`
	Description   string  `json:"description,omitempty"`
	LastPrice     float64 `json:"lastPrice,omitempty"`
	Change        float64 `json:"change,omitempty"`
	PercentChange float64 `json:"percentChange,omitempty"`
	TotalVolume   int     `json:"totalVolume,omitempty"`
}

// MarketHoursResponse represents response for market hours
type MarketHoursResponse struct {
	MarketHours map[string]*MarketHourInfo `json:"marketHours,omitempty"`
}

// MarketHourInfo represents market hour information
type MarketHourInfo struct {
	MarketType   string        `json:"marketType,omitempty"`
	IsOpen       bool          `json:"isOpen,omitempty"`
	SessionHours *SessionHours `json:"sessionHours,omitempty"`
}

// SessionHours represents session hours
type SessionHours struct {
	PreMarket     *Hours `json:"preMarket,omitempty"`
	RegularMarket *Hours `json:"regularMarket,omitempty"`
	PostMarket    *Hours `json:"postMarket,omitempty"`
}

// Hours represents time range
type Hours struct {
	Start string `json:"start,omitempty"`
	End   string `json:"end,omitempty"`
}

// MarketHourResponse represents response for a single market hour
type MarketHourResponse struct {
	MarketHourInfo *MarketHourInfo `json:"marketHourInfo,omitempty"`
}

// InstrumentsResponse represents response for instruments
type InstrumentsResponse struct {
	Instruments []Instrument `json:"instruments,omitempty"`
}

// Instrument represents an instrument
type Instrument struct {
	Symbol         string  `json:"symbol,omitempty"`
	Description    string  `json:"description,omitempty"`
	Exchange       string  `json:"exchange,omitempty"`
	AssetType      string  `json:"assetType,omitempty"`
	SymbolType     string  `json:"symbolType,omitempty"`
	InstrumentType string  `json:"instrumentType,omitempty"`
	IsDeliverable  bool    `json:"isDeliverable,omitempty"`
	HighPrice      float64 `json:"highPrice,omitempty"`
	LowPrice       float64 `json:"lowPrice,omitempty"`
	MinPrice       float64 `json:"minPrice,omitempty"`
	MaxPrice       float64 `json:"maxPrice,omitempty"`
	PriceIncrement float64 `json:"priceIncrement,omitempty"`
	ContractSize   int     `json:"contractSize,omitempty"`
	ExpirationDate string  `json:"expirationDate,omitempty"`
	OptionType     string  `json:"optionType,omitempty"`
	StrikePrice    float64 `json:"strikePrice,omitempty"`
	Industry       string  `json:"industry,omitempty"`
	Sector         string  `json:"sector,omitempty"`
}

// InstrumentCusipResponse represents response for instrument by CUSIP
type InstrumentCusipResponse struct {
	Instrument *Instrument `json:"instrument,omitempty"`
}
