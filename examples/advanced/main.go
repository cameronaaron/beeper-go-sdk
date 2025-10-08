package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	beeperdesktop "github.com/cameronaaron/beeper-go-sdk"
	"github.com/cameronaaron/beeper-go-sdk/resources"
)

func main() {
	// Create client with custom configuration
	client, err := beeperdesktop.New(
		beeperdesktop.WithTimeout(15*time.Second),
		beeperdesktop.WithMaxRetries(3),
	)
	if err != nil {
		log.Fatal("Failed to create client:", err)
	}

	ctx := context.Background()

	// Demonstrate error handling
	fmt.Println("=== Error Handling Demo ===")
	_, err = client.Chats.Retrieve(ctx, resources.ChatRetrieveParams{
		ChatID: "nonexistent-chat-id",
	})
	if err != nil {
		switch e := err.(type) {
		case *beeperdesktop.NotFoundError:
			fmt.Printf("Chat not found (expected): %s\n", e.Message)
		case *beeperdesktop.AuthenticationError:
			fmt.Printf("Authentication error: %s\n", e.Message)
		default:
			fmt.Printf("Other error: %v\n", err)
		}
	}

	// Advanced message search with filters
	fmt.Println("\n=== Advanced Message Search ===")
	searchParams := resources.MessageSearchParams{
		Query:              beeperdesktop.StringPtr("hello"),
		Limit:              beeperdesktop.IntPtr(5),
		ExcludeLowPriority: beeperdesktop.BoolPtr(true),
		Direction:          beeperdesktop.StringPtr("before"),
	}

	messages, err := client.Messages.Search(ctx, searchParams)
	if err != nil {
		log.Printf("Message search failed: %v", err)
	} else {
		fmt.Printf("Found %d messages:\n", len(messages.Items))
		for i, msg := range messages.Items {
			text := "<no text>"
			if msg.Text != nil {
				text = *msg.Text
				if len(text) > 100 {
					text = text[:100] + "..."
				}
			}
			fmt.Printf("  %d. %s: %s\n", i+1, msg.SenderID, text)
		}
	}

	// Chat management demo
	fmt.Println("\n=== Chat Management Demo ===")

	// Search for chats
	chatSearchParams := resources.ChatSearchParams{
		Limit:        beeperdesktop.IntPtr(3),
		IncludeMuted: beeperdesktop.BoolPtr(false),
		ChatType:     beeperdesktop.StringPtr("single"),
	}

	chats, err := client.Chats.Search(ctx, chatSearchParams)
	if err != nil {
		log.Printf("Chat search failed: %v", err)
	} else {
		fmt.Printf("Found %d chats:\n", len(chats.Items))
		for i, chat := range chats.Items {
			title := chat.Title
			if strings.TrimSpace(title) == "" {
				title = "<no title>"
			}
			fmt.Printf("  %d. %s (%s) - %s\n", i+1, chat.ID, chat.Type, title)

			// Demonstrate chat archiving (commented out to avoid side effects)
			/*
				if !chat.IsArchived {
					fmt.Printf("    Archiving chat %s\n", chat.ID)
					_, err := client.Chats.Archive(ctx, resources.ChatArchiveParams{
						ChatID:   chat.ID,
						Archived: true,
					})
					if err != nil {
						fmt.Printf("    Failed to archive: %v\n", err)
					} else {
						fmt.Printf("    Chat archived successfully\n")
					}
				}
			*/
		}
	}

	// App operations demo
	fmt.Println("\n=== App Operations Demo ===")

	// Global search
	appSearchParams := resources.AppSearchParams{
		Query:            "test",
		Limit:            beeperdesktop.IntPtr(3),
		MessageLimit:     beeperdesktop.IntPtr(2),
		ParticipantLimit: beeperdesktop.IntPtr(5),
	}

	searchResults, err := client.App.Search(ctx, appSearchParams)
	if err != nil {
		log.Printf("App search failed: %v", err)
	} else {
		fmt.Printf("Global search results:\n")
		fmt.Printf("  Chats: %d\n", len(searchResults.Chats))
		fmt.Printf("  Messages: %d\n", len(searchResults.Messages))

		for i, chatResult := range searchResults.Chats {
			fmt.Printf("  Chat %d: %s (%d participants, %d messages)\n",
				i+1, chatResult.Chat.ID, len(chatResult.Participants), len(chatResult.Messages))
		}
	}

	// Context cancellation demo
	fmt.Println("\n=== Context Cancellation Demo ===")

	// Create a context with a short timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	// This might timeout if the request takes too long
	_, err = client.Accounts.List(timeoutCtx)
	if err != nil {
		if timeoutCtx.Err() == context.DeadlineExceeded {
			fmt.Println("Request timed out (expected in some cases)")
		} else {
			fmt.Printf("Request failed: %v\n", err)
		}
	} else {
		fmt.Println("Request completed within timeout")
	}

	// Token info demo
	fmt.Println("\n=== Token Information ===")
	tokenInfo, err := client.Token.Info(ctx)
	if err != nil {
		log.Printf("Failed to get token info: %v", err)
	} else {
		fmt.Printf("Token details:\n")
		fmt.Printf("  Subject: %s\n", tokenInfo.Sub)
		fmt.Printf("  Scope: %s\n", tokenInfo.Scope)
		fmt.Printf("  Issued at: %s\n", time.Unix(tokenInfo.Iat, 0).Format(time.RFC3339))
		if tokenInfo.Exp != nil {
			fmt.Printf("  Expires at: %s\n", time.Unix(*tokenInfo.Exp, 0).Format(time.RFC3339))
		}
	}

	fmt.Println("\n=== Advanced Example Completed ===")
}
