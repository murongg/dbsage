package main

import (
	"log"
	"os"

	"dbsage/internal/ai"
	"dbsage/internal/ui"
	"dbsage/pkg/database"
)

func main() {
	// Get environment variables
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	baseURL := os.Getenv("OPENAI_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	// Initialize connection service (will automatically load connections from config file)
	connService := database.NewConnectionService()
	defer connService.Close()

	// Get current database tools
	dbTools := connService.GetCurrentTools()

	if dbTools == nil {
		log.Println("No database connections available.")
		log.Println("Use '/add <name>' command to add database connections.")
	}

	// Initialize OpenAI client with dynamic database tools
	openaiClient := ai.NewClientWithDynamicTools(apiKey, baseURL, func() *database.DatabaseTools {
		return connService.GetCurrentTools()
	})

	// Run Bubble Tea TUI
	if err := ui.Run(openaiClient, dbTools, connService); err != nil {
		log.Fatalf("TUI error: %v", err)
	}
}
