package resources

import (
	"context"
	"net/url"
	"strconv"
	"time"
)

// Messages handles message-related API operations
type Messages struct {
	client ClientInterface
}

// NewMessages creates a new Messages resource client
func NewMessages(client ClientInterface) *Messages {
	return &Messages{client: client}
}

// MessageSearchParams represents parameters for searching messages
type MessageSearchParams struct {
	AccountIDs         []string   `json:"accountIDs,omitempty"`
	ChatIDs            []string   `json:"chatIDs,omitempty"`
	ChatType           *string    `json:"chatType,omitempty"`
	Cursor             *string    `json:"cursor,omitempty"`
	DateAfter          *time.Time `json:"dateAfter,omitempty"`
	DateBefore         *time.Time `json:"dateBefore,omitempty"`
	Direction          *string    `json:"direction,omitempty"`
	ExcludeLowPriority *bool      `json:"excludeLowPriority,omitempty"`
	IncludeMuted       *bool      `json:"includeMuted,omitempty"`
	Limit              *int       `json:"limit,omitempty"`
	MediaTypes         []string   `json:"mediaTypes,omitempty"`
	Query              *string    `json:"query,omitempty"`
	SenderIDs          []string   `json:"senderIDs,omitempty"`
}

// MessageSendParams represents parameters for sending a message
type MessageSendParams struct {
	ChatID     string  `json:"chatID"`
	Text       string  `json:"text"`
	ReplyToID  *string `json:"replyToId,omitempty"`
	Attachment *string `json:"attachment,omitempty"`
}

// MessageSendResponse represents the response from sending a message
type MessageSendResponse struct {
	MessageID string `json:"messageID"`
	Deeplink  string `json:"deeplink"`
	Success   bool   `json:"success"`
	Error     string `json:"error,omitempty"`
}

// Search searches messages across chats using Beeper's message index
func (m *Messages) Search(ctx context.Context, params MessageSearchParams) (*MessagesCursor, error) {
	var result MessagesCursor
	path := "/v0/search-messages"

	query := url.Values{}

	if len(params.AccountIDs) > 0 {
		for idx, id := range params.AccountIDs {
			query.Add("accountIDs["+strconv.Itoa(idx)+"]", id)
		}
	}

	if len(params.ChatIDs) > 0 {
		for idx, id := range params.ChatIDs {
			query.Add("chatIDs["+strconv.Itoa(idx)+"]", id)
		}
	}

	if params.ChatType != nil {
		query.Set("chatType", *params.ChatType)
	}

	if params.Cursor != nil {
		query.Set("cursor", *params.Cursor)
	}

	if params.DateAfter != nil {
		query.Set("dateAfter", params.DateAfter.Format(time.RFC3339Nano))
	}

	if params.DateBefore != nil {
		query.Set("dateBefore", params.DateBefore.Format(time.RFC3339Nano))
	}

	if params.Direction != nil {
		query.Set("direction", *params.Direction)
	}

	if params.ExcludeLowPriority != nil {
		query.Set("excludeLowPriority", strconv.FormatBool(*params.ExcludeLowPriority))
	}

	if params.IncludeMuted != nil {
		query.Set("includeMuted", strconv.FormatBool(*params.IncludeMuted))
	}

	if params.Limit != nil {
		query.Set("limit", strconv.Itoa(*params.Limit))
	}

	if len(params.MediaTypes) > 0 {
		for idx, mediaType := range params.MediaTypes {
			query.Add("mediaTypes["+strconv.Itoa(idx)+"]", mediaType)
		}
	}

	if params.Query != nil {
		query.Set("query", *params.Query)
	}

	if len(params.SenderIDs) > 0 {
		for idx, senderID := range params.SenderIDs {
			query.Add("senderIDs["+strconv.Itoa(idx)+"]", senderID)
		}
	}

	if len(query) > 0 {
		path += "?" + query.Encode()
	}

	err := m.client.DoRequest(ctx, "GET", path, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Send sends a text message to a specific chat
func (m *Messages) Send(ctx context.Context, params MessageSendParams) (*MessageSendResponse, error) {
	var result MessageSendResponse
	err := m.client.DoRequest(ctx, "POST", "/v0/send-message", params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
