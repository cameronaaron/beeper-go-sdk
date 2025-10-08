use crate::error::{Error, Result};
use crate::version::VERSION;
use std::time::Duration;
use reqwest::Client as HttpClient;

/// Configuration for the Beeper Desktop API client
#[derive(Debug, Clone)]
pub struct Config {
    /// Access token for authentication
    pub access_token: String,
    /// Base URL for the API
    pub base_url: String,
    /// Request timeout
    pub timeout: Duration,
    /// Maximum number of retries
    pub max_retries: u32,
    /// User agent string
    pub user_agent: String,
    /// HTTP client (optional)
    pub http_client: Option<HttpClient>,
}

impl Config {
    /// Create a new configuration builder
    pub fn builder() -> ConfigBuilder {
        ConfigBuilder::new()
    }

    /// Create a default configuration using environment variables
    pub fn from_env() -> Result<Self> {
        let access_token = std::env::var("BEEPER_ACCESS_TOKEN")
            .map_err(|_| Error::config("BEEPER_ACCESS_TOKEN environment variable is required"))?;
        
        let base_url = std::env::var("BEEPER_DESKTOP_BASE_URL")
            .unwrap_or_else(|_| "http://localhost:23373".to_string());

        Ok(Self {
            access_token,
            base_url,
            timeout: Duration::from_secs(30),
            max_retries: 2,
            user_agent: format!("beeper-desktop-api-rust/{}", VERSION),
            http_client: None,
        })
    }

    /// Validate the configuration
    pub fn validate(&self) -> Result<()> {
        if self.access_token.is_empty() {
            return Err(Error::config("access token is required"));
        }

        if self.base_url.is_empty() {
            return Err(Error::config("base URL is required"));
        }

        // Parse URL to validate format
        url::Url::parse(&self.base_url)
            .map_err(|e| Error::config(format!("invalid base URL: {}", e)))?;

        Ok(())
    }
}

/// Builder for creating Config instances
#[derive(Debug, Default)]
pub struct ConfigBuilder {
    access_token: Option<String>,
    base_url: Option<String>,
    timeout: Option<Duration>,
    max_retries: Option<u32>,
    user_agent: Option<String>,
    http_client: Option<HttpClient>,
}

impl ConfigBuilder {
    /// Create a new configuration builder
    pub fn new() -> Self {
        Self::default()
    }

    /// Set the access token
    pub fn access_token(mut self, token: impl Into<String>) -> Self {
        self.access_token = Some(token.into());
        self
    }

    /// Set the base URL
    pub fn base_url(mut self, url: impl Into<String>) -> Self {
        self.base_url = Some(url.into());
        self
    }

    /// Set the request timeout
    pub fn timeout(mut self, timeout: Duration) -> Self {
        self.timeout = Some(timeout);
        self
    }

    /// Set the maximum number of retries
    pub fn max_retries(mut self, retries: u32) -> Self {
        self.max_retries = Some(retries);
        self
    }

    /// Set the user agent string
    pub fn user_agent(mut self, agent: impl Into<String>) -> Self {
        self.user_agent = Some(agent.into());
        self
    }

    /// Set a custom HTTP client
    pub fn http_client(mut self, client: HttpClient) -> Self {
        self.http_client = Some(client);
        self
    }

    /// Build the configuration
    pub fn build(self) -> Result<Config> {
        let config = Config {
            access_token: self.access_token.unwrap_or_else(|| {
                std::env::var("BEEPER_ACCESS_TOKEN").unwrap_or_default()
            }),
            base_url: self.base_url.unwrap_or_else(|| {
                std::env::var("BEEPER_DESKTOP_BASE_URL")
                    .unwrap_or_else(|_| "http://localhost:23373".to_string())
            }),
            timeout: self.timeout.unwrap_or(Duration::from_secs(30)),
            max_retries: self.max_retries.unwrap_or(2),
            user_agent: self.user_agent.unwrap_or_else(|| {
                format!("beeper-desktop-api-rust/{}", VERSION)
            }),
            http_client: self.http_client,
        };

        config.validate()?;
        Ok(config)
    }
}