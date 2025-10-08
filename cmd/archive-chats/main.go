package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	beeperdesktop "github.com/cameronaaron/beeper-go-sdk"
	"github.com/cameronaaron/beeper-go-sdk/resources"
)

const (
	archiveDir = "chat-archives"
	timeFormat = "2006-01-02 15:04:05"
)

func main() {
	fmt.Println("📦 Beeper Chat Archive Tool")
	fmt.Println("============================")
	fmt.Println()

	// Create client
	client, err := beeperdesktop.New(
		beeperdesktop.WithTimeout(30*time.Second),
		beeperdesktop.WithMaxRetries(3),
	)
	if err != nil {
		log.Fatal("❌ Failed to create client:", err)
	}

	ctx := context.Background()

	// Get all chats
	fmt.Println("🔍 Fetching chats...")
	chats, err := fetchAllChats(ctx, client)
	if err != nil {
		log.Fatal("❌ Failed to fetch chats:", err)
	}

	if len(chats) == 0 {
		fmt.Println("ℹ️  No chats found to archive")
		return
	}

	fmt.Printf("✓ Found %d chats\n\n", len(chats))

	// Let user select chats to archive
	selectedChats := selectChatsInteractive(chats)
	if len(selectedChats) == 0 {
		fmt.Println("No chats selected for archiving")
		return
	}

	// Create archive directory
	if err := os.MkdirAll(archiveDir, 0755); err != nil {
		log.Fatal("❌ Failed to create archive directory:", err)
	}

	// Archive each selected chat
	fmt.Println("\n📝 Archiving chats...")
	for i, chat := range selectedChats {
		fmt.Printf("[%d/%d] Archiving: %s\n", i+1, len(selectedChats), chat.Title)

		if err := archiveChat(ctx, client, chat); err != nil {
			fmt.Printf("  ⚠️  Warning: %v\n", err)
			continue
		}

		fmt.Printf("  ✓ Archived to: %s\n", getArchiveFilename(chat))
	}

	fmt.Println("\n✅ Archiving complete!")
	fmt.Printf("📁 Archives saved to: %s/\n", archiveDir)
}

func fetchAllChats(ctx context.Context, client *beeperdesktop.BeeperDesktop) ([]resources.Chat, error) {
	var allChats []resources.Chat

	params := resources.ChatSearchParams{
		Limit: beeperdesktop.IntPtr(100),
	}

	result, err := client.Chats.Search(ctx, params)
	if err != nil {
		return nil, err
	}

	allChats = append(allChats, result.Items...)

	// Fetch more pages if available
	for result.Pagination != nil && result.Pagination.HasMore {
		params.Cursor = result.Pagination.Cursor
		result, err = client.Chats.Search(ctx, params)
		if err != nil {
			return nil, err
		}
		allChats = append(allChats, result.Items...)
	}

	return allChats, nil
}

func selectChatsInteractive(chats []resources.Chat) []resources.Chat {
	fmt.Println("Select chats to archive:")
	fmt.Println("  [a] Archive all chats")
	fmt.Println("  [#] Enter chat numbers (comma-separated)")
	fmt.Println("  [q] Quit")
	fmt.Println()

	// Display chats
	for i, chat := range chats {
		participants := fmt.Sprintf("%d participants", chat.Participants.Total)
		unread := ""
		if chat.UnreadCount > 0 {
			unread = fmt.Sprintf(" [%d unread]", chat.UnreadCount)
		}

		fmt.Printf("  %3d. %-40s %-20s %s%s\n",
			i+1,
			truncateString(chat.Title, 40),
			fmt.Sprintf("(%s)", chat.Network),
			participants,
			unread,
		)
	}

	fmt.Print("\nSelection: ")
	var input string
	fmt.Scanln(&input)

	input = strings.TrimSpace(input)

	switch input {
	case "a", "A", "all":
		return chats
	case "q", "Q", "quit":
		return nil
	default:
		// Parse numbers
		selected := []resources.Chat{}
		parts := strings.Split(input, ",")
		for _, part := range parts {
			var num int
			if _, err := fmt.Sscanf(strings.TrimSpace(part), "%d", &num); err == nil {
				if num > 0 && num <= len(chats) {
					selected = append(selected, chats[num-1])
				}
			}
		}
		return selected
	}
}

