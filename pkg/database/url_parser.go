package database

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"dbsage/pkg/dbinterfaces"
)

// ParseDatabaseURL parses a database URL and returns a ConnectionConfig
// Supports formats like:
// postgres://user:password@host:port/database?sslmode=disable
// postgresql://user:password@host:port/database?sslmode=disable
func ParseDatabaseURL(databaseURL string) (*dbinterfaces.ConnectionConfig, error) {
	// Parse the URL
	u, err := url.Parse(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid database URL format: %w", err)
	}

	// Check scheme
	if u.Scheme != "postgres" && u.Scheme != "postgresql" {
		return nil, fmt.Errorf("unsupported database scheme: %s (only postgres/postgresql supported)", u.Scheme)
	}

	// Extract host and port
	host := u.Hostname()
	if host == "" {
		host = "localhost"
	}

	port := 5432 // Default PostgreSQL port
	if u.Port() != "" {
		if p, err := strconv.Atoi(u.Port()); err == nil {
			port = p
		}
	}

	// Extract database name
	database := strings.TrimPrefix(u.Path, "/")
	if database == "" {
		database = "postgres" // Default database
	}

	// Extract username and password
	username := ""
	password := ""
	if u.User != nil {
		username = u.User.Username()
		if p, ok := u.User.Password(); ok {
			password = p
		}
	}

	// Extract SSL mode
	sslMode := "disable" // Default
	if u.Query().Get("sslmode") != "" {
		sslMode = u.Query().Get("sslmode")
	}

	config := &dbinterfaces.ConnectionConfig{
		Host:     host,
		Port:     port,
		Database: database,
		Username: username,
		Password: password,
		SSLMode:  sslMode,
	}

	return config, nil
}

// BuildDatabaseURL builds a database URL from a ConnectionConfig
func BuildDatabaseURL(config *dbinterfaces.ConnectionConfig) string {
	// Build user info
	userInfo := config.Username
	if config.Password != "" {
		userInfo = fmt.Sprintf("%s:%s", config.Username, config.Password)
	}

	// Build the URL
	dbURL := fmt.Sprintf("postgres://%s@%s:%d/%s",
		userInfo, config.Host, config.Port, config.Database)

	// Add SSL mode if not default
	if config.SSLMode != "" && config.SSLMode != "disable" {
		dbURL += fmt.Sprintf("?sslmode=%s", config.SSLMode)
	} else if config.SSLMode == "disable" {
		dbURL += "?sslmode=disable"
	}

	return dbURL
}

// ValidateDatabaseURL validates if a string looks like a database URL
func ValidateDatabaseURL(input string) bool {
	return strings.HasPrefix(input, "postgres://") || strings.HasPrefix(input, "postgresql://")
}
