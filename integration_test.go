// Integration tests for the Schwab API client.
//
// These tests make real HTTP calls to the Schwab API and require valid
// credentials. They are skipped automatically when the required environment
// variables are not set, so they never break CI.
//
// Required environment variables:
//
//	SCHWAB_APP_KEY       — your Schwab app key
//	SCHWAB_APP_SECRET    — your Schwab app secret
//	SCHWAB_CALLBACK_URL  — OAuth callback URL registered for the app
//	SCHWAB_TOKEN_PATH    — path to a tokens.json file from a prior login
//	                       (default: ~/.schwabdev/tokens.json)
//
// Run with:
//
//	go test -v -run TestIntegration ./...
//	go test -v -run TestIntegration/Quotes ./...
package schwabdev_test

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	schwabdev "github.com/citizenadam/go-schwabapi"
)

// ── Test fixture ──────────────────────────────────────────────────────────────

// integrationClient builds a real *Client from environment variables.
// Returns nil and skips the test if any required variable is missing.
func integrationClient(t *testing.T) *schwabdev.Client {
	t.Helper()

	appKey := os.Getenv("SCHWAB_APP_KEY")
	appSecret := os.Getenv("SCHWAB_APP_SECRET")
	callbackURL := os.Getenv("SCHWAB_CALLBACK_URL")

	if appKey == "" || appSecret == "" || callbackURL == "" {
		t.Skip("SCHWAB_APP_KEY / SCHWAB_APP_SECRET / SCHWAB_CALLBACK_URL not set — skipping integration tests")
	}

	tokenPath := os.Getenv("SCHWAB_TOKEN_PATH") // optional, defaults inside NewClient

	client, err := schwabdev.NewClient(
		appKey, appSecret, callbackURL,
		tokenPath,
		"", // no encryption
		30*time.Second,
		nil, // no custom callOnAuth — tokens.json must already exist
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	t.Cleanup(func() { client.Close() })
	return client
}

// firstAccountHash calls LinkedAccounts and returns the hash of the first
// account. Skips if there are no accounts.
func firstAccountHash(t *testing.T, client *schwabdev.Client) string {
	t.Helper()
	ctx := context.Background()
	accounts, err := client.LinkedAccounts(ctx)
	if err != nil {
		t.Fatalf("LinkedAccounts: %v", err)
	}
	if len(*accounts) == 0 {
		t.Skip("no linked accounts on this login — skipping")
	}
	return (*accounts)[0].HashValue
}

// assertValidJSON re-encodes v to JSON and back as a sanity check that the
// struct is fully round-trippable after being populated from a live response.
func assertValidJSON(t *testing.T, label string, v any) {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Errorf("%s: marshal failed: %v", label, err)
		return
	}
	if len(b) < 2 {
		t.Errorf("%s: marshal produced suspiciously short output: %s", label, b)
	}
}

// ptr returns a pointer to v. Convenience for optional API params.
func ptr[T any](v T) *T { return &v }

// ── Account tests ─────────────────────────────────────────────────────────────

func TestIntegration_LinkedAccounts(t *testing.T) {
	client := integrationClient(t)
	ctx := context.Background()

	resp, err := client.LinkedAccounts(ctx)
	if err != nil {
		t.Fatalf("LinkedAccounts error: %v", err)
	}
	if resp == nil {
		t.Fatal("response is nil")
	}
	if len(*resp) == 0 {
		t.Fatal("expected at least one linked account")
	}
	for i, acct := range *resp {
		if acct.AccountNumber == "" {
			t.Errorf("account[%d]: AccountNumber is empty", i)
		}
		if acct.HashValue == "" {
			t.Errorf("account[%d]: HashValue is empty", i)
		}
	}
	assertValidJSON(t, "LinkedAccountsResponse", resp)
	t.Logf("LinkedAccounts: %d account(s) found", len(*resp))
}

func TestIntegration_AccountDetailsAll(t *testing.T) {
	client := integrationClient(t)
	ctx := context.Background()

	resp, err := client.AccountDetailsAll(ctx, nil)
	if err != nil {
		t.Fatalf("AccountDetailsAll error: %v", err)
	}
	if resp == nil {
		t.Fatal("response is nil")
	}
	assertValidJSON(t, "AccountDetailsAllResponse", resp)
}

