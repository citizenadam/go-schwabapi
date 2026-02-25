# Schwab API Full Compatibility Implementation

## TL;DR

> **Quick Summary**: Implement all 31 missing REST API methods and 14 streaming service methods to achieve 1:1 compatibility with tylerebowers/Schwabdev Python library.
>
> **Deliverables**:
> - 31 REST API methods in accounts, transactions, market, and orders modules
> - 14 streaming service methods in stream module
> - Supporting types (requests/responses) and helper functions
> - TDD test suite with coverage for all methods
>
> **Estimated Effort**: Large (6-9 days)
> **Parallel Execution**: YES - 10 waves with 5-6 parallel tasks each
> **Critical Path**: Type definitions → Foundation → REST APIs → Streaming Services → QA

---

## Context

### Original Request
User requested implementation of all missing methods to achieve 1:1 compatibility with Schwabdev Python library.

### Interview Summary
**Key Discussions**:
- **Scope**: All 31 missing methods + 14 streaming services (45 total)
- **Test Strategy**: TDD (Test-Driven Development)
- **Test Infrastructure**: Existing with stretchr/testify and httptest patterns
- **Streaming Services**: All 14 services implemented as individual methods

**Research Findings**:
- Python library uses trailing `/` in account_details_all endpoint
- Pre-market and post-market hours via MarketHours(symbols, date)
- _get_streamer_info() from Python for streaming authentication
- Token refresh pattern (auto-refresh 61s before expiry)
- Date formats: ISO8601, EPOCH, EPOCH_MS, YYYY-MM-DD

**Metis Review** (addressed):
- **Issue**: Need auto-token refresh before API calls → Added to OAuthClient implementation
- **Issue**: Streaming auto-resubscribe → Using Manager.RecordRequest() pattern
- **Guardrails**: No new dependencies, no caching, no validation, follow existing patterns
- **Defaults Applied**: Auto-refresh yes, individual methods, pagination yes, Go-idiomatic errors

---

## Work Objectives

### Core Objective
Implement 45 missing methods (31 REST + 14 streaming) to achieve 1:1 API compatibility with Schwabdev Python library using TDD.

### Concrete Deliverables
- 6 new methods in Accounts: AccountDetailsAll, Preferences, AccountOrdersAll
- 2 new methods in Transactions: Transactions, TransactionDetails
- 7 new methods in Market: Movers, MarketHours, MarketHour, Instruments, InstrumentCusip, OptionExpirationChain, PreviewOrder
- 1 method in OAuth: GetStreamerInfo
- 14 streaming service methods in pkg/stream/services.go
- 31 request/response type definitions
- Helper functions: date/time formatting, key formatting
- TDD test suite with all 45 methods

### Definition of Done
- [ ] All 45 methods implemented and passing tests
- [ ] All methods have success/error/edge case tests
- [ ] Streaming auto-resubscribe works after reconnection
- [ ] Token auto-refresh works before API calls
- [ ] NoBreaking changes to existing methods
- [ ] Documentation comments on all public APIs
- [ ] All evidence files in .sisyphus/evidence/ captured

### Must Have
- 1:1 API compatibility with Schwabdev Python library
- TDD workflow for all methods
- Auto-token refresh before API calls
- Auto-resubscribe for streaming services
- Pagination support for bulk operations
- No breaking changes to existing code

### Must NOT Have (Guardrails)
- New external dependencies
- Client-side validation (let API handle)
- Caching or optimization layers
- Builder patterns or fluent APIs
- Custom error hierarchies
- Configuration options or flags
- Pagination helper functions
- Logging abstractions
- Auto-documentation tools

---

## Verification Strategy (MANDATORY)

> **ZERO HUMAN INTERVENTION** — ALL verification is agent-executed.

### Test Decision
- **Infrastructure exists**: YES (stretchr/testify, httptest)
- **Automated tests**: YES (TDD)
- **Framework**: go test
- **If TDD**: Each task follows RED (failing test) → GREEN (minimal impl) → REFACTOR

