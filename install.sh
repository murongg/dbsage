#!/bin/bash

# DBSage One-Click Installation Script
# Database AI Assistant - One-Click Installation Script
# 
# Supported OS: Linux, macOS
# Supported Architecture: amd64, arm64
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/murongg/dbsage/main/install.sh | bash
#   or download and run: bash install.sh

set -e

# Color definitions
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Icon definitions
SUCCESS="✅"
ERROR="❌"
INFO="ℹ️"
WARNING="⚠️"
ROCKET="🚀"
WRENCH="🔧"
PACKAGE="📦"
SPARKLES="✨"

# Configuration variables
REPO_URL="https://github.com/murongg/dbsage"
BINARY_NAME="dbsage"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="$HOME/.dbsage"
VERSION="latest"
FORCE_INSTALL=false
INSTALL_GLOBAL=true

# Print colored messages
print_message() {
    local color=$1
    local icon=$2
    local message=$3
    echo -e "${color}${icon} ${message}${NC}"
}

print_success() {
    print_message "$GREEN" "$SUCCESS" "$1"
}

print_error() {
    print_message "$RED" "$ERROR" "$1"
}

print_info() {
    print_message "$BLUE" "$INFO" "$1"
}

print_warning() {
    print_message "$YELLOW" "$WARNING" "$1"
}

print_header() {
    echo -e "${PURPLE}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "                    ${SPARKLES} DBSage Database AI Assistant ${SPARKLES}"
    echo "                         One-Click Installation"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo -e "${NC}"
}

# Show help information
show_help() {
    echo "DBSage One-Click Installation Script"
    echo ""
    echo "Usage: $0 [options]"
    echo ""
    echo "Options:"
    echo "  -h, --help          Show help information"
    echo "  -v, --version VER   Specify version (default: latest)"
    echo "  -d, --dir DIR       Specify installation directory (default: /usr/local/bin)"
    echo "  -f, --force         Force reinstallation"
    echo "  --local             Install to current directory"
    echo "  --no-config         Skip configuration file creation"
    echo ""
    echo "Examples:"
    echo "  $0                  # Standard installation"
    echo "  $0 --local          # Local installation"
    echo "  $0 -f               # Force reinstallation"
    echo ""
}

# Detect operating system and architecture
detect_platform() {
    local os arch

    # Detect operating system
    case "$(uname -s)" in
        Linux*)     os="linux" ;;
        Darwin*)    os="darwin" ;;
        CYGWIN*|MINGW*|MSYS*) os="windows" ;;
        *)          print_error "Unsupported operating system: $(uname -s)"; exit 1 ;;
    esac

    # Detect architecture
    case "$(uname -m)" in
        x86_64|amd64)   arch="amd64" ;;
        arm64|aarch64)  arch="arm64" ;;
        armv7l)         arch="arm" ;;
        *)              print_error "Unsupported architecture: $(uname -m)"; exit 1 ;;
    esac

    echo "${os}_${arch}"
}

# Check if already installed
check_existing_installation() {
    if command -v "$BINARY_NAME" >/dev/null 2>&1 && [ "$FORCE_INSTALL" = false ]; then
        print_warning "DBSage is already installed on the system"
        local existing_version
        existing_version=$("$BINARY_NAME" --version 2>/dev/null || echo "Unknown version")
        echo "Current version: $existing_version"
        echo ""
        echo "Use --force option to reinstall"
        exit 0
    fi
}

# Check permissions
check_permissions() {
    if [ "$INSTALL_GLOBAL" = true ] && [ ! -w "$INSTALL_DIR" ]; then
        print_warning "Administrator privileges required to write to $INSTALL_DIR"
        print_info "Please enter administrator password or use --local option for local installation"
        return 1
    fi
    return 0
}

