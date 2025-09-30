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

	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull, pk int
		var defaultValue sql.NullString

		err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk)
		if err != nil {
			continue
		}

		// Suggest indexes for non-primary key columns that might be used in WHERE clauses
		if pk == 0 && (strings.Contains(strings.ToLower(dataType), "int") ||
			strings.Contains(strings.ToLower(dataType), "text") ||
			strings.Contains(strings.ToLower(dataType), "varchar")) {

			suggestion := models.IndexSuggestion{
				TableName: tableName,
				IndexName: fmt.Sprintf("idx_%s_%s", tableName, name),
				Columns:   []string{name},
				IndexType: "btree",
				Reason:    fmt.Sprintf("Column %s might benefit from an index if used in WHERE clauses", name),
				Impact:    "medium",
			}
			suggestions = append(suggestions, suggestion)
		}
	}

	return suggestions, nil
}

func (o *SQLiteQueryOptimizer) suggestCompositeIndexes(tableName string) ([]models.IndexSuggestion, error) {
	suggestions := []models.IndexSuggestion{}

	// This is a simplified heuristic - in practice, you'd analyze actual query patterns
	suggestion := models.IndexSuggestion{
		TableName: tableName,
		IndexName: fmt.Sprintf("idx_%s_composite", tableName),
		Columns:   []string{"column1", "column2"}, // Placeholder
		IndexType: "btree",
		Reason:    "Consider composite indexes for columns frequently used together in WHERE clauses",
		Impact:    "high",
	}
	suggestions = append(suggestions, suggestion)

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
