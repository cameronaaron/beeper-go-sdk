use beeper_desktop_api::{BeeperDesktop, Config, Error};
use beeper_desktop_api::resources::{
    MessageSearchParams, MessageSendParams, ChatSearchParams, ChatCreateParams,
    ContactSearchParams, AppSearchParams,
};
use std::io::{self, Write};
use std::time::Duration;

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    // Initialize logging
    tracing_subscriber::fmt().init();

    println!("=================================");
    println!("  Beeper Desktop API Test App");
    println!("=================================\n");

    // Create client
    let client = match create_client().await {
        Ok(client) => {
            println!("✓ Client created successfully");
            client
        }
        Err(e) => {
            eprintln!("✗ Failed to create client: {}", e);
            return Ok(());
        }
    };

    loop {
        print_menu();
        
        let choice = get_user_input("\nEnter your choice: ")?;
        
        match choice.trim() {
            "1" => test_token_info(&client).await,
            "2" => test_accounts_list(&client).await,
            "3" => test_chats_search(&client).await,
            "4" => test_messages_search(&client).await,
            "5" => test_contacts_search(&client).await,
            "6" => test_send_message(&client).await,
            "7" => test_create_chat(&client).await,
            "8" => test_app_search(&client).await,
            "9" => test_pagination(&client).await,
            "10" => test_error_handling(&client).await,
            "q" | "Q" | "quit" => {
                println!("Goodbye!");
                break;
            }
            _ => println!("Invalid choice. Please try again."),
        }
        
        println!("\nPress Enter to continue...");
        let _ = get_user_input("");
    }

    Ok(())
}

async fn create_client() -> Result<BeeperDesktop, Box<dyn std::error::Error>> {
    // Try environment variables first
    match BeeperDesktop::new().await {
        Ok(client) => Ok(client),
        Err(beeper_desktop_api::Error::Config { .. }) => {
            // If env vars not set, prompt user
            println!("Environment variables not set. Please provide configuration:");
            
            let access_token = get_user_input("Access Token: ")?;
            let base_url = get_user_input_with_default(
                "Base URL", 
                "http://localhost:23373"
            )?;
            
            let config = Config::builder()
                .access_token(access_token.trim())
                .base_url(base_url.trim())
                .timeout(Duration::from_secs(30))
                .max_retries(3)
                .build()?;
                
            Ok(BeeperDesktop::with_config(config).await?)
        }
        Err(e) => Err(Box::new(e)),
    }
}

fn print_menu() {
    println!("\n=== Test Menu ===");
    println!("1.  Test Token Info");
    println!("2.  List Accounts");
    println!("3.  Search Chats");
    println!("4.  Search Messages");
    println!("5.  Search Contacts");
    println!("6.  Send Message");
    println!("7.  Create Chat");
    println!("8.  App Search");
    println!("9.  Test Pagination");
    println!("10. Test Error Handling");
    println!("Q.  Quit");
}

fn get_user_input(prompt: &str) -> Result<String, Box<dyn std::error::Error>> {
    print!("{}", prompt);
    io::stdout().flush()?;
    
    let mut input = String::new();
    io::stdin().read_line(&mut input)?;
    Ok(input)
}

fn get_user_input_with_default(prompt: &str, default: &str) -> Result<String, Box<dyn std::error::Error>> {
    let full_prompt = format!("{} [{}]: ", prompt, default);
    let input = get_user_input(&full_prompt)?;
    
    let trimmed = input.trim();
    if trimmed.is_empty() {
        Ok(default.to_string())
    } else {
        Ok(trimmed.to_string())
    }
}