# Check dependencies
check_dependencies() {
    print_info "Checking system dependencies..."
    
    local missing_deps=()
    
    # Check curl or wget
    if ! command -v curl >/dev/null 2>&1 && ! command -v wget >/dev/null 2>&1; then
        missing_deps+=("curl_or_wget")
    fi
    
    # Check tar
    if ! command -v tar >/dev/null 2>&1; then
        missing_deps+=("tar")
    fi
    
    # Check unzip (for potential future use)
    if ! command -v unzip >/dev/null 2>&1; then
        missing_deps+=("unzip")
    fi
    
    if [ ${#missing_deps[@]} -gt 0 ]; then
        print_error "Missing required dependencies:"
        for dep in "${missing_deps[@]}"; do
            case $dep in
                "curl_or_wget")
                    echo "  - curl or wget: Please install curl or wget for downloading files"
                    ;;
                "tar")
                    echo "  - tar: Please install tar for extracting archives"
                    ;;
                "unzip")
                    echo "  - unzip: Please install unzip for extracting archives"
                    ;;
            esac
        done
        exit 1
    fi
    
    print_success "All dependency checks passed"
}

# Get latest release version
get_latest_version() {
    local api_url="https://api.github.com/repos/murongg/dbsage/releases/latest"
    local version
    
    if command -v curl >/dev/null 2>&1; then
        version=$(curl -s "$api_url" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    elif command -v wget >/dev/null 2>&1; then
        version=$(wget -qO- "$api_url" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    else
        print_error "Neither curl nor wget is available"
        exit 1
    fi
    
    if [ -z "$version" ]; then
        print_error "Failed to get latest version from GitHub API"
        exit 1
    fi
    
    echo "$version"
}

# Create temporary directory
create_temp_dir() {
    local temp_dir
    temp_dir=$(mktemp -d 2>/dev/null || mktemp -d -t 'dbsage_install')
    echo "$temp_dir"
}

# Download binary from GitHub releases
download_binary() {
    local temp_dir=$1
    local version=$2
    local platform=$3
    local original_dir=$(pwd)
    
    print_info "Downloading DBSage binary for $platform..."
    
    cd "$temp_dir"
    
    # Determine archive name and format based on platform
    local archive_name
    local extract_cmd
    if [[ "$platform" == *"windows"* ]]; then
        archive_name="dbsage_${platform}.zip"
        extract_cmd="unzip -q"
    else
        archive_name="dbsage_${platform}.tar.gz"
        extract_cmd="tar -xzf"
    fi
    
    local download_url="https://github.com/murongg/dbsage/releases/download/${version}/${archive_name}"
    local cache_dir="$HOME/.cache/dbsage"
    local cached_file="$cache_dir/${version}_${archive_name}"
    
    # Create cache directory if it doesn't exist
    mkdir -p "$cache_dir"
    
    # Check if we have a cached version
    if [ -f "$cached_file" ]; then
        print_info "Using cached binary: $cached_file"
        cp "$cached_file" "$archive_name"
    else
        print_info "Download URL: $download_url"
        
        # Download the archive
        if command -v curl >/dev/null 2>&1; then
            curl -L -o "$archive_name" "$download_url"
        elif command -v wget >/dev/null 2>&1; then
            wget -O "$archive_name" "$download_url"
        else
            print_error "Neither curl nor wget is available"
            exit 1
        fi
        
        # Cache the downloaded file for future use
        if [ $? -eq 0 ]; then
            cp "$archive_name" "$cached_file"
            print_info "Binary cached for future installations"
        fi
    fi
    
    if [ $? -ne 0 ]; then
        print_error "Failed to download binary from $download_url"
        print_info "Available releases: https://github.com/murongg/dbsage/releases"
        exit 1
    fi
    
    # Extract the archive
    print_info "Extracting binary..."
    $extract_cmd "$archive_name"
    
    if [ $? -ne 0 ]; then
        print_error "Failed to extract archive"
        exit 1
    fi
    
    # Handle Windows executable extension
    local binary_file="$BINARY_NAME"
    if [[ "$platform" == *"windows"* ]]; then
        binary_file="${BINARY_NAME}.exe"
    fi
    
    # Make binary executable
    chmod +x "$binary_file"
    
    # Return to original directory
    cd "$original_dir"
    
    print_success "Binary download and extraction completed"
}

# Install binary
install_binary() {
    local temp_dir=$1
    local platform=$2
    
    # Determine source binary name (handle Windows .exe extension)
    local source_binary_name="$BINARY_NAME"
    if [[ "$platform" == *"windows"* ]]; then
        source_binary_name="${BINARY_NAME}.exe"
    fi
    local source_binary="$temp_dir/$source_binary_name"
    
    # Verify source binary exists
    if [ ! -f "$source_binary" ]; then
        print_error "Source binary not found: $source_binary"
        print_info "Available files in temp directory:"
        ls -la "$temp_dir"
        exit 1
    fi
    
    # Determine target binary name
    local target_binary_name="$BINARY_NAME"
    if [[ "$platform" == *"windows"* ]]; then
        target_binary_name="${BINARY_NAME}.exe"
    fi
    
    if [ "$INSTALL_GLOBAL" = true ]; then
        print_info "Installing DBSage to $INSTALL_DIR..."
        
        if [ ! -w "$INSTALL_DIR" ]; then
            sudo cp "$source_binary" "$INSTALL_DIR/$target_binary_name"
            sudo chmod +x "$INSTALL_DIR/$target_binary_name"
        else
            cp "$source_binary" "$INSTALL_DIR/$target_binary_name"
            chmod +x "$INSTALL_DIR/$target_binary_name"
        fi
        
        print_success "DBSage installed to $INSTALL_DIR/$target_binary_name"
    else
        print_info "Installing DBSage to current directory..."
        # Remove existing binary if it exists to avoid conflicts
        rm -f "./$target_binary_name"
        
        # Copy binary to current directory
        if cp "$source_binary" "./$target_binary_name"; then
            chmod +x "./$target_binary_name"
            print_success "DBSage installed to $(pwd)/$target_binary_name"
        else
            print_error "Failed to install binary to $(pwd)/$target_binary_name"
            exit 1
        fi
    fi
}

# Create configuration files
create_config() {
    if [ "$SKIP_CONFIG" = true ]; then
        return
    fi
    
    print_info "Creating configuration directory and files..."
    
    # Create configuration directory
    mkdir -p "$CONFIG_DIR"
    
    # Create example configuration file
    cat > "$CONFIG_DIR/config.env" << 'EOF'
# DBSage Configuration File
# Please modify the following configuration as needed

# OpenAI API Configuration
OPENAI_API_KEY=your_openai_api_key_here
OPENAI_BASE_URL=https://api.openai.com/v1

# Database Configuration (optional, can also be added at runtime)
# DATABASE_URL=postgres://username:password@localhost:5432/database?sslmode=disable

# Log Level (optional)
# LOG_LEVEL=info

# Other Configuration
# MAX_CONNECTIONS=10
# TIMEOUT=30s
EOF
    
    # Create connection configuration file
    cat > "$CONFIG_DIR/connections.json" << 'EOF'
{}
EOF
    
    print_success "Configuration files created in $CONFIG_DIR/"
    print_info "Please edit $CONFIG_DIR/config.env file to set your OpenAI API Key"
}

# Setup environment variables
setup_environment() {
    local shell_profile=""
    
    # Detect current shell and determine configuration file
    case "$SHELL" in
        */bash)
            if [ -f "$HOME/.bashrc" ]; then
                shell_profile="$HOME/.bashrc"
            elif [ -f "$HOME/.bash_profile" ]; then
                shell_profile="$HOME/.bash_profile"
            fi
            ;;
        */zsh)
            shell_profile="$HOME/.zshrc"
            ;;
        */fish)
            shell_profile="$HOME/.config/fish/config.fish"
            ;;
    esac
    
    if [ -n "$shell_profile" ] && [ "$INSTALL_GLOBAL" = false ]; then
        print_info "Adding DBSage to PATH..."
        local dbsage_path="$(pwd)"
        
        # Check if already added
        if ! grep -q "# DBSage" "$shell_profile" 2>/dev/null; then
            echo "" >> "$shell_profile"
            echo "# DBSage - Database AI Assistant" >> "$shell_profile"
            echo "export PATH=\"$dbsage_path:\$PATH\"" >> "$shell_profile"
            print_success "PATH updated, please run 'source $shell_profile' or reopen terminal"
        fi
    fi
}

