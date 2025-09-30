package optimization

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"

	"dbsage/internal/models"
)

// PostgreSQLQueryOptimizer provides query optimization functionality for PostgreSQL
type PostgreSQLQueryOptimizer struct {
	db *sql.DB
}

// NewPostgreSQLQueryOptimizer creates a new PostgreSQL query optimizer
func NewPostgreSQLQueryOptimizer(db *sql.DB) *PostgreSQLQueryOptimizer {
	return &PostgreSQLQueryOptimizer{db: db}
}

// AnalyzeQueryPerformance analyzes the performance of a given query
func (o *PostgreSQLQueryOptimizer) AnalyzeQueryPerformance(query string) (*models.PerformanceAnalysis, error) {
	analysis := &models.PerformanceAnalysis{
		AnalysisDate:    time.Now().Format(time.RFC3339),
		Bottlenecks:     []string{},
		Recommendations: []models.QueryOptimizationSuggestion{},
	}

	// Get database size
	var dbSize string
	err := o.db.QueryRow("SELECT pg_size_pretty(pg_database_size(current_database()))").Scan(&dbSize)
	if err == nil {
		analysis.DatabaseSize = dbSize
	}

	// Get table count
	var tableCount int
	err = o.db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public'").Scan(&tableCount)
	if err == nil {
		analysis.TableCount = tableCount
	}

	// Get index count
	var indexCount int
	err = o.db.QueryRow("SELECT COUNT(*) FROM pg_indexes WHERE schemaname = 'public'").Scan(&indexCount)
	if err == nil {
		analysis.IndexCount = indexCount
	}

	// Analyze the specific query
	queryAnalysis, err := o.analyzeSpecificQuery(query)
	if err == nil {
		analysis.Recommendations = append(analysis.Recommendations, queryAnalysis...)
	}

	// Get slow query count
	slowQueryCount, err := o.getSlowQueryCount()
	if err == nil {
		analysis.SlowQueryCount = slowQueryCount
	}

	// Calculate overall score
	analysis.OverallScore = o.calculateOverallScore(analysis)

	return analysis, nil
}

// SuggestIndexes suggests indexes for a specific table
func (o *PostgreSQLQueryOptimizer) SuggestIndexes(tableName string) ([]models.IndexSuggestion, error) {
	suggestions := []models.IndexSuggestion{}

	// Check for missing indexes on foreign keys
	fkSuggestions, err := o.suggestForeignKeyIndexes(tableName)
	if err == nil {
		suggestions = append(suggestions, fkSuggestions...)
	}

	// Check for columns frequently used in WHERE clauses
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

// GetQueryPatterns analyzes query patterns from pg_stat_statements
func (o *PostgreSQLQueryOptimizer) GetQueryPatterns() ([]models.QueryPattern, error) {
	patterns := []models.QueryPattern{}

	// Check if pg_stat_statements extension is available
	var extExists bool
	err := o.db.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_extension WHERE extname = 'pg_stat_statements')").Scan(&extExists)
	if err != nil || !extExists {
		return patterns, fmt.Errorf("pg_stat_statements extension is not installed")
	}

	query := `
		SELECT 
			query,
			calls,
			total_exec_time as total_time,
			mean_exec_time as avg_time
		FROM pg_stat_statements 
		WHERE calls > 10 
		ORDER BY total_exec_time DESC 
		LIMIT 20
	`

	rows, err := o.db.Query(query)
	if err != nil {
		return patterns, err
	}
	defer rows.Close()

	for rows.Next() {
		var pattern models.QueryPattern
		err := rows.Scan(&pattern.Query, &pattern.Count, &pattern.TotalTime, &pattern.AvgTime)
		if err != nil {
			continue
		}

		// Determine pattern type
		if pattern.AvgTime > 1000 { // > 1 second
			pattern.PatternType = "slow"
		} else if pattern.Count > 100 {
			pattern.PatternType = "frequent"
		} else {
			pattern.PatternType = "complex"
		}

		// Extract tables from query
		pattern.Tables = o.extractTablesFromQuery(pattern.Query)

		// Generate suggestions for this pattern
		pattern.Suggestions = o.generatePatternSuggestions(pattern)

		patterns = append(patterns, pattern)
	}

	return patterns, nil
}

