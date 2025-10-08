package main

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html"
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

		archivePath, err := archiveChat(ctx, client, chat)
		if err != nil {
			fmt.Printf("  ⚠️  Warning: %v\n", err)
			continue
		}

		fmt.Printf("  ✓ Archived to: %s\n", archivePath)
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

func archiveChat(ctx context.Context, client *beeperdesktop.BeeperDesktop, chat resources.Chat) (string, error) {
	// Fetch all messages for this chat
	messages, err := fetchChatMessages(ctx, client, chat)
	if err != nil {
		return "", fmt.Errorf("failed to fetch messages: %w", err)
	}

	// Sort messages by timestamp
	sort.Slice(messages, func(i, j int) bool {
		return messages[i].Timestamp.Before(messages[j].Timestamp)
	})

	archivedAt := time.Now()
	folder := getArchiveFolder(chat, archivedAt)
	baseName := getArchiveBaseName(chat)
	chatDir := filepath.Join(archiveDir, folder)
	if err := os.RemoveAll(chatDir); err != nil {
		return "", fmt.Errorf("failed to reset chat directory: %w", err)
	}
	if err := os.MkdirAll(chatDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create chat directory: %w", err)
	}

	markdown := generateMarkdown(chat, messages, archivedAt)
	if err := os.WriteFile(filepath.Join(chatDir, baseName+".md"), []byte(markdown), 0644); err != nil {
		return "", fmt.Errorf("failed to write markdown: %w", err)
	}

	if err := writeJSONArchive(chatDir, baseName, chat, messages, archivedAt); err != nil {
		return "", err
	}

	htmlContent := generateHTML(chat, messages, archivedAt)
	if err := os.WriteFile(filepath.Join(chatDir, baseName+".html"), []byte(htmlContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write HTML: %w", err)
	}

	textContent := generatePlainText(chat, messages, archivedAt)
	if err := os.WriteFile(filepath.Join(chatDir, baseName+".txt"), []byte(textContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write text export: %w", err)
	}

	return filepath.Join(archiveDir, folder), nil
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

func generateMarkdown(chat resources.Chat, messages []resources.Message, archivedAt time.Time) string {
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
	md.WriteString(fmt.Sprintf("**Archived:** %s\n\n", archivedAt.Format(timeFormat)))
	md.WriteString(fmt.Sprintf("**Total Messages:** %d\n\n", len(messages)))
	md.WriteString("*Generated by Beeper Chat Archive Tool*\n")

	return md.String()
}

func generateHTML(chat resources.Chat, messages []resources.Message, archivedAt time.Time) string {
	var htmlBuilder strings.Builder
	htmlBuilder.WriteString("<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n<meta charset=\"utf-8\">\n")
	htmlBuilder.WriteString(fmt.Sprintf("<title>%s</title>\n", html.EscapeString(chat.Title)))
	htmlBuilder.WriteString("<style>body{font-family:system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;margin:2rem;max-width:960px;} h1,h2,h3{margin-top:2rem;} .message{border-top:1px solid #ddd;padding:1rem 0;} .meta{color:#555;font-size:0.9rem;} blockquote{background:#f8f8f8;border-left:4px solid #ccc;padding:0.75rem;margin:0.75rem 0;} table{border-collapse:collapse;} </style>\n</head>\n<body>\n")

	htmlBuilder.WriteString(fmt.Sprintf("<h1>%s</h1>\n", html.EscapeString(chat.Title)))
	htmlBuilder.WriteString("<section>\n<ul>\n")
	htmlBuilder.WriteString(fmt.Sprintf("<li><strong>Network:</strong> %s</li>\n", html.EscapeString(chat.Network)))
	htmlBuilder.WriteString(fmt.Sprintf("<li><strong>Chat ID:</strong> <code>%s</code></li>\n", html.EscapeString(chat.ID)))
	htmlBuilder.WriteString(fmt.Sprintf("<li><strong>Participants:</strong> %d</li>\n", chat.Participants.Total))
	if chat.LastActivity != nil {
		htmlBuilder.WriteString(fmt.Sprintf("<li><strong>Last Activity:</strong> %s</li>\n", html.EscapeString(*chat.LastActivity)))
	}
	htmlBuilder.WriteString(fmt.Sprintf("<li><strong>Total Messages:</strong> %d</li>\n", len(messages)))
	htmlBuilder.WriteString("</ul>\n</section>\n")

	if len(chat.Participants.Items) > 0 {
		htmlBuilder.WriteString("<section>\n<h2>Participants</h2>\n<ol>\n")
		for _, participant := range chat.Participants.Items {
			name := "Unknown"
			if participant.FullName != nil {
				name = *participant.FullName
			}
			htmlBuilder.WriteString(fmt.Sprintf("<li><strong>%s</strong> (<code>%s</code>)</li>\n", html.EscapeString(name), html.EscapeString(participant.ID)))
		}
		htmlBuilder.WriteString("</ol>\n</section>\n")
	}

	htmlBuilder.WriteString("<section>\n<h2>Messages</h2>\n")
	currentDate := ""
	for i, msg := range messages {
		msgDate := msg.Timestamp.Format("2006-01-02")
		if msgDate != currentDate {
			currentDate = msgDate
			htmlBuilder.WriteString(fmt.Sprintf("<h3>%s</h3>\n", html.EscapeString(msg.Timestamp.Format("Monday, January 2, 2006"))))
		}

		senderName := "Unknown"
		if msg.SenderName != nil {
			senderName = *msg.SenderName
		}

		htmlBuilder.WriteString("<div class=\"message\">\n")
		htmlBuilder.WriteString(fmt.Sprintf("<div class=\"meta\"><strong>Message #%d</strong> &middot; From %s at %s &middot; ID <code>%s</code></div>\n", i+1, html.EscapeString(senderName), html.EscapeString(msg.Timestamp.Format("15:04:05")), html.EscapeString(msg.MessageID)))

		if msg.Text != nil && *msg.Text != "" {
			escaped := html.EscapeString(*msg.Text)
			htmlBuilder.WriteString(fmt.Sprintf("<blockquote>%s</blockquote>\n", strings.ReplaceAll(escaped, "\n", "<br>")))
		} else {
			htmlBuilder.WriteString("<blockquote><em>No text content</em></blockquote>\n")
		}

		if len(msg.Attachments) > 0 {
			htmlBuilder.WriteString("<div><strong>Attachments:</strong><ul>\n")
			for _, att := range msg.Attachments {
				name := "unknown"
				if att.FileName != nil {
					name = *att.FileName
				}
				htmlBuilder.WriteString("<li>")
				htmlBuilder.WriteString(html.EscapeString(name))
				htmlBuilder.WriteString(fmt.Sprintf(" (%s)", html.EscapeString(att.Type)))
				if att.FileSize != nil {
					htmlBuilder.WriteString(fmt.Sprintf(" &middot; %s", html.EscapeString(formatFileSize(*att.FileSize))))
				}
				if att.SrcURL != nil {
					htmlBuilder.WriteString(fmt.Sprintf(" &middot; <a href=\"%s\">Download</a>", html.EscapeString(*att.SrcURL)))
				}
				htmlBuilder.WriteString("</li>\n")
			}
			htmlBuilder.WriteString("</ul></div>\n")
		}

		if len(msg.Reactions) > 0 {
			htmlBuilder.WriteString("<div><strong>Reactions:</strong> ")
			for idx, reaction := range msg.Reactions {
				if idx > 0 {
					htmlBuilder.WriteString(", ")
				}
				htmlBuilder.WriteString(html.EscapeString(reaction.ReactionKey))
			}
			htmlBuilder.WriteString("</div>\n")
		}

		htmlBuilder.WriteString("</div>\n")
	}
	htmlBuilder.WriteString("</section>\n")

	htmlBuilder.WriteString("<footer><p><strong>Archived:</strong> " + html.EscapeString(archivedAt.Format(timeFormat)) + "</p>")
	htmlBuilder.WriteString(fmt.Sprintf("<p><strong>Total Messages:</strong> %d</p>", len(messages)))
	htmlBuilder.WriteString("<p><em>Generated by Beeper Chat Archive Tool</em></p></footer>\n")
	htmlBuilder.WriteString("</body>\n</html>\n")

	return htmlBuilder.String()
}

func generatePlainText(chat resources.Chat, messages []resources.Message, archivedAt time.Time) string {
	var textBuilder strings.Builder
	textBuilder.WriteString(fmt.Sprintf("Chat: %s\n", chat.Title))
	textBuilder.WriteString(fmt.Sprintf("Network: %s\n", chat.Network))
	textBuilder.WriteString(fmt.Sprintf("Chat ID: %s\n", chat.ID))
	textBuilder.WriteString(fmt.Sprintf("Participants: %d\n", chat.Participants.Total))
	if chat.LastActivity != nil {
		textBuilder.WriteString(fmt.Sprintf("Last Activity: %s\n", *chat.LastActivity))
	}
	textBuilder.WriteString(fmt.Sprintf("Total Messages: %d\n", len(messages)))
	textBuilder.WriteString("----------------------------------------\n\n")

	currentDate := ""
	for i, msg := range messages {
		msgDate := msg.Timestamp.Format("2006-01-02")
		if msgDate != currentDate {
			currentDate = msgDate
			textBuilder.WriteString(fmt.Sprintf("%s\n", msg.Timestamp.Format("Monday, January 2, 2006")))
			textBuilder.WriteString(strings.Repeat("=", 40) + "\n")
		}

		senderName := "Unknown"
		if msg.SenderName != nil {
			senderName = *msg.SenderName
		}

		textBuilder.WriteString(fmt.Sprintf("Message #%d\n", i+1))
		textBuilder.WriteString(fmt.Sprintf("From: %s\n", senderName))
		textBuilder.WriteString(fmt.Sprintf("Time: %s\n", msg.Timestamp.Format(timeFormat)))
		textBuilder.WriteString(fmt.Sprintf("Message ID: %s\n", msg.MessageID))

		if msg.Text != nil && *msg.Text != "" {
			textBuilder.WriteString(*msg.Text + "\n")
		} else {
			textBuilder.WriteString("[No text content]\n")
		}

		if len(msg.Attachments) > 0 {
			textBuilder.WriteString("Attachments:\n")
			for _, att := range msg.Attachments {
				name := "unknown"
				if att.FileName != nil {
					name = *att.FileName
				}
				line := fmt.Sprintf("- %s (%s)", name, att.Type)
				if att.FileSize != nil {
					line += " " + formatFileSize(*att.FileSize)
				}
				if att.SrcURL != nil {
					line += " " + *att.SrcURL
				}
				textBuilder.WriteString(line + "\n")
			}
		}

		if len(msg.Reactions) > 0 {
			reactions := make([]string, 0, len(msg.Reactions))
			for _, reaction := range msg.Reactions {
				reactions = append(reactions, reaction.ReactionKey)
			}
			textBuilder.WriteString("Reactions: " + strings.Join(reactions, ", ") + "\n")
		}

		textBuilder.WriteString("----------------------------------------\n")
	}

	textBuilder.WriteString(fmt.Sprintf("Archived: %s\n", archivedAt.Format(timeFormat)))
	textBuilder.WriteString(fmt.Sprintf("Total Messages: %d\n", len(messages)))
	textBuilder.WriteString("Generated by Beeper Chat Archive Tool\n")

	return textBuilder.String()
}

func writeJSONArchive(chatDir, baseName string, chat resources.Chat, messages []resources.Message, archivedAt time.Time) error {
	payload := struct {
		Title        string                     `json:"title"`
		Network      string                     `json:"network"`
		ChatID       string                     `json:"chat_id"`
		Participants resources.ChatParticipants `json:"participants"`
		LastActivity *string                    `json:"last_activity,omitempty"`
		ArchivedAt   string                     `json:"archived_at"`
		Messages     []resources.Message        `json:"messages"`
	}{
		Title:        chat.Title,
		Network:      chat.Network,
		ChatID:       chat.ID,
		Participants: chat.Participants,
		LastActivity: chat.LastActivity,
		ArchivedAt:   archivedAt.Format(time.RFC3339),
		Messages:     messages,
	}

	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.WriteFile(filepath.Join(chatDir, baseName+".json"), data, 0644); err != nil {
		return fmt.Errorf("failed to write JSON: %w", err)
	}

	return nil
}

func getArchiveFolder(chat resources.Chat, archivedAt time.Time) string {
	timestamp := archivedAt.Format("2006-01-02")
	network := sanitizeFilename(chat.Network)
	title := sanitizeFilename(chat.Title)
	if len(title) > 48 {
		title = title[:48]
	}

	identifier := buildChatIdentifier(chat)
	folderName := fmt.Sprintf("%s_%s", title, identifier)

	return filepath.Join(timestamp, network, folderName)
}

func getArchiveBaseName(chat resources.Chat) string {
	network := sanitizeFilename(chat.Network)
	title := sanitizeFilename(chat.Title)
	if len(title) > 48 {
		title = title[:48]
	}

	identifier := buildChatIdentifier(chat)
	return fmt.Sprintf("%s_%s_%s_messages", network, title, identifier)
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

func buildChatIdentifier(chat resources.Chat) string {
	idSlug := sanitizeFilename(chat.ID)
	if len(idSlug) > 16 {
		idSlug = idSlug[:16]
	}
	idSlug = strings.Trim(idSlug, "-")
	if idSlug == "" {
		idSlug = "chat"
	}

	checksum := sha1.Sum([]byte(chat.ID))
	hash := hex.EncodeToString(checksum[:])[:6]

	return fmt.Sprintf("%s-%s", idSlug, hash)
}