# Verify installation
verify_installation() {
    print_info "Verifying installation..."
    
    local dbsage_cmd
    if [ "$INSTALL_GLOBAL" = true ]; then
        dbsage_cmd="$BINARY_NAME"
    else
        dbsage_cmd="./$BINARY_NAME"
    fi
    
    if command -v "$dbsage_cmd" >/dev/null 2>&1; then
        local version_output
        version_output=$("$dbsage_cmd" --version 2>/dev/null || echo "DBSage installed successfully")
        print_success "Installation verification passed: $version_output"
        return 0
    else
        print_error "Installation verification failed"
        return 1
    fi
}

# Show post-installation instructions
show_post_install_instructions() {
    echo ""
    print_success "🎉 DBSage installation completed!"
    echo ""
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${YELLOW}${ROCKET} Quick Start:${NC}"
    echo ""
    echo "1. ${BLUE}Configure OpenAI API Key:${NC}"
    echo "   Edit configuration file: ${GREEN}$CONFIG_DIR/config.env${NC}"
    echo "   Set: ${CYAN}OPENAI_API_KEY=your_actual_api_key${NC}"
    echo ""
    echo "2. ${BLUE}Start DBSage:${NC}"
    if [ "$INSTALL_GLOBAL" = true ]; then
        echo "   ${GREEN}dbsage${NC}"
    else
        echo "   ${GREEN}./dbsage${NC}"
        echo "   Or reopen terminal and run directly: ${GREEN}dbsage${NC}"
    fi
    echo ""
    echo "3. ${BLUE}Add database connection:${NC}"
    echo "   Run in DBSage: ${CYAN}/add mydb${NC}"
    echo "   Then follow the prompts to enter database connection information"
    echo ""
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${YELLOW}${INFO} More Information:${NC}"
    echo ""
    echo "• ${BLUE}Configuration directory:${NC} $CONFIG_DIR"
    echo "• ${BLUE}Documentation:${NC} https://github.com/murongg/dbsage"
    echo "• ${BLUE}Issue reports:${NC} https://github.com/murongg/dbsage/issues"
    echo ""
    echo -e "${GREEN}Thank you for using DBSage!${NC} ${SPARKLES}"
    echo ""
}

