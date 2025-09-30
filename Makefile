# Database AI Assistant - Go Version Makefile

.PHONY: help setup build run clean install dev test release release-patch release-minor release-major

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
	@echo "  test      - 🧪 Run tests"
	@echo "  install   - 📦 Install/update dependencies"
	@echo "  clean     - 🧹 Clean build artifacts"
	@echo ""
	@echo "Release commands:"
	@echo "  release-patch  - 🏷️  Create patch release (x.x.X)"
	@echo "  release-minor  - 🏷️  Create minor release (x.X.0)"
	@echo "  release-major  - 🏷️  Create major release (X.0.0)"
	@echo ""
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

# Run tests
test:
	@echo "🧪 Running tests..."
	@go test -v ./...

# Release commands
release-patch:
	@echo "🏷️ Creating patch release..."
	@./scripts/release.sh patch

release-minor:
	@echo "🏷️ Creating minor release..."
	@./scripts/release.sh minor

release-major:
	@echo "🏷️ Creating major release..."
	@./scripts/release.sh major

# Quick start (setup + run)
start: setup
	@echo ""
	@echo "🚨 Please edit the .env file with your actual credentials, then run:"
	@echo "   make run"