async fn test_token_info(client: &BeeperDesktop) {
    println!("\n--- Testing Token Info ---");
    
    match client.token().info().await {
        Ok(token_info) => {
            println!("✓ Token Info Retrieved:");
            println!("  Subject: {}", token_info.sub);
            println!("  Scope: {}", token_info.scope);
            println!("  Token Use: {}", token_info.token_use);
            println!("  Issued At: {}", token_info.iat);
            
            if let Some(exp) = token_info.exp {
                println!("  Expires: {}", exp);
            }
            if let Some(aud) = &token_info.aud {
                println!("  Audience: {}", aud);
            }
        }
        Err(e) => println!("✗ Error: {}", e),
    }
}

async fn test_accounts_list(client: &BeeperDesktop) {
    println!("\n--- Testing Accounts List ---");
    
    match client.accounts().list().await {
        Ok(accounts) => {
            println!("✓ Found {} account(s):", accounts.len());
            
            for (i, account) in accounts.iter().enumerate() {
                println!("  {}. {} ({})", i + 1, account.account_id, account.network);
                println!("     User: {}", account.user.full_name.as_deref().unwrap_or("No name"));
                if let Some(email) = &account.user.email {
                    println!("     Email: {}", email);
                }
                if let Some(phone) = &account.user.phone_number {
                    println!("     Phone: {}", phone);
                }
            }
        }
        Err(e) => println!("✗ Error: {}", e),
    }
}

async fn test_chats_search(client: &BeeperDesktop) {
    println!("\n--- Testing Chats Search ---");
    
    let limit_str = get_user_input_with_default("Limit", "10").unwrap();
    let limit = limit_str.parse::<i32>().unwrap_or(10);
    
    let mut params = ChatSearchParams::new();
    params.limit = Some(limit);
    
    match client.chats().search(&params).await {
        Ok(chats) => {
            println!("✓ Found {} chat(s):", chats.items.len());
            
            for (i, chat) in chats.items.iter().enumerate() {
                println!("  {}. {} - {} ({})", 
                    i + 1, 
                    chat.id, 
                    chat.title, 
                    chat.chat_type
                );
                println!("     Network: {}", chat.network);
                println!("     Unread: {}", chat.unread_count);
                println!("     Participants: {}", chat.participants.total);
            }
            
            if let Some(pagination) = &chats.pagination {
                if pagination.has_more {
                    println!("  (More results available)");
                }
            }
        }
        Err(e) => println!("✗ Error: {}", e),
    }
}

async fn test_messages_search(client: &BeeperDesktop) {
    println!("\n--- Testing Messages Search ---");
    
    let query = get_user_input("Search query (optional): ").unwrap();
    let limit_str = get_user_input_with_default("Limit", "10").unwrap();
    let limit = limit_str.parse::<i32>().unwrap_or(10);
    
    let mut params = MessageSearchParams::new();
    if !query.trim().is_empty() {
        params.query = Some(query.trim().to_string());
    }
    params.limit = Some(limit);
    
    match client.messages().search(&params).await {
        Ok(messages) => {
            println!("✓ Found {} message(s):", messages.items.len());
            
            for (i, message) in messages.items.iter().enumerate() {
                println!("  {}. {} ({})", i + 1, message.message_id, message.chat_id);
                if let Some(text) = &message.text {
                    let preview = if text.len() > 50 {
                        format!("{}...", &text[..47])
                    } else {
                        text.clone()
                    };
                    println!("     Text: {}", preview);
                }
                if let Some(sender_name) = &message.sender_name {
                    println!("     Sender: {}", sender_name);
                }
            }
            
            if let Some(pagination) = &messages.pagination {
                if pagination.has_more {
                    println!("  (More results available)");
                }
            }
        }
        Err(e) => println!("✗ Error: {}", e),
    }
}

