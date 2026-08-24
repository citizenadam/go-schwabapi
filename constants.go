package schwabdev

import "time"

// HTTP Client Constants
const (
	// DefaultHTTPRequestTimeout is the default timeout for HTTP requests to the Schwab API
	DefaultHTTPRequestTimeout = 10 * time.Second

	// OAuthTokenRequestTimeout is the timeout for OAuth token request operations
	OAuthTokenRequestTimeout = 30 * time.Second
)

// Token Management Constants
const (
	// OAuthTokenEndpoint is the Schwab OAuth token endpoint. TokenManager
	// posts authorization_code and refresh_token grants here. Tests may point
	// a TokenManager's unexported oauthTokenURL field at a local server.
	OAuthTokenEndpoint = "https://api.schwabapi.com/v1/oauth/token"

	// AccessTokenValidity is the validity period for access tokens (30 minutes)
	AccessTokenValidity = 1800 * time.Second

	// RefreshTokenValidity is the validity period for refresh tokens (7 days)
	RefreshTokenValidity = 604800 * time.Second

	// AccessTokenRefreshThreshold is the time before expiry to refresh access token (61 seconds)
	AccessTokenRefreshThreshold = 61 * time.Second

	// RefreshTokenRefreshThreshold is the time before expiry to refresh refresh token (60.5 minutes)
	RefreshTokenRefreshThreshold = 3630 * time.Second
)

// WebSocket Streaming Constants
const (
	// WSPingInterval is the interval between WebSocket ping messages
	WSPingInterval = 20 * time.Second

	// WSPingTimeout is how long to wait for pong responses from the server
	WSPingTimeout = 30 * time.Second

	// WSReconnectBackoffInitial is the initial backoff time for reconnection attempts
	WSReconnectBackoffInitial = 2 * time.Second

	// WSReconnectBackoffMax is the maximum backoff time for reconnection attempts
	WSReconnectBackoffMax = 120 * time.Second

	// WSCrashThreshold is the threshold for detecting stream crashes
	WSCrashThreshold = 90 * time.Second

	// WSLoopReadyWait is the timeout for waiting for the event loop to be ready
	WSLoopReadyWait = 4 * time.Second

	// WSCloseTimeout is the timeout for closing the WebSocket connection
	WSCloseTimeout = 5 * time.Second

	// WSThreadJoinTimeout is the timeout for joining the streaming thread
	WSThreadJoinTimeout = 5 * time.Second
)

// Background Task Constants
const (
	// TokenCheckerSleep is the sleep interval for the token checker background task
	TokenCheckerSleep = 30 * time.Second

	// AutoCheckerSleep is the sleep interval for the auto checker background task
	AutoCheckerSleep = 30 * time.Second
)

// Validation Constants
const (
	// AppKeyLength1 is the first valid length for app keys
	AppKeyLength1 = 32

	// AppKeyLength2 is the second valid length for app keys
	AppKeyLength2 = 48

	// AppSecretLength1 is the first valid length for app secrets
	AppSecretLength1 = 16

	// AppSecretLength2 is the second valid length for app secrets
	AppSecretLength2 = 64
)

// Encryption Constants
const (
	// EncryptionPrefix is the prefix added to encrypted token values
	EncryptionPrefix = "enc:"
)

// ============================================================================
// OpenAPI Enum Constants — Trader API
// The following constants correspond to the enum values defined in the
// Schwab Trader API OpenAPI specification.
// ============================================================================

// Instruction enum — the action of the order on a leg.
const (
	InstructionBuy             = "BUY"
	InstructionSell            = "SELL"
	InstructionBuyToCover      = "BUY_TO_COVER"
	InstructionSellShort       = "SELL_SHORT"
	InstructionBuyToOpen       = "BUY_TO_OPEN"
	InstructionBuyToClose      = "BUY_TO_CLOSE"
	InstructionSellToOpen      = "SELL_TO_OPEN"
	InstructionSellToClose     = "SELL_TO_CLOSE"
	InstructionExchange        = "EXCHANGE"
	InstructionSellShortExempt = "SELL_SHORT_EXEMPT"
)

