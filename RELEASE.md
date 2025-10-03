# Release Process

This document describes how to create a new release of DBSage.

## Quick Release

The easiest way to create a release is using the automated release script:

```bash
# Auto-increment patch version (1.0.0 -> 1.0.1)
./scripts/release.sh patch

# Auto-increment minor version (1.0.0 -> 1.1.0)
./scripts/release.sh minor

# Auto-increment major version (1.0.0 -> 2.0.0)
./scripts/release.sh major

# Specific version
./scripts/release.sh v1.2.3
```

## What the Release Script Does

1. **Validation**
   - Checks git status (clean working directory)
   - Validates dependencies (git)
   - Confirms target version

2. **Git Operations**
   - Creates annotated git tag
   - Pushes tag to remote repository

3. **GitHub Integration**
   - Tag push triggers GitHub Action
   - Action automatically:
     - Builds binaries for all platforms (Linux, macOS, Windows)
     - Creates archives (.tar.gz for Unix, .zip for Windows)
     - Generates SHA256 checksums
     - Creates GitHub Release
     - Uploads all build artifacts

## Manual Release Process

If you prefer manual control or need to troubleshoot:

### 1. Prepare Release

```bash
# Ensure clean working directory
git status

# Pull latest changes
git pull origin main

# Run tests
go test ./...
```

### 2. Create Git Tag

```bash
# Create and push tag
git tag -a v1.2.3 -m "Release v1.2.3"
git push origin v1.2.3
```

### 3. Monitor GitHub Action

The GitHub Action will automatically:
- Create a GitHub Release
- Upload all build artifacts
- Generate release notes

Or manually via GitHub web interface:
1. Go to https://github.com/murongg/dbsage/releases
2. Click "Create a new release"
3. Select the tag you just pushed
4. Upload artifacts from `release/` directory
5. Add release notes

## Release Script Options

```bash
# Show help
./scripts/release.sh --help

# Dry run (show what would happen)
./scripts/release.sh --dry-run patch

# Skip git operations (for testing)
./scripts/release.sh --skip-git v1.2.3
```

## Supported Platforms

The release builds for these platforms:

| OS      | Architecture | Archive Format |
|---------|-------------|----------------|
| Linux   | AMD64       | .tar.gz        |
| Linux   | ARM64       | .tar.gz        |
| macOS   | AMD64       | .tar.gz        |
| macOS   | ARM64       | .tar.gz        |
| Windows | AMD64       | .zip           |
| Windows | ARM64       | .zip           |

## Version Naming

- Use semantic versioning (MAJOR.MINOR.PATCH)
- Always prefix with 'v' (e.g., v1.2.3)
- Examples:
  - v1.0.0 - Initial release
  - v1.0.1 - Bug fixes
  - v1.1.0 - New features (backward compatible)
  - v2.0.0 - Breaking changes

## Post-Release Checklist

After creating a release:

1. ✅ Verify GitHub Release was created
2. ✅ Test installation scripts with new version
3. ✅ Update documentation if needed
4. ✅ Announce release (if applicable)

## Troubleshooting

### GitHub Action Not Triggered

If the GitHub Action doesn't run after pushing a tag:

1. Check the tag format (must start with 'v')
2. Verify the workflow file syntax
3. Check repository permissions

### Build Failures

If builds fail:

1. Ensure Go version compatibility
2. Check for platform-specific issues
3. Verify all dependencies are available

### Installation Script Issues

Test the installation scripts after each release:

```bash
# Test global installation (Linux/macOS)
curl -fsSL https://raw.githubusercontent.com/murongg/dbsage/main/install.sh | sudo bash

# Test local installation (Linux/macOS)
curl -fsSL https://raw.githubusercontent.com/murongg/dbsage/main/install.sh | bash -s -- --local

# Test Windows (in PowerShell as Administrator)
# Download and run install.bat
```

## Emergency Hotfix Process

For critical bugs requiring immediate release:

1. Create hotfix branch from the release tag
2. Apply minimal fix
3. Follow normal release process with patch increment
4. Consider backporting to main branch

## Release Automation

The entire release process is automated via GitHub Actions when you push a tag:

```bash
# This triggers the full release pipeline
git tag v1.2.3
git push origin v1.2.3
```

The action will:
- Build all platform binaries
- Create release archives
- Generate checksums
- Create GitHub Release
- Upload all artifacts

No manual intervention required! 🚀
