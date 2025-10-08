# 📦 Beeper Chat Archive Tool

**A beautiful, feature-rich chat archiving tool for Beeper Desktop**

Export your Beeper chats to clean, readable markdown files with complete message history, metadata, and formatting.

---

## ✨ Key Features

### 📝 **Beautiful Markdown Export**
- Clean, professional formatting
- Chronologically organized by date
- Readable block quotes for messages
- Automatic filename sanitization
- Human-readable file sizes

### 🎯 **Complete Data Preservation**
- ✅ All messages in chronological order
- ✅ Full sender information
- ✅ Precise timestamps (HH:MM:SS)
- ✅ Message IDs for reference
- ✅ Attachment metadata (name, type, size, URL)
- ✅ Reactions and emoji
- ✅ Participant lists
- ✅ Chat metadata (network, ID, last activity)

### 🚀 **User-Friendly Interface**
- Interactive chat selection
- Preview all chats with participant counts
- See unread message indicators
- Archive all or select specific chats
- Progress indicators for large chats
- Automatic retry on failures

---

## 🎨 Sample Output

```markdown
# Beeper Updates

**Network:** Beeper (Matrix)
**Chat ID:** `!updates:beeper.local`
**Participants:** 6
**Last Activity:** 2025-10-06T20:31:50.543Z
**Total Messages:** 20

---

## Participants

1. **Beeper Help** (`@help:beeper.com`)
2. **Brad Murray** (`@brad:beeper.com`)
3. **Eric Migicovsky** (`@eric:beeper.com`)

---

## Messages

### 📅 Wednesday, October 8, 2025

#### Message #1

**From:** Alex Example  
**Time:** 02:17:55  
**Message ID:** `msg_0001`

> Excited to try the new archive tool tonight!

**Reactions:** 👍, 🎉

---

#### Message #2

**From:** Sam Sample  
**Time:** 02:19:01  
**Message ID:** `msg_0002`

> Same here—love how clean the markdown looks.

**Attachments:**

- Attachment 1: `design.png` (img) - 21.8 KB
  - URL: file:///path/to/file

---
```

---

## 🚀 Quick Start

### Installation
```bash
cd ~/projects/desktop-api-go/Golang/cmd/archive-chats
go build -o archive-chats
```

### Run
```bash
export BEEPER_ACCESS_TOKEN="your-token-here"
./archive-chats
```

### Select Chats
- Type `a` to archive all chats
- Enter numbers like `1,3,5` for specific chats
- Type `q` to quit

---

## 📂 Output Structure

Archives are saved to `chat-archives/` with organized filenames:

```
chat-archives/
├── 2025-10-07_instagram_team-planning.md
├── 2025-10-07_whatsapp_project-sync.md
├── 2025-10-07_beeper-matrix_updates.md
└── 2025-10-07_google-messages_vendor-checkins.md
```

**Filename Format:** `DATE_NETWORK_CHAT-TITLE.md`

---

## 🎯 Use Cases

| Use Case | Description |
|----------|-------------|
| 🗂️ **Backup** | Regular backups of important conversations |
| 📜 **Legal/Compliance** | Document retention for legal purposes |
| 🔄 **Migration** | Preserve history when switching platforms |
| 🔍 **Search/Analysis** | Full-text search with grep or other tools |
| 💾 **Long-term Storage** | Archive chats before account deletion |
| 📚 **Memory Keeping** | Save sentimental group chats |
| 🔬 **Data Analysis** | Export for analysis or research |

---

## 🎨 What Makes It Beautiful?

### Visual Organization
- **Date Headers**: `📅 Monday, October 7, 2025`
- **Message Numbers**: Easy reference (`Message #1`, `Message #2`)
- **Clean Separators**: Markdown horizontal rules
- **Emoji Support**: Full Unicode emoji preservation

### Smart Formatting
- **Block Quotes**: Messages indented with `>` for clarity
- **Bold Headers**: Sender names and metadata stand out
- **Code Blocks**: IDs and technical info in monospace
- **Lists**: Structured participants and attachments

### Data Richness
- **Timestamps**: Precise time for every message
- **File Sizes**: Human-readable (21.8 KB, not 22323 bytes)
- **URLs**: Direct links to attachments
- **Reactions**: Emoji reactions preserved