// Duration enum — how long the order stays active.
const (
	DurationDay               = "DAY"
	DurationGoodTillCancel    = "GOOD_TILL_CANCEL"
	DurationFillOrKill        = "FILL_OR_KILL"
	DurationImmediateOrCancel = "IMMEDIATE_OR_CANCEL"
	DurationEndOfWeek         = "END_OF_WEEK"
	DurationEndOfMonth        = "END_OF_MONTH"
	DurationNextEndOfMonth    = "NEXT_END_OF_MONTH"
	DurationUnknown           = "UNKNOWN"
)

// OrderType enum — the type of the order (response, includes UNKNOWN).
const (
	OrderTypeMarket            = "MARKET"
	OrderTypeLimit             = "LIMIT"
	OrderTypeStop              = "STOP"
	OrderTypeStopLimit         = "STOP_LIMIT"
	OrderTypeTrailingStop      = "TRAILING_STOP"
	OrderTypeCabinet           = "CABINET"
	OrderTypeNonMarketable     = "NON_MARKETABLE"
	OrderTypeMarketOnClose     = "MARKET_ON_CLOSE"
	OrderTypeExercise          = "EXERCISE"
	OrderTypeTrailingStopLimit = "TRAILING_STOP_LIMIT"
	OrderTypeNetDebit          = "NET_DEBIT"
	OrderTypeNetCredit         = "NET_CREDIT"
	OrderTypeNetZero           = "NET_ZERO"
	OrderTypeLimitOnClose      = "LIMIT_ON_CLOSE"
	OrderTypeUnknown           = "UNKNOWN"
)

// OrderTypeRequest enum — the type of the order for request bodies.
// Same as OrderType but excludes UNKNOWN (which is only valid in responses).
const (
	OrderTypeRequestMarket            = "MARKET"
	OrderTypeRequestLimit             = "LIMIT"
	OrderTypeRequestStop              = "STOP"
	OrderTypeRequestStopLimit         = "STOP_LIMIT"
	OrderTypeRequestTrailingStop      = "TRAILING_STOP"
	OrderTypeRequestCabinet           = "CABINET"
	OrderTypeRequestNonMarketable     = "NON_MARKETABLE"
	OrderTypeRequestMarketOnClose     = "MARKET_ON_CLOSE"
	OrderTypeRequestExercise          = "EXERCISE"
	OrderTypeRequestTrailingStopLimit = "TRAILING_STOP_LIMIT"
	OrderTypeRequestNetDebit          = "NET_DEBIT"
	OrderTypeRequestNetCredit         = "NET_CREDIT"
	OrderTypeRequestNetZero           = "NET_ZERO"
	OrderTypeRequestLimitOnClose      = "LIMIT_ON_CLOSE"
)

// OrderStrategyType enum — the strategy type for the order.
const (
	OrderStrategyTypeSingle     = "SINGLE"
	OrderStrategyTypeCancel     = "CANCEL"
	OrderStrategyTypeRecall     = "RECALL"
	OrderStrategyTypePair       = "PAIR"
	OrderStrategyTypeFlatten    = "FLATTEN"
	OrderStrategyTypeTwoDaySwap = "TWO_DAY_SWAP"
	OrderStrategyTypeBlastAll   = "BLAST_ALL"
	OrderStrategyTypeOCO        = "OCO"
	OrderStrategyTypeTrigger    = "TRIGGER"
)

// ComplexOrderStrategyType enum — the complex strategy type for the order.
const (
	ComplexOrderStrategyNone                   = "NONE"
	ComplexOrderStrategyCovered                = "COVERED"
	ComplexOrderStrategyVertical               = "VERTICAL"
	ComplexOrderStrategyBackRatio              = "BACK_RATIO"
	ComplexOrderStrategyCalendar               = "CALENDAR"
	ComplexOrderStrategyDiagonal               = "DIAGONAL"
	ComplexOrderStrategyStraddle               = "STRADDLE"
	ComplexOrderStrategyStrangle               = "STRANGLE"
	ComplexOrderStrategyCollarSynthetic        = "COLLAR_SYNTHETIC"
	ComplexOrderStrategyButterfly              = "BUTTERFLY"
	ComplexOrderStrategyCondor                 = "CONDOR"
	ComplexOrderStrategyIronCondor             = "IRON_CONDOR"
	ComplexOrderStrategyVerticalRoll           = "VERTICAL_ROLL"
	ComplexOrderStrategyCollarWithStock        = "COLLAR_WITH_STOCK"
	ComplexOrderStrategyDoubleDiagonal         = "DOUBLE_DIAGONAL"
	ComplexOrderStrategyUnbalancedButterfly    = "UNBALANCED_BUTTERFLY"
	ComplexOrderStrategyUnbalancedCondor       = "UNBALANCED_CONDOR"
	ComplexOrderStrategyUnbalancedIronCondor   = "UNBALANCED_IRON_CONDOR"
	ComplexOrderStrategyUnbalancedVerticalRoll = "UNBALANCED_VERTICAL_ROLL"
	ComplexOrderStrategyMutualFundSwap         = "MUTUAL_FUND_SWAP"
	ComplexOrderStrategyCustom                 = "CUSTOM"
)

