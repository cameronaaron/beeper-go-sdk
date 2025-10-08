use crate::client::BeeperDesktop;
use crate::error::Result;
use crate::resources::shared::User;
use reqwest::Method;
use serde::{Deserialize, Serialize};

/// Contacts handles contact-related API operations
#[derive(Debug, Clone)]
pub struct Contacts {
    client: BeeperDesktop,
}

impl Contacts {
    /// Create a new Contacts resource client
    pub fn new(client: BeeperDesktop) -> Self {
        Self { client }
    }

    /// Search searches for contacts/users
    pub async fn search(&self, params: &ContactSearchParams) -> Result<ContactSearchResponse> {
        let query_params = vec![
            ("accountID", params.account_id.as_str()),
            ("query", params.query.as_str()),
        ];

        self.client
            .do_request_with_query(Method::GET, "/v0/search-users", &query_params)
            .await
    }
}

/// ContactSearchParams represents parameters for searching contacts
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct ContactSearchParams {
    pub account_id: String,
    pub query: String,
}

/// ContactSearchResponse represents the response from searching contacts
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ContactSearchResponse {
    pub items: Vec<User>,
}