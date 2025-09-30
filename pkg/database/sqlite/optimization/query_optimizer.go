package optimization

import (
	"database/sql"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"dbsage/internal/models"
)

// SQLiteQueryOptimizer provides query optimization functionality for SQLite
type SQLiteQueryOptimizer struct {
	db *sql.DB
}

// NewSQLiteQueryOptimizer creates a new SQLite query optimizer
func NewSQLiteQueryOptimizer(db *sql.DB) *SQLiteQueryOptimizer {
	return &SQLiteQueryOptimizer{db: db}
}

// AnalyzeQueryPerformance analyzes the performance of a given query
func (o *SQLiteQueryOptimizer) AnalyzeQueryPerformance(query string) (*models.PerformanceAnalysis, error) {
	analysis := &models.PerformanceAnalysis{
		AnalysisDate:    time.Now().Format(time.RFC3339),
		Bottlenecks:     []string{},
		Recommendations: []models.QueryOptimizationSuggestion{},
	}

	// Get database size
	var dbPath string
	err := o.db.QueryRow("PRAGMA database_list").Scan(nil, nil, &dbPath)
	if err == nil && dbPath != ":memory:" {
		if fileInfo, err := os.Stat(dbPath); err == nil {
			sizeMB := float64(fileInfo.Size()) / (1024 * 1024)
			analysis.DatabaseSize = fmt.Sprintf("%.1f MB", sizeMB)
		}
	} else {
		analysis.DatabaseSize = "In-Memory"
	}

	// Get table count
	var tableCount int
	err = o.db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name NOT LIKE 'sqlite_%'").Scan(&tableCount)
	if err == nil {
		analysis.TableCount = tableCount
	}

	// Get index count
	var indexCount int
	err = o.db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type = 'index' AND name NOT LIKE 'sqlite_%'").Scan(&indexCount)
	if err == nil {
		analysis.IndexCount = indexCount
	}

	// Analyze the specific query
	queryAnalysis, err := o.analyzeSpecificQuery(query)
	if err == nil {
		analysis.Recommendations = append(analysis.Recommendations, queryAnalysis...)
	}

	// SQLite doesn't have a built-in slow query log
	analysis.SlowQueryCount = 0

	// Calculate overall score
	analysis.OverallScore = o.calculateOverallScore(analysis)

	return analysis, nil
}

// SuggestIndexes suggests indexes for a specific table
func (o *SQLiteQueryOptimizer) SuggestIndexes(tableName string) ([]models.IndexSuggestion, error) {
	suggestions := []models.IndexSuggestion{}

	// Check for missing indexes on foreign keys
	fkSuggestions, err := o.suggestForeignKeyIndexes(tableName)
	if err == nil {
		suggestions = append(suggestions, fkSuggestions...)
	}

	// Check for columns frequently used in WHERE clauses (basic heuristics)
	whereSuggestions, err := o.suggestWhereClauseIndexes(tableName)
	if err == nil {
		suggestions = append(suggestions, whereSuggestions...)
	}

	// Check for composite index opportunities
	compositeSuggestions, err := o.suggestCompositeIndexes(tableName)
	if err == nil {
		suggestions = append(suggestions, compositeSuggestions...)
	}

	return suggestions, nil
}

// GetQueryPatterns analyzes query patterns (limited in SQLite)
func (o *SQLiteQueryOptimizer) GetQueryPatterns() ([]models.QueryPattern, error) {
	// SQLite doesn't have built-in query pattern analysis like MySQL/PostgreSQL
	// We'll return basic patterns based on table structure
	patterns := []models.QueryPattern{}

	// Get all tables
	query := "SELECT name FROM sqlite_master WHERE type = 'table' AND name NOT LIKE 'sqlite_%'"
	rows, err := o.db.Query(query)
	if err != nil {
		return patterns, err
	}
	defer rows.Close()

	for rows.Next() {
		var tableName string
		err := rows.Scan(&tableName)
		if err != nil {
			continue
		}

		// Create a basic pattern for each table
		pattern := models.QueryPattern{
			Query:       fmt.Sprintf("SELECT * FROM %s", tableName),
			Count:       1,   // Placeholder
			TotalTime:   0.1, // Placeholder
			AvgTime:     0.1, // Placeholder
			PatternType: "basic",
			Tables:      []string{tableName},
			Suggestions: []models.QueryOptimizationSuggestion{
				{
					Type:        "index",
					Description: fmt.Sprintf("Consider adding indexes to frequently queried columns in %s", tableName),
					Impact:      "medium",
				},
			},
		}

		patterns = append(patterns, pattern)
	}

	return patterns, nil
}

