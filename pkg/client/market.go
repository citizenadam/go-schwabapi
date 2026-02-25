package client

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/citizenadam/go-schwabapi/pkg/types"
)

// Market handles market data API endpoints
type Market struct {
	httpClient  *Client
	logger      *slog.Logger
	tokenGetter TokenGetter
}

// NewMarket creates a new Market client
func NewMarket(httpClient *Client, logger *slog.Logger, tokenGetter TokenGetter) *Market {
	return &Market{
		httpClient:  httpClient,
		logger:      logger,
		tokenGetter: tokenGetter,
	}
}

// Quotes retrieves quotes for a list of symbols
// Endpoint: GET /marketdata/v1/quotes
func (m *Market) Quotes(ctx context.Context, symbols []string, fields string, indicative bool) (*types.QuotesResponse, error) {
	// Create context with deadline to prevent indefinite blocking
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	apiURL := fmt.Sprintf("%s/marketdata/v1/quotes", baseAPIURL)

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", m.tokenGetter.GetAccessToken()),
		"Accept":        "application/json",
	}

	// Build query parameters
	params := url.Values{}
	params.Add("symbols", strings.Join(symbols, ","))
	if fields != "" {
		params.Add("fields", fields)
	}
	if indicative {
		params.Add("indicative", "true")
	}

	// Append query string to URL
	apiURL = fmt.Sprintf("%s?%s", apiURL, params.Encode())

	resp, err := m.httpClient.Get(ctx, apiURL, headers)
	if err != nil {
		m.logger.Error("failed to get quotes",
			"url", apiURL,
			"symbols", symbols,
			"error", err,
		)
		return nil, fmt.Errorf("failed to get quotes: %w", err)
	}

	var result types.QuotesResponse
	if err := m.httpClient.DecodeJSON(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode quotes response: %w", err)
	}

	m.logger.Info("successfully retrieved quotes",
		"count", len(result.Quotes),
	)

	return &result, nil
}

// Quote retrieves a quote for a single symbol
// Endpoint: GET /marketdata/v1/{symbol}/quotes
func (m *Market) Quote(ctx context.Context, symbol string, fields string) (*types.QuoteResponse, error) {
	// Create context with deadline to prevent indefinite blocking
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	apiURL := fmt.Sprintf("%s/marketdata/v1/%s/quotes", baseAPIURL, url.PathEscape(symbol))

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", m.tokenGetter.GetAccessToken()),
		"Accept":        "application/json",
	}

	// Build query parameters
	params := url.Values{}
	if fields != "" {
		params.Add("fields", fields)
	}

	// Append query string to URL if we have parameters
	if len(params) > 0 {
		apiURL = fmt.Sprintf("%s?%s", apiURL, params.Encode())
	}

	resp, err := m.httpClient.Get(ctx, apiURL, headers)
	if err != nil {
		m.logger.Error("failed to get quote",
			"url", apiURL,
			"symbol", symbol,
			"error", err,
		)
		return nil, fmt.Errorf("failed to get quote: %w", err)
	}

	var result types.QuoteResponse
	if err := m.httpClient.DecodeJSON(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode quote response: %w", err)
	}

	m.logger.Info("successfully retrieved quote",
		"symbol", symbol,
	)

	return &result, nil
}

// Movers retrieves market movers for an index
// Endpoint: GET /marketdata/v1/movers
func (m *Market) Movers(ctx context.Context, index string, direction string, change string) (*types.MoversResponse, error) {
	// Create context with deadline to prevent indefinite blocking
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	apiURL := fmt.Sprintf("%s/marketdata/v1/movers", baseAPIURL)

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", m.tokenGetter.GetAccessToken()),
		"Accept":        "application/json",
	}

	// Build query parameters
	params := url.Values{}
	params.Add("index", index)
	params.Add("direction", direction)
	params.Add("change", change)

	// Append query string to URL
	apiURL = fmt.Sprintf("%s?%s", apiURL, params.Encode())

	resp, err := m.httpClient.Get(ctx, apiURL, headers)
	if err != nil {
		m.logger.Error("failed to get movers",
			"url", apiURL,
			"index", index,
			"error", err,
		)
		return nil, fmt.Errorf("failed to get movers: %w", err)
	}

	var result types.MoversResponse
	if err := m.httpClient.DecodeJSON(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode movers response: %w", err)
	}

	m.logger.Info("successfully retrieved movers",
		"index", index,
		"count", len(result.Movers),
	)

	return &result, nil
}

