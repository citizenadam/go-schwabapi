package schwabdev

import "encoding/json"

// ============================================================================
// ACCOUNTS & TRADING API RESPONSE TYPES
// Aligned with Schwab Trader API OpenAPI specification.
// ============================================================================

// AccountNumberHash represents a linked account with its hash value.
// GET /trader/v1/accounts/accountNumbers
type AccountNumberHash struct {
	AccountNumber string `json:"accountNumber"`
	HashValue     string `json:"hashValue"`
}

// LinkedAccountsResponse is the response for GET /trader/v1/accounts/accountNumbers
type LinkedAccountsResponse []AccountNumberHash

// LinkedAccount is an alias for AccountNumberHash (backward compatibility).
type LinkedAccount = AccountNumberHash

// Account is the wrapper returned by GET /trader/v1/accounts and
// GET /trader/v1/accounts/{accountNumber}.
type Account struct {
	SecuritiesAccount *SecuritiesAccount `json:"securitiesAccount,omitempty"`
	AggregatedBalance *AggregatedBalance `json:"aggregatedBalance,omitempty"`
}

// AccountDetailsAllResponse is the response for GET /trader/v1/accounts/
type AccountDetailsAllResponse = Account

// AccountDetailsResponse is the response for GET /trader/v1/accounts/{accountHash}
type AccountDetailsResponse = Account

// SecuritiesAccount is a oneOf discriminator on the "type" field.
// It embeds SecuritiesAccountBase and is populated as either a MarginAccount
// or CashAccount at decode time via the Type field.
type SecuritiesAccount struct {
	Type                    string          `json:"type"`
	AccountNumber           string          `json:"accountNumber"`
	RoundTrips              int             `json:"roundTrips"`
	IsDayTrader             bool            `json:"isDayTrader"`
	IsClosingOnlyRestricted bool            `json:"isClosingOnlyRestricted"`
	PfcbFlag                bool            `json:"pfcbFlag"`
	InitialBalances         json.RawMessage `json:"initialBalances,omitempty"`
	CurrentBalances         json.RawMessage `json:"currentBalances,omitempty"`
	ProjectedBalances       json.RawMessage `json:"projectedBalances,omitempty"`
	Positions               []*Position     `json:"positions,omitempty"`
}

// SecuritiesAccountBase corresponds to the OpenAPI SecuritiesAccountBase schema.
type SecuritiesAccountBase struct {
	Type                    string      `json:"type"`
	AccountNumber           string      `json:"accountNumber"`
	RoundTrips              int         `json:"roundTrips"`
	IsDayTrader             bool        `json:"isDayTrader"`
	IsClosingOnlyRestricted bool        `json:"isClosingOnlyRestricted"`
	PfcbFlag                bool        `json:"pfcbFlag"`
	Positions               []*Position `json:"positions,omitempty"`
}

// CashAccount extends SecuritiesAccountBase with cash-specific balances.
type CashAccount struct {
	SecuritiesAccountBase
	InitialBalances   *CashInitialBalance `json:"initialBalances,omitempty"`
	CurrentBalances   *CashBalance        `json:"currentBalances,omitempty"`
	ProjectedBalances *CashBalance        `json:"projectedBalances,omitempty"`
}

// MarginAccount extends SecuritiesAccountBase with margin-specific balances.
type MarginAccount struct {
	SecuritiesAccountBase
	InitialBalances   *MarginInitialBalance `json:"initialBalances,omitempty"`
	CurrentBalances   *MarginBalance        `json:"currentBalances,omitempty"`
	ProjectedBalances *MarginBalance        `json:"projectedBalances,omitempty"`
}

// CashInitialBalance represents opening balance information for a cash account.
type CashInitialBalance struct {
	AccruedInterest            float64 `json:"accruedInterest"`
	CashAvailableForTrading    float64 `json:"cashAvailableForTrading"`
	CashAvailableForWithdrawal float64 `json:"cashAvailableForWithdrawal"`
	CashBalance                float64 `json:"cashBalance"`
	BondValue                  float64 `json:"bondValue"`
	CashReceipts               float64 `json:"cashReceipts"`
	LiquidationValue           float64 `json:"liquidationValue"`
	LongOptionMarketValue      float64 `json:"longOptionMarketValue"`
	LongStockValue             float64 `json:"longStockValue"`
	MoneyMarketFund            float64 `json:"moneyMarketFund"`
	MutualFundValue            float64 `json:"mutualFundValue"`
	ShortOptionMarketValue     float64 `json:"shortOptionMarketValue"`
	ShortStockValue            float64 `json:"shortStockValue"`
	IsInCall                   float64 `json:"isInCall"`
	UnsettledCash              float64 `json:"unsettledCash"`
	CashDebitCallValue         float64 `json:"cashDebitCallValue"`
	PendingDeposits            float64 `json:"pendingDeposits"`
	AccountValue               float64 `json:"accountValue"`
}

// MarginInitialBalance represents opening balance information for a margin account.
type MarginInitialBalance struct {
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
	IsInCall                         float64 `json:"isInCall"`
	UnsettledCash                    float64 `json:"unsettledCash"`
	PendingDeposits                  float64 `json:"pendingDeposits"`
	MarginBalance                    float64 `json:"marginBalance"`
	ShortBalance                     float64 `json:"shortBalance"`
	AccountValue                     float64 `json:"accountValue"`
}

// InitialBalances is an alias for MarginInitialBalance (backward compat).
// The API returns MarginInitialBalance for margin accounts and CashInitialBalance
// for cash accounts.
type InitialBalances = MarginInitialBalance

// CashBalance represents current balances for a cash account.
type CashBalance struct {
	CashAvailableForTrading      float64 `json:"cashAvailableForTrading"`
	CashAvailableForWithdrawal   float64 `json:"cashAvailableForWithdrawal"`
	CashCall                     float64 `json:"cashCall"`
	LongNonMarginableMarketValue float64 `json:"longNonMarginableMarketValue"`
	TotalCash                    float64 `json:"totalCash"`
	CashDebitCallValue           float64 `json:"cashDebitCallValue"`
	UnsettledCash                float64 `json:"unsettledCash"`
}

// MarginBalance represents current balances for a margin account.
type MarginBalance struct {
	AvailableFunds                   float64 `json:"availableFunds"`
	AvailableFundsNonMarginableTrade float64 `json:"availableFundsNonMarginableTrade"`
	BuyingPower                      float64 `json:"buyingPower"`
	BuyingPowerNonMarginableTrade    float64 `json:"buyingPowerNonMarginableTrade"`
	DayTradingBuyingPower            float64 `json:"dayTradingBuyingPower"`
	DayTradingBuyingPowerCall        float64 `json:"dayTradingBuyingPowerCall"`
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
	IsInCall                         float64 `json:"isInCall"`
	StockBuyingPower                 float64 `json:"stockBuyingPower"`
	OptionBuyingPower                float64 `json:"optionBuyingPower"`
}

// CurrentBalances is an alias for MarginBalance (backward compat).
type CurrentBalances = MarginBalance

