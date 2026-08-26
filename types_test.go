package schwabdev_test

import (
	"encoding/json"
	"testing"

	schwabdev "github.com/citizenadam/go-schwabapi"
)

// ── helpers ───────────────────────────────────────────────────────────────────

// roundtrip marshals v to JSON then unmarshals into a new value of the same
// type and returns it. It fails the test on any error.
func roundtrip[T any](t *testing.T, v T) T {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got T
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	return got
}

// mustUnmarshal decodes raw JSON into T and fails the test on error.
func mustUnmarshal[T any](t *testing.T, raw string) T {
	t.Helper()
	var v T
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		t.Fatalf("unmarshal %T: %v", v, err)
	}
	return v
}

// ── Accounts ──────────────────────────────────────────────────────────────────

func TestLinkedAccountsResponse_RoundTrip(t *testing.T) {
	input := schwabdev.LinkedAccountsResponse{
		{AccountNumber: "123456789", HashValue: "abc123hash"},
		{AccountNumber: "987654321", HashValue: "xyz789hash"},
	}
	got := roundtrip(t, input)
	if len(got) != 2 {
		t.Fatalf("want 2 accounts, got %d", len(got))
	}
	if got[0].AccountNumber != "123456789" {
		t.Errorf("AccountNumber: want 123456789, got %s", got[0].AccountNumber)
	}
	if got[0].HashValue != "abc123hash" {
		t.Errorf("HashValue: want abc123hash, got %s", got[0].HashValue)
	}
}

func TestLinkedAccountsResponse_UnmarshalFromAPI(t *testing.T) {
	raw := `[
		{"accountNumber": "111222333", "hashValue": "hash111"},
		{"accountNumber": "444555666", "hashValue": "hash444"}
	]`
	got := mustUnmarshal[schwabdev.LinkedAccountsResponse](t, raw)
	if len(got) != 2 {
		t.Fatalf("want 2, got %d", len(got))
	}
	if got[1].HashValue != "hash444" {
		t.Errorf("want hash444, got %s", got[1].HashValue)
	}
}

func TestAccountDetailsAllResponse_RoundTrip(t *testing.T) {
	input := schwabdev.AccountDetailsAllResponse{
		SecuritiesAccount: &schwabdev.SecuritiesAccount{
			Type:          "MARGIN",
			AccountNumber: "ACC001",
			RoundTrips:    2,
			IsDayTrader:   false,
			Positions: []*schwabdev.Position{
				{
					LongQuantity: 100,
					AveragePrice: 150.25,
					MarketValue:  17500.00,
					Instrument: &schwabdev.Instrument{
						Symbol:       "AAPL",
						AssetType:    "EQUITY",
						Cusip:        "037833100",
						InstrumentID: 1234567,
					},
				},
			},
		},
		AggregatedBalance: &schwabdev.AggregatedBalance{
			CurrentLiquidationValue: 33500.00,
			LiquidationValue:        34000.00,
		},
	}
	got := roundtrip(t, input)
	if got.SecuritiesAccount == nil {
		t.Fatal("SecuritiesAccount is nil after roundtrip")
	}
	if got.SecuritiesAccount.Type != "MARGIN" {
		t.Errorf("Type: want MARGIN, got %s", got.SecuritiesAccount.Type)
	}
	if len(got.SecuritiesAccount.Positions) != 1 {
		t.Fatalf("want 1 position, got %d", len(got.SecuritiesAccount.Positions))
	}
	if got.SecuritiesAccount.Positions[0].Instrument == nil {
		t.Fatal("Position instrument is nil")
	}
	if got.SecuritiesAccount.Positions[0].Instrument.Symbol != "AAPL" {
		t.Errorf("Position symbol: want AAPL, got %s", got.SecuritiesAccount.Positions[0].Instrument.Symbol)
	}
	if got.AggregatedBalance == nil {
		t.Fatal("AggregatedBalance is nil after roundtrip")
	}
}

func TestAccountDetailsAllResponse_UnmarshalFromAPI(t *testing.T) {
	raw := `{
		"securitiesAccount": {
			"type": "CASH",
			"accountNumber": "ACCT123",
			"roundTrips": 0,
			"isDayTrader": false,
			"isClosingOnlyRestricted": false,
			"pfcbFlag": false,
			"currentBalances": {
				"cashBalance": 5000.00,
				"buyingPower": 5000.00,
				"equity": 5000.00,
				"liquidationValue": 5000.00
			}
		}
	}`
	got := mustUnmarshal[schwabdev.AccountDetailsAllResponse](t, raw)
	if got.SecuritiesAccount == nil {
		t.Fatal("SecuritiesAccount unexpectedly nil")
	}
	if got.SecuritiesAccount.AccountNumber != "ACCT123" {
		t.Errorf("want ACCT123, got %s", got.SecuritiesAccount.AccountNumber)
	}
	if len(got.SecuritiesAccount.CurrentBalances) == 0 {
		t.Fatal("CurrentBalances unexpectedly empty")
	}
}

func TestAccountDetailsAllResponse_NilOptionals(t *testing.T) {
	// Ensure optional pointer fields decode as nil when absent.
	raw := `{"securitiesAccount": {"type": "CASH", "accountNumber": "X", "roundTrips": 0}}`
	got := mustUnmarshal[schwabdev.AccountDetailsAllResponse](t, raw)
	if len(got.SecuritiesAccount.InitialBalances) != 0 {
		t.Error("InitialBalances should be empty when absent")
	}
	if len(got.SecuritiesAccount.Positions) != 0 {
		t.Error("Positions should be empty when absent")
	}
	if got.AggregatedBalance != nil {
		t.Error("AggregatedBalance should be nil when absent")
	}
}

// ── Orders ────────────────────────────────────────────────────────────────────

func TestOrder_RoundTrip(t *testing.T) {
	tag := "my-algo"
	input := schwabdev.Order{
		Session:                  "NORMAL",
		Duration:                 "DAY",
		OrderType:                "LIMIT",
		ComplexOrderStrategyType: "NONE",
		Quantity:                 10,
		FilledQuantity:           10,
		RemainingQuantity:        0,
		RequestedDestination:     "AUTO",
		Price:                    155.50,
		OrderStrategyType:        "SINGLE",
		OrderID:                  9876543210,
		Cancelable:               false,
		Editable:                 false,
		Status:                   "FILLED",
		EnteredTime:              "2024-01-15T10:30:00+0000",
		Tag:                      &tag,
		AccountNumber:            111222333,
		OrderLegCollection: []*schwabdev.OrderLegCollection{
			{
				OrderLegType: "EQUITY",
				LegID:        1,
				Instruction:  "BUY",
				Quantity:     10,
				Instrument: &schwabdev.Instrument{
					AssetType:    "EQUITY",
					Symbol:       "MSFT",
					Cusip:        "594918104",
					InstrumentID: 7654321,
				},
			},
		},
		OrderActivityCollection: []*schwabdev.OrderActivity{
			{
				ActivityType:           "EXECUTION",
				ExecutionType:          "FILL",
				Quantity:               10,
				OrderRemainingQuantity: 0,
				ExecutionLegs: []*schwabdev.ExecutionLeg{
					{
						LegID:    1,
						Quantity: 10,
						Price:    155.48,
						Time:     "2024-01-15T10:30:01+0000",
					},
				},
			},
		},
	}
	got := roundtrip(t, input)
	if got.OrderID != 9876543210 {
		t.Errorf("OrderID: want 9876543210, got %d", got.OrderID)
	}
	if got.Tag == nil || *got.Tag != "my-algo" {
		t.Errorf("Tag: want 'my-algo', got %v", got.Tag)
	}
	if len(got.OrderLegCollection) != 1 {
		t.Fatalf("want 1 order leg, got %d", len(got.OrderLegCollection))
	}
	if got.OrderLegCollection[0].Instrument.Symbol != "MSFT" {
		t.Errorf("Instrument symbol: want MSFT, got %s", got.OrderLegCollection[0].Instrument.Symbol)
	}
	if len(got.OrderActivityCollection) != 1 {
		t.Fatalf("want 1 activity, got %d", len(got.OrderActivityCollection))
	}
}