// OptimizeQuery provides optimization suggestions for a specific query
func (o *SQLiteQueryOptimizer) OptimizeQuery(query string) ([]models.QueryOptimizationSuggestion, error) {
	suggestions := []models.QueryOptimizationSuggestion{}

	// Analyze query structure
	queryLower := strings.ToLower(strings.TrimSpace(query))

	// Check for SELECT *
	if strings.Contains(queryLower, "select *") {
		suggestions = append(suggestions, models.QueryOptimizationSuggestion{
			Type:        "query_structure",
			Description: "Avoid SELECT * - specify only needed columns to reduce I/O",
			Impact:      "medium",
		})
	}

	// Check for missing WHERE clause in SELECT
	if strings.HasPrefix(queryLower, "select") && !strings.Contains(queryLower, "where") && !strings.Contains(queryLower, "limit") {
		suggestions = append(suggestions, models.QueryOptimizationSuggestion{
			Type:        "query_structure",
			Description: "Consider adding WHERE clause or LIMIT to avoid full table scan",
			Impact:      "high",
		})
	}

	// Check for LIKE with leading wildcard
	if strings.Contains(queryLower, "like '%") {
		suggestions = append(suggestions, models.QueryOptimizationSuggestion{
			Type:        "query_structure",
			Description: "LIKE with leading wildcard prevents index usage - consider full-text search",
			Impact:      "high",
			Priority:    "medium",
		})
	}

	// Check for ORDER BY without LIMIT
	if strings.Contains(queryLower, "order by") && !strings.Contains(queryLower, "limit") {
		suggestions = append(suggestions, models.QueryOptimizationSuggestion{
			Type:        "query_structure",
			Description: "ORDER BY without LIMIT may sort unnecessary rows - consider adding LIMIT",
			Impact:      "medium",
		})
	}

	// Check for subqueries that could be JOINs
	if strings.Contains(queryLower, "select") && strings.Count(queryLower, "select") > 1 {
		suggestions = append(suggestions, models.QueryOptimizationSuggestion{
			Type:        "query_structure",
			Description: "Consider converting subqueries to JOINs for better performance",
			Impact:      "medium",
			Priority:    "medium",
		})
	}

	// Extract tables and suggest indexes
	tables := o.extractTablesFromQuery(query)
	for _, table := range tables {
		suggestions = append(suggestions, models.QueryOptimizationSuggestion{
			Type:        "index",
			Description: fmt.Sprintf("Ensure appropriate indexes exist on %s for columns used in WHERE, JOIN, and ORDER BY clauses", table),
			Impact:      "high",
		})
	}

	return suggestions, nil
}

// AnalyzeTablePerformance analyzes performance issues specific to a table
func (o *SQLiteQueryOptimizer) AnalyzeTablePerformance(tableName string) (*models.PerformanceAnalysis, error) {
	analysis := &models.PerformanceAnalysis{
		AnalysisDate:    time.Now().Format(time.RFC3339),
		Bottlenecks:     []string{},
		Recommendations: []models.QueryOptimizationSuggestion{},
	}

	// Get table row count
	var rowCount int64
	err := o.db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)).Scan(&rowCount)
	if err == nil {
		if rowCount > 100000 {
			analysis.Bottlenecks = append(analysis.Bottlenecks, fmt.Sprintf("Large table (%d rows) - consider partitioning or archiving", rowCount))
		}
	}

	// Check for indexes on this table
	indexQuery := fmt.Sprintf("PRAGMA index_list(%s)", tableName)
	rows, err := o.db.Query(indexQuery)
	if err == nil {
		defer rows.Close()
		indexCount := 0
		for rows.Next() {
			indexCount++
		}

		if indexCount == 0 {
			analysis.Recommendations = append(analysis.Recommendations, models.QueryOptimizationSuggestion{
				Type:        "index",
				Description: fmt.Sprintf("Table %s has no indexes - consider adding indexes on frequently queried columns", tableName),
				Impact:      "high",
			})
		}
	}

	// Check for foreign keys without indexes
	fkQuery := fmt.Sprintf("PRAGMA foreign_key_list(%s)", tableName)
	fkRows, err := o.db.Query(fkQuery)
	if err == nil {
		defer fkRows.Close()
		for fkRows.Next() {
			analysis.Recommendations = append(analysis.Recommendations, models.QueryOptimizationSuggestion{
				Type:        "index",
				Description: fmt.Sprintf("Consider adding indexes on foreign key columns in %s", tableName),
				Impact:      "medium",
			})
		}
	}

	analysis.OverallScore = o.calculateOverallScore(analysis)

	return analysis, nil
}

