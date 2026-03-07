package schwabdev

// ============================================================================
// ACCOUNTS & TRADING API RESPONSE TYPES
// ============================================================================

// LinkedAccount represents a single linked account with its hash value
type LinkedAccount struct {
	AccountNumber string `json:"accountNumber"`
	HashValue     string `json:"hashValue"`
}

// LinkedAccountsResponse is the response for GET /trader/v1/accounts/accountNumbers
type LinkedAccountsResponse []LinkedAccount

// AccountDetailsAllResponse is the response for GET /trader/v1/accounts/
type AccountDetailsAllResponse struct {
	SecuritiesAccount *SecuritiesAccount `json:"securitiesAccount,omitempty"`
	AggregatedBalance *AggregatedBalance `json:"aggregatedBalance,omitempty"`
}

// AccountDetailsResponse is the response for GET /trader/v1/accounts/{accountHash}
type AccountDetailsResponse AccountDetailsAllResponse

// SecuritiesAccount represents detailed account information
type SecuritiesAccount struct {
	Type                    string             `json:"type"`
	AccountNumber           string             `json:"accountNumber"`
	RoundTrips              int                `json:"roundTrips"`
	IsDayTrader             bool               `json:"isDayTrader"`
	IsClosingOnlyRestricted bool               `json:"isClosingOnlyRestricted"`
	PfcbFlag                bool               `json:"pfcbFlag"`
	InitialBalances         *InitialBalances   `json:"initialBalances,omitempty"`
	CurrentBalances         *CurrentBalances   `json:"currentBalances,omitempty"`
	ProjectedBalances       *ProjectedBalances `json:"projectedBalances,omitempty"`
	Positions               []*Position        `json:"positions,omitempty"`
}

// InitialBalances represents opening balance information
type InitialBalances struct {
	AccruedInterest                  float64 `json:"accruedInterest"`
	AvailableFundsNonMarginableTrade float64 `json:"availableFundsNonMarginableTrade"`
	BondValue                        float64 `json:"bondValue"`
	BuyingPower                      float64 `json:"buyingPower"`
	CashBalance                      float64 `json:"cashBalance"`
	CashAvailableForTrading          float64 `json:"cashAvailableForTrading"`
	CashReceipts                     float64 `json:"cashReceipts"`
	DayTradingBuyingPower            float64 `json:"dayTradingBuyingPower"`
	DayTradingBuyingPowerCall        float64 `json:"dayTradingBuyingPowerCall"`
	DayTradingEquityCall             float64 `json:"dayTradingEquityCall"`
	Equity                           float64 `json:"equity"`
	EquityPercentage                 float64 `json:"equityPercentage"`
	LiquidationValue                 float64 `json:"liquidationValue"`
	LongMarginValue                  float64 `json:"longMarginValue"`
	LongOptionMarketValue            float64 `json:"longOptionMarketValue"`
	LongStockValue                   float64 `json:"longStockValue"`
	MaintenanceCall                  float64 `json:"maintenanceCall"`
	MaintenanceRequirement           float64 `json:"maintenanceRequirement"`
	Margin                           float64 `json:"margin"`
	MarginEquity                     float64 `json:"marginEquity"`
	MoneyMarketFund                  float64 `json:"moneyMarketFund"`
	MutualFundValue                  float64 `json:"mutualFundValue"`
	RegTCall                         float64 `json:"regTCall"`
	ShortMarginValue                 float64 `json:"shortMarginValue"`
	ShortOptionMarketValue           float64 `json:"shortOptionMarketValue"`
	ShortStockValue                  float64 `json:"shortStockValue"`
	TotalCash                        float64 `json:"totalCash"`
	IsInCall                         bool    `json:"isInCall"`
	PendingDeposits                  float64 `json:"pendingDeposits"`
	MarginBalance                    float64 `json:"marginBalance"`
	ShortBalance                     float64 `json:"shortBalance"`
	AccountValue                     float64 `json:"accountValue"`
}