func TestOrder_UnmarshalFromAPI(t *testing.T) {
	raw := `{
		"session": "NORMAL",
		"duration": "DAY",
		"orderType": "MARKET",
		"complexOrderStrategyType": "NONE",
		"quantity": 5.0,
		"filledQuantity": 5.0,
		"remainingQuantity": 0.0,
		"requestedDestination": "AUTO",
		"destinationLinkName": "NITE",
		"price": 0.0,
		"orderLegCollection": [
			{
				"orderLegType": "EQUITY",
				"legId": 1,
				"instruction": "SELL",
				"positionEffect": "CLOSING",
				"quantity": 5.0,
				"instrument": {
					"assetType": "EQUITY",
					"cusip": "037833100",
					"symbol": "AAPL",
					"instrumentId": 1234567
				}
			}
		],
		"orderStrategyType": "SINGLE",
		"orderId": 1122334455,
		"cancelable": false,
		"editable": false,
		"status": "FILLED",
		"enteredTime": "2024-06-01T14:00:00+0000",
		"accountNumber": 999888777
	}`
	got := mustUnmarshal[schwabdev.Order](t, raw)
	if got.Status != "FILLED" {
		t.Errorf("Status: want FILLED, got %s", got.Status)
	}
	if got.OrderID != 1122334455 {
		t.Errorf("OrderID: want 1122334455, got %d", got.OrderID)
	}
	if len(got.OrderLegCollection) != 1 {
		t.Fatalf("want 1 leg, got %d", len(got.OrderLegCollection))
	}
	if got.OrderLegCollection[0].Instrument == nil {
		t.Fatal("Instrument is nil")
	}
	if got.OrderLegCollection[0].Instrument.Symbol != "AAPL" {
		t.Errorf("Symbol: want AAPL, got %s", got.OrderLegCollection[0].Instrument.Symbol)
	}
	if got.CancelTime != nil {
		t.Error("CancelTime should be nil when absent")
	}
}

func TestAccountOrdersResponse_RoundTrip(t *testing.T) {
	input := schwabdev.AccountOrdersResponse{
		{OrderID: 1, Status: "WORKING", OrderType: "LIMIT"},
		{OrderID: 2, Status: "FILLED", OrderType: "MARKET"},
	}
	got := roundtrip(t, input)
	if len(got) != 2 {
		t.Fatalf("want 2 orders, got %d", len(got))
	}
	if got[1].Status != "FILLED" {
		t.Errorf("want FILLED, got %s", got[1].Status)
	}
}

// ── Transactions ──────────────────────────────────────────────────────────────

func TestTransaction_RoundTrip(t *testing.T) {
	input := schwabdev.Transaction{
		ActivityID:    12345,
		Type:          "TRADE",
		Description:   "Bought 5 GOOGL @ 175.30",
		AccountNumber: "ACC123",
		NetAmount:     876.50,
	}
	got := roundtrip(t, input)
	if got.ActivityID != 12345 {
		t.Errorf("ActivityID: want 12345, got %d", got.ActivityID)
	}
	if got.NetAmount != 876.50 {
		t.Errorf("NetAmount: want 876.50, got %f", got.NetAmount)
	}
}

func TestTransactionsResponse_UnmarshalFromAPI(t *testing.T) {
	raw := `[
		{"activityId": 1, "type": "TRADE", "description": "NVDA trade", "netAmount": -5001.50},
		{"activityId": 2, "type": "DIVIDEND_OR_INTEREST", "description": "MSFT div", "netAmount": 25.00}
	]`
	got := mustUnmarshal[schwabdev.TransactionsResponse](t, raw)
	if len(got) != 2 {
		t.Fatalf("want 2, got %d", len(got))
	}
	if got[0].Type != "TRADE" {
		t.Errorf("want TRADE, got %s", got[0].Type)
	}
	if got[1].NetAmount != 25.00 {
		t.Errorf("want 25.00, got %f", got[1].NetAmount)
	}
}

// ── Quotes ────────────────────────────────────────────────────────────────────

func TestQuote_RoundTrip(t *testing.T) {
	input := schwabdev.Quote{
		AssetMainType: "EQUITY",
		AssetSubType:  "COE",
		Symbol:        "AAPL",
		RealTime:      true,
		SSID:          1234567890,
		QuoteData: &schwabdev.QuoteData{
			AskPrice:         182.50,
			BidPrice:         182.48,
			LastPrice:        182.49,
			OpenPrice:        181.00,
			ClosePrice:       180.75,
			HighPrice:        183.20,
			LowPrice:         180.50,
			TotalVolume:      45678901,
			NetChange:        1.74,
			NetPercentChange: 0.96,
			Mark:             182.49,
			SecurityStatus:   "Normal",
			Volatility:       25.5,
		},
		Fundamental: &schwabdev.Fundamental{
			PeRatio:         28.5,
			Eps:             6.43,
			DivYield:        0.54,
			DivAmount:       0.96,
			Avg10DaysVolume: 55000000,
			FundStrategy:    "A",
		},
		Reference: &schwabdev.Reference{
			Cusip:        "037833100",
			Description:  "Apple Inc",
			Exchange:     "Q",
			ExchangeName: "NASDAQ",
			IsShortable:  true,
		},
		Regular: &schwabdev.Regular{
			RegularMarketLastPrice:     182.49,
			RegularMarketNetChange:     1.74,
			RegularMarketPercentChange: 0.96,
			RegularMarketLastSize:      100,
		},
	}
	got := roundtrip(t, input)
	if got.Symbol != "AAPL" {
		t.Errorf("Symbol: want AAPL, got %s", got.Symbol)
	}
	if got.QuoteData == nil {
		t.Fatal("QuoteData is nil after roundtrip")
	}
	if got.QuoteData.AskPrice != 182.50 {
		t.Errorf("AskPrice: want 182.50, got %f", got.QuoteData.AskPrice)
	}
	if got.QuoteData.Volatility != 25.5 {
		t.Errorf("Volatility: want 25.5, got %f", got.QuoteData.Volatility)
	}
	if got.Fundamental == nil {
		t.Fatal("Fundamental is nil after roundtrip")
	}
	if got.Fundamental.PeRatio != 28.5 {
		t.Errorf("PeRatio: want 28.5, got %f", got.Fundamental.PeRatio)
	}
	if got.Fundamental.FundStrategy != "A" {
		t.Errorf("FundStrategy: want A, got %s", got.Fundamental.FundStrategy)
	}
	if got.Reference == nil {
		t.Fatal("Reference is nil after roundtrip")
	}
	if got.Regular == nil {
		t.Fatal("Regular is nil after roundtrip")
	}
}

