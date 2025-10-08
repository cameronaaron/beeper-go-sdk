# Chat Archive Tool - Quick Start Guide

## Installation

```bash
cd Golang/cmd/archive-chats
go build -o archive-chats
```

## Quick Usage Examples

### Archive All Chats

```bash
export BEEPER_ACCESS_TOKEN="your-token-here"
echo "a" | ./archive-chats
```

### Archive Specific Chats

```bash
# Archive chats #1, #5, and #10
export BEEPER_ACCESS_TOKEN="your-token-here"
echo "1,5,10" | ./archive-chats
```

### Interactive Mode

```bash
export BEEPER_ACCESS_TOKEN="your-token-here"
./archive-chats
# Then choose your chats interactively
```

## Output Location

All archives are saved to `chat-archives/` directory with filenames like:

- `2025-10-07_instagram_sample-chat.md`
- `2025-10-07_beeper-matrix_product-updates.md`
- `2025-10-07_whatsapp_project-team.md`

## Features Included in Each Archive

✅ **Complete Chat History**

- All messages in chronological order
- Sender names and timestamps
- Message IDs for reference

✅ **Rich Metadata**

- Chat title, network, and ID
- Full participant list
- Last activity timestamp
- Unread message counts

✅ **Media Information**

- Attachment filenames
- File sizes (formatted)
- Media types
- Source URLs

✅ **Interactions**

- Reactions on messages
- Multi-line messages preserved

✅ **Beautiful Formatting**

- Date headers (e.g., "📅 Monday, October 7, 2025")
- Message numbers for easy reference
- Block quotes for readability
- Clean separators

## Sample Output

Here's what a chat archive looks like:

```markdown
# Beeper Updates

**Network:** Beeper (Matrix)
**Chat ID:** `!updates:beeper.local`
**Participants:** 6
**Total Messages:** 20

---

## Participants

1. **Beeper Help** (`@help:beeper.com`)
2. **Product Lead** (`@product:beeper.com`)
...

---

## Messages

### 📅 Monday, October 7, 2025

#### Message #1

**From:** User One  
**Time:** 14:30:15  
**Message ID:** `msg_123456`

> Hello! This is my message.

---
```

## Tips

1. **Large Chats**: The tool shows progress for chats with 500+ messages
2. **Filename Safety**: Special characters are automatically sanitized
3. **Version Control**: Add `chat-archives/` to `.gitignore` so private exports stay local
4. **Markdown Viewers**: Archives look great in:
   - VS Code (with Markdown Preview)
   - Obsidian
   - Typora
   - GitHub/GitLab (if you push them)
   - Any text editor

## What Gets Archived?

| Feature | Included | Format |
|---------|----------|--------|
| Message Text | ✅ | Block quote |
| Timestamps | ✅ | HH:MM:SS |
| Sender Name | ✅ | Bold header |
| Attachments | ✅ | List with details |
| Reactions | ✅ | Emoji list |
| Participants | ✅ | Numbered list |
| Message IDs | ✅ | Code format |
| File Sizes | ✅ | Human-readable (KB/MB) |

## Automation

You can automate archiving with a cron job:

```bash
# Archive all chats daily at 2 AM
0 2 * * * cd /path/to/archive-chats && echo "a" | BEEPER_ACCESS_TOKEN="your-token" ./archive-chats
```

## Troubleshooting

**Chat list appears empty:**

- Ensure Beeper Desktop is running
- Check your access token is valid

**Archive fails midway:**

- Network issue - the tool will retry automatically
- Large chat timeout - try archiving fewer chats at once

**Special characters in filename:**

- These are automatically replaced with underscores
- Network names are lowercase with dashes

## Advanced Usage

### Archive only unread chats

Currently, you'll need to note the chat numbers with "[X unread]" and enter those numbers manually.

### Archive by network

Note the chat numbers for specific networks (Instagram, WhatsApp, etc.) and enter those.

### Batch processing

```bash
# Archive chats 1-10
echo "1,2,3,4,5,6,7,8,9,10" | ./archive-chats

# Or use a loop for larger batches
for i in {1..50}; do
  echo "$i" | BEEPER_ACCESS_TOKEN="your-token" ./archive-chats
  sleep 1  # Be nice to the API
done
```

## Example Real-World Use Cases

1. **Before Account Deletion**: Archive everything before closing an account
2. **Legal/Compliance**: Keep records for documentation
3. **Backup**: Regular backups of important conversations
4. **Migration**: Moving to a new system but want to keep history
5. **Analysis**: Export for data analysis or searching
6. **Memory Keeping**: Save important group chats or personal conversations

## Viewing Archives

The markdown format means you can:

- Search with `grep` or any text search tool
- Convert to PDF with pandoc
- Import into note-taking apps
- View in browser with markdown preview
- Edit/annotate in any text editor
- Version control with git

Enjoy archiving your chats! 📦✨