// CurrentBalances represents current account balances
type CurrentBalances struct {
	AccruedInterest                  float64 `json:"accruedInterest"`
	CashBalance                      float64 `json:"cashBalance"`
	CashReceipts                     float64 `json:"cashReceipts"`
	LongOptionMarketValue            float64 `json:"longOptionMarketValue"`
	LiquidationValue                 float64 `json:"liquidationValue"`
	LongMarketValue                  float64 `json:"longMarketValue"`
	MoneyMarketFund                  float64 `json:"moneyMarketFund"`
	Savings                          float64 `json:"savings"`
	ShortMarketValue                 float64 `json:"shortMarketValue"`
	PendingDeposits                  float64 `json:"pendingDeposits"`
	MutualFundValue                  float64 `json:"mutualFundValue"`
	BondValue                        float64 `json:"bondValue"`
	ShortOptionMarketValue           float64 `json:"shortOptionMarketValue"`
	AvailableFunds                   float64 `json:"availableFunds"`
	AvailableFundsNonMarginableTrade float64 `json:"availableFundsNonMarginableTrade"`
	BuyingPower                      float64 `json:"buyingPower"`
	BuyingPowerNonMarginableTrade    float64 `json:"buyingPowerNonMarginableTrade"`
	DayTradingBuyingPower            float64 `json:"dayTradingBuyingPower"`
	Equity                           float64 `json:"equity"`
	EquityPercentage                 float64 `json:"equityPercentage"`
	LongMarginValue                  float64 `json:"longMarginValue"`
	MaintenanceCall                  float64 `json:"maintenanceCall"`
	MaintenanceRequirement           float64 `json:"maintenanceRequirement"`
	MarginBalance                    float64 `json:"marginBalance"`
	RegTCall                         float64 `json:"regTCall"`
	ShortBalance                     float64 `json:"shortBalance"`
	ShortMarginValue                 float64 `json:"shortMarginValue"`
	Sma                              float64 `json:"sma"`
}

// ProjectedBalances represents projected balances after pending transactions
type ProjectedBalances struct {
	AvailableFunds                   float64 `json:"availableFunds"`
	AvailableFundsNonMarginableTrade float64 `json:"availableFundsNonMarginableTrade"`
	BuyingPower                      float64 `json:"buyingPower"`
	DayTradingBuyingPower            float64 `json:"dayTradingBuyingPower"`
	DayTradingBuyingPowerCall        float64 `json:"dayTradingBuyingPowerCall"`
	MaintenanceCall                  float64 `json:"maintenanceCall"`
	RegTCall                         float64 `json:"regTCall"`
	IsInCall                         bool    `json:"isInCall"`
	StockBuyingPower                 float64 `json:"stockBuyingPower"`
}

// AggregatedBalance represents aggregate balance across accounts
type AggregatedBalance struct {
	CurrentLiquidationValue float64 `json:"currentLiquidationValue"`
	LiquidationValue        float64 `json:"liquidationValue"`
}

// Position represents a position held in an account
type Position struct {
	ShortQuantity                float64 `json:"shortQuantity"`
	AveragePrice                 float64 `json:"averagePrice"`
	MarketValue                  float64 `json:"marketValue"`
	LongQuantity                 float64 `json:"longQuantity"`
	PreviousSessionLongQuantity  float64 `json:"previousSessionLongQuantity"`
	PreviousSessionShortQuantity float64 `json:"previousSessionShortQuantity"`
	ChangedSinceLastSession      bool    `json:"changedSinceLastSession"`
	AssetType                    string  `json:"assetType"`
	Cusip                        string  `json:"cusip"`
	Symbol                       string  `json:"symbol"`
	InstrumentID                 int64   `json:"instrumentId"`
}

// AccountOrdersResponse is the response for GET /trader/v1/accounts/{accountHash}/orders
type AccountOrdersResponse []Order

// Order represents an order object
type Order struct {
	Session                  string           `json:"session"`
	Duration                 string           `json:"duration"`
	OrderType                string           `json:"orderType"`
	CancelTime               *string          `json:"cancelTime,omitempty"`
	ComplexOrderStrategyType string           `json:"complexOrderStrategyType"`
	Quantity                 float64          `json:"quantity"`
	FilledQuantity           float64          `json:"filledQuantity"`
	RemainingQuantity        float64          `json:"remainingQuantity"`
	RequestedDestination     string           `json:"requestedDestination"`
	DestinationLinkName      string           `json:"destinationLinkName"`
	Price                    float64          `json:"price"`
	OrderLegCollection       []*OrderLeg      `json:"orderLegCollection"`
	OrderStrategyType        string           `json:"orderStrategyType"`
	OrderID                  int64            `json:"orderId"`
	Cancelable               bool             `json:"cancelable"`
	Editable                 bool             `json:"editable"`
	Status                   string           `json:"status"`
	EnteredTime              string           `json:"enteredTime"`
	CloseTime                *string          `json:"closeTime,omitempty"`
	Tag                      *string          `json:"tag,omitempty"`
	AccountNumber            int64            `json:"accountNumber"`
	OrderActivityCollection  []*OrderActivity `json:"orderActivityCollection,omitempty"`
}

