# Beeper Desktop API Go SDK

[![Go Reference](https://pkg.go.dev/badge/github.com/cameronaaron/beeper-go-sdk.svg)](https://pkg.go.dev/github.com/cameronaaron/beeper-go-sdk)
[![Go Report Card](https://goreportcard.com/badge/github.com/cameronaaron/beeper-go-sdk)](https://goreportcard.com/report/github.com/cameronaaron/beeper-go-sdk)

This library provides convenient access to the Beeper Desktop API from Go applications.

The documentation for Beeper Desktop API can be found on [developers.beeper.com/desktop-api](https://developers.beeper.com/desktop-api/).

## Installation

```bash
go get github.com/cameronaaron/beeper-go-sdk
```

## Usage

```go
package main

import (
    "context"
    "fmt"
    "log"

    beeperdesktop "github.com/cameronaaron/beeper-go-sdk"
    "github.com/cameronaaron/beeper-go-sdk/resources"
)

func main() {
    // Create client with access token from environment variable
    client, err := beeperdesktop.New()
    if err != nil {
        log.Fatal(err)
    }

    // Or create client with explicit access token
    client, err = beeperdesktop.New(
        beeperdesktop.WithAccessToken("your-access-token"),
    )
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    // List connected accounts
    accounts, err := client.Accounts.List(ctx)
    if err != nil {
        log.Fatal(err)
    }

    for _, account := range *accounts {
        fmt.Printf("Account: %s (%s)\n", account.AccountID, account.Network)
    }

    // Search chats
    chats, err := client.Chats.Search(ctx, resources.ChatSearchParams{
        Limit: beeperdesktop.IntPtr(10),
    })
    if err != nil {
        log.Fatal(err)
    }

    for _, chat := range chats.Items {
        fmt.Printf("Chat: %s\n", chat.ID)
    }

    // Send a message
    response, err := client.Messages.Send(ctx, resources.MessageSendParams{
        ChatID: "your-chat-id",
        Text:   "Hello from Go!",
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Message sent: %s\n", response.MessageID)
}
```

## Configuration Options

The client can be configured with various options:

```go
client, err := beeperdesktop.New(
    beeperdesktop.WithAccessToken("your-access-token"),
    beeperdesktop.WithBaseURL("http://localhost:23373"),
    beeperdesktop.WithTimeout(30*time.Second),
    beeperdesktop.WithMaxRetries(3),
    beeperdesktop.WithUserAgent("my-app/1.0"),
    beeperdesktop.WithHTTPClient(customHTTPClient),
)
```

## Error Handling

The SDK provides typed errors for different HTTP status codes:

```go
accounts, err := client.Accounts.List(ctx)
if err != nil {
    switch e := err.(type) {
    case *beeperdesktop.AuthenticationError:
        fmt.Println("Authentication failed:", e.Message)
    case *beeperdesktop.NotFoundError:
        fmt.Println("Resource not found:", e.Message)
    case *beeperdesktop.RateLimitError:
        fmt.Println("Rate limited:", e.Message)
    default:
        fmt.Println("Other error:", err)
    }
    return
}
```

## Pagination

For paginated endpoints, you can iterate through all results:

```go
// Search messages with pagination
params := resources.MessageSearchParams{
    Query: beeperdesktop.StringPtr("hello"),
    Limit: beeperdesktop.IntPtr(10),
}

for {
    messages, err := client.Messages.Search(ctx, params)
    if err != nil {
        log.Fatal(err)
    }

    for _, message := range messages.Items {
        fmt.Printf("Message: %s\n", *message.Text)
    }

    // Check if there are more pages
    if messages.Pagination == nil || !messages.Pagination.HasMore {
        break
    }

    // Set cursor for next page
    params.Cursor = messages.Pagination.Cursor
}
```

## Context Support

All API methods accept a `context.Context` for cancellation and timeouts:

```go
// With timeout
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

accounts, err := client.Accounts.List(ctx)

// With cancellation
ctx, cancel := context.WithCancel(context.Background())
go func() {
    time.Sleep(5 * time.Second)
    cancel() // Cancel the request
}()

accounts, err := client.Accounts.List(ctx)
```

## Web Chat Experience

Run the bundled web client for a fully modern chatting surface backed by this SDK:

```bash
go run ./cmd/webchat
```

Open `http://localhost:8080`, drop in your Beeper API access token (and optional base URL), and enjoy a sleek, responsive interface. The implementation lives in `cmd/webchat/` and embeds all assets, so the binary remains self-contained.

## Environment Variables

The SDK respects the following environment variables:

- `BEEPER_ACCESS_TOKEN`: Access token for authentication
- `BEEPER_DESKTOP_BASE_URL`: Base URL for the API (defaults to `http://localhost:23373`)

## Requirements

- Go 1.21 or later

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