### QA Policy
Every task MUST include agent-executed QA scenarios.

| Deliverable Type | Verification Tool | Method |
|------------------|-------------------|--------|
| REST API | Bash (go test) | Test suite with httptest mocks |
| Streaming Services | Bash (go test) + mock WebSocket | Test subscription, message handling |
| Integration | Bash (go test) | End-to-end flows |

---

## Execution Strategy

### Parallel Execution Waves

> Maximize throughput with 5-6 tasks per wave. Target 70% parallel speedup.

```
Wave 1 (Foundation — types, helpers, max parallel):
├── Task 1: Response type definitions (31 methods) [quick]
├── Task 2: Request type definitions (31 methods) [quick]
├── Task 3: Date/time helper functions [quick]
├── Task 4: Streaming key helper functions [quick]
├── Task 5: Token auto-refresh in OAuthClient [quick]
└── Task 6: GetStreamerInfo for streaming auth [quick]

Wave 2 (Account Methods — 3 methods):
├── Task 7: AccountDetailsAll REST method [unspecified-low]
├── Task 8: Preferences REST method [unspecified-low]
├── Task 9: AccountOrdersAll REST method [unspecified-low]
├── Task 10: Test suite: Account methods [quick]
├── Task 11: Test suite: Account orders [quick]
└── Task 12: Test suite: Preferences [quick]

Wave 3 (Transaction Methods — 2 methods):
├── Task 13: Transactions REST method [unspecified-low]
├── Task 14: TransactionDetails REST method [unspecified-low]
├── Task 15: Test suite: Transactions [quick]
├── Task 16: Test suite: TransactionDetails [quick]
└── Task 17: Test suite: Integration [quick]

Wave 4 (Market Data Part 1 — 4 methods):
├── Task 18: Movers REST method [unspecified-low]
├── Task 19: MarketHours REST method (plural) [unspecified-low]
├── Task 20: MarketHour REST method (singular) [unspecified-low]
├── Task 21: Test suite: Market hours [quick]
├── Task 22: Test suite: Movers [quick]
└── Task 23: Test suite: Date formatting [quick]

Wave 5 (Market Data Part 2 — 3 methods):
├── Task 24: Instruments REST method [unspecified-low]
├── Task 25: InstrumentCusip REST method [unspecified-low]
├── Task 26: OptionExpirationChain REST method [unspecified-low]
├── Task 27: Test suite: Instruments [quick]
├── Task 28: Test suite: Option chains [quick]
└── Task 29: Test suite: Error handling [quick]

Wave 6 (Order Preview — 1 method):
├── Task 30: PreviewOrder REST method [unspecified-low]
├── Task 31: Test suite: Preview order [quick]
├── Task 32: Test suite: Edge cases [quick]
├── Task 33: Test suite: Integration with Orders [quick]
└── Task 34: Test data fixtures [quick]

Wave 7 (Level One Streaming — 5 services):
├── Task 35: LevelOneEquities streaming service [unspecified-low]
├── Task 36: LevelOneOptions streaming service [unspecified-low]
├── Task 37: LevelOneFutures streaming service [unspecified-low]
├── Task 38: LevelOneFuturesOptions service [unspecified-low]
└── Task 39: LevelOneForex streaming service [unspecified-low]

Wave 8 (Book Services — 3 services):
├── Task 40: NyseBook streaming service [unspecified-low]
├── Task 41: NasdaqBook streaming service [unspecified-low]
└── Task 42: OptionsBook streaming service [unspecified-low]

Wave 9 (Chart Services — 2 services):
├── Task 43: ChartEquity streaming service [unspecified-low]
└── Task 44: ChartFutures streaming service [unspecified-low]

Wave 10 (Screener Services — 3 services):
├── Task 45: ScreenerEquity streaming service [unspecified-low]
├── Task 46: ScreenerOptions streaming service [unspecified-low]
└── Task 47: ScreenerOption streaming service [unspecified-low]

Wave 11 (Activity Service — 1 service):
├── Task 48: AccountActivity streaming service [unspecified-low]
├── Task 49: Test suite: Level One [quick]
├── Task 50: Test suite: Book services [quick]
├── Task 51: Test suite: Charts & Screeners [quick]
└── Task 52: Test suite: Auto-resubscribe [quick]

Wave 12 (Final Verification):
├── Task 53: Integration tests (all methods) [deep]
├── Task 54: Code review (check patterns) [unspecified-high]
├── Task 55: Compatibility audit (vs Python library) [deep]
├── Task 56: Git cleanup and summary [git]
└── Task 57: Final QA scenarios [unspecified-high]

Critical Path: Task 1 → Task 5 → Task 7→ Task 13 → Task 18 → Task 30 → Task 35 → Task 53 → F1-F4
Parallel Speedup: ~70% faster than sequential
Max Concurrent: 6 (Waves 1-11)
```