// OptionChainsRequest represents parameters for option chains request
type OptionChainsRequest struct {
	Symbol                 string
	ContractType           string
	StrikeCount            int
	IncludeUnderlyingQuote bool
	Strategy               string
	Interval               string
	Strike                 float64
	Range                  string
	FromDate               string
	ToDate                 string
	Volatility             float64
	UnderlyingPrice        float64
	InterestRate           float64
	DaysToExpiration       int
	ExpMonth               string
	OptionType             string
	Entitlement            string
}

// OptionChains retrieves option chains for a symbol
// Endpoint: GET /marketdata/v1/chains
func (m *Market) OptionChains(ctx context.Context, req *OptionChainsRequest) (*types.OptionChainsResponse, error) {
	// Create context with deadline to prevent indefinite blocking
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	apiURL := fmt.Sprintf("%s/marketdata/v1/chains", baseAPIURL)

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", m.tokenGetter.GetAccessToken()),
		"Accept":        "application/json",
	}

	// Build query parameters
	params := url.Values{}
	params.Add("symbol", req.Symbol)
	if req.ContractType != "" {
		params.Add("contractType", req.ContractType)
	}
	if req.StrikeCount > 0 {
		params.Add("strikeCount", fmt.Sprintf("%d", req.StrikeCount))
	}
	if req.IncludeUnderlyingQuote {
		params.Add("includeUnderlyingQuote", "true")
	}
	if req.Strategy != "" {
		params.Add("strategy", req.Strategy)
	}
	if req.Interval != "" {
		params.Add("interval", req.Interval)
	}
	if req.Strike > 0 {
		params.Add("strike", fmt.Sprintf("%f", req.Strike))
	}
	if req.Range != "" {
		params.Add("range", req.Range)
	}
	if req.FromDate != "" {
		params.Add("fromDate", req.FromDate)
	}
	if req.ToDate != "" {
		params.Add("toDate", req.ToDate)
	}
	if req.Volatility > 0 {
		params.Add("volatility", fmt.Sprintf("%f", req.Volatility))
	}
	if req.UnderlyingPrice > 0 {
		params.Add("underlyingPrice", fmt.Sprintf("%f", req.UnderlyingPrice))
	}
	if req.InterestRate > 0 {
		params.Add("interestRate", fmt.Sprintf("%f", req.InterestRate))
	}
	if req.DaysToExpiration > 0 {
		params.Add("daysToExpiration", fmt.Sprintf("%d", req.DaysToExpiration))
	}
	if req.ExpMonth != "" {
		params.Add("expMonth", req.ExpMonth)
	}
	if req.OptionType != "" {
		params.Add("optionType", req.OptionType)
	}
	if req.Entitlement != "" {
		params.Add("entitlement", req.Entitlement)
	}

	// Append query string to URL
	apiURL = fmt.Sprintf("%s?%s", apiURL, params.Encode())

	resp, err := m.httpClient.Get(ctx, apiURL, headers)
	if err != nil {
		m.logger.Error("failed to get option chains",
			"url", apiURL,
			"symbol", req.Symbol,
			"error", err,
		)
		return nil, fmt.Errorf("failed to get option chains: %w", err)
	}

	var result types.OptionChainsResponse
	if err := m.httpClient.DecodeJSON(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode option chains response: %w", err)
	}

	m.logger.Info("successfully retrieved option chains",
		"symbol", req.Symbol,
		"numberOfContracts", result.NumberOfContracts,
	)

	return &result, nil
}

// PriceHistoryRequest represents parameters for price history request
type PriceHistoryRequest struct {
	Symbol                string
	PeriodType            string
	Period                string
	FrequencyType         string
	Frequency             string
	StartDate             string
	EndDate               string
	NeedExtendedHoursData bool
	NeedPreviousClose     bool
}

