# Database AI Assistant - Go Version Makefile

.PHONY: help setup build run clean install dev

# Default target
help:
	@echo "🤖 Database AI Assistant - Go Version"
	@echo "===================================="
	@echo ""
	@echo "Available commands:"
	@echo "  setup     - 🔧 Run initial setup and create .env file"
	@echo "  build     - 🔨 Build the application"
	@echo "  run       - 🚀 Run the application"
	@echo "  dev       - 💻 Run in development mode (go run)"
	@echo "  install   - 📦 Install/update dependencies"
	@echo "  clean     - 🧹 Clean build artifacts"
	@echo "  help      - 📖 Show this help message"
	@echo ""

# Setup environment
setup:
	@./scripts/setup.sh

# Build the application
build:
	@echo "🔨 Building dbsage..."
	@go build -o dbsage cmd/dbsage/main.go
	@echo "✅ Build complete! Binary: ./dbsage"

# Run the application
run: build
	@./scripts/run.sh

# Development mode - run with go run
dev:
	@echo "💻 Running in development mode..."
	@go run cmd/dbsage/main.go

# Install/update dependencies
install:
	@echo "📦 Installing dependencies..."
	@go mod tidy
	@go mod download
	@echo "✅ Dependencies installed!"

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -f dbsage
	@go clean
	@echo "✅ Clean complete!"

# Quick start (setup + run)
start: setup
	@echo ""
	@echo "🚨 Please edit the .env file with your actual credentials, then run:"
	@echo "   make run"