func TestQuotesResponse_UnmarshalFromAPI(t *testing.T) {
	raw := `{
		"AAPL": {
			"assetMainType": "EQUITY",
			"symbol": "AAPL",
			"realtime": true,
			"ssid": 1001,
			"quote": {
				"askPrice": 182.50,
				"bidPrice": 182.45,
				"lastPrice": 182.48,
				"totalVolume": 40000000,
				"closePrice": 180.00,
				"highPrice": 183.00,
				"lowPrice": 181.00,
				"openPrice": 181.50,
				"netChange": 2.48,
				"netPercentChange": 1.38,
				"mark": 182.48,
				"securityStatus": "Normal",
				"volatility": 22.5
			}
		},
		"MSFT": {
			"assetMainType": "EQUITY",
			"symbol": "MSFT",
			"realtime": true,
			"ssid": 1002,
			"quote": {
				"askPrice": 415.00,
				"bidPrice": 414.95,
				"lastPrice": 414.98,
				"totalVolume": 20000000,
				"closePrice": 410.00,
				"highPrice": 416.00,
				"lowPrice": 413.00,
				"openPrice": 413.50,
				"netChange": 4.98,
				"netPercentChange": 1.21,
				"mark": 414.98,
				"securityStatus": "Normal"
			}
		}
	}`
	got := mustUnmarshal[schwabdev.QuotesResponse](t, raw)
	if len(got) != 2 {
		t.Fatalf("want 2 quotes, got %d", len(got))
	}
	aapl, ok := got["AAPL"]
	if !ok {
		t.Fatal("AAPL missing from response")
	}
	if aapl.QuoteData == nil {
		t.Fatal("AAPL QuoteData is nil")
	}
	if aapl.QuoteData.AskPrice != 182.50 {
		t.Errorf("AAPL AskPrice: want 182.50, got %f", aapl.QuoteData.AskPrice)
	}
	if aapl.QuoteData.Volatility != 22.5 {
		t.Errorf("AAPL Volatility: want 22.5, got %f", aapl.QuoteData.Volatility)
	}
	msft := got["MSFT"]
	if msft.QuoteData.LastPrice != 414.98 {
		t.Errorf("MSFT LastPrice: want 414.98, got %f", msft.QuoteData.LastPrice)
	}
}

func TestQuoteData_52WeekFieldNames(t *testing.T) {
	// Verify the non-standard JSON field names for 52-week high/low decode correctly.
	raw := `{"52WeekHigh": 199.62, "52WeekLow": 124.17}`
	got := mustUnmarshal[schwabdev.QuoteData](t, raw)
	if got.FiftyTwoWeekHigh != 199.62 {
		t.Errorf("52WeekHigh: want 199.62, got %f", got.FiftyTwoWeekHigh)
	}
	if got.FiftyTwoWeekLow != 124.17 {
		t.Errorf("52WeekLow: want 124.17, got %f", got.FiftyTwoWeekLow)
	}
}

func TestQuote_WithExtendedMarket(t *testing.T) {
	// Verify that the Extended field decodes as ExtendedMarket (not bool).
	raw := `{
		"assetMainType": "EQUITY",
		"symbol": "AAPL",
		"realtime": true,
		"ssid": 1001,
		"extended": {
			"askPrice": 182.55,
			"askSize": 500,
			"bidPrice": 182.50,
			"bidSize": 300,
			"lastPrice": 182.52,
			"lastSize": 100,
			"mark": 182.51,
			"quoteTime": 1700000000000,
			"totalVolume": 5000000,
			"tradeTime": 1700000001000
		}
	}`
	got := mustUnmarshal[schwabdev.Quote](t, raw)
	if got.Extended == nil {
		t.Fatal("Extended is nil after unmarshal")
	}
	if got.Extended.AskPrice != 182.55 {
		t.Errorf("Extended AskPrice: want 182.55, got %f", got.Extended.AskPrice)
	}
	if got.Extended.BidSize != 300 {
		t.Errorf("Extended BidSize: want 300, got %d", got.Extended.BidSize)
	}
}

// ── Option Chains ─────────────────────────────────────────────────────────────

func TestOptionChainsResponse_RoundTrip(t *testing.T) {
	input := schwabdev.OptionChainsResponse{
		Symbol:          "SPY",
		Status:          "SUCCESS",
		Strategy:        "SINGLE",
		IsDelayed:       false,
		IsIndex:         false,
		InterestRate:    5.25,
		UnderlyingPrice: 450.75,
		Volatility:      15.5,
		CallExpDateMap: map[string]map[string][]schwabdev.OptionContract{
			"2024-01-19:4": {
				"450.0": {
					{
						PutCall:          "CALL",
						Symbol:           "SPY   240119C00450000",
						Description:      "SPY Jan 19 2024 450 Call",
						BidPrice:         3.50,
						AskPrice:         3.55,
						LastPrice:        3.52,
						MarkPrice:        3.525,
						Delta:            0.52,
						Gamma:            0.03,
						Theta:            -0.15,
						Vega:             0.25,
						StrikePrice:      450.0,
						ExpirationDate:   "2024-01-19T00:00:00.000+0000",
						DaysToExpiration: 4,
						OpenInterest:     15000,
						TotalVolume:      5000,
					},
				},
			},
		},
		PutExpDateMap: map[string]map[string][]schwabdev.OptionContract{},
	}
	got := roundtrip(t, input)
	if got.Symbol != "SPY" {
		t.Errorf("Symbol: want SPY, got %s", got.Symbol)
	}
	calls, ok := got.CallExpDateMap["2024-01-19:4"]
	if !ok {
		t.Fatal("expiry key missing from CallExpDateMap")
	}
	contracts := calls["450.0"]
	if len(contracts) != 1 {
		t.Fatalf("want 1 contract, got %d", len(contracts))
	}
	if contracts[0].Delta != 0.52 {
		t.Errorf("Delta: want 0.52, got %f", contracts[0].Delta)
	}
}