// Session enum — the market session for the order.
const (
	SessionNormal   = "NORMAL"
	SessionAM       = "AM"
	SessionPM       = "PM"
	SessionSeamless = "SEAMLESS"
)

// Status enum — the current status of the order.
const (
	OrderStatusAwaitingParentOrder    = "AWAITING_PARENT_ORDER"
	OrderStatusAwaitingCondition      = "AWAITING_CONDITION"
	OrderStatusAwaitingStopCondition  = "AWAITING_STOP_CONDITION"
	OrderStatusAwaitingManualReview   = "AWAITING_MANUAL_REVIEW"
	OrderStatusAccepted               = "ACCEPTED"
	OrderStatusAwaitingURouting       = "AWAITING_UR_OUT"
	OrderStatusPendingActivation      = "PENDING_ACTIVATION"
	OrderStatusQueued                 = "QUEUED"
	OrderStatusWorking                = "WORKING"
	OrderStatusRejected                = "REJECTED"
	OrderStatusPendingCancel          = "PENDING_CANCEL"
	OrderStatusCanceled                = "CANCELED"
	OrderStatusPendingReplace         = "PENDING_REPLACE"
	OrderStatusReplaced               = "REPLACED"
	OrderStatusFilled                 = "FILLED"
	OrderStatusExpired                = "EXPIRED"
	OrderStatusNew                    = "NEW"
	OrderStatusAwaitingReleaseTime    = "AWAITING_RELEASE_TIME"
	OrderStatusPendingAcknowledgement = "PENDING_ACKNOWLEDGEMENT"
	OrderStatusPendingRecall          = "PENDING_RECALL"
	OrderStatusUnknown                = "UNKNOWN"
)

// AssetType enum — the type of asset for the instrument.
const (
	AssetTypeEquity               = "EQUITY"
	AssetTypeMutualFund           = "MUTUAL_FUND"
	AssetTypeOption               = "OPTION"
	AssetTypeFuture               = "FUTURE"
	AssetTypeForex                = "FOREX"
	AssetTypeIndex                = "INDEX"
	AssetTypeCashEquivalent       = "CASH_EQUIVALENT"
	AssetTypeFixedIncome          = "FIXED_INCOME"
	AssetTypeProduct              = "PRODUCT"
	AssetTypeCurrency             = "CURRENCY"
	AssetTypeCollectiveInvestment = "COLLECTIVE_INVESTMENT"
)

// SpecialInstruction enum — special handling instructions for the order.
const (
	SpecialInstructionAllOrNone            = "ALL_OR_NONE"
	SpecialInstructionDoNotReduce          = "DO_NOT_REDUCE"
	SpecialInstructionAllOrNoneDoNotReduce = "ALL_OR_NONE_DO_NOT_REDUCE"
)

// StopPriceLinkBasis enum — the basis for linking the stop price.
const (
	StopPriceLinkBasisManual  = "MANUAL"
	StopPriceLinkBasisBase    = "BASE"
	StopPriceLinkBasisTrigger = "TRIGGER"
	StopPriceLinkBasisLast    = "LAST"
	StopPriceLinkBasisBid     = "BID"
	StopPriceLinkBasisAsk     = "ASK"
	StopPriceLinkBasisAskBid  = "ASK_BID"
	StopPriceLinkBasisMark    = "MARK"
	StopPriceLinkBasisAverage = "AVERAGE"
)

// StopPriceLinkType enum — the type of link for the stop price.
const (
	StopPriceLinkTypeValue   = "VALUE"
	StopPriceLinkTypePercent = "PERCENT"
	StopPriceLinkTypeTick    = "TICK"
)

