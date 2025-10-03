package models

// ConnectionConfig holds database connection configuration
type ConnectionConfig struct {
	Name     string `json:"name"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`
	SSLMode  string `json:"sslmode"`
}

// QueryResult represents the result of a SQL query
type QueryResult struct {
	Columns  []string        `json:"columns"`
	Rows     [][]interface{} `json:"rows"`
	RowCount int             `json:"row_count"`
	Duration string          `json:"duration"`
}

// TableInfo represents basic table information
type TableInfo struct {
	TableName   string `json:"table_name"`
	Schema      string `json:"schema"`
	TableType   string `json:"table_type"`
	Description string `json:"description"`
}

// ColumnInfo represents column information
type ColumnInfo struct {
	ColumnName    string  `json:"column_name"`
	DataType      string  `json:"data_type"`
	IsNullable    string  `json:"is_nullable"`
	DefaultValue  *string `json:"default_value,omitempty"`
	CharMaxLength *int    `json:"character_maximum_length,omitempty"`
	NumPrecision  *int    `json:"numeric_precision,omitempty"`
	NumScale      *int    `json:"numeric_scale,omitempty"`
	IsPrimaryKey  bool    `json:"is_primary_key"`
	IsForeignKey  bool    `json:"is_foreign_key"`
	Description   string  `json:"description"`
}

// IndexInfo represents index information
type IndexInfo struct {
	IndexName   string   `json:"index_name"`
	IsUnique    bool     `json:"is_unique"`
	IsPrimary   bool     `json:"is_primary"`
	Columns     []string `json:"columns"`
	IndexType   string   `json:"index_type"`
	TableSpace  string   `json:"tablespace"`
	Description string   `json:"description"`
}
