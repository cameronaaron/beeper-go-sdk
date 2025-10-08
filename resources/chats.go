package resources

import (
	"context"
	"time"
)

// Chats handles chat-related API operations
type Chats struct {
	client    ClientInterface
	Reminders *Reminders
}

// NewChats creates a new Chats resource client
func NewChats(client ClientInterface) *Chats {
	return &Chats{
		client:    client,
		Reminders: NewReminders(client),
	}
}

// Chat represents a chat/conversation
type Chat struct {
	ID                     string           `json:"id"`
	AccountID              string           `json:"accountID"`
	Network                string           `json:"network"`
	Title                  string           `json:"title"`
	Type                   string           `json:"type"` // single, group
	UnreadCount            int              `json:"unreadCount"`
	Participants           ChatParticipants `json:"participants"`
	IsArchived             *bool            `json:"isArchived,omitempty"`
	IsMuted                *bool            `json:"isMuted,omitempty"`
	IsPinned               *bool            `json:"isPinned,omitempty"`
	LastActivity           *string          `json:"lastActivity,omitempty"`
	LastReadMessageSortKey interface{}      `json:"lastReadMessageSortKey,omitempty"` // string or number
	LocalChatID            *string          `json:"localChatID,omitempty"`
}

// ChatParticipants represents chat participants information
type ChatParticipants struct {
	HasMore bool   `json:"hasMore"`
	Items   []User `json:"items"`
	Total   int    `json:"total"`
}

// ChatCreateParams represents parameters for creating a chat
type ChatCreateParams struct {
	AccountID      string   `json:"accountID"`
	ParticipantIDs []string `json:"participantIDs"`
	Type           string   `json:"type"` // single, group
	Title          *string  `json:"title,omitempty"`
}

// ChatCreateResponse represents the response from creating a chat
type ChatCreateResponse struct {
	Chat    Chat   `json:"chat"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// ChatRetrieveParams represents parameters for retrieving a chat
type ChatRetrieveParams struct {
	ChatID string `json:"chatID"`
}

// ChatArchiveParams represents parameters for archiving a chat
type ChatArchiveParams struct {
	ChatID   string `json:"chatID"`
	Archived bool   `json:"archived"`
}

// ChatSearchParams represents parameters for searching chats
type ChatSearchParams struct {
	AccountIDs   []string `json:"accountIDs,omitempty"`
	ChatType     *string  `json:"chatType,omitempty"`
	IncludeMuted *bool    `json:"includeMuted,omitempty"`
	Limit        *int     `json:"limit,omitempty"`
	Cursor       *string  `json:"cursor,omitempty"`
	Scope        *string  `json:"scope,omitempty"`
	Query        *string  `json:"query,omitempty"`
}

// ChatsCursor represents paginated chat results
type ChatsCursor = Cursor[Chat]

// Create creates a single or group chat on a specific account
func (c *Chats) Create(ctx context.Context, params ChatCreateParams) (*ChatCreateResponse, error) {
	var result ChatCreateResponse
	err := c.client.DoRequest(ctx, "POST", "/v0/create-chat", params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Retrieve gets chat details including metadata, participants, and latest message
func (c *Chats) Retrieve(ctx context.Context, params ChatRetrieveParams) (*Chat, error) {
	var result Chat
	err := c.client.DoRequest(ctx, "GET", "/v0/get-chat", params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Archive archives or unarchives a chat
func (c *Chats) Archive(ctx context.Context, params ChatArchiveParams) (*BaseResponse, error) {
	var result BaseResponse
	err := c.client.DoRequest(ctx, "POST", "/v0/archive-chat", params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Search searches chats by title/network or participants
func (c *Chats) Search(ctx context.Context, params ChatSearchParams) (*ChatsCursor, error) {
	var result ChatsCursor
	err := c.client.DoRequest(ctx, "GET", "/v0/search-chats", params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Reminders handles chat reminder operations
type Reminders struct {
	client ClientInterface
}

// NewReminders creates a new Reminders resource client
func NewReminders(client ClientInterface) *Reminders {
	return &Reminders{client: client}
}

// ReminderCreateParams represents parameters for creating a reminder
type ReminderCreateParams struct {
	ChatID    string    `json:"chatID"`
	Timestamp time.Time `json:"timestamp"`
	Message   *string   `json:"message,omitempty"`
}

// ReminderDeleteParams represents parameters for deleting a reminder
type ReminderDeleteParams struct {
	ChatID string `json:"chatID"`
}

// Create sets a reminder for a chat at a specific time
func (r *Reminders) Create(ctx context.Context, params ReminderCreateParams) (*BaseResponse, error) {
	var result BaseResponse
	err := r.client.DoRequest(ctx, "POST", "/v0/set-chat-reminder", params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete clears a chat reminder
func (r *Reminders) Delete(ctx context.Context, params ReminderDeleteParams) (*BaseResponse, error) {
	var result BaseResponse
	err := r.client.DoRequest(ctx, "POST", "/v0/clear-chat-reminder", params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