// Helper methods

func (o *SQLiteQueryOptimizer) analyzeSpecificQuery(query string) ([]models.QueryOptimizationSuggestion, error) {
	suggestions := []models.QueryOptimizationSuggestion{}

	// Run EXPLAIN QUERY PLAN
	explainQuery := fmt.Sprintf("EXPLAIN QUERY PLAN %s", query)
	rows, err := o.db.Query(explainQuery)
	if err != nil {
		return suggestions, err
	}
	defer rows.Close()

	hasFullTableScan := false
	for rows.Next() {
		var id, parent, notused int
		var detail string
		err := rows.Scan(&id, &parent, &notused, &detail)
		if err != nil {
			continue
		}

		if strings.Contains(strings.ToLower(detail), "scan table") {
			hasFullTableScan = true
		}
	}

	if hasFullTableScan {
		suggestions = append(suggestions, models.QueryOptimizationSuggestion{
			Type:        "index",
			Description: "Query performs full table scan - consider adding appropriate indexes",
			Impact:      "high",
		})
	}

	return suggestions, nil
}

func (o *SQLiteQueryOptimizer) suggestForeignKeyIndexes(tableName string) ([]models.IndexSuggestion, error) {
	suggestions := []models.IndexSuggestion{}

	query := fmt.Sprintf("PRAGMA foreign_key_list(%s)", tableName)
	rows, err := o.db.Query(query)
	if err != nil {
		return suggestions, err
	}
	defer rows.Close()

	for rows.Next() {
		var id, seq int
		var table, from, to, onUpdate, onDelete, match string
		err := rows.Scan(&id, &seq, &table, &from, &to, &onUpdate, &onDelete, &match)
		if err != nil {
			continue
		}

		suggestion := models.IndexSuggestion{
			TableName: tableName,
			IndexName: fmt.Sprintf("idx_%s_%s", tableName, from),
			Columns:   []string{from},
			IndexType: "btree",
			Reason:    "Foreign key column should have an index for better JOIN performance",
			Impact:    "medium",
		}
		suggestions = append(suggestions, suggestion)
	}

	return suggestions, nil
}