func TestOptionChainsResponse_UnmarshalFromAPI(t *testing.T) {
	raw := `{
		"symbol": "AAPL",
		"status": "SUCCESS",
		"strategy": "SINGLE",
		"interval": 0.0,
		"isDelayed": false,
		"isIndex": false,
		"interestRate": 5.3,
		"underlyingPrice": 182.5,
		"volatility": 25.0,
		"daysToExpiration": 0.0,
		"numberOfContracts": 10,
		"callExpDateMap": {
			"2024-02-16:32": {
				"185.0": [
					{
						"putCall": "CALL",
						"symbol": "AAPL  240216C00185000",
						"description": "AAPL Feb 16 2024 185 Call",
						"bidPrice": 2.10,
						"askPrice": 2.15,
						"lastPrice": 2.12,
						"markPrice": 2.125,
						"strikePrice": 185.0,
						"expirationDate": "2024-02-16T00:00:00.000+0000",
						"daysToExpiration": 32,
						"delta": 0.38,
						"openInterest": 8000,
						"inTheMoney": false
					}
				]
			}
		},
		"putExpDateMap": {}
	}`
	got := mustUnmarshal[schwabdev.OptionChainsResponse](t, raw)
	if got.Symbol != "AAPL" {
		t.Errorf("Symbol: want AAPL, got %s", got.Symbol)
	}
	expiry := got.CallExpDateMap["2024-02-16:32"]
	if expiry == nil {
		t.Fatal("expiry key missing")
	}
	contracts := expiry["185.0"]
	if len(contracts) != 1 {
		t.Fatalf("want 1 contract, got %d", len(contracts))
	}
	if contracts[0].Delta != 0.38 {
		t.Errorf("Delta: want 0.38, got %f", contracts[0].Delta)
	}
}

func TestOptionExpirationChainResponse_RoundTrip(t *testing.T) {
	input := schwabdev.OptionExpirationChainResponse{
		ExpirationList: []*schwabdev.ExpirationDate{
			{
				Expiration:       "2024-01-19",
				DaysToExpiration: 4,
				ExpirationType:   "W",
				SettlementType:   "P",
				Standard:         true,
			},
			{
				Expiration:       "2024-02-16",
				DaysToExpiration: 32,
				ExpirationType:   "R",
				SettlementType:   "P",
				Standard:         true,
			},
		},
	}
	got := roundtrip(t, input)
	if len(got.ExpirationList) != 2 {
		t.Fatalf("want 2, got %d", len(got.ExpirationList))
	}
	if got.ExpirationList[0].DaysToExpiration != 4 {
		t.Errorf("DaysToExpiration: want 4, got %d", got.ExpirationList[0].DaysToExpiration)
	}
}

func TestExpirationDate_OptionRootsString(t *testing.T) {
	// Verify that optionRoots is a string (per OpenAPI spec), not an array.
	raw := `{"expiration": "2024-01-19", "optionRoots": "AAPL,GOOGL"}`
	got := mustUnmarshal[schwabdev.ExpirationDate](t, raw)
	if got.OptionRoots != "AAPL,GOOGL" {
		t.Errorf("OptionRoots: want 'AAPL,GOOGL', got %s", got.OptionRoots)
	}
}

// ── Price History ─────────────────────────────────────────────────────────────

func TestPriceHistoryResponse_RoundTrip(t *testing.T) {
	input := schwabdev.PriceHistoryResponse{
		Symbol: "TSLA",
		Empty:  false,
		Candles: []*schwabdev.Candle{
			{Open: 200.0, High: 215.5, Low: 198.3, Close: 212.0, Volume: 80000000, Datetime: 1700000000000},
			{Open: 212.0, High: 220.0, Low: 210.0, Close: 218.5, Volume: 70000000, Datetime: 1700086400000},
		},
	}
	got := roundtrip(t, input)
	if got.Symbol != "TSLA" {
		t.Errorf("Symbol: want TSLA, got %s", got.Symbol)
	}
	if len(got.Candles) != 2 {
		t.Fatalf("want 2 candles, got %d", len(got.Candles))
	}
	if got.Candles[0].Close != 212.0 {
		t.Errorf("Close: want 212.0, got %f", got.Candles[0].Close)
	}
	if got.Candles[1].Volume != 70000000 {
		t.Errorf("Volume: want 70000000, got %d", got.Candles[1].Volume)
	}
}

func TestPriceHistoryResponse_UnmarshalFromAPI(t *testing.T) {
	raw := `{
		"candles": [
			{"open": 150.0, "high": 155.0, "low": 149.0, "close": 153.5, "volume": 30000000, "datetime": 1705622400000},
			{"open": 153.5, "high": 158.0, "low": 152.0, "close": 157.0, "volume": 28000000, "datetime": 1705708800000}
		],
		"symbol": "AAPL",
		"empty": false
	}`
	got := mustUnmarshal[schwabdev.PriceHistoryResponse](t, raw)
	if got.Symbol != "AAPL" {
		t.Errorf("Symbol: want AAPL, got %s", got.Symbol)
	}
	if got.Empty {
		t.Error("Empty should be false")
	}
	if len(got.Candles) != 2 {
		t.Fatalf("want 2 candles, got %d", len(got.Candles))
	}
	if got.Candles[0].Datetime != 1705622400000 {
		t.Errorf("Datetime: want 1705622400000, got %d", got.Candles[0].Datetime)
	}
}

func TestPriceHistoryResponse_Empty(t *testing.T) {
	raw := `{"symbol": "FAKE", "empty": true, "candles": []}`
	got := mustUnmarshal[schwabdev.PriceHistoryResponse](t, raw)
	if !got.Empty {
		t.Error("Empty should be true")
	}
	if len(got.Candles) != 0 {
		t.Errorf("want 0 candles, got %d", len(got.Candles))
	}
}

func TestPriceHistoryResponse_PreviousCloseDateInt64(t *testing.T) {
	// Verify that previousCloseDate is int64 (per OpenAPI CandleList spec).
	raw := `{
		"candles": [],
		"symbol": "AAPL",
		"empty": false,
		"previousClose": 182.50,
		"previousCloseDate": 1705180800000,
		"previousCloseDateISO8601": "2024-01-14"
	}`
	got := mustUnmarshal[schwabdev.PriceHistoryResponse](t, raw)
	if got.PreviousClose != 182.50 {
		t.Errorf("PreviousClose: want 182.50, got %f", got.PreviousClose)
	}
	if got.PreviousCloseDate != 1705180800000 {
		t.Errorf("PreviousCloseDate: want 1705180800000, got %d", got.PreviousCloseDate)
	}
}

// ── Movers ────────────────────────────────────────────────────────────────────

func TestMoversResponse_RoundTrip(t *testing.T) {
	input := schwabdev.MoversResponse{
		{Symbol: "NVDA", Description: "NVIDIA Corp", LastPrice: 650.00, Change: 25.50, TotalVolume: 45000000},
		{Symbol: "AMD", Description: "Adv Micro Devices", LastPrice: 180.00, Change: -3.20, TotalVolume: 30000000},
	}
	got := roundtrip(t, input)
	if len(got) != 2 {
		t.Fatalf("want 2 movers, got %d", len(got))
	}
	if got[0].Change != 25.50 {
		t.Errorf("Change: want 25.50, got %f", got[0].Change)
	}
	if got[1].TotalVolume != 30000000 {
		t.Errorf("TotalVolume: want 30000000, got %d", got[1].TotalVolume)
	}
}

// ── Market Hours ──────────────────────────────────────────────────────────────

