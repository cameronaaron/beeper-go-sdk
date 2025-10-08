use beeper_desktop_api::{BeeperDesktop, Config};
use beeper_desktop_api::resources::{ChatSearchParams, MessageSearchParams};
use chrono::Utc;
use std::fs::{create_dir_all, File};
use std::io::{self, Write};
use std::path::Path;
use std::time::Duration;

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    // Initialize logging
    tracing_subscriber::fmt()
        .with_env_filter("beeper_desktop_api=info,archive_chats=info")
        .init();

    println!("============================");
    println!("  Beeper Chat Archive Tool");
    println!("============================\n");

    // Create client
    let client = match create_client().await {
        Ok(client) => {
            println!("✓ Connected to Beeper Desktop API");
            client
        }
        Err(e) => {
            eprintln!("✗ Failed to connect: {}", e);
            return Ok(());
        }
    };

    // Verify connection
    match client.token().info().await {
        Ok(token_info) => {
            println!("✓ Authentication successful (Subject: {})", token_info.sub);
        }
        Err(e) => {
            eprintln!("✗ Authentication failed: {}", e);
            return Ok(());
        }
    }

    // Create output directory
    let output_dir = "chat-archives";
    create_dir_all(output_dir)?;
    println!("✓ Created output directory: {}", output_dir);

    // Get list of chats
    println!("\nFetching chats...");
    let chats = match get_all_chats(&client).await {
        Ok(chats) => {
            println!("✓ Found {} chats", chats.len());
            chats
        }
        Err(e) => {
            eprintln!("✗ Failed to fetch chats: {}", e);
            return Ok(());
        }
    };

    // Interactive chat selection or archive all
    let selected_chats = if get_user_confirmation("\nArchive all chats? (y/n): ")? {
        chats
    } else {
        select_chats_interactively(chats)?
    };

    if selected_chats.is_empty() {
        println!("No chats selected for archiving.");
        return Ok(());
    }

    println!("\nArchiving {} chat(s)...", selected_chats.len());

    // Archive each selected chat
    for (i, chat) in selected_chats.iter().enumerate() {
        println!("\n[{}/{}] Archiving: {}", i + 1, selected_chats.len(), chat.title);
        
        match archive_chat(&client, chat, output_dir).await {
            Ok(message_count) => {
                println!("  ✓ Archived {} messages", message_count);
            }
            Err(e) => {
                eprintln!("  ✗ Failed to archive chat: {}", e);
            }
        }
    }

    println!("\n✓ Archive process completed!");
    println!("Archives saved to: {}", output_dir);

    Ok(())
}

async fn create_client() -> Result<BeeperDesktop, Box<dyn std::error::Error>> {
    // Try environment variables first
    match BeeperDesktop::new().await {
        Ok(client) => Ok(client),
        Err(_) => {
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
                .timeout(Duration::from_secs(60)) // Longer timeout for archiving
                .max_retries(3)
                .build()?;
                
            Ok(BeeperDesktop::with_config(config).await?)
        }
    }
}

async fn get_all_chats(client: &BeeperDesktop) -> Result<Vec<beeper_desktop_api::Chat>, Box<dyn std::error::Error>> {
    let mut all_chats = Vec::new();
    let mut params = ChatSearchParams::new();
    params.limit = Some(50);

    loop {
        let chats = client.chats().search(&params).await?;
        all_chats.extend(chats.items);

        if let Some(pagination) = &chats.pagination {
            if pagination.has_more {
                params.cursor = pagination.cursor.clone();
            } else {
                break;
            }
        } else {
            break;
        }
    }

    Ok(all_chats)
}

fn select_chats_interactively(
    chats: Vec<beeper_desktop_api::Chat>,
) -> Result<Vec<beeper_desktop_api::Chat>, Box<dyn std::error::Error>> {
    println!("\nAvailable chats:");
    for (i, chat) in chats.iter().enumerate() {
        println!("  {}. {} ({} - {} messages)", 
            i + 1, 
            chat.title, 
            chat.network,
            chat.unread_count
        );
    }

    println!("\nEnter chat numbers to archive (comma-separated, or 'all'):");
    let input = get_user_input("Selection: ")?;
    
    if input.trim().to_lowercase() == "all" {
        return Ok(chats);
    }

    let mut selected = Vec::new();
    for part in input.split(',') {
        if let Ok(index) = part.trim().parse::<usize>() {
            if index > 0 && index <= chats.len() {
                selected.push(chats[index - 1].clone());
            }
        }
    }

    Ok(selected)
}