// ProjectedBalances represents projected balances after pending transactions.
// Uses the same structure as current balances (MarginBalance or CashBalance
// depending on account type).
type ProjectedBalances = MarginBalance

// AggregatedBalance represents aggregate balance across accounts.
type AggregatedBalance struct {
	CurrentLiquidationValue float64 `json:"currentLiquidationValue"`
	LiquidationValue        float64 `json:"liquidationValue"`
}

// AccountsBaseInstrument is the base instrument for accounts responses.
type AccountsBaseInstrument struct {
	AssetType    string  `json:"assetType"`
	Cusip        string  `json:"cusip"`
	Symbol       string  `json:"symbol"`
	Description  string  `json:"description"`
	InstrumentID int64   `json:"instrumentId"`
	NetChange    float64 `json:"netChange"`
}

// Instrument represents a financial instrument (accounts context).
// This is the concrete type used in OrderLeg, Position, etc.
type Instrument struct {
	AssetType    string  `json:"assetType"`
	Cusip        string  `json:"cusip,omitempty"`
	Symbol       string  `json:"symbol"`
	Description  string  `json:"description,omitempty"`
	InstrumentID int64   `json:"instrumentId,omitempty"`
	NetChange    float64 `json:"netChange,omitempty"`
}

// AccountOption extends AccountsBaseInstrument with option-specific fields.
type AccountOption struct {
	AccountsBaseInstrument
	OptionDeliverables []*AccountAPIOptionDeliverable `json:"optionDeliverables,omitempty"`
	PutCall            string                         `json:"putCall,omitempty"`
	OptionMultiplier   int                            `json:"optionMultiplier,omitempty"`
	Type               string                         `json:"type,omitempty"`
	UnderlyingSymbol   string                         `json:"underlyingSymbol,omitempty"`
}

// AccountAPIOptionDeliverable represents an option deliverable in accounts.
type AccountAPIOptionDeliverable struct {
	Symbol           string  `json:"symbol"`
	DeliverableUnits float64 `json:"deliverableUnits"`
	APICurrencyType  string  `json:"apiCurrencyType,omitempty"`
	AssetType        string  `json:"assetType,omitempty"`
}

// AccountFixedIncome extends AccountsBaseInstrument with fixed-income fields.
type AccountFixedIncome struct {
	AccountsBaseInstrument
	MaturityDate string  `json:"maturityDate,omitempty"`
	Factor       float64 `json:"factor,omitempty"`
	VariableRate float64 `json:"variableRate,omitempty"`
}

// AccountCashEquivalent extends AccountsBaseInstrument.
type AccountCashEquivalent struct {
	AccountsBaseInstrument
	Type string `json:"type,omitempty"`
}

// Position represents a position held in an account.
type Position struct {
	ShortQuantity                  float64     `json:"shortQuantity"`
	AveragePrice                   float64     `json:"averagePrice"`
	CurrentDayProfitLoss           float64     `json:"currentDayProfitLoss"`
	CurrentDayProfitLossPercentage float64     `json:"currentDayProfitLossPercentage"`
	LongQuantity                   float64     `json:"longQuantity"`
	SettledLongQuantity            float64     `json:"settledLongQuantity"`
	SettledShortQuantity           float64     `json:"settledShortQuantity"`
	AgedQuantity                   float64     `json:"agedQuantity"`
	Instrument                     *Instrument `json:"instrument,omitempty"`
	MarketValue                    float64     `json:"marketValue"`
	MaintenanceRequirement         float64     `json:"maintenanceRequirement"`
	AverageLongPrice               float64     `json:"averageLongPrice"`
	AverageShortPrice              float64     `json:"averageShortPrice"`
	TaxLotAverageLongPrice         float64     `json:"taxLotAverageLongPrice"`
	TaxLotAverageShortPrice        float64     `json:"taxLotAverageShortPrice"`
	LongOpenProfitLoss             float64     `json:"longOpenProfitLoss"`
	ShortOpenProfitLoss            float64     `json:"shortOpenProfitLoss"`
	PreviousSessionLongQuantity    float64     `json:"previousSessionLongQuantity"`
	PreviousSessionShortQuantity   float64     `json:"previousSessionShortQuantity"`
	CurrentDayCost                 float64     `json:"currentDayCost"`
}

// AccountOrdersResponse is the response for GET /trader/v1/accounts/{accountHash}/orders
type AccountOrdersResponse []Order

// Order represents an order object as returned by the API.
type Order struct {
	Session                  string                `json:"session,omitempty"`
	Duration                 string                `json:"duration,omitempty"`
	OrderType                string                `json:"orderType,omitempty"`
	CancelTime               *string               `json:"cancelTime,omitempty"`
	ComplexOrderStrategyType string                `json:"complexOrderStrategyType,omitempty"`
	Quantity                 float64               `json:"quantity,omitempty"`
	FilledQuantity           float64               `json:"filledQuantity,omitempty"`
	RemainingQuantity        float64               `json:"remainingQuantity,omitempty"`
	RequestedDestination     string                `json:"requestedDestination,omitempty"`
	DestinationLinkName      string                `json:"destinationLinkName,omitempty"`
	ReleaseTime              *string               `json:"releaseTime,omitempty"`
	StopPrice                float64               `json:"stopPrice,omitempty"`
	StopPriceLinkBasis       string                `json:"stopPriceLinkBasis,omitempty"`
	StopPriceLinkType        string                `json:"stopPriceLinkType,omitempty"`
	StopPriceOffset          float64               `json:"stopPriceOffset,omitempty"`
	StopType                 string                `json:"stopType,omitempty"`
	PriceLinkBasis           string                `json:"priceLinkBasis,omitempty"`
	PriceLinkType            string                `json:"priceLinkType,omitempty"`
	Price                    float64               `json:"price,omitempty"`
	TaxLotMethod             string                `json:"taxLotMethod,omitempty"`
	OrderLegCollection       []*OrderLegCollection `json:"orderLegCollection,omitempty"`
	ActivationPrice          float64               `json:"activationPrice,omitempty"`
	SpecialInstruction       string                `json:"specialInstruction,omitempty"`
	OrderStrategyType        string                `json:"orderStrategyType,omitempty"`
	OrderID                  int64                 `json:"orderId,omitempty"`
	Cancelable               bool                  `json:"cancelable,omitempty"`
	Editable                 bool                  `json:"editable,omitempty"`
	Status                   string                `json:"status,omitempty"`
	EnteredTime              string                `json:"enteredTime,omitempty"`
	CloseTime                *string               `json:"closeTime,omitempty"`
	Tag                      *string               `json:"tag,omitempty"`
	AccountNumber            int64                 `json:"accountNumber,omitempty"`
	OrderActivityCollection  []*OrderActivity      `json:"orderActivityCollection,omitempty"`
	ReplacingOrderCollection []*Order              `json:"replacingOrderCollection,omitempty"`
	ChildOrderStrategies     []*Order              `json:"childOrderStrategies,omitempty"`
	StatusDescription        string                `json:"statusDescription,omitempty"`
}