// OrderLeg represents a leg of an order
type OrderLeg struct {
	OrderLegType   string      `json:"orderLegType"`
	LegID          int         `json:"legId"`
	Instrument     *Instrument `json:"instrument"`
	Instruction    string      `json:"instruction"`
	PositionEffect string      `json:"positionEffect"`
	Quantity       float64     `json:"quantity"`
}

// Instrument represents a financial instrument
type Instrument struct {
	AssetType    string `json:"assetType"`
	Cusip        string `json:"cusip,omitempty"`
	Symbol       string `json:"symbol"`
	InstrumentID int64  `json:"instrumentId,omitempty"`
}

// OrderActivity represents order execution activity
type OrderActivity struct {
	ActivityType           string          `json:"activityType"`
	ActivityID             int64           `json:"activityId"`
	ExecutionType          string          `json:"executionType"`
	Quantity               float64         `json:"quantity"`
	OrderRemainingQuantity float64         `json:"orderRemainingQuantity"`
	ExecutionLegs          []*ExecutionLeg `json:"executionLegs"`
}

// ExecutionLeg represents execution details for a leg
type ExecutionLeg struct {
	LegID             int     `json:"legId"`
	Quantity          float64 `json:"quantity"`
	MismarkedQuantity float64 `json:"mismarkedQuantity"`
	Price             float64 `json:"price"`
	Time              string  `json:"time"`
	InstrumentID      int64   `json:"instrumentId"`
}

// PlaceOrderResponse is the response for POST /trader/v1/accounts/{accountHash}/orders
// Note: Order ID is returned in the Location header, response body is empty
type PlaceOrderResponse struct {
	OrderID string // Extracted from Location header
}

// OrderDetailsResponse is the response for GET /trader/v1/accounts/{accountHash}/orders/{orderId}
type OrderDetailsResponse Order

// CancelOrderResponse is the response for DELETE /trader/v1/accounts/{accountHash}/orders/{orderId}
// Note: Empty response body on success (HTTP 200)
type CancelOrderResponse struct{}

// ReplaceOrderResponse is the response for PUT /trader/v1/accounts/{accountHash}/orders/{orderId}
// Note: Empty response body on success (HTTP 200)
type ReplaceOrderResponse struct{}

// AccountOrdersAllResponse is the response for GET /trader/v1/orders
type AccountOrdersAllResponse []Order

// PreviewOrderResponse is the response for POST /trader/v1/accounts/{accountHash}/previewOrder
type PreviewOrderResponse struct {
	OrderID               int64                  `json:"orderId"`
	OrderStrategy         *PreviewOrderStrategy  `json:"orderStrategy"`
	OrderValidationResult *OrderValidationResult `json:"orderValidationResult"`
	CommissionAndFee      *CommissionAndFee      `json:"commissionAndFee"`
}

// PreviewOrderStrategy represents the order strategy in preview
type PreviewOrderStrategy struct {
	AccountNumber          string             `json:"accountNumber"`
	AdvancedOrderType      string             `json:"advancedOrderType"`
	CloseTime              string             `json:"closeTime"`
	EnteredTime            string             `json:"enteredTime"`
	OrderBalance           *OrderBalance      `json:"orderBalance"`
	OrderStrategyType      string             `json:"orderStrategyType"`
	OrderVersion           int                `json:"orderVersion"`
	Session                string             `json:"session"`
	Status                 string             `json:"status"`
	Discretionary          bool               `json:"discretionary"`
	Duration               string             `json:"duration"`
	FilledQuantity         float64            `json:"filledQuantity"`
	OrderType              string             `json:"orderType"`
	OrderValue             float64            `json:"orderValue"`
	Price                  float64            `json:"price"`
	Quantity               float64            `json:"quantity"`
	RemainingQuantity      float64            `json:"remainingQuantity"`
	SellNonMarginableFirst bool               `json:"sellNonMarginableFirst"`
	Strategy               string             `json:"strategy"`
	AmountIndicator        string             `json:"amountIndicator"`
	OrderLegs              []*PreviewOrderLeg `json:"orderLegs"`
}

