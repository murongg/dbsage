# Database AI Assistant - Go Version Makefile

.PHONY: help setup build run clean install dev test test-coverage test-race test-utils test-models test-database test-ai test-ui benchmark benchmark-utils benchmark-ai benchmark-ui lint check fmt release release-patch release-minor release-major

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
	@echo ""
	@echo "Code quality:"
	@echo "  lint      - 🔍 Run code quality checks"
	@echo "  fmt       - 📝 Format code"
	@echo "  check     - 🎯 Run lint + tests"
	@echo ""
	@echo "Testing:"
	@echo "  test           - 🧪 Run all tests"
	@echo "  test-coverage  - 📊 Run tests with coverage report"
	@echo "  test-race      - 🏃 Run tests with race detection"
	@echo "  test-utils     - 🧪 Run utils tests only"
	@echo "  test-models    - 🧪 Run models tests only"
	@echo "  test-database  - 🧪 Run database tests only"
	@echo "  test-ai        - 🧪 Run AI tests only"
	@echo "  test-ui        - 🧪 Run UI tests only"
	@echo ""
	@echo "Benchmarking:"
	@echo "  benchmark      - 🚀 Run all benchmarks"
	@echo "  benchmark-*    - 🚀 Run specific package benchmarks"
	@echo ""
	@echo "Maintenance:"
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
	@rm -f coverage.out coverage.html
	@go clean
	@echo "✅ Clean complete!"

# Run linting and code quality checks
lint:
	@echo "🔍 Running code quality checks..."
	@echo "  📝 Checking code format..."
	@if [ -n "$$(gofmt -l .)" ]; then echo "❌ Code formatting issues found. Run 'make fmt' to fix."; gofmt -l .; exit 1; else echo "✅ Code format OK"; fi
	@echo "  🔍 Running go vet..."
	@go vet ./... && echo "✅ go vet passed" || (echo "❌ go vet failed" && exit 1)
	@echo "  🔨 Checking compilation..."
	@go build ./... && echo "✅ Build successful" || (echo "❌ Build failed" && exit 1)
	@echo "  📋 Checking imports..."
	@go mod tidy && echo "✅ Imports clean" || (echo "❌ Import issues" && exit 1)
	@echo "✅ All lint checks completed!"

# Comprehensive check including tests
check: lint test
	@echo "🎯 All checks passed!"

# Format code
fmt:
	@echo "📝 Formatting code..."
	@gofmt -w .
	@echo "✅ Code formatted!"

# Run tests
test:
	@echo "🧪 Running all tests..."
	@go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "🧪 Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

# Run tests for specific packages
test-utils:
	@echo "🧪 Running utils tests..."
	@go test -v ./internal/utils/...

test-models:
	@echo "🧪 Running models tests..."
	@go test -v ./internal/models/...

test-database:
	@echo "🧪 Running database tests..."
	@go test -v ./pkg/database/...

test-ai:
	@echo "🧪 Running AI tests..."
	@go test -v ./internal/ai/...

test-ui:
	@echo "🧪 Running UI tests..."
	@go test -v ./internal/ui/...

# Run tests with race detection
test-race:
	@echo "🧪 Running tests with race detection..."
	@go test -race -v ./...

# Run benchmarks
benchmark:
	@echo "🚀 Running benchmarks..."
	@go test -bench=. -benchmem ./...

# Run specific benchmark
benchmark-utils:
	@echo "🚀 Running utils benchmarks..."
	@go test -bench=. -benchmem ./internal/utils/...

benchmark-ai:
	@echo "🚀 Running AI benchmarks..."
	@go test -bench=. -benchmem ./internal/ai/...

benchmark-ui:
	@echo "🚀 Running UI benchmarks..."
	@go test -bench=. -benchmem ./internal/ui/...

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
