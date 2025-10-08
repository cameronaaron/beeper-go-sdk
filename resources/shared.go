package resources

import "time"

// Attachment represents a file attachment in a message
type Attachment struct {
	Type        string          `json:"type"` // unknown, img, video, audio
	Duration    *int            `json:"duration,omitempty"`
	FileName    *string         `json:"fileName,omitempty"`
	FileSize    *int64          `json:"fileSize,omitempty"`
	IsGif       *bool           `json:"isGif,omitempty"`
	IsSticker   *bool           `json:"isSticker,omitempty"`
	IsVoiceNote *bool           `json:"isVoiceNote,omitempty"`
	MimeType    *string         `json:"mimeType,omitempty"`
	PosterImg   *string         `json:"posterImg,omitempty"`
	Size        *AttachmentSize `json:"size,omitempty"`
	SrcURL      *string         `json:"srcURL,omitempty"`
}

// AttachmentSize represents pixel dimensions of an attachment
type AttachmentSize struct {
	Height *int `json:"height,omitempty"`
	Width  *int `json:"width,omitempty"`
}

// BaseResponse represents a basic API response
type BaseResponse struct {
	Success bool    `json:"success"`
	Error   *string `json:"error,omitempty"`
}

// Error represents an API error response
type ErrorResponse struct {
	Error   string            `json:"error"`
	Code    *string           `json:"code,omitempty"`
	Details map[string]string `json:"details,omitempty"`
}

// Message represents a chat message
type Message struct {
	ID          string       `json:"id"`
	AccountID   string       `json:"accountID"`
	ChatID      string       `json:"chatID"`
	MessageID   string       `json:"messageID"`
	SenderID    string       `json:"senderID"`
	SortKey     interface{}  `json:"sortKey"` // string or number
	Timestamp   time.Time    `json:"timestamp"`
	Attachments []Attachment `json:"attachments,omitempty"`
	IsSender    *bool        `json:"isSender,omitempty"`
	IsUnread    *bool        `json:"isUnread,omitempty"`
	Reactions   []Reaction   `json:"reactions,omitempty"`
	SenderName  *string      `json:"senderName,omitempty"`
	Text        *string      `json:"text,omitempty"`
}

// Reaction represents a message reaction
type Reaction struct {
	ID            string  `json:"id"`
	ParticipantID string  `json:"participantID"`
	ReactionKey   string  `json:"reactionKey"`
	Emoji         *bool   `json:"emoji,omitempty"`
	ImgURL        *string `json:"imgURL,omitempty"`
}

// User represents a person on or reachable through Beeper
type User struct {
	ID            string  `json:"id"`
	CannotMessage *bool   `json:"cannotMessage,omitempty"`
	Email         *string `json:"email,omitempty"`
	FullName      *string `json:"fullName,omitempty"`
	ImgURL        *string `json:"imgURL,omitempty"`
	IsSelf        *bool   `json:"isSelf,omitempty"`
	PhoneNumber   *string `json:"phoneNumber,omitempty"`
	Username      *string `json:"username,omitempty"`
}

// Cursor represents a paginated response
type Cursor[T any] struct {
	Items      []T             `json:"items"`
	Pagination *PaginationInfo `json:"pagination,omitempty"`
}

// PaginationInfo contains pagination metadata
type PaginationInfo struct {
	Cursor    *string `json:"cursor,omitempty"`
	Limit     *int    `json:"limit,omitempty"`
	Direction *string `json:"direction,omitempty"`
	HasMore   bool    `json:"has_more"`
}

// MessagesCursor is a type alias for message pagination
type MessagesCursor = Cursor[Message]
