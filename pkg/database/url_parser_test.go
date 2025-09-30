package database

import (
	"testing"

	"dbsage/pkg/dbinterfaces"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseDatabaseURL(t *testing.T) {
	tests := []struct {
		name          string
		url           string
		expected      *dbinterfaces.ConnectionConfig
		expectedError bool
	}{
		{
			name: "complete postgres URL",
			url:  "postgres://user:password@localhost:5432/testdb?sslmode=require",
			expected: &dbinterfaces.ConnectionConfig{
				Host:     "localhost",
				Port:     5432,
				Database: "testdb",
				Username: "user",
				Password: "password",
				SSLMode:  "require",
			},
			expectedError: false,
		},
		{
			name: "postgresql scheme",
			url:  "postgresql://user:password@localhost:5432/testdb",
			expected: &dbinterfaces.ConnectionConfig{
				Host:     "localhost",
				Port:     5432,
				Database: "testdb",
				Username: "user",
				Password: "password",
				SSLMode:  "disable",
			},
			expectedError: false,
		},
		{
			name: "URL without password",
			url:  "postgres://user@localhost:5432/testdb",
			expected: &dbinterfaces.ConnectionConfig{
				Host:     "localhost",
				Port:     5432,
				Database: "testdb",
				Username: "user",
				Password: "",
				SSLMode:  "disable",
			},
			expectedError: false,
		},
		{
			name: "URL without user info",
			url:  "postgres://localhost:5432/testdb",
			expected: &dbinterfaces.ConnectionConfig{
				Host:     "localhost",
				Port:     5432,
				Database: "testdb",
				Username: "",
				Password: "",
				SSLMode:  "disable",
			},
			expectedError: false,
		},
		{
			name: "URL without port (default)",
			url:  "postgres://user:password@localhost/testdb",
			expected: &dbinterfaces.ConnectionConfig{
				Host:     "localhost",
				Port:     5432,
				Database: "testdb",
				Username: "user",
				Password: "password",
				SSLMode:  "disable",
			},
			expectedError: false,
		},
		{
			name: "URL without database (default)",
			url:  "postgres://user:password@localhost:5432",
			expected: &dbinterfaces.ConnectionConfig{
				Host:     "localhost",
				Port:     5432,
				Database: "postgres",
				Username: "user",
				Password: "password",
				SSLMode:  "disable",
			},
			expectedError: false,
		},
		{
			name: "URL without host (default)",
			url:  "postgres://user:password@:5432/testdb",
			expected: &dbinterfaces.ConnectionConfig{
				Host:     "localhost",
				Port:     5432,
				Database: "testdb",
				Username: "user",
				Password: "password",
				SSLMode:  "disable",
			},
			expectedError: false,
		},
		{
			name: "URL with custom port",
			url:  "postgres://user:password@localhost:5433/testdb",
			expected: &dbinterfaces.ConnectionConfig{
				Host:     "localhost",
				Port:     5433,
				Database: "testdb",
				Username: "user",
				Password: "password",
				SSLMode:  "disable",
			},
			expectedError: false,
		},
		{
			name: "URL with SSL mode prefer",
			url:  "postgres://user:password@localhost:5432/testdb?sslmode=prefer",
			expected: &dbinterfaces.ConnectionConfig{
				Host:     "localhost",
				Port:     5432,
				Database: "testdb",
				Username: "user",
				Password: "password",
				SSLMode:  "prefer",
			},
			expectedError: false,
		},
		{
			name:          "invalid URL format",
			url:           "not-a-url",
			expected:      nil,
			expectedError: true,
		},
		{
			name:          "unsupported scheme",
			url:           "mysql://user:password@localhost:3306/testdb",
			expected:      nil,
			expectedError: true,
		},
		{
			name:          "empty URL",
			url:           "",
			expected:      nil,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseDatabaseURL(tt.url)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.expected.Host, result.Host)
				assert.Equal(t, tt.expected.Port, result.Port)
				assert.Equal(t, tt.expected.Database, result.Database)
				assert.Equal(t, tt.expected.Username, result.Username)
				assert.Equal(t, tt.expected.Password, result.Password)
				assert.Equal(t, tt.expected.SSLMode, result.SSLMode)
			}
		})
	}
}

