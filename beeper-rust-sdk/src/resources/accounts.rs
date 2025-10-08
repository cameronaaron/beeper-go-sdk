use crate::client::BeeperDesktop;
use crate::error::Result;
use crate::resources::shared::Account;
use reqwest::Method;

/// Accounts handles account-related API operations
#[derive(Debug, Clone)]
pub struct Accounts {
    client: BeeperDesktop,
}

impl Accounts {
    /// Create a new Accounts resource client
    pub fn new(client: BeeperDesktop) -> Self {
        Self { client }
    }

    /// List retrieves all connected Beeper accounts available on this device
    pub async fn list(&self) -> Result<Vec<Account>> {
        self.client
            .do_request(Method::GET, "/v0/get-accounts", None::<&()>)
            .await
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::resources::shared::{Account, User};
    use crate::{Config, BeeperDesktop};
    use wiremock::{
        matchers::{header, method, path},
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
            .max_retries(0) // Disable retries for tests
            .build()
            .unwrap();
        
        let client = BeeperDesktop::with_config(config).await.unwrap();
        
        (mock_server, client)
    }

    #[tokio::test]
    async fn test_accounts_list_success() {
        let (mock_server, client) = setup_mock_server().await;

        let expected_accounts = vec![
            Account {
                account_id: "account1".to_string(),
                network: "discord".to_string(),
                user: User {
                    id: "user1".to_string(),
                    cannot_message: None,
                    email: Some("user1@example.com".to_string()),
                    full_name: Some("User One".to_string()),
                    img_url: None,
                    is_self: Some(true),
                    phone_number: None,
                    username: Some("user1".to_string()),
                },
            },
            Account {
                account_id: "account2".to_string(),
                network: "telegram".to_string(),
                user: User {
                    id: "user2".to_string(),
                    cannot_message: None,
                    email: None,
                    full_name: Some("User Two".to_string()),
                    img_url: Some("https://example.com/avatar.png".to_string()),
                    is_self: Some(false),
                    phone_number: Some("+1234567890".to_string()),
                    username: None,
                },
            },
        ];

        Mock::given(method("GET"))
            .and(path("/v0/get-accounts"))
            .and(header("authorization", "Bearer test-token"))
            .respond_with(ResponseTemplate::new(200).set_body_json(&expected_accounts))
            .mount(&mock_server)
            .await;

        let accounts = client.accounts().list().await.unwrap();

        assert_eq!(accounts.len(), 2);
        assert_eq!(accounts[0].account_id, "account1");
        assert_eq!(accounts[0].network, "discord");
        assert_eq!(accounts[0].user.full_name, Some("User One".to_string()));
        assert_eq!(accounts[1].account_id, "account2");
        assert_eq!(accounts[1].network, "telegram");
        assert_eq!(accounts[1].user.phone_number, Some("+1234567890".to_string()));
    }

    #[tokio::test]
    async fn test_accounts_list_authentication_error() {
        let (mock_server, client) = setup_mock_server().await;

        Mock::given(method("GET"))
            .and(path("/v0/get-accounts"))
            .respond_with(ResponseTemplate::new(401).set_body_json(json!({
                "error": "Invalid token",
                "code": "INVALID_TOKEN"
            })))
            .mount(&mock_server)
            .await;

        let result = client.accounts().list().await;

        assert!(result.is_err());
        match result.unwrap_err() {
            crate::Error::Authentication { message, code, .. } => {
                assert_eq!(message, "Invalid token");
                assert_eq!(code, Some("INVALID_TOKEN".to_string()));
            }
            _ => panic!("Expected authentication error"),
        }
    }

    #[tokio::test]
    async fn test_accounts_list_empty() {
        let (mock_server, client) = setup_mock_server().await;

        Mock::given(method("GET"))
            .and(path("/v0/get-accounts"))
            .and(header("authorization", "Bearer test-token"))
            .respond_with(ResponseTemplate::new(200).set_body_json(Vec::<Account>::new()))
            .mount(&mock_server)
            .await;

        let accounts = client.accounts().list().await.unwrap();

        assert_eq!(accounts.len(), 0);
    }

    #[tokio::test]
    async fn test_accounts_list_server_error() {
        let (mock_server, client) = setup_mock_server().await;

        Mock::given(method("GET"))
            .and(path("/v0/get-accounts"))
            .respond_with(ResponseTemplate::new(500).set_body_json(json!({
                "error": "Internal server error"
            })))
            .mount(&mock_server)
            .await;

        let result = client.accounts().list().await;

        assert!(result.is_err());
        match result.unwrap_err() {
            crate::Error::InternalServer { message, .. } => {
                assert_eq!(message, "Internal server error");
            }
            _ => panic!("Expected internal server error"),
        }
    }
}