func archiveChat(ctx context.Context, client *beeperdesktop.BeeperDesktop, chat resources.Chat) error {
	// Fetch all messages for this chat
	messages, err := fetchChatMessages(ctx, client, chat)
	if err != nil {
		return fmt.Errorf("failed to fetch messages: %w", err)
	}

	// Sort messages by timestamp
	sort.Slice(messages, func(i, j int) bool {
		return messages[i].Timestamp.Before(messages[j].Timestamp)
	})

	// Generate markdown
	markdown := generateMarkdown(chat, messages)

	// Save to file
	filename := filepath.Join(archiveDir, getArchiveFilename(chat))
	if err := os.WriteFile(filename, []byte(markdown), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func fetchChatMessages(ctx context.Context, client *beeperdesktop.BeeperDesktop, chat resources.Chat) ([]resources.Message, error) {
	var allMessages []resources.Message

	params := resources.MessageSearchParams{
		ChatIDs:    []string{chat.ID},
		AccountIDs: []string{chat.AccountID},
		Limit:      beeperdesktop.IntPtr(100),
		Direction:  beeperdesktop.StringPtr("before"),
	}

	result, err := client.Messages.Search(ctx, params)
	if err != nil {
		return nil, err
	}

	allMessages = append(allMessages, result.Items...)

	// Fetch more pages if available
	for result.Pagination != nil && result.Pagination.HasMore {
		params.Cursor = result.Pagination.Cursor
		result, err = client.Messages.Search(ctx, params)
		if err != nil {
			return nil, err
		}
		allMessages = append(allMessages, result.Items...)

		// Progress indicator for large chats
		if len(allMessages)%500 == 0 {
			fmt.Printf("  → Fetched %d messages...\n", len(allMessages))
		}
	}

	return allMessages, nil
}

func generateMarkdown(chat resources.Chat, messages []resources.Message) string {
	var md strings.Builder

	// Header
	md.WriteString(fmt.Sprintf("# %s\n\n", chat.Title))
	md.WriteString(fmt.Sprintf("**Network:** %s\n\n", chat.Network))
	md.WriteString(fmt.Sprintf("**Chat ID:** `%s`\n\n", chat.ID))
	md.WriteString(fmt.Sprintf("**Participants:** %d\n\n", chat.Participants.Total))

	if chat.LastActivity != nil {
		md.WriteString(fmt.Sprintf("**Last Activity:** %s\n\n", *chat.LastActivity))
	}

	md.WriteString(fmt.Sprintf("**Total Messages:** %d\n\n", len(messages)))
	md.WriteString("---\n\n")

	// Table of contents
	if len(messages) > 100 {
		md.WriteString("## Table of Contents\n\n")
		md.WriteString("*Messages organized chronologically*\n\n")
		md.WriteString("---\n\n")
	}

	// Participants section
	if len(chat.Participants.Items) > 0 {
		md.WriteString("## Participants\n\n")
		for i, participant := range chat.Participants.Items {
			name := "Unknown"
			if participant.FullName != nil {
				name = *participant.FullName
			}
			md.WriteString(fmt.Sprintf("%d. **%s** (`%s`)\n", i+1, name, participant.ID))
		}
		md.WriteString("\n---\n\n")
	}

	// Messages section
	md.WriteString("## Messages\n\n")

	var currentDate string
	for i, msg := range messages {
		// Add date header when date changes
		msgDate := msg.Timestamp.Format("2006-01-02")
		if msgDate != currentDate {
			currentDate = msgDate
			md.WriteString(fmt.Sprintf("\n### 📅 %s\n\n", msg.Timestamp.Format("Monday, January 2, 2006")))
		}

		// Message metadata
		timestamp := msg.Timestamp.Format("15:04:05")
		senderName := "Unknown"
		if msg.SenderName != nil {
			senderName = *msg.SenderName
		}

		// Message number for reference
		md.WriteString(fmt.Sprintf("#### Message #%d\n\n", i+1))
		md.WriteString(fmt.Sprintf("**From:** %s  \n", senderName))
		md.WriteString(fmt.Sprintf("**Time:** %s  \n", timestamp))
		md.WriteString(fmt.Sprintf("**Message ID:** `%s`\n\n", msg.MessageID))

		// Message text
		if msg.Text != nil && *msg.Text != "" {
			text := *msg.Text
			// Escape markdown special characters in message text
			text = strings.ReplaceAll(text, "\\", "\\\\")
			md.WriteString(fmt.Sprintf("> %s\n\n", strings.ReplaceAll(text, "\n", "\n> ")))
		} else {
			md.WriteString("> *[No text content]*\n\n")
		}

		// Attachments
		if len(msg.Attachments) > 0 {
			md.WriteString("**Attachments:**\n\n")
			for j, att := range msg.Attachments {
				attType := att.Type
				fileName := "unknown"
				if att.FileName != nil {
					fileName = *att.FileName
				}

				md.WriteString(fmt.Sprintf("- Attachment %d: `%s` (%s)", j+1, fileName, attType))

				if att.FileSize != nil {
					md.WriteString(fmt.Sprintf(" - %s", formatFileSize(*att.FileSize)))
				}

				if att.SrcURL != nil {
					md.WriteString(fmt.Sprintf("\n  - URL: %s", *att.SrcURL))
				}

				md.WriteString("\n")
			}
			md.WriteString("\n")
		}

		// Reactions
		if len(msg.Reactions) > 0 {
			md.WriteString("**Reactions:** ")
			for j, reaction := range msg.Reactions {
				if j > 0 {
					md.WriteString(", ")
				}
				md.WriteString(reaction.ReactionKey)
			}
			md.WriteString("\n\n")
		}

		md.WriteString("---\n\n")
	}

	// Footer
	md.WriteString("\n## Archive Information\n\n")
	md.WriteString(fmt.Sprintf("**Archived:** %s\n\n", time.Now().Format(timeFormat)))
	md.WriteString(fmt.Sprintf("**Total Messages:** %d\n\n", len(messages)))
	md.WriteString("*Generated by Beeper Chat Archive Tool*\n")

	return md.String()
}

func getArchiveFilename(chat resources.Chat) string {
	timestamp := time.Now().Format("2006-01-02")
	network := sanitizeFilename(chat.Network)
	title := sanitizeFilename(chat.Title)

	if len(title) > 50 {
		title = title[:50]
	}

	return fmt.Sprintf("%s_%s_%s.md", timestamp, network, title)
}

func sanitizeFilename(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	lastDash := false
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			lastDash = false
		case r == ' ' || r == '-' || r == '_':
			if !lastDash {
				b.WriteRune('-')
				lastDash = true
			}
		default:
			if !lastDash {
				b.WriteRune('-')
				lastDash = true
			}
		}
	}

	result := strings.Trim(b.String(), "-")
	if result == "" {
		return "chat"
	}
	return result
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