async fn test_contacts_search(client: &BeeperDesktop) {
    println!("\n--- Testing Contacts Search ---");
    
    // First get accounts
    let accounts = match client.accounts().list().await {
        Ok(accounts) => accounts,
        Err(e) => {
            println!("✗ Error getting accounts: {}", e);
            return;
        }
    };
    
    if accounts.is_empty() {
        println!("✗ No accounts found");
        return;
    }
    
    println!("Available accounts:");
    for (i, account) in accounts.iter().enumerate() {
        println!("  {}. {} ({})", i + 1, account.account_id, account.network);
    }
    
    let account_index_str = get_user_input("Select account (number): ").unwrap();
    let account_index = account_index_str.trim().parse::<usize>().unwrap_or(1) - 1;
    
    if account_index >= accounts.len() {
        println!("✗ Invalid account selection");
        return;
    }
    
    let selected_account = &accounts[account_index];
    let query = get_user_input("Search query: ").unwrap();
    
    let params = ContactSearchParams {
        account_id: selected_account.account_id.clone(),
        query: query.trim().to_string(),
    };
    
    match client.contacts().search(&params).await {
        Ok(contacts) => {
            println!("✓ Found {} contact(s):", contacts.items.len());
            
            for (i, contact) in contacts.items.iter().enumerate() {
                println!("  {}. {} ({})", i + 1, contact.id, 
                    contact.full_name.as_deref().unwrap_or("No name"));
                if let Some(username) = &contact.username {
                    println!("     Username: {}", username);
                }
                if let Some(email) = &contact.email {
                    println!("     Email: {}", email);
                }
            }
        }
        Err(e) => println!("✗ Error: {}", e),
    }
}

async fn test_send_message(client: &BeeperDesktop) {
    println!("\n--- Testing Send Message ---");
    
    // First get chats
    let mut search_params = ChatSearchParams::new();
    search_params.limit = Some(5);
    
    let chats = match client.chats().search(&search_params).await {
        Ok(chats) => chats,
        Err(e) => {
            println!("✗ Error getting chats: {}", e);
            return;
        }
    };
    
    if chats.items.is_empty() {
        println!("✗ No chats found");
        return;
    }
    
    println!("Available chats:");
    for (i, chat) in chats.items.iter().enumerate() {
        println!("  {}. {} - {}", i + 1, chat.title, chat.id);
    }
    
    let chat_index_str = get_user_input("Select chat (number): ").unwrap();
    let chat_index = chat_index_str.trim().parse::<usize>().unwrap_or(1) - 1;
    
    if chat_index >= chats.items.len() {
        println!("✗ Invalid chat selection");
        return;
    }
    
    let selected_chat = &chats.items[chat_index];
    let message_text = get_user_input("Message text: ").unwrap();
    
    let params = MessageSendParams {
        chat_id: selected_chat.id.clone(),
        text: message_text.trim().to_string(),
        reply_to_id: None,
        attachment: None,
    };
    
    match client.messages().send(&params).await {
        Ok(response) => {
            if response.success {
                println!("✓ Message sent successfully!");
                println!("  Message ID: {}", response.message_id);
                println!("  Deeplink: {}", response.deeplink);
            } else {
                println!("✗ Failed to send message: {:?}", response.error);
            }
        }
        Err(e) => println!("✗ Error: {}", e),
    }
}

async fn test_create_chat(client: &BeeperDesktop) {
    println!("\n--- Testing Create Chat ---");
    
    // Get accounts first
    let accounts = match client.accounts().list().await {
        Ok(accounts) => accounts,
        Err(e) => {
            println!("✗ Error getting accounts: {}", e);
            return;
        }
    };
    
    if accounts.is_empty() {
        println!("✗ No accounts found");
        return;
    }
    
    println!("Available accounts:");
    for (i, account) in accounts.iter().enumerate() {
        println!("  {}. {} ({})", i + 1, account.account_id, account.network);
    }
    
    let account_index_str = get_user_input("Select account (number): ").unwrap();
    let account_index = account_index_str.trim().parse::<usize>().unwrap_or(1) - 1;
    
    if account_index >= accounts.len() {
        println!("✗ Invalid account selection");
        return;
    }
    
    let selected_account = &accounts[account_index];
    let participant_id = get_user_input("Participant ID: ").unwrap();
    let chat_type = get_user_input_with_default("Chat type", "single").unwrap();
    
    let params = ChatCreateParams {
        account_id: selected_account.account_id.clone(),
        participant_ids: vec![participant_id.trim().to_string()],
        chat_type: chat_type.trim().to_string(),
        title: None,
    };
    
    match client.chats().create(&params).await {
        Ok(response) => {
            if response.success {
                println!("✓ Chat created successfully!");
                println!("  Chat ID: {}", response.chat.id);
                println!("  Chat Title: {}", response.chat.title);
            } else {
                println!("✗ Failed to create chat: {:?}", response.error);
            }
        }
        Err(e) => println!("✗ Error: {}", e),
    }
}