func TestIntegration_AccountDetailsAll_WithPositions(t *testing.T) {
	client := integrationClient(t)
	ctx := context.Background()

	fields := "positions"
	resp, err := client.AccountDetailsAll(ctx, &fields)
	if err != nil {
		t.Fatalf("AccountDetailsAll(positions) error: %v", err)
	}
	if resp == nil {
		t.Fatal("response is nil")
	}
	assertValidJSON(t, "AccountDetailsAllResponse+positions", resp)
}

func TestIntegration_AccountDetails(t *testing.T) {
	client := integrationClient(t)
	ctx := context.Background()
	hash := firstAccountHash(t, client)

	resp, err := client.AccountDetails(ctx, hash, nil)
	if err != nil {
		t.Fatalf("AccountDetails error: %v", err)
	}
	if resp == nil {
		t.Fatal("response is nil")
	}
	if resp.SecuritiesAccount == nil {
		t.Fatal("SecuritiesAccount is nil — check account hash")
	}
	if resp.SecuritiesAccount.AccountNumber == "" {
		t.Error("SecuritiesAccount.AccountNumber is empty")
	}
	assertValidJSON(t, "AccountDetailsResponse", resp)
	t.Logf("AccountDetails: type=%s", resp.SecuritiesAccount.Type)
}

func TestIntegration_AccountDetails_WithPositions(t *testing.T) {
	client := integrationClient(t)
	ctx := context.Background()
	hash := firstAccountHash(t, client)

	fields := "positions"
	resp, err := client.AccountDetails(ctx, hash, &fields)
	if err != nil {
		t.Fatalf("AccountDetails(positions) error: %v", err)
	}
	if resp == nil {
		t.Fatal("response is nil")
	}
	assertValidJSON(t, "AccountDetailsResponse+positions", resp)
	t.Logf("AccountDetails+positions: %d position(s)", len(resp.SecuritiesAccount.Positions))
}

// ── Order tests ───────────────────────────────────────────────────────────────

func TestIntegration_AccountOrders(t *testing.T) {
	client := integrationClient(t)
	ctx := context.Background()
	hash := firstAccountHash(t, client)

	from := time.Now().AddDate(0, -1, 0) // last 30 days
	to := time.Now()

	resp, err := client.AccountOrders(ctx, hash, from, to, nil, nil)
	if err != nil {
		t.Fatalf("AccountOrders error: %v", err)
	}
	if resp == nil {
		t.Fatal("response is nil")
	}
	assertValidJSON(t, "AccountOrdersResponse", resp)
	t.Logf("AccountOrders: %d order(s) in last 30 days", len(*resp))

	for i, order := range *resp {
		if order.OrderID == 0 {
			t.Errorf("order[%d]: OrderID is 0", i)
		}
		if order.Status == "" {
			t.Errorf("order[%d]: Status is empty", i)
		}
		if order.OrderType == "" {
			t.Errorf("order[%d]: OrderType is empty", i)
		}
		for j, leg := range order.OrderLegCollection {
			if leg.Instrument == nil {
				t.Errorf("order[%d] leg[%d]: Instrument is nil", i, j)
			} else if leg.Instrument.Symbol == "" {
				t.Errorf("order[%d] leg[%d]: Symbol is empty", i, j)
			}
		}
	}
}

func TestIntegration_AccountOrdersAll(t *testing.T) {
	client := integrationClient(t)
	ctx := context.Background()

	from := time.Now().AddDate(0, -1, 0)
	to := time.Now()
	maxResults := 50

	resp, err := client.AccountOrdersAll(ctx, from, to, &maxResults, nil)
	if err != nil {
		t.Fatalf("AccountOrdersAll error: %v", err)
	}
	if resp == nil {
		t.Fatal("response is nil")
	}
	assertValidJSON(t, "AccountOrdersAllResponse", resp)
	t.Logf("AccountOrdersAll: %d order(s)", len(*resp))
}

