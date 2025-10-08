#!/bin/bash

# Beeper Chat Archive Tool - Demo Script
# This script demonstrates various ways to use the archive tool

echo "🎬 Beeper Chat Archive Tool - Demo"
echo "====================================="
echo ""

# Check if built
if [ ! -f "./archive-chats" ]; then
    echo "Building archive tool..."
    go build -o archive-chats
    echo "✅ Build complete"
    echo ""
fi

# Check for token
if [ -z "$BEEPER_ACCESS_TOKEN" ]; then
    echo "❌ Error: BEEPER_ACCESS_TOKEN not set"
    echo ""
    echo "Please set your token:"
    echo "  export BEEPER_ACCESS_TOKEN=\"your-token-here\""
    echo ""
    exit 1
fi

echo "Select a demo mode:"
echo ""
echo "1. Interactive mode (select chats manually)"
echo "2. Archive first 5 chats"
echo "3. Archive all chats"
echo "4. Show chat list only (quit without archiving)"
echo ""
read -p "Choice (1-4): " choice

case $choice in
    1)
        echo ""
        echo "🎯 Interactive Mode"
        echo "==================="
        echo "Select chats by number(s) or type 'a' for all"
        echo ""
        ./archive-chats
        ;;
    2)
        echo ""
        echo "🎯 Archiving First 5 Chats"
        echo "=========================="
        echo ""
        echo "1,2,3,4,5" | ./archive-chats
        ;;
    3)
        echo ""
        echo "🎯 Archiving All Chats"
        echo "====================="
        echo "⚠️  This may take a while for many chats!"
        echo ""
        read -p "Are you sure? (y/n): " confirm
        if [ "$confirm" = "y" ] || [ "$confirm" = "Y" ]; then
            echo "a" | ./archive-chats
        else
            echo "Cancelled."
        fi
        ;;
    4)
        echo ""
        echo "🎯 Chat List Preview"
        echo "==================="
        echo ""
        echo "q" | ./archive-chats
        ;;
    *)
        echo "Invalid choice"
        exit 1
        ;;
esac

echo ""
echo "📊 Archive Summary"
echo "=================="
if [ -d "chat-archives" ]; then
    count=$(ls -1 chat-archives/*.md 2>/dev/null | wc -l)
    echo "Total archives: $count"
    echo ""
    
    if [ $count -gt 0 ]; then
        echo "Recent archives:"
        ls -lht chat-archives/*.md | head -5
        echo ""
        
        echo "📁 View archives:"
        echo "  cd chat-archives"
        echo "  open *.md          # macOS"
        echo "  code *.md          # VS Code"
        echo ""
        
        echo "🔍 Search archives:"
        echo "  grep -r \"search term\" chat-archives/"
        echo ""
    fi
else
    echo "No archives yet!"
fi

echo "✨ Done!"
