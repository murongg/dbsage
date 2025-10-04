#!/bin/bash

# Database AI Assistant - Go Version

echo "🚀 Starting Database AI Assistant (Go Version)..."

# Try to load .env file if it exists
if [ -f ".env" ]; then
    echo "📄 Loading environment variables from .env file..."
    set -a; source .env; set +a
fi

# Check configuration status and provide helpful information
config_warnings=0

if [ -z "$OPENAI_API_KEY" ]; then
    echo "⚠️  Warning: OPENAI_API_KEY not set - AI features will be disabled"
    echo "   You can still use database commands (/add, /list, /switch)"
    echo "   To enable AI: export OPENAI_API_KEY=\"your_api_key_here\""
    config_warnings=$((config_warnings + 1))
fi

# Check if there's a base URL override
if [ -n "$OPENAI_BASE_URL" ]; then
    echo "🔗 Using custom OpenAI base URL: $OPENAI_BASE_URL"
fi

# Note: DATABASE_URL is no longer required as connections are managed through the app
if [ -z "$DATABASE_URL" ] && [ $config_warnings -eq 0 ]; then
    echo "💡 Tip: Add database connections using '/add <name>' command in the app"
fi

if [ $config_warnings -gt 0 ]; then
    echo ""
    echo "📋 Quick setup guide:"
    echo "   1. Set environment variable: export OPENAI_API_KEY=your_api_key_here"
    echo "   2. Start DBSage and use '/add mydb' to add database connections"
    echo "   3. Follow the in-app guidance for complete setup"
    echo ""
fi

# Install dependencies if go.mod exists
if [ -f "go.mod" ]; then
    echo "📦 Installing Go dependencies..."
    go mod tidy
    go mod download
fi

# Build and run
echo "🔨 Building application..."
go build -o dbsage ./cmd/dbsage/main.go

if [ $? -eq 0 ]; then
    echo "✅ Build successful! Starting application..."
    echo ""
    echo "📱 DBSage is starting..."
    if [ -z "$OPENAI_API_KEY" ]; then
        echo "   → You'll see guidance on setting up your API key"
    fi
    echo "   → Press 'q' to dismiss guidance messages"
    echo "   → Type '?' or '/help' for available commands"
    echo "   → Press Ctrl+C to exit"
    echo ""
    ./dbsage
else
    echo "❌ Build failed!"
    exit 1
fi