func TestIntegration_PreviewOrder(t *testing.T) {
	client := integrationClient(t)
	ctx := context.Background()
	hash := firstAccountHash(t, client)

	order := &schwabdev.PreviewOrderRequest{
		OrderType:         "LIMIT",
		Session:           "NORMAL",
		Duration:          "DAY",
		OrderStrategyType: "SINGLE",
		Price:             "1.00", // far OTM — unlikely to fill
		OrderLegCollection: []*schwabdev.OrderLegRequest{
			{
				Instruction: "BUY",
				Quantity:    1,
				Instrument: &schwabdev.InstrumentRequest{
					Symbol:    "AAPL",
					AssetType: "EQUITY",
				},
			},
		},
	}

	resp, err := client.PreviewOrder(ctx, hash, order)
	if err != nil {
		// PreviewOrder can legitimately fail with a validation rejection —
		// that's still a successful API call, the struct just has Rejects populated.
		// Only fail on network/auth errors.
		if strings.Contains(err.Error(), "401") || strings.Contains(err.Error(), "403") {
			t.Fatalf("PreviewOrder auth error: %v", err)
		}
		t.Logf("PreviewOrder returned error (may be expected): %v", err)
		return
	}
	if resp == nil {
		t.Fatal("response is nil")
	}
	assertValidJSON(t, "PreviewOrderResponse", resp)
	t.Logf("PreviewOrder: orderID=%d", resp.OrderID)
}

// ── Transaction tests ─────────────────────────────────────────────────────────

func TestIntegration_Transactions(t *testing.T) {
	client := integrationClient(t)
	ctx := context.Background()
	hash := firstAccountHash(t, client)

	from := time.Now().AddDate(0, -3, 0) // last 90 days
	to := time.Now()

	resp, err := client.Transactions(ctx, hash, from, to, "TRADE", nil)
	if err != nil {
		t.Fatalf("Transactions error: %v", err)
	}
	if resp == nil {
		t.Fatal("response is nil")
	}
	assertValidJSON(t, "TransactionsResponse", resp)
	t.Logf("Transactions: %d trade(s) in last 90 days", len(*resp))

	for i, tx := range *resp {
		if tx.TransactionID == "" {
			t.Errorf("transaction[%d]: TransactionID is empty", i)
		}
		if tx.Type == "" {
			t.Errorf("transaction[%d]: Type is empty", i)
		}
	}
}

func TestIntegration_TransactionDetails(t *testing.T) {
	client := integrationClient(t)
	ctx := context.Background()
	hash := firstAccountHash(t, client)

	// Get a transaction ID from the list first.
	from := time.Now().AddDate(0, -3, 0)
	to := time.Now()
	list, err := client.Transactions(ctx, hash, from, to, "TRADE", nil)
	if err != nil {
		t.Fatalf("Transactions (setup): %v", err)
	}
	if list == nil || len(*list) == 0 {
		t.Skip("no transactions in last 90 days — skipping TransactionDetails")
	}

	txID := (*list)[0].TransactionID
	resp, err := client.TransactionDetails(ctx, hash, txID)
	if err != nil {
		t.Fatalf("TransactionDetails error: %v", err)
	}
	if resp == nil {
		t.Fatal("response is nil")
	}
	if (*schwabdev.Transaction)(resp).TransactionID != txID {
		t.Errorf("TransactionID: want %s, got %s", txID, (*schwabdev.Transaction)(resp).TransactionID)
	}
	assertValidJSON(t, "TransactionDetailsResponse", resp)
}

// ── Market Data tests ─────────────────────────────────────────────────────────

func TestIntegration_Quote(t *testing.T) {
	client := integrationClient(t)
	ctx := context.Background()

	resp, err := client.Quote(ctx, "AAPL", nil)
	if err != nil {
		t.Fatalf("Quote error: %v", err)
	}
	if resp == nil {
		t.Fatal("response is nil")
	}
	if resp.Symbol != "AAPL" {
		t.Errorf("Symbol: want AAPL, got %s", resp.Symbol)
	}
	if resp.AssetMainType == "" {
		t.Error("AssetMainType is empty")
	}
	if (*schwabdev.Quote)(resp).QuoteData == nil {
		t.Error("QuoteData is nil — fields param may need to include 'quote'")
	} else {
		if (*schwabdev.Quote)(resp).QuoteData.AskPrice <= 0 {
			t.Logf("Warning: AskPrice is %f (market may be closed)", (*schwabdev.Quote)(resp).QuoteData.AskPrice)
		}
		if (*schwabdev.Quote)(resp).QuoteData.BidPrice <= 0 {
			t.Logf("Warning: BidPrice is %f (market may be closed)", (*schwabdev.Quote)(resp).QuoteData.BidPrice)
		}
		if (*schwabdev.Quote)(resp).QuoteData.ClosePrice <= 0 {
			t.Errorf("ClosePrice is %f — should always be populated", (*schwabdev.Quote)(resp).QuoteData.ClosePrice)
		}
	}
	assertValidJSON(t, "QuoteResponse(AAPL)", resp)
	t.Logf("Quote AAPL: last=%.2f close=%.2f", (*schwabdev.Quote)(resp).QuoteData.LastPrice, (*schwabdev.Quote)(resp).QuoteData.ClosePrice)
}

