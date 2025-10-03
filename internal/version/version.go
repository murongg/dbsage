package version

import "fmt"

// Build information variables that will be set by ldflags during build
var (
	Version   = "dev"
	GitCommit = "unknown"
)

// GetVersionInfo returns formatted version information
func GetVersionInfo() string {
	return fmt.Sprintf("dbsage version %s\nGit commit: %s", Version, GitCommit)
}

// GetVersionString returns just the version string
func GetVersionString() string {
	return Version
}

// GetFullVersionInfo returns detailed version information
func GetFullVersionInfo() (string, string) {
	return Version, GitCommit
}