func TestMarketHoursResponse_RoundTrip(t *testing.T) {
	input := schwabdev.MarketHoursResponse{
		"equity": schwabdev.MarketHour{
			Category:    "EQUITY",
			Date:        "2024-01-15",
			Exchange:    "NYSE",
			IsOpen:      true,
			MarketType:  "EQUITY",
			Product:     "EQO",
			ProductName: "equity",
			SessionHours: &schwabdev.SessionHours{
				RegularMarket: []*schwabdev.Interval{
					{Start: "2024-01-15T09:30:00-05:00", End: "2024-01-15T16:00:00-05:00"},
				},
			},
		},
	}
	got := roundtrip(t, input)
	equity, ok := got["equity"]
	if !ok {
		t.Fatal("equity key missing")
	}
	if !equity.IsOpen {
		t.Error("IsOpen should be true")
	}
	if equity.SessionHours == nil {
		t.Fatal("SessionHours is nil")
	}
}

func TestMarketHoursResponse_UnmarshalFromAPI(t *testing.T) {
	raw := `{
		"equity": {
			"EQO": {
				"date": "2024-01-15",
				"marketType": "EQUITY",
				"exchange": "NULL",
				"category": "NULL",
				"isOpen": true,
				"product": "EQO",
				"productName": "equity",
				"sessionHours": {
					"preMarket": [{"start": "2024-01-15T07:00:00-05:00", "end": "2024-01-15T09:30:00-05:00"}],
					"regularMarket": [{"start": "2024-01-15T09:30:00-05:00", "end": "2024-01-15T16:00:00-05:00"}]
				}
			}
		}
	}`
	// The API wraps by market type then product — test that outer map decodes.
	var raw2 map[string]map[string]schwabdev.MarketHour
	if err := json.Unmarshal([]byte(raw), &raw2); err != nil {
		t.Fatalf("unmarshal nested market hours: %v", err)
	}
	equityMap := raw2["equity"]
	mh := equityMap["EQO"]
	if !mh.IsOpen {
		t.Error("IsOpen should be true")
	}
}

func TestMarketHour_Closed(t *testing.T) {
	raw := `{"category": "EQUITY", "date": "2024-01-13", "isOpen": false, "marketType": "EQUITY", "product": "EQO", "productName": "equity"}`
	got := mustUnmarshal[schwabdev.MarketHour](t, raw)
	if got.IsOpen {
		t.Error("IsOpen should be false for weekend")
	}
	if got.SessionHours != nil {
		t.Error("SessionHours should be nil when market is closed")
	}
}

// ── Instruments ───────────────────────────────────────────────────────────────

func TestInstrumentsResponse_RoundTrip(t *testing.T) {
	input := schwabdev.InstrumentsResponse{
		{Symbol: "AAPL", Description: "Apple Inc", AssetType: "EQUITY", Cusip: "037833100", Exchange: "NASDAQ"},
		{Symbol: "AAPL1", Description: "Apple Inc Adj", AssetType: "EQUITY", Cusip: "037833209", Exchange: "NASDAQ"},
	}
	got := roundtrip(t, input)
	if len(got) != 2 {
		t.Fatalf("want 2, got %d", len(got))
	}
	if got[0].Cusip != "037833100" {
		t.Errorf("Cusip: want 037833100, got %s", got[0].Cusip)
	}
}

func TestInstrumentCUSIPResponse_RoundTrip(t *testing.T) {
	input := schwabdev.InstrumentCUSIPResponse{
		{
			Cusip:       "037833100",
			Symbol:      "AAPL",
			Description: "Apple Inc",
			Exchange:    "NASDAQ",
			AssetType:   "EQUITY",
		},
	}
	got := roundtrip(t, input)
	if len(got) != 1 {
		t.Fatalf("want 1, got %d", len(got))
	}
	if got[0].Symbol != "AAPL" {
		t.Errorf("Symbol: want AAPL, got %s", got[0].Symbol)
	}
}

func TestInstrumentSearch_EnrichedFields(t *testing.T) {
	// Verify that the enriched InstrumentSearch (matching InstrumentResponse) decodes
	// the additional fields: bondFactor, bondMultiplier, bondPrice, type, fundamental, instrumentInfo, bondInstrumentInfo.
	raw := `{
		"cusip": "037833100",
		"symbol": "AAPL",
		"description": "Apple Inc",
		"exchange": "NASDAQ",
		"assetType": "EQUITY",
		"bondFactor": "1.0",
		"bondMultiplier": "100",
		"bondPrice": 101.50,
		"type": "COMMON_STOCK",
		"fundamental": {
			"symbol": "AAPL",
			"high52": 199.62,
			"low52": 124.17,
			"peRatio": 28.5,
			"beta": 1.28,
			"eps": 6.43,
			"marketCap": 2800000000000,
			"dividendFreq": 4
		},
		"instrumentInfo": {
			"cusip": "037833100",
			"symbol": "AAPL",
			"description": "Apple Inc",
			"exchange": "NASDAQ",
			"assetType": "EQUITY",
			"type": "COMMON_STOCK"
		}
	}`
	got := mustUnmarshal[schwabdev.InstrumentSearch](t, raw)
	if got.BondFactor != "1.0" {
		t.Errorf("BondFactor: want '1.0', got %s", got.BondFactor)
	}
	if got.BondMultiplier != "100" {
		t.Errorf("BondMultiplier: want '100', got %s", got.BondMultiplier)
	}
	if got.BondPrice != 101.50 {
		t.Errorf("BondPrice: want 101.50, got %f", got.BondPrice)
	}
	if got.Type != "COMMON_STOCK" {
		t.Errorf("Type: want COMMON_STOCK, got %s", got.Type)
	}
	if got.Fundamental == nil {
		t.Fatal("Fundamental is nil")
	}
	if got.Fundamental.PeRatio != 28.5 {
		t.Errorf("Fundamental PeRatio: want 28.5, got %f", got.Fundamental.PeRatio)
	}
	if got.Fundamental.Beta != 1.28 {
		t.Errorf("Fundamental Beta: want 1.28, got %f", got.Fundamental.Beta)
	}
	if got.Fundamental.MarketCap != 2800000000000 {
		t.Errorf("Fundamental MarketCap: want 2800000000000, got %f", got.Fundamental.MarketCap)
	}
	if got.Fundamental.DividendFreq != 4 {
		t.Errorf("Fundamental DividendFreq: want 4, got %d", got.Fundamental.DividendFreq)
	}
	if got.InstrumentInfo == nil {
		t.Fatal("InstrumentInfo is nil")
	}
	if got.InstrumentInfo.Symbol != "AAPL" {
		t.Errorf("InstrumentInfo Symbol: want AAPL, got %s", got.InstrumentInfo.Symbol)
	}
	if got.InstrumentInfo.Type != "COMMON_STOCK" {
		t.Errorf("InstrumentInfo Type: want COMMON_STOCK, got %s", got.InstrumentInfo.Type)
	}
}

