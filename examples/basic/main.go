package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	beeperdesktop "github.com/cameronaaron/beeper-go-sdk"
	"github.com/cameronaaron/beeper-go-sdk/resources"
)

func main() {
	// Create client with access token from environment variable BEEPER_ACCESS_TOKEN
	client, err := beeperdesktop.New()
	if err != nil {
		log.Fatal("Failed to create client:", err)
	}

	ctx := context.Background()

	// List connected accounts
	fmt.Println("Listing connected accounts...")
	accounts, err := client.Accounts.List(ctx)
	if err != nil {
		log.Fatal("Failed to list accounts:", err)
	}

	fmt.Printf("Found %d accounts:\n", len(*accounts))
	for i, account := range *accounts {
		fullName := ""
		if account.User.FullName != nil {
			fullName = *account.User.FullName
		}
		fmt.Printf("  %d. %s (%s) - %s\n", i+1, account.AccountID, account.Network, fullName)
	}

	if len(*accounts) == 0 {
		fmt.Println("No accounts found. Make sure Beeper Desktop is running and you have accounts connected.")
		return
	}

	// Search for chats
	fmt.Println("\nSearching for chats...")
	chats, err := client.Chats.Search(ctx, resources.ChatSearchParams{
		Limit:        beeperdesktop.IntPtr(5),
		IncludeMuted: beeperdesktop.BoolPtr(true),
	})
	if err != nil {
		log.Fatal("Failed to search chats:", err)
	}

	fmt.Printf("Found %d chats:\n", len(chats.Items))
	for i, chat := range chats.Items {
		chatTitle := chat.Title
		if strings.TrimSpace(chatTitle) == "" {
			chatTitle = "<no title>"
		}
		fmt.Printf("  %d. %s (%s) - %s\n", i+1, chat.ID, chat.Type, chatTitle)
	}

	// Search for messages
	fmt.Println("\nSearching for recent messages...")
	messages, err := client.Messages.Search(ctx, resources.MessageSearchParams{
		Limit:              beeperdesktop.IntPtr(5),
		ExcludeLowPriority: beeperdesktop.BoolPtr(true),
	})
	if err != nil {
		log.Fatal("Failed to search messages:", err)
	}

	fmt.Printf("Found %d messages:\n", len(messages.Items))
	for i, message := range messages.Items {
		text := "<no text>"
		if message.Text != nil {
			text = *message.Text
			if len(text) > 50 {
				text = text[:50] + "..."
			}
		}
		senderName := "<unknown>"
		if message.SenderName != nil {
			senderName = *message.SenderName
		}
		fmt.Printf("  %d. From %s: %s\n", i+1, senderName, text)
	}

	// Get token info
	fmt.Println("\nGetting token info...")
	tokenInfo, err := client.Token.Info(ctx)
	if err != nil {
		log.Fatal("Failed to get token info:", err)
	}

	fmt.Printf("Token info:\n")
	fmt.Printf("  Subject: %s\n", tokenInfo.Sub)
	fmt.Printf("  Scope: %s\n", tokenInfo.Scope)
	fmt.Printf("  Token Use: %s\n", tokenInfo.TokenUse)
	if tokenInfo.ClientID != nil {
		fmt.Printf("  Client ID: %s\n", *tokenInfo.ClientID)
	}

	fmt.Println("\nExample completed successfully!")
}
