use thiserror::Error;
use std::collections::HashMap;

/// Main error type for the Beeper Desktop API SDK
#[derive(Error, Debug)]
pub enum Error {
    /// HTTP client error (connection, timeout, etc.)
    #[error("HTTP client error: {0}")]
    Http(#[from] reqwest::Error),

    /// JSON serialization/deserialization error
    #[error("JSON error: {0}")]
    Json(#[from] serde_json::Error),

    /// URL parsing error
    #[error("URL error: {0}")]
    Url(#[from] url::ParseError),

    /// Configuration error
    #[error("Configuration error: {message}")]
    Config { message: String },

    /// API error response (4xx/5xx status codes)
    #[error("API error {status}: {message}")]
    Api {
        status: u16,
        message: String,
        code: Option<String>,
        details: Option<HashMap<String, String>>,
    },

    /// Authentication error (401)
    #[error("Authentication error: {message}")]
    Authentication {
        message: String,
        code: Option<String>,
        details: Option<HashMap<String, String>>,
    },

    /// Bad request error (400)
    #[error("Bad request: {message}")]
    BadRequest {
        message: String,
        code: Option<String>,
        details: Option<HashMap<String, String>>,
    },

    /// Permission denied error (403)
    #[error("Permission denied: {message}")]
    PermissionDenied {
        message: String,
        code: Option<String>,
        details: Option<HashMap<String, String>>,
    },

    /// Not found error (404)
    #[error("Not found: {message}")]
    NotFound {
        message: String,
        code: Option<String>,
        details: Option<HashMap<String, String>>,
    },

    /// Conflict error (409)
    #[error("Conflict: {message}")]
    Conflict {
        message: String,
        code: Option<String>,
        details: Option<HashMap<String, String>>,
    },

    /// Unprocessable entity error (422)
    #[error("Unprocessable entity: {message}")]
    UnprocessableEntity {
        message: String,
        code: Option<String>,
        details: Option<HashMap<String, String>>,
    },

    /// Rate limit error (429)
    #[error("Rate limited: {message}")]
    RateLimit {
        message: String,
        code: Option<String>,
        details: Option<HashMap<String, String>>,
    },

    /// Internal server error (5xx)
    #[error("Internal server error: {message}")]
    InternalServer {
        message: String,
        code: Option<String>,
        details: Option<HashMap<String, String>>,
    },
}

impl Error {
    /// Check if the error is retryable
    pub fn is_retryable(&self) -> bool {
        match self {
            Error::Http(e) => {
                // Connection errors are retryable
                e.is_connect() || e.is_timeout() || e.is_request()
            }
            Error::Conflict { .. } => true,
            Error::RateLimit { .. } => true,
            Error::InternalServer { .. } => true,
            Error::Api { status, .. } => *status == 408 || *status >= 500,
            _ => false,
        }
    }

    /// Create an authentication error
    pub fn authentication(message: impl Into<String>) -> Self {
        Self::Authentication {
            message: message.into(),
            code: None,
            details: None,
        }
    }

    /// Create a configuration error
    pub fn config(message: impl Into<String>) -> Self {
        Self::Config {
            message: message.into(),
        }
    }
}

/// Result type alias for the SDK
pub type Result<T> = std::result::Result<T, Error>;