func TestMarketDataBond_RoundTrip(t *testing.T) {
	input := schwabdev.MarketDataBond{
		Cusip:          "345370100",
		Symbol:         "GOVT10Y",
		Description:    "US Treasury 10 Year",
		Exchange:       "GOVT",
		AssetType:      "BOND",
		BondFactor:     "1.0",
		BondMultiplier: "1000",
		BondPrice:      98.75,
		Type:           "US_TREASURY_NOTE",
	}
	got := roundtrip(t, input)
	if got.BondPrice != 98.75 {
		t.Errorf("BondPrice: want 98.75, got %f", got.BondPrice)
	}
	if got.Type != "US_TREASURY_NOTE" {
		t.Errorf("Type: want US_TREASURY_NOTE, got %s", got.Type)
	}
}

// ── Streamer Info ─────────────────────────────────────────────────────────────

func TestPreferencesResponse_RoundTrip(t *testing.T) {
	input := schwabdev.PreferencesResponse{
		StreamerInfo: []*schwabdev.StreamerInfo{
			{
				StreamerURL:            "wss://streamer.schwab.com/ws",
				SchwabClientCustomerID: "customer-xyz",
				SchwabClientCorrelID:   "correl-abc-123",
				SchwabClientChannel:    "IO",
				SchwabClientFunctionID: "APIAPP",
			},
		},
	}
	got := roundtrip(t, input)
	if len(got.StreamerInfo) != 1 {
		t.Fatalf("want 1 streamer info, got %d", len(got.StreamerInfo))
	}
	if got.StreamerInfo[0].StreamerURL != "wss://streamer.schwab.com/ws" {
		t.Errorf("StreamerURL: want wss://streamer.schwab.com/ws, got %s", got.StreamerInfo[0].StreamerURL)
	}
}

func TestPreferencesResponse_UnmarshalFromAPI(t *testing.T) {
	raw := `{
		"streamerInfo": [
			{
				"streamerSocketUrl": "wss://streamer.schwab.com/ws",
				"schwabClientCorrelId": "abc-correl-123",
				"schwabClientChannel": "IO",
				"schwabClientFunctionId": "APIAPP",
				"schwabClientCustomerId": "customer-xyz"
			}
		]
	}`
	got := mustUnmarshal[schwabdev.PreferencesResponse](t, raw)
	if len(got.StreamerInfo) != 1 {
		t.Fatalf("want 1, got %d", len(got.StreamerInfo))
	}
}

// ── Order Requests (marshalling out to API) ───────────────────────────────────

func TestOrderRequest_MarshalToAPI(t *testing.T) {
	input := schwabdev.OrderRequest{
		OrderType:         "LIMIT",
		Session:           "NORMAL",
		Duration:          "DAY",
		OrderStrategyType: "SINGLE",
		Price:             150.00,
		OrderLegCollection: []*schwabdev.OrderLegRequest{
			{
				Instruction: "BUY",
				Quantity:    10,
				Instrument: &schwabdev.Instrument{
					Symbol:    "AAPL",
					AssetType: "EQUITY",
				},
			},
		},
	}
	b, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	// Unmarshal back and verify key fields survive.
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("unmarshal to map: %v", err)
	}
	if m["orderType"] != "LIMIT" {
		t.Errorf("orderType: want LIMIT, got %v", m["orderType"])
	}
	if m["price"] != 150.0 {
		t.Errorf("price: want 150.00, got %v", m["price"])
	}
	legs, ok := m["orderLegCollection"].([]any)
	if !ok || len(legs) != 1 {
		t.Fatalf("orderLegCollection: want 1 leg, got %v", m["orderLegCollection"])
	}
}

func TestOrderRequest_OmitsEmptyOptionals(t *testing.T) {
	// stopPrice and complexOrderStrategyType are omitempty — verify they're
	// absent from the JSON when zero-valued.
	input := schwabdev.OrderRequest{
		OrderType:         "MARKET",
		Session:           "NORMAL",
		Duration:          "DAY",
		OrderStrategyType: "SINGLE",
		OrderLegCollection: []*schwabdev.OrderLegRequest{
			{Instruction: "SELL", Quantity: 5, Instrument: &schwabdev.Instrument{Symbol: "TSLA", AssetType: "EQUITY"}},
		},
	}
	b, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var m map[string]any
	json.Unmarshal(b, &m)
	if _, present := m["stopPrice"]; present {
		t.Error("stopPrice should be omitted when empty")
	}
	if _, present := m["complexOrderStrategyType"]; present {
		t.Error("complexOrderStrategyType should be omitted when empty")
	}
}

// ── Preview Order ─────────────────────────────────────────────────────────────

func TestPreviewOrderResponse_RoundTrip(t *testing.T) {
	input := schwabdev.PreviewOrderResponse{
		OrderID: 999,
		OrderValidationResult: &schwabdev.OrderValidationResult{
			Rejects: []*schwabdev.OrderValidationDetail{},
		},
		CommissionAndFee: &schwabdev.CommissionAndFee{
			Commission: &schwabdev.Commission{
				CommissionLegs: []*schwabdev.CommissionLeg{
					{CommissionValues: []*schwabdev.CommissionValue{{Value: 0.0, Type: "COMMISSION"}}},
				},
			},
			Fee: &schwabdev.Fee{
				FeeLegs: []*schwabdev.FeeLeg{
					{FeeValues: []*schwabdev.FeeValue{{Value: 0.00055, Type: "SEC_FEE"}}},
				},
			},
		},
	}
	got := roundtrip(t, input)
	if got.OrderID != 999 {
		t.Errorf("OrderID: want 999, got %d", got.OrderID)
	}
	if got.CommissionAndFee == nil {
		t.Fatal("CommissionAndFee is nil after roundtrip")
	}
	if got.CommissionAndFee.Fee.FeeLegs[0].FeeValues[0].Value != 0.00055 {
		t.Errorf("FeeValue: want 0.00055, got %f", got.CommissionAndFee.Fee.FeeLegs[0].FeeValues[0].Value)
	}
}

// ── Position ──────────────────────────────────────────────────────────────────

func TestPosition_RoundTrip(t *testing.T) {
	input := schwabdev.Position{
		ShortQuantity:                0,
		AveragePrice:                 155.25,
		MarketValue:                  15525.00,
		LongQuantity:                 100,
		PreviousSessionLongQuantity:  100,
		PreviousSessionShortQuantity: 0,
		Instrument: &schwabdev.Instrument{
			AssetType:    "EQUITY",
			Cusip:        "594918104",
			Symbol:       "MSFT",
			InstrumentID: 7654321,
		},
	}
	got := roundtrip(t, input)
	if got.Instrument == nil || got.Instrument.Symbol != "MSFT" {
		t.Errorf("Instrument symbol: want MSFT, got %v", got.Instrument)
	}
	if got.LongQuantity != 100 {
		t.Errorf("LongQuantity: want 100, got %f", got.LongQuantity)
	}
	if got.AveragePrice != 155.25 {
		t.Errorf("AveragePrice: want 155.25, got %f", got.AveragePrice)
	}
}

// ── TokenRecord (storage) ─────────────────────────────────────────────────────

