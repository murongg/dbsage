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

# Create .env file from example
if [ ! -f ".env" ]; then
    if [ -f "configs/config.example" ]; then
        echo "📄 Creating .env file from configs/config.example..."
        cp configs/config.example .env
        echo "✅ .env file created!"
    else
        echo "📄 Creating .env file with default template..."
        cat > .env << 'EOF'
# OpenAI Configuration (Required for AI features)
# Get your API key from: https://platform.openai.com/api-keys
OPENAI_API_KEY=your_openai_api_key_here
OPENAI_BASE_URL=https://api.openai.com/v1

# Database Configuration (Optional - can be set up through the app)
# Note: You can add database connections using '/add <name>' command in the app
# DATABASE_URL=postgres://username:password@localhost:5432/database?sslmode=disable

# Optional: Application Configuration
# LOG_LEVEL=info
EOF
        echo "✅ .env file created with default template!"
    fi
else
    echo "⚠️  .env file already exists, skipping creation"
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
echo "1. 🔑 Get your OpenAI API key:"
echo "   → Visit: https://platform.openai.com/api-keys"
echo "   → Edit .env file and set: OPENAI_API_KEY=your_actual_key"
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
