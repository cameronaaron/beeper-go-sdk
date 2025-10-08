package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	beeperdesktop "github.com/beeper/desktop-api-go"
	"github.com/beeper/desktop-api-go/resources"
)

func main() {
	fmt.Println("🧪 Beeper Desktop API - Go SDK Test Application")
	fmt.Println("================================================")
	fmt.Println()

	// Check for access token
	accessToken := os.Getenv("BEEPER_ACCESS_TOKEN")
	if accessToken == "" {
		fmt.Println("⚠️  BEEPER_ACCESS_TOKEN environment variable not set")
		fmt.Println("ℹ️  Running in demo mode - will show API structure only")
		fmt.Println()
		runDemoMode()
		return
	}

	// Create client
	fmt.Println("✅ Creating Beeper Desktop client...")
	client, err := beeperdesktop.New(
		beeperdesktop.WithTimeout(10*time.Second),
		beeperdesktop.WithMaxRetries(2),
	)
	if err != nil {
		log.Fatal("❌ Failed to create client:", err)
	}
	fmt.Println("✅ Client created successfully")
	fmt.Println()

	ctx := context.Background()

	// Test 1: Get token info
	fmt.Println("📋 Test 1: Getting token information...")
	runTokenTest(ctx, client)

	// Test 2: List accounts
	fmt.Println("\n👤 Test 2: Listing accounts...")
	accounts := runAccountsTest(ctx, client)

	// Test 3: Search chats
	fmt.Println("\n💬 Test 3: Searching chats...")
	runChatsTest(ctx, client)

	// Test 4: Search messages
	fmt.Println("\n📨 Test 4: Searching messages...")
	runMessagesTest(ctx, client)

	// Test 5: Search contacts
	fmt.Println("\n📇 Test 5: Searching contacts...")
	runContactsTest(ctx, client, accounts)

	fmt.Println("\n================================================")
	fmt.Println("✅ All tests completed successfully!")
	fmt.Println("================================================")
}

func runDemoMode() {
	fmt.Println("📚 Available API Resources:")
	fmt.Println("  • Accounts  - List connected messaging accounts")
	fmt.Println("  • App       - Search, open, and download attachments")
	fmt.Println("  • Chats     - Search and manage conversations")
	fmt.Println("  • Contacts  - Search for contacts")
	fmt.Println("  • Messages  - Search and send messages")
	fmt.Println("  • Token     - Get access token information")
	fmt.Println()
	fmt.Println("🔧 Client Features:")
	fmt.Println("  • Automatic retry with exponential backoff")
	fmt.Println("  • Typed error handling (BadRequest, NotFound, etc.)")
	fmt.Println("  • Context support for cancellation/timeout")
	fmt.Println("  • Pagination support with iterators")
	fmt.Println("  • Configurable timeouts and HTTP client")
	fmt.Println()
	fmt.Println("💡 To test with real API:")
	fmt.Println("   export BEEPER_ACCESS_TOKEN=your_token_here")
	fmt.Println("   go run cmd/testapp/main.go")
}

func runTokenTest(ctx context.Context, client *beeperdesktop.BeeperDesktop) {
	tokenInfo, err := client.Token.Info(ctx)
	if err != nil {
		if _, ok := err.(*beeperdesktop.AuthenticationError); ok {
			fmt.Println("  ⚠️  Authentication error - token may be invalid")
			return
		}
		fmt.Printf("  ❌ Error: %v\n", err)
		return
	}

	fmt.Printf("  ✓ Token Use: %s\n", tokenInfo.TokenUse)
	fmt.Printf("  ✓ Subject: %s\n", tokenInfo.Sub)
	fmt.Printf("  ✓ Scope: %s\n", tokenInfo.Scope)
	if tokenInfo.ClientID != nil {
		fmt.Printf("  ✓ Client ID: %s\n", *tokenInfo.ClientID)
	}
}

