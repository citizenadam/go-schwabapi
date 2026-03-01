# go-schwabapi

A Go client library for the Charles Schwab API. This library provides a simple and idiomatic way to interact with Schwab's trading and market data APIs.

## Installation

```bash
go get github.com/citizenadam/go-schwabapi
```

## Prerequisites

Before using this library, you need to:

1. Register a developer application with Schwab at [developer.schwab.com](https://developer.schwab.com)
2. Obtain your application credentials:
   - `APP_KEY` - Your application's client ID
   - `APP_SECRET` - Your application's client secret
   - `CALLBACK_URL` - Your registered callback URL (e.g., `https://127.0.0.1`)

## OAuth Authentication

The Schwab API uses OAuth 2.0 for authentication. The flow consists of:

1. **Authorization** - Redirect user to Schwab's login page
2. **Callback** - User authorizes and is redirected back with an authorization code
3. **Token Exchange** - Exchange the authorization code for access and refresh tokens
4. **Token Refresh** - Automatically refresh tokens before they expire

### Environment Setup

Set your credentials as environment variables, preferrable using a secret manager:

```bash
export APP_KEY=your_app_key
export APP_SECRET=your_app_secret
export CALLBACK_URL=https://127.0.0.1
```

### Basic OAuth Flow

```go
package main

import (
    "context"
    "fmt"
    "log/slog"
    "net/http"
    "os"

    "github.com/citizenadam/go-schwabapi/pkg/client"
    "github.com/citizenadam/go-schwabapi/pkg/token"
)

func main() {
    ctx := context.Background()
    logger := slog.Default()

    // Initialize HTTP client
    httpClient := client.NewClient(logger)

    // Initialize token manager (persists tokens to SQLite)
    tokenManager, err := token.NewManager("tokens.db", logger)
    if err != nil {
        slog.Error("Failed to create token manager", "error", err)
        os.Exit(1)
    }
    defer tokenManager.Close()

    // Create OAuth client
    oauthClient := client.NewOAuthClient(
        httpClient,
        logger,
        os.Getenv("APP_KEY"),
        os.Getenv("APP_SECRET"),
        os.Getenv("CALLBACK_URL"),
        tokenManager,
    )

    // Step 1: Generate authorization URL
    authURL, err := oauthClient.Authorize(ctx)
    if err != nil {
        logger.Error("Failed to generate auth URL", "error", err)
        os.Exit(1)
    }

    fmt.Println("Visit this URL to authorize:", authURL)

    // Step 2: Start a local server to receive the callback
    // The callback will contain an authorization code
    // Example: https://127.0.0.1?code=AUTHORIZATION_CODE

    // Step 3: Exchange authorization code for tokens
    // (This would typically be done in your callback handler)
    authCode := "authorization_code_from_callback"
    token, err := oauthClient.Exchange(ctx, authCode)
    if err != nil {
        logger.Error("Failed to exchange code for token", "error", err)
        os.Exit(1)
    }

    fmt.Printf("Access Token: %s\n", token.AccessToken)
    fmt.Printf("Refresh Token: %s\n", token.RefreshToken)

    // Step 4: Refresh tokens when needed
    newToken, err := oauthClient.RefreshToken(ctx, token.RefreshToken)
    if err != nil {
        logger.Error("Failed to refresh token", "error", err)
        os.Exit(1)
    }

    fmt.Printf("New Access Token: %s\n", newToken.AccessToken)
}
```

## Example: Retrieve Account Information

This example shows how to retrieve all linked accounts and their details:

```go
package main

import (
    "context"
    "fmt"
    "log/slog"
    "os"

    "github.com/citizenadam/go-schwabapi/pkg/client"
    "github.com/citizenadam/go-schwabapi/pkg/token"
)

// TokenProvider implements client.TokenGetter interface
type TokenProvider struct {
    accessToken string
}

func (t *TokenProvider) GetAccessToken() string {
    return t.accessToken
}

func main() {
    ctx := context.Background()
    logger := slog.Default()

    // Initialize HTTP client
    httpClient := client.NewClient(logger)

    // Initialize token manager
    tokenManager, err := token.NewManager("tokens.db", logger)
    if err != nil {
        logger.Error("Failed to create token manager", "error", err)
        os.Exit(1)
    }
    defer tokenManager.Close()

    // Create token provider
    tokenProvider := &TokenProvider{
        accessToken: tokenManager.GetAccessToken(),
    }

    // Create Accounts client
    accountsClient := client.NewAccounts(httpClient, logger, tokenProvider)

    // Get all linked accounts
    linkedAccounts, err := accountsClient.LinkedAccounts(ctx)
    if err != nil {
        logger.Error("Failed to get linked accounts", "error", err)
        os.Exit(1)
    }

    fmt.Println("Linked Accounts:")
    for _, acct := range linkedAccounts.AccountNumbers {
        fmt.Printf("  Account Number: %s, Hash: %s\n", acct.AccountNumber, acct.AccountHash)
    }

    // Get detailed account information for all accounts
    accountDetails, err := accountsClient.AccountDetailsAll(ctx, "")
    if err != nil {
        logger.Error("Failed to get account details", "error", err)
        os.Exit(1)
    }

    fmt.Println("\nAccount Details:")
    for _, acct := range accountDetails.Accounts {
        fmt.Printf("  Account: %s (%s)\n", acct.AccountNumber, acct.AccountType)
        fmt.Printf("  Status: %s\n", acct.AccountStatus)
        fmt.Printf("  Hash: %s\n", acct.AccountHash)
    }
}
```

## Example: Retrieve Stock Quote (AAPL)

This example shows how to retrieve a stock quote for Apple (AAPL):

```go
package main

import (
    "context"
    "fmt"
    "log/slog"
    "os"

    "github.com/citizenadam/go-schwabapi/pkg/client"
    "github.com/citizenadam/go-schwabapi/pkg/token"
)

// TokenProvider implements client.TokenGetter interface
type TokenProvider struct {
    accessToken string
}

func (t *TokenProvider) GetAccessToken() string {
    return t.accessToken
}

func main() {
    ctx := context.Background()
    logger := slog.Default()

    // Initialize HTTP client
    httpClient := client.NewClient(logger)

    // Initialize token manager
    tokenManager, err := token.NewManager("tokens.db", logger)
    if err != nil {
        logger.Error("Failed to create token manager", "error", err)
        os.Exit(1)
    }
    defer tokenManager.Close()

    // Create token provider
    tokenProvider := &TokenProvider{
        accessToken: tokenManager.GetAccessToken(),
    }

    // Create Market client
    marketClient := client.NewMarket(httpClient, logger, tokenProvider)

    // Get quote for AAPL
    quote, err := marketClient.Quote(ctx, "AAPL", "")
    if err != nil {
        logger.Error("Failed to get quote", "error", err)
        os.Exit(1)
    }

    fmt.Println("AAPL Quote:")
    fmt.Printf("  Symbol: %s\n", quote.Quote.Symbol)
    fmt.Printf("  Description: %s\n", quote.Quote.Description)
    fmt.Printf("  Last Price: $%.2f\n", quote.Quote.LastPrice)
    fmt.Printf("  Bid: $%.2f x %d\n", quote.Quote.BidPrice, quote.Quote.BidSize)
    fmt.Printf("  Ask: $%.2f x %d\n", quote.Quote.AskPrice, quote.Quote.AskSize)
    fmt.Printf("  Volume: %d\n", quote.Quote.Volume)
    fmt.Printf("  Open: $%.2f\n", quote.Quote.OpenPrice)
    fmt.Printf("  High: $%.2f\n", quote.Quote.HighPrice)
    fmt.Printf("  Low: $%.2f\n", quote.Quote.LowPrice)
    fmt.Printf("  Previous Close: $%.2f\n", quote.Quote.PreviousClose)
    fmt.Printf("  Change: $%.2f (%.2f%%)\n", quote.Quote.Change, quote.Quote.PercentChange)

    // Get quotes for multiple symbols
    quotes, err := marketClient.Quotes(ctx, []string{"AAPL", "MSFT", "GOOGL"}, "", false)
    if err != nil {
        logger.Error("Failed to get quotes", "error", err)
        os.Exit(1)
    }

    fmt.Println("\nMultiple Quotes:")
    for symbol, q := range quotes.Quotes {
        fmt.Printf("  %s: $%.2f\n", symbol, q.LastPrice)
    }
}
```

## Example: Retrieve Option Chain (AAPL)

This example shows how to retrieve option chain data for Apple (AAPL):

```go
package main

import (
    "context"
    "fmt"
    "log/slog"
    "os"

    "github.com/citizenadam/go-schwabapi/pkg/client"
    "github.com/citizenadam/go-schwabapi/pkg/token"
)

// TokenProvider implements client.TokenGetter interface
type TokenProvider struct {
    accessToken string
}

func (t *TokenProvider) GetAccessToken() string {
    return t.accessToken
}

func main() {
    ctx := context.Background()
    logger := slog.Default()

    // Initialize HTTP client
    httpClient := client.NewClient(logger)

    // Initialize token manager
    tokenManager, err := token.NewManager("tokens.db", logger)
    if err != nil {
        logger.Error("Failed to create token manager", "error", err)
        os.Exit(1)
    }
    defer tokenManager.Close()

    // Create token provider
    tokenProvider := &TokenProvider{
        accessToken: tokenManager.GetAccessToken(),
    }

    // Create Market client
    marketClient := client.NewMarket(httpClient, logger, tokenProvider)

    // Get option chain for AAPL
    optionChain, err := marketClient.OptionChains(ctx, &client.OptionChainsRequest{
        Symbol:                 "AAPL",
        ContractType:           "ALL",           // ALL, CALL, or PUT
        StrikeCount:           10,              // Number of strikes to return
        IncludeUnderlyingQuote: true,            // Include underlying stock quote
        Strategy:              "SINGLE",        // SINGLE, ANALYTICAL, etc.
        Range:                 "ATM",           // ITM, OTM, ATM, etc.
    })
    if err != nil {
        logger.Error("Failed to get option chain", "error", err)
        os.Exit(1)
    }

    fmt.Println("AAPL Option Chain:")
    fmt.Printf("  Symbol: %s\n", optionChain.Symbol)
    fmt.Printf("  Underlying Price: $%.2f\n", optionChain.UnderlyingPrice)
    fmt.Printf("  Number of Contracts: %d\n", optionChain.NumberOfContracts)
    fmt.Printf("  Days to Expiration: %d\n", optionChain.DaysToExpiration)
    fmt.Printf("  Implied Volatility: %.2f%%\n", optionChain.Volatility*100)

    // Print call options
    fmt.Println("\nCall Options:")
    for expDate, strikes := range optionChain.CallExpDateMap {
        fmt.Printf("  Expiration: %s\n", expDate)
        for strike, contracts := range strikes {
            for _, contract := range contracts {
                fmt.Printf("    Strike $%s: Bid $%.2f / Ask $%.2f, Vol: %d, OI: %d\n",
                    strike,
                    contract.BidPrice,
                    contract.AskPrice,
                    contract.Volume,
                    contract.OpenInterest,
                )
            }
        }
    }

    // Print put options
    fmt.Println("\nPut Options:")
    for expDate, strikes := range optionChain.PutExpDateMap {
        fmt.Printf("  Expiration: %s\n", expDate)
        for strike, contracts := range strikes {
            for _, contract := range contracts {
                fmt.Printf("    Strike $%s: Bid $%.2f / Ask $%.2f, Vol: %d, OI: %d\n",
                    strike,
                    contract.BidPrice,
                    contract.AskPrice,
                    contract.Volume,
                    contract.OpenInterest,
                )
            }
        }
    }

    // Get option expiration chain (available expiration dates)
    expChain, err := marketClient.OptionExpirationChain(ctx, "AAPL", "", 0, 0)
    if err != nil {
        logger.Error("Failed to get expiration chain", "error", err)
        os.Exit(1)
    }

    fmt.Println("\nAvailable Expirations:")
    for _, exp := range expChain.ExpirationList {
        fmt.Printf("  Date: %s, Type: %s, Days: %d\n",
            exp.ExpirationDate,
            exp.ExpirationType,
            exp.DaysToExpiration,
        )
    }
}
```

## API Reference

### Account Operations

| Method | Description |
|--------|-------------|
| `LinkedAccounts(ctx)` | Get all linked account numbers and hashes |
| `AccountDetails(ctx, accountHash, fields)` | Get details for a specific account |
| `AccountDetailsAll(ctx, fields)` | Get details for all linked accounts |
| `AccountOrders(ctx, accountHash, ...)` | Get orders for a specific account |
| `AccountOrdersAll(ctx, ...)` | Get orders for all accounts |
| `Transactions(ctx, accountHash, ...)` | Get transaction history |
| `Preferences(ctx)` | Get user preferences including streamer info |

### Market Data Operations

| Method | Description |
|--------|-------------|
| `Quote(ctx, symbol, fields)` | Get quote for a single symbol |
| `Quotes(ctx, symbols, fields, indicative)` | Get quotes for multiple symbols |
| `OptionChains(ctx, request)` | Get option chain for a symbol |
| `OptionExpirationChain(ctx, symbol, ...)` | Get available option expirations |
| `PriceHistory(ctx, request)` | Get historical price data |
| `Movers(ctx, index, direction, change)` | Get market movers |
| `MarketHours(ctx, markets, date)` | Get market hours |
| `Instruments(ctx, symbols, projection)` | Get instrument information |

### OAuth Operations

| Method | Description |
|--------|-------------|
| `Authorize(ctx)` | Generate authorization URL |
| `Exchange(ctx, code)` | Exchange auth code for tokens |
| `RefreshToken(ctx, refreshToken)` | Refresh access token |
| `RevokeToken(ctx, token, tokenType)` | Revoke a token |
| `GetStreamerInfo(ctx)` | Get streaming connection info |

## Token Management

The library includes a token manager that handles:

- **Persistence**: Tokens are stored in a SQLite database
- **Auto-refresh**: Access tokens are automatically refreshed before expiry
- **Thread-safe**: Safe for concurrent use across multiple goroutines

```go
// Create token manager
tokenManager, err := token.NewManager("tokens.db", logger)

// Get current access token (auto-refreshes if needed)
accessToken := tokenManager.GetAccessToken()

// Get refresh token
refreshToken := tokenManager.GetRefreshToken()

// Manually trigger token update
tokenManager.UpdateTokens(ctx, forceAccessToken, forceRefreshToken)

// Close when done
defer tokenManager.Close()
```

## Error Handling

All API methods return errors that should be handled appropriately:

```go
quote, err := marketClient.Quote(ctx, "AAPL", "")
if err != nil {
    logger.Error("Failed to get quote", "error", err)
    // Handle error - check if it's an auth error, rate limit, etc.
    return
}
```

## Rate Limits

The Schwab API has rate limits. The library uses sensible timeouts and connection pooling:

- Request timeout: 60 seconds
- Dial timeout: 30 seconds
- Connection pool: 100 max idle connections

## License

MIT License
