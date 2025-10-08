use crate::client::BeeperDesktop;
use crate::error::Result;
use crate::resources::shared::{MessagesCursor};
use crate::utils::slice_to_indexed_params;
use chrono::{DateTime, Utc};
use reqwest::Method;
use serde::{Deserialize, Serialize};

/// Messages handles message-related API operations
#[derive(Debug, Clone)]
pub struct Messages {
    client: BeeperDesktop,
}

impl Messages {
    /// Create a new Messages resource client
    pub fn new(client: BeeperDesktop) -> Self {
        Self { client }
    }

    /// Search searches messages across chats using Beeper's message index
    pub async fn search(&self, params: &MessageSearchParams) -> Result<MessagesCursor> {
        let mut query_params = Vec::new();

        // Handle account IDs
        if !params.account_ids.is_empty() {
            let account_params = slice_to_indexed_params("accountIDs", &params.account_ids);
            query_params.extend(account_params.into_iter().map(|(k, v)| (k, v)));
        }

        // Handle chat IDs
        if !params.chat_ids.is_empty() {
            let chat_params = slice_to_indexed_params("chatIDs", &params.chat_ids);
            query_params.extend(chat_params.into_iter().map(|(k, v)| (k, v)));
        }

        // Handle sender IDs
        if !params.sender_ids.is_empty() {
            let sender_params = slice_to_indexed_params("senderIDs", &params.sender_ids);
            query_params.extend(sender_params.into_iter().map(|(k, v)| (k, v)));
        }

        // Handle media types
        if !params.media_types.is_empty() {
            let media_params = slice_to_indexed_params("mediaTypes", &params.media_types);
            query_params.extend(media_params.into_iter().map(|(k, v)| (k, v)));
        }

        // Add other parameters
        if let Some(chat_type) = &params.chat_type {
            query_params.push(("chatType".to_string(), chat_type.clone()));
        }
        if let Some(cursor) = &params.cursor {
            query_params.push(("cursor".to_string(), cursor.clone()));
        }
        if let Some(date_after) = &params.date_after {
            query_params.push(("dateAfter".to_string(), date_after.to_rfc3339()));
        }
        if let Some(date_before) = &params.date_before {
            query_params.push(("dateBefore".to_string(), date_before.to_rfc3339()));
        }
        if let Some(direction) = &params.direction {
            query_params.push(("direction".to_string(), direction.clone()));
        }
        if let Some(exclude_low_priority) = params.exclude_low_priority {
            query_params.push(("excludeLowPriority".to_string(), exclude_low_priority.to_string()));
        }
        if let Some(include_muted) = params.include_muted {
            query_params.push(("includeMuted".to_string(), include_muted.to_string()));
        }
        if let Some(limit) = params.limit {
            query_params.push(("limit".to_string(), limit.to_string()));
        }
        if let Some(query) = &params.query {
            query_params.push(("query".to_string(), query.clone()));
        }

        // Convert to the format expected by do_request_with_query
        let query_refs: Vec<(&str, &str)> = query_params
            .iter()
            .map(|(k, v)| (k.as_str(), v.as_str()))
            .collect();

        self.client
            .do_request_with_query(Method::GET, "/v0/search-messages", &query_refs)
            .await
    }

    /// Send sends a text message to a specific chat
    pub async fn send(&self, params: &MessageSendParams) -> Result<MessageSendResponse> {
        self.client
            .do_request(Method::POST, "/v0/send-message", Some(params))
            .await
    }
}

/// MessageSearchParams represents parameters for searching messages
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct MessageSearchParams {
    pub account_ids: Vec<String>,
    pub chat_ids: Vec<String>,
    pub chat_type: Option<String>,
    pub cursor: Option<String>,
    pub date_after: Option<DateTime<Utc>>,
    pub date_before: Option<DateTime<Utc>>,
    pub direction: Option<String>,
    pub exclude_low_priority: Option<bool>,
    pub include_muted: Option<bool>,
    pub limit: Option<i32>,
    pub media_types: Vec<String>,
    pub query: Option<String>,
    pub sender_ids: Vec<String>,
}

impl MessageSearchParams {
    /// Create a new MessageSearchParams with default values
    pub fn new() -> Self {
        Self::default()
    }
}

/// MessageSendParams represents parameters for sending a message
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct MessageSendParams {
    pub chat_id: String,
    pub text: String,
    pub reply_to_id: Option<String>,
    pub attachment: Option<String>,
}

