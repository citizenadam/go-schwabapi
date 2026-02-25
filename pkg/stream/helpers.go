package stream

import "fmt"

// FormatEquityKey formats a streaming key for equity symbols.
// The symbol is padded to 6 characters with spaces.
func FormatEquityKey(symbol string) string {
	// Pad symbol to 6 characters with spaces
	return fmt.Sprintf("%-6s", symbol)
}

// FormatOptionKey formats a streaming key for option symbols.
// Format: SYMBOL(6 chars, space-padded) + EXPIRY(YYMMDD) + C/P + STRIKE(8 digits)
// Example: "AAPL  240809C00095000" for AAPL Aug 9 2024 $95 Call
func FormatOptionKey(symbol string, expiry string, callPut string, strike float64) string {
	// Pad symbol to 6 characters with spaces
	paddedSymbol := fmt.Sprintf("%-6s", symbol)
	// Format strike as 8 digits (5 before decimal, 3 after)
	strikeStr := fmt.Sprintf("%08d", int(strike*1000))
	return paddedSymbol + expiry + callPut + strikeStr
}

// FormatFutureKey formats a streaming key for futures symbols.
// Format: SYMBOL + MONTH_CODE + YEAR_CODE
// Example: "ESH24" for E-mini S&P 500 March 2024
func FormatFutureKey(symbol string, monthCode string, yearCode string) string {
	return symbol + monthCode + yearCode
}

// FormatForexKey formats a streaming key for forex symbols.
// Format: BASE/QUOTE (e.g., "EUR/USD")
func FormatForexKey(base string, quote string) string {
	return base + "/" + quote
}