// PriceHistory retrieves price history for a symbol
// Endpoint: GET /marketdata/v1/pricehistory
func (m *Market) PriceHistory(ctx context.Context, req *PriceHistoryRequest) (*types.PriceHistoryResponse, error) {
	// Create context with deadline to prevent indefinite blocking
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	apiURL := fmt.Sprintf("%s/marketdata/v1/pricehistory", baseAPIURL)

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", m.tokenGetter.GetAccessToken()),
		"Accept":        "application/json",
	}

	// Build query parameters
	params := url.Values{}
	params.Add("symbol", req.Symbol)
	if req.PeriodType != "" {
		params.Add("periodType", req.PeriodType)
	}
	if req.Period != "" {
		params.Add("period", req.Period)
	}
	if req.FrequencyType != "" {
		params.Add("frequencyType", req.FrequencyType)
	}
	if req.Frequency != "" {
		params.Add("frequency", req.Frequency)
	}
	if req.StartDate != "" {
		params.Add("startDate", req.StartDate)
	}
	if req.EndDate != "" {
		params.Add("endDate", req.EndDate)
	}
	if req.NeedExtendedHoursData {
		params.Add("needExtendedHoursData", "true")
	}
	if req.NeedPreviousClose {
		params.Add("needPreviousClose", "true")
	}

	// Append query string to URL
	apiURL = fmt.Sprintf("%s?%s", apiURL, params.Encode())

	resp, err := m.httpClient.Get(ctx, apiURL, headers)
	if err != nil {
		m.logger.Error("failed to get price history",
			"url", apiURL,
			"symbol", req.Symbol,
			"error", err,
		)
		return nil, fmt.Errorf("failed to get price history: %w", err)
	}

	var result types.PriceHistoryResponse
	if err := m.httpClient.DecodeJSON(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode price history response: %w", err)
	}

	m.logger.Info("successfully retrieved price history",
		"symbol", req.Symbol,
		"candlesCount", len(result.Candles),
	)

	return &result, nil
}

// MarketHours retrieves hours for multiple markets
// Endpoint: GET /marketdata/v1/markets
func (m *Market) MarketHours(ctx context.Context, markets []string, date string) (*types.MarketHoursResponse, error) {
	// Create context with deadline to prevent indefinite blocking
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	apiURL := fmt.Sprintf("%s/marketdata/v1/markets", baseAPIURL)

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", m.tokenGetter.GetAccessToken()),
		"Accept":        "application/json",
	}

	// Build query parameters
	params := url.Values{}
	params.Add("markets", strings.Join(markets, ","))
	if date != "" {
		params.Add("date", date)
	}

	// Append query string to URL
	apiURL = fmt.Sprintf("%s?%s", apiURL, params.Encode())

	resp, err := m.httpClient.Get(ctx, apiURL, headers)
	if err != nil {
		m.logger.Error("failed to get market hours",
			"url", apiURL,
			"markets", markets,
			"error", err,
		)
		return nil, fmt.Errorf("failed to get market hours: %w", err)
	}

	var result types.MarketHoursResponse
	if err := m.httpClient.DecodeJSON(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode market hours response: %w", err)
	}

	m.logger.Info("successfully retrieved market hours",
		"count", len(result.MarketHours),
	)

	return &result, nil
}

// MarketHour retrieves hours for a single market
// Endpoint: GET /marketdata/v1/markethours/{marketId}
func (m *Market) MarketHour(ctx context.Context, marketId string, date string) (*types.MarketHourResponse, error) {
	// Create context with deadline to prevent indefinite blocking
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	apiURL := fmt.Sprintf("%s/marketdata/v1/markethours/%s", baseAPIURL, url.PathEscape(marketId))

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", m.tokenGetter.GetAccessToken()),
		"Accept":        "application/json",
	}

	// Build query parameters
	params := url.Values{}
	if date != "" {
		params.Add("date", date)
	}

	// Append query string to URL if we have parameters
	if len(params) > 0 {
		apiURL = fmt.Sprintf("%s?%s", apiURL, params.Encode())
	}

	resp, err := m.httpClient.Get(ctx, apiURL, headers)
	if err != nil {
		m.logger.Error("failed to get market hour",
			"url", apiURL,
			"marketId", marketId,
			"error", err,
		)
		return nil, fmt.Errorf("failed to get market hour: %w", err)
	}

	var result types.MarketHourResponse
	if err := m.httpClient.DecodeJSON(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode market hour response: %w", err)
	}

	m.logger.Info("successfully retrieved market hour",
		"marketId", marketId,
	)

	return &result, nil
}