// OrderBalance represents order balance information
type OrderBalance struct {
	OrderValue             float64 `json:"orderValue"`
	ProjectedAvailableFund float64 `json:"projectedAvailableFund"`
	ProjectedBuyingPower   float64 `json:"projectedBuyingPower"`
	ProjectedCommission    float64 `json:"projectedCommission"`
}

// PreviewOrderLeg represents a leg in preview order
type PreviewOrderLeg struct {
	AskPrice            float64     `json:"askPrice"`
	BidPrice            float64     `json:"bidPrice"`
	LastPrice           float64     `json:"lastPrice"`
	MarkPrice           float64     `json:"markPrice"`
	ProjectedCommission float64     `json:"projectedCommission"`
	FinalSymbol         string      `json:"finalSymbol"`
	LegID               int         `json:"legId"`
	AssetType           string      `json:"assetType"`
	Instruction         string      `json:"instruction"`
	PositionEffect      string      `json:"positionEffect"`
	Instrument          *Instrument `json:"instrument"`
}

// OrderValidationResult represents validation results for an order
type OrderValidationResult struct {
	Rejects []*OrderReject `json:"rejects"`
}

// OrderReject represents a rejection reason
type OrderReject struct {
	ActivityMessage  string `json:"activityMessage"`
	OriginalSeverity string `json:"originalSeverity"`
}

// CommissionAndFee represents commission and fee details
type CommissionAndFee struct {
	Commission     *Commission `json:"commission"`
	Fee            *Fee        `json:"fee"`
	TrueCommission *Commission `json:"trueCommission"`
}

// Commission represents commission details
type Commission struct {
	CommissionLegs []*CommissionLeg `json:"commissionLegs"`
}

// CommissionLeg represents a commission leg
type CommissionLeg struct {
	CommissionValues []*CommissionValue `json:"commissionValues"`
}

// CommissionValue represents a commission value
type CommissionValue struct {
	Value float64 `json:"value"`
	Type  string  `json:"type"`
}

// Fee represents fee details
type Fee struct {
	FeeLegs []*FeeLeg `json:"feeLegs"`
}

// FeeLeg represents a fee leg
type FeeLeg struct {
	FeeValues []*FeeValue `json:"feeValues"`
}

// FeeValue represents a fee value
type FeeValue struct {
	Value float64 `json:"value"`
	Type  string  `json:"type"`
}

// TransactionsResponse is the response for GET /trader/v1/accounts/{accountHash}/transactions
type TransactionsResponse []Transaction

// Transaction represents a transaction
type Transaction struct {
	TransactionID string  `json:"transactionId"`
	Type          string  `json:"type"`
	Symbol        string  `json:"symbol"`
	Date          string  `json:"date"`
	Quantity      float64 `json:"quantity"`
	Price         float64 `json:"price"`
	NetAmount     float64 `json:"netAmount"`
}

// TransactionDetailsResponse is the response for GET /trader/v1/accounts/{accountHash}/transactions/{transactionId}
type TransactionDetailsResponse Transaction

// PreferencesResponse is the response for GET /trader/v1/userPreference
type PreferencesResponse struct {
	StreamerInfo []*StreamerInfo `json:"streamerInfo,omitempty"`
}

// StreamerInfo represents streamer configuration
type StreamerInfo struct {
	StreamerURL            string `json:"streamerUrl"`
	SchwabClientCorrelID   string `json:"schwabClientCorrelId"`
	SchwabClientChannel    string `json:"schwabClientChannel"`
	SchwabClientFunctionID string `json:"schwabClientFunctionId"`
}

// ============================================================================
// MARKET DATA API RESPONSE TYPES
// ============================================================================

// QuotesResponse is the response for GET /marketdata/v1/quotes
type QuotesResponse map[string]Quote

// QuoteResponse is the response for GET /marketdata/v1/{symbol_id}/quotes
type QuoteResponse Quote

// Quote represents a complete quote with all data sections
type Quote struct {
	AssetMainType string       `json:"assetMainType"`
	AssetSubType  string       `json:"assetSubType"`
	QuoteType     string       `json:"quoteType"`
	Realtime      bool         `json:"realtime"`
	Ssid          int64        `json:"ssid"`
	Symbol        string       `json:"symbol"`
	Fundamental   *Fundamental `json:"fundamental,omitempty"`
	QuoteData     *QuoteData   `json:"quote,omitempty"`
	Reference     *Reference   `json:"reference,omitempty"`
	Regular       *Regular     `json:"regular,omitempty"`
}