### Dependency Matrix (abbreviated)

| Task | Depends On | Blocks | Wave |
|------|------------|--------|------|
| 1-6 | — | 7-57 | 1 |
| 7-12 | 2, 3, 5 | 13-17, 10-12 | 2 |
| 13-17 | 2, 3, 5 | 18-29, 15-17 | 3 |
| 18-29 | 2, 3, 5 | 30-39, 21-23 | 4 |
| 30-34 | 2, 3, 5 | 35-44, 31-34 | 5 |
| 35-44 | 4, 5, 6 | 45-50, 35-44 | 6-9 |
| 45-52 | 4, 5, 6 | 53-57, 49-52 | 10-11 |
| 53-57 | 7-52 | — | 12 |

> Full matrix for all 57 tasks will be included in complete plan.

### Agent Dispatch Summary

| Wave | # Parallel | Tasks → Agent Category |
|------|------------|----------------------|
| 1 | **6** | T1-T6 → `quick` |
| 2 | **6** | T7-T9 → `unspecified-low`, T10-T12 → `quick` |
| 3 | **5** | T13-T14 → `unspecified-low`, T15-T17 → `quick` |
| 4 | **6** | T18-T20 → `unspecified-low`, T21-T23 → `quick` |
| 5 | **6** | T24-T26 → `unspecified-low`, T27-T29 → `quick` |
| 6 | **6** | T30 → `unspecified-low`, T31-T34 → `quick` |
| 7 | **5** | T35-T39 → `unspecified-low` |
| 8 | **3** | T40-T42 → `unspecified-low` |
| 9 | **2** | T43-T44 → `unspecified-low` |
| 10 | **3** | T45-T47 → `unspecified-low` |
| 11 | **5** | T48 → `unspecified-low`, T49-T52 → `quick` |
| 12 | **5** | T53, F1 → `deep`, T54, T57 → `unspecified-high`, T55 → `deep`, T56 → `git` |

---

## TODOs

> Implementation + Test = ONE Task. Every task has QA scenarios.
> **A task WITHOUT QA Scenarios is INCOMPLETE.**

---

### WAVE 1: Foundation (Types & Helpers)

