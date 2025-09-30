package optimization

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"

	"dbsage/internal/models"
)

// MySQLQueryOptimizer provides query optimization functionality for MySQL
type MySQLQueryOptimizer struct {
	db *sql.DB
}

// NewMySQLQueryOptimizer creates a new MySQL query optimizer
func NewMySQLQueryOptimizer(db *sql.DB) *MySQLQueryOptimizer {
	return &MySQLQueryOptimizer{db: db}
}

// AnalyzeQueryPerformance analyzes the performance of a given query
func (o *MySQLQueryOptimizer) AnalyzeQueryPerformance(query string) (*models.PerformanceAnalysis, error) {
	analysis := &models.PerformanceAnalysis{
		AnalysisDate:    time.Now().Format(time.RFC3339),
		Bottlenecks:     []string{},
		Recommendations: []models.QueryOptimizationSuggestion{},
	}

	// Get database size
	var dbSize int64
	err := o.db.QueryRow("SELECT ROUND(SUM(data_length + index_length) / 1024 / 1024, 1) AS 'DB Size in MB' FROM information_schema.tables WHERE table_schema = DATABASE()").Scan(&dbSize)
	if err == nil {
		analysis.DatabaseSize = fmt.Sprintf("%.1f MB", float64(dbSize))
	}

	// Get table count
	var tableCount int
	err = o.db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE()").Scan(&tableCount)
	if err == nil {
		analysis.TableCount = tableCount
	}

	// Get index count
	var indexCount int
	err = o.db.QueryRow("SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema = DATABASE()").Scan(&indexCount)
	if err == nil {
		analysis.IndexCount = indexCount
	}

	// Analyze the specific query
	queryAnalysis, err := o.analyzeSpecificQuery(query)
	if err == nil {
		analysis.Recommendations = append(analysis.Recommendations, queryAnalysis...)
	}

	// Get slow query count (from performance_schema if available)
	slowQueryCount, err := o.getSlowQueryCount()
	if err == nil {
		analysis.SlowQueryCount = slowQueryCount
	}

	// Calculate overall score
	analysis.OverallScore = o.calculateOverallScore(analysis)

	return analysis, nil
}

