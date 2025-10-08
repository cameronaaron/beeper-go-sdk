//! # Beeper Desktop API Rust SDK
//!
//! This library provides convenient access to the Beeper Desktop API from Rust applications.
//!
//! ## Quick Start
//!
//! ```rust,no_run
//! use beeper_desktop_api::{BeeperDesktop, Config};
//!
//! #[tokio::main]
//! async fn main() -> Result<(), Box<dyn std::error::Error>> {
//!     // Create client with access token from environment variable
//!     let client = BeeperDesktop::new().await?;
//!
//!     // List connected accounts
//!     let accounts = client.accounts().list().await?;
//!     for account in accounts {
//!         println!("Account: {} ({})", account.account_id, account.network);
//!     }
//!
//!     Ok(())
//! }
//! ```

pub mod client;
pub mod config;
pub mod error;
pub mod resources;
pub mod utils;
pub mod version;

pub use client::BeeperDesktop;
pub use config::Config;
pub use error::{Error, Result};
pub use version::VERSION;

// Re-export commonly used types
pub use resources::{
    Account, Chat, Message, User, 
    MessageSearchParams, MessageSendParams,
    ChatSearchParams, ChatCreateParams,
    ContactSearchParams,
    AppSearchParams, AppOpenParams, AppDownloadAssetParams,
};