// Fundamental represents fundamental data
type Fundamental struct {
	Avg10DaysVolume    float64 `json:"avg10DaysVolume"`
	Avg1YearVolume     float64 `json:"avg1YearVolume"`
	DeclarationDate    string  `json:"declarationDate,omitempty"`
	DivAmount          float64 `json:"divAmount"`
	DivExDate          string  `json:"divExDate,omitempty"`
	DivFreq            int     `json:"divFreq"`
	DivPayAmount       float64 `json:"divPayAmount"`
	DivPayDate         string  `json:"divPayDate,omitempty"`
	DivYield           float64 `json:"divYield"`
	Eps                float64 `json:"eps"`
	FundLeverageFactor float64 `json:"fundLeverageFactor"`
	LastEarningsDate   string  `json:"lastEarningsDate,omitempty"`
	NextDivExDate      string  `json:"nextDivExDate,omitempty"`
	NextDivPayDate     string  `json:"nextDivPayDate,omitempty"`
	PeRatio            float64 `json:"peRatio"`
}

// QuoteData represents real-time quote data
type QuoteData struct {
	FiftyTwoWeekHigh        float64 `json:"52WeekHigh"`
	FiftyTwoWeekLow         float64 `json:"52WeekLow"`
	AskMICId                string  `json:"askMICId,omitempty"`
	AskPrice                float64 `json:"askPrice"`
	AskSize                 int     `json:"askSize"`
	AskTime                 int64   `json:"askTime"`
	BidMICId                string  `json:"bidMICId,omitempty"`
	BidPrice                float64 `json:"bidPrice"`
	BidSize                 int     `json:"bidSize"`
	BidTime                 int64   `json:"bidTime"`
	ClosePrice              float64 `json:"closePrice"`
	HighPrice               float64 `json:"highPrice"`
	LastMICId               string  `json:"lastMICId,omitempty"`
	LastPrice               float64 `json:"lastPrice"`
	LastSize                int     `json:"lastSize"`
	LowPrice                float64 `json:"lowPrice"`
	Mark                    float64 `json:"mark"`
	MarkChange              float64 `json:"markChange"`
	MarkPercentChange       float64 `json:"markPercentChange"`
	NetChange               float64 `json:"netChange"`
	NetPercentChange        float64 `json:"netPercentChange"`
	OpenPrice               float64 `json:"openPrice"`
	PostMarketChange        float64 `json:"postMarketChange"`
	PostMarketPercentChange float64 `json:"postMarketPercentChange"`
	QuoteTime               int64   `json:"quoteTime"`
	SecurityStatus          string  `json:"securityStatus"`
	TotalVolume             int64   `json:"totalVolume"`
	TradeTime               int64   `json:"tradeTime"`
}

// Reference represents reference data
type Reference struct {
	Cusip          string  `json:"cusip"`
	Description    string  `json:"description"`
	Exchange       string  `json:"exchange"`
	ExchangeName   string  `json:"exchangeName"`
	IsHardToBorrow bool    `json:"isHardToBorrow"`
	IsShortable    bool    `json:"isShortable"`
	HtbQuantity    int64   `json:"htbQuantity"`
	HtbRate        float64 `json:"htbRate"`
}

// Regular represents regular market trading data
type Regular struct {
	RegularMarketLastPrice     float64 `json:"regularMarketLastPrice"`
	RegularMarketLastSize      int64   `json:"regularMarketLastSize"`
	RegularMarketNetChange     float64 `json:"regularMarketNetChange"`
	RegularMarketPercentChange float64 `json:"regularMarketPercentChange"`
	RegularMarketTradeTime     int64   `json:"regularMarketTradeTime"`
}

