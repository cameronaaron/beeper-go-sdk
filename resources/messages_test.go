package resources_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	beeperdesktop "github.com/cameronaaron/beeper-go-sdk"
	"github.com/cameronaaron/beeper-go-sdk/resources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMessagesSearchQueryEncoding(t *testing.T) {
	var capturedURL *url.URL

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resources.MessagesCursor{})
	}))
	defer server.Close()

	client, err := beeperdesktop.New(
		beeperdesktop.WithAccessToken("token"),
		beeperdesktop.WithBaseURL(server.URL),
		beeperdesktop.WithMaxRetries(0),
	)
	require.NoError(t, err)

	_, err = client.Messages.Search(context.Background(), resources.MessageSearchParams{
		AccountIDs:   []string{"account-1", "account-2"},
		ChatIDs:      []string{"chat-1"},
		Limit:        beeperdesktop.IntPtr(25),
		Direction:    beeperdesktop.StringPtr("before"),
		IncludeMuted: beeperdesktop.BoolPtr(true),
	})
	require.NoError(t, err)

	require.NotNil(t, capturedURL)
	values := capturedURL.Query()

	assert.Equal(t, "account-1", values.Get("accountIDs[0]"))
	assert.Equal(t, "account-2", values.Get("accountIDs[1]"))
	assert.Equal(t, "chat-1", values.Get("chatIDs[0]"))
	assert.Equal(t, "25", values.Get("limit"))
	assert.Equal(t, "before", values.Get("direction"))
	assert.Equal(t, "true", values.Get("includeMuted"))
}

func TestMessagesSendPayload(t *testing.T) {
	type sendPayload struct {
		ChatID    string  `json:"chatID"`
		Text      string  `json:"text"`
		ReplyToID *string `json:"replyToId"`
	}

	var captured sendPayload

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		json.NewDecoder(r.Body).Decode(&captured)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resources.MessageSendResponse{MessageID: "msg_123", Success: true})
	}))
	defer server.Close()

	client, err := beeperdesktop.New(
		beeperdesktop.WithAccessToken("token"),
		beeperdesktop.WithBaseURL(server.URL),
		beeperdesktop.WithMaxRetries(0),
	)
	require.NoError(t, err)

	replyTo := "msg_parent"
	resp, err := client.Messages.Send(context.Background(), resources.MessageSendParams{
		ChatID:    "chat-123",
		Text:      "hello world",
		ReplyToID: &replyTo,
	})
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, "msg_123", resp.MessageID)

	assert.Equal(t, "chat-123", captured.ChatID)
	assert.Equal(t, "hello world", captured.Text)
	require.NotNil(t, captured.ReplyToID)
	assert.Equal(t, "msg_parent", *captured.ReplyToID)
}
