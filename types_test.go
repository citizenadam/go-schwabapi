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
			InitialBalances: &schwabdev.InitialBalances{
				CashBalance:      10000.50,
				BuyingPower:      20000.00,
				AccountValue:     35000.75,
				LiquidationValue: 34000.00,
			},
			CurrentBalances: &schwabdev.CurrentBalances{
				CashBalance:      9500.00,
				BuyingPower:      19000.00,
				LiquidationValue: 33500.00,
				Equity:           33500.00,
			},
			ProjectedBalances: &schwabdev.ProjectedBalances{
				BuyingPower:    18500.00,
				AvailableFunds: 9000.00,
			},
			Positions: []*schwabdev.Position{
				{
					Symbol:       "AAPL",
					LongQuantity: 100,
					AveragePrice: 150.25,
					MarketValue:  17500.00,
					AssetType:    "EQUITY",
					Cusip:        "037833100",
					InstrumentID: 1234567,
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
	if got.SecuritiesAccount.Positions[0].Symbol != "AAPL" {
		t.Errorf("Position symbol: want AAPL, got %s", got.SecuritiesAccount.Positions[0].Symbol)
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
	if got.SecuritiesAccount.CurrentBalances == nil {
		t.Fatal("CurrentBalances unexpectedly nil")
	}
	if got.SecuritiesAccount.CurrentBalances.CashBalance != 5000.00 {
		t.Errorf("CashBalance: want 5000.00, got %f", got.SecuritiesAccount.CurrentBalances.CashBalance)
	}
}

func TestAccountDetailsAllResponse_NilOptionals(t *testing.T) {
	// Ensure optional pointer fields decode as nil when absent.
	raw := `{"securitiesAccount": {"type": "CASH", "accountNumber": "X", "roundTrips": 0}}`
	got := mustUnmarshal[schwabdev.AccountDetailsAllResponse](t, raw)
	if got.SecuritiesAccount.InitialBalances != nil {
		t.Error("InitialBalances should be nil when absent")
	}
	if got.SecuritiesAccount.Positions != nil {
		t.Error("Positions should be nil when absent")
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
		OrderLegCollection: []*schwabdev.OrderLeg{
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
				ActivityID:             111,
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
		TransactionID: "TXN-001",
		Type:          "TRADE",
		Symbol:        "GOOGL",
		Date:          "2024-03-15",
		Quantity:      5,
		Price:         175.30,
		NetAmount:     876.50,
	}
	got := roundtrip(t, input)
	if got.TransactionID != "TXN-001" {
		t.Errorf("TransactionID: want TXN-001, got %s", got.TransactionID)
	}
	if got.NetAmount != 876.50 {
		t.Errorf("NetAmount: want 876.50, got %f", got.NetAmount)
	}
}

func TestTransactionsResponse_UnmarshalFromAPI(t *testing.T) {
	raw := `[
		{"transactionId": "T1", "type": "TRADE", "symbol": "NVDA", "date": "2024-01-10", "quantity": 10, "price": 500.00, "netAmount": -5001.50},
		{"transactionId": "T2", "type": "DIVIDEND", "symbol": "MSFT", "date": "2024-01-12", "quantity": 0, "price": 0, "netAmount": 25.00}
	]`
	got := mustUnmarshal[schwabdev.TransactionsResponse](t, raw)
	if len(got) != 2 {
		t.Fatalf("want 2, got %d", len(got))
	}
	if got[0].Symbol != "NVDA" {
		t.Errorf("want NVDA, got %s", got[0].Symbol)
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
		Realtime:      true,
		Ssid:          1234567890,
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
		},
		Fundamental: &schwabdev.Fundamental{
			PeRatio:         28.5,
			Eps:             6.43,
			DivYield:        0.54,
			DivAmount:       0.96,
			Avg10DaysVolume: 55000000,
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
	if got.Fundamental == nil {
		t.Fatal("Fundamental is nil after roundtrip")
	}
	if got.Fundamental.PeRatio != 28.5 {
		t.Errorf("PeRatio: want 28.5, got %f", got.Fundamental.PeRatio)
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
				"securityStatus": "Normal"
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

// ── Option Chains ─────────────────────────────────────────────────────────────

func TestOptionChainsResponse_RoundTrip(t *testing.T) {
	input := schwabdev.OptionChainsResponse{
		Symbol:            "SPY",
		Status:            "SUCCESS",
		Strategy:          "SINGLE",
		IsDelayed:         false,
		IsIndex:           false,
		InterestRate:      5.25,
		UnderlyingPrice:   450.75,
		Volatility:        15.5,
		NumberOfContracts: 200,
		CallExpDateMap: map[string]map[string][]schwabdev.OptionContract{
			"2024-01-19:4": {
				"450.0": {
					{
						PutCall:          "CALL",
						Symbol:           "SPY   240119C00450000",
						Description:      "SPY Jan 19 2024 450 Call",
						Bid:              3.50,
						Ask:              3.55,
						Last:             3.52,
						Mark:             3.525,
						Delta:            0.52,
						Gamma:            0.03,
						Theta:            -0.15,
						Vega:             0.25,
						StrikePrice:      450.0,
						ExpirationDate:   "2024-01-19T00:00:00.000+0000",
						DaysToExpiration: 4,
						OpenInterest:     15000,
						TotalVolume:      5000,
						InTheMoney:       true,
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
	if !contracts[0].InTheMoney {
		t.Error("InTheMoney should be true")
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
						"bid": 2.10,
						"ask": 2.15,
						"last": 2.12,
						"mark": 2.125,
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
				ExpirationDate:   "2024-01-19",
				DaysToExpiration: 4,
				ExpirationType:   "W",
				SettlementType:   "P",
				Standard:         true,
			},
			{
				ExpirationDate:   "2024-02-16",
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

// ── Movers ────────────────────────────────────────────────────────────────────

func TestMoversResponse_RoundTrip(t *testing.T) {
	input := schwabdev.MoversResponse{
		{Symbol: "NVDA", Description: "NVIDIA Corp", LastPrice: 650.00, Change: 25.50, PercentChange: 4.08, Volume: 45000000},
		{Symbol: "AMD", Description: "Adv Micro Devices", LastPrice: 180.00, Change: -3.20, PercentChange: -1.75, Volume: 30000000},
	}
	got := roundtrip(t, input)
	if len(got) != 2 {
		t.Fatalf("want 2 movers, got %d", len(got))
	}
	if got[0].PercentChange != 4.08 {
		t.Errorf("PercentChange: want 4.08, got %f", got[0].PercentChange)
	}
	if got[1].Change != -3.20 {
		t.Errorf("Change: want -3.20, got %f", got[1].Change)
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
				SessionDuration: []*schwabdev.SessionDuration{
					{StartDateTime: "2024-01-15T09:30:00-05:00", EndDateTime: "2024-01-15T16:00:00-05:00"},
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
		Instruments: []*schwabdev.InstrumentCUSIP{
			{
				Cusip:       "037833100",
				Symbol:      "AAPL",
				Description: "Apple Inc",
				Exchange:    "NASDAQ",
				AssetType:   "EQUITY",
			},
		},
	}
	got := roundtrip(t, input)
	if len(got.Instruments) != 1 {
		t.Fatalf("want 1, got %d", len(got.Instruments))
	}
	if got.Instruments[0].Symbol != "AAPL" {
		t.Errorf("Symbol: want AAPL, got %s", got.Instruments[0].Symbol)
	}
}

// ── Streamer Info ─────────────────────────────────────────────────────────────

func TestPreferencesResponse_RoundTrip(t *testing.T) {
	input := schwabdev.PreferencesResponse{
		StreamerInfo: []*schwabdev.StreamerInfo{
			{
				StreamerURL:            "wss://streamer.schwab.com/ws",
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
	// Note: streamerSocketUrl vs streamerUrl — verify which tag the struct uses.
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
		Price:             "150.00",
		OrderLegCollection: []*schwabdev.OrderLegRequest{
			{
				Instruction: "BUY",
				Quantity:    10,
				Instrument: &schwabdev.InstrumentRequest{
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
	if m["price"] != "150.00" {
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
			{Instruction: "SELL", Quantity: 5, Instrument: &schwabdev.InstrumentRequest{Symbol: "TSLA", AssetType: "EQUITY"}},
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
			Rejects: []*schwabdev.OrderReject{},
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
		ChangedSinceLastSession:      false,
		AssetType:                    "EQUITY",
		Cusip:                        "594918104",
		Symbol:                       "MSFT",
		InstrumentID:                 7654321,
	}
	got := roundtrip(t, input)
	if got.Symbol != "MSFT" {
		t.Errorf("Symbol: want MSFT, got %s", got.Symbol)
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
