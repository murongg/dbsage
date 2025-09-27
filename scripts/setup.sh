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
# OpenAI Configuration
OPENAI_API_KEY=your_openai_api_key_here
OPENAI_BASE_URL=https://api.openai.com/v1

# Database Configuration
DATABASE_URL=postgres://username:password@localhost:5432/database?sslmode=disable

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
echo "Next steps:"
echo "1. Edit the .env file with your actual credentials:"
echo "   - Set your OPENAI_API_KEY"
echo "   - Set your DATABASE_URL"
echo ""
echo "2. Run the application:"
echo "   make run    # or"
echo "   make dev    # for development mode"
echo ""
echo "For help, run: make help"