func TestTokenRecord_RoundTrip(t *testing.T) {
	// Verify the storage record itself round-trips cleanly through JSON,
	// which is how FileTokenStorage persists it.
	now := mustUnmarshal[schwabdev.TokenRecord](t, `{
		"access_token_issued":  "2024-01-15T10:00:00Z",
		"refresh_token_issued": "2024-01-10T10:00:00Z",
		"access_token":         "eyJhbGciOiJSUzI1NiJ9.test",
		"refresh_token":        "refreshtokenvalue",
		"id_token":             "idtokenvalue",
		"expires_in":           1800,
		"token_type":           "Bearer",
		"scope":                "api"
	}`)

	b, err := json.Marshal(now)
	if err != nil {
		t.Fatalf("marshal TokenRecord: %v", err)
	}
	var got schwabdev.TokenRecord
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("unmarshal TokenRecord: %v", err)
	}
	if got.AccessToken != "eyJhbGciOiJSUzI1NiJ9.test" {
		t.Errorf("AccessToken mismatch: %s", got.AccessToken)
	}
	if got.ExpiresIn != 1800 {
		t.Errorf("ExpiresIn: want 1800, got %d", got.ExpiresIn)
	}
	if got.AccessTokenIssued.IsZero() {
		t.Error("AccessTokenIssued should not be zero")
	}
}

// ── Spec-Aligned Constants ────────────────────────────────────────────────────

func TestConstants_TraderAPI(t *testing.T) {
	// Verify key enum values match the OpenAPI spec exactly.
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"InstructionSellShortExempt", schwabdev.InstructionSellShortExempt, "SELL_SHORT_EXEMPT"},
		{"DurationImmediateOrCancel", schwabdev.DurationImmediateOrCancel, "IMMEDIATE_OR_CANCEL"},
		{"DurationEndOfWeek", schwabdev.DurationEndOfWeek, "END_OF_WEEK"},
		{"OrderTypeCabinet", schwabdev.OrderTypeCabinet, "CABINET"},
		{"OrderTypeNonMarketable", schwabdev.OrderTypeNonMarketable, "NON_MARKETABLE"},
		{"OrderTypeLimitOnClose", schwabdev.OrderTypeLimitOnClose, "LIMIT_ON_CLOSE"},
		{"OrderTypeUnknown", schwabdev.OrderTypeUnknown, "UNKNOWN"},
		{"OrderStrategyTypeCancel", schwabdev.OrderStrategyTypeCancel, "CANCEL"},
		{"OrderStrategyTypeFlatten", schwabdev.OrderStrategyTypeFlatten, "FLATTEN"},
		{"ComplexOrderStrategyButterfly", schwabdev.ComplexOrderStrategyButterfly, "BUTTERFLY"},
		{"ComplexOrderStrategyBackRatio", schwabdev.ComplexOrderStrategyBackRatio, "BACK_RATIO"},
		{"ComplexOrderStrategyVerticalRoll", schwabdev.ComplexOrderStrategyVerticalRoll, "VERTICAL_ROLL"},
		{"ComplexOrderStrategyIronCondor", schwabdev.ComplexOrderStrategyIronCondor, "IRON_CONDOR"},
		{"ComplexOrderStrategyDoubleDiagonal", schwabdev.ComplexOrderStrategyDoubleDiagonal, "DOUBLE_DIAGONAL"},
		{"PositionEffectOpening", schwabdev.PositionEffectOpening, "OPENING"},
		{"PositionEffectClosing", schwabdev.PositionEffectClosing, "CLOSING"},
		{"PositionEffectAutomatic", schwabdev.PositionEffectAutomatic, "AUTOMATIC"},
		{"AssetTypeFuture", schwabdev.AssetTypeFuture, "FUTURE"},
		{"AssetTypeForex", schwabdev.AssetTypeForex, "FOREX"},
		{"AssetTypeCashEquivalent", schwabdev.AssetTypeCashEquivalent, "CASH_EQUIVALENT"},
		{"AssetTypeCollectiveInvestment", schwabdev.AssetTypeCollectiveInvestment, "COLLECTIVE_INVESTMENT"},
		{"OrderStatusNew", schwabdev.OrderStatusNew, "NEW"},
		{"OrderStatusUnknown", schwabdev.OrderStatusUnknown, "UNKNOWN"},
		{"OrderStatusPendingRecall", schwabdev.OrderStatusPendingRecall, "PENDING_RECALL"},
		{"TransactionTypeMemorandum", schwabdev.TransactionTypeMemorandum, "MEMORANDUM"},
		{"TransactionTypeSmaAdjustment", schwabdev.TransactionTypeSmaAdjustment, "SMA_ADJUSTMENT"},
		{"TransactionTypeMarginCall", schwabdev.TransactionTypeMarginCall, "MARGIN_CALL"},
		{"RequestedDestinationECNArca", schwabdev.RequestedDestinationECNArca, "ECN_ARCA"},
		{"SpecialInstructionAllOrNone", schwabdev.SpecialInstructionAllOrNone, "ALL_OR_NONE"},
		{"StopPriceLinkBasisAskBid", schwabdev.StopPriceLinkBasisAskBid, "ASK_BID"},
		{"TaxLotMethodLossHarvester", schwabdev.TaxLotMethodLossHarvester, "LOSS_HARVESTER"},
		{"FeeTypeFTT", schwabdev.FeeTypeFTT, "FTT"},
		{"FeeTypeTefraTax", schwabdev.FeeTypeTefraTax, "TEFRA_TAX"},
		{"TransferItemPositionEffectUnknown", schwabdev.TransferItemPositionEffectUnknown, "UNKNOWN"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Errorf("%s: want %q, got %q", tc.name, tc.want, tc.got)
			}
		})
	}
}

func TestConstants_OrderTypeRequest(t *testing.T) {
	// Verify that OrderTypeRequest constants exist and match OrderType values (without UNKNOWN).
	tests := []struct {
		name     string
		reqConst string
		want     string
	}{
		{"Market", schwabdev.OrderTypeRequestMarket, "MARKET"},
		{"Limit", schwabdev.OrderTypeRequestLimit, "LIMIT"},
		{"Stop", schwabdev.OrderTypeRequestStop, "STOP"},
		{"StopLimit", schwabdev.OrderTypeRequestStopLimit, "STOP_LIMIT"},
		{"TrailingStop", schwabdev.OrderTypeRequestTrailingStop, "TRAILING_STOP"},
		{"Cabinet", schwabdev.OrderTypeRequestCabinet, "CABINET"},
		{"NonMarketable", schwabdev.OrderTypeRequestNonMarketable, "NON_MARKETABLE"},
		{"MarketOnClose", schwabdev.OrderTypeRequestMarketOnClose, "MARKET_ON_CLOSE"},
		{"Exercise", schwabdev.OrderTypeRequestExercise, "EXERCISE"},
		{"TrailingStopLimit", schwabdev.OrderTypeRequestTrailingStopLimit, "TRAILING_STOP_LIMIT"},
		{"NetDebit", schwabdev.OrderTypeRequestNetDebit, "NET_DEBIT"},
		{"NetCredit", schwabdev.OrderTypeRequestNetCredit, "NET_CREDIT"},
		{"NetZero", schwabdev.OrderTypeRequestNetZero, "NET_ZERO"},
		{"LimitOnClose", schwabdev.OrderTypeRequestLimitOnClose, "LIMIT_ON_CLOSE"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.reqConst != tc.want {
				t.Errorf("%s: want %q, got %q", tc.name, tc.want, tc.reqConst)
			}
		})
	}
}