// StopType enum — the type of stop for the order.
const (
	StopTypeStandard = "STANDARD"
	StopTypeBid      = "BID"
	StopTypeAsk      = "ASK"
	StopTypeLast     = "LAST"
	StopTypeMark     = "MARK"
)

// PriceLinkBasis enum — the basis for linking the price.
const (
	PriceLinkBasisManual  = "MANUAL"
	PriceLinkBasisBase    = "BASE"
	PriceLinkBasisTrigger = "TRIGGER"
	PriceLinkBasisLast    = "LAST"
	PriceLinkBasisBid     = "BID"
	PriceLinkBasisAsk     = "ASK"
	PriceLinkBasisAskBid  = "ASK_BID"
	PriceLinkBasisMark    = "MARK"
	PriceLinkBasisAverage = "AVERAGE"
)

// PriceLinkType enum — the type of link for the price.
const (
	PriceLinkTypeValue   = "VALUE"
	PriceLinkTypePercent = "PERCENT"
	PriceLinkTypeTick    = "TICK"
)

// TaxLotMethod enum — the method used for tax lot selection.
const (
	TaxLotMethodFIFO          = "FIFO"
	TaxLotMethodLIFO          = "LIFO"
	TaxLotMethodHighCost      = "HIGH_COST"
	TaxLotMethodLowCost       = "LOW_COST"
	TaxLotMethodAverageCost   = "AVERAGE_COST"
	TaxLotMethodSpecificLot   = "SPECIFIC_LOT"
	TaxLotMethodLossHarvester = "LOSS_HARVESTER"
)

// RequestedDestination enum — the requested routing destination for the order.
const (
	RequestedDestinationINET    = "INET"
	RequestedDestinationECNArca = "ECN_ARCA"
	RequestedDestinationCBOE    = "CBOE"
	RequestedDestinationAMEX    = "AMEX"
	RequestedDestinationPHLX    = "PHLX"
	RequestedDestinationISE     = "ISE"
	RequestedDestinationBOX     = "BOX"
	RequestedDestinationNYSE    = "NYSE"
	RequestedDestinationNASDAQ  = "NASDAQ"
	RequestedDestinationBATS    = "BATS"
	RequestedDestinationC2      = "C2"
	RequestedDestinationAuto    = "AUTO"
)

// TransactionType enum — the type of account transaction.
const (
	TransactionTypeTrade              = "TRADE"
	TransactionTypeReceiveAndDeliver  = "RECEIVE_AND_DELIVER"
	TransactionTypeDividendOrInterest = "DIVIDEND_OR_INTEREST"
	TransactionTypeACHReceipt         = "ACH_RECEIPT"
	TransactionTypeACHDisbursement    = "ACH_DISBURSEMENT"
	TransactionTypeCashReceipt        = "CASH_RECEIPT"
	TransactionTypeCashDisbursement   = "CASH_DISBURSEMENT"
	TransactionTypeElectronicFund     = "ELECTRONIC_FUND"
	TransactionTypeWireOut            = "WIRE_OUT"
	TransactionTypeWireIn             = "WIRE_IN"
	TransactionTypeJournal            = "JOURNAL"
	TransactionTypeMemorandum         = "MEMORANDUM"
	TransactionTypeMarginCall         = "MARGIN_CALL"
	TransactionTypeMoneyMarket        = "MONEY_MARKET"
	TransactionTypeSmaAdjustment      = "SMA_ADJUSTMENT"
)

// PositionEffect enum (OrderLegCollection) — the effect on the position.
const (
	PositionEffectOpening   = "OPENING"
	PositionEffectClosing   = "CLOSING"
	PositionEffectAutomatic = "AUTOMATIC"
)

// QuantityType enum (OrderLegCollection) — the type of quantity specified.
const (
	QuantityTypeAllShares = "ALL_SHARES"
	QuantityTypeDollars   = "DOLLARS"
	QuantityTypeShares    = "SHARES"
)

// DivCapGains enum (OrderLegCollection) — dividend / capital gains handling.
const (
	DivCapGainsReinvest = "REINVEST"
	DivCapGainsPayout   = "PAYOUT"
)

// SubAccount enum (Transaction) — the sub-account type.
const (
	SubAccountCash    = "CASH"
	SubAccountMargin  = "MARGIN"
	SubAccountShort   = "SHORT"
	SubAccountDiv     = "DIV"
	SubAccountIncome  = "INCOME"
	SubAccountUnknown = "UNKNOWN"
)