func (o *SQLiteQueryOptimizer) suggestWhereClauseIndexes(tableName string) ([]models.IndexSuggestion, error) {
	suggestions := []models.IndexSuggestion{}

	// Get table schema
	query := fmt.Sprintf("PRAGMA table_info(%s)", tableName)
	rows, err := o.db.Query(query)
	if err != nil {
		return suggestions, err
	}
	defer rows.Close()

	// Check existing indexes to avoid duplicates
	existingIndexes := make(map[string]bool)
	indexQuery := fmt.Sprintf("PRAGMA index_list(%s)", tableName)
	indexRows, err := o.db.Query(indexQuery)
	if err == nil {
		defer indexRows.Close()
		for indexRows.Next() {
			var seq int
			var name, unique, origin string
			var partial int
			err := indexRows.Scan(&seq, &name, &unique, &origin, &partial)
			if err == nil {
				existingIndexes[name] = true
			}
		}
	}

	for rows.Next() {
		var cid int
		var columnName, dataType string
		var notNull, pk int
		var defaultValue sql.NullString

		err := rows.Scan(&cid, &columnName, &dataType, &notNull, &defaultValue, &pk)
		if err != nil {
			continue
		}

		// Skip primary key columns (they already have indexes)
		if pk == 1 {
			continue
		}

		// Check if column already has an index
		indexName := fmt.Sprintf("idx_%s_%s", tableName, columnName)
		if existingIndexes[indexName] {
			continue
		}

		// Suggest indexes based on data type and common usage patterns
		shouldIndex := false
		reason := ""
		impact := "medium"

		dataTypeLower := strings.ToLower(dataType)
		columnNameLower := strings.ToLower(columnName)

		switch {
		case strings.Contains(dataTypeLower, "integer") || strings.Contains(dataTypeLower, "int"):
			shouldIndex = true
			reason = fmt.Sprintf("Integer column '%s' could benefit from an index for equality and range queries", columnName)
			impact = "medium"

		case strings.Contains(dataTypeLower, "text"):
			shouldIndex = true
			reason = fmt.Sprintf("Text column '%s' could benefit from an index for equality queries and sorting", columnName)
			impact = "medium"

			// For text columns that might contain large content, suggest they might not need indexing
			if strings.Contains(columnNameLower, "content") ||
				strings.Contains(columnNameLower, "description") ||
				strings.Contains(columnNameLower, "body") {
				reason = fmt.Sprintf("Text column '%s' might not need an index if it contains large text content", columnName)
				impact = "low"
			}

		case strings.Contains(dataTypeLower, "varchar") || strings.Contains(dataTypeLower, "char"):
			shouldIndex = true
			reason = fmt.Sprintf("String column '%s' could benefit from an index for equality queries", columnName)
			impact = "medium"

		case strings.Contains(dataTypeLower, "real") || strings.Contains(dataTypeLower, "numeric") || strings.Contains(dataTypeLower, "decimal"):
			shouldIndex = true
			reason = fmt.Sprintf("Numeric column '%s' could benefit from an index for range queries and sorting", columnName)
			impact = "medium"

		case strings.Contains(dataTypeLower, "date") || strings.Contains(dataTypeLower, "time"):
			shouldIndex = true
			reason = fmt.Sprintf("Date/time column '%s' could benefit from an index for range queries and sorting", columnName)
			impact = "high"

		case strings.Contains(dataTypeLower, "blob"):
			// Generally don't index BLOB columns
			shouldIndex = false
		}

		// Special cases based on column names (common patterns)
		if shouldIndex {
			if strings.Contains(columnNameLower, "email") {
				reason = fmt.Sprintf("Email column '%s' should have an index for user lookups", columnName)
				impact = "high"
			} else if strings.Contains(columnNameLower, "status") || strings.Contains(columnNameLower, "state") {
				reason = fmt.Sprintf("Status column '%s' should have an index for filtering queries", columnName)
				impact = "high"
			} else if strings.Contains(columnNameLower, "created") || strings.Contains(columnNameLower, "updated") {
				reason = fmt.Sprintf("Timestamp column '%s' should have an index for date range queries", columnName)
				impact = "high"
			} else if strings.Contains(columnNameLower, "user") && strings.Contains(columnNameLower, "id") {
				reason = fmt.Sprintf("User ID column '%s' should have an index for user-specific queries", columnName)
				impact = "high"
			}
		}

		if shouldIndex {
			suggestion := models.IndexSuggestion{
				TableName:     tableName,
				IndexName:     indexName,
				Columns:       []string{columnName},
				IndexType:     "btree",
				Reason:        reason,
				Impact:        impact,
				CreateSQL:     fmt.Sprintf("CREATE INDEX %s ON %s (%s)", indexName, tableName, columnName),
				EstimatedSize: "Small to Medium",
			}
			suggestions = append(suggestions, suggestion)
		}
	}

	return suggestions, nil
}