- [x] 1. Response Type Definitions (31 methods)

  **What to do**:
  - Add struct definitions to `pkg/types/responses.go` for all 31 missing methods
  - Follow existing pattern: JSON tags, pointer types for optional fields
  - Types needed: AccountDetailsAllResponse, PreferencesResponse, AccountOrdersAllResponse, TransactionsResponse, TransactionDetailsResponse, MoversResponse, MarketHoursResponse, InstrumentsResponse, OptionExpirationChainResponse, PreviewOrderResponse

  **Must NOT do**:
  - Add validation methods
  - Create helper structs beyond response schemas

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Straightforward type definitions, no complex logic
  - **Skills**: None needed

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 2-6)
  - **Blocks**: 7-57 (all depend on response types)

  **References**:
  - `pkg/types/responses.go:15-100` - Existing response type patterns (MarketDataResponse, AccountDetailsResponse)
  - Use JSON tags with omitempty for optional fields
  - Use pointer types (*string, *int) for nullable fields

  **Acceptance Criteria**:
  - [ ] All 31 response types defined in pkg/types/responses.go
  - [ ] go test ./pkg/types - PASS (compile check)
  - [ ] go vet ./pkg/types - PASS

  **QA Scenarios**:

  ```
  Scenario: Type definitions compile
    Tool: Bash (go test)
    Preconditions: None
    Steps:
      1. go test ./pkg/types
      2. go vet ./pkg/types
    Expected Result: No compilation errors, no vet warnings
    Failure Indicators: compile errors showing missing types
    Evidence: .sisyphus/evidence/task-1-type-compile.log
  ```

  **Commit**: YES
  - Message: `feat(types): add response types for 31 new methods`
  - Files: pkg/types/responses.go
  - Pre-commit: go test ./pkg/types

- [x] 2. Request Type Definitions (31 methods)

  **What to do**:
  - Add struct definitions to `pkg/types/requests.go` for all 31 missing methods
  - Types needed: AccountDetailsAllRequest, GetPreferencesRequest, AccountOrdersAllRequest, TransactionsRequest, MoversRequest, MarketHoursRequest, InstrumentsRequest, OptionExpirationChainRequest, PreviewOrderRequest
  - Include query parameter tags

  **Must NOT do**:
  - Add validation logic
  - Create builder patterns

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Type definitions only, similar to Task 1
  - **Skills**: None needed

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1, 3-6)
  - **Blocks**: 7-57

  **References**:
  - `pkg/types/requests.go:1-50` - Existing request type patterns (OptionChainsRequest, PriceHistoryRequest)
  - Use `url` tag for query parameters (e.g., `url:"fields"`)

  **Acceptance Criteria**:
  - [ ] All 31 request types defined in pkg/types/requests.go
  - [ ] go test ./pkg/types - PASS

  **QA Scenarios**:

  ```
  Scenario: Request types compile
    Tool: Bash (go test)
    Preconditions: None
    Steps:
      1. go build ./pkg/types
    Expected Result: No compilation errors
    Evidence: .sisyphus/evidence/task-2-requests-compile.log
  ```

  **Commit**: YES
  - Message: `feat(types): add request types for 31 new methods`
  - Files: pkg/types/requests.go
  - Pre-commit: go test ./pkg/types

- [x] 3. Date/Time Helper Functions

  **What to do**:
  - Add helper functions to `pkg/client/helpers.go` (create if needed)
  - Implement: FormatISO8601Date, FormatEPOCH, FormatEPOCH_MS, FormatYYYYMMDD
  - Convert Go time.Time to string formats matching Python

  **Must NOT do**:
  - Add timezone handling beyond UTC
  - Create validation logic

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Utility functions, straightforward

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1

  **References**:
  - Python time_convert() from Schwabdev client.py
  - Standard Go time package: time.Now(), time.Unix(), .Format(time.RFC3339)

  **Acceptance Criteria**:
  - [ ] 4 helper functions defined in pkg/client/helpers.go
  - [ ] go test pkg/client/helpers*_test.go - PASS

  **QA Scenarios**:

  ```
  Scenario: Date formatting works correctly
    Tool: Bash (go test)
    Preconditions: None
    Steps:
      1. go test -v ./pkg/client -run TestDateHelpers
      2. Verify test compares: 2024-03-15T12:34:56Z for ISO8601
    Expected Result: All tests pass, correct formats
    Evidence: .sisyphus/evidence/task-3-date-helpers.log
  ```

  **Commit**: YES
  - Message: `feat(client): add date/time formatting helpers`
  - Files: pkg/client/helpers.go, pkg/client/helpers_test.go

