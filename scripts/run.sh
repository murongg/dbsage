#!/bin/bash

# Database AI Assistant - Go Version

echo "🚀 Starting Database AI Assistant (Go Version)..."

# Try to load .env file if it exists
if [ -f ".env" ]; then
    echo "📄 Loading environment variables from .env file..."
    export $(cat .env | grep -v '^#' | grep -v '^$' | xargs)
fi

# Check if environment variables are set
if [ -z "$OPENAI_API_KEY" ]; then
    echo "❌ Error: OPENAI_API_KEY environment variable is not set"
    echo ""
    echo "Please either:"
    echo "1. Set it with: export OPENAI_API_KEY=\"your_api_key_here\""
    echo "2. Create a .env file with: OPENAI_API_KEY=your_api_key_here"
    echo "3. Copy configs/config.example to .env and edit it"
    exit 1
fi

if [ -z "$DATABASE_URL" ]; then
    echo "❌ Error: DATABASE_URL environment variable is not set"
    echo ""
    echo "Please either:"
    echo "1. Set it with: export DATABASE_URL=\"postgres://user:pass@localhost:5432/db?sslmode=disable\""
    echo "2. Create a .env file with the DATABASE_URL"
    echo "3. Copy configs/config.example to .env and edit it"
    exit 1
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
    ./dbsage
else
    echo "❌ Build failed!"
    exit 1
fi
