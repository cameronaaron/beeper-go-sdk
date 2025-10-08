use crate::client::BeeperDesktop;
use crate::error::Result;
use reqwest::Method;
use serde::{Deserialize, Serialize};

/// Token handles token-related API operations
#[derive(Debug, Clone)]
pub struct Token {
    client: BeeperDesktop,
}

impl Token {
    /// Create a new Token resource client
    pub fn new(client: BeeperDesktop) -> Self {
        Self { client }
    }

    /// Info returns information about the authenticated user/token
    pub async fn info(&self) -> Result<UserInfo> {
        self.client
            .do_request(Method::GET, "/oauth/userinfo", None::<&()>)
            .await
    }
}

/// RevokeRequest represents a token revocation request
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RevokeRequest {
    pub token: String,
    pub token_type_hint: Option<String>,
}

/// UserInfo represents information about the authenticated user/token
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct UserInfo {
    /// Issued at timestamp (Unix epoch seconds)
    pub iat: i64,
    /// Granted scopes
    pub scope: String,
    /// Subject identifier (token ID)
    pub sub: String,
    /// Token type
    pub token_use: String,
    /// Audience (client ID)
    pub aud: Option<String>,
    /// Client identifier
    pub client_id: Option<String>,
    /// Expiration timestamp (Unix epoch seconds)
    pub exp: Option<i64>,
}