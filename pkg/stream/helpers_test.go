package stream

import (
	"testing"
)

func TestFormatEquityKey(t *testing.T) {
	tests := []struct {
		name     string
		symbol   string
		expected string
	}{
		{
			name:     "4-char symbol",
			symbol:   "AAPL",
			expected: "AAPL  ",
		},
		{
			name:     "5-char symbol",
			symbol:   "GOOGL",
			expected: "GOOGL ",
		},
		{
			name:     "6-char symbol",
			symbol:   "SPY500",
			expected: "SPY500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatEquityKey(tt.symbol)
			if result != tt.expected {
				t.Errorf("FormatEquityKey(%q) = %q, want %q", tt.symbol, result, tt.expected)
			}
		})
	}
}

func TestFormatOptionKey(t *testing.T) {
	tests := []struct {
		name     string
		symbol   string
		expiry   string
		callPut  string
		strike   float64
		expected string
	}{
		{
			name:     "AAPL Aug 9 2024 $95 Call",
			symbol:   "AAPL",
			expiry:   "240809",
			callPut:  "C",
			strike:   95.00,
			expected: "AAPL  240809C00095000",
		},
		{
			name:     "AAPL Aug 9 2024 $95 Put",
			symbol:   "AAPL",
			expiry:   "240809",
			callPut:  "P",
			strike:   95.00,
			expected: "AAPL  240809P00095000",
		},
		{
			name:     "TSLA Dec 20 2024 $200 Call",
			symbol:   "TSLA",
			expiry:   "241220",
			callPut:  "C",
			strike:   200.00,
			expected: "TSLA  241220C00200000",
		},
		{
			name:     "SPY Jan 17 2025 $500 Call",
			symbol:   "SPY",
			expiry:   "250117",
			callPut:  "C",
			strike:   500.00,
			expected: "SPY   250117C00500000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatOptionKey(tt.symbol, tt.expiry, tt.callPut, tt.strike)
			if result != tt.expected {
				t.Errorf("FormatOptionKey(%q, %q, %q, %.2f) = %q, want %q",
					tt.symbol, tt.expiry, tt.callPut, tt.strike, result, tt.expected)
			}
		})
	}
}

func TestFormatFutureKey(t *testing.T) {
	tests := []struct {
		name      string
		symbol    string
		monthCode string
		yearCode  string
		expected  string
	}{
		{
			name:      "E-mini S&P 500 March 2024",
			symbol:    "ES",
			monthCode: "H",
			yearCode:  "24",
			expected:  "ESH24",
		},
		{
			name:      "Crude Oil June 2024",
			symbol:    "CL",
			monthCode: "M",
			yearCode:  "24",
			expected:  "CLM24",
		},
		{
			name:      "Gold December 2024",
			symbol:    "GC",
			monthCode: "Z",
			yearCode:  "24",
			expected:  "GCZ24",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatFutureKey(tt.symbol, tt.monthCode, tt.yearCode)
			if result != tt.expected {
				t.Errorf("FormatFutureKey(%q, %q, %q) = %q, want %q",
					tt.symbol, tt.monthCode, tt.yearCode, result, tt.expected)
			}
		})
	}
}

func TestFormatForexKey(t *testing.T) {
	tests := []struct {
		name     string
		base     string
		quote    string
		expected string
	}{
		{
			name:     "EUR/USD",
			base:     "EUR",
			quote:    "USD",
			expected: "EUR/USD",
		},
		{
			name:     "GBP/USD",
			base:     "GBP",
			quote:    "USD",
			expected: "GBP/USD",
		},
		{
			name:     "USD/JPY",
			base:     "USD",
			quote:    "JPY",
			expected: "USD/JPY",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatForexKey(tt.base, tt.quote)
			if result != tt.expected {
				t.Errorf("FormatForexKey(%q, %q) = %q, want %q",
					tt.base, tt.quote, result, tt.expected)
			}
		})
	}
}
