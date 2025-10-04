#!/bin/bash

# Database AI Assistant - Setup Script

echo "🔧 Setting up Database AI Assistant (Go Version)..."

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21+ first."
    echo "Visit: https://golang.org/dl/"
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
REQUIRED_VERSION="1.21"
if ! printf '%s\n%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V -C; then
    echo "❌ Go version $GO_VERSION is too old. Please upgrade to Go 1.21+ first."
    exit 1
fi

echo "✅ Go version $GO_VERSION detected"

# Check if .env file exists (for backwards compatibility)
if [ -f ".env" ]; then
    echo "📄 Found existing .env file - loading environment variables..."
    set -a
    source .env
    set +a
    echo "✅ Environment variables loaded from .env"
else
    echo "ℹ️  No .env file found - will use system environment variables"
    echo "   Set OPENAI_API_KEY environment variable to enable AI features"
fi

# Install dependencies
echo "📦 Installing Go dependencies..."
go mod tidy
go mod download

if [ $? -eq 0 ]; then
    echo "✅ Dependencies installed successfully!"
else
    echo "❌ Failed to install dependencies"
    exit 1
fi

# Make scripts executable
echo "🔧 Making scripts executable..."
chmod +x scripts/run.sh
chmod +x scripts/setup.sh

echo ""
echo "🎉 Setup completed successfully!"
echo ""
echo "🚀 Quick Start Guide:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "1. 🔑 Configure OpenAI API Key:"
echo "   → Visit: https://platform.openai.com/api-keys"
echo "   → Set environment variable:"
echo "     export OPENAI_API_KEY=your_actual_key"
echo "   → For persistent configuration, add to ~/.zshrc or ~/.bashrc:"
echo "     echo 'export OPENAI_API_KEY=your_key' >> ~/.zshrc"
echo "     source ~/.zshrc"
echo ""
echo "2. 🏃 Run DBSage:"
echo "   → ./scripts/run.sh    # or make run"
echo ""
echo "3. 📚 First time usage:"
echo "   → The app will show setup guidance automatically"
echo "   → Use '/add mydb' to add database connections"
echo "   → Press 'q' to dismiss guidance messages"
echo "   → Type '?' for help anytime"
echo ""
echo "💡 Pro tip: You can start DBSage even without API key or database!"
echo "   The app will guide you through the setup process."
echo ""
echo "For advanced options, run: make help"