---

## 📊 Archive Contents

Each archive includes:

### 📌 Header Section
- Chat title (large heading)
- Network type
- Chat ID (code format)
- Participant count
- Last activity timestamp
- Total message count

### 👥 Participants Section
- Numbered list of all participants
- Full names (when available)
- User IDs

### 💬 Messages Section
- **Date Headers**: New section for each day
- **Message Metadata**:
  - Message number
  - Sender name
  - Timestamp (HH:MM:SS)
  - Message ID
- **Content**:
  - Full message text (block quoted)
  - Attachments list with details
  - Reactions
- **Separators**: Clean horizontal rules

### 📋 Footer Section
- Archive generation timestamp
- Message count summary
- Tool attribution

---

## 🔧 Advanced Features

### Automatic Pagination
The tool automatically fetches all messages, handling pagination transparently:
- Fetches 100 messages per request
- Continues until all messages retrieved
- Shows progress for chats with 500+ messages

### Error Handling
- Automatic retry on network errors (up to 3 attempts)
- 30-second timeout per request
- Graceful error messages
- Continues with other chats if one fails

### Filename Sanitization
- Invalid characters → underscores
- Network names → lowercase with dashes
- Title truncation (50 chars max)
- Date prefix for organization

---

## 🎓 Tips & Tricks

### Viewing Archives

**Best Markdown Viewers:**
- VS Code with Markdown Preview
- Obsidian
- Typora
- GitHub/GitLab (if pushed to repo)
- Any text editor

### Searching Archives
```bash
# Find all messages containing "important"
grep -r "important" chat-archives/

# Find messages from a specific person
grep -r "From: Alex" chat-archives/

# Count messages in an archive
grep -c "Message #" chat-archives/filename.md
```

### Converting to PDF
```bash
# Using pandoc
pandoc chat-archives/filename.md -o output.pdf
```

### Version Control
```bash
# Track archive history
git add chat-archives/
git commit -m "Archived chats from $(date +%Y-%m-%d)"
```

---

## 🔒 Privacy & Security

- Archives are saved locally only
- Contains full message text and metadata
- Includes attachment URLs (local file paths)
- Store in encrypted location if sensitive
- Consider `.gitignore` for private archives

---

## 📈 Performance

| Metric | Value |
|--------|-------|
| Messages per request | 100 |
| Request timeout | 30s |
| Max retries | 3 |
| Progress updates | Every 500 messages |
| Filename max length | 50 chars (title) |

**Typical Performance:**
- Small chat (< 100 messages): < 5 seconds
- Medium chat (500 messages): ~10 seconds
- Large chat (2000 messages): ~30-60 seconds

---

## 🆘 Troubleshooting

| Issue | Solution |
|-------|----------|
| "Failed to create client" | Set `BEEPER_ACCESS_TOKEN` environment variable |
| "Failed to fetch chats" | Ensure Beeper Desktop is running |
| Empty chat list | Verify access token is valid |
| Timeout on large chats | Archive fewer chats at once |
| Special chars in filename | Automatically sanitized to underscores |

---

## 🎉 Why Use This Tool?

### vs Manual Copy-Paste
- ✅ Preserves all metadata
- ✅ Maintains chronological order
- ✅ Includes reactions and attachments
- ✅ Formats beautifully
- ✅ Processes thousands of messages

### vs Screenshots
- ✅ Searchable text
- ✅ Version controllable
- ✅ Smaller file size
- ✅ Easy to share/backup
- ✅ Accessible format

### vs Database Export
- ✅ Human-readable
- ✅ No special tools needed
- ✅ Works with any text editor
- ✅ Beautiful formatting
- ✅ Easy to browse

---

## 🏗️ Built With

- **Language**: Go 1.21+
- **SDK**: Beeper Desktop API (Go port)
- **Format**: Markdown
- **Dependencies**: Minimal (only SDK)

---

## 📄 License

Same as Beeper Desktop API project.

---

## 🎁 Example Archive

See [`chat-archives/2025-10-07_beeper-(matrix)_Beeper Updates.md`](chat-archives/2025-10-07_beeper-(matrix)_Beeper%20Updates.md) for a real example!

---

Made with ❤️ using the Beeper Desktop API (Go port)