// TransactionStatus enum (Transaction) — the status of a transaction.
const (
	TransactionStatusValid   = "VALID"
	TransactionStatusInvalid = "INVALID"
	TransactionStatusPending = "PENDING"
	TransactionStatusUnknown = "UNKNOWN"
)

// TransactionActivityType enum (Transaction) — the activity type of a transaction.
const (
	TransactionActivityTypeActivityCorrection = "ACTIVITY_CORRECTION"
	TransactionActivityTypeExecution          = "EXECUTION"
	TransactionActivityTypeOrderAction        = "ORDER_ACTION"
	TransactionActivityTypeTransfer           = "TRANSFER"
	TransactionActivityTypeUnknown            = "UNKNOWN"
)

// AdvancedOrderType enum (OrderStrategy) — the type of advanced order.
const (
	AdvancedOrderTypeNone     = "NONE"
	AdvancedOrderTypeOTO      = "OTO"
	AdvancedOrderTypeOCO      = "OCO"
	AdvancedOrderTypeOTOCO    = "OTOCO"
	AdvancedOrderTypeOT2OCO   = "OT2OCO"
	AdvancedOrderTypeOT3OCO   = "OT3OCO"
	AdvancedOrderTypeBlastAll = "BLAST_ALL"
	AdvancedOrderTypeOTA      = "OTA"
	AdvancedOrderTypePair     = "PAIR"
)

// AmountIndicator enum (OrderStrategy) — indicates the amount basis.
const (
	AmountIndicatorDollars    = "DOLLARS"
	AmountIndicatorShares     = "SHARES"
	AmountIndicatorAllShares  = "ALL_SHARES"
	AmountIndicatorPercentage = "PERCENTAGE"
	AmountIndicatorUnknown    = "UNKNOWN"
)

// SettlementInstruction enum (OrderStrategy) — the settlement instruction.
const (
	SettlementInstructionRegular = "REGULAR"
	SettlementInstructionCash    = "CASH"
	SettlementInstructionNextDay = "NEXT_DAY"
	SettlementInstructionUnknown = "UNKNOWN"
)

// ActivityType enum (OrderActivity) — the type of order activity.
const (
	ActivityTypeExecution   = "EXECUTION"
	ActivityTypeOrderAction = "ORDER_ACTION"
)

// ExecutionType enum (OrderActivity) — the type of execution.
const (
	ExecutionTypeFill = "FILL"
)

// OptionPutCall enum — whether an option is a put or a call.
const (
	OptionPutCallPut     = "PUT"
	OptionPutCallCall    = "CALL"
	OptionPutCallUnknown = "UNKNOWN"
)

// OptionType enum — the style/type of the option.
const (
	OptionTypeVanilla = "VANILLA"
	OptionTypeBinary  = "BINARY"
	OptionTypeBarrier = "BARRIER"
	OptionTypeUnknown = "UNKNOWN"
)

// ApiRuleAction enum (OrderValidationDetail) — the action taken by an API rule.
const (
	ApiRuleActionAccept  = "ACCEPT"
	ApiRuleActionAlert   = "ALERT"
	ApiRuleActionReject  = "REJECT"
	ApiRuleActionReview  = "REVIEW"
	ApiRuleActionUnknown = "UNKNOWN"
)

// FeeType enum — the type of fee applied to an order.
const (
	FeeTypeCommission             = "COMMISSION"
	FeeTypeSecFee                 = "SEC_FEE"
	FeeTypeStrFee                 = "STR_FEE"
	FeeTypeRFee                   = "R_FEE"
	FeeTypeCdscFee                = "CDSC_FEE"
	FeeTypeOptRegFee              = "OPT_REG_FEE"
	FeeTypeAdditionalFee          = "ADDITIONAL_FEE"
	FeeTypeMiscellaneousFee       = "MISCELLANEOUS_FEE"
	FeeTypeFTT                    = "FTT"
	FeeTypeFuturesClearingFee     = "FUTURES_CLEARING_FEE"
	FeeTypeFuturesDeskOfficeFee   = "FUTURES_DESK_OFFICE_FEE"
	FeeTypeFuturesExchangeFee     = "FUTURES_EXCHANGE_FEE"
	FeeTypeFuturesGlobexFee       = "FUTURES_GLOBEX_FEE"
	FeeTypeFuturesNfaFee          = "FUTURES_NFA_FEE"
	FeeTypeFuturesPitBrokerageFee = "FUTURES_PIT_BROKERAGE_FEE"
	FeeTypeFuturesTransactionFee  = "FUTURES_TRANSACTION_FEE"
	FeeTypeLowProceedsCommission  = "LOW_PROCEEDS_COMMISSION"
	FeeTypeBaseCharge             = "BASE_CHARGE"
	FeeTypeGeneralCharge          = "GENERAL_CHARGE"
	FeeTypeGstFee                 = "GST_FEE"
	FeeTypeTafFee                 = "TAF_FEE"
	FeeTypeIndexOptionFee         = "INDEX_OPTION_FEE"
	FeeTypeTefraTax               = "TEFRA_TAX"
	FeeTypeStateTax               = "STATE_TAX"
	FeeTypeUnknown                = "UNKNOWN"
)

