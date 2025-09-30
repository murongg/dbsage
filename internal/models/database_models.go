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

// TableStats represents table statistics
type TableStats struct {
	TableName   string `json:"table_name"`
	RowCount    int64  `json:"row_count"`
	TableSize   string `json:"table_size"`
	IndexSize   string `json:"index_size"`
	TotalSize   string `json:"total_size"`
	LastVacuum  string `json:"last_vacuum"`
	LastAnalyze string `json:"last_analyze"`
	SeqScan     int64  `json:"seq_scan"`
	SeqTupRead  int64  `json:"seq_tup_read"`
	IdxScan     int64  `json:"idx_scan"`
	IdxTupFetch int64  `json:"idx_tup_fetch"`
}

// SlowQuery represents a slow query
type SlowQuery struct {
	Query      string  `json:"query"`
	Calls      int64   `json:"calls"`
	TotalTime  float64 `json:"total_time"`
	MeanTime   float64 `json:"mean_time"`
	MinTime    float64 `json:"min_time"`
	MaxTime    float64 `json:"max_time"`
	StddevTime float64 `json:"stddev_time"`
	Rows       int64   `json:"rows"`
}

// DatabaseSize represents database size information
type DatabaseSize struct {
	DatabaseName string `json:"database_name"`
	Size         string `json:"size"`
	SizeBytes    int64  `json:"size_bytes"`
}

// ActiveConnection represents an active database connection
type ActiveConnection struct {
	PID        int    `json:"pid"`
	Username   string `json:"username"`
	Database   string `json:"database"`
	ClientAddr string `json:"client_addr"`
	State      string `json:"state"`
	Query      string `json:"query"`
	Duration   string `json:"duration"`
}

// QueryOptimizationSuggestion represents a query optimization suggestion
type QueryOptimizationSuggestion struct {
	Type        string  `json:"type"`        // "index", "rewrite", "structure"
	Priority    string  `json:"priority"`    // "high", "medium", "low"
	Description string  `json:"description"` // Human readable description
	Details     string  `json:"details"`     // Technical details
	Impact      string  `json:"impact"`      // Expected performance impact
	SQLBefore   string  `json:"sql_before"`  // Original SQL (if applicable)
	SQLAfter    string  `json:"sql_after"`   // Suggested SQL (if applicable)
	Cost        float64 `json:"cost"`        // Estimated cost reduction
}

// IndexSuggestion represents an index suggestion
type IndexSuggestion struct {
	TableName     string   `json:"table_name"`
	IndexName     string   `json:"index_name"`
	Columns       []string `json:"columns"`
	IndexType     string   `json:"index_type"`     // "btree", "hash", "gin", "gist"
	Reason        string   `json:"reason"`         // Why this index is suggested
	Impact        string   `json:"impact"`         // Expected performance impact
	CreateSQL     string   `json:"create_sql"`     // SQL to create the index
	EstimatedSize string   `json:"estimated_size"` // Estimated index size
}

// QueryPattern represents a common query pattern
type QueryPattern struct {
	PatternType string                        `json:"pattern_type"` // "frequent", "slow", "complex"
	Query       string                        `json:"query"`
	Count       int64                         `json:"count"`
	AvgTime     float64                       `json:"avg_time"`
	TotalTime   float64                       `json:"total_time"`
	Tables      []string                      `json:"tables"`
	Suggestions []QueryOptimizationSuggestion `json:"suggestions"`
}

// PerformanceAnalysis represents overall performance analysis
type PerformanceAnalysis struct {
	AnalysisDate     string                        `json:"analysis_date"`
	DatabaseSize     string                        `json:"database_size"`
	TableCount       int                           `json:"table_count"`
	IndexCount       int                           `json:"index_count"`
	SlowQueryCount   int                           `json:"slow_query_count"`
	Bottlenecks      []string                      `json:"bottlenecks"`
	IndexSuggestions []IndexSuggestion             `json:"index_suggestions"`
	QueryPatterns    []QueryPattern                `json:"query_patterns"`
	OverallScore     int                           `json:"overall_score"` // 0-100
	Recommendations  []QueryOptimizationSuggestion `json:"recommendations"`
}
