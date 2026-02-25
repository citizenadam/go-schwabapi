# Learnings - Task 3: Date/Time Helper Functions

## Implementation Details

### Functions Implemented
1. **FormatISO8601Date** - Formats time.Time to RFC3339 format (e.g., "2024-03-15T12:34:56Z")
2. **FormatEPOCH** - Formats time.Time to Unix timestamp in seconds (e.g., "1710506096")
3. **FormatEPOCH_MS** - Formats time.Time to Unix timestamp in milliseconds (e.g., "1710506096000")
4. **FormatYYYYMMDD** - Formats time.Time to YYYY-MM-DD format (e.g., "2024-03-15")

### Key Patterns
- Used `time.RFC3339` for ISO8601 format (standard Go constant)
- Used `strconv.FormatInt(t.Unix(), 10)` for epoch timestamps
- Used `t.Format("2006-01-02")` for custom date format (Go's reference time: Mon Jan 2 15:04:05 MST 2006)

### Testing Approach
- Fixed test time for reproducibility: `time.Date(2024, 3, 15, 12, 34, 56, 0, time.UTC)`
- Tests verify exact output matches expected format
- Additional tests verify current time formatting works correctly
- All 8 tests pass (4 fixed-time tests + 4 current-time tests)

### Gotchas
- Initial test expectations were wrong (1710502496 vs actual 1710506096)
- Always verify Unix timestamps with actual calculation before writing tests
- Use `strconv.FormatInt` instead of `string(rune())` for integer-to-string conversion

### Files Created/Modified
- `pkg/client/helpers.go` - New file with 4 helper functions
- `pkg/client/helpers_test.go` - New file with 8 test cases
- `.sisyphus/evidence/task-3-date-helpers.log` - Test output evidence

### Verification
- All tests pass: `go test -v ./pkg/client -run "Format"`
- No regressions: All existing client tests still pass
- Build succeeds: `go build ./pkg/client`
---

# Learnings - Task 4: Streaming Key Helper Functions

## Implementation Details

### Functions Implemented
1. **FormatEquityKey** - Formats equity symbols with 6-char space padding (e.g., "AAPL  ")
2. **FormatOptionKey** - Formats option keys: SYMBOL(6) + EXPIRY(YYMMDD) + C/P + STRIKE(8 digits)
3. **FormatFutureKey** - Formats futures keys: SYMBOL + MONTH_CODE + YEAR_CODE
4. **FormatForexKey** - Formats forex keys: BASE/QUOTE (e.g., "EUR/USD")

### Key Patterns
- Option key format: `SYMBOL(6 chars, space-padded) + EXPIRY(YYMMDD) + C/P + STRIKE(8 digits)`
- Example: "AAPL  240809C00095000" for AAPL Aug 9 2024 $95 Call
- Strike formatting: `strike * 1000` then format as 8 digits with leading zeros
- Equity padding: `fmt.Sprintf("%-6s", symbol)` for left-aligned 6-char padding

### Testing Approach
- Test cases cover all 4 key format functions
- Option key tests verify exact format matching Python schwabdev library
- Future key tests use standard futures month codes (H=March, M=June, Z=December)
- All 13 tests pass (3 equity + 4 option + 3 future + 3 forex)

### Gotchas
- Initial test expectations for strike formatting were wrong
- Strike format is `strike * 1000` (not `strike * 100`) to get 8-digit format
- Example: $95.00 → 95000 → "00095000"
- Example: $200.00 → 200000 → "00200000"
- Example: $500.00 → 500000 → "00500000"

### Files Created/Modified
- `pkg/stream/helpers.go` - New file with 4 key format helpers
- `pkg/stream/helpers_test.go` - New file with 13 test cases
- `.sisyphus/evidence/task-4-key-helpers.log` - Test output evidence

### Verification
- All tests pass: `go test -v ./pkg/stream`
- No regressions: All existing stream tests still pass
- Build succeeds: `go build ./pkg/stream`