// OptimizeQuery provides optimization suggestions for a specific query
func (o *PostgreSQLQueryOptimizer) OptimizeQuery(query string) ([]models.QueryOptimizationSuggestion, error) {
	suggestions := []models.QueryOptimizationSuggestion{}

	// Analyze query structure
	structureSuggestions := o.analyzeQueryStructure(query)
	suggestions = append(suggestions, structureSuggestions...)

	// Check for common anti-patterns
	antiPatternSuggestions := o.checkAntiPatterns(query)
	suggestions = append(suggestions, antiPatternSuggestions...)

	// Analyze EXPLAIN output for the query
	explainSuggestions, err := o.analyzeExplainOutput(query)
	if err == nil {
		suggestions = append(suggestions, explainSuggestions...)
	}

	return suggestions, nil
}

// AnalyzeTablePerformance analyzes performance issues specific to a table
func (o *PostgreSQLQueryOptimizer) AnalyzeTablePerformance(tableName string) (*models.PerformanceAnalysis, error) {
	analysis := &models.PerformanceAnalysis{
		AnalysisDate:    time.Now().Format(time.RFC3339),
		Bottlenecks:     []string{},
		Recommendations: []models.QueryOptimizationSuggestion{},
	}

	// Get table-specific statistics
	tableStats, err := o.getTableSpecificStats(tableName)
	if err == nil {
		analysis.Recommendations = append(analysis.Recommendations, tableStats...)
	}

	// Suggest indexes for this table
	indexSuggestions, err := o.SuggestIndexes(tableName)
	if err == nil {
		analysis.IndexSuggestions = indexSuggestions
	}

	// Check for table-specific bottlenecks
	bottlenecks := o.identifyTableBottlenecks(tableName)
	analysis.Bottlenecks = bottlenecks

	// Calculate overall score for this table
	analysis.OverallScore = o.calculateTableScore(tableName)

	return analysis, nil
}

// Helper methods

func (o *PostgreSQLQueryOptimizer) analyzeSpecificQuery(query string) ([]models.QueryOptimizationSuggestion, error) {
	suggestions := []models.QueryOptimizationSuggestion{}

	// Run EXPLAIN ANALYZE
	explainQuery := fmt.Sprintf("EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON) %s", query)
	rows, err := o.db.Query(explainQuery)
	if err != nil {
		return suggestions, err
	}
	defer rows.Close()

	// Parse EXPLAIN output and generate suggestions
	// This is a simplified version - in a real implementation, you'd parse the JSON
	if rows.Next() {
		var explainOutput string
		if err := rows.Scan(&explainOutput); err != nil {
			return suggestions, fmt.Errorf("failed to scan EXPLAIN output: %w", err)
		}

		if strings.Contains(explainOutput, "Seq Scan") {
			suggestions = append(suggestions, models.QueryOptimizationSuggestion{
				Type:        "index",
				Priority:    "high",
				Description: "Query is using sequential scan, consider adding an index",
				Details:     "Sequential scans are inefficient for large tables",
				Impact:      "High performance improvement expected",
			})
		}

		if strings.Contains(explainOutput, "Sort") && strings.Contains(explainOutput, "external") {
			suggestions = append(suggestions, models.QueryOptimizationSuggestion{
				Type:        "structure",
				Priority:    "medium",
				Description: "Query requires external sorting, consider increasing work_mem",
				Details:     "External sorting indicates insufficient memory allocation",
				Impact:      "Medium performance improvement expected",
			})
		}
	}

	return suggestions, nil
}

