# Query Optimization Guide

DBSage's advanced query optimization features help you identify performance bottlenecks, suggest improvements, and automatically optimize your database queries.

## 🔍 Quick Start

### Basic Performance Analysis
```
You: Analyze the performance of this query: SELECT * FROM users WHERE email = 'john@example.com'

DBSage: I'll analyze this query's performance for you...

[Analyzes query performance]

Performance Analysis Results:
- Overall Score: 75/100
- Execution Time: 12ms
- Issues Found: 2
- Recommendations: 3

Issues:
1. Using SELECT * - consider specifying only needed columns
2. Missing index on email column for efficient lookups

Recommendations:
1. Create index on email column: CREATE INDEX idx_users_email ON users (email);
2. Rewrite query to select specific columns
3. Consider adding LIMIT if not all results are needed
```

### Index Suggestions
```
You: Suggest indexes for my products table

DBSage: I'll analyze the products table and suggest optimal indexes...

[Analyzing table structure and query patterns]

Index Suggestions for 'products':

1. **Primary Recommendation**
   CREATE INDEX idx_products_category_price ON products (category_id, price);
   Reason: Composite index for category filtering with price sorting
   Impact: High - will significantly improve category browsing queries
   Estimated Size: Medium

2. **Secondary Recommendation**
   CREATE INDEX idx_products_sku ON products (sku);
   Reason: Unique product lookups by SKU
   Impact: High - will improve product lookup performance
   Estimated Size: Small

3. **Optional Recommendation**
   CREATE INDEX idx_products_created_at ON products (created_at);
   Reason: Date range queries and reporting
   Impact: Medium - useful for admin dashboards
   Estimated Size: Small
```

## 🔧 Advanced Features

### Query Pattern Analysis
```
You: Show me query patterns

DBSage: Analyzing query patterns from database statistics...

Query Patterns Found:

1. **Frequent Pattern** (127 executions)
   SELECT * FROM orders WHERE customer_id = ?
   Avg Time: 45ms
   Suggestion: Add index on customer_id

2. **Slow Pattern** (23 executions)
   SELECT o.*, c.name FROM orders o JOIN customers c ON o.customer_id = c.id WHERE o.created_at > ?
   Avg Time: 1.2s
   Suggestions: 
   - Add composite index on (created_at, customer_id)
   - Consider pre-joining data for reports

3. **Complex Pattern** (8 executions)
   SELECT COUNT(*) FROM products WHERE category_id IN (SELECT id FROM categories WHERE parent_id = ?)
   Avg Time: 800ms
   Suggestion: Rewrite as JOIN for better performance
```

### Table Performance Analysis
```
You: Analyze users table performance

DBSage: Analyzing performance for the 'users' table...

Table Performance Analysis - 'users':

Overview:
- Table Size: 2.3 GB
- Row Count: 1,245,678
- Index Count: 4
- Performance Score: 82/100

Bottlenecks Identified:
1. High sequential scan ratio (78% vs 22% index scans)
2. Last analyzed 3 days ago - statistics may be outdated
3. Large table without partitioning

Recommendations:
1. **HIGH PRIORITY**: Add index on frequently queried columns
   CREATE INDEX idx_users_email ON users (email);
   CREATE INDEX idx_users_status_created ON users (status, created_at);

2. **MEDIUM PRIORITY**: Update table statistics
   ANALYZE users;

3. **LOW PRIORITY**: Consider partitioning by created_at for very large datasets
```

## 📊 Performance Scoring

DBSage provides a performance score (0-100) based on various factors:

- **90-100**: Excellent performance, well-optimized
- **80-89**: Good performance, minor optimizations possible
- **70-79**: Fair performance, some improvements needed
- **60-69**: Poor performance, significant optimizations required
- **0-59**: Critical performance issues, immediate attention needed

### Scoring Factors:
- Number of slow queries
- Missing indexes on foreign keys
- Sequential scan ratios
- Table fragmentation
- Query pattern efficiency

## 🛠️ Database-Specific Features

### PostgreSQL
- **pg_stat_statements** integration for query pattern analysis
- **EXPLAIN ANALYZE** with buffer analysis
- **Autovacuum** and statistics recommendations
- **Parallel query** optimization suggestions

### MySQL
- **Performance Schema** integration
- **InnoDB** optimization recommendations
- **Query cache** efficiency analysis
- **Storage engine** suggestions

## 💡 Best Practices

### Query Writing
1. **Avoid SELECT \*** - specify only needed columns
2. **Use LIMIT** when appropriate to reduce result sets
3. **Prefer JOINs** over subqueries when possible
4. **Index WHERE clause columns** frequently used in searches
5. **Use prepared statements** to improve query planning

### Index Strategy
1. **Single column indexes** for simple equality searches
2. **Composite indexes** for multi-column WHERE clauses
3. **Covering indexes** to avoid table lookups
4. **Partial indexes** for filtered queries (PostgreSQL)
5. **Regular maintenance** - analyze and reindex when needed

### Performance Monitoring
1. **Regular analysis** - run performance checks weekly
2. **Monitor slow queries** - set up alerts for queries >1s
3. **Track index usage** - remove unused indexes
4. **Analyze query patterns** - optimize frequent queries first
5. **Capacity planning** - monitor growth trends

## 🚀 Integration Examples

### Automated Optimization Workflow
```bash
# Daily performance check
You: What are today's performance bottlenecks?

# Weekly index review
You: Suggest indexes for all tables with slow queries

# Monthly comprehensive analysis
You: Analyze overall database performance and provide optimization roadmap
```

### Development Workflow
```bash
# Before deploying new queries
You: Optimize this query: [your new SQL]

# After schema changes
You: Analyze performance impact of recent schema changes

# Performance regression testing
You: Compare current performance with last week's baseline
```

## 🔧 Troubleshooting

### Common Issues

**Issue**: "pg_stat_statements extension not found"
**Solution**: Enable the extension:
```sql
CREATE EXTENSION pg_stat_statements;
```

**Issue**: "Performance schema not available"
**Solution**: Enable performance schema in MySQL:
```sql
SET GLOBAL performance_schema = ON;
```

**Issue**: "No query patterns found"
**Solution**: 
- Ensure your database has been running with queries
- Check if performance monitoring is enabled
- Run some queries and try again

### Performance Tips

1. **Run ANALYZE** regularly to keep statistics up to date
2. **Monitor index usage** - unused indexes waste space and slow writes
3. **Consider partitioning** for tables >100M rows
4. **Regular maintenance** - vacuum, reindex as needed
5. **Test optimizations** in staging before production

## 📚 Additional Resources

- [PostgreSQL Performance Tuning](https://www.postgresql.org/docs/current/performance-tips.html)
- [MySQL Performance Optimization](https://dev.mysql.com/doc/refman/8.0/en/optimization.html)
- [Database Indexing Best Practices](docs/indexing-guide.md)
- [Query Optimization Patterns](docs/query-patterns.md)

---

For more help, ask DBSage: "How can I optimize my database performance?"