async fn test_app_search(client: &BeeperDesktop) {
    println!("\n--- Testing App Search ---");
    
    let query = get_user_input("Search query: ").unwrap();
    let limit_str = get_user_input_with_default("Limit", "10").unwrap();
    let limit = limit_str.parse::<i32>().unwrap_or(10);
    
    let params = AppSearchParams {
        query: query.trim().to_string(),
        account_ids: None,
        chat_type: None,
        include_muted: None,
        limit: Some(limit),
        message_limit: Some(5),
        participant_limit: Some(5),
    };
    
    match client.app().search(&params).await {
        Ok(response) => {
            println!("✓ Search completed:");
            println!("  Chats: {}", response.chats.len());
            println!("  Messages: {}", response.messages.len());
            
            for (i, chat_result) in response.chats.iter().enumerate() {
                println!("  Chat {}: {} - {}", 
                    i + 1, 
                    chat_result.chat.title, 
                    chat_result.chat.id
                );
            }
            
            for (i, message_result) in response.messages.iter().enumerate() {
                println!("  Message {}: {} in {}", 
                    i + 1, 
                    message_result.message.message_id,
                    message_result.chat.title
                );
            }
        }
        Err(e) => println!("✗ Error: {}", e),
    }
}

async fn test_pagination(client: &BeeperDesktop) {
    println!("\n--- Testing Pagination ---");
    
    let mut params = MessageSearchParams::new();
    params.limit = Some(5);
    
    let mut total_messages = 0;
    let mut page_count = 0;
    let max_pages = 3;
    
    loop {
        match client.messages().search(&params).await {
            Ok(messages) => {
                page_count += 1;
                total_messages += messages.items.len();
                
                println!("Page {}: {} messages", page_count, messages.items.len());
                
                // Check if there are more pages and we haven't hit our limit
                if let Some(pagination) = &messages.pagination {
                    if pagination.has_more && page_count < max_pages {
                        params.cursor = pagination.cursor.clone();
                        continue;
                    }
                }
                
                break;
            }
            Err(e) => {
                println!("✗ Error: {}", e);
                break;
            }
        }
    }
    
    println!("✓ Pagination test completed: {} messages across {} pages", 
        total_messages, page_count);
}

async fn test_error_handling(client: &BeeperDesktop) {
    println!("\n--- Testing Error Handling ---");
    
    // Test with invalid chat ID
    println!("Testing with invalid chat ID...");
    let params = MessageSendParams {
        chat_id: "invalid-chat-id-12345".to_string(),
        text: "This should fail".to_string(),
        reply_to_id: None,
        attachment: None,
    };
    
    match client.messages().send(&params).await {
        Ok(response) => {
            if !response.success {
                println!("✓ API returned error as expected: {:?}", response.error);
            } else {
                println!("? Unexpected success");
            }
        }
        Err(Error::NotFound { message, .. }) => {
            println!("✓ Got expected NotFound error: {}", message);
        }
        Err(Error::BadRequest { message, .. }) => {
            println!("✓ Got expected BadRequest error: {}", message);
        }
        Err(e) => {
            println!("? Got different error type: {}", e);
        }
    }
}