async fn archive_chat(
    client: &BeeperDesktop,
    chat: &beeper_desktop_api::Chat,
    output_dir: &str,
) -> Result<usize, Box<dyn std::error::Error>> {
    // Sanitize filename
    let filename = sanitize_filename(&format!("{}_{}.md", chat.network, chat.title));
    let filepath = Path::new(output_dir).join(filename);

    // Create markdown file
    let mut file = File::create(&filepath)?;

    // Write header
    writeln!(file, "# Chat Archive: {}\n", chat.title)?;
    writeln!(file, "- **Network:** {}", chat.network)?;
    writeln!(file, "- **Chat ID:** {}", chat.id)?;
    writeln!(file, "- **Type:** {}", chat.chat_type)?;
    writeln!(file, "- **Participants:** {}", chat.participants.total)?;
    writeln!(file, "- **Archived on:** {}\n", Utc::now().format("%Y-%m-%d %H:%M:%S UTC"))?;

    if !chat.participants.items.is_empty() {
        writeln!(file, "## Participants\n")?;
        for participant in &chat.participants.items {
            writeln!(file, "- **{}** ({})", 
                participant.full_name.as_deref().unwrap_or("Unknown"),
                participant.id
            )?;
        }
        writeln!(file)?;
    }

    writeln!(file, "## Messages\n")?;

    // Fetch all messages for this chat
    let mut params = MessageSearchParams::new();
    params.chat_ids = vec![chat.id.clone()];
    params.limit = Some(100);

    let mut message_count = 0;
    
    loop {
        match client.messages().search(&params).await {
            Ok(messages) => {
                // Write messages in chronological order
                let mut sorted_messages = messages.items;
                sorted_messages.sort_by(|a, b| a.timestamp.cmp(&b.timestamp));

                for message in &sorted_messages {
                    write_message_to_file(&mut file, message)?;
                    message_count += 1;
                }

                // Check for more pages
                if let Some(pagination) = &messages.pagination {
                    if pagination.has_more {
                        params.cursor = pagination.cursor.clone();
                        continue;
                    }
                }
                break;
            }
            Err(e) => {
                eprintln!("  Warning: Failed to fetch messages: {}", e);
                break;
            }
        }
    }

    writeln!(file, "\n---\n*Archive generated by Beeper Chat Archive Tool*")?;

    println!("  ✓ Saved to: {}", filepath.display());
    Ok(message_count)
}

fn write_message_to_file(
    file: &mut File,
    message: &beeper_desktop_api::Message,
) -> Result<(), Box<dyn std::error::Error>> {
    let timestamp = message.timestamp.format("%Y-%m-%d %H:%M:%S");
    let sender = message.sender_name.as_deref().unwrap_or(&message.sender_id);
    
    writeln!(file, "### {} - {}", sender, timestamp)?;
    
    if let Some(text) = &message.text {
        writeln!(file, "{}\n", text)?;
    }
    
    // Handle attachments
    if let Some(attachments) = &message.attachments {
        if !attachments.is_empty() {
            writeln!(file, "**Attachments:**")?;
            for attachment in attachments {
                let file_name = attachment.file_name.as_deref().unwrap_or("Unknown");
                let file_type = &attachment.attachment_type;
                writeln!(file, "- {} ({})", file_name, file_type)?;
                
                if let Some(src_url) = &attachment.src_url {
                    writeln!(file, "  - URL: {}", src_url)?;
                }
            }
            writeln!(file)?;
        }
    }
    
    // Handle reactions
    if let Some(reactions) = &message.reactions {
        if !reactions.is_empty() {
            write!(file, "**Reactions:** ")?;
            for (i, reaction) in reactions.iter().enumerate() {
                if i > 0 {
                    write!(file, ", ")?;
                }
                write!(file, "{}", reaction.reaction_key)?;
            }
            writeln!(file, "\n")?;
        }
    }

    Ok(())
}

fn sanitize_filename(name: &str) -> String {
    name.chars()
        .map(|c| {
            if c.is_alphanumeric() || c == '_' || c == '-' || c == '.' {
                c
            } else {
                '_'
            }
        })
        .collect()
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

fn get_user_confirmation(prompt: &str) -> Result<bool, Box<dyn std::error::Error>> {
    let input = get_user_input(prompt)?;
    let trimmed = input.trim().to_lowercase();
    Ok(trimmed == "y" || trimmed == "yes")
}