// TransferItemPositionEffect enum (TransferItem) — the effect on the position for a transfer.
const (
	TransferItemPositionEffectOpening   = "OPENING"
	TransferItemPositionEffectClosing   = "CLOSING"
	TransferItemPositionEffectAutomatic = "AUTOMATIC"
	TransferItemPositionEffectUnknown   = "UNKNOWN"
)

// ============================================================================
// OpenAPI Enum Constants — Market Data API
// The following constants correspond to the enum values defined in the
// Schwab Market Data API OpenAPI specification.
// ============================================================================

// AssetMainType enum — the main asset type for quote responses.
const (
	AssetMainTypeBond         = "BOND"
	AssetMainTypeEquity       = "EQUITY"
	AssetMainTypeForex        = "FOREX"
	AssetMainTypeFuture       = "FUTURE"
	AssetMainTypeFutureOption = "FUTURE_OPTION"
	AssetMainTypeIndex        = "INDEX"
	AssetMainTypeMutualFund   = "MUTUAL_FUND"
	AssetMainTypeOption       = "OPTION"
)

// EquityAssetSubType enum — asset sub-type for equity quotes.
const (
	EquityAssetSubTypeCOE = "COE"
	EquityAssetSubTypePRF = "PRF"
	EquityAssetSubTypeADR = "ADR"
	EquityAssetSubTypeGDR = "GDR"
	EquityAssetSubTypeCEF = "CEF"
	EquityAssetSubTypeETF = "ETF"
	EquityAssetSubTypeETN = "ETN"
	EquityAssetSubTypeUIT = "UIT"
	EquityAssetSubTypeWAR = "WAR"
	EquityAssetSubTypeRGT = "RGT"
)

// MutualFundAssetSubType enum — asset sub-type for mutual fund quotes.
const (
	MutualFundAssetSubTypeOEF = "OEF"
	MutualFundAssetSubTypeCEF = "CEF"
	MutualFundAssetSubTypeMMF = "MMF"
)

// QuoteType enum — the quote type for quote responses.
const (
	QuoteTypeNBBO = "NBBO"
	QuoteTypeNFL  = "NFL"
)

// DivFreq enum — dividend frequency for fundamental data.
const (
	DivFreqAnnually      = 1
	DivFreqSemiAnnually  = 2
	DivFreqThreeTimesYr  = 3
	DivFreqQuarterly     = 4
	DivFreqBiMonthly     = 6
	DivFreqElevenTimesYr = 11
	DivFreqMonthly       = 12
)

// FundStrategy enum — fund strategy for fundamental data.
const (
	FundStrategyActive      = "A"
	FundStrategyLeveraged    = "L"
	FundStrategyPassive      = "P"
	FundStrategyQuantitative = "Q"
	FundStrategyShort        = "S"
)

// MDContractType enum — option contract type (Put/Call) for market data responses.
// Note: This is different from the chain query parameter ContractType (CALL/PUT/ALL).
const (
	MDContractTypePut  = "P"
	MDContractTypeCall = "C"
)

// MDExpirationType enum — expiration type for market data option responses.
const (
	MDExpirationTypeMonthly   = "M"
	MDExpirationTypeQuarterly = "Q"
	MDExpirationTypeStandard  = "S"
	MDExpirationTypeWeekly    = "W"
)

// MDSettlementType enum — settlement type for market data option responses.
const (
	MDSettlementTypeAM = "A"
	MDSettlementTypePM = "P"
)

