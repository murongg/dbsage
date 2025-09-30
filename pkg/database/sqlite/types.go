package sqlite

// ConnectionConfig represents a SQLite database connection configuration
type ConnectionConfig struct {
	Name        string `json:"name"`
	Database    string `json:"database"` // Path to SQLite database file
	Mode        string `json:"mode"`     // SQLite specific: ro, rw, rwc, memory
	Cache       string `json:"cache"`    // SQLite specific: shared, private
	Timeout     string `json:"timeout"`  // Connection timeout
	Description string `json:"description"`
	LastUsed    string `json:"last_used,omitempty"` // ISO 8601 timestamp
}
