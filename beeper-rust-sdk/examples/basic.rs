use beeper_desktop_api::{BeeperDesktop, Config};
use beeper_desktop_api::resources::{MessageSendParams, ChatSearchParams};
use std::time::Duration;

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    // Initialize tracing for logging
    tracing_subscriber::fmt::init();

    // Create client with default configuration from environment variables
    let client = BeeperDesktop::new().await?;

    // Or create client with explicit configuration
    let _config = Config::builder()
        .access_token("your-access-token")
        .base_url("http://localhost:23373")
        .timeout(Duration::from_secs(30))
        .max_retries(3)
        .build()?;

    // List connected accounts
    println!("Fetching accounts...");
    let accounts = client.accounts().list().await?;
    for account in &accounts {
        println!("Account: {} ({})", account.account_id, account.network);
    }

    // Get token info
    println!("\nFetching token info...");
    let token_info = client.token().info().await?;
    println!("Token subject: {}", token_info.sub);
    println!("Token scope: {}", token_info.scope);

    // Search chats
    println!("\nSearching chats...");
    let mut search_params = ChatSearchParams::new();
    search_params.limit = Some(10);
    
    let chats = client.chats().search(&search_params).await?;
    println!("Found {} chats", chats.items.len());
    
    for chat in &chats.items {
        println!("Chat: {} - {} ({})", chat.id, chat.title, chat.chat_type);
    }

    // Send a message (if there are any chats)
    if let Some(first_chat) = chats.items.first() {
        println!("\nSending message to chat: {}", first_chat.title);
        
        let send_params = MessageSendParams {
            chat_id: first_chat.id.clone(),
            text: "Hello from Rust SDK!".to_string(),
            reply_to_id: None,
            attachment: None,
        };
        
        match client.messages().send(&send_params).await {
            Ok(response) => {
                if response.success {
                    println!("Message sent successfully: {}", response.message_id);
                } else {
                    println!("Failed to send message: {:?}", response.error);
                }
            }
            Err(e) => {
                println!("Error sending message: {}", e);
            }
        }
    } else {
        println!("No chats found to send message to");
    }

    println!("\nExample completed successfully!");
    Ok(())
}