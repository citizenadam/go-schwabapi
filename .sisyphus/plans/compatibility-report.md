# Schwab API Library Compatibility Report

## Executive Summary

**Reference Implementation**: [tylerebowers/Schwabdev](https://github.com/tylerebowers/Schwabdev) (Python)
**Target Implementation**: [citizenadam/go-schwabapi](https://github.com/citizenadam/go-schwabapi) (Go)

**Overall Compatibility Status**: **PARTIAL COMPATIBILITY (~40%)**

The Go implementation covers core functionality but is missing significant portions of the Python API. Implementation is ongoing.

---

## Detailed Comparison

### Client Methods Comparison

#### Account Methods

| Python Method | Go Method | Status | Signature Match |
|--------------|-----------|--------|-----------------|
| `linked_accounts()` | `LinkedAccounts(ctx)` | ✅ DONE | Compatible (context added) |
| `account_details_all(fields)` | ❌ MISSING | 🔴 N/A | Not implemented |
| `account_details(accountHash, fields)` | `AccountDetails(ctx, accountHash, fields)` | ✅ DONE | Compatible |
| `account_orders(accountHash, from, to, maxResults, status)` | `AccountOrders(ctx, accountHash, from, to, maxResults, status)` | ✅ DONE | Compatible |
| `preferences()` | ❌ MISSING | 🔴 N/A | Not implemented |

**Coverage: 4/6 (67%)**

#### Order Methods

| Python Method | Go Method | Status | Signature Match |
|--------------|-----------|--------|-----------------|
| `place_order(accountHash, order)` | `PlaceOrder(ctx, accountHash, order)` | ✅ DONE | Compatible |
| `order_details(accountHash, orderId)` | `OrderDetails(ctx, accountHash, orderId)` | ✅ DONE | Compatible |
| `cancel_order(accountHash, orderId)` | `CancelOrder(ctx, accountHash, orderId)` | ✅ DONE | Compatible |
| `replace_order(accountHash, orderId, order)` | `ReplaceOrder(ctx, accountHash, orderId, order)` | ✅ DONE | Compatible |
| `account_orders_all(from, to, max, status)` | ❌ MISSING | 🔴 N/A | Not implemented |
| `preview_order(accountHash, order)` | ❌ MISSING | 🔴 N/A | Not implemented |

**Coverage: 4/6 (67%)**

#### Transaction Methods

| Python Method | Go Method | Status | Signature Match |
|--------------|-----------|--------|-----------------|
| `transactions(accountHash, start, end, types, symbol)` | ❌ MISSING | 🔴 N/A | Not implemented |
| `transaction_details(accountHash, transactionId)` | ❌ MISSING | 🔴 N/A | Not implemented |

**Coverage: 0/2 (0%)**

#### Market Data Methods

| Python Method | Go Method | Status | Signature Match |
|--------------|-----------|--------|-----------------|
| `quotes(symbols, fields, indicative)` | `Quotes(ctx, symbols, fields, indicative)` | ✅ DONE | Compatible |
| `quote(symbol_id, fields)` | `Quote(ctx, symbol, fields)` | ✅ DONE | Compatible |
| `option_chains(symbol, ...)` | `OptionChains(ctx, req)` | ✅ DONE | Struct-based (compatible) |
| `option_expiration_chain(symbol)` | ❌ MISSING | 🔴 N/A | Not implemented |
| `price_history(symbol, ...)` | `PriceHistory(ctx, req)` | ✅ DONE | Struct-based (compatible) |
| `movers(symbol, sort, frequency)` | ❌ MISSING | 🔴 N/A | Not implemented |
| `market_hours(symbols, date)` | ❌ MISSING | 🔴 N/A | Not implemented |
| `market_hour(market_id, date)` | ❌ MISSING | 🔴 N/A | Not implemented |
| `instruments(symbols, projection)` | ❌ MISSING | 🔴 N/A | Not implemented |
| `instrument_cusip(cusip_id)` | ❌ MISSING | 🔴 N/A | Not implemented |

**Coverage: 3/10 (30%)**

#### OAuth/Token Methods

| Python Method | Go Method | Status | Signature Match |
|--------------|-----------|--------|-----------------|
| `_get_streamer_info()` | ❌ MISSING | 🔴 N/A | Not implemented (private method) |
| `update_tokens(force_access, force_refresh)` | ❌ MISSING | 🔴 N/A | Handled internally in Go |
| `Authorize(ctx)` | `Authorize(ctx)` | ✅ DONE | **Different signature** (Python has noAuthorize method) |
| `RefreshToken(ctx, refreshToken)` | `RefreshToken(ctx, refreshToken)` | ✅ DONE | Compatible |
| `RevokeToken(ctx, token, tokenType)` | `RevokeToken(ctx, token, tokenType)` | ✅ DONE | Compatible |
| `ExchangeCode(ctx, code)` | `ExchangeCode(ctx, code)` | ✅ DONE | Compatible |

**Note**: Python uses a `Tokens` class for token management, while Go uses `OAuthClient` with explicit methods.

**Coverage: 4/6 (67%)**

### Streaming Methods Comparison

#### Stream Base Methods

| Python Method | Go Method | Status | Signature Match |
|--------------|-----------|--------|-----------------|
| `basic_request(service, command, parameters)` | ❌ MISSING | 🔴 N/A | Go uses different approach |
| `_run_streamer(receiver, ping_timeout)` | ❌ MISSING | 🔴 N/A | Go uses different architecture |
| `_wait_for_backoff()` | ❌ MISSING | 🔴 N/A | Internal method |
| `_record_request(request)` | `RecordRequest(ctx, req)` | ✅ DONE | Compatible |
| `_list_to_string(ls)` | ❌ MISSING | 🔴 N/A | Not needed in Go |
| `start(receiver, daemon, ...)` | `Connect(ctx, url)` + `start()` logic | 🔶 DONE | **Different architecture** |
| `stop(clear_subscriptions)` | `Stop()` | ✅ DONE | Simpler in Go |
| `send(requests, record)` | `Write(data, record)` | ✅ DONE | Compatible concept |

#### Stream Service Methods

| Python Service | Go Equivalent | Status | Notes |
|----------------|--------------|--------|-------|
| `level_one_equities()` | ❌ MISSING | 🔴 N/A | |
| `level_one_options()` | ❌ MISSING | 🔴 N/A | |
| `level_one_futures()` | ❌ MISSING | 🔴 N/A | |
| `level_one_futures_options()` | ❌ MISSING | 🔴 N/A | |
| `level_one_forex()` | ❌ MISSING | 🔴 N/A | |
| `nyse_book()` | ❌ MISSING | 🔴 N/A | |
| `nasdaq_book()` | ❌ MISSING | 🔴 N/A | |
| `options_book()` | ❌ MISSING | 🔴 N/A | |
| `chart_equity()` | ❌ MISSING | 🔴 N/A | |
| `chart_futures()` | ❌ MISSING | 🔴 N/A | |
| `screener_equity()` | ❌ MISSING | 🔴 N/A | |
| `screener_options()` | ❌ MISSING | 🔴 N/A | |
| `screener_option()` | ❌ MISSING | 🔴 N/A | |
| `account_activity()` | ❌ MISSING | 🔴 N/A | |

**Coverage: 0/14 (0%)** - Streaming services not implemented

#### Streaming Infrastructure

| Python Feature | Go Feature | Status |
|----------------|------------|--------|
| Streaming WebSocket client | `Client` with `Conn` | ✅ DONE |
| Automatic reconnection | `Client:Connect()` with backoff | ✅ DONE |
| Subscription management | `Manager` with `RecordRequest()` | ✅ DONE |
| Message handling | `Handler` with `ParseMessage()` | ✅ DONE |
| Auto-start on market hours | ❌ MISSING | 🔴 Not implemented |
| Login/Logout commands | `Handler` supports LOGIN/LOGOUT | ✅ DONE |
| Subscription commands (ADD, SUBS, UNSUBS, VIEW) | `Handler` supports all | ✅ DONE |

**Streaming Infrastructure Coverage: 5/7 (71%)**
**Streaming Service Coverage: 0/14 (0%)**
**Overall Streaming: Minimal**

---

## Key Differences

### 1. Architecture Differences

| Aspect | Python | Go |
|--------|--------|-----|
| Client Interface | Single `Client` class with all methods | Separate structs: `Accounts`, `OrdersClient`, `Market`, `OAuthClient` |
| Context Usage | No context | All methods take `context.Context` for cancellation/timeout |
| Async Support | Separate `ClientAsync` class | Go uses goroutines (native concurrency) |
| Request Parameters | Keyword arguments in function signatures | Struct-based parameters (e.g., `OptionChainsRequest`) |
| Response Parsing | Automatic based on content-type | Explicit `DecodeJSON()` calls |
| Error Handling | Exceptions | Error returns |

### 2. Missing Functionality

**Critical Missing Methods:**
1. `account_details_all()` - Bulk account details
2. `account_orders_all()` - Bulk order retrieval
3. `preview_order()` - Order preview
4. `transactions()` and `transaction_details()` - Transaction history
5. `preferences()` - User preferences with streamer info
6. All streaming service methods (level_one_equities, etc.)
7. All market methods except quotes: movers, market_hours, instruments, etc.
8. `option_expiration_chain()`

### 3. Architectural Divergences

| Python Pattern | Go Pattern | Impact |
|----------------|-----------|--------|
| One client with all methods | Multiple specialized clients | More code, better organization |
| Direct method calls | Context on every method | More thread-safety control |
| Dict-based method params | Struct-based params | Type safety vs flexibility |
| No context pattern | Context pattern everywhere | Better control in Go |
| Manual streaming via threads | Native goroutines | Go idiomatic streaming |

---

## Compatibility Matrix

### By Category

| Category | Total | Implemented | Missing | Coverage |
|----------|-------|-------------|---------|----------|
| Accounts | 6 | 4 | 2 | 67% |
| Orders | 6 | 4 | 2 | 67% |
| Transactions | 2 | 0 | 2 | 0% |
| Market Data | 10 | 3 | 7 | 30% |
| OAuth | 6 | 4 | 2 | 67% |
| Streaming - Infrastructure | 7 | 5 | 2 | 71% |
| Streaming - Services | 14 | 0 | 14 | 0% |
| **Total** | **51** | **20** | **31** | **39%** |

### By Priority

| Priority | Methods | Status |
|----------|---------|--------|
| **Critical** | place_order, cancel_order, replace_order, quote, quotes | ✅ DONE |
| **High** | account_details, account_orders, order_details, option_chains, price_history | ✅ DONE |
| **Medium** | account_details_all, transactions, instruments, movers, market_hours | ❌ MISSING |
| **Low** | preview_order, streaming services, option_expiration_chain | ❌ MISSING |

---

## Recommendations

### 1. Short-term (Priority: HIGH)

1. **Implement missing critical account methods**:
   - `AccountDetailsAll(fields)` combines all accounts
   - `Preferences()` for streamer info

2. **Implement bulk order operations**:
   - `AccountOrdersAll(from, to, max, status)`

3. **Implement transaction history**:
   - `Transactions(accountHash, start, end, types, symbol)`
   - `TransactionDetails(accountHash, transactionId)`

### 2. Medium-term (Priority: MEDIUM)

1. **Complete market data coverage**:
   - `Movers(symbol, sort, frequency)`
   - `MarketHours(symbols, date)`
   - `Instruments(symbols, projection)`
   - `OptionExpirationChain(symbol)`
   - `PreviewOrder(accountHash, order)`

2. **Streaming service implementations**:
   - Implement at least: `level_one_equities`, `level_one_options`
   - Consider implementing others based on demand

### 3. Long-term (Priority: LOW)

1. **Streaming auto-start**:
   - Implement market-hours-based auto-start

2. **Streaming service methods**:
   - All 14 streaming service methods (NYSE_BOOK, NASDAQ_BOOK, etc.)

---

## Conclusion

The Go implementation has solid foundations with **~40% compatibility** with the Python reference. Core ordering, account details, and basic market data functionality are implemented and working. However:

**Strengths:**
- ✅ Core ordering operations complete
- ✅ Account details and orders for individual accounts working
- ✅ Basic market data (quotes) implemented
- ✅ Streaming infrastructure (connections, reconnection, subscriptions) solid

**Weaknesses:**
- ❌ Transaction history completely missing
- ❌ Bulk operations missing (account_details_all, account_orders_all)
- ❌ Most market data endpoints missing (movers, instruments)
- ❌ All streaming service methods missing (level_one, books, screeners)
- ❌ Market hours and expiration chains missing

**Verdict**: The implementation is **NOT 1:1 compatible**. It's approximately **40% complete** with good coverage of core functionality but missing significant portions of the API. This is a work-in-progress implementation.

---

**Generated**: 2025-02-24
**Reference**: https://github.com/tylerebowers/Schwabdev
**Target**: citizenadam/go-schwabapi