- [x] 4. Streaming Key Helper Functions

  **What to do**:
  - Add key format helpers to `pkg/stream/helpers.go` (create if needed)
  - Implement: FormatEquityKey, FormatOptionKey, FormatFutureKey, FormatForexKey
  - Match Python key formats: 6-char symbols, 6-char expiry, C/P, strike codes

  **Must NOT do**:
  - Add client-side validation
  - Create key parser

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Format utilities

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1

  **References**:
  - Python schwabdev/translation.py or stream key format documentation
  - Option key example: "AAPL  240809C00095000" (symbol 6-chars, space padding, date YYMMDD, C/P, strike)

  **Acceptance Criteria**:
  - [ ] 4 key format helpers defined in pkg/stream/helpers.go
  - [ ] go test - PASS

  **QA Scenarios**:

  ```
  Scenario: Key formatting matches Python
    Tool: Bash (go test)
    Preconditions: None
    Steps:
      1. go test -v ./pkg/stream -run TestKeyFormatting
      2. Verify output against Python examples
    Expected Result: Key matches "AAPL  240809C00095000" format
    Evidence: .sisyphus/evidence/task-4-key-helpers.log
  ```

  **Commit**: YES
  - Message: `feat(stream): add key formatting helpers`
  - Files: pkg/stream/helpers.go, pkg/stream/helpers_test.go

- [x] 5. Token Auto-Refresh in OAuthClient

  **What to do**:
  - Modify `pkg/client/oauth.go:TokenGetter.GetAccessToken(ctx)`
  - Add auto-refresh logic: if token expires in <61 seconds, refresh
  - Use existing SQLite storage from current implementation

  **Must NOT do**:
  - Change existing storage logic
  - Add new dependencies

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Small change to existing method

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1

  **References**:
  - `pkg/client/oauth.go:TokenGetter` interface
  - Python tokens.py: update_tokens() threshold at 61 seconds

  **Acceptance Criteria**:
  - [ ] GetAccessToken() auto-refreshes before expiry
  - [ ] go test - PASS with mock time
  - [ ] No changes to existing RefreshToken() method

  **QA Scenarios**:

  ```
  Scenario: Token refreshes before expiry
    Tool: Bash (go test)
    Preconditions: Mock token with 60s expiry
    Steps:
      1. Call GetAccessToken(ctx)
      2. Verify RefreshToken() called
      3. Verify new token returned
    Expected Result: Token refreshed and returned
    Evidence: .sisyphus/evidence/task-5-token-refresh.log

  Scenario: No refresh for fresh token
    Tool: Bash (go test)
    Preconditions: Mock token with >61s expiry
    Steps:
      1. Call GetAccessToken(ctx)
      2. Verify RefreshToken() NOT called
    Expected Result: Returns existing token without refresh
    Evidence: .sisyphus/evidence/task-5-no-refresh.log
  ```

  **Commit**: YES
  - Message: `feat(oauth): add auto-refresh before API calls`
  - Files: pkg/client/oauth.go
  - Pre-commit: go test -v ./pkg/client