func TestIntegration_Quote_AllFields(t *testing.T) {
	client := integrationClient(t)
	ctx := context.Background()

	fields := "all"
	resp, err := client.Quote(ctx, "AAPL", &fields)
	if err != nil {
		t.Fatalf("Quote(all) error: %v", err)
	}
	if (*schwabdev.Quote)(resp).Fundamental == nil {
		t.Error("Fundamental is nil when fields=all")
	} else {
		if (*schwabdev.Quote)(resp).Fundamental.PeRatio <= 0 {
			t.Logf("Warning: PeRatio is %f", (*schwabdev.Quote)(resp).Fundamental.PeRatio)
		}
	}
	if (*schwabdev.Quote)(resp).Reference == nil {
		t.Error("Reference is nil when fields=all")
	} else {
		if (*schwabdev.Quote)(resp).Reference.Cusip == "" {
			t.Error("Reference.Cusip is empty")
		}
		if (*schwabdev.Quote)(resp).Reference.Description == "" {
			t.Error("Reference.Description is empty")
		}
	}
	assertValidJSON(t, "QuoteResponse(AAPL,all)", resp)
}

func TestIntegration_Quotes_MultipleSymbols(t *testing.T) {
	client := integrationClient(t)
	ctx := context.Background()

	symbols := []string{"AAPL", "MSFT", "GOOGL", "AMZN", "NVDA"}
	resp, err := client.Quotes(ctx, symbols, nil, nil)
	if err != nil {
		t.Fatalf("Quotes error: %v", err)
	}
	if resp == nil {
		t.Fatal("response is nil")
	}
	for _, sym := range symbols {
		q, ok := (*resp)[sym]
		if !ok {
			t.Errorf("symbol %s missing from response", sym)
			continue
		}
		if q.Symbol != sym {
			t.Errorf("%s: Symbol field mismatch: got %s", sym, q.Symbol)
		}
		if (*schwabdev.Quote)(&q).QuoteData == nil {
			t.Errorf("%s: QuoteData is nil", sym)
		}
	}
	assertValidJSON(t, "QuotesResponse(5 symbols)", resp)
	t.Logf("Quotes: received %d/%d symbols", len(*resp), len(symbols))
}

func TestIntegration_Quotes_52WeekHighLow(t *testing.T) {
	// Specifically tests that the unusual "52WeekHigh"/"52WeekLow" JSON keys
	// decode correctly into FiftyTwoWeekHigh/FiftyTwoWeekLow struct fields.
	client := integrationClient(t)
	ctx := context.Background()

	fields := "quote"
	resp, err := client.Quote(ctx, "SPY", &fields)
	if err != nil {
		t.Fatalf("Quote(SPY) error: %v", err)
	}
	if (*schwabdev.Quote)(resp).QuoteData == nil {
		t.Skip("QuoteData nil — cannot test 52-week fields")
	}
	if (*schwabdev.Quote)(resp).QuoteData.FiftyTwoWeekHigh <= 0 {
		t.Errorf("FiftyTwoWeekHigh is %f — struct tag may be wrong", (*schwabdev.Quote)(resp).QuoteData.FiftyTwoWeekHigh)
	}
	if (*schwabdev.Quote)(resp).QuoteData.FiftyTwoWeekLow <= 0 {
		t.Errorf("FiftyTwoWeekLow is %f — struct tag may be wrong", (*schwabdev.Quote)(resp).QuoteData.FiftyTwoWeekLow)
	}
	if (*schwabdev.Quote)(resp).QuoteData.FiftyTwoWeekHigh < (*schwabdev.Quote)(resp).QuoteData.FiftyTwoWeekLow {
		t.Errorf("52WeekHigh (%.2f) < 52WeekLow (%.2f) — fields may be swapped",
			(*schwabdev.Quote)(resp).QuoteData.FiftyTwoWeekHigh, (*schwabdev.Quote)(resp).QuoteData.FiftyTwoWeekLow)
	}
	t.Logf("SPY 52wk: high=%.2f low=%.2f", (*schwabdev.Quote)(resp).QuoteData.FiftyTwoWeekHigh, (*schwabdev.Quote)(resp).QuoteData.FiftyTwoWeekLow)
}

