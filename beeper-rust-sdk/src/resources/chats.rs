use crate::client::BeeperDesktop;
use crate::error::Result;
use crate::resources::shared::{Chat, ChatsCursor, BaseResponse};
use chrono::{DateTime, Utc};
use reqwest::Method;
use serde::{Deserialize, Serialize};

/// Chats handles chat-related API operations
#[derive(Debug, Clone)]
pub struct Chats {
    client: BeeperDesktop,
    pub reminders: Reminders,
}

impl Chats {
    /// Create a new Chats resource client
    pub fn new(client: BeeperDesktop) -> Self {
        let reminders = Reminders::new(client.clone());
        Self { client, reminders }
    }

    /// Create creates a single or group chat on a specific account
    pub async fn create(&self, params: &ChatCreateParams) -> Result<ChatCreateResponse> {
        self.client
            .do_request(Method::POST, "/v0/create-chat", Some(params))
            .await
    }

    /// Retrieve gets chat details including metadata, participants, and latest message
    pub async fn retrieve(&self, params: &ChatRetrieveParams) -> Result<Chat> {
        self.client
            .do_request(Method::GET, "/v0/get-chat", Some(params))
            .await
    }

    /// Archive archives or unarchives a chat
    pub async fn archive(&self, params: &ChatArchiveParams) -> Result<BaseResponse> {
        self.client
            .do_request(Method::POST, "/v0/archive-chat", Some(params))
            .await
    }

    /// Search searches chats by title/network or participants
    pub async fn search(&self, params: &ChatSearchParams) -> Result<ChatsCursor> {
        self.client
            .do_request(Method::GET, "/v0/search-chats", Some(params))
            .await
    }
}

/// Reminders handles chat reminder operations
#[derive(Debug, Clone)]
pub struct Reminders {
    client: BeeperDesktop,
}

impl Reminders {
    /// Create a new Reminders resource client
    pub fn new(client: BeeperDesktop) -> Self {
        Self { client }
    }

    /// Create sets a reminder for a chat at a specific time
    pub async fn create(&self, params: &ReminderCreateParams) -> Result<BaseResponse> {
        self.client
            .do_request(Method::POST, "/v0/set-chat-reminder", Some(params))
            .await
    }

    /// Delete clears a chat reminder
    pub async fn delete(&self, params: &ReminderDeleteParams) -> Result<BaseResponse> {
        self.client
            .do_request(Method::POST, "/v0/clear-chat-reminder", Some(params))
            .await
    }
}

/// ChatCreateParams represents parameters for creating a chat
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct ChatCreateParams {
    pub account_id: String,
    pub participant_ids: Vec<String>,
    /// Type of chat (single, group)
    #[serde(rename = "type")]
    pub chat_type: String,
    pub title: Option<String>,
}

/// ChatCreateResponse represents the response from creating a chat
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ChatCreateResponse {
    pub chat: Chat,
    pub success: bool,
    pub error: Option<String>,
}

/// ChatRetrieveParams represents parameters for retrieving a chat
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct ChatRetrieveParams {
    pub chat_id: String,
}

/// ChatArchiveParams represents parameters for archiving a chat
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct ChatArchiveParams {
    pub chat_id: String,
    pub archived: bool,
}

/// ChatSearchParams represents parameters for searching chats
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct ChatSearchParams {
    pub account_ids: Option<Vec<String>>,
    pub chat_type: Option<String>,
    pub include_muted: Option<bool>,
    pub limit: Option<i32>,
    pub cursor: Option<String>,
    pub scope: Option<String>,
    pub query: Option<String>,
}

impl ChatSearchParams {
    /// Create a new ChatSearchParams with default values
    pub fn new() -> Self {
        Self::default()
    }
}

/// ReminderCreateParams represents parameters for creating a reminder
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct ReminderCreateParams {
    pub chat_id: String,
    pub timestamp: DateTime<Utc>,
    pub message: Option<String>,
}

/// ReminderDeleteParams represents parameters for deleting a reminder
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct ReminderDeleteParams {
    pub chat_id: String,
}