- [x] 6. GetStreamerInfo for Streaming Authentication

  **What to do**:
  - Add GetStreamerInfo(ctx) to `pkg/client/oauth.go`
  - Extract streamerInfo from preferences response (index [0] array access)
  - Return: SchwabClientChannel, SchwabClientFunctionId, SchwabClientCustomerId, SchwabClientCorrelId

  **Must NOT do**:
  - Create new struct for streamerInfo (use existing)

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Simple extraction method

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1

  **References**:
  - Python client.py: _get_streamer_info() method
  - Python endpoint: GET /trader/v1/userPreference
  - Response array access: response.json().get('streamerInfo', None)[0]

  **Acceptance Criteria**:
  - [ ] GetStreamerInfo() defined in OAuthClient
  - [ ] Returns streamerInfo from preferences[0]
  - [ ] go test - PASS

  **QA Scenarios**:

  ```
  Scenario: Extracts streamerInfo correctly
    Tool: Bash (go test)
    Preconditions: Mock preferences response
    Steps:
      1. Call GetStreamerInfo(ctx)
      2. Verify streamerInfo fields extracted
      3. Verify array index [0] used
    Expected Result: All 4 fields returned correctly
    Evidence: .sisyphus/evidence/task-6-streamer-info.log
  ```

  **Commit**: YES
  - Message: `feat(oauth): add GetStreamerInfo for streaming auth`
  - Files: pkg/client/oauth.go

---

### WAVE 2: Account Methods (3 Methods)

- [x] 7. AccountDetailsAll REST Method

  **What to do**:
  - Add AccountDetailsAll(ctx, fields string) to `pkg/client/accounts.go`
  - Endpoint: GET `/trader/v1/accounts/` (trailing slash)
  - Query parameter: `fields` (comma-separated options: "positions")
  - Returns all linked accounts with balances/positions

  **Must NOT do**:
  - Implement pagination (defer to later if needed)
  - Add validation for fields parameter

  **Recommended Agent Profile**:
  - **Category**: `unspecified-low`
    - Reason: Simple REST wrapper, follows existing pattern
  - **Skills**: None needed

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 8-12)
  - **Blocks**: 10, 21, 53

  **References**:
  - `pkg/client/accounts.go:LinkedAccounts()` for pattern
  - Schwabdev: client.py account_details_all()
  - Endpoint: `/trader/v1/accounts/` with trailing slash
  - Fields: "positions" or nil for default

  **Acceptance Criteria**:
  - [ ] AccountDetailsAll() implemented
  - [ ] Uses correct endpoint with trailing slash
  - [ ] go test - PASS (success, error, edge cases)

  **QA Scenarios**:

  ```
  Scenario: Retrieve all accounts successfully
    Tool: Bash (go test)
    Preconditions: Mock httptest server with accounts response
    Steps:
      1. client.AccountDetailsAll(ctx, "positions")
      2. Assert response contains all accounts
      3. Assert positions field included
    Expected Result: AccountDetailsAllResponse with multiple accounts
    Failure Indicators: Empty response, wrong endpoint called
    Evidence: .sisyphus/evidence/task-7-account-details-all-success.log

  Scenario: Handle authentication error
    Tool: Bash (go test)
    Preconditions: Mock server returning 401
    Steps:
      1. Call AccountDetailsAll(ctx, "positions")
      2. Verify error returned
    Expected Result: Error contains "unauthorized" or "401"
    Evidence: .sisyphus/evidence/task-7-account-details-all-error.log

  Scenario: Handle empty response
    Tool: Bash (go test)
    Preconditions: Mock server returning empty array
    Steps:
      1. Call AccountDetailsAll(ctx, "")
      2. Verify empty response returned gracefully
    Expected Result: Empty slice, not error
    Evidence: .sisyphus/evidence/task-7-account-details-all-edge.log
  ```

  **Commit**: YES (group with Tasks 8-9)
  - Message: `feat(accounts): add AccountDetailsAll, Preferences, AccountOrdersAll`
  - Files: pkg/client/accounts.go
  - Pre-commit: go test -v ./pkg/client

**Commit**: YES (group with Tasks 8-9)
  - Message: `feat(accounts): add AccountDetailsAll, Preferences, AccountOrdersAll`
  - Files: pkg/client/accounts.go
  - Pre-commit: go test -v ./pkg/client

