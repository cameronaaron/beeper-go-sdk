# Go SDK Feature Parity

This Go SDK provides feature parity with the TypeScript version with Go-idiomatic patterns.

## Key Differences from TypeScript SDK

### 1. **Context Support**
- All API methods accept `context.Context` as the first parameter
- Enables cancellation, timeouts, and request scoping
- More explicit than TypeScript's implicit timeout handling

### 2. **Error Handling** 
- Uses Go's explicit error handling with typed errors
- Each HTTP status code maps to a specific error type
- Errors can be type-asserted for specific handling

### 3. **Pointer Helpers**
- Go requires explicit pointers for optional fields
- Utility functions like `StringPtr()`, `IntPtr()`, `BoolPtr()` provided
- Matches JSON omitempty behavior

### 4. **Generics for Pagination**
- Uses Go 1.18+ generics for type-safe pagination
- `Cursor[T]` type provides compile-time safety
- Iterator pattern for streaming large result sets

### 5. **Resource Structure**
- Each resource (Accounts, Chats, Messages, etc.) is a separate struct
- Maintains the same hierarchical organization as TypeScript
- Resource clients are initialized with the main client

### 6. **Configuration**
- Functional options pattern for client configuration
- Environment variable defaults maintained
- More explicit than TypeScript's object-based config

## API Compatibility

All endpoints from the TypeScript version are supported:

- ✅ **Accounts**: `List()`
- ✅ **App**: `DownloadAsset()`, `Open()`, `Search()`
- ✅ **Chats**: `Create()`, `Retrieve()`, `Archive()`, `Search()`
- ✅ **Chat Reminders**: `Create()`, `Delete()`
- ✅ **Contacts**: `Search()`
- ✅ **Messages**: `Search()`, `Send()`
- ✅ **Token**: `Info()`

## Usage Patterns

### Basic Usage
```go
client, err := beeperdesktop.New()
if err != nil {
    log.Fatal(err)
}

accounts, err := client.Accounts.List(context.Background())
```

### With Configuration
```go
client, err := beeperdesktop.New(
    beeperdesktop.WithAccessToken("token"),
    beeperdesktop.WithTimeout(30*time.Second),
    beeperdesktop.WithMaxRetries(3),
)
```

### Error Handling
```go
_, err := client.Chats.Retrieve(ctx, params)
if err != nil {
    switch e := err.(type) {
    case *beeperdesktop.NotFoundError:
        // Handle 404
    case *beeperdesktop.AuthenticationError:
        // Handle 401
    }
}
```

### Pagination
```go
params := resources.MessageSearchParams{...}
messages, err := client.Messages.Search(ctx, params)

// Manual pagination
for messages.Pagination != nil && messages.Pagination.HasMore {
    params.Cursor = messages.Pagination.Cursor
    nextPage, err := client.Messages.Search(ctx, params)
    // Process nextPage...
}
```

## Type Safety

The Go SDK provides compile-time type safety for:
- Request parameters (structs with proper JSON tags)
- Response types (strongly typed structs)
- Error types (specific error structs)
- Pagination cursors (generic types)

## Performance Considerations

- HTTP connection pooling (via Go's http.Client)
- Automatic retries with exponential backoff
- Context-based cancellation prevents resource leaks
- Memory-efficient pagination for large datasets