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
		rows.Scan(&explainOutput)

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
	// This is a simplified implementation
	// In a real scenario, you'd analyze actual query patterns from pg_stat_statements
	return []models.IndexSuggestion{}, nil
}

func (o *PostgreSQLQueryOptimizer) suggestCompositeIndexes(tableName string) ([]models.IndexSuggestion, error) {
	// This would analyze query patterns to suggest composite indexes
	return []models.IndexSuggestion{}, nil
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