- [x] 8. Preferences REST Method

  **What to do**:
  - Add Preferences(ctx) to `pkg/client/accounts.go`
  - Endpoint: GET `/trader/v1/userPreference` (NOT plural)
  - Use GetStreamerInfo() to extract streamerInfo

  **Must NOT do**:
  - Duplicate streamerInfo extraction logic

  **Recommended Agent Profile**:
  - **Category**: `unspecified-low`
  - **Parallel Group**: Wave 2
  - **Blocks**: 11

  **References**:
  - Task 6: GetStreamerInfo() for streamerInfo extraction
  - Schwabdev: client.py preferences()

  **Acceptance Criteria**:
  - [ ] Preferences() implemented
  - [ ] Uses correct userPreference endpoint
  - [ ] go test - PASS

  **QA Scenarios**:

  ```
  Scenario: Retrieves preferences successfully
    Tool: Bash (go test)
    Preconditions: Mock server with preferences response
    Steps:
      1. client.Preferences(ctx)
      2. Verify streamerInfo[0] access
    Expected Result: PreferencesResponse with streamerInfo
    Evidence: .sisyphus/evidence/task-8-preferences-success.log
  ```

- [x] 9. AccountOrdersAll REST Method

  **What to do**:
  - Add AccountOrdersAll(ctx, from, to, maxResults, status) to `pkg/client/accounts.go`
  - Endpoint: GET `/trader/v1/orders`
  - Query params: fromEnteredTime, toEnteredTime, maxResults, status

  **Must NOT do**:
  - Implement pagination

  **Recommended Agent Profile**:
  - **Category**: `unspecified-low`
  - **Parallel Group**: Wave 2

  **References**:
  - `pkg/client/orders.go:AccountOrders()` for pattern
  - Schwabdev: client.py account_orders_all()

  **Acceptance Criteria**:
  - [ ] AccountOrdersAll() implemented
  - [ ] Query params formatted correctly
  - [ ] go test - PASS

  **QA Scenarios**:

  ```
  Scenario: Retrieves all orders successfully
    Tool: Bash (go test)
    Preconditions: Mock server with orders response
    Steps:
      1. client.AccountOrdersAll(ctx, "2024-01-01", "2024-03-15", 100, "FILLED")
      2. Verify date formatting in params
    Expected Result: AccountOrdersAllResponse
    Evidence: .sisyphus/evidence/task-9-orders-all-success.log
  ```

---

### NOTE: Tasks 10-57 follow identical patterns

**Standard Template for Remaining Tasks** (scheduling consideration):

For efficiency, tasks 8-57 follow the same template as Tasks 1-7:
- Each REST method: Implementation + 3 test scenarios (success/error/edge)
- Each streaming service: Implementation + subscription/reconnection tests
- Pattern references from existing codebase
- Agent profiles: `unspecified-low` for methods, `quick` for tests

**Tasks 10-12**: Account method tests
**Tasks 13-14**: Transactions methods (Transactions, TransactionDetails)
**Tasks 15-17**: Transaction tests
**Tasks 18-20**: Market Data Part 1 (Movers, MarketHours, MarketHour)
**Tasks 21-23**: Market Data tests
**Tasks 24-26**: Market Data Part 2 (Instruments, InstrumentCusip, OptionExpirationChain)
**Tasks 27-29**: Instrument tests
**Tasks 30-34**: Order Preview (PreviewOrder + tests)
**Tasks 35-44**: Level One + Book + Chart streaming services
**Tasks 45-47**: Screener streaming services
**Tasks 48-52**: Activity service + streaming tests
**Tasks 53-57**: Final Verification Wave

Each task structure mirrors Tasks 1-7 with:
- Dependencies blocked/waited correctly
- QA scenarios with specific steps
- Evidence capture paths
- Commit grouping patterns

---

## Final Verification Wave (MANDATORY — after ALL implementation tasks)

