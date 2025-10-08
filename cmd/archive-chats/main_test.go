package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cameronaaron/beeper-go-sdk/resources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"simple", "Project Team", "project-team"},
		{"mixed punctuation", "Beeper (Matrix)", "beeper-matrix"},
		{"emoji and symbols", "🔥 Launch!", "launch"},
		{"repeated separators", "Name__With--Various   Spaces", "name-with-various-spaces"},
		{"empty result fallback", "???", "chat"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, sanitizeFilename(tt.input))
		})
	}
}

func TestTruncateString(t *testing.T) {
	assert.Equal(t, "short", truncateString("short", 10))
	assert.Equal(t, "short", truncateString("short", 5))
	assert.Equal(t, "hell...", truncateString("hello world", 7))
}

func TestFormatFileSize(t *testing.T) {
	assert.Equal(t, "512 B", formatFileSize(512))
	assert.Equal(t, "2.0 KB", formatFileSize(2048))
	assert.Equal(t, "1.5 KB", formatFileSize(1536))
	assert.Equal(t, "5.0 MB", formatFileSize(5*1024*1024))
}

func TestArchiveFolderAndBaseName(t *testing.T) {
	archivedAt := time.Date(2025, 10, 8, 12, 0, 0, 0, time.UTC)
	chat := resources.Chat{
		ID:      "!project:beeper.local",
		Network: "Matrix",
		Title:   "Project Updates 🚀",
	}

	folder := getArchiveFolder(chat, archivedAt)
	parts := strings.Split(folder, string(filepath.Separator))
	require.Len(t, parts, 3)
	assert.Equal(t, "2025-10-08", parts[0])
	assert.Equal(t, "matrix", parts[1])
	identifier := buildChatIdentifier(chat)
	expectedFolder := fmt.Sprintf("%s_%s", sanitizeFilename(chat.Title), identifier)
	assert.Equal(t, expectedFolder, parts[2])

	baseName := getArchiveBaseName(chat)
	assert.Equal(t, fmt.Sprintf("matrix_%s_%s_messages", sanitizeFilename(chat.Title), identifier), baseName)
	assert.True(t, strings.HasPrefix(baseName, "matrix_"))
	assert.True(t, strings.Contains(baseName, "_messages"))

	longTitle := strings.Repeat("A", 60)
	chat.Title = longTitle
	chat.ID = "!1234567890:example"

	baseName = getArchiveBaseName(chat)
	parts = strings.Split(baseName, "_")
	require.Equal(t, 4, len(parts))
	assert.Equal(t, sanitizeFilename(chat.Network), parts[0])
	assert.LessOrEqual(t, len(parts[1]), 48)
	assert.True(t, strings.HasSuffix(baseName, "_messages"))
	identifierParts := strings.Split(parts[2], "-")
	require.GreaterOrEqual(t, len(identifierParts), 2)
	assert.LessOrEqual(t, len(strings.Join(identifierParts[:len(identifierParts)-1], "-")), 16)
	hashPart := identifierParts[len(identifierParts)-1]
	assert.Len(t, hashPart, 6)
	for _, r := range hashPart {
		require.True(t, (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f'))
	}
}

func TestGenerateMarkdown(t *testing.T) {
	lastActivity := "2025-10-07T12:34:56Z"
	chat := resources.Chat{
		ID:        "!updates:beeper.local",
		AccountID: "acc_123",
		Network:   "Beeper (Matrix)",
		Title:     "Product Updates",
		Participants: resources.ChatParticipants{
			Total: 2,
			Items: []resources.User{
				{
					ID:       "@alice:beeper.com",
					FullName: ptr("Alice"),
				},
				{
					ID: "@bob:beeper.com",
				},
			},
		},
		LastActivity: &lastActivity,
	}

	ts1 := time.Date(2025, 10, 7, 12, 0, 0, 0, time.UTC)
	ts2 := ts1.Add(5 * time.Minute)

	messages := []resources.Message{
		{
			MessageID:  "msg_1",
			Timestamp:  ts1,
			SenderName: ptr("Alice"),
			Text:       ptr("Hello\nWorld"),
			Attachments: []resources.Attachment{
				{
					Type:     "img",
					FileName: ptr("photo.png"),
					FileSize: ptr(int64(2048)),
					SrcURL:   ptr("https://example.com/photo.png"),
				},
			},
			Reactions: []resources.Reaction{{ReactionKey: "👍"}},
		},
		{
			MessageID:  "msg_2",
			Timestamp:  ts2,
			SenderName: ptr("Bob"),
		},
	}

	archivedAt := time.Date(2025, 10, 8, 15, 0, 0, 0, time.UTC)
	markdown := generateMarkdown(chat, messages, archivedAt)

	assert.Contains(t, markdown, "# Product Updates")
	assert.Contains(t, markdown, "**Network:** Beeper (Matrix)")
	assert.Contains(t, markdown, "**Chat ID:** `!updates:beeper.local`")
	assert.Contains(t, markdown, "## Participants")
	assert.Contains(t, markdown, "1. **Alice** (`@alice:beeper.com`)")
	assert.Contains(t, markdown, "2. **Unknown** (`@bob:beeper.com`)")

	expectedDate := fmt.Sprintf("### 📅 %s", ts1.Format("Monday, January 2, 2006"))
	assert.Contains(t, markdown, expectedDate)
	assert.Contains(t, markdown, "#### Message #1")
	assert.Contains(t, markdown, "> Hello\n> World")
	assert.Contains(t, markdown, "- Attachment 1: `photo.png` (img) - 2.0 KB")
	assert.Contains(t, markdown, "  - URL: https://example.com/photo.png")
	assert.Contains(t, markdown, "**Reactions:** 👍")
	assert.Contains(t, markdown, "> *[No text content]*")
	assert.Contains(t, markdown, "## Archive Information")
	assert.Contains(t, markdown, archivedAt.Format(timeFormat))
	assert.Contains(t, markdown, "*Generated by Beeper Chat Archive Tool*")
}

func ptr[T any](v T) *T {
	return &v
}

func TestBuildChatIdentifier(t *testing.T) {
	chat := resources.Chat{ID: "!project:beeper.local"}
	identifier := buildChatIdentifier(chat)
	parts := strings.Split(identifier, "-")
	require.GreaterOrEqual(t, len(parts), 2)
	prefix := strings.Join(parts[:len(parts)-1], "-")
	assert.NotEmpty(t, prefix)
	assert.LessOrEqual(t, len(prefix), 16)
	hashPart := parts[len(parts)-1]
	assert.Len(t, hashPart, 6)
	for _, r := range hashPart {
		require.True(t, (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f'))
	}

	chat.ID = "short"
	identifier = buildChatIdentifier(chat)
	assert.True(t, strings.HasPrefix(identifier, "short-"))
}
