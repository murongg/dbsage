#!/bin/bash

# DBSage Release Script
# Automates the process of creating a new release
#
# Usage:
#   ./scripts/release.sh [version]
#   ./scripts/release.sh --help
#
# Examples:
#   ./scripts/release.sh v1.0.0
#   ./scripts/release.sh patch    # Auto-increment patch version
#   ./scripts/release.sh minor    # Auto-increment minor version
#   ./scripts/release.sh major    # Auto-increment major version

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

# Icons
SUCCESS="✅"
ERROR="❌"
INFO="ℹ️"
WARNING="⚠️"
ROCKET="🚀"
PACKAGE="📦"
TAG="🏷️"

# Configuration
REPO_NAME="murongg/dbsage"
BINARY_NAME="dbsage"
BUILD_DIR="dist"
ARCHIVE_DIR="release"

# Print functions
print_success() {
    echo -e "${GREEN}${SUCCESS} $1${NC}"
}

print_error() {
    echo -e "${RED}${ERROR} $1${NC}"
}

print_info() {
    echo -e "${BLUE}${INFO} $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}${WARNING} $1${NC}"
}

print_header() {
    echo -e "${PURPLE}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "                    ${ROCKET} DBSage Release Script ${ROCKET}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo -e "${NC}"
}

# Show help
show_help() {
    echo "DBSage Release Script"
    echo ""
    echo "Usage: $0 [version|increment]"
    echo ""
    echo "Arguments:"
    echo "  version     Specific version (e.g., v1.0.0, 1.2.3)"
    echo "  patch       Auto-increment patch version (x.x.X)"
    echo "  minor       Auto-increment minor version (x.X.0)"
    echo "  major       Auto-increment major version (X.0.0)"
    echo ""
    echo "Options:"
    echo "  -h, --help  Show this help message"
    echo "  --dry-run   Show what would be done without making changes"
    echo "  --skip-git  Skip git operations (for testing)"
    echo ""
    echo "Examples:"
    echo "  $0 v1.0.0           # Release version 1.0.0"
    echo "  $0 1.2.3            # Release version 1.2.3 (v prefix added automatically)"
    echo "  $0 patch            # Increment patch: 1.0.0 -> 1.0.1"
    echo "  $0 minor            # Increment minor: 1.0.0 -> 1.1.0"
    echo "  $0 major            # Increment major: 1.0.0 -> 2.0.0"
    echo "  $0 --dry-run patch  # Show what patch increment would do"
    echo ""
}

# Get current version from git tags
get_current_version() {
    local current_version
    current_version=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
    echo "$current_version"
}

# Parse version string
parse_version() {
    local version=$1
    # Remove 'v' prefix if present
    version=${version#v}
    
    # Split version into parts
    local IFS='.'
    local parts=($version)
    
    if [ ${#parts[@]} -ne 3 ]; then
        print_error "Invalid version format: $1 (expected: x.y.z)"
        exit 1
    fi
    
    MAJOR=${parts[0]}
    MINOR=${parts[1]}
    PATCH=${parts[2]}
}

# Increment version
increment_version() {
    local increment_type=$1
    local current_version=$(get_current_version)
    
    # Print to stderr to avoid mixing with return value
    print_info "Current version: $current_version" >&2
    
    # Parse the current version
    parse_version "$current_version"
    
    case $increment_type in
        patch)
            PATCH=$((PATCH + 1))
            ;;
        minor)
            MINOR=$((MINOR + 1))
            PATCH=0
            ;;
        major)
            MAJOR=$((MAJOR + 1))
            MINOR=0
            PATCH=0
            ;;
        *)
            print_error "Invalid increment type: $increment_type"
            exit 1
            ;;
    esac
    
    local new_version="v${MAJOR}.${MINOR}.${PATCH}"
    echo "$new_version"
}

