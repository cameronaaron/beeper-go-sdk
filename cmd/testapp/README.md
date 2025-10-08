# Beeper Desktop API - Go SDK Test Application

This is a comprehensive test application that demonstrates all major features of the Beeper Desktop API Go SDK.

## Features

The test application validates the following SDK functionality:

- ✅ **Client Initialization** - Creates a client with custom configuration
- ✅ **Token Information** - Retrieves and displays OAuth token details
- ✅ **Account Management** - Lists all connected messaging accounts
- ✅ **Chat Operations** - Searches for and displays conversations
- ✅ **Message Search** - Searches messages across all chats
- ✅ **Contact Search** - Finds contacts across accounts
- ✅ **Error Handling** - Demonstrates typed error handling
- ✅ **Context Support** - Uses context for request management

## Running the Test App

### Demo Mode (No API Access Required)

Run without an access token to see the API structure and available features:

```bash
# From the Golang directory
make demo

# Or directly
go run cmd/testapp/main.go
```

### With Real API Access

If you have a Beeper Desktop instance running and an access token:

```bash
# Set your access token
export BEEPER_ACCESS_TOKEN=your_token_here

# Run the test app
go run cmd/testapp/main.go

# Or build and run
make build-testapp
./testapp
```

### Using the Makefile

```bash
# Build the test app binary
make build-testapp

# Build and run
make run-testapp

# Run in demo mode
make demo
```

## What It Tests

### 1. Token Information
- Validates authentication
- Retrieves token metadata (scope, subject, client ID)
- Demonstrates error handling for invalid tokens

### 2. Account Listing
- Lists all connected messaging accounts
- Displays account IDs and network types
- Shows account count

### 3. Chat Search
- Searches for recent conversations
- Displays chat names and IDs
- Limits results for readability (shows first 3)

### 4. Message Search
- Searches messages containing specific text ("hello")
- Displays message timestamps and IDs
- Demonstrates query parameter usage

### 5. Contact Search
- Searches for contacts across accounts
- Displays contact names and IDs
- Shows pagination support

## Expected Output

### Demo Mode
```
🧪 Beeper Desktop API - Go SDK Test Application
================================================

⚠️  BEEPER_ACCESS_TOKEN environment variable not set
ℹ️  Running in demo mode - will show API structure only

📚 Available API Resources:
  • Accounts  - List connected messaging accounts
  • App       - Search, open, and download attachments
  • Chats     - Search and manage conversations
  • Contacts  - Search for contacts
  • Messages  - Search and send messages
  • Token     - Get access token information

🔧 Client Features:
  • Automatic retry with exponential backoff
  • Typed error handling (BadRequest, NotFound, etc.)
  • Context support for cancellation/timeout
  • Pagination support with iterators
  • Configurable timeouts and HTTP client
```

### With Valid Token
```
🧪 Beeper Desktop API - Go SDK Test Application
================================================

✅ Creating Beeper Desktop client...
✅ Client created successfully

📋 Test 1: Getting token information...
  ✓ Token Use: access
  ✓ Subject: user_abc123
  ✓ Scope: openid profile

👤 Test 2: Listing accounts...
  ✓ Found 3 account(s)
    1. whatsapp_123 (whatsapp)
    2. signal_456 (signal)
    3. telegram_789 (telegram)

💬 Test 3: Searching chats...
  ✓ Found 15 chat(s)
    1. Family Group (ID: chat_abc)
    2. Work Team (ID: chat_def)
    3. Friends (ID: chat_ghi)

📨 Test 4: Searching messages...
  ✓ Found 3 message(s) containing 'hello'
    1. [2025-10-07 14:30] MessageID: msg_001
    2. [2025-10-06 09:15] MessageID: msg_002
    3. [2025-10-05 18:45] MessageID: msg_003

📇 Test 5: Searching contacts...
  ✓ Found 3 contact(s)
    1. John Doe (ID: user_001)
    2. Jane Smith (ID: user_002)
    3. Bob Johnson (ID: user_003)

================================================
✅ All tests completed successfully!
================================================
```

## Configuration

The test app uses the following client configuration:

```go
client, err := beeperdesktop.New(
    beeperdesktop.WithTimeout(10*time.Second),
    beeperdesktop.WithMaxRetries(2),
)
```

You can modify `cmd/testapp/main.go` to test different configurations:

- `WithBaseURL(url)` - Custom API endpoint
- `WithHTTPClient(client)` - Custom HTTP client
- `WithUserAgent(ua)` - Custom user agent
- `WithMaxRetries(n)` - Maximum retry attempts

## Error Handling

The test app demonstrates proper error handling:

```go
if err != nil {
    if _, ok := err.(*beeperdesktop.AuthenticationError); ok {
        fmt.Println("  ⚠️  Authentication error - token may be invalid")
        return
    }
    fmt.Printf("  ❌ Error: %v\n", err)
    return
}
```

All errors are gracefully handled, and the app continues to run even if individual tests fail.

## Customization

To add more tests or modify existing ones:

1. Add a new test function following the pattern `runXxxTest()`
2. Call it from `main()` after creating the client
3. Use appropriate error handling and output formatting

Example:
```go
func runMyCustomTest(ctx context.Context, client *beeperdesktop.BeeperDesktop) {
    fmt.Println("\n🔬 Test X: My custom test...")
    // Your test code here
}
```

## Troubleshooting

### "BEEPER_ACCESS_TOKEN environment variable not set"
- This is expected if you haven't set the token
- The app will run in demo mode showing available features
- To use the real API, set the environment variable

### Connection Errors
- Ensure Beeper Desktop is running on localhost:23373
- Check if the access token is valid
- Verify network connectivity

### Empty Results
- Some queries may return no results if your account is new
- This is normal and not an error
- Try different search queries or check if you have data

## Development

To modify or extend the test app:

```bash
# Edit the source
vim cmd/testapp/main.go

# Test your changes
go run cmd/testapp/main.go

# Build for distribution
go build -o testapp cmd/testapp/main.go
```

## See Also

- [Main README](../../README.md) - Full SDK documentation
- [Basic Example](../../examples/basic/) - Simple usage example
- [Advanced Example](../../examples/advanced/) - Complex features
- [API Documentation](../../GOLANG_FEATURES.md) - Go-specific patterns
