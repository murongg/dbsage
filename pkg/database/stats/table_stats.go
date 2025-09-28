package stats

import (
	"database/sql"
	"fmt"

	"dbsage/internal/models"
)

// TableStatsCollector collects table statistics
type TableStatsCollector struct {
	db *sql.DB
}

func NewTableStatsCollector(db *sql.DB) *TableStatsCollector {
	return &TableStatsCollector{db: db}
}

// GetTableStats returns detailed statistics for a specific table
func (c *TableStatsCollector) GetTableStats(tableName string) (*models.TableStats, error) {
	query := `
		SELECT 
			schemaname || '.' || tablename as table_name,
			n_tup_ins + n_tup_upd + n_tup_del as total_changes,
			n_tup_ins as inserts,
			n_tup_upd as updates,
			n_tup_del as deletes,
			n_live_tup as live_tuples,
			n_dead_tup as dead_tuples,
			last_vacuum,
			last_autovacuum,
			last_analyze,
			last_autoanalyze,
			vacuum_count,
			autovacuum_count,
			analyze_count,
			autoanalyze_count,
			seq_scan,
			seq_tup_read,
			idx_scan,
			idx_tup_fetch,
			n_tup_ins,
			n_tup_upd,
			n_tup_del,
			n_tup_hot_upd
		FROM pg_stat_user_tables 
		WHERE tablename = $1
	`

	var stats models.TableStats
	var lastVacuum, lastAnalyze sql.NullString
	var seqScan, seqTupRead, idxScan, idxTupFetch sql.NullInt64
	var totalChanges, inserts, updates, deletes, liveTuples, deadTuples sql.NullInt64
	var vacuumCount, autovacuumCount, analyzeCount, autoanalyzeCount sql.NullInt64
	var tupIns, tupUpd, tupDel, tupHotUpd sql.NullInt64
	var lastAutovacuum, lastAutoanalyze sql.NullString

	err := c.db.QueryRow(query, tableName).Scan(
		&stats.TableName,
		&totalChanges,
		&inserts,
		&updates,
		&deletes,
		&liveTuples,
		&deadTuples,
		&lastVacuum,
		&lastAutovacuum,
		&lastAnalyze,
		&lastAutoanalyze,
		&vacuumCount,
		&autovacuumCount,
		&analyzeCount,
		&autoanalyzeCount,
		&seqScan,
		&seqTupRead,
		&idxScan,
		&idxTupFetch,
		&tupIns,
		&tupUpd,
		&tupDel,
		&tupHotUpd,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get table stats: %w", err)
	}

	// Convert nullable values
	if lastVacuum.Valid {
		stats.LastVacuum = lastVacuum.String
	}
	if lastAnalyze.Valid {
		stats.LastAnalyze = lastAnalyze.String
	}
	if seqScan.Valid {
		stats.SeqScan = seqScan.Int64
	}
	if seqTupRead.Valid {
		stats.SeqTupRead = seqTupRead.Int64
	}
	if idxScan.Valid {
		stats.IdxScan = idxScan.Int64
	}
	if idxTupFetch.Valid {
		stats.IdxTupFetch = idxTupFetch.Int64
	}
	if liveTuples.Valid {
		stats.RowCount = liveTuples.Int64
	}

	// Get table size information
	sizeQuery := `
		SELECT 
			pg_size_pretty(pg_total_relation_size($1)) as total_size,
			pg_size_pretty(pg_relation_size($1)) as table_size,
			pg_size_pretty(pg_total_relation_size($1) - pg_relation_size($1)) as index_size
	`

	err = c.db.QueryRow(sizeQuery, tableName).Scan(
		&stats.TotalSize,
		&stats.TableSize,
		&stats.IndexSize,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get table size: %w", err)
	}

	return &stats, nil
}

// GetTableSizes returns size information for all tables
func (c *TableStatsCollector) GetTableSizes() ([]map[string]interface{}, error) {
	query := `
		SELECT 
			schemaname,
			tablename,
			pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size,
			pg_total_relation_size(schemaname||'.'||tablename) as size_bytes
		FROM pg_tables 
		WHERE schemaname NOT IN ('information_schema', 'pg_catalog')
		ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC
	`

	rows, err := c.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query table sizes: %w", err)
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var schema, table, size string
		var sizeBytes int64

		err := rows.Scan(&schema, &table, &size, &sizeBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to scan table size row: %w", err)
		}

		results = append(results, map[string]interface{}{
			"schema":     schema,
			"table_name": table,
			"size":       size,
			"size_bytes": sizeBytes,
		})
	}

	return results, nil
}

// GetSlowQueries returns slow query information
func (c *TableStatsCollector) GetSlowQueries() ([]models.SlowQuery, error) {
	query := `
		SELECT 
			query,
			calls,
			total_exec_time,
			mean_exec_time,
			min_exec_time,
			max_exec_time,
			stddev_exec_time,
			rows
		FROM pg_stat_statements 
		ORDER BY total_exec_time DESC 
		LIMIT 20
	`

	rows, err := c.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query slow queries: %w", err)
	}
	defer rows.Close()

	var results []models.SlowQuery
	for rows.Next() {
		var query models.SlowQuery

		err := rows.Scan(
			&query.Query,
			&query.Calls,
			&query.TotalTime,
			&query.MeanTime,
			&query.MinTime,
			&query.MaxTime,
			&query.StddevTime,
			&query.Rows,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan slow query row: %w", err)
		}

		results = append(results, query)
	}

	return results, nil
}
