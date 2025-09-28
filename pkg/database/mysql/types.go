package mysql

// ConnectionConfig represents a MySQL database connection configuration
type ConnectionConfig struct {
	Name        string `json:"name"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Database    string `json:"database"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	Charset     string `json:"charset"`   // MySQL specific: utf8mb4, utf8, etc.
	Collation   string `json:"collation"` // MySQL specific collation
	Timeout     string `json:"timeout"`   // Connection timeout
	Description string `json:"description"`
	LastUsed    string `json:"last_used,omitempty"` // ISO 8601 timestamp
}