// Instruments retrieves instruments by symbols
// Endpoint: GET /trader/v1/instruments/instruments
func (m *Market) Instruments(ctx context.Context, symbols string, projection string) (*types.InstrumentsResponse, error) {
	// Create context with deadline to prevent indefinite blocking
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	apiURL := fmt.Sprintf("%s/trader/v1/instruments/instruments", baseAPIURL)

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", m.tokenGetter.GetAccessToken()),
		"Accept":        "application/json",
	}

	// Build query parameters
	params := url.Values{}
	params.Add("symbols", symbols)
	if projection != "" {
		params.Add("projection", projection)
	}

	// Append query string to URL
	apiURL = fmt.Sprintf("%s?%s", apiURL, params.Encode())

	resp, err := m.httpClient.Get(ctx, apiURL, headers)
	if err != nil {
		m.logger.Error("failed to get instruments",
			"url", apiURL,
			"symbols", symbols,
			"error", err,
		)
		return nil, fmt.Errorf("failed to get instruments: %w", err)
	}

	var result types.InstrumentsResponse
	if err := m.httpClient.DecodeJSON(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode instruments response: %w", err)
	}

	m.logger.Info("successfully retrieved instruments",
		"count", len(result.Instruments),
	)

	return &result, nil
}

// InstrumentCusip retrieves instrument by CUSIP
// Endpoint: GET /trader/v1/instruments/cusip/{cusip}
func (m *Market) InstrumentCusip(ctx context.Context, cusip string) (*types.InstrumentCusipResponse, error) {
	// Create context with deadline to prevent indefinite blocking
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	apiURL := fmt.Sprintf("%s/trader/v1/instruments/cusip/%s", baseAPIURL, url.PathEscape(cusip))

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", m.tokenGetter.GetAccessToken()),
		"Accept":        "application/json",
	}

	resp, err := m.httpClient.Get(ctx, apiURL, headers)
	if err != nil {
		m.logger.Error("failed to get instrument by CUSIP",
			"url", apiURL,
			"cusip", cusip,
			"error", err,
		)
		return nil, fmt.Errorf("failed to get instrument by CUSIP: %w", err)
	}

	var result types.InstrumentCusipResponse
	if err := m.httpClient.DecodeJSON(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode instrument CUSIP response: %w", err)
	}

	m.logger.Info("successfully retrieved instrument by CUSIP",
		"cusip", cusip,
	)

	return &result, nil
}

// OptionExpirationChain retrieves option expiration chain
// Endpoint: GET /marketdata/v1/expirationchain
func (m *Market) OptionExpirationChain(ctx context.Context, symbol string, putCall string, strikePriceFrom float64, strikePriceTo float64) (*types.OptionExpirationChainResponse, error) {
	// Create context with deadline to prevent indefinite blocking
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	apiURL := fmt.Sprintf("%s/marketdata/v1/expirationchain", baseAPIURL)

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", m.tokenGetter.GetAccessToken()),
		"Accept":        "application/json",
	}

	// Build query parameters
	params := url.Values{}
	params.Add("symbol", symbol)
	if putCall != "" {
		params.Add("putCall", putCall)
	}
	if strikePriceFrom > 0 {
		params.Add("strikePriceFrom", fmt.Sprintf("%f", strikePriceFrom))
	}
	if strikePriceTo > 0 {
		params.Add("strikePriceTo", fmt.Sprintf("%f", strikePriceTo))
	}

	// Append query string to URL
	apiURL = fmt.Sprintf("%s?%s", apiURL, params.Encode())

	resp, err := m.httpClient.Get(ctx, apiURL, headers)
	if err != nil {
		m.logger.Error("failed to get option expiration chain",
			"url", apiURL,
			"symbol", symbol,
			"error", err,
		)
		return nil, fmt.Errorf("failed to get option expiration chain: %w", err)
	}

	var result types.OptionExpirationChainResponse
	if err := m.httpClient.DecodeJSON(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to decode option expiration chain response: %w", err)
	}

	m.logger.Info("successfully retrieved option expiration chain",
		"symbol", symbol,
	)

	return &result, nil
}