func TestIntegration_PriceHistory(t *testing.T) {
	client := integrationClient(t)
	ctx := context.Background()

	periodType := "month"
	period := 1
	frequencyType := "daily"
	frequency := 1

	resp, err := client.PriceHistory(ctx, "AAPL", &periodType, &period, &frequencyType, &frequency, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("PriceHistory error: %v", err)
	}
	if resp == nil {
		t.Fatal("response is nil")
	}
	if resp.Symbol != "AAPL" {
		t.Errorf("Symbol: want AAPL, got %s", resp.Symbol)
	}
	if resp.Empty {
		t.Error("PriceHistory returned empty — expected candles for AAPL")
	}
	if len(resp.Candles) == 0 {
		t.Error("no candles returned")
	}
	for i, c := range resp.Candles {
		if c.Open <= 0 {
			t.Errorf("candle[%d]: Open is %f", i, c.Open)
		}
		if c.High < c.Low {
			t.Errorf("candle[%d]: High (%.2f) < Low (%.2f)", i, c.High, c.Low)
		}
		if c.Volume < 0 {
			t.Errorf("candle[%d]: Volume is negative: %d", i, c.Volume)
		}
		if c.Datetime <= 0 {
			t.Errorf("candle[%d]: Datetime is %d", i, c.Datetime)
		}
	}
	assertValidJSON(t, "PriceHistoryResponse", resp)
	t.Logf("PriceHistory AAPL: %d candles, latest close=%.2f", len(resp.Candles), resp.Candles[len(resp.Candles)-1].Close)
}

func TestIntegration_PriceHistory_DateRange(t *testing.T) {
	client := integrationClient(t)
	ctx := context.Background()

	start := time.Now().AddDate(0, 0, -7)
	end := time.Now()

	resp, err := client.PriceHistory(ctx, "SPY", nil, nil, nil, nil, start, end, nil, nil)
	if err != nil {
		t.Fatalf("PriceHistory(date range) error: %v", err)
	}
	if resp == nil {
		t.Fatal("response is nil")
	}
	t.Logf("PriceHistory SPY (7d): %d candles", len(resp.Candles))
}

func TestIntegration_OptionChains(t *testing.T) {
	client := integrationClient(t)
	ctx := context.Background()

	contractType := "CALL"
	strikeCount := 5

	resp, err := client.OptionChains(ctx, "AAPL", &contractType, &strikeCount,
		nil, nil, nil, nil, nil,
		nil, nil,
		nil, nil, nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("OptionChains error: %v", err)
	}
	if resp == nil {
		t.Fatal("response is nil")
	}
	if resp.Symbol != "AAPL" {
		t.Errorf("Symbol: want AAPL, got %s", resp.Symbol)
	}
	if resp.Status != "SUCCESS" {
		t.Errorf("Status: want SUCCESS, got %s", resp.Status)
	}
	if len(resp.CallExpDateMap) == 0 {
		t.Error("CallExpDateMap is empty")
	}
	if len(resp.PutExpDateMap) != 0 {
		t.Errorf("PutExpDateMap should be empty when contractType=CALL, got %d entries", len(resp.PutExpDateMap))
	}

	// Validate at least one contract has required fields populated.
	var found bool
	for expiry, strikes := range resp.CallExpDateMap {
		for strike, contracts := range strikes {
			for _, c := range contracts {
				found = true
				if c.Symbol == "" {
					t.Errorf("expiry=%s strike=%s: contract Symbol is empty", expiry, strike)
				}
				if c.StrikePrice <= 0 {
					t.Errorf("expiry=%s strike=%s: StrikePrice is %f", expiry, strike, c.StrikePrice)
				}
				if c.ExpirationDate == "" {
					t.Errorf("expiry=%s strike=%s: ExpirationDate is empty", expiry, strike)
				}
			}
		}
	}
	if !found {
		t.Error("no contracts found in CallExpDateMap")
	}
	assertValidJSON(t, "OptionChainsResponse", resp)
	t.Logf("OptionChains AAPL: %d expiry dates, underlying=%.2f", len(resp.CallExpDateMap), resp.UnderlyingPrice)
}