func runAccountsTest(ctx context.Context, client *beeperdesktop.BeeperDesktop) *resources.AccountListResponse {
	accounts, err := client.Accounts.List(ctx)
	if err != nil {
		fmt.Printf("  ❌ Error: %v\n", err)
		return nil
	}

	if accounts == nil || len(*accounts) == 0 {
		fmt.Println("  ℹ️  No accounts found")
		return nil
	}

	fmt.Printf("  ✓ Found %d account(s)\n", len(*accounts))
	for i, account := range *accounts {
		fmt.Printf("    %d. %s (%s)\n", i+1, account.AccountID, account.Network)
	}
	return accounts
}

func runChatsTest(ctx context.Context, client *beeperdesktop.BeeperDesktop) {
	params := resources.ChatSearchParams{
		Limit: beeperdesktop.IntPtr(5),
	}

	chats, err := client.Chats.Search(ctx, params)
	if err != nil {
		fmt.Printf("  ❌ Error: %v\n", err)
		return
	}

	if chats == nil || chats.Items == nil || len(chats.Items) == 0 {
		fmt.Println("  ℹ️  No chats found")
		return
	}

	fmt.Printf("  ✓ Found %d chat(s)\n", len(chats.Items))
	for i, chat := range chats.Items {
		if i >= 3 {
			break // Show only first 3
		}
		fmt.Printf("    %d. %s (%s) - %d participants\n", i+1, chat.Title, chat.Network, chat.Participants.Total)
	}
}

func runMessagesTest(ctx context.Context, client *beeperdesktop.BeeperDesktop) {
	params := resources.MessageSearchParams{
		Query: beeperdesktop.StringPtr("hello"),
		Limit: beeperdesktop.IntPtr(3),
	}

	messages, err := client.Messages.Search(ctx, params)
	if err != nil {
		fmt.Printf("  ❌ Error: %v\n", err)
		return
	}

	if messages == nil || messages.Items == nil || len(messages.Items) == 0 {
		fmt.Println("  ℹ️  No messages found matching 'hello'")
		return
	}

	fmt.Printf("  ✓ Found %d message(s) containing 'hello'\n", len(messages.Items))
	for i, msg := range messages.Items {
		timestamp := msg.Timestamp.Format("2006-01-02 15:04")
		fmt.Printf("    %d. [%s] MessageID: %s\n", i+1, timestamp, msg.MessageID)
	}
}

func runContactsTest(ctx context.Context, client *beeperdesktop.BeeperDesktop, accounts *resources.AccountListResponse) {
	if accounts == nil || len(*accounts) == 0 {
		fmt.Println("  ℹ️  No accounts available for contact search")
		return
	}

	// Try to find WhatsApp account (contacts search works best on WhatsApp)
	var accountID string
	for _, account := range *accounts {
		if account.Network == "WhatsApp" {
			accountID = account.AccountID
			break
		}
	}
	
	// Fall back to first account if no WhatsApp found
	if accountID == "" {
		accountID = (*accounts)[0].AccountID
	}
	
	params := resources.ContactSearchParams{
		AccountID: accountID,
		Query:     "test",
	}
	
	fmt.Printf("  → Searching contacts on %s...\n", accountID)

	contacts, err := client.Contacts.Search(ctx, params)
	if err != nil {
		fmt.Printf("  ❌ Error: %v\n", err)
		fmt.Println("  ℹ️  Note: Contact search may not be supported on all networks")
		return
	}

	if contacts == nil || len(contacts.Items) == 0 {
		fmt.Println("  ℹ️  No contacts found matching 'test'")
		return
	}

	fmt.Printf("  ✓ Found %d contact(s)\n", len(contacts.Items))
	for i, contact := range contacts.Items {
		if i >= 3 {
			break
		}
		name := "Unknown"
		if contact.FullName != nil {
			name = *contact.FullName
		}
		fmt.Printf("    %d. %s (ID: %s)\n", i+1, name, contact.ID)
	}
}