func TestBuildDatabaseURL(t *testing.T) {
	tests := []struct {
		name     string
		config   *dbinterfaces.ConnectionConfig
		expected string
	}{
		{
			name: "complete config",
			config: &dbinterfaces.ConnectionConfig{
				Host:     "localhost",
				Port:     5432,
				Database: "testdb",
				Username: "user",
				Password: "password",
				SSLMode:  "require",
			},
			expected: "postgres://user:password@localhost:5432/testdb?sslmode=require",
		},
		{
			name: "config without password",
			config: &dbinterfaces.ConnectionConfig{
				Host:     "localhost",
				Port:     5432,
				Database: "testdb",
				Username: "user",
				Password: "",
				SSLMode:  "disable",
			},
			expected: "postgres://user@localhost:5432/testdb?sslmode=disable",
		},
		{
			name: "config with disable SSL mode",
			config: &dbinterfaces.ConnectionConfig{
				Host:     "localhost",
				Port:     5432,
				Database: "testdb",
				Username: "user",
				Password: "password",
				SSLMode:  "disable",
			},
			expected: "postgres://user:password@localhost:5432/testdb?sslmode=disable",
		},
		{
			name: "config with empty SSL mode",
			config: &dbinterfaces.ConnectionConfig{
				Host:     "localhost",
				Port:     5432,
				Database: "testdb",
				Username: "user",
				Password: "password",
				SSLMode:  "",
			},
			expected: "postgres://user:password@localhost:5432/testdb",
		},
		{
			name: "config with custom port",
			config: &dbinterfaces.ConnectionConfig{
				Host:     "example.com",
				Port:     5433,
				Database: "mydb",
				Username: "admin",
				Password: "secret",
				SSLMode:  "prefer",
			},
			expected: "postgres://admin:secret@example.com:5433/mydb?sslmode=prefer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildDatabaseURL(tt.config)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateDatabaseURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{
			name:     "valid postgres URL",
			url:      "postgres://user:password@localhost:5432/testdb",
			expected: true,
		},
		{
			name:     "valid postgresql URL",
			url:      "postgresql://user:password@localhost:5432/testdb",
			expected: true,
		},
		{
			name:     "invalid mysql URL",
			url:      "mysql://user:password@localhost:3306/testdb",
			expected: false,
		},
		{
			name:     "invalid http URL",
			url:      "http://example.com",
			expected: false,
		},
		{
			name:     "empty string",
			url:      "",
			expected: false,
		},
		{
			name:     "random string",
			url:      "not-a-url",
			expected: false,
		},
		{
			name:     "postgres prefix but invalid",
			url:      "postgres-like-string",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateDatabaseURL(tt.url)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseAndBuildRoundTrip(t *testing.T) {
	// Test that parsing a URL and then building it back gives a consistent result
	originalURL := "postgres://user:password@localhost:5432/testdb?sslmode=require"

	config, err := ParseDatabaseURL(originalURL)
	require.NoError(t, err)

	rebuiltURL := BuildDatabaseURL(config)
	assert.Equal(t, originalURL, rebuiltURL)
}

func TestParseURLWithSpecialCharacters(t *testing.T) {
	// Test URLs with special characters in password
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "password with special chars",
			url:      "postgres://user:p%40ssw%3Ard@localhost:5432/testdb",
			expected: "p@ssw:rd", // URL decoded
		},
		{
			name:     "username with special chars",
			url:      "postgres://us%40er:password@localhost:5432/testdb",
			expected: "us@er", // URL decoded
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := ParseDatabaseURL(tt.url)
			require.NoError(t, err)

			if tt.name == "password with special chars" {
				assert.Equal(t, tt.expected, config.Password)
			} else {
				assert.Equal(t, tt.expected, config.Username)
			}
		})
	}
}

// Benchmark tests
func BenchmarkParseDatabaseURL(b *testing.B) {
	url := "postgres://user:password@localhost:5432/testdb?sslmode=require"
	for i := 0; i < b.N; i++ {
		ParseDatabaseURL(url)
	}
}

func BenchmarkBuildDatabaseURL(b *testing.B) {
	config := &dbinterfaces.ConnectionConfig{
		Host:     "localhost",
		Port:     5432,
		Database: "testdb",
		Username: "user",
		Password: "password",
		SSLMode:  "require",
	}
	for i := 0; i < b.N; i++ {
		BuildDatabaseURL(config)
	}
}

func BenchmarkValidateDatabaseURL(b *testing.B) {
	url := "postgres://user:password@localhost:5432/testdb"
	for i := 0; i < b.N; i++ {
		ValidateDatabaseURL(url)
	}
}
