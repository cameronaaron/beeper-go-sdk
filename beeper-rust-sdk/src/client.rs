use crate::config::Config;
use crate::error::{Error, Result};
use crate::resources::{Accounts, App, Chats, Contacts, Messages, Token};
use reqwest::{Client as HttpClient, Method, Response, StatusCode};
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::sync::Arc;
use std::time::Duration;
use tracing::{debug, warn};
use url::Url;

/// Error response from the API
#[derive(Debug, Deserialize)]
struct ErrorResponse {
    error: Option<String>,
    code: Option<String>,
    details: Option<HashMap<String, String>>,
}

/// Main API client for the Beeper Desktop API
#[derive(Debug, Clone)]
pub struct BeeperDesktop {
    config: Arc<Config>,
    http_client: HttpClient,
    base_url: Url,
}

impl BeeperDesktop {
    /// Create a new BeeperDesktop client with default configuration from environment
    pub async fn new() -> Result<Self> {
        let config = Config::from_env()?;
        Self::with_config(config).await
    }

    /// Create a new BeeperDesktop client with the given configuration
    pub async fn with_config(config: Config) -> Result<Self> {
        config.validate()?;

        // Ensure base URL ends with /
        let base_url_str = if config.base_url.ends_with('/') {
            config.base_url.clone()
        } else {
            format!("{}/", config.base_url)
        };

        let base_url = Url::parse(&base_url_str)?;

        let http_client = if let Some(client) = config.http_client.clone() {
            client
        } else {
            HttpClient::builder()
                .timeout(config.timeout)
                .build()?
        };

        Ok(Self {
            config: Arc::new(config),
            http_client,
            base_url,
        })
    }

    /// Get the accounts resource client
    pub fn accounts(&self) -> Accounts {
        Accounts::new(self.clone())
    }

    /// Get the app resource client
    pub fn app(&self) -> App {
        App::new(self.clone())
    }

    /// Get the chats resource client
    pub fn chats(&self) -> Chats {
        Chats::new(self.clone())
    }

    /// Get the contacts resource client
    pub fn contacts(&self) -> Contacts {
        Contacts::new(self.clone())
    }

    /// Get the messages resource client
    pub fn messages(&self) -> Messages {
        Messages::new(self.clone())
    }

    /// Get the token resource client
    pub fn token(&self) -> Token {
        Token::new(self.clone())
    }

    /// Make an HTTP request with retry logic
    pub async fn do_request<T, R>(&self, method: Method, path: &str, body: Option<&T>) -> Result<R>
    where
        T: Serialize + ?Sized,
        R: for<'de> Deserialize<'de>,
    {
        let mut retries_left = self.config.max_retries;
        
        loop {
            match self.do_request_once(method.clone(), path, body).await {
                Ok(result) => return Ok(result),
                Err(error) if retries_left > 0 && error.is_retryable() => {
                    warn!("Request failed with retryable error: {}. Retrying...", error);
                    
                    // Exponential backoff
                    let delay = Duration::from_millis(1000 * (self.config.max_retries - retries_left + 1) as u64);
                    tokio::time::sleep(delay).await;
                    
                    retries_left -= 1;
                }
                Err(error) => return Err(error),
            }
        }
    }

    /// Make an HTTP request with query parameters
    pub async fn do_request_with_query<R>(
        &self,
        method: Method,
        path: &str,
        query: &[(&str, &str)],
    ) -> Result<R>
    where
        R: for<'de> Deserialize<'de>,
    {
        let mut retries_left = self.config.max_retries;
        
        loop {
            match self.do_request_with_query_once(method.clone(), path, query).await {
                Ok(result) => return Ok(result),
                Err(error) if retries_left > 0 && error.is_retryable() => {
                    warn!("Request failed with retryable error: {}. Retrying...", error);
                    
                    // Exponential backoff
                    let delay = Duration::from_millis(1000 * (self.config.max_retries - retries_left + 1) as u64);
                    tokio::time::sleep(delay).await;
                    
                    retries_left -= 1;
                }
                Err(error) => return Err(error),
            }
        }
    }