// MDExerciseType enum — exercise type for market data option responses.
const (
	MDExerciseTypeAmerican  = "A"
	MDExerciseTypeEuropean  = "E"
)

// Market ID Constants (for /marketdata/v1/markets/{market_id})
const (
	MarketIDEquity = "equity"
	MarketIDOption = "option"
	MarketIDBond   = "bond"
	MarketIDFuture = "future"
	MarketIDForex  = "forex"
)

// Option Chain Contract Type Constants (for /marketdata/v1/chains contractType query param)
const (
	ContractTypeCall = "CALL"
	ContractTypePut  = "PUT"
	ContractTypeAll  = "ALL"
)

// ChainStrategy enum — strategy for option chain query parameter.
const (
	ChainStrategySingle     = "SINGLE"
	ChainStrategyAnalytical = "ANALYTICAL"
	ChainStrategyCovered    = "COVERED"
	ChainStrategyVertical   = "VERTICAL"
	ChainStrategyCalendar   = "CALENDAR"
	ChainStrategyStrangle   = "STRANGLE"
	ChainStrategyStraddle   = "STRADDLE"
	ChainStrategyButterfly  = "BUTTERFLY"
	ChainStrategyCondor     = "CONDOR"
	ChainStrategyDiagonal   = "DIAGONAL"
	ChainStrategyCollar     = "COLLAR"
	ChainStrategyRoll       = "ROLL"
)

// ExpMonth enum — expiration month for option chain query parameter.
const (
	ExpMonthJan = "JAN"
	ExpMonthFeb = "FEB"
	ExpMonthMar = "MAR"
	ExpMonthApr = "APR"
	ExpMonthMay = "MAY"
	ExpMonthJun = "JUN"
	ExpMonthJul = "JUL"
	ExpMonthAug = "AUG"
	ExpMonthSep = "SEP"
	ExpMonthOct = "OCT"
	ExpMonthNov = "NOV"
	ExpMonthDec = "DEC"
	ExpMonthAll = "ALL"
)

// Entitlement enum — entitlement for option chain query parameter.
const (
	EntitlementPayingPro = "PN"
	EntitlementNonPaying = "NP"
	EntitlementPro       = "PP"
)

// SortType enum — sort type for movers query parameter.
const (
	SortTypeVolume            = "VOLUME"
	SortTypeTrades            = "TRADES"
	SortTypePercentChangeUp   = "PERCENT_CHANGE_UP"
	SortTypePercentChangeDown = "PERCENT_CHANGE_DOWN"
)

// MoverFrequency enum — frequency for movers query parameter.
const (
	MoverFrequency0  = 0
	MoverFrequency1  = 1
	MoverFrequency5  = 5
	MoverFrequency10 = 10
	MoverFrequency30 = 30
	MoverFrequency60 = 60
)

// MoverSymbol enum — symbol for movers path parameter.
const (
	MoverSymbolDJI        = "$DJI"
	MoverSymbolCOMPX      = "$COMPX"
	MoverSymbolSPX        = "$SPX"
	MoverSymbolNYSE       = "NYSE"
	MoverSymbolNASDAQ     = "NASDAQ"
	MoverSymbolOTCBB      = "OTCBB"
	MoverSymbolIndexAll   = "INDEX_ALL"
	MoverSymbolEquityAll  = "EQUITY_ALL"
	MoverSymbolOptionAll  = "OPTION_ALL"
	MoverSymbolOptionPut  = "OPTION_PUT"
	MoverSymbolOptionCall = "OPTION_CALL"
)

// PeriodType enum — period type for price history query parameter.
const (
	PeriodTypeDay   = "day"
	PeriodTypeMonth = "month"
	PeriodTypeYear  = "year"
	PeriodTypeYTD   = "ytd"
)

// FrequencyType enum — frequency type for price history query parameter.
const (
	FrequencyTypeMinute  = "minute"
	FrequencyTypeDaily   = "daily"
	FrequencyTypeWeekly  = "weekly"
	FrequencyTypeMonthly = "monthly"
)

// Instrument Projection Constants
const (
	ProjectionSymbolSearch = "symbol-search"
	ProjectionSymbolRegex  = "symbol-regex"
	ProjectionDescSearch   = "desc-search"
	ProjectionDescRegex    = "desc-regex"
	ProjectionFundamental  = "fundamental"
	ProjectionSearch       = "search"
)