/// MessageSendResponse represents the response from sending a message
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct MessageSendResponse {
    pub message_id: String,
    pub deeplink: String,
    pub success: bool,
    pub error: Option<String>,
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::{Config, BeeperDesktop};
    use crate::resources::shared::{Message, User, PaginationInfo, Cursor};
    use chrono::{DateTime, Utc};
    use wiremock::{
        matchers::{header, method, path, query_param},
        Mock, MockServer, ResponseTemplate,
    };
    use serde_json::json;
    use std::time::Duration;

    async fn setup_mock_server() -> (MockServer, BeeperDesktop) {
        let mock_server = MockServer::start().await;
        
        let config = Config::builder()
            .access_token("test-token")
            .base_url(mock_server.uri())
            .timeout(Duration::from_secs(5))
            .max_retries(0)
            .build()
            .unwrap();
        
        let client = BeeperDesktop::with_config(config).await.unwrap();
        
        (mock_server, client)
    }

    #[tokio::test]
    async fn test_message_search_query_encoding() {
        let (mock_server, client) = setup_mock_server().await;

        Mock::given(method("GET"))
            .and(path("/v0/search-messages"))
            .and(query_param("accountIDs[0]", "account-1"))
            .and(query_param("accountIDs[1]", "account-2"))
            .and(query_param("chatIDs[0]", "chat-1"))
            .and(query_param("limit", "25"))
            .and(query_param("direction", "before"))
            .and(query_param("includeMuted", "true"))
            .respond_with(ResponseTemplate::new(200).set_body_json(json!({
                "items": [],
                "pagination": null
            })))
            .mount(&mock_server)
            .await;

        let mut params = MessageSearchParams::new();
        params.account_ids = vec!["account-1".to_string(), "account-2".to_string()];
        params.chat_ids = vec!["chat-1".to_string()];
        params.limit = Some(25);
        params.direction = Some("before".to_string());
        params.include_muted = Some(true);

        let result = client.messages().search(&params).await;
        assert!(result.is_ok());
    }

    #[tokio::test]
    async fn test_message_send_payload() {
        let (mock_server, client) = setup_mock_server().await;

        let expected_response = MessageSendResponse {
            message_id: "msg_123".to_string(),
            deeplink: "https://beeper.com/chat/123".to_string(),
            success: true,
            error: None,
        };

        Mock::given(method("POST"))
            .and(path("/v0/send-message"))
            .and(header("authorization", "Bearer test-token"))
            .respond_with(ResponseTemplate::new(200).set_body_json(&expected_response))
            .mount(&mock_server)
            .await;

        let send_params = MessageSendParams {
            chat_id: "chat-123".to_string(),
            text: "hello world".to_string(),
            reply_to_id: Some("msg_parent".to_string()),
            attachment: None,
        };

        let response = client.messages().send(&send_params).await.unwrap();
        
        assert!(response.success);
        assert_eq!(response.message_id, "msg_123");
        assert_eq!(response.deeplink, "https://beeper.com/chat/123");
    }

    #[tokio::test]
    async fn test_message_search_with_pagination() {
        let (mock_server, client) = setup_mock_server().await;

        let mock_message = Message {
            id: "msg_1".to_string(),
            account_id: "account_1".to_string(),
            chat_id: "chat_1".to_string(),
            message_id: "msg_1".to_string(),
            sender_id: "user_1".to_string(),
            sort_key: json!("1234567890"),
            timestamp: Utc::now(),
            attachments: None,
            is_sender: Some(false),
            is_unread: Some(true),
            reactions: None,
            sender_name: Some("Test User".to_string()),
            text: Some("Hello world".to_string()),
        };

        let expected_cursor = MessagesCursor {
            items: vec![mock_message],
            pagination: Some(PaginationInfo {
                cursor: Some("next_cursor".to_string()),
                limit: Some(10),
                direction: Some("after".to_string()),
                has_more: true,
            }),
        };

        Mock::given(method("GET"))
            .and(path("/v0/search-messages"))
            .and(query_param("query", "hello"))
            .and(query_param("limit", "10"))
            .respond_with(ResponseTemplate::new(200).set_body_json(&expected_cursor))
            .mount(&mock_server)
            .await;

        let mut params = MessageSearchParams::new();
        params.query = Some("hello".to_string());
        params.limit = Some(10);

        let result = client.messages().search(&params).await.unwrap();
        
        assert_eq!(result.items.len(), 1);
        assert_eq!(result.items[0].text, Some("Hello world".to_string()));
        assert!(result.pagination.is_some());
        
        let pagination = result.pagination.unwrap();
        assert_eq!(pagination.cursor, Some("next_cursor".to_string()));
        assert!(pagination.has_more);
    }
}