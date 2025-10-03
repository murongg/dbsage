package version

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// GitHubRelease represents a GitHub release
type GitHubRelease struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Body        string    `json:"body"`
	HTMLURL     string    `json:"html_url"`
	PublishedAt time.Time `json:"published_at"`
	Draft       bool      `json:"draft"`
	Prerelease  bool      `json:"prerelease"`
}

// UpdateInfo contains information about available updates
type UpdateInfo struct {
	HasUpdate      bool
	CurrentVersion string
	LatestVersion  string
	ReleaseURL     string
	ReleaseNotes   string
}

// Checker handles version checking functionality
type Checker struct {
	githubOwner string
	githubRepo  string
	client      *http.Client
}

// NewChecker creates a new version checker
func NewChecker(owner, repo string) *Checker {
	return &Checker{
		githubOwner: owner,
		githubRepo:  repo,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// CheckForUpdates checks if there's a newer version available
func (c *Checker) CheckForUpdates(ctx context.Context) (*UpdateInfo, error) {
	currentVersion := GetVersionString()
	if currentVersion == "dev" {
		// Don't check for updates in development mode
		return &UpdateInfo{
			HasUpdate:      false,
			CurrentVersion: currentVersion,
		}, nil
	}

	latestRelease, err := c.getLatestRelease(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest release: %w", err)
	}

	if latestRelease == nil {
		return &UpdateInfo{
			HasUpdate:      false,
			CurrentVersion: currentVersion,
		}, nil
	}

	updateInfo := &UpdateInfo{
		CurrentVersion: currentVersion,
		LatestVersion:  latestRelease.TagName,
		ReleaseURL:     latestRelease.HTMLURL,
		ReleaseNotes:   latestRelease.Body,
	}

	// Compare versions
	hasUpdate, err := c.isNewerVersion(currentVersion, latestRelease.TagName)
	if err != nil {
		return updateInfo, fmt.Errorf("failed to compare versions: %w", err)
	}

	updateInfo.HasUpdate = hasUpdate
	return updateInfo, nil
}

// getLatestRelease fetches the latest release from GitHub API
func (c *Checker) getLatestRelease(ctx context.Context) (*GitHubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", c.githubOwner, c.githubRepo)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", fmt.Sprintf("dbsage/%s", GetVersionString()))

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		// No releases found
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	// Skip draft and prerelease versions
	if release.Draft || release.Prerelease {
		return nil, nil
	}

	return &release, nil
}

// isNewerVersion compares two semantic version strings
func (c *Checker) isNewerVersion(current, latest string) (bool, error) {
	// Remove 'v' prefix if present
	current = strings.TrimPrefix(current, "v")
	latest = strings.TrimPrefix(latest, "v")

	if current == latest {
		return false, nil
	}

	// Simple semantic version comparison
	currentParts, err := parseVersion(current)
	if err != nil {
		return false, fmt.Errorf("invalid current version format: %w", err)
	}

	latestParts, err := parseVersion(latest)
	if err != nil {
		return false, fmt.Errorf("invalid latest version format: %w", err)
	}

	// Compare major, minor, patch
	for i := 0; i < 3; i++ {
		if latestParts[i] > currentParts[i] {
			return true, nil
		}
		if latestParts[i] < currentParts[i] {
			return false, nil
		}
	}

	return false, nil
}

// parseVersion parses a semantic version string into [major, minor, patch]
func parseVersion(version string) ([3]int, error) {
	var parts [3]int
	versionParts := strings.Split(version, ".")

	if len(versionParts) != 3 {
		return parts, fmt.Errorf("version must have exactly 3 parts (major.minor.patch)")
	}

	for i, part := range versionParts {
		var num int
		if _, err := fmt.Sscanf(part, "%d", &num); err != nil {
			return parts, fmt.Errorf("invalid version part '%s': %w", part, err)
		}
		parts[i] = num
	}

	return parts, nil
}

// GetDefaultChecker returns a checker for the dbsage repository
func GetDefaultChecker() *Checker {
	return NewChecker("murongg", "dbsage")
}