func TestConstants_MarketDataAPI(t *testing.T) {
	// Verify market data enum values match the OpenAPI spec.
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"AssetMainTypeBond", schwabdev.AssetMainTypeBond, "BOND"},
		{"AssetMainTypeFutureOption", schwabdev.AssetMainTypeFutureOption, "FUTURE_OPTION"},
		{"EquityAssetSubTypeETF", schwabdev.EquityAssetSubTypeETF, "ETF"},
		{"MutualFundAssetSubTypeMMF", schwabdev.MutualFundAssetSubTypeMMF, "MMF"},
		{"QuoteTypeNBBO", schwabdev.QuoteTypeNBBO, "NBBO"},
		{"QuoteTypeNFL", schwabdev.QuoteTypeNFL, "NFL"},
		{"FundStrategyActive", schwabdev.FundStrategyActive, "A"},
		{"FundStrategyQuantitative", schwabdev.FundStrategyQuantitative, "Q"},
		{"MDContractTypePut", schwabdev.MDContractTypePut, "P"},
		{"MDContractTypeCall", schwabdev.MDContractTypeCall, "C"},
		{"MDExpirationTypeMonthly", schwabdev.MDExpirationTypeMonthly, "M"},
		{"MDExpirationTypeWeekly", schwabdev.MDExpirationTypeWeekly, "W"},
		{"MDSettlementTypeAM", schwabdev.MDSettlementTypeAM, "A"},
		{"MDSettlementTypePM", schwabdev.MDSettlementTypePM, "P"},
		{"MDExerciseTypeAmerican", schwabdev.MDExerciseTypeAmerican, "A"},
		{"MDExerciseTypeEuropean", schwabdev.MDExerciseTypeEuropean, "E"},
		{"SortTypeVolume", schwabdev.SortTypeVolume, "VOLUME"},
		{"PeriodTypeDay", schwabdev.PeriodTypeDay, "day"},
		{"FrequencyTypeMinute", schwabdev.FrequencyTypeMinute, "minute"},
		{"EntitlementPayingPro", schwabdev.EntitlementPayingPro, "PN"},
		{"ExpMonthJan", schwabdev.ExpMonthJan, "JAN"},
		{"ExpMonthAll", schwabdev.ExpMonthAll, "ALL"},
		{"MoverSymbolDJI", schwabdev.MoverSymbolDJI, "$DJI"},
		{"MoverSymbolOptionPut", schwabdev.MoverSymbolOptionPut, "OPTION_PUT"},
		{"ChainStrategyButterfly", schwabdev.ChainStrategyButterfly, "BUTTERFLY"},
		{"ChainStrategyAnalytical", schwabdev.ChainStrategyAnalytical, "ANALYTICAL"},
		{"ProjectionSearch", schwabdev.ProjectionSearch, "search"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Errorf("%s: want %q, got %q", tc.name, tc.want, tc.got)
			}
		})
	}
}

func TestConstants_DivFreq(t *testing.T) {
	// DivFreq is an integer enum.
	tests := []struct {
		name string
		got  int
		want int
	}{
		{"Annually", schwabdev.DivFreqAnnually, 1},
		{"Quarterly", schwabdev.DivFreqQuarterly, 4},
		{"Monthly", schwabdev.DivFreqMonthly, 12},
		{"BiMonthly", schwabdev.DivFreqBiMonthly, 6},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Errorf("%s: want %d, got %d", tc.name, tc.want, tc.got)
			}
		})
	}
}

func TestConstants_MoverFrequency(t *testing.T) {
	// MoverFrequency is an integer enum.
	tests := []struct {
		name string
		got  int
		want int
	}{
		{"Freq0", schwabdev.MoverFrequency0, 0},
		{"Freq5", schwabdev.MoverFrequency5, 5},
		{"Freq60", schwabdev.MoverFrequency60, 60},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Errorf("%s: want %d, got %d", tc.name, tc.want, tc.got)
			}
		})
	}
}

// TestQuoteNestedQuoteObject is a regression test for the v0.0.19 fix:
// Schwab nests quote data under the "quote" object, so QuoteData must be
// tagged json:"quote". Before the fix it was an anonymous embedded field,
// which flattens the fields to the top level of the symbol object, leaving
// quote.lastPrice (etc.) unpopulated. Index symbols (no "extended" section,
// e.g. $RVX) then reported LastPrice=0 and were dropped by consumers.
func TestQuoteNestedQuoteObject(t *testing.T) {
	body := []byte(`{
		"$RVX": {
			"assetMainType": "INDEX",
			"quote": {
				"lastPrice": 17.21,
				"bidPrice": 0,
				"askPrice": 0,
				"netChange": 0.11,
				"volatility": 0,
				"totalVolume": 0,
				"quoteTime": 1756075200000
			},
			"reference": {"description": "CBOE Russell 1000 Volatility Index"}
		}
	}`)

	var resp schwabdev.QuotesResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("unmarshal quotes: %v", err)
	}
	q, ok := resp["$RVX"]
	if !ok {
		t.Fatalf("symbol $RVX missing from response")
	}
	if q.QuoteData == nil {
		t.Fatal("QuoteData is nil — the \"quote\" object was not parsed")
	}
	if q.QuoteData.LastPrice != 17.21 {
		t.Fatalf("LastPrice = %v, want 17.21", q.QuoteData.LastPrice)
	}
	if q.Extended != nil {
		t.Fatalf("index symbol unexpectedly has an extended section: %+v", q.Extended)
	}
}

// TestOptionChainsNumericLastTradingDay is a regression test for the v0.0.19
// fix: the live Schwab API returns lastTradingDay as a number (epoch millis),
// not a string as the OpenAPI spec declares. json.Number accepts both.
func TestOptionChainsNumericLastTradingDay(t *testing.T) {
	body := []byte(`{
		"symbol": "SPX",
		"status": "SUCCESS",
		"underlyingPrice": 6012.5,
		"callExpDateMap": {
			"2026-08-26:0": {
				"0.6000.0": [{
					"putCall": "CALL",
					"symbol": "SPXW   260826C06000000",
					"lastTradingDay": 1784937600000,
					"expirationDate": "2026-08-26",
					"daysToExpiration": 0,
					"strikePrice": 6000
				}]
			}
		}
	}`)

	var resp schwabdev.OptionChainsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("unmarshal chains: %v", err)
	}
	contract := resp.CallExpDateMap["2026-08-26:0"]["0.6000.0"][0]
	if contract.LastTradingDay.String() != "1784937600000" {
		t.Fatalf("LastTradingDay = %q, want 1784937600000", contract.LastTradingDay.String())
	}
}
