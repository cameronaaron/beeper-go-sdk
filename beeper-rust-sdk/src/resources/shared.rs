use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use std::collections::HashMap;

/// Attachment represents a file attachment in a message
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct Attachment {
    /// Type of attachment (unknown, img, video, audio)
    #[serde(rename = "type")]
    pub attachment_type: String,
    pub duration: Option<i32>,
    pub file_name: Option<String>,
    pub file_size: Option<i64>,
    pub is_gif: Option<bool>,
    pub is_sticker: Option<bool>,
    pub is_voice_note: Option<bool>,
    pub mime_type: Option<String>,
    pub poster_img: Option<String>,
    pub size: Option<AttachmentSize>,
    pub src_url: Option<String>,
}

/// AttachmentSize represents pixel dimensions of an attachment
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AttachmentSize {
    pub height: Option<i32>,
    pub width: Option<i32>,
}

/// BaseResponse represents a basic API response
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct BaseResponse {
    pub success: bool,
    pub error: Option<String>,
}

/// Error represents an API error response
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ErrorResponse {
    pub error: String,
    pub code: Option<String>,
    pub details: Option<HashMap<String, String>>,
}

/// Message represents a chat message
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Message {
    pub id: String,
    #[serde(rename = "accountID")]
    pub account_id: String,
    #[serde(rename = "chatID")]
    pub chat_id: String,
    #[serde(rename = "messageID")]
    pub message_id: String,
    #[serde(rename = "senderID")]
    pub sender_id: String,
    #[serde(rename = "sortKey")]
    pub sort_key: serde_json::Value, // Can be string or number
    pub timestamp: DateTime<Utc>,
    pub attachments: Option<Vec<Attachment>>,
    #[serde(rename = "isSender")]
    pub is_sender: Option<bool>,
    #[serde(rename = "isUnread")]
    pub is_unread: Option<bool>,
    pub reactions: Option<Vec<Reaction>>,
    #[serde(rename = "senderName")]
    pub sender_name: Option<String>,
    pub text: Option<String>,
}

/// Reaction represents a message reaction
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct Reaction {
    pub id: String,
    pub participant_id: String,
    pub reaction_key: String,
    pub emoji: Option<bool>,
    pub img_url: Option<String>,
}

/// User represents a person on or reachable through Beeper
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct User {
    pub id: String,
    #[serde(rename = "cannotMessage")]
    pub cannot_message: Option<bool>,
    pub email: Option<String>,
    #[serde(rename = "fullName")]
    pub full_name: Option<String>,
    #[serde(rename = "imgURL")]
    pub img_url: Option<String>,
    #[serde(rename = "isSelf")]
    pub is_self: Option<bool>,
    #[serde(rename = "phoneNumber")]
    pub phone_number: Option<String>,
    pub username: Option<String>,
}

/// Cursor represents a paginated response
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Cursor<T> {
    pub items: Vec<T>,
    pub pagination: Option<PaginationInfo>,
}

/// PaginationInfo contains pagination metadata
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PaginationInfo {
    pub cursor: Option<String>,
    pub limit: Option<i32>,
    pub direction: Option<String>,
    pub has_more: bool,
}

/// MessagesCursor is a type alias for message pagination
pub type MessagesCursor = Cursor<Message>;

/// ChatsCursor represents paginated chat results
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ChatsCursor {
    pub items: Vec<Chat>,
    pub pagination: Option<PaginationInfo>,
}

/// Chat represents a chat/conversation
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Chat {
    pub id: String,
    #[serde(rename = "accountID")]
    pub account_id: String,
    pub network: String,
    pub title: String,
    /// Type of chat (single, group)
    #[serde(rename = "type")]
    pub chat_type: String,
    #[serde(rename = "unreadCount")]
    pub unread_count: i32,
    pub participants: ChatParticipants,
    #[serde(rename = "isArchived")]
    pub is_archived: Option<bool>,
    #[serde(rename = "isMuted")]
    pub is_muted: Option<bool>,
    #[serde(rename = "isPinned")]
    pub is_pinned: Option<bool>,
    #[serde(rename = "lastActivity")]
    pub last_activity: Option<String>,
    #[serde(rename = "lastReadMessageSortKey")]
    pub last_read_message_sort_key: Option<serde_json::Value>, // Can be string or number
    #[serde(rename = "localChatID")]
    pub local_chat_id: Option<String>,
}

/// ChatParticipants represents chat participants information
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ChatParticipants {
    #[serde(rename = "hasMore")]
    pub has_more: bool,
    pub items: Vec<User>,
    pub total: i32,
}

/// Account represents a chat account added to Beeper
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Account {
    #[serde(rename = "accountID")]
    pub account_id: String,
    pub network: String,
    pub user: User,
}