func (o *PostgreSQLQueryOptimizer) suggestForeignKeyIndexes(tableName string) ([]models.IndexSuggestion, error) {
	suggestions := []models.IndexSuggestion{}

	query := `
		SELECT 
			tc.constraint_name,
			kcu.column_name
		FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kcu 
			ON tc.constraint_name = kcu.constraint_name
		WHERE tc.constraint_type = 'FOREIGN KEY' 
			AND tc.table_name = $1
			AND NOT EXISTS (
				SELECT 1 FROM pg_indexes 
				WHERE tablename = $1 
				AND indexdef LIKE '%' || kcu.column_name || '%'
			)
	`

	rows, err := o.db.Query(query, tableName)
	if err != nil {
		return suggestions, err
	}
	defer rows.Close()

	for rows.Next() {
		var constraintName, columnName string
		if err := rows.Scan(&constraintName, &columnName); err != nil {
			continue
		}

		suggestions = append(suggestions, models.IndexSuggestion{
			TableName:     tableName,
			IndexName:     fmt.Sprintf("idx_%s_%s", tableName, columnName),
			Columns:       []string{columnName},
			IndexType:     "btree",
			Reason:        "Foreign key column without index",
			Impact:        "High - will significantly improve join performance",
			CreateSQL:     fmt.Sprintf("CREATE INDEX idx_%s_%s ON %s (%s);", tableName, columnName, tableName, columnName),
			EstimatedSize: "Small to Medium",
		})
	}

	return suggestions, nil
}

func (o *PostgreSQLQueryOptimizer) suggestWhereClauseIndexes(tableName string) ([]models.IndexSuggestion, error) {
	suggestions := []models.IndexSuggestion{}

	// Get table columns that might benefit from indexes
	query := `
		SELECT 
			c.column_name,
			c.data_type,
			c.is_nullable,
			CASE WHEN i.indexname IS NOT NULL THEN 'indexed' ELSE 'not_indexed' END as index_status
		FROM information_schema.columns c
		LEFT JOIN pg_indexes i ON i.tablename = c.table_name 
			AND i.indexdef LIKE '%' || c.column_name || '%'
			AND i.schemaname = c.table_schema
		WHERE c.table_schema = 'public' 
		AND c.table_name = $1
		AND i.indexname IS NULL  -- Only non-indexed columns
		ORDER BY c.ordinal_position
	`

	rows, err := o.db.Query(query, tableName)
	if err != nil {
		return suggestions, err
	}
	defer rows.Close()

	for rows.Next() {
		var columnName, dataType, isNullable, indexStatus string
		err := rows.Scan(&columnName, &dataType, &isNullable, &indexStatus)
		if err != nil {
			continue
		}

		// Suggest indexes for columns that are commonly used in WHERE clauses
		// Focus on integer, text, varchar, and timestamp columns
		shouldIndex := false
		indexType := "btree"
		reason := ""

		switch {
		case strings.Contains(strings.ToLower(dataType), "integer") ||
			strings.Contains(strings.ToLower(dataType), "bigint") ||
			strings.Contains(strings.ToLower(dataType), "smallint"):
			shouldIndex = true
			reason = fmt.Sprintf("Integer column '%s' could benefit from a B-tree index for equality and range queries", columnName)

		case strings.Contains(strings.ToLower(dataType), "text") ||
			strings.Contains(strings.ToLower(dataType), "varchar") ||
			strings.Contains(strings.ToLower(dataType), "character"):
			shouldIndex = true
			reason = fmt.Sprintf("Text column '%s' could benefit from a B-tree index for equality queries", columnName)
			// Also suggest GIN index for full-text search if it's a text column
			if strings.Contains(strings.ToLower(dataType), "text") {
				ginSuggestion := models.IndexSuggestion{
					TableName:     tableName,
					IndexName:     fmt.Sprintf("idx_%s_%s_gin", tableName, columnName),
					Columns:       []string{columnName},
					IndexType:     "gin",
					Reason:        fmt.Sprintf("GIN index on text column '%s' for full-text search capabilities", columnName),
					Impact:        "high",
					CreateSQL:     fmt.Sprintf("CREATE INDEX idx_%s_%s_gin ON %s USING gin(to_tsvector('english', %s))", tableName, columnName, tableName, columnName),
					EstimatedSize: "Medium to Large",
				}
				suggestions = append(suggestions, ginSuggestion)
			}

		case strings.Contains(strings.ToLower(dataType), "timestamp") ||
			strings.Contains(strings.ToLower(dataType), "date") ||
			strings.Contains(strings.ToLower(dataType), "time"):
			shouldIndex = true
			reason = fmt.Sprintf("Date/time column '%s' could benefit from a B-tree index for range queries and sorting", columnName)

		case strings.Contains(strings.ToLower(dataType), "uuid"):
			shouldIndex = true
			reason = fmt.Sprintf("UUID column '%s' could benefit from a B-tree index for equality queries", columnName)

		case strings.Contains(strings.ToLower(dataType), "boolean"):
			// Boolean columns usually don't need indexes unless the distribution is very skewed
			if strings.Contains(strings.ToLower(columnName), "active") ||
				strings.Contains(strings.ToLower(columnName), "enabled") ||
				strings.Contains(strings.ToLower(columnName), "deleted") {
				shouldIndex = true
				reason = fmt.Sprintf("Boolean column '%s' might benefit from a partial index if the distribution is skewed", columnName)
			}
		}

		if shouldIndex {
			suggestion := models.IndexSuggestion{
				TableName:     tableName,
				IndexName:     fmt.Sprintf("idx_%s_%s", tableName, columnName),
				Columns:       []string{columnName},
				IndexType:     indexType,
				Reason:        reason,
				Impact:        "medium",
				CreateSQL:     fmt.Sprintf("CREATE INDEX idx_%s_%s ON %s (%s)", tableName, columnName, tableName, columnName),
				EstimatedSize: "Small to Medium",
			}
			suggestions = append(suggestions, suggestion)
		}
	}

	return suggestions, nil
}

