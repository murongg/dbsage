package main

import (
	"log"
	"os"

	"dbsage/internal/ai"
	"dbsage/internal/ui"
	"dbsage/pkg/database"
	"dbsage/pkg/dbinterfaces"
)

func main() {
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