// SuggestIndexes suggests indexes for a specific table
func (o *MySQLQueryOptimizer) SuggestIndexes(tableName string) ([]models.IndexSuggestion, error) {
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

// GetQueryPatterns analyzes query patterns from performance_schema
func (o *MySQLQueryOptimizer) GetQueryPatterns() ([]models.QueryPattern, error) {
	patterns := []models.QueryPattern{}

	// Check if performance_schema is available
	var schemaExists int
	err := o.db.QueryRow("SELECT COUNT(*) FROM information_schema.schemata WHERE schema_name = 'performance_schema'").Scan(&schemaExists)
	if err != nil || schemaExists == 0 {
		return patterns, fmt.Errorf("performance_schema is not available")
	}

	query := `
		SELECT 
			DIGEST_TEXT as query,
			COUNT_STAR as calls,
			SUM_TIMER_WAIT/1000000000 as total_time,
			AVG_TIMER_WAIT/1000000000 as avg_time
		FROM performance_schema.events_statements_summary_by_digest 
		WHERE COUNT_STAR > 10 
		ORDER BY SUM_TIMER_WAIT DESC 
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
		if pattern.AvgTime > 1.0 { // > 1 second
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
func (o *MySQLQueryOptimizer) OptimizeQuery(query string) ([]models.QueryOptimizationSuggestion, error) {
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
func (o *MySQLQueryOptimizer) AnalyzeTablePerformance(tableName string) (*models.PerformanceAnalysis, error) {
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

func (o *MySQLQueryOptimizer) analyzeSpecificQuery(query string) ([]models.QueryOptimizationSuggestion, error) {
	suggestions := []models.QueryOptimizationSuggestion{}

	// Run EXPLAIN
	explainQuery := fmt.Sprintf("EXPLAIN FORMAT=JSON %s", query)
	rows, err := o.db.Query(explainQuery)
	if err != nil {
		return suggestions, err
	}
	defer rows.Close()

	// Parse EXPLAIN output and generate suggestions
	if rows.Next() {
		var explainOutput string
		rows.Scan(&explainOutput)

		if strings.Contains(explainOutput, "ALL") {
			suggestions = append(suggestions, models.QueryOptimizationSuggestion{
				Type:        "index",
				Priority:    "high",
				Description: "Query is using full table scan, consider adding an index",
				Details:     "Full table scans are inefficient for large tables",
				Impact:      "High performance improvement expected",
			})
		}

		if strings.Contains(explainOutput, "Using filesort") {
			suggestions = append(suggestions, models.QueryOptimizationSuggestion{
				Type:        "index",
				Priority:    "medium",
				Description: "Query requires filesort, consider adding an index for ORDER BY",
				Details:     "Filesort operations can be expensive for large result sets",
				Impact:      "Medium performance improvement expected",
			})
		}

		if strings.Contains(explainOutput, "Using temporary") {
			suggestions = append(suggestions, models.QueryOptimizationSuggestion{
				Type:        "structure",
				Priority:    "medium",
				Description: "Query uses temporary table, consider optimizing GROUP BY or DISTINCT",
				Details:     "Temporary tables require additional memory and I/O",
				Impact:      "Medium performance improvement expected",
			})
		}
	}

	return suggestions, nil
}

func (o *MySQLQueryOptimizer) suggestForeignKeyIndexes(tableName string) ([]models.IndexSuggestion, error) {
	suggestions := []models.IndexSuggestion{}

	query := `
		SELECT 
			COLUMN_NAME
		FROM information_schema.KEY_COLUMN_USAGE
		WHERE TABLE_SCHEMA = DATABASE()
			AND TABLE_NAME = ?
			AND REFERENCED_TABLE_NAME IS NOT NULL
			AND COLUMN_NAME NOT IN (
				SELECT COLUMN_NAME 
				FROM information_schema.STATISTICS 
				WHERE TABLE_SCHEMA = DATABASE() 
				AND TABLE_NAME = ?
			)
	`

	rows, err := o.db.Query(query, tableName, tableName)
	if err != nil {
		return suggestions, err
	}
	defer rows.Close()

	for rows.Next() {
		var columnName string
		if err := rows.Scan(&columnName); err != nil {
			continue
		}

		suggestions = append(suggestions, models.IndexSuggestion{
			TableName:     tableName,
			IndexName:     fmt.Sprintf("idx_%s_%s", tableName, columnName),
			Columns:       []string{columnName},
			IndexType:     "BTREE",
			Reason:        "Foreign key column without index",
			Impact:        "High - will significantly improve join performance",
			CreateSQL:     fmt.Sprintf("CREATE INDEX idx_%s_%s ON %s (%s);", tableName, columnName, tableName, columnName),
			EstimatedSize: "Small to Medium",
		})
	}

	return suggestions, nil
}

func (o *MySQLQueryOptimizer) suggestWhereClauseIndexes(tableName string) ([]models.IndexSuggestion, error) {
	// This is a simplified implementation
	// In a real scenario, you'd analyze actual query patterns from performance_schema
	return []models.IndexSuggestion{}, nil
}

func (o *MySQLQueryOptimizer) suggestCompositeIndexes(tableName string) ([]models.IndexSuggestion, error) {
	// This would analyze query patterns to suggest composite indexes
	return []models.IndexSuggestion{}, nil
}

func (o *MySQLQueryOptimizer) extractTablesFromQuery(query string) []string {
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

func (o *MySQLQueryOptimizer) generatePatternSuggestions(pattern models.QueryPattern) []models.QueryOptimizationSuggestion {
	suggestions := []models.QueryOptimizationSuggestion{}

	if pattern.PatternType == "slow" {
		suggestions = append(suggestions, models.QueryOptimizationSuggestion{
			Type:        "rewrite",
			Priority:    "high",
			Description: "This query is consistently slow and should be optimized",
			Details:     fmt.Sprintf("Average execution time: %.2fs", pattern.AvgTime),
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

func (o *MySQLQueryOptimizer) analyzeQueryStructure(query string) []models.QueryOptimizationSuggestion {
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
			Details:     "Consider using full-text search for this pattern",
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

func (o *MySQLQueryOptimizer) checkAntiPatterns(query string) []models.QueryOptimizationSuggestion {
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

	// Check for OR conditions that could be optimized
	if strings.Contains(upperQuery, " OR ") {
		suggestions = append(suggestions, models.QueryOptimizationSuggestion{
			Type:        "rewrite",
			Priority:    "medium",
			Description: "Consider rewriting OR conditions with UNION for better performance",
			Details:     "OR conditions can prevent efficient index usage",
			Impact:      "Medium performance improvement expected",
		})
	}

	return suggestions
}

func (o *MySQLQueryOptimizer) analyzeExplainOutput(query string) ([]models.QueryOptimizationSuggestion, error) {
	// This would analyze the EXPLAIN output in detail
	// For now, return empty suggestions
	return []models.QueryOptimizationSuggestion{}, nil
}

func (o *MySQLQueryOptimizer) getSlowQueryCount() (int, error) {
	var count int
	err := o.db.QueryRow("SELECT COUNT(*) FROM performance_schema.events_statements_summary_by_digest WHERE AVG_TIMER_WAIT/1000000000 > 1").Scan(&count)
	return count, err
}

func (o *MySQLQueryOptimizer) calculateOverallScore(analysis *models.PerformanceAnalysis) int {
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

func (o *MySQLQueryOptimizer) getTableSpecificStats(tableName string) ([]models.QueryOptimizationSuggestion, error) {
	suggestions := []models.QueryOptimizationSuggestion{}

	// Check table statistics freshness
	var updateTime sql.NullTime
	err := o.db.QueryRow("SELECT UPDATE_TIME FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?", tableName).Scan(&updateTime)
	if err == nil && updateTime.Valid {
		if time.Since(updateTime.Time) > 24*time.Hour {
			suggestions = append(suggestions, models.QueryOptimizationSuggestion{
				Type:        "structure",
				Priority:    "low",
				Description: "Table statistics may be outdated",
				Details:     fmt.Sprintf("Last updated: %s", updateTime.Time.Format("2006-01-02 15:04:05")),
				Impact:      "Low - consider running ANALYZE TABLE",
			})
		}
	}

	// Check for table fragmentation
	var dataFree int64
	err = o.db.QueryRow("SELECT DATA_FREE FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?", tableName).Scan(&dataFree)
	if err == nil && dataFree > 1024*1024*100 { // > 100MB free space
		suggestions = append(suggestions, models.QueryOptimizationSuggestion{
			Type:        "structure",
			Priority:    "medium",
			Description: "Table may be fragmented",
			Details:     fmt.Sprintf("Free space: %.2f MB", float64(dataFree)/(1024*1024)),
			Impact:      "Medium - consider running OPTIMIZE TABLE",
		})
	}

	return suggestions, nil
}

func (o *MySQLQueryOptimizer) identifyTableBottlenecks(tableName string) []string {
	bottlenecks := []string{}

	// Check for large table without appropriate indexes
	var rowCount int64
	err := o.db.QueryRow("SELECT TABLE_ROWS FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?", tableName).Scan(&rowCount)
	if err == nil && rowCount > 100000 {
		bottlenecks = append(bottlenecks, "Large table may need partitioning")
	}

	// Check for MyISAM tables (if any)
	var engine string
	err = o.db.QueryRow("SELECT ENGINE FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?", tableName).Scan(&engine)
	if err == nil && engine == "MyISAM" {
		bottlenecks = append(bottlenecks, "MyISAM engine may cause locking issues, consider InnoDB")
	}

	return bottlenecks
}

func (o *MySQLQueryOptimizer) calculateTableScore(tableName string) int {
	score := 100

	// This would implement table-specific scoring logic
	// For now, return a default score
	return score
}