// OrderLegCollection represents a leg in an order (both response and request).
// Used for complex strategies like Butterfly, Vertical Spreads, etc.
type OrderLegCollection struct {
	OrderLegType   string      `json:"orderLegType,omitempty"`
	LegID          int64       `json:"legId,omitempty"`
	Instrument     *Instrument `json:"instrument,omitempty"`
	Instruction    string      `json:"instruction"`
	PositionEffect string      `json:"positionEffect,omitempty"`
	Quantity       float64     `json:"quantity"`
	QuantityType   string      `json:"quantityType,omitempty"`
	DivCapGains    string      `json:"divCapGains,omitempty"`
	ToSymbol       string      `json:"toSymbol,omitempty"`
}

// OrderLeg represents a leg in an OrderStrategy (preview/strategy context).
type OrderLeg struct {
	AskPrice            float64 `json:"askPrice,omitempty"`
	BidPrice            float64 `json:"bidPrice,omitempty"`
	LastPrice           float64 `json:"lastPrice,omitempty"`
	MarkPrice           float64 `json:"markPrice,omitempty"`
	ProjectedCommission float64 `json:"projectedCommission,omitempty"`
	Quantity            float64 `json:"quantity,omitempty"`
	FinalSymbol         string  `json:"finalSymbol,omitempty"`
	LegID               int64   `json:"legId,omitempty"`
	AssetType           string  `json:"assetType,omitempty"`
	Instruction         string  `json:"instruction,omitempty"`
}

// OrderActivity represents order execution activity.
type OrderActivity struct {
	ActivityType           string          `json:"activityType,omitempty"`
	ExecutionType          string          `json:"executionType,omitempty"`
	Quantity               float64         `json:"quantity,omitempty"`
	OrderRemainingQuantity float64         `json:"orderRemainingQuantity,omitempty"`
	ExecutionLegs          []*ExecutionLeg `json:"executionLegs,omitempty"`
}