func (o *SQLiteQueryOptimizer) suggestCompositeIndexes(tableName string) ([]models.IndexSuggestion, error) {
	suggestions := []models.IndexSuggestion{}

	// Get table schema to analyze column patterns
	query := fmt.Sprintf("PRAGMA table_info(%s)", tableName)
	rows, err := o.db.Query(query)
	if err != nil {
		return suggestions, err
	}
	defer rows.Close()

	var columns []struct {
		name     string
		dataType string
		pk       int
	}

	for rows.Next() {
		var cid int
		var columnName, dataType string
		var notNull, pk int
		var defaultValue sql.NullString

		err := rows.Scan(&cid, &columnName, &dataType, &notNull, &defaultValue, &pk)
		if err != nil {
			continue
		}

		columns = append(columns, struct {
			name     string
			dataType string
			pk       int
		}{columnName, dataType, pk})
	}

	// Check existing indexes to avoid duplicates
	existingIndexes := make(map[string]bool)
	indexQuery := fmt.Sprintf("PRAGMA index_list(%s)", tableName)
	indexRows, err := o.db.Query(indexQuery)
	if err == nil {
		defer indexRows.Close()
		for indexRows.Next() {
			var seq int
			var name, unique, origin string
			var partial int
			err := indexRows.Scan(&seq, &name, &unique, &origin, &partial)
			if err == nil {
				existingIndexes[name] = true
			}
		}
	}

	// Analyze foreign key relationships
	fkQuery := fmt.Sprintf("PRAGMA foreign_key_list(%s)", tableName)
	fkRows, err := o.db.Query(fkQuery)
	if err == nil {
		defer fkRows.Close()

		var fkColumns []string
		for fkRows.Next() {
			var id, seq int
			var table, from, to, onUpdate, onDelete, match string
			err := fkRows.Scan(&id, &seq, &table, &from, &to, &onUpdate, &onDelete, &match)
			if err == nil {
				fkColumns = append(fkColumns, from)
			}
		}

		// If we have multiple foreign keys, suggest a composite index
		if len(fkColumns) >= 2 {
			indexName := fmt.Sprintf("idx_%s_fk_composite", tableName)
			if !existingIndexes[indexName] {
				suggestion := models.IndexSuggestion{
					TableName:     tableName,
					IndexName:     indexName,
					Columns:       fkColumns[:2], // Take first two FK columns
					IndexType:     "btree",
					Reason:        "Composite index on foreign key columns can improve JOIN performance",
					Impact:        "high",
					CreateSQL:     fmt.Sprintf("CREATE INDEX %s ON %s (%s, %s)", indexName, tableName, fkColumns[0], fkColumns[1]),
					EstimatedSize: "Medium",
				}
				suggestions = append(suggestions, suggestion)
			}
		}
	}

	// Look for common column name patterns that often work well together
	var statusCol, dateCol, userCol, createdCol string

	for _, col := range columns {
		if col.pk == 1 {
			continue // Skip primary keys
		}

		colNameLower := strings.ToLower(col.name)

		// Look for status/state columns
		if statusCol == "" && (strings.Contains(colNameLower, "status") ||
			strings.Contains(colNameLower, "state") ||
			strings.Contains(colNameLower, "type")) {
			statusCol = col.name
		}

		// Look for date/time columns
		if dateCol == "" && (strings.Contains(colNameLower, "date") ||
			strings.Contains(colNameLower, "time") ||
			strings.Contains(colNameLower, "created") ||
			strings.Contains(colNameLower, "updated")) {
			dateCol = col.name
		}

		// Look for user-related columns
		if userCol == "" && (strings.Contains(colNameLower, "user") ||
			strings.Contains(colNameLower, "owner") ||
			strings.Contains(colNameLower, "author")) {
			userCol = col.name
		}

		// Look specifically for created_at/created_date columns
		if createdCol == "" && (strings.Contains(colNameLower, "created")) {
			createdCol = col.name
		}
	}

	// Suggest status + date composite index
	if statusCol != "" && dateCol != "" {
		indexName := fmt.Sprintf("idx_%s_%s_%s", tableName, statusCol, dateCol)
		if !existingIndexes[indexName] {
			suggestion := models.IndexSuggestion{
				TableName:     tableName,
				IndexName:     indexName,
				Columns:       []string{statusCol, dateCol},
				IndexType:     "btree",
				Reason:        fmt.Sprintf("Composite index on '%s' and '%s' can improve queries filtering by status and date", statusCol, dateCol),
				Impact:        "high",
				CreateSQL:     fmt.Sprintf("CREATE INDEX %s ON %s (%s, %s)", indexName, tableName, statusCol, dateCol),
				EstimatedSize: "Medium",
			}
			suggestions = append(suggestions, suggestion)
		}
	}

	// Suggest user + created composite index
	if userCol != "" && createdCol != "" && userCol != createdCol {
		indexName := fmt.Sprintf("idx_%s_%s_%s", tableName, userCol, createdCol)
		if !existingIndexes[indexName] {
			suggestion := models.IndexSuggestion{
				TableName:     tableName,
				IndexName:     indexName,
				Columns:       []string{userCol, createdCol},
				IndexType:     "btree",
				Reason:        fmt.Sprintf("Composite index on '%s' and '%s' can improve user-specific queries with date filtering", userCol, createdCol),
				Impact:        "high",
				CreateSQL:     fmt.Sprintf("CREATE INDEX %s ON %s (%s, %s)", indexName, tableName, userCol, createdCol),
				EstimatedSize: "Medium to Large",
			}
			suggestions = append(suggestions, suggestion)
		}
	}

	// Look for integer columns that might work well together (like category_id + priority)
	var intColumns []string
	for _, col := range columns {
		if col.pk == 1 {
			continue
		}
		if strings.Contains(strings.ToLower(col.dataType), "int") {
			intColumns = append(intColumns, col.name)
		}
	}

	// If we have multiple integer columns, suggest a composite index for the first two
	if len(intColumns) >= 2 {
		indexName := fmt.Sprintf("idx_%s_%s_%s", tableName, intColumns[0], intColumns[1])
		if !existingIndexes[indexName] {
			suggestion := models.IndexSuggestion{
				TableName:     tableName,
				IndexName:     indexName,
				Columns:       intColumns[:2],
				IndexType:     "btree",
				Reason:        fmt.Sprintf("Composite index on integer columns '%s' and '%s' might improve multi-column filtering", intColumns[0], intColumns[1]),
				Impact:        "medium",
				CreateSQL:     fmt.Sprintf("CREATE INDEX %s ON %s (%s, %s)", indexName, tableName, intColumns[0], intColumns[1]),
				EstimatedSize: "Small to Medium",
			}
			suggestions = append(suggestions, suggestion)
		}
	}

	// SQLite-specific: Suggest covering indexes for common SELECT patterns
	// This is a heuristic based on common application patterns
	if len(columns) >= 3 {
		var smallColumns []string
		for _, col := range columns {
			if col.pk == 1 {
				continue
			}
			// Include small columns that are often selected together
			colNameLower := strings.ToLower(col.name)
			dataTypeLower := strings.ToLower(col.dataType)

			if (strings.Contains(dataTypeLower, "int") ||
				strings.Contains(dataTypeLower, "text") ||
				strings.Contains(dataTypeLower, "varchar")) &&
				(strings.Contains(colNameLower, "name") ||
					strings.Contains(colNameLower, "title") ||
					strings.Contains(colNameLower, "status") ||
					strings.Contains(colNameLower, "type")) {
				smallColumns = append(smallColumns, col.name)
			}
		}

		if len(smallColumns) >= 2 {
			indexName := fmt.Sprintf("idx_%s_covering", tableName)
			if !existingIndexes[indexName] {
				suggestion := models.IndexSuggestion{
					TableName:     tableName,
					IndexName:     indexName,
					Columns:       smallColumns[:2],
					IndexType:     "btree",
					Reason:        fmt.Sprintf("Covering index on '%s' and '%s' can improve SELECT performance by avoiding table lookups", smallColumns[0], smallColumns[1]),
					Impact:        "medium",
					CreateSQL:     fmt.Sprintf("CREATE INDEX %s ON %s (%s, %s)", indexName, tableName, smallColumns[0], smallColumns[1]),
					EstimatedSize: "Medium",
				}
				suggestions = append(suggestions, suggestion)
			}
		}
	}

	return suggestions, nil
}

func (o *SQLiteQueryOptimizer) extractTablesFromQuery(query string) []string {
	tables := []string{}

	// Simple regex to extract table names - this is a basic implementation
	re := regexp.MustCompile(`(?i)(?:from|join|update|into)\s+([a-zA-Z_][a-zA-Z0-9_]*)`)
	matches := re.FindAllStringSubmatch(query, -1)

	for _, match := range matches {
		if len(match) > 1 {
			tableName := match[1]
			// Avoid duplicates
			found := false
			for _, existing := range tables {
				if existing == tableName {
					found = true
					break
				}
			}
			if !found {
				tables = append(tables, tableName)
			}
		}
	}

	return tables
}

func (o *SQLiteQueryOptimizer) calculateOverallScore(analysis *models.PerformanceAnalysis) int {
	score := 100

	// Deduct points for bottlenecks
	score -= len(analysis.Bottlenecks) * 10

	// Deduct points for high-impact recommendations
	for _, rec := range analysis.Recommendations {
		if rec.Impact == "high" {
			score -= 15
		} else if rec.Impact == "medium" {
			score -= 10
		} else {
			score -= 5
		}
	}

	if score < 0 {
		score = 0
	}

	return score
}