// OptionChainsResponse is the response for GET /marketdata/v1/chains
type OptionChainsResponse struct {
	Symbol            string                                 `json:"symbol"`
	Status            string                                 `json:"status"`
	Strategy          string                                 `json:"strategy"`
	Interval          float64                                `json:"interval"`
	IsDelayed         bool                                   `json:"isDelayed"`
	IsIndex           bool                                   `json:"isIndex"`
	InterestRate      float64                                `json:"interestRate"`
	UnderlyingPrice   float64                                `json:"underlyingPrice"`
	Volatility        float64                                `json:"volatility"`
	DaysToExpiration  float64                                `json:"daysToExpiration"`
	NumberOfContracts int                                    `json:"numberOfContracts"`
	AssetMainType     string                                 `json:"assetMainType"`
	AssetSubType      string                                 `json:"assetSubType"`
	IsChainTruncated  bool                                   `json:"isChainTruncated"`
	CallExpDateMap    map[string]map[string][]OptionContract `json:"callExpDateMap"`
	PutExpDateMap     map[string]map[string][]OptionContract `json:"putExpDateMap"`
}

// OptionContract represents an option contract
type OptionContract struct {
	PutCall                string               `json:"putCall"`
	Symbol                 string               `json:"symbol"`
	Description            string               `json:"description"`
	ExchangeName           string               `json:"exchangeName"`
	Bid                    float64              `json:"bid"`
	Ask                    float64              `json:"ask"`
	Last                   float64              `json:"last"`
	Mark                   float64              `json:"mark"`
	BidSize                int                  `json:"bidSize"`
	AskSize                int                  `json:"askSize"`
	BidAskSize             string               `json:"bidAskSize"`
	LastSize               int                  `json:"lastSize"`
	HighPrice              float64              `json:"highPrice"`
	LowPrice               float64              `json:"lowPrice"`
	OpenPrice              float64              `json:"openPrice"`
	ClosePrice             float64              `json:"closePrice"`
	TotalVolume            int                  `json:"totalVolume"`
	TradeTimeInLong        int64                `json:"tradeTimeInLong"`
	QuoteTimeInLong        int64                `json:"quoteTimeInLong"`
	NetChange              float64              `json:"netChange"`
	Volatility             float64              `json:"volatility"`
	Delta                  float64              `json:"delta"`
	Gamma                  float64              `json:"gamma"`
	Theta                  float64              `json:"theta"`
	Vega                   float64              `json:"vega"`
	Rho                    float64              `json:"rho"`
	OpenInterest           int                  `json:"openInterest"`
	TimeValue              float64              `json:"timeValue"`
	TheoreticalOptionValue float64              `json:"theoreticalOptionValue"`
	TheoreticalVolatility  float64              `json:"theoreticalVolatility"`
	OptionDeliverablesList []*OptionDeliverable `json:"optionDeliverablesList"`
	StrikePrice            float64              `json:"strikePrice"`
	ExpirationDate         string               `json:"expirationDate"`
	DaysToExpiration       int                  `json:"daysToExpiration"`
	ExpirationType         string               `json:"expirationType"`
	LastTradingDay         int64                `json:"lastTradingDay"`
	Multiplier             float64              `json:"multiplier"`
	SettlementType         string               `json:"settlementType"`
	DeliverableNote        string               `json:"deliverableNote"`
	PercentChange          float64              `json:"percentChange"`
	MarkChange             float64              `json:"markChange"`
	MarkPercentChange      float64              `json:"markPercentChange"`
	IntrinsicValue         float64              `json:"intrinsicValue"`
	ExtrinsicValue         float64              `json:"extrinsicValue"`
	OptionRoot             string               `json:"optionRoot"`
	ExerciseType           string               `json:"exerciseType"`
	High52Week             float64              `json:"high52Week"`
	Low52Week              float64              `json:"low52Week"`
	NonStandard            bool                 `json:"nonStandard"`
	PennyPilot             bool                 `json:"pennyPilot"`
	InTheMoney             bool                 `json:"inTheMoney"`
	Mini                   bool                 `json:"mini"`
}

// OptionDeliverable represents option deliverable information
type OptionDeliverable struct {
	Symbol           string  `json:"symbol"`
	AssetType        string  `json:"assetType"`
	DeliverableUnits float64 `json:"deliverableUnits"`
}

// OptionExpirationChainResponse is the response for GET /marketdata/v1/expirationchain
type OptionExpirationChainResponse struct {
	ExpirationList []*ExpirationDate `json:"expirationList"`
}

// ExpirationDate represents an option expiration date
type ExpirationDate struct {
	ExpirationDate   string `json:"expirationDate"`
	DaysToExpiration int    `json:"daysToExpiration"`
	ExpirationType   string `json:"expirationType"`
	SettlementType   string `json:"settlementType"`
	OptionRoots      string `json:"optionRoots"`
	Standard         bool   `json:"standard"`
}