# Clean up temporary files
cleanup() {
    local temp_dir=$1
    if [ -n "$temp_dir" ] && [ -d "$temp_dir" ]; then
        print_info "Cleaning up temporary files..."
        rm -rf "$temp_dir"
    fi
}

# Main installation function
main() {
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            -v|--version)
                VERSION="$2"
                shift 2
                ;;
            -d|--dir)
                INSTALL_DIR="$2"
                shift 2
                ;;
            -f|--force)
                FORCE_INSTALL=true
                shift
                ;;
            --local)
                INSTALL_GLOBAL=false
                INSTALL_DIR="."
                shift
                ;;
            --no-config)
                SKIP_CONFIG=true
                shift
                ;;
            *)
                print_error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    # Start installation process
    print_header
    
    # Check existing installation
    check_existing_installation
    
    # Check system platform
    local platform
    platform=$(detect_platform)
    print_info "Detected platform: $platform"
    
    # Check dependencies
    check_dependencies
    
    # Get version to download
    local target_version
    if [ "$VERSION" = "latest" ]; then
        print_info "Getting latest release version..."
        target_version=$(get_latest_version)
        print_info "Latest version: $target_version"
    else
        target_version="$VERSION"
        print_info "Target version: $target_version"
    fi
    
    # Check permissions
    if ! check_permissions; then
        print_warning "Switching to local installation mode"
        INSTALL_GLOBAL=false
        INSTALL_DIR="."
    fi
    
    # Create temporary directory
    local temp_dir
    temp_dir=$(create_temp_dir)
    
    # Set cleanup trap
    trap "cleanup '$temp_dir'" EXIT
    
    # Download binary
    download_binary "$temp_dir" "$target_version" "$platform"
    
    # Install
    install_binary "$temp_dir" "$platform"
    
    # Create configuration
    create_config
    
    # Setup environment
    setup_environment
    
    # Verify installation
    if verify_installation; then
        show_post_install_instructions
    else
        print_error "Installation may have issues, please check error messages"
        exit 1
    fi
}

# Check if running as script
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