- [x] 53. Integration Tests (All Methods)

  **What to do**:
  - Create integration tests in `pkg/integration_test.go`
  - Test workflows: Account → Order → Transaction paths
  - Test streaming: Subscribe → Reconnect → Auto-resubscribe

  **Acceptance Criteria**:
  - [ ] All 45 methods work together
  - [ ] Streaming auto-resubscribes after reconnect
  - [ ] Token refresh works before API calls

  **QA Scenarios**:

  ```
  Scenario: Full account workflow
    Tool: Bash (go test)
    Preconditions: Mock server responses
    Steps:
      1. AccountDetailsAll() → Get all accounts
      2. Preferences() → Get streamerInfo
      3. PlaceOrder() → Create order
      4. OrderDetails() → Verify order created
      5. Transactions() → Check transaction history
    Expected Result: All operations succeed in sequence
    Evidence: .sisyphus/evidence/task-53-full-workflow.log
  ```

- [ ] F1. Plan Compliance Audit — `oracle`
  Read plan end-to-end. Verify all 45 methods implemented. Check evidence files exist.
  Output: `Methods [45/45] | Evidence [N/N] | VERDICT: APPROVE/REJECT`

- [ ] F2. Code Quality Review — `unspecified-high`
  Run `tsc --noEmit` + linter + `go test`. Check all changes for `as any`/`@ts-ignore`, empty catches, console.log.
  Output: `Build [PASS/FAIL] | Lint [PASS/FAIL] | Tests [N pass/N fail] | VERDICT`

- [ ] F3. Real Manual QA — `unspecified-high`
  Execute EVERY QA scenario from ALL tasks. Test edge cases and error paths.
  Output: `Scenarios [N/N pass] | Edge Cases [N/N] | VERDICT`

- [ ] F4. Scope Fidelity Check — `deep`
  Verify everything in spec was built (no missing). Verify nothing beyond spec (no creep).
  Output: `Methods [45/45] | Unaccounted [CLEAN] | VERDICT`

---

## Success Criteria

### Verification Commands
```bash
# Build
go build ./...

# Test all
go test -v ./...

# Check all methods
grep -r "func.*AccountDetailsAll\|func.*Preferences\|func.*Transactions" ./pkg/client

# Run integration
go test -v ./pkg -run TestIntegration

# Check streaming services
grep -r "func.*LevelOne\|func.*NyseBook\|func.*Chart" ./pkg/stream
```

### Final Checklist
- [ ] All 45 methods implemented
- [ ] All 31 REST API methods work
- [ ] All 14 streaming services work
- [ ] Token auto-refresh works
- [ ] Streaming auto-resubscribe works
- [ ] TDD tests pass for all methods
- [ ] No breaking changes to existing code
- [ ] Evidence files captured
- [ ] Documentation comments present
- [ ] Go build passes

---

## Commit Strategy

| Group | Message | Tasks |
|-------|---------|-------|
| 1 | `feat: add response/request types for 31 new methods` | 1-2 |
| 2 | `feat(client): add date/time and key formatting helpers` | 3-4 |
| 3 | `feat(oauth): add token auto-refresh and streamer info` | 5-6 |
| 4 | `feat(accounts): add AccountDetailsAll, Preferences, OrdersAll` | 7-9 |
| 5 | `feat(accounts): add transaction methods` | 13-14 |
| 6 | `feat(market): add market data getters (part 1)` | 18-20 |
| 7 | `feat(market): add instruments and option chains` | 24-26 |
| 8 | `feat(orders): add preview order` | 30 |
| 9 | `feat(stream): add level one streaming services` | 35-39 |
| 10 | `feat(stream): add book and chart services` | 40-44 |
| 11 | `feat(stream): add screener and activity services` | 45-48 |
| 12 | `test: add integration tests and QA` | 53, F1-F4 |
| 13 | `chore: final cleanup and documentation` | 54-57 |

---

**Generated**: 2025-02-24
**User**: Accepted all recommended defaults
**Scope**: 31 REST methods + 14 streaming services
**Test Strategy**: TDD with stretchr/testify
**Estimated time**: 6-9 days