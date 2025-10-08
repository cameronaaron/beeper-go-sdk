# Beeper Desktop API Rust SDK

[![Crates.io](https://img.shields.io/crates/v/beeper-desktop-api.svg)](https://crates.io/crates/beeper-desktop-api)
[![Documentation](https://docs.rs/beeper-desktop-api/badge.svg)](https://docs.rs/beeper-desktop-api)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

This library provides convenient access to the Beeper Desktop API from Rust applications.

The documentation for Beeper Desktop API can be found on [developers.beeper.com/desktop-api](https://developers.beeper.com/desktop-api/).

## Installation

Add this to your `Cargo.toml`:

```toml
[dependencies]
beeper-desktop-api = "0.1.0"
```

## Usage

```rust
use beeper_desktop_api::{BeeperDesktop, Config};
use beeper_desktop_api::resources::{MessageSendParams, ChatSearchParams};

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    // Create client with access token from environment variable
    let client = BeeperDesktop::new().await?;

    // Or create client with explicit configuration
    let config = Config::builder()
        .access_token("your-access-token")
        .base_url("http://localhost:23373")
        .timeout(std::time::Duration::from_secs(30))
        .max_retries(3)
        .build()?;
    
    let client = BeeperDesktop::with_config(config).await?;

    // List connected accounts
    let accounts = client.accounts().list().await?;
    for account in accounts {
        println!("Account: {} ({})", account.account_id, account.network);
    }

    // Search chats
    let mut search_params = ChatSearchParams::new();
    search_params.limit = Some(10);
    
    let chats = client.chats().search(&search_params).await?;
    for chat in chats.items {
        println!("Chat: {}", chat.id);
    }

    // Send a message
    let send_params = MessageSendParams {
        chat_id: "your-chat-id".to_string(),
        text: "Hello from Rust!".to_string(),
        reply_to_id: None,
        attachment: None,
    };
    
    let response = client.messages().send(&send_params).await?;
    println!("Message sent: {}", response.message_id);

    Ok(())
}
```

## Configuration Options

The client can be configured with various options using the builder pattern:

```rust
use beeper_desktop_api::{BeeperDesktop, Config};
use std::time::Duration;

let config = Config::builder()
    .access_token("your-access-token")
    .base_url("http://localhost:23373")
    .timeout(Duration::from_secs(30))
    .max_retries(3)
    .user_agent("my-app/1.0")
    .build()?;

let client = BeeperDesktop::with_config(config).await?;
```

## Error Handling

The SDK provides typed errors for different HTTP status codes:

```rust
use beeper_desktop_api::Error;

match client.accounts().list().await {
    Ok(accounts) => {
        // Handle success
    }
    Err(Error::Authentication(e)) => {
        println!("Authentication failed: {}", e.message);
    }
    Err(Error::NotFound(e)) => {
        println!("Resource not found: {}", e.message);
    }
    Err(Error::RateLimit(e)) => {
        println!("Rate limited: {}", e.message);
    }
    Err(e) => {
        println!("Other error: {}", e);
    }
}
```

## Pagination

For paginated endpoints, you can iterate through all results:

```rust
use beeper_desktop_api::resources::MessageSearchParams;

let mut params = MessageSearchParams::new();
params.query = Some("hello".to_string());
params.limit = Some(10);

loop {
    let messages = client.messages().search(&params).await?;
    
    for message in &messages.items {
        if let Some(text) = &message.text {
            println!("Message: {}", text);
        }
    }
    
    // Check if there are more pages
    if let Some(pagination) = &messages.pagination {
        if !pagination.has_more {
            break;
        }
        params.cursor = pagination.cursor.clone();
    } else {
        break;
    }
}
```

## Async Support

All API methods are async and use Tokio as the async runtime:

```rust
use tokio::time::{timeout, Duration};

// With timeout
let accounts = timeout(Duration::from_secs(10), client.accounts().list()).await??;

// Concurrent requests
let (accounts, chats) = tokio::join!(
    client.accounts().list(),
    client.chats().search(&ChatSearchParams::new())
);
```

## Environment Variables

The SDK respects the following environment variables:

- `BEEPER_ACCESS_TOKEN`: Access token for authentication
- `BEEPER_DESKTOP_BASE_URL`: Base URL for the API (defaults to `http://localhost:23373`)

## Requirements

- Rust 1.70.0 or later

## Features

- `default`: Includes `rustls-tls` for TLS support
- `rustls-tls`: Use rustls for TLS (recommended)
- `native-tls`: Use native system TLS

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.