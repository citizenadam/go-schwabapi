package types

// TimeFormat represents the time format for API requests
type TimeFormat string

const (
	// TimeFormatISO8601 is ISO 8601 format
	TimeFormatISO8601 TimeFormat = "8601"
	// TimeFormatEpoch is Unix epoch format
	TimeFormatEpoch TimeFormat = "epoch"
	// TimeFormatEpochMS is Unix epoch in milliseconds
	TimeFormatEpochMS TimeFormat = "epoch_ms"
	// TimeFormatYYYYMMDD is YYYY-MM-DD format
	TimeFormatYYYYMMDD TimeFormat = "YYYY-MM-DD"
)

// Service represents the streaming service type
type Service string

const (
	// ServiceAdmin is the admin service
	ServiceAdmin Service = "ADMIN"
	// ServiceLevelOneEquities is level one equities service
	ServiceLevelOneEquities Service = "LEVELONE_EQUITIES"
	// ServiceLevelOneOptions is level one options service
	ServiceLevelOneOptions Service = "LEVELONE_OPTIONS"
	// ServiceLevelOneFutures is level one futures service
	ServiceLevelOneFutures Service = "LEVELONE_FUTURES"
	// ServiceLevelOneFuturesOptions is level one futures options service
	ServiceLevelOneFuturesOptions Service = "LEVELONE_FUTURES_OPTIONS"
	// ServiceLevelOneForex is level one forex service
	ServiceLevelOneForex Service = "LEVELONE_FOREX"
	// ServiceNYSEBook is NYSE book service
	ServiceNYSEBook Service = "NYSE_BOOK"
	// ServiceNASDAQBook is NASDAQ book service
	ServiceNASDAQBook Service = "NASDAQ_BOOK"
	// ServiceOptionsBook is options book service
	ServiceOptionsBook Service = "OPTIONS_BOOK"
	// ServiceChartEquity is chart equity service
	ServiceChartEquity Service = "CHART_EQUITY"
	// ServiceChartFutures is chart futures service
	ServiceChartFutures Service = "CHART_FUTURES"
	// ServiceScreenerEquity is screener equity service
	ServiceScreenerEquity Service = "SCREENER_EQUITY"
	// ServiceScreenerOption is screener option service
	ServiceScreenerOption Service = "SCREENER_OPTION"
	// ServiceAccountActivity is account activity service
	ServiceAccountActivity Service = "ACCT_ACTIVITY"
)

// Command represents the streaming command type
type Command string

const (
	// CommandLogin is login command
	CommandLogin Command = "LOGIN"
	// CommandLogout is logout command
	CommandLogout Command = "LOGOUT"
	// CommandAdd is add subscription command
	CommandAdd Command = "ADD"
	// CommandSubs is subscribe command
	CommandSubs Command = "SUBS"
	// CommandUnsubs is unsubscribe command
	CommandUnsubs Command = "UNSUBS"
	// CommandView is view command
	CommandView Command = "VIEW"
)