func (o *PostgreSQLQueryOptimizer) suggestCompositeIndexes(tableName string) ([]models.IndexSuggestion, error) {
	suggestions := []models.IndexSuggestion{}

	// Get foreign key columns - these are often used together in JOINs
	fkQuery := `
		SELECT 
			kcu.column_name,
			kcu.constraint_name,
			ccu.table_name AS foreign_table_name,
			ccu.column_name AS foreign_column_name
		FROM information_schema.key_column_usage kcu
		JOIN information_schema.constraint_column_usage ccu 
			ON kcu.constraint_name = ccu.constraint_name
		JOIN information_schema.table_constraints tc 
			ON kcu.constraint_name = tc.constraint_name
		WHERE tc.constraint_type = 'FOREIGN KEY'
		AND kcu.table_schema = 'public'
		AND kcu.table_name = $1
		ORDER BY kcu.ordinal_position
	`

	fkRows, err := o.db.Query(fkQuery, tableName)
	if err == nil {
		defer fkRows.Close()

		var fkColumns []string
		for fkRows.Next() {
			var columnName, constraintName, foreignTable, foreignColumn string
			err := fkRows.Scan(&columnName, &constraintName, &foreignTable, &foreignColumn)
			if err == nil {
				fkColumns = append(fkColumns, columnName)
			}
		}

		// If we have multiple foreign keys, suggest a composite index
		if len(fkColumns) >= 2 {
			suggestion := models.IndexSuggestion{
				TableName:     tableName,
				IndexName:     fmt.Sprintf("idx_%s_fk_composite", tableName),
				Columns:       fkColumns[:2], // Take first two FK columns
				IndexType:     "btree",
				Reason:        "Composite index on foreign key columns can significantly improve JOIN performance",
				Impact:        "high",
				CreateSQL:     fmt.Sprintf("CREATE INDEX idx_%s_fk_composite ON %s (%s)", tableName, tableName, strings.Join(fkColumns[:2], ", ")),
				EstimatedSize: "Medium",
			}
			suggestions = append(suggestions, suggestion)
		}
	}

	// Look for common patterns: status + timestamp columns
	statusDateQuery := `
		SELECT 
			c1.column_name as status_col,
			c2.column_name as date_col,
			c1.data_type as status_type,
			c2.data_type as date_type
		FROM information_schema.columns c1
		JOIN information_schema.columns c2 ON c1.table_name = c2.table_name AND c1.table_schema = c2.table_schema
		WHERE c1.table_schema = 'public' 
		AND c1.table_name = $1
		AND (c1.column_name ILIKE '%status%' OR c1.column_name ILIKE '%state%' OR c1.column_name ILIKE '%type%')
		AND (c2.column_name ILIKE '%date%' OR c2.column_name ILIKE '%time%' OR c2.column_name ILIKE '%created%' OR c2.column_name ILIKE '%updated%')
		AND c1.column_name != c2.column_name
		LIMIT 1
	`

	var statusCol, dateCol, statusType, dateType string
	err = o.db.QueryRow(statusDateQuery, tableName).Scan(&statusCol, &dateCol, &statusType, &dateType)
	if err == nil {
		suggestion := models.IndexSuggestion{
			TableName:     tableName,
			IndexName:     fmt.Sprintf("idx_%s_%s_%s", tableName, statusCol, dateCol),
			Columns:       []string{statusCol, dateCol},
			IndexType:     "btree",
			Reason:        fmt.Sprintf("Composite index on '%s' and '%s' can improve queries filtering by status and date/time", statusCol, dateCol),
			Impact:        "high",
			CreateSQL:     fmt.Sprintf("CREATE INDEX idx_%s_%s_%s ON %s (%s, %s)", tableName, statusCol, dateCol, tableName, statusCol, dateCol),
			EstimatedSize: "Medium",
		}
		suggestions = append(suggestions, suggestion)
	}

	// Look for user_id + created_at pattern (very common in web applications)
	userDateQuery := `
		SELECT 
			c1.column_name as user_col,
			c2.column_name as date_col
		FROM information_schema.columns c1
		JOIN information_schema.columns c2 ON c1.table_name = c2.table_name AND c1.table_schema = c2.table_schema
		WHERE c1.table_schema = 'public' 
		AND c1.table_name = $1
		AND (c1.column_name ILIKE '%user%' OR c1.column_name ILIKE '%owner%' OR c1.column_name ILIKE '%author%' OR c1.column_name ILIKE '%creator%')
		AND (c2.column_name ILIKE '%created%' OR c2.column_name ILIKE '%date%' OR c2.column_name ILIKE '%time%')
		AND c1.column_name != c2.column_name
		LIMIT 1
	`

	var userCol, createdCol string
	err = o.db.QueryRow(userDateQuery, tableName).Scan(&userCol, &createdCol)
	if err == nil {
		suggestion := models.IndexSuggestion{
			TableName:     tableName,
			IndexName:     fmt.Sprintf("idx_%s_%s_%s", tableName, userCol, createdCol),
			Columns:       []string{userCol, createdCol},
			IndexType:     "btree",
			Reason:        fmt.Sprintf("Composite index on '%s' and '%s' can improve user-specific queries with date filtering", userCol, createdCol),
			Impact:        "high",
			CreateSQL:     fmt.Sprintf("CREATE INDEX idx_%s_%s_%s ON %s (%s, %s)", tableName, userCol, createdCol, tableName, userCol, createdCol),
			EstimatedSize: "Medium to Large",
		}
		suggestions = append(suggestions, suggestion)
	}

	// Look for JSONB columns that might benefit from GIN indexes
	jsonbQuery := `
		SELECT column_name
		FROM information_schema.columns
		WHERE table_schema = 'public'
		AND table_name = $1
		AND data_type = 'jsonb'
	`

	jsonbRows, err := o.db.Query(jsonbQuery, tableName)
	if err == nil {
		defer jsonbRows.Close()

		for jsonbRows.Next() {
			var columnName string
			err := jsonbRows.Scan(&columnName)
			if err == nil {
				suggestion := models.IndexSuggestion{
					TableName:     tableName,
					IndexName:     fmt.Sprintf("idx_%s_%s_gin", tableName, columnName),
					Columns:       []string{columnName},
					IndexType:     "gin",
					Reason:        fmt.Sprintf("GIN index on JSONB column '%s' for efficient JSON queries and operations", columnName),
					Impact:        "high",
					CreateSQL:     fmt.Sprintf("CREATE INDEX idx_%s_%s_gin ON %s USING gin(%s)", tableName, columnName, tableName, columnName),
					EstimatedSize: "Medium to Large",
				}
				suggestions = append(suggestions, suggestion)
			}
		}
	}

	// Look for array columns that might benefit from GIN indexes
	arrayQuery := `
		SELECT column_name, data_type
		FROM information_schema.columns
		WHERE table_schema = 'public'
		AND table_name = $1
		AND data_type LIKE '%[]'
	`

	arrayRows, err := o.db.Query(arrayQuery, tableName)
	if err == nil {
		defer arrayRows.Close()

		for arrayRows.Next() {
			var columnName, dataType string
			err := arrayRows.Scan(&columnName, &dataType)
			if err == nil {
				suggestion := models.IndexSuggestion{
					TableName:     tableName,
					IndexName:     fmt.Sprintf("idx_%s_%s_gin", tableName, columnName),
					Columns:       []string{columnName},
					IndexType:     "gin",
					Reason:        fmt.Sprintf("GIN index on array column '%s' (%s) for efficient array operations and containment queries", columnName, dataType),
					Impact:        "high",
					CreateSQL:     fmt.Sprintf("CREATE INDEX idx_%s_%s_gin ON %s USING gin(%s)", tableName, columnName, tableName, columnName),
					EstimatedSize: "Medium",
				}
				suggestions = append(suggestions, suggestion)
			}
		}
	}

	// Suggest partial indexes for boolean columns with likely skewed distribution
	partialIndexQuery := `
		SELECT column_name
		FROM information_schema.columns
		WHERE table_schema = 'public'
		AND table_name = $1
		AND data_type = 'boolean'
		AND (column_name ILIKE '%active%' OR column_name ILIKE '%enabled%' OR column_name ILIKE '%deleted%' OR column_name ILIKE '%published%')
	`

	partialRows, err := o.db.Query(partialIndexQuery, tableName)
	if err == nil {
		defer partialRows.Close()

		for partialRows.Next() {
			var columnName string
			err := partialRows.Scan(&columnName)
			if err == nil {
				// Suggest partial index for TRUE values (usually the minority)
				suggestion := models.IndexSuggestion{
					TableName:     tableName,
					IndexName:     fmt.Sprintf("idx_%s_%s_true", tableName, columnName),
					Columns:       []string{columnName},
					IndexType:     "btree",
					Reason:        fmt.Sprintf("Partial index on '%s' for TRUE values can be very efficient if most records are FALSE", columnName),
					Impact:        "medium",
					CreateSQL:     fmt.Sprintf("CREATE INDEX idx_%s_%s_true ON %s (%s) WHERE %s = true", tableName, columnName, tableName, columnName, columnName),
					EstimatedSize: "Small",
				}
				suggestions = append(suggestions, suggestion)
			}
		}
	}

	return suggestions, nil
}