// PriceHistoryResponse is the response for GET /marketdata/v1/pricehistory
type PriceHistoryResponse struct {
	Candles []*Candle `json:"candles"`
	Symbol  string    `json:"symbol"`
	Empty   bool      `json:"empty"`
}

// Candle represents a price history candle
type Candle struct {
	Open     float64 `json:"open"`
	High     float64 `json:"high"`
	Low      float64 `json:"low"`
	Close    float64 `json:"close"`
	Volume   int64   `json:"volume"`
	Datetime int64   `json:"datetime"`
}

// MoversResponse is the response for GET /marketdata/v1/movers/{symbol}
type MoversResponse []Mover

// Mover represents a market mover
type Mover struct {
	Symbol        string  `json:"symbol"`
	Description   string  `json:"description"`
	LastPrice     float64 `json:"lastPrice"`
	Change        float64 `json:"change"`
	PercentChange float64 `json:"percentChange"`
	Volume        int64   `json:"volume"`
}

// MarketHoursResponse is the response for GET /marketdata/v1/markets
type MarketHoursResponse map[string]MarketHour

// MarketHourResponse is the response for GET /marketdata/v1/markets/{market_id}
type MarketHourResponse MarketHour

// MarketHour represents market hours for a specific market
type MarketHour struct {
	Category     string        `json:"category"`
	Date         string        `json:"date"`
	Exchange     string        `json:"exchange"`
	IsOpen       bool          `json:"isOpen"`
	MarketType   string        `json:"marketType"`
	Product      string        `json:"product"`
	ProductName  string        `json:"productName"`
	SessionHours *SessionHours `json:"sessionHours,omitempty"`
}

// SessionHours represents market session hours
type SessionHours struct {
	SessionDuration []*SessionDuration `json:"sessionDuration,omitempty"`
	StartEndTime    []*StartEndTime    `json:"startEndTime,omitempty"`
}

// SessionDuration represents a session duration
type SessionDuration struct {
	StartDateTime string `json:"startDateTime"`
	EndDateTime   string `json:"endDateTime"`
}

// StartEndTime represents start and end times
type StartEndTime struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// InstrumentsResponse is the response for GET /marketdata/v1/instruments
type InstrumentsResponse []InstrumentSearch

// InstrumentSearch represents an instrument search result
type InstrumentSearch struct {
	Symbol      string `json:"symbol"`
	Description string `json:"description"`
	AssetType   string `json:"assetType"`
	Cusip       string `json:"cusip"`
	Exchange    string `json:"exchange"`
}

// InstrumentCUSIPResponse is the response for GET /marketdata/v1/instruments/{cusip_id}
type InstrumentCUSIPResponse struct {
	Instruments []*InstrumentCUSIP `json:"instruments"`
}

// InstrumentCUSIP represents an instrument by CUSIP
type InstrumentCUSIP struct {
	Cusip       string `json:"cusip"`
	Symbol      string `json:"symbol"`
	Description string `json:"description"`
	Exchange    string `json:"exchange"`
	AssetType   string `json:"assetType"`
}

// ============================================================================
// REQUEST TYPES FOR POST/PUT OPERATIONS
// ============================================================================

// OrderRequest represents an order request for place_order and replace_order
type OrderRequest struct {
	OrderType                string             `json:"orderType"`
	Session                  string             `json:"session"`
	Duration                 string             `json:"duration"`
	OrderStrategyType        string             `json:"orderStrategyType"`
	Price                    string             `json:"price,omitempty"`
	StopPrice                string             `json:"stopPrice,omitempty"`
	OrderLegCollection       []*OrderLegRequest `json:"orderLegCollection"`
	ComplexOrderStrategyType string             `json:"complexOrderStrategyType,omitempty"`
	Quantity                 float64            `json:"quantity,omitempty"`
}

// OrderLegRequest represents a leg in an order request
type OrderLegRequest struct {
	Instruction string             `json:"instruction"`
	Quantity    int                `json:"quantity"`
	Instrument  *InstrumentRequest `json:"instrument"`
}

// InstrumentRequest represents an instrument in an order request
type InstrumentRequest struct {
	Symbol    string `json:"symbol"`
	AssetType string `json:"assetType"`
}

// PreviewOrderRequest represents a preview order request
type PreviewOrderRequest OrderRequest
