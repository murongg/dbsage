package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"dbsage/internal/ai"
	"dbsage/internal/ui"
	"dbsage/internal/version"
	"dbsage/pkg/database"
	"dbsage/pkg/dbinterfaces"
)

// showVersion displays version information
func showVersion() {
	version.Version = Version
	version.GitCommit = GitCommit

	fmt.Print(version.GetVersionInfo())
	fmt.Println()
}

// Build information - these will be set via ldflags
var (
	Version   = "dev"
	GitCommit = "unknown"
)

func main() {
	// Set version information globally
	version.Version = Version
	version.GitCommit = GitCommit

	// Parse command line flags
	versionFlag := flag.Bool("version", false, "Show version information")
	flag.BoolVar(versionFlag, "v", false, "Show version information (short)")
	flag.Parse()

	// Handle version flag
	if *versionFlag {
		showVersion()
		return
	}

	// Get environment variables
	apiKey := os.Getenv("OPENAI_API_KEY")
	baseURL := os.Getenv("OPENAI_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	// Initialize connection service (will automatically load connections from config file)
	connService := database.NewDefaultConnectionService()
	defer connService.Close()

	// Get current database tools
	dbTools := connService.GetCurrentTools()

	var openaiClient *ai.Client
	if apiKey != "" {
		// Initialize OpenAI client with dynamic database tools
		openaiClient = ai.NewClientWithDynamicTools(apiKey, baseURL, func() dbinterfaces.DatabaseInterface {
			return connService.GetCurrentTools()
		})
	}

	// Run Bubble Tea TUI - it will handle the case when apiKey is empty
	if err := ui.Run(openaiClient, dbTools, connService); err != nil {
		log.Fatalf("TUI error: %v", err)
	}
}
