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

func TestChatsSearchPayload(t *testing.T) {
	var capturedURL *url.URL
	var capturedBody map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL
		defer r.Body.Close()
		_ = json.NewDecoder(r.Body).Decode(&capturedBody)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resources.ChatsCursor{})
	}))
	defer server.Close()

	client, err := beeperdesktop.New(
		beeperdesktop.WithAccessToken("token"),
		beeperdesktop.WithBaseURL(server.URL),
		beeperdesktop.WithMaxRetries(0),
	)
	require.NoError(t, err)

	_, err = client.Chats.Search(context.Background(), resources.ChatSearchParams{
		AccountIDs:   []string{"account-1"},
		IncludeMuted: beeperdesktop.BoolPtr(true),
		Limit:        beeperdesktop.IntPtr(10),
		Scope:        beeperdesktop.StringPtr("titles"),
		Query:        beeperdesktop.StringPtr("updates"),
	})
	require.NoError(t, err)

	require.NotNil(t, capturedURL)
	assert.Equal(t, "/v0/search-chats", capturedURL.Path)

	require.NotNil(t, capturedBody)

	accountIDs, ok := capturedBody["accountIDs"].([]interface{})
	require.True(t, ok)
	require.Len(t, accountIDs, 1)
	assert.Equal(t, "account-1", accountIDs[0])

	assert.Equal(t, true, capturedBody["includeMuted"])
	assert.Equal(t, float64(10), capturedBody["limit"])
	assert.Equal(t, "titles", capturedBody["scope"])
	assert.Equal(t, "updates", capturedBody["query"])
}

func TestChatsCreatePayload(t *testing.T) {
	type createPayload struct {
		AccountID      string   `json:"accountID"`
		ParticipantIDs []string `json:"participantIDs"`
		Type           string   `json:"type"`
		Title          *string  `json:"title"`
	}

	var captured createPayload

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		json.NewDecoder(r.Body).Decode(&captured)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resources.ChatCreateResponse{Success: true, Chat: resources.Chat{ID: "chat-1"}})
	}))
	defer server.Close()

	client, err := beeperdesktop.New(
		beeperdesktop.WithAccessToken("token"),
		beeperdesktop.WithBaseURL(server.URL),
		beeperdesktop.WithMaxRetries(0),
	)
	require.NoError(t, err)

	title := "Project Updates"
	resp, err := client.Chats.Create(context.Background(), resources.ChatCreateParams{
		AccountID:      "account-1",
		ParticipantIDs: []string{"user-1", "user-2"},
		Type:           "group",
		Title:          &title,
	})
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, "chat-1", resp.Chat.ID)

	assert.Equal(t, "account-1", captured.AccountID)
	assert.Equal(t, []string{"user-1", "user-2"}, captured.ParticipantIDs)
	assert.Equal(t, "group", captured.Type)
	require.NotNil(t, captured.Title)
	assert.Equal(t, "Project Updates", *captured.Title)
}