# Validate git status
check_git_status() {
    if [ "$SKIP_GIT" = "true" ]; then
        return 0
    fi
    
    print_info "Checking git status..."
    
    # Check if we're in a git repository
    if ! git rev-parse --git-dir > /dev/null 2>&1; then
        print_error "Not in a git repository"
        exit 1
    fi
    
    # Check if working directory is clean
    if ! git diff-index --quiet HEAD --; then
        print_error "Working directory is not clean. Please commit your changes first."
        git status --porcelain
        exit 1
    fi
    
    # Check if we're on main/master branch
    local current_branch=$(git branch --show-current)
    if [[ "$current_branch" != "main" && "$current_branch" != "master" ]]; then
        print_warning "You're not on main/master branch (current: $current_branch)"
        read -p "Continue anyway? [y/N]: " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    fi
    
    print_success "Git status check passed"
}

# Check dependencies
check_dependencies() {
    print_info "Checking dependencies..."
    
    local missing_deps=()
    
    # Check required tools
    for tool in go git tar zip; do
        if ! command -v "$tool" >/dev/null 2>&1; then
            missing_deps+=("$tool")
        fi
    done
    
    if [ ${#missing_deps[@]} -gt 0 ]; then
        print_error "Missing dependencies: ${missing_deps[*]}"
        exit 1
    fi
    
    print_success "All dependencies available"
}

# Build binaries for all platforms
build_binaries() {
    print_info "Building binaries for all platforms..."
    
    # Clean and create build directory
    rm -rf "$BUILD_DIR"
    mkdir -p "$BUILD_DIR"
    
    # Build matrix
    local platforms=(
        "linux:amd64"
        "linux:arm64"
        "darwin:amd64"
        "darwin:arm64"
        "windows:amd64"
        "windows:arm64"
    )
    
    for platform in "${platforms[@]}"; do
        local goos=${platform%:*}
        local goarch=${platform#*:}
        local binary_name="$BINARY_NAME"
        
        if [ "$goos" = "windows" ]; then
            binary_name="${binary_name}.exe"
        fi
        
        print_info "Building ${goos}/${goarch}..."
        
        env GOOS="$goos" GOARCH="$goarch" CGO_ENABLED=0 \
            go build -ldflags="-w -s -X main.Version=${NEW_VERSION}" \
            -o "${BUILD_DIR}/${binary_name}" \
            ./cmd/dbsage/main.go
        
        if [ $? -ne 0 ]; then
            print_error "Failed to build for ${goos}/${goarch}"
            exit 1
        fi
        
        # Move binary to platform-specific directory
        local platform_dir="${BUILD_DIR}/${goos}_${goarch}"
        mkdir -p "$platform_dir"
        mv "${BUILD_DIR}/${binary_name}" "$platform_dir/"
        
        print_success "Built ${goos}/${goarch}"
    done
    
    print_success "All binaries built successfully"
}

# Create release archives
create_archives() {
    print_info "Creating release archives..."
    
    # Clean and create archive directory
    rm -rf "$ARCHIVE_DIR"
    mkdir -p "$ARCHIVE_DIR"
    
    # Create archives for each platform
    for platform_dir in "$BUILD_DIR"/*; do
        if [ ! -d "$platform_dir" ]; then
            continue
        fi
        
        local platform=$(basename "$platform_dir")
        local archive_name
        
        print_info "Creating archive for $platform..."
        
        # Copy additional files
        cp README.md "$platform_dir/" 2>/dev/null || true
        cp LICENSE "$platform_dir/" 2>/dev/null || true
        
        # Create archive
        if [[ "$platform" == windows_* ]]; then
            archive_name="dbsage_${platform}.zip"
            (cd "$platform_dir" && zip -q "../$ARCHIVE_DIR/$archive_name" *)
        else
            archive_name="dbsage_${platform}.tar.gz"
            tar -czf "$ARCHIVE_DIR/$archive_name" -C "$platform_dir" .
        fi
        
        print_success "Created $archive_name"
    done
    
    print_success "All archives created"
}

# Generate checksums
generate_checksums() {
    print_info "Generating checksums..."
    
    cd "$ARCHIVE_DIR"
    sha256sum * > checksums.txt
    cd - > /dev/null
    
    print_success "Checksums generated"
}

# Create git tag
create_git_tag() {
    if [ "$SKIP_GIT" = "true" ]; then
        print_info "Skipping git tag creation (--skip-git)"
        return 0
    fi
    
    print_info "Creating git tag: $NEW_VERSION"
    
    if [ "$DRY_RUN" = "true" ]; then
        print_info "[DRY RUN] Would create tag: $NEW_VERSION"
        return 0
    fi
    
    # Create annotated tag
    git tag -a "$NEW_VERSION" -m "Release $NEW_VERSION"
    
    print_success "Git tag created: $NEW_VERSION"
}

# Push to remote
push_to_remote() {
    if [ "$SKIP_GIT" = "true" ]; then
        print_info "Skipping git push (--skip-git)"
        return 0
    fi
    
    print_info "Pushing to remote repository..."
    
    if [ "$DRY_RUN" = "true" ]; then
        print_info "[DRY RUN] Would push tag: $NEW_VERSION"
        return 0
    fi
    
    git push origin "$NEW_VERSION"
    
    print_success "Tag pushed to remote"
}

# Show release summary
show_summary() {
    echo ""
    print_success "🎉 Release $NEW_VERSION completed!"
    echo ""
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${YELLOW}${PACKAGE} Release Summary:${NC}"
    echo ""
    echo "• ${BLUE}Version:${NC} $NEW_VERSION"
    echo "• ${BLUE}Tag:${NC} $NEW_VERSION"
    echo "• ${BLUE}Archives:${NC} $(ls $ARCHIVE_DIR/*.tar.gz $ARCHIVE_DIR/*.zip 2>/dev/null | wc -l) files"
    echo "• ${BLUE}Checksums:${NC} $ARCHIVE_DIR/checksums.txt"
    echo ""
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${YELLOW}${INFO} Next Steps:${NC}"
    echo ""
    echo "1. ${BLUE}GitHub Release:${NC} Visit https://github.com/$REPO_NAME/releases"
    echo "2. ${BLUE}Upload Assets:${NC} Upload files from $ARCHIVE_DIR/ directory"
    echo "3. ${BLUE}Release Notes:${NC} Add release notes and changelog"
    echo ""
    echo -e "${GREEN}The GitHub Action should automatically create the release!${NC}"
    echo ""
}

# Main function
main() {
    # Parse arguments
    DRY_RUN=false
    SKIP_GIT=false
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            --dry-run)
                DRY_RUN=true
                shift
                ;;
            --skip-git)
                SKIP_GIT=true
                shift
                ;;
            *)
                VERSION_ARG="$1"
                shift
                ;;
        esac
    done
    
    # Show header
    print_header
    
    if [ "$DRY_RUN" = "true" ]; then
        print_warning "DRY RUN MODE - No changes will be made"
        echo ""
    fi
    
    # Validate input
    if [ -z "$VERSION_ARG" ]; then
        print_error "Version argument required"
        echo ""
        show_help
        exit 1
    fi
    
    # Determine new version
    case $VERSION_ARG in
        patch|minor|major)
            NEW_VERSION=$(increment_version "$VERSION_ARG")
            ;;
        v*)
            NEW_VERSION="$VERSION_ARG"
            ;;
        *)
            # Add 'v' prefix if not present
            NEW_VERSION="v$VERSION_ARG"
            ;;
    esac
    
    print_info "Target version: $NEW_VERSION"
    
    # Validate version format
    parse_version "$NEW_VERSION"
    
    # Pre-flight checks
    check_dependencies
    check_git_status
    
    # Confirm release
    if [ "$DRY_RUN" != "true" ]; then
        echo ""
        print_warning "Ready to create release $NEW_VERSION"
        read -p "Continue? [y/N]: " -n 1 -r
        echo ""
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            print_info "Release cancelled"
            exit 0
        fi
        echo ""
    fi
    
    # Build and package
    if [ "$DRY_RUN" != "true" ]; then
        build_binaries
        create_archives
        generate_checksums
    else
        print_info "[DRY RUN] Would build binaries and create archives"
    fi
    
    # Git operations
    create_git_tag
    push_to_remote
    
    # Show summary
    show_summary
}

# Run main function
main "$@"
