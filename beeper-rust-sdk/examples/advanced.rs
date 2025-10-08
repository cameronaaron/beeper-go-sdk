use beeper_desktop_api::{BeeperDesktop, Config, Error};
use beeper_desktop_api::resources::{MessageSearchParams, ChatSearchParams};
use std::time::Duration;
use tokio::time::timeout;

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    // Initialize tracing for detailed logging
    tracing_subscriber::fmt()
        .with_env_filter("beeper_desktop_api=debug,advanced=info")
        .init();

    println!("Advanced Beeper Desktop API Example");
    println!("===================================\n");

    // Create client with custom configuration
    let config = Config::builder()
        .timeout(Duration::from_secs(30))
        .max_retries(3)
        .user_agent("advanced-example/1.0")
        .build()?;

    let client = BeeperDesktop::with_config(config).await?;

    // Demonstrate error handling
    println!("1. Error Handling Demonstration");
    println!("-------------------------------");
    
    match client.accounts().list().await {
        Ok(accounts) => {
            println!("✓ Successfully fetched {} accounts", accounts.len());
            for account in &accounts {
                println!("  - {} ({}): {}", 
                    account.account_id, 
                    account.network, 
                    account.user.full_name.as_deref().unwrap_or("No name")
                );
            }
        }
        Err(Error::Authentication { message, .. }) => {
            println!("✗ Authentication failed: {}", message);
            return Ok(());
        }
        Err(Error::NotFound { message, .. }) => {
            println!("✗ Resource not found: {}", message);
        }
        Err(Error::RateLimit { message, .. }) => {
            println!("✗ Rate limited: {}", message);
        }
        Err(e) => {
            println!("✗ Other error: {}", e);
            return Err(e.into());
        }
    }

    // Demonstrate timeout handling
    println!("\n2. Timeout Handling");
    println!("-------------------");
    
    match timeout(Duration::from_secs(5), client.token().info()).await {
        Ok(Ok(token_info)) => {
            println!("✓ Token info retrieved within timeout");
            println!("  - Subject: {}", token_info.sub);
            println!("  - Scope: {}", token_info.scope);
            if let Some(exp) = token_info.exp {
                println!("  - Expires: {}", exp);
            }
        }
        Ok(Err(e)) => {
            println!("✗ API error: {}", e);
        }
        Err(_) => {
            println!("✗ Request timed out");
        }
    }

    // Demonstrate concurrent requests
    println!("\n3. Concurrent Requests");
    println!("----------------------");
    
    let mut search_params = ChatSearchParams::new();
    search_params.limit = Some(5);
    
    let accounts_client = client.accounts();
    let chats_client = client.chats();
    let token_client = client.token();
    
    let (accounts_result, chats_result, token_result) = tokio::join!(
        accounts_client.list(),
        chats_client.search(&search_params),
        token_client.info()
    );

    match (accounts_result, chats_result, token_result) {
        (Ok(accounts), Ok(chats), Ok(token_info)) => {
            println!("✓ All concurrent requests successful");
            println!("  - Accounts: {}", accounts.len());
            println!("  - Chats: {}", chats.items.len());
            println!("  - Token subject: {}", token_info.sub);
        }
        _ => {
            println!("✗ Some concurrent requests failed");
        }
    }

    // Demonstrate pagination
    println!("\n4. Pagination Example");
    println!("---------------------");
    
    let mut message_params = MessageSearchParams::new();
    message_params.limit = Some(5);
    
    let mut total_messages = 0;
    let mut page_count = 0;
    let max_pages = 3; // Limit to avoid too much output
    
    loop {
        match client.messages().search(&message_params).await {
            Ok(messages) => {
                page_count += 1;
                total_messages += messages.items.len();
                
                println!("Page {}: {} messages", page_count, messages.items.len());
                
                for (i, message) in messages.items.iter().enumerate() {
                    if let Some(text) = &message.text {
                        let preview = if text.len() > 50 {
                            format!("{}...", &text[..47])
                        } else {
                            text.clone()
                        };
                        println!("  {}. {}: {}", i + 1, message.sender_id, preview);
                    }
                }
                
                // Check if there are more pages and we haven't hit our limit
                if let Some(pagination) = &messages.pagination {
                    if pagination.has_more && page_count < max_pages {
                        message_params.cursor = pagination.cursor.clone();
                        continue;
                    }
                }
                
                break;
            }
            Err(e) => {
                println!("✗ Error searching messages: {}", e);
                break;
            }
        }
    }
    
    println!("Total messages retrieved: {} across {} pages", total_messages, page_count);

    // Demonstrate search functionality
    println!("\n5. Search Functionality");
    println!("-----------------------");
    
    let mut search_params = MessageSearchParams::new();
    search_params.query = Some("hello".to_string());
    search_params.limit = Some(10);
    
    match client.messages().search(&search_params).await {
        Ok(messages) => {
            println!("✓ Found {} messages containing 'hello'", messages.items.len());
            
            for message in messages.items.iter().take(3) {
                if let Some(text) = &message.text {
                    println!("  - {}: {}", message.sender_id, text);
                }
            }
            
            if messages.items.len() > 3 {
                println!("  ... and {} more", messages.items.len() - 3);
            }
        }
        Err(e) => {
            println!("✗ Error searching messages: {}", e);
        }
    }

    println!("\n6. Contact Search");
    println!("-----------------");
    
    // First get an account to search in
    if let Ok(accounts) = client.accounts().list().await {
        if let Some(account) = accounts.first() {
            let contact_params = beeper_desktop_api::resources::ContactSearchParams {
                account_id: account.account_id.clone(),
                query: "test".to_string(),
            };
            
            match client.contacts().search(&contact_params).await {
                Ok(contacts) => {
                    println!("✓ Found {} contacts matching 'test'", contacts.items.len());
                    
                    for contact in contacts.items.iter().take(5) {
                        println!("  - {}: {}", 
                            contact.id, 
                            contact.full_name.as_deref().unwrap_or("No name")
                        );
                    }
                }
                Err(e) => {
                    println!("✗ Error searching contacts: {}", e);
                }
            }
        }
    }

    println!("\nAdvanced example completed!");
    Ok(())
}