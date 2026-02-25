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

# Learnings - Task 6: GetStreamerInfo for Streaming Authentication

## Implementation Details

### Method Implemented
- **GetStreamerInfo(ctx)** - Added to OAuthClient in pkg/client/oauth.go
- Endpoint: GET /trader/v1/userPreference
- Returns: *types.StreamerInfo containing authentication details for streaming services

### Key Changes
1. Added `tokenGetter TokenGetter` field to OAuthClient struct
2. Added `baseURL string` field to OAuthClient struct for configurable API base URL
3. Updated NewOAuthClient() signature to accept tokenGetter parameter
4. Implemented GetStreamerInfo() method that:
   - Calls /trader/v1/userPreference endpoint
   - Extracts streamerInfo from PreferencesResponse
   - Returns StreamerInfo struct with all authentication fields

### StreamerInfo Fields
The StreamerInfo struct contains:
- AccountID
- AccountIDType
- Token
- TokenTimestamp
- UserID
- AppID
- Secret
- AccessLevel

### Testing Approach
- Created oauth_test.go with 3 test scenarios:
  1. Successfully retrieves streamer info
  2. Handles missing streamer info
  3. Handles HTTP error (401 unauthorized)
- Used httptest.NewServer() for mocking API responses
- Created mockTokenGetter for testing token retrieval
- All 3 tests pass

### Gotchas
1. **OAuthClient needed tokenGetter**: Unlike Accounts struct, OAuthClient didn't have a tokenGetter field. Added it to enable GetStreamerInfo() to get access tokens.
2. **Hardcoded URL issue**: GetStreamerInfo() initially had hardcoded API URL. Added baseURL field to OAuthClient to make it configurable for testing.
3. **Test server URL**: Had to set oauthClient.baseURL = server.URL in tests to use the mock server instead of real API.
4. **Error handling**: When HTTP status is not OK, the response body is still decoded. Need to check if StreamerInfo is nil after decoding.

### Files Created/Modified
- `pkg/client/oauth.go` - Added GetStreamerInfo() method, added tokenGetter and baseURL fields
- `pkg/client/oauth_test.go` - New file with 3 test cases
- `.sisyphus/evidence/task-6-streamer-info.log` - Test output evidence

### Verification
- All tests pass: `go test -v ./pkg/client -run TestGetStreamerInfo`
- No regressions: All existing client tests still pass
- Build succeeds: `go build ./pkg/client`
- No vet warnings: `go vet ./pkg/client`

### Notes on Task Dependencies
- Task 6 depends on Task 8 (Preferences method) according to the plan, but Preferences() is not yet implemented
- GetStreamerInfo() calls the /trader/v1/userPreference endpoint directly instead of using a Preferences() method
- This approach works for now, but may need to be refactored when Preferences() is implemented in Task 8

# Learnings - Task 7: AccountDetailsAll REST Method

## Implementation Details

### Method Implemented
- **AccountDetailsAll(ctx, fields string)** - Added to Accounts struct in pkg/client/accounts.go
- Endpoint: GET /trader/v1/accounts/ (with trailing slash)
- Query parameter: `fields` (comma-separated options: "positions")
- Returns: *types.AccountDetailsAllResponse containing all linked accounts with balances/positions

### Key Patterns
- Follows existing LinkedAccounts() pattern from pkg/client/accounts.go
- Uses context with 30-second timeout to prevent indefinite blocking
- Constructs API URL with baseAPIURL constant
- Sets Authorization header with Bearer token from tokenGetter.GetAccessToken()
- Uses url.Values{} for query parameter building
- Appends query string to URL only if parameters exist
- Logs success/error with slog.Logger

### Implementation Steps
1. Created context with timeout (30 seconds)
2. Built API URL: `/trader/v1/accounts/` (with trailing slash as specified)
3. Added query parameter `fields` if provided
4. Set Authorization header with Bearer token
5. Made GET request via a.httpClient.Get()
6. Decoded JSON response into AccountDetailsAllResponse
7. Logged success with account count

### Testing Approach
- All existing tests pass: `go test -v ./pkg/client`
- No new tests were created for this task (tests will be added in Task 10)
- Build succeeds: `go build ./pkg/client`
- No diagnostics errors

### Gotchas
1. **Trailing slash requirement**: The endpoint must include trailing slash `/trader/v1/accounts/` (not `/trader/v1/accounts`)
2. **Query parameter handling**: Only append query string if parameters exist (len(params) > 0)
3. **Response type**: Uses AccountDetailsAllResponse (from Task 1) which has same structure as AccountDetailsResponse (both contain []Account)

### Files Created/Modified
- `pkg/client/accounts.go` - Added AccountDetailsAll() method (lines 73-108)

### Verification
- All tests pass: `go test -v ./pkg/client`
- No regressions: All existing client tests still pass
- Build succeeds: `go build ./pkg/client`
- No diagnostics errors

### Notes on Task Dependencies
- Task 7 depends on Task 1 (AccountDetailsAllResponse type) - COMPLETED
- Task 7 depends on Task 5 (Token auto-refresh) - COMPLETED
- Task 7 will be tested in Task 10 (Account method tests)

# Learnings - Task 8: Preferences REST Method

## Implementation Details

### Method Implemented
- **Preferences(ctx)** - Added to Accounts struct in pkg/client/accounts.go
- Endpoint: GET /trader/v1/userPreference (NOT plural)
- Returns: *types.PreferencesResponse containing user preferences including streamerInfo

### Key Patterns
- Follows existing AccountDetails() pattern from pkg/client/accounts.go
- Uses context with 30-second timeout to prevent indefinite blocking
- Constructs API URL with baseAPIURL constant
- Sets Authorization header with Bearer token from tokenGetter.GetAccessToken()
- Logs success/error with slog.Logger
- Returns PreferencesResponse which contains StreamerInfo field

### Implementation Steps
1. Created context with timeout (30 seconds)
2. Built API URL: `/trader/v1/userPreference` (singular, NOT plural)
3. Set Authorization header with Bearer token
4. Made GET request via a.httpClient.Get()
5. Decoded JSON response into PreferencesResponse
6. Logged success

### Testing Approach
- All existing tests pass: `go test -v ./pkg/client`
- No new tests were created for this task (tests will be added in Task 12)
- Build succeeds: `go build ./pkg/client`
- No vet warnings: `go vet ./pkg/client`
- No diagnostics errors

### Gotchas
1. **Endpoint is singular**: The endpoint is `/trader/v1/userPreference` (NOT `/trader/v1/userPreferences`)
2. **StreamerInfo access**: The PreferencesResponse type already has StreamerInfo as a field, so no array access [0] is needed in Go (unlike Python)
3. **Documentation pattern**: Followed existing docstring pattern in the file (method description + endpoint comment)

### Files Created/Modified
- `pkg/client/accounts.go` - Added Preferences() method (lines 156-182)

### Verification
- All tests pass: `go test -v ./pkg/client`
- No regressions: All existing client tests still pass
- Build succeeds: `go build ./pkg/client`
- No vet warnings: `go vet ./pkg/client`
- No diagnostics errors

### Notes on Task Dependencies
- Task 8 depends on Task 1 (PreferencesResponse type) - COMPLETED
- Task 8 depends on Task 5 (Token auto-refresh) - COMPLETED
- Task 8 will be tested in Task 12 (Account method tests)
