use crate::client::BeeperDesktop;
use crate::error::Result;
use crate::resources::shared::{Chat, Message, User};
use reqwest::Method;
use serde::{Deserialize, Serialize};

/// App handles app-related API operations
#[derive(Debug, Clone)]
pub struct App {
    client: BeeperDesktop,
}

impl App {
    /// Create a new App resource client
    pub fn new(client: BeeperDesktop) -> Self {
        Self { client }
    }

    /// DownloadAsset downloads an asset from a URL
    pub async fn download_asset(&self, params: &AppDownloadAssetParams) -> Result<AppDownloadAssetResponse> {
        self.client
            .do_request(Method::POST, "/v0/download-asset", Some(params))
            .await
    }

    /// Open opens Beeper Desktop and optionally navigates to a specific chat
    pub async fn open(&self, params: &AppOpenParams) -> Result<AppOpenResponse> {
        self.client
            .do_request(Method::POST, "/v0/open-app", Some(params))
            .await
    }

    /// Search searches for chats and messages in one call
    pub async fn search(&self, params: &AppSearchParams) -> Result<AppSearchResponse> {
        self.client
            .do_request(Method::GET, "/v0/search", Some(params))
            .await
    }
}

/// AppDownloadAssetParams represents parameters for downloading an asset
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct AppDownloadAssetParams {
    pub asset_url: String,
}

/// AppDownloadAssetResponse represents the response from downloading an asset
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct AppDownloadAssetResponse {
    pub local_path: String,
    pub success: bool,
    pub error: Option<String>,
}

/// AppOpenParams represents parameters for opening the app
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct AppOpenParams {
    pub chat_id: Option<String>,
    pub message_id: Option<String>,
    pub draft_text: Option<String>,
    pub draft_attachment: Option<String>,
}

/// AppOpenResponse represents the response from opening the app
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AppOpenResponse {
    pub success: bool,
    pub error: Option<String>,
}

/// AppSearchParams represents parameters for searching
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct AppSearchParams {
    pub query: String,
    pub account_ids: Option<Vec<String>>,
    pub chat_type: Option<String>,
    pub include_muted: Option<bool>,
    pub limit: Option<i32>,
    pub message_limit: Option<i32>,
    pub participant_limit: Option<i32>,
}

/// AppSearchResponse represents the response from searching
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AppSearchResponse {
    pub chats: Vec<ChatSearchResult>,
    pub messages: Vec<MessageSearchResult>,
}

/// ChatSearchResult represents a chat in search results
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ChatSearchResult {
    pub chat: Chat,
    pub participants: Vec<User>,
    pub messages: Vec<Message>,
}

/// MessageSearchResult represents a message in search results
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MessageSearchResult {
    pub message: Message,
    pub chat: Chat,
}