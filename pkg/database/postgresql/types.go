package postgresql

// ConnectionConfig represents a database connection configuration
type ConnectionConfig struct {
	Name        string `json:"name"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Database    string `json:"database"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	SSLMode     string `json:"ssl_mode"`
	Description string `json:"description"`
	LastUsed    string `json:"last_used,omitempty"` // ISO 8601 timestamp
}
