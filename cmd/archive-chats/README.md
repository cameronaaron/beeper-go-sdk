# 📦 Beeper Chat Archive Tool

A beautiful command-line tool to archive your Beeper chats to markdown format.

## Features

✨ **Beautiful Markdown Export**
- Clean, readable markdown format
- Organized by date with headers
- Message metadata preserved
- Reactions and attachments included
- Participant information

📊 **Complete Data Preservation**
- All messages chronologically sorted
- Sender information
- Timestamps
- Attachments (with file info)
- Reactions
- Message IDs for reference

🎯 **User-Friendly**
- Interactive chat selection
- Progress indicators for large chats
- Sanitized filenames
- Organized output directory

## Installation

```bash
cd Golang/cmd/archive-chats
go build -o archive-chats
```

## Usage

1. Make sure Beeper Desktop is running
2. Set your access token (if not using default):
   ```bash
   export BEEPER_ACCESS_TOKEN="your-token-here"
   ```

3. Run the tool:
   ```bash
   ./archive-chats
   ```

4. Follow the interactive prompts:
   - View all your chats numbered
   - Type `a` to archive all chats
   - Or enter specific numbers: `1,3,5` to archive selected chats
   - Type `q` to quit

## Output

Archives are saved to the `chat-archives/` directory with filenames like:

```
chat-archives/
├── 2025-10-07_instagram_sample-chat.md
├── 2025-10-07_beeper-matrix_product-updates.md
└── 2025-10-07_whatsapp_project-team.md
```

## Markdown Format

Each archive includes:

### Header Section
- Chat title, network, ID
- Participant count
- Last activity timestamp
- Total message count

### Participants List
- All chat participants with names and IDs

### Messages (Chronologically Sorted)
- Date headers for organization
- Message numbers for reference
- Sender name and timestamp
- Full message text (quoted format)
- Attachments with file info
- Reactions

### Footer
- Archive generation timestamp
- Summary statistics

## Example Archive

```markdown
# Sample DM

**Network:** instagram

**Chat ID:** `!chat123:beeper.local`

**Participants:** 2

**Last Activity:** 2025-10-07T14:30:00Z

**Total Messages:** 150

---

## Participants

1. **User One** (`@user1:beeper.local`)
2. **User Two** (`@user2:beeper.local`)

---

## Messages

### 📅 Monday, October 7, 2025

#### Message #1

**From:** User One  
**Time:** 09:15:23  
**Message ID:** `msg_123456`

> Hello! How are you?

---

#### Message #2

**From:** User Two  
**Time:** 09:16:45  
**Message ID:** `msg_123457`

> I'm doing great, thanks!

**Reactions:** 👍, ❤️

---
```

## Features in Detail

### Interactive Selection
The tool displays all your chats with:
- Chat title (truncated to 40 chars)
- Network type (Instagram, WhatsApp, etc.)
- Participant count
- Unread message count (if any)

### Progress Indicators
For large chats (500+ messages), progress updates are shown:
```
[1/5] Archiving: Team Standup
  → Fetched 500 messages...
  → Fetched 1000 messages...
   ✓ Archived to: 2025-10-07_instagram_team-standup.md
```

### Filename Sanitization
- Invalid characters replaced with underscores
- Network name in lowercase with dashes
- Date prefix for organization
- Truncated titles (max 50 chars)

### Markdown Features
- Properly escaped special characters
- Block quotes for message text
- Multi-line messages preserved
- File size formatting (KB, MB, GB)
- Clean separator lines

## Requirements

- Go 1.21 or higher
- Beeper Desktop running locally
- Valid access token

## Tips

- Archive chats regularly to preserve history
- Add `chat-archives/` to `.gitignore` to avoid committing private exports
- Large chats may take a few minutes to fetch all messages
- Archives can be viewed in any markdown viewer or editor

## Troubleshooting

**"Failed to create client"**
- Make sure `BEEPER_ACCESS_TOKEN` is set
- Verify Beeper Desktop is running

**"Failed to fetch chats"**
- Check your network connection
- Verify your access token is valid
- Make sure Beeper Desktop API is accessible

**Large chats timing out**
- The tool has a 30-second timeout per request
- For very large chats, consider archiving fewer at once

## License

Same as the main Beeper Desktop API project.