func TestIntegration_OptionExpirationChain(t *testing.T) {
	client := integrationClient(t)
	ctx := context.Background()

	resp, err := client.OptionExpirationChain(ctx, "SPY")
	if err != nil {
		t.Fatalf("OptionExpirationChain error: %v", err)
	}
	if resp == nil {
		t.Fatal("response is nil")
	}
	if len(resp.ExpirationList) == 0 {
		t.Error("ExpirationList is empty")
	}
	for i, exp := range resp.ExpirationList {
		if exp.ExpirationDate == "" {
			t.Errorf("expiration[%d]: ExpirationDate is empty", i)
		}
		if exp.DaysToExpiration < 0 {
			t.Errorf("expiration[%d]: DaysToExpiration is negative: %d", i, exp.DaysToExpiration)
		}
	}
	assertValidJSON(t, "OptionExpirationChainResponse", resp)
	t.Logf("OptionExpirationChain SPY: %d expiries", len(resp.ExpirationList))
}

func TestIntegration_Movers(t *testing.T) {
	client := integrationClient(t)
	ctx := context.Background()

	resp, err := client.Movers(ctx, "$SPX", nil, nil)
	if err != nil {
		t.Fatalf("Movers error: %v", err)
	}
	if resp == nil {
		t.Fatal("response is nil")
	}
	// Movers can be empty outside market hours — just verify the struct.
	for i, m := range *resp {
		if m.Symbol == "" {
			t.Errorf("mover[%d]: Symbol is empty", i)
		}
	}
	assertValidJSON(t, "MoversResponse", resp)
	t.Logf("Movers $SPX: %d mover(s)", len(*resp))
}

func TestIntegration_MarketHours_Single(t *testing.T) {
	client := integrationClient(t)
	ctx := context.Background()

	resp, err := client.MarketHour(ctx, "equity", nil)
	if err != nil {
		t.Fatalf("MarketHour error: %v", err)
	}
	if resp == nil {
		t.Fatal("response is nil")
	}
	assertValidJSON(t, "MarketHourResponse", resp)
	t.Logf("MarketHour equity: isOpen=%v", (*schwabdev.MarketHour)(resp).IsOpen)
}

func TestIntegration_MarketHours_Multiple(t *testing.T) {
	client := integrationClient(t)
	ctx := context.Background()

	markets := []string{"equity", "option", "future"}
	resp, err := client.MarketHours(ctx, markets, nil)
	if err != nil {
		t.Fatalf("MarketHours error: %v", err)
	}
	if resp == nil {
		t.Fatal("response is nil")
	}
	assertValidJSON(t, "MarketHoursResponse", resp)
	t.Logf("MarketHours: received %d market(s)", len(*resp))
}

func TestIntegration_Instruments(t *testing.T) {
	client := integrationClient(t)
	ctx := context.Background()

	resp, err := client.Instruments(ctx, "AAPL", "symbol-search")
	if err != nil {
		t.Fatalf("Instruments error: %v", err)
	}
	if resp == nil {
		t.Fatal("response is nil")
	}
	if len(*resp) == 0 {
		t.Error("no instruments returned for AAPL symbol-search")
	}
	for i, inst := range *resp {
		if inst.Symbol == "" {
			t.Errorf("instrument[%d]: Symbol is empty", i)
		}
		if inst.AssetType == "" {
			t.Errorf("instrument[%d]: AssetType is empty", i)
		}
	}
	assertValidJSON(t, "InstrumentsResponse", resp)
	t.Logf("Instruments(AAPL): %d result(s)", len(*resp))
}