    /// Make a single HTTP request without retry
    async fn do_request_once<T, R>(&self, method: Method, path: &str, body: Option<&T>) -> Result<R>
    where
        T: Serialize + ?Sized,
        R: for<'de> Deserialize<'de>,
    {
        let url = self.base_url.join(path.trim_start_matches('/'))?;
        
        debug!("Making {} request to {}", method, url);

        let mut request = self.http_client
            .request(method, url)
            .header("Authorization", format!("Bearer {}", self.config.access_token))
            .header("User-Agent", &self.config.user_agent)
            .header("Accept", "application/json");

        if let Some(body) = body {
            let json_body = serde_json::to_string(body)?;
            debug!("Request body: {}", json_body);
            request = request
                .header("Content-Type", "application/json")
                .body(json_body);
        }

        let response = request.send().await?;
        self.handle_response(response).await
    }

    /// Make a single HTTP request with query parameters without retry
    async fn do_request_with_query_once<R>(
        &self,
        method: Method,
        path: &str,
        query: &[(&str, &str)],
    ) -> Result<R>
    where
        R: for<'de> Deserialize<'de>,
    {
        let mut url = self.base_url.join(path.trim_start_matches('/'))?;
        
        // Add query parameters
        if !query.is_empty() {
            let mut url_query = url.query_pairs_mut();
            for (key, value) in query {
                url_query.append_pair(key, value);
            }
            url_query.finish();
        }
        
        debug!("Making {} request to {}", method, url);

        let request = self.http_client
            .request(method, url)
            .header("Authorization", format!("Bearer {}", self.config.access_token))
            .header("User-Agent", &self.config.user_agent)
            .header("Accept", "application/json");

        let response = request.send().await?;
        self.handle_response(response).await
    }



    /// Handle HTTP response and convert to typed result
    async fn handle_response<R>(&self, response: Response) -> Result<R>
    where
        R: for<'de> Deserialize<'de>,
    {
        let status = response.status();
        let response_text = response.text().await?;

        debug!("Response status: {}, body: {}", status, response_text);

        if status.is_success() {
            serde_json::from_str(&response_text)
                .map_err(|e| {
                    warn!("Failed to parse successful response: {}", e);
                    Error::Json(e)
                })
        } else {
            self.handle_error_response(status, &response_text)
        }
    }

    /// Convert HTTP error response to typed error
    fn handle_error_response<R>(&self, status: StatusCode, body: &str) -> Result<R> {
        let error_response: ErrorResponse = serde_json::from_str(body).unwrap_or_else(|_| {
            ErrorResponse {
                error: Some(body.to_string()),
                code: None,
                details: None,
            }
        });

        let message = error_response.error.unwrap_or_else(|| body.to_string());

        match status {
            StatusCode::BAD_REQUEST => Err(Error::BadRequest {
                message,
                code: error_response.code,
                details: error_response.details,
            }),
            StatusCode::UNAUTHORIZED => Err(Error::Authentication {
                message,
                code: error_response.code,
                details: error_response.details,
            }),
            StatusCode::FORBIDDEN => Err(Error::PermissionDenied {
                message,
                code: error_response.code,
                details: error_response.details,
            }),
            StatusCode::NOT_FOUND => Err(Error::NotFound {
                message,
                code: error_response.code,
                details: error_response.details,
            }),
            StatusCode::CONFLICT => Err(Error::Conflict {
                message,
                code: error_response.code,
                details: error_response.details,
            }),
            StatusCode::UNPROCESSABLE_ENTITY => Err(Error::UnprocessableEntity {
                message,
                code: error_response.code,
                details: error_response.details,
            }),
            StatusCode::TOO_MANY_REQUESTS => Err(Error::RateLimit {
                message,
                code: error_response.code,
                details: error_response.details,
            }),
            status if status.is_server_error() => Err(Error::InternalServer {
                message,
                code: error_response.code,
                details: error_response.details,
            }),
            _ => Err(Error::Api {
                status: status.as_u16(),
                message,
                code: error_response.code,
                details: error_response.details,
            }),
        }
    }
}