func (o *PostgreSQLQueryOptimizer) extractTablesFromQuery(query string) []string {
	tables := []string{}

	// Simple regex to extract table names (this could be more sophisticated)
	re := regexp.MustCompile(`(?i)FROM\s+(\w+)|JOIN\s+(\w+)`)
	matches := re.FindAllStringSubmatch(query, -1)

	for _, match := range matches {
		for i := 1; i < len(match); i++ {
			if match[i] != "" {
				tables = append(tables, match[i])
			}
		}
	}

	return tables
}

func (o *PostgreSQLQueryOptimizer) generatePatternSuggestions(pattern models.QueryPattern) []models.QueryOptimizationSuggestion {
	suggestions := []models.QueryOptimizationSuggestion{}

	if pattern.PatternType == "slow" {
		suggestions = append(suggestions, models.QueryOptimizationSuggestion{
			Type:        "rewrite",
			Priority:    "high",
			Description: "This query is consistently slow and should be optimized",
			Details:     fmt.Sprintf("Average execution time: %.2fms", pattern.AvgTime),
			Impact:      "High performance improvement expected",
		})
	}

	if pattern.PatternType == "frequent" {
		suggestions = append(suggestions, models.QueryOptimizationSuggestion{
			Type:        "index",
			Priority:    "medium",
			Description: "This query runs frequently and could benefit from better indexing",
			Details:     fmt.Sprintf("Executed %d times", pattern.Count),
			Impact:      "Medium performance improvement expected",
		})
	}

	return suggestions
}

func (o *PostgreSQLQueryOptimizer) analyzeQueryStructure(query string) []models.QueryOptimizationSuggestion {
	suggestions := []models.QueryOptimizationSuggestion{}

	// Check for SELECT *
	if strings.Contains(strings.ToUpper(query), "SELECT *") {
		suggestions = append(suggestions, models.QueryOptimizationSuggestion{
			Type:        "rewrite",
			Priority:    "low",
			Description: "Avoid using SELECT *, specify only needed columns",
			Details:     "SELECT * transfers unnecessary data and reduces query cache efficiency",
			Impact:      "Low to medium performance improvement",
		})
	}

	// Check for LIKE patterns starting with %
	if strings.Contains(query, "LIKE '%") {
		suggestions = append(suggestions, models.QueryOptimizationSuggestion{
			Type:        "index",
			Priority:    "medium",
			Description: "LIKE pattern starting with % cannot use regular indexes",
			Details:     "Consider using full-text search or trigram indexes for this pattern",
			Impact:      "Medium performance improvement expected",
		})
	}

	// Check for subqueries in SELECT clause
	if strings.Contains(strings.ToUpper(query), "SELECT") && strings.Count(query, "SELECT") > 1 {
		suggestions = append(suggestions, models.QueryOptimizationSuggestion{
			Type:        "rewrite",
			Priority:    "medium",
			Description: "Consider rewriting subqueries as JOINs",
			Details:     "JOINs are often more efficient than correlated subqueries",
			Impact:      "Medium performance improvement expected",
		})
	}

	return suggestions
}

func (o *PostgreSQLQueryOptimizer) checkAntiPatterns(query string) []models.QueryOptimizationSuggestion {
	suggestions := []models.QueryOptimizationSuggestion{}

	upperQuery := strings.ToUpper(query)

	// Check for DISTINCT without necessity
	if strings.Contains(upperQuery, "DISTINCT") {
		suggestions = append(suggestions, models.QueryOptimizationSuggestion{
			Type:        "rewrite",
			Priority:    "medium",
			Description: "Verify if DISTINCT is necessary",
			Details:     "DISTINCT can be expensive; ensure it's required and consider using GROUP BY if appropriate",
			Impact:      "Medium performance improvement if unnecessary",
		})
	}

	// Check for UNION instead of UNION ALL
	if strings.Contains(upperQuery, "UNION") && !strings.Contains(upperQuery, "UNION ALL") {
		suggestions = append(suggestions, models.QueryOptimizationSuggestion{
			Type:        "rewrite",
			Priority:    "low",
			Description: "Consider using UNION ALL instead of UNION if duplicates are acceptable",
			Details:     "UNION performs duplicate elimination which is expensive",
			Impact:      "Low to medium performance improvement",
		})
	}

	// Check for functions in WHERE clause
	if regexp.MustCompile(`(?i)WHERE\s+\w+\s*\(`).MatchString(query) {
		suggestions = append(suggestions, models.QueryOptimizationSuggestion{
			Type:        "rewrite",
			Priority:    "high",
			Description: "Avoid using functions on columns in WHERE clause",
			Details:     "Functions on columns prevent index usage",
			Impact:      "High performance improvement expected",
		})
	}

	return suggestions
}

func (o *PostgreSQLQueryOptimizer) analyzeExplainOutput(query string) ([]models.QueryOptimizationSuggestion, error) {
	// This would analyze the EXPLAIN output in detail
	// For now, return empty suggestions
	return []models.QueryOptimizationSuggestion{}, nil
}

func (o *PostgreSQLQueryOptimizer) getSlowQueryCount() (int, error) {
	var count int
	err := o.db.QueryRow("SELECT COUNT(*) FROM pg_stat_statements WHERE mean_exec_time > 1000").Scan(&count)
	return count, err
}

func (o *PostgreSQLQueryOptimizer) calculateOverallScore(analysis *models.PerformanceAnalysis) int {
	score := 100

	// Deduct points for various issues
	if analysis.SlowQueryCount > 10 {
		score -= 20
	} else if analysis.SlowQueryCount > 5 {
		score -= 10
	}

	if len(analysis.IndexSuggestions) > 5 {
		score -= 15
	} else if len(analysis.IndexSuggestions) > 2 {
		score -= 8
	}

	if len(analysis.Bottlenecks) > 0 {
		score -= len(analysis.Bottlenecks) * 5
	}

	// Ensure score doesn't go below 0
	if score < 0 {
		score = 0
	}

	return score
}

func (o *PostgreSQLQueryOptimizer) getTableSpecificStats(tableName string) ([]models.QueryOptimizationSuggestion, error) {
	suggestions := []models.QueryOptimizationSuggestion{}

	// Check table statistics freshness
	var lastAnalyze sql.NullTime
	err := o.db.QueryRow("SELECT last_analyze FROM pg_stat_user_tables WHERE relname = $1", tableName).Scan(&lastAnalyze)
	if err == nil && lastAnalyze.Valid {
		if time.Since(lastAnalyze.Time) > 24*time.Hour {
			suggestions = append(suggestions, models.QueryOptimizationSuggestion{
				Type:        "structure",
				Priority:    "medium",
				Description: "Table statistics are outdated",
				Details:     fmt.Sprintf("Last analyzed: %s", lastAnalyze.Time.Format("2006-01-02 15:04:05")),
				Impact:      "Medium - outdated statistics affect query planning",
			})
		}
	}

	// Check for bloat
	// This is a simplified bloat check
	var seqScan, idxScan int64
	err = o.db.QueryRow("SELECT seq_scan, idx_scan FROM pg_stat_user_tables WHERE relname = $1", tableName).Scan(&seqScan, &idxScan)
	if err == nil && seqScan > idxScan*2 {
		suggestions = append(suggestions, models.QueryOptimizationSuggestion{
			Type:        "index",
			Priority:    "high",
			Description: "Table has high sequential scan ratio",
			Details:     fmt.Sprintf("Sequential scans: %d, Index scans: %d", seqScan, idxScan),
			Impact:      "High - consider adding appropriate indexes",
		})
	}

	return suggestions, nil
}

func (o *PostgreSQLQueryOptimizer) identifyTableBottlenecks(tableName string) []string {
	bottlenecks := []string{}

	// Check for large table without appropriate indexes
	var rowCount int64
	err := o.db.QueryRow("SELECT n_tup_ins + n_tup_upd + n_tup_del FROM pg_stat_user_tables WHERE relname = $1", tableName).Scan(&rowCount)
	if err == nil && rowCount > 100000 {
		bottlenecks = append(bottlenecks, "Large table may need partitioning")
	}

	// Check for frequent updates
	var updates int64
	err = o.db.QueryRow("SELECT n_tup_upd FROM pg_stat_user_tables WHERE relname = $1", tableName).Scan(&updates)
	if err == nil && updates > rowCount/2 {
		bottlenecks = append(bottlenecks, "High update frequency may cause bloat")
	}

	return bottlenecks
}

func (o *PostgreSQLQueryOptimizer) calculateTableScore(tableName string) int {
	score := 100

	// This would implement table-specific scoring logic
	// For now, return a default score
	return score
}