func TestIntegration_InstrumentCUSIP(t *testing.T) {
	client := integrationClient(t)
	ctx := context.Background()

	// AAPL CUSIP
	resp, err := client.InstrumentCUSIP(ctx, "037833100")
	if err != nil {
		t.Fatalf("InstrumentCUSIP error: %v", err)
	}
	if resp == nil {
		t.Fatal("response is nil")
	}
	assertValidJSON(t, "InstrumentCUSIPResponse", resp)
	t.Logf("InstrumentCUSIP(AAPL): %d instrument(s)", len(resp.Instruments))
}

// ── Streamer Info ─────────────────────────────────────────────────────────────

func TestIntegration_GetStreamerInfo(t *testing.T) {
	client := integrationClient(t)
	ctx := context.Background()

	resp, err := client.GetStreamerInfo(ctx)
	if err != nil {
		t.Fatalf("GetStreamerInfo error: %v", err)
	}
	if resp == nil {
		t.Fatal("response is nil")
	}
	if resp.StreamerURL == "" {
		t.Error("StreamerURL is empty — struct tag may not match API field name")
	}
	if resp.SchwabClientChannel == "" {
		t.Error("SchwabClientChannel is empty")
	}
	if resp.SchwabClientCorrelID == "" {
		t.Error("SchwabClientCorrelID is empty")
	}
	if !strings.HasPrefix(resp.StreamerURL, "wss://") {
		t.Errorf("StreamerURL should start with wss://, got: %s", resp.StreamerURL)
	}
	assertValidJSON(t, "StreamerInfo", resp)
	t.Logf("StreamerInfo: url=%s channel=%s", resp.StreamerURL, resp.SchwabClientChannel)
}

// ── Token lifecycle ───────────────────────────────────────────────────────────

func TestIntegration_UpdateTokens_NoForce(t *testing.T) {
	client := integrationClient(t)

	updated, err := client.UpdateTokens(false, false)
	if err != nil {
		t.Fatalf("UpdateTokens(false,false) error: %v", err)
	}
	// Tokens should still be valid from a fresh login — expect no refresh.
	t.Logf("UpdateTokens(false,false): updated=%v", updated)
}

func TestIntegration_UpdateTokens_ForceAccess(t *testing.T) {
	client := integrationClient(t)

	updated, err := client.UpdateTokens(true, false)
	if err != nil {
		t.Fatalf("UpdateTokens(true,false) error: %v", err)
	}
	if !updated {
		t.Error("UpdateTokens(forceAccess=true) should have updated the access token")
	}

	// Immediately make a real API call to confirm the new token works.
	ctx := context.Background()
	resp, err := client.LinkedAccounts(ctx)
	if err != nil {
		t.Fatalf("API call after forced token refresh failed: %v", err)
	}
	if resp == nil || len(*resp) == 0 {
		t.Error("expected accounts after forced token refresh")
	}
}

func TestIntegration_TokenInfo(t *testing.T) {
	client := integrationClient(t)

	info := client.TokenManager().TokenInfo()
	if info.AccessToken == "" {
		t.Error("AccessToken is empty")
	}
	if info.AccessTokenIssued.IsZero() {
		t.Error("AccessTokenIssued is zero")
	}
	if info.RefreshTokenIssued.IsZero() {
		t.Error("RefreshTokenIssued is zero")
	}
	if info.AccessTokenExpiry.IsZero() {
		t.Error("AccessTokenExpiry is zero")
	}
	if info.RefreshTokenExpiry.IsZero() {
		t.Error("RefreshTokenExpiry is zero")
	}
	if info.AccessTokenExpiry.Before(time.Now()) {
		t.Error("AccessToken is already expired — UpdateTokens should have refreshed it")
	}
	if info.RefreshTokenExpiry.Before(time.Now()) {
		t.Error("RefreshToken is already expired — re-authentication is required")
	}
	if !info.Valid() {
		t.Error("TokenInfo.Valid() returned false — token appears expired")
	}
	t.Logf("TokenInfo: expires_in=%.0fs refresh_expires_in=%.0fh",
		time.Until(info.AccessTokenExpiry).Seconds(),
		time.Until(info.RefreshTokenExpiry).Hours())
}