// ExecutionLeg represents execution details for a leg.
type ExecutionLeg struct {
	LegID             int64   `json:"legId,omitempty"`
	Price             float64 `json:"price,omitempty"`
	Quantity          float64 `json:"quantity,omitempty"`
	MismarkedQuantity float64 `json:"mismarkedQuantity,omitempty"`
	InstrumentID      int64   `json:"instrumentId,omitempty"`
	Time              string  `json:"time,omitempty"`
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

// ============================================================================
// PREVIEW ORDER
// ============================================================================

// PreviewOrderResponse is the response for POST /trader/v1/accounts/{accountHash}/previewOrder
type PreviewOrderResponse struct {
	OrderID               int64                  `json:"orderId"`
	OrderStrategy         *OrderStrategy         `json:"orderStrategy,omitempty"`
	OrderValidationResult *OrderValidationResult `json:"orderValidationResult,omitempty"`
	CommissionAndFee      *CommissionAndFee      `json:"commissionAndFee,omitempty"`
}

// OrderStrategy represents the order strategy in preview and order strategy responses.
type OrderStrategy struct {
	AccountNumber          string        `json:"accountNumber,omitempty"`
	AdvancedOrderType      string        `json:"advancedOrderType,omitempty"`
	CloseTime              *string       `json:"closeTime,omitempty"`
	EnteredTime            string        `json:"enteredTime,omitempty"`
	OrderBalance           *OrderBalance `json:"orderBalance,omitempty"`
	OrderStrategyType      string        `json:"orderStrategyType,omitempty"`
	OrderVersion           float64       `json:"orderVersion,omitempty"`
	Session                string        `json:"session,omitempty"`
	Status                 string        `json:"status,omitempty"`
	AllOrNone              bool          `json:"allOrNone,omitempty"`
	Discretionary          bool          `json:"discretionary,omitempty"`
	Duration               string        `json:"duration,omitempty"`
	FilledQuantity         float64       `json:"filledQuantity,omitempty"`
	OrderType              string        `json:"orderType,omitempty"`
	OrderValue             float64       `json:"orderValue,omitempty"`
	Price                  float64       `json:"price,omitempty"`
	Quantity               float64       `json:"quantity,omitempty"`
	RemainingQuantity      float64       `json:"remainingQuantity,omitempty"`
	SellNonMarginableFirst bool          `json:"sellNonMarginableFirst,omitempty"`
	SettlementInstruction  string        `json:"settlementInstruction,omitempty"`
	Strategy               string        `json:"strategy,omitempty"`
	AmountIndicator        string        `json:"amountIndicator,omitempty"`
	OrderLegs              []*OrderLeg   `json:"orderLegs,omitempty"`
}

// PreviewOrderStrategy is an alias for OrderStrategy (backward compat).
type PreviewOrderStrategy = OrderStrategy

// PreviewOrderLeg is an alias for OrderLeg (backward compat).
type PreviewOrderLeg = OrderLeg

// OrderBalance represents order balance information.
type OrderBalance struct {
	OrderValue             float64 `json:"orderValue,omitempty"`
	ProjectedAvailableFund float64 `json:"projectedAvailableFund,omitempty"`
	ProjectedBuyingPower   float64 `json:"projectedBuyingPower,omitempty"`
	ProjectedCommission    float64 `json:"projectedCommission,omitempty"`
}

// OrderValidationResult represents validation results for an order.
type OrderValidationResult struct {
	Alerts  []*OrderValidationDetail `json:"alerts,omitempty"`
	Accepts []*OrderValidationDetail `json:"accepts,omitempty"`
	Rejects []*OrderValidationDetail `json:"rejects,omitempty"`
	Reviews []*OrderValidationDetail `json:"reviews,omitempty"`
	Warns   []*OrderValidationDetail `json:"warns,omitempty"`
}

// OrderValidationDetail represents a single validation detail entry.
type OrderValidationDetail struct {
	ValidationRuleName string `json:"validationRuleName,omitempty"`
	Message            string `json:"message,omitempty"`
	ActivityMessage    string `json:"activityMessage,omitempty"`
	OriginalSeverity   string `json:"originalSeverity,omitempty"`
	OverrideName       string `json:"overrideName,omitempty"`
	OverrideSeverity   string `json:"overrideSeverity,omitempty"`
}

// OrderReject is an alias for OrderValidationDetail (backward compat).
type OrderReject = OrderValidationDetail

// CommissionAndFee represents commission and fee details.
type CommissionAndFee struct {
	Commission     *Commission `json:"commission,omitempty"`
	Fee            *Fees       `json:"fee,omitempty"`
	TrueCommission *Commission `json:"trueCommission,omitempty"`
}

// Commission represents commission details.
type Commission struct {
	CommissionLegs []*CommissionLeg `json:"commissionLegs,omitempty"`
}

// CommissionLeg represents a commission leg.
type CommissionLeg struct {
	CommissionValues []*CommissionValue `json:"commissionValues,omitempty"`
}

// CommissionValue represents a commission value.
type CommissionValue struct {
	Value float64 `json:"value,omitempty"`
	Type  string  `json:"type,omitempty"`
}

// Fees represents fee details.
type Fees struct {
	FeeLegs []*FeeLeg `json:"feeLegs,omitempty"`
}

// Fee is an alias for Fees (backward compat).
type Fee = Fees

// FeeLeg represents a fee leg.
type FeeLeg struct {
	FeeValues []*FeeValue `json:"feeValues,omitempty"`
}

// FeeValue represents a fee value.
type FeeValue struct {
	Value float64 `json:"value,omitempty"`
	Type  string  `json:"type,omitempty"`
}

// ============================================================================
// TRANSACTIONS
// ============================================================================

// TransactionsResponse is the response for GET /trader/v1/accounts/{accountNumber}/transactions
type TransactionsResponse []Transaction

// Transaction represents a transaction in an account's history.
type Transaction struct {
	ActivityID     int64           `json:"activityId,omitempty"`
	Time           string          `json:"time,omitempty"`
	User           *UserDetails    `json:"user,omitempty"`
	Description    string          `json:"description,omitempty"`
	AccountNumber  string          `json:"accountNumber,omitempty"`
	Type           string          `json:"type,omitempty"`
	Status         string          `json:"status,omitempty"`
	SubAccount     string          `json:"subAccount,omitempty"`
	TradeDate      string          `json:"tradeDate,omitempty"`
	SettlementDate string          `json:"settlementDate,omitempty"`
	PositionID     int64           `json:"positionId,omitempty"`
	OrderID        int64           `json:"orderId,omitempty"`
	NetAmount      float64         `json:"netAmount,omitempty"`
	ActivityType   string          `json:"activityType,omitempty"`
	TransferItems  []*TransferItem `json:"transferItems,omitempty"`

	// Legacy compatibility fields (used by older code paths)
	TransactionID string `json:"transactionId,omitempty"`
}

// TransactionDetailsResponse is the response for GET /trader/v1/accounts/{accountNumber}/transactions/{transactionId}
type TransactionDetailsResponse Transaction

// TransferItem represents a transfer item in a transaction.
type TransferItem struct {
	Instrument     *TransactionInstrument `json:"instrument,omitempty"`
	Amount         float64                `json:"amount,omitempty"`
	Cost           float64                `json:"cost,omitempty"`
	Price          float64                `json:"price,omitempty"`
	FeeType        string                 `json:"feeType,omitempty"`
	PositionEffect string                 `json:"positionEffect,omitempty"`
}

// TransactionItem is an alias for TransferItem (backward compat).
type TransactionItem = TransferItem

// TransactionFees represents fees associated with a transaction item.
// This is a convenience struct for code that accesses fees by name.
type TransactionFees struct {
	SEC        float64 `json:"secFee,omitempty"`
	TAF        float64 `json:"tafFee,omitempty"`
	Commission float64 `json:"commission,omitempty"`
	Additional float64 `json:"additionalFee,omitempty"`
	OptRegFee  float64 `json:"optRegFee,omitempty"`
}

// TransactionBaseInstrument is the base instrument for transaction responses.
type TransactionBaseInstrument struct {
	AssetType    string  `json:"assetType"`
	Cusip        string  `json:"cusip,omitempty"`
	Symbol       string  `json:"symbol,omitempty"`
	Description  string  `json:"description,omitempty"`
	InstrumentID int64   `json:"instrumentId,omitempty"`
	NetChange    float64 `json:"netChange,omitempty"`
}

// TransactionInstrument is the instrument embedded in TransferItem.
// It maps to the oneOf AccountsInstrument / TransactionInstrument schema.
type TransactionInstrument struct {
	TransactionBaseInstrument
}

// TransactionOption extends TransactionBaseInstrument with option-specific fields.
type TransactionOption struct {
	TransactionBaseInstrument
	ExpirationDate          string                             `json:"expirationDate,omitempty"`
	OptionDeliverables      []*TransactionAPIOptionDeliverable `json:"optionDeliverables,omitempty"`
	OptionPremiumMultiplier int64                              `json:"optionPremiumMultiplier,omitempty"`
	PutCall                 string                             `json:"putCall,omitempty"`
	StrikePrice             float64                            `json:"strikePrice,omitempty"`
	Type                    string                             `json:"type,omitempty"`
	UnderlyingSymbol        string                             `json:"underlyingSymbol,omitempty"`
	UnderlyingCusip         string                             `json:"underlyingCusip,omitempty"`
	Deliverable             *TransactionInstrument             `json:"deliverable,omitempty"`
}

// TransactionAPIOptionDeliverable represents an option deliverable in transactions.
type TransactionAPIOptionDeliverable struct {
	RootSymbol        string                 `json:"rootSymbol,omitempty"`
	StrikePercent     int64                  `json:"strikePercent,omitempty"`
	DeliverableNumber int64                  `json:"deliverableNumber,omitempty"`
	DeliverableUnits  float64                `json:"deliverableUnits,omitempty"`
	Deliverable       *TransactionInstrument `json:"deliverable,omitempty"`
	AssetType         string                 `json:"assetType,omitempty"`
}

// TransactionFixedIncome extends TransactionBaseInstrument with fixed-income fields.
type TransactionFixedIncome struct {
	TransactionBaseInstrument
	Type         string  `json:"type,omitempty"`
	MaturityDate string  `json:"maturityDate,omitempty"`
	Factor       float64 `json:"factor,omitempty"`
	Multiplier   float64 `json:"multiplier,omitempty"`
	VariableRate float64 `json:"variableRate,omitempty"`
}

// TransactionMutualFund extends TransactionBaseInstrument with mutual fund fields.
type TransactionMutualFund struct {
	TransactionBaseInstrument
	FundFamilyName       string `json:"fundFamilyName,omitempty"`
	FundFamilySymbol     string `json:"fundFamilySymbol,omitempty"`
	FundGroup            string `json:"fundGroup,omitempty"`
	Type                 string `json:"type,omitempty"`
	ExchangeCutoffTime   string `json:"exchangeCutoffTime,omitempty"`
	PurchaseCutoffTime   string `json:"purchaseCutoffTime,omitempty"`
	RedemptionCutoffTime string `json:"redemptionCutoffTime,omitempty"`
}

// TransactionCashEquivalent extends TransactionBaseInstrument.
type TransactionCashEquivalent struct {
	TransactionBaseInstrument
	Type string `json:"type,omitempty"`
}

// TransactionEquity extends TransactionBaseInstrument.
type TransactionEquity struct {
	TransactionBaseInstrument
	Type string `json:"type,omitempty"`
}

// ============================================================================
// USER PREFERENCES
// ============================================================================

// PreferencesResponse is the response for GET /trader/v1/userPreference
type PreferencesResponse struct {
	Accounts     []*UserPreferenceAccount `json:"accounts,omitempty"`
	StreamerInfo []*StreamerInfo          `json:"streamerInfo,omitempty"`
	Offers       []*Offer                 `json:"offers,omitempty"`
}

// StreamerInfo represents streamer configuration.
type StreamerInfo struct {
	StreamerURL            string `json:"streamerSocketUrl"`
	SchwabClientCustomerID string `json:"schwabClientCustomerId,omitempty"`
	SchwabClientCorrelID   string `json:"schwabClientCorrelId,omitempty"`
	SchwabClientChannel    string `json:"schwabClientChannel,omitempty"`
	SchwabClientFunctionID string `json:"schwabClientFunctionId,omitempty"`
}

// UserPreferenceAccount represents a user's account preference.
type UserPreferenceAccount struct {
	AccountNumber      string `json:"accountNumber,omitempty"`
	PrimaryAccount     bool   `json:"primaryAccount,omitempty"`
	Type               string `json:"type,omitempty"`
	NickName           string `json:"nickName,omitempty"`
	AccountColor       string `json:"accountColor,omitempty"`
	DisplayAcctID      string `json:"displayAcctId,omitempty"`
	AutoPositionEffect bool   `json:"autoPositionEffect,omitempty"`
}

// Offer represents an offer/permission level.
type Offer struct {
	Level2Permissions bool   `json:"level2Permissions,omitempty"`
	MktDataPermission string `json:"mktDataPermission,omitempty"`
}

// UserDetails represents the user associated with a transaction.
type UserDetails struct {
	CDDomainID     string `json:"cdDomainId,omitempty"`
	Login          string `json:"login,omitempty"`
	Type           string `json:"type,omitempty"`
	UserID         int64  `json:"userId,omitempty"`
	SystemUserName string `json:"systemUserName,omitempty"`
	FirstName      string `json:"firstName,omitempty"`
	LastName       string `json:"lastName,omitempty"`
	BrokerRepCode  string `json:"brokerRepCode,omitempty"`
}

// ============================================================================
// MARKET DATA API RESPONSE TYPES
// ============================================================================

// QuotesResponse is the response for GET /marketdata/v1/quotes
type QuotesResponse map[string]Quote

// QuoteResponse is the response for GET /marketdata/v1/{symbol_id}/quotes
type QuoteResponse Quote

// ExtendedMarket represents extended-hours market data for a quote.
// Corresponds to the OpenAPI ExtendedMarket schema.
type ExtendedMarket struct {
	AskPrice   float64 `json:"askPrice,omitempty"`
	AskSize    int     `json:"askSize,omitempty"`
	BidPrice   float64 `json:"bidPrice,omitempty"`
	BidSize    int     `json:"bidSize,omitempty"`
	LastPrice  float64 `json:"lastPrice,omitempty"`
	LastSize   int     `json:"lastSize,omitempty"`
	Mark       float64 `json:"mark,omitempty"`
	QuoteTime  int64   `json:"quoteTime,omitempty"`
	TotalVolume int64  `json:"totalVolume,omitempty"`
	TradeTime  int64   `json:"tradeTime,omitempty"`
}

// Quote represents a complete quote with all data sections.
// The API returns different sub-objects (quote, fundamental, reference, regular)
// as a flat structure; we embed them for convenience.
type Quote struct {
	AssetMainType string `json:"assetMainType,omitempty"`
	AssetSubType  string `json:"assetSubType,omitempty"`
	SSID          int64  `json:"ssid,omitempty"`
	Symbol        string `json:"symbol,omitempty"`
	RealTime      bool   `json:"realtime,omitempty"`
	QuoteType     string `json:"quoteType,omitempty"`
	Extended      *ExtendedMarket `json:"extended,omitempty"`

	// Quote data (QuoteEquity / QuoteOption / etc.)
	*QuoteData

	// Sub-sections
	Fundamental *Fundamental `json:"fundamental,omitempty"`
	Reference   *Reference   `json:"reference,omitempty"`
	Regular     *Regular     `json:"regular,omitempty"`
}

// QuoteData represents real-time quote data.
type QuoteData struct {
	FiftyTwoWeekHigh  float64 `json:"52WeekHigh,omitempty"`
	FiftyTwoWeekLow   float64 `json:"52WeekLow,omitempty"`
	AskMICId          string  `json:"askMICId,omitempty"`
	AskPrice          float64 `json:"askPrice,omitempty"`
	AskSize           int     `json:"askSize,omitempty"`
	AskTime           int64   `json:"askTime,omitempty"`
	BidMICId          string  `json:"bidMICId,omitempty"`
	BidPrice          float64 `json:"bidPrice,omitempty"`
	BidSize           int     `json:"bidSize,omitempty"`
	BidTime           int64   `json:"bidTime,omitempty"`
	ClosePrice        float64 `json:"closePrice,omitempty"`
	HighPrice         float64 `json:"highPrice,omitempty"`
	LastMICId         string  `json:"lastMICId,omitempty"`
	LastPrice         float64 `json:"lastPrice,omitempty"`
	LastSize          int     `json:"lastSize,omitempty"`
	LowPrice          float64 `json:"lowPrice,omitempty"`
	Mark              float64 `json:"mark,omitempty"`
	MarkChange        float64 `json:"markChange,omitempty"`
	MarkPercentChange float64 `json:"markPercentChange,omitempty"`
	NetChange         float64 `json:"netChange,omitempty"`
	NetPercentChange  float64 `json:"netPercentChange,omitempty"`
	OpenPrice         float64 `json:"openPrice,omitempty"`
	QuoteTime         int64   `json:"quoteTime,omitempty"`
	SecurityStatus    string  `json:"securityStatus,omitempty"`
	TotalVolume       int64   `json:"totalVolume,omitempty"`
	TradeTime         int64   `json:"tradeTime,omitempty"`
	Volatility        float64 `json:"volatility,omitempty"`
}

// Fundamental represents fundamental data.
type Fundamental struct {
	Avg10DaysVolume    float64 `json:"avg10DaysVolume,omitempty"`
	Avg1YearVolume     float64 `json:"avg1YearVolume,omitempty"`
	DeclarationDate    string  `json:"declarationDate,omitempty"`
	DivAmount          float64 `json:"divAmount,omitempty"`
	DivExDate          string  `json:"divExDate,omitempty"`
	DivFreq            int     `json:"divFreq,omitempty"`
	DivPayAmount       float64 `json:"divPayAmount,omitempty"`
	DivPayDate         string  `json:"divPayDate,omitempty"`
	DivYield           float64 `json:"divYield,omitempty"`
	Eps                float64 `json:"eps,omitempty"`
	FundLeverageFactor float64 `json:"fundLeverageFactor,omitempty"`
	NextDivExDate      string  `json:"nextDivExDate,omitempty"`
	NextDivPayDate     string  `json:"nextDivPayDate,omitempty"`
	PeRatio            float64 `json:"peRatio,omitempty"`
	FundStrategy       string  `json:"fundStrategy,omitempty"`
}

// Reference represents reference data (ReferenceEquity).
type Reference struct {
	Cusip          string  `json:"cusip,omitempty"`
	Description    string  `json:"description,omitempty"`
	Exchange       string  `json:"exchange,omitempty"`
	ExchangeName   string  `json:"exchangeName,omitempty"`
	FSIDesc        string  `json:"fsiDesc,omitempty"`
	HtbQuantity    int64   `json:"htbQuantity,omitempty"`
	HtbRate        float64 `json:"htbRate,omitempty"`
	IsHardToBorrow bool    `json:"isHardToBorrow,omitempty"`
	IsShortable    bool    `json:"isShortable,omitempty"`
	OTCMarketTier  string  `json:"otcMarketTier,omitempty"`
}

// Regular represents regular market trading data.
type Regular struct {
	RegularMarketLastPrice     float64 `json:"regularMarketLastPrice,omitempty"`
	RegularMarketLastSize      int64   `json:"regularMarketLastSize,omitempty"`
	RegularMarketNetChange     float64 `json:"regularMarketNetChange,omitempty"`
	RegularMarketPercentChange float64 `json:"regularMarketPercentChange,omitempty"`
	RegularMarketTradeTime     int64   `json:"regularMarketTradeTime,omitempty"`
}

// OptionChainsResponse is the response for GET /marketdata/v1/chains
type OptionChainsResponse struct {
	Symbol           string                                 `json:"symbol,omitempty"`
	Status           string                                 `json:"status,omitempty"`
	Underlying       *Underlying                            `json:"underlying,omitempty"`
	Strategy         string                                 `json:"strategy,omitempty"`
	Interval         float64                                `json:"interval,omitempty"`
	IsDelayed        bool                                   `json:"isDelayed,omitempty"`
	IsIndex          bool                                   `json:"isIndex,omitempty"`
	DaysToExpiration float64                                `json:"daysToExpiration,omitempty"`
	InterestRate     float64                                `json:"interestRate,omitempty"`
	UnderlyingPrice  float64                                `json:"underlyingPrice,omitempty"`
	Volatility       float64                                `json:"volatility,omitempty"`
	CallExpDateMap   map[string]map[string][]OptionContract `json:"callExpDateMap,omitempty"`
	PutExpDateMap    map[string]map[string][]OptionContract `json:"putExpDateMap,omitempty"`
}

// Underlying represents underlying data in an option chain.
type Underlying struct {
	Ask               float64 `json:"ask,omitempty"`
	AskSize           int     `json:"askSize,omitempty"`
	Bid               float64 `json:"bid,omitempty"`
	BidSize           int     `json:"bidSize,omitempty"`
	Change            float64 `json:"change,omitempty"`
	Close             float64 `json:"close,omitempty"`
	Delayed           bool    `json:"delayed,omitempty"`
	Description       string  `json:"description,omitempty"`
	ExchangeName      string  `json:"exchangeName,omitempty"`
	FiftyTwoWeekHigh  float64 `json:"fiftyTwoWeekHigh,omitempty"`
	FiftyTwoWeekLow   float64 `json:"fiftyTwoWeekLow,omitempty"`
	HighPrice         float64 `json:"highPrice,omitempty"`
	Last              float64 `json:"last,omitempty"`
	LowPrice          float64 `json:"lowPrice,omitempty"`
	Mark              float64 `json:"mark,omitempty"`
	MarkChange        float64 `json:"markChange,omitempty"`
	MarkPercentChange float64 `json:"markPercentChange,omitempty"`
	OpenPrice         float64 `json:"openPrice,omitempty"`
	PercentChange     float64 `json:"percentChange,omitempty"`
	QuoteTime         int64   `json:"quoteTime,omitempty"`
	Symbol            string  `json:"symbol,omitempty"`
	TotalVolume       int64   `json:"totalVolume,omitempty"`
	TradeTime         int64   `json:"tradeTime,omitempty"`
}

// OptionContract represents an option contract.
type OptionContract struct {
	PutCall                string               `json:"putCall,omitempty"`
	Symbol                 string               `json:"symbol,omitempty"`
	Description            string               `json:"description,omitempty"`
	ExchangeName           string               `json:"exchangeName,omitempty"`
	BidPrice               float64              `json:"bidPrice,omitempty"`
	AskPrice               float64              `json:"askPrice,omitempty"`
	LastPrice              float64              `json:"lastPrice,omitempty"`
	MarkPrice              float64              `json:"markPrice,omitempty"`
	BidSize                int                  `json:"bidSize,omitempty"`
	AskSize                int                  `json:"askSize,omitempty"`
	LastSize               int                  `json:"lastSize,omitempty"`
	HighPrice              float64              `json:"highPrice,omitempty"`
	LowPrice               float64              `json:"lowPrice,omitempty"`
	OpenPrice              float64              `json:"openPrice,omitempty"`
	ClosePrice             float64              `json:"closePrice,omitempty"`
	TotalVolume            int64                `json:"totalVolume,omitempty"`
	TradeDate              string               `json:"tradeDate,omitempty"`
	QuoteTimeInLong        int64                `json:"quoteTimeInLong,omitempty"`
	TradeTimeInLong        int64                `json:"tradeTimeInLong,omitempty"`
	NetChange              float64              `json:"netChange,omitempty"`
	Volatility             float64              `json:"volatility,omitempty"`
	Delta                  float64              `json:"delta,omitempty"`
	Gamma                  float64              `json:"gamma,omitempty"`
	Theta                  float64              `json:"theta,omitempty"`
	Vega                   float64              `json:"vega,omitempty"`
	Rho                    float64              `json:"rho,omitempty"`
	OpenInterest           int                  `json:"openInterest,omitempty"`
	TimeValue              float64              `json:"timeValue,omitempty"`
	TheoreticalOptionValue float64              `json:"theoreticalOptionValue,omitempty"`
	TheoreticalVolatility  float64              `json:"theoreticalVolatility,omitempty"`
	OptionDeliverablesList []*OptionDeliverable `json:"optionDeliverablesList,omitempty"`
	StrikePrice            float64              `json:"strikePrice,omitempty"`
	ExpirationDate         string               `json:"expirationDate,omitempty"`
	DaysToExpiration       int                  `json:"daysToExpiration,omitempty"`
	ExpirationType         string               `json:"expirationType,omitempty"`
	LastTradingDay         string               `json:"lastTradingDay,omitempty"`
	Multiplier             float64              `json:"multiplier,omitempty"`
	SettlementType         string               `json:"settlementType,omitempty"`
	DeliverableNote        string               `json:"deliverableNote,omitempty"`
	IsIndexOption          bool                 `json:"isIndexOption,omitempty"`
	PercentChange          float64              `json:"percentChange,omitempty"`
	MarkChange             float64              `json:"markChange,omitempty"`
	MarkPercentChange      float64              `json:"markPercentChange,omitempty"`
	IsPennyPilot           bool                 `json:"isPennyPilot,omitempty"`
	IntrinsicValue         float64              `json:"intrinsicValue,omitempty"`
	OptionRoot             string               `json:"optionRoot,omitempty"`
	IsInTheMoney           bool                 `json:"isInTheMoney,omitempty"`
	IsMini                 bool                 `json:"isMini,omitempty"`
	IsNonStandard          bool                 `json:"isNonStandard,omitempty"`
}

// OptionDeliverable represents option deliverable information.
type OptionDeliverable struct {
	Symbol           string  `json:"symbol,omitempty"`
	AssetType        string  `json:"assetType,omitempty"`
	DeliverableUnits float64 `json:"deliverableUnits,omitempty"`
	CurrencyType     string  `json:"currencyType,omitempty"`
}

// OptionExpirationChainResponse is the response for GET /marketdata/v1/expirationchain
type OptionExpirationChainResponse struct {
	Status         string            `json:"status,omitempty"`
	ExpirationList []*ExpirationDate `json:"expirationList,omitempty"`
}

// ExpirationDate represents an option expiration date.
type ExpirationDate struct {
	DaysToExpiration int      `json:"daysToExpiration,omitempty"`
	Expiration       string   `json:"expiration,omitempty"`
	ExpirationType   string   `json:"expirationType,omitempty"`
	Standard         bool     `json:"standard,omitempty"`
	SettlementType   string   `json:"settlementType,omitempty"`
	OptionRoots      string  `json:"optionRoots,omitempty"`
}

// PriceHistoryResponse is the response for GET /marketdata/v1/pricehistory
type PriceHistoryResponse struct {
	Candles                  []*Candle `json:"candles,omitempty"`
	Empty                    bool      `json:"empty,omitempty"`
	Symbol                   string    `json:"symbol,omitempty"`
	PreviousClose            float64   `json:"previousClose,omitempty"`
	PreviousCloseDate        int64     `json:"previousCloseDate,omitempty"`
	PreviousCloseDateISO8601 string    `json:"previousCloseDateISO8601,omitempty"`
}

// Candle represents a price history candle.
type Candle struct {
	Close           float64 `json:"close,omitempty"`
	Datetime        int64   `json:"datetime,omitempty"`
	DatetimeISO8601 string  `json:"datetimeISO8601,omitempty"`
	High            float64 `json:"high,omitempty"`
	Low             float64 `json:"low,omitempty"`
	Open            float64 `json:"open,omitempty"`
	Volume          int64   `json:"volume,omitempty"`
}

// MoversResponse is the response for GET /marketdata/v1/movers/{symbol}
type MoversResponse []Mover

// Mover represents a market mover.
type Mover struct {
	Symbol      string  `json:"symbol,omitempty"`
	Description string  `json:"description,omitempty"`
	Change      float64 `json:"change,omitempty"`
	Direction   string  `json:"direction,omitempty"`
	LastPrice   float64 `json:"last,omitempty"`
	TotalVolume int64   `json:"totalVolume,omitempty"`
}

// MarketHoursResponse is the response for GET /marketdata/v1/markets
type MarketHoursResponse map[string]MarketHour

// MarketHourResponse is the response for GET /marketdata/v1/markets/{market_id}
type MarketHourResponse MarketHour

// MarketHour represents market hours for a specific market.
type MarketHour struct {
	Date         string        `json:"date,omitempty"`
	MarketType   string        `json:"marketType,omitempty"`
	Exchange     string        `json:"exchange,omitempty"`
	Category     string        `json:"category,omitempty"`
	Product      string        `json:"product,omitempty"`
	ProductName  string        `json:"productName,omitempty"`
	IsOpen       bool          `json:"isOpen,omitempty"`
	SessionHours *SessionHours `json:"sessionHours,omitempty"`
}

// SessionHours represents market session hours.
type SessionHours struct {
	PreMarket     []*Interval `json:"preMarket,omitempty"`
	RegularMarket []*Interval `json:"regularMarket,omitempty"`
}

// Interval represents a market session time interval.
type Interval struct {
	Start string `json:"start,omitempty"`
	End   string `json:"end,omitempty"`
}

// SessionDuration is an alias for Interval (backward compat).
type SessionDuration = Interval

// StartEndTime is an alias for Interval (backward compat).
type StartEndTime = Interval

// InstrumentsResponse is the response for GET /marketdata/v1/instruments
type InstrumentsResponse []InstrumentSearch

// MarketDataInstrument represents instrument info for a market-data instrument search.
// Corresponds to the OpenAPI Instrument schema (cusip, symbol, description, exchange, assetType, type).
// This is distinct from the Trader API Instrument type which embeds AccountsBaseInstrument.
type MarketDataInstrument struct {
	Cusip       string `json:"cusip,omitempty"`
	Symbol      string `json:"symbol,omitempty"`
	Description string `json:"description,omitempty"`
	Exchange    string `json:"exchange,omitempty"`
	AssetType   string `json:"assetType,omitempty"`
	Type        string `json:"type,omitempty"`
}

// MarketDataBond represents bond instrument info for a market-data instrument search.
// Corresponds to the OpenAPI Bond schema.
type MarketDataBond struct {
	Cusip         string  `json:"cusip,omitempty"`
	Symbol        string  `json:"symbol,omitempty"`
	Description   string  `json:"description,omitempty"`
	Exchange      string  `json:"exchange,omitempty"`
	AssetType     string  `json:"assetType,omitempty"`
	BondFactor    string  `json:"bondFactor,omitempty"`
	BondMultiplier string  `json:"bondMultiplier,omitempty"`
	BondPrice     float64 `json:"bondPrice,omitempty"`
	Type          string  `json:"type,omitempty"`
}

// FundamentalInst represents detailed fundamental data for an instrument search.
// Corresponds to the OpenAPI FundamentalInst schema.
type FundamentalInst struct {
	Symbol              string  `json:"symbol,omitempty"`
	High52              float64 `json:"high52,omitempty"`
	Low52               float64 `json:"low52,omitempty"`
	DividendAmount      float64 `json:"dividendAmount,omitempty"`
	DividendYield       float64 `json:"dividendYield,omitempty"`
	DividendDate        string  `json:"dividendDate,omitempty"`
	PeRatio             float64 `json:"peRatio,omitempty"`
	PegRatio            float64 `json:"pegRatio,omitempty"`
	PbRatio             float64 `json:"pbRatio,omitempty"`
	PrRatio             float64 `json:"prRatio,omitempty"`
	PcfRatio            float64 `json:"pcfRatio,omitempty"`
	GrossMarginTTM      float64 `json:"grossMarginTTM,omitempty"`
	GrossMarginMRQ      float64 `json:"grossMarginMRQ,omitempty"`
	NetProfitMarginTTM  float64 `json:"netProfitMarginTTM,omitempty"`
	NetProfitMarginMRQ  float64 `json:"netProfitMarginMRQ,omitempty"`
	OperatingMarginTTM  float64 `json:"operatingMarginTTM,omitempty"`
	OperatingMarginMRQ  float64 `json:"operatingMarginMRQ,omitempty"`
	ReturnOnEquity      float64 `json:"returnOnEquity,omitempty"`
	ReturnOnAssets      float64 `json:"returnOnAssets,omitempty"`
	ReturnOnInvestment  float64 `json:"returnOnInvestment,omitempty"`
	QuickRatio          float64 `json:"quickRatio,omitempty"`
	CurrentRatio        float64 `json:"currentRatio,omitempty"`
	InterestCoverage    float64 `json:"interestCoverage,omitempty"`
	TotalDebtToCapital  float64 `json:"totalDebtToCapital,omitempty"`
	LtDebtToEquity      float64 `json:"ltDebtToEquity,omitempty"`
	TotalDebtToEquity   float64 `json:"totalDebtToEquity,omitempty"`
	EpsTTM              float64 `json:"epsTTM,omitempty"`
	EpsChangePercentTTM float64 `json:"epsChangePercentTTM,omitempty"`
	EpsChangeYear       float64 `json:"epsChangeYear,omitempty"`
	EpsChange            float64 `json:"epsChange,omitempty"`
	RevChangeYear       float64 `json:"revChangeYear,omitempty"`
	RevChangeTTM        float64 `json:"revChangeTTM,omitempty"`
	RevChangeIn         float64 `json:"revChangeIn,omitempty"`
	SharesOutstanding   float64 `json:"sharesOutstanding,omitempty"`
	MarketCapFloat      float64 `json:"marketCapFloat,omitempty"`
	MarketCap           float64 `json:"marketCap,omitempty"`
	BookValuePerShare   float64 `json:"bookValuePerShare,omitempty"`
	ShortIntToFloat     float64 `json:"shortIntToFloat,omitempty"`
	ShortIntDayToCover  float64 `json:"shortIntDayToCover,omitempty"`
	DivGrowthRate3Year  float64 `json:"divGrowthRate3Year,omitempty"`
	DividendPayAmount   float64 `json:"dividendPayAmount,omitempty"`
	DividendPayDate     string  `json:"dividendPayDate,omitempty"`
	Beta                float64 `json:"beta,omitempty"`
	Vol1DayAvg          float64 `json:"vol1DayAvg,omitempty"`
	Vol10DayAvg         float64 `json:"vol10DayAvg,omitempty"`
	Vol3MonthAvg        float64 `json:"vol3MonthAvg,omitempty"`
	Avg10DaysVolume     int64   `json:"avg10DaysVolume,omitempty"`
	Avg1DayVolume       int64   `json:"avg1DayVolume,omitempty"`
	Avg3MonthVolume     int64   `json:"avg3MonthVolume,omitempty"`
	DeclarationDate     string  `json:"declarationDate,omitempty"`
	DividendFreq        int     `json:"dividendFreq,omitempty"`
	Eps                 float64 `json:"eps,omitempty"`
	CorporationActionDate string `json:"corporationActionDate,omitempty"`
	DtnVolume           int64   `json:"dtnVolume,omitempty"`
	NextDividendPayDate string  `json:"nextDividendPayDate,omitempty"`
	NextDividendDate    string  `json:"nextDividendDate,omitempty"`
	FundLeverageFactor  float64 `json:"fundLeverageFactor,omitempty"`
	FundStrategy        string  `json:"fundStrategy,omitempty"`
}

// InstrumentSearch represents an instrument search result.
// Corresponds to the OpenAPI InstrumentResponse schema.
type InstrumentSearch struct {
	Cusip             string               `json:"cusip,omitempty"`
	Symbol            string               `json:"symbol,omitempty"`
	Description        string               `json:"description,omitempty"`
	Exchange          string               `json:"exchange,omitempty"`
	AssetType         string               `json:"assetType,omitempty"`
	BondFactor        string               `json:"bondFactor,omitempty"`
	BondMultiplier    string               `json:"bondMultiplier,omitempty"`
	BondPrice         float64              `json:"bondPrice,omitempty"`
	Type              string               `json:"type,omitempty"`
	Fundamental       *FundamentalInst     `json:"fundamental,omitempty"`
	InstrumentInfo    *MarketDataInstrument `json:"instrumentInfo,omitempty"`
	BondInstrumentInfo *MarketDataBond     `json:"bondInstrumentInfo,omitempty"`
}

// InstrumentCUSIPResponse is the response for GET /marketdata/v1/instruments/{cusip_id}
type InstrumentCUSIPResponse []*InstrumentSearch

// InstrumentCUSIP represents an instrument by CUSIP.
// Deprecated: Use InstrumentSearch instead.
type InstrumentCUSIP = InstrumentSearch

// ============================================================================
// REQUEST TYPES FOR POST/PUT OPERATIONS
// ============================================================================

// OrderRequest represents an order request for place_order and replace_order.
// Fields not needed for a given order type can be left zero-valued.
type OrderRequest struct {
	Session                  string                `json:"session,omitempty"`
	Duration                 string                `json:"duration,omitempty"`
	OrderType                string                `json:"orderType,omitempty"`
	CancelTime               *string               `json:"cancelTime,omitempty"`
	ComplexOrderStrategyType string                `json:"complexOrderStrategyType,omitempty"`
	Quantity                 float64               `json:"quantity,omitempty"`
	FilledQuantity           float64               `json:"filledQuantity,omitempty"`
	RemainingQuantity        float64               `json:"remainingQuantity,omitempty"`
	DestinationLinkName      string                `json:"destinationLinkName,omitempty"`
	ReleaseTime              *string               `json:"releaseTime,omitempty"`
	StopPrice                float64               `json:"stopPrice,omitempty"`
	StopPriceLinkBasis       string                `json:"stopPriceLinkBasis,omitempty"`
	StopPriceLinkType        string                `json:"stopPriceLinkType,omitempty"`
	StopPriceOffset          float64               `json:"stopPriceOffset,omitempty"`
	StopType                 string                `json:"stopType,omitempty"`
	PriceLinkBasis           string                `json:"priceLinkBasis,omitempty"`
	PriceLinkType            string                `json:"priceLinkType,omitempty"`
	Price                    float64               `json:"price,omitempty"`
	TaxLotMethod             string                `json:"taxLotMethod,omitempty"`
	OrderLegCollection       []*OrderLegCollection `json:"orderLegCollection,omitempty"`
	ActivationPrice          float64               `json:"activationPrice,omitempty"`
	SpecialInstruction       string                `json:"specialInstruction,omitempty"`
	OrderStrategyType        string                `json:"orderStrategyType,omitempty"`
	OrderID                  int64                 `json:"orderId,omitempty"`
	Cancelable               bool                  `json:"cancelable,omitempty"`
	Editable                 bool                  `json:"editable,omitempty"`
	Status                   string                `json:"status,omitempty"`
	EnteredTime              string                `json:"enteredTime,omitempty"`
	CloseTime                *string               `json:"closeTime,omitempty"`
	AccountNumber            int64                 `json:"accountNumber,omitempty"`
	OrderActivityCollection  []*OrderActivity      `json:"orderActivityCollection,omitempty"`
	ReplacingOrderCollection []*OrderRequest       `json:"replacingOrderCollection,omitempty"`
	ChildOrderStrategies     []*OrderRequest       `json:"childOrderStrategies,omitempty"`
	StatusDescription        string                `json:"statusDescription,omitempty"`
}

// OrderLegRequest is an alias for OrderLegCollection for single-leg orders.
// Deprecated: Use OrderLegCollection for new code.
type OrderLegRequest = OrderLegCollection

// InstrumentRequest represents an instrument in an order request.
type InstrumentRequest struct {
	Symbol    string `json:"symbol"`
	AssetType string `json:"assetType"`
}

// PreviewOrderRequest represents a preview order request
type PreviewOrderRequest OrderRequest
