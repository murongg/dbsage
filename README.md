# DBSage - Database Sage

A PostgreSQL database AI management tool built with Go, focused on intelligent database operations and analysis.

> **Note**: This is currently an MVP (Minimum Viable Product) version. Features and functionality are under active development.

## Core Features

### AI Intelligence
- **Natural Language Queries** - Describe your needs in natural language, automatically generates SQL
- **Smart Tool Selection** - Automatically chooses appropriate database tools
- **Multi-step Analysis** - Supports complex recursive task processing
- **Safety Checks** - Automatically identifies dangerous operations and provides alerts

### Database Operations
- **SQL Execution** - Safely execute SQL queries
- **Table Management** - View all tables and table structures
- **Performance Analysis** - EXPLAIN query analysis and slow query detection
- **Index Management** - View and analyze table indexes
- **Data Statistics** - Get table statistics and data analysis

### Advanced Query Optimization
- **Query Performance Analysis** - Comprehensive analysis of query execution plans
- **Index Suggestions** - AI-powered recommendations for optimal indexing
- **Query Pattern Analysis** - Identify and optimize frequent query patterns
- **Query Optimization** - Automated suggestions for query improvements
- **Table Performance Analysis** - Deep dive into table-specific performance issues
- **Bottleneck Detection** - Identify and resolve database performance bottlenecks

## Quick Start

### Requirements
- Go 1.21+
- OpenAI API Key

### Installation

```bash
# 1. Enter project directory
cd go/

# 2. Initialize setup
make setup

# 3. Edit configuration file
nano .env
# Set your OPENAI_API_KEY and OPENAI_BASE_URL

# 4. Run the program
make run
```

### Configuration Example

```bash
# Environment variables
export OPENAI_API_KEY="your_openai_api_key_here"
export OPENAI_BASE_URL="your_openai_base_url_here"

# Run directly
go run cmd/dbsage/main.go
```

## Usage Examples

```
You: Help me query the first 10 records from the users table
DBSage: I'll query the first 10 records from the users table for you.

Execute SQL: SELECT * FROM users LIMIT 10;

Query results: ...

You: Analyze the performance bottlenecks of this database
DBSage: I'll help you analyze database performance, starting with slow queries...

Get slow queries list
Get all table sizes
Get active connection information

Based on the analysis results, the following performance issues were found: ...

You: Suggest indexes for my users table
DBSage: I'll analyze the users table and suggest optimal indexes...

Analyze table performance for 'users'
Suggest indexes for 'users' table

Index suggestions:
1. CREATE INDEX idx_users_email ON users (email) - for login queries
2. CREATE INDEX idx_users_created_at ON users (created_at) - for date range queries
3. CREATE INDEX idx_users_status_created ON users (status, created_at) - composite index

You: Optimize this slow query: SELECT * FROM orders WHERE created_at > '2024-01-01'
DBSage: I'll analyze and optimize this query...

Analyzing query performance...
Optimizing query...

Optimization suggestions:
1. Avoid SELECT *, specify only needed columns
2. Add index on created_at column
3. Consider partitioning if table is large

Optimized query:
SELECT id, customer_id, total, status FROM orders 
WHERE created_at > '2024-01-01' 
ORDER BY created_at DESC;
```

## Built-in Commands

### Basic Commands
- `help` - Show feature list
- `clear` - Clear screen and redisplay welcome message
- `exit` / `quit` - Exit program

### Query Optimization Commands
You can ask DBSage to help you with these optimization tasks using natural language:

- **"Analyze the performance of this query: [SQL]"** - Get detailed performance analysis
- **"Suggest indexes for [table_name]"** - Get AI-powered index recommendations
- **"Show me query patterns"** - Analyze frequent and slow query patterns
- **"Optimize this query: [SQL]"** - Get optimization suggestions
- **"Analyze [table_name] performance"** - Deep dive into table performance issues
- **"What are the database bottlenecks?"** - Identify system-wide performance issues

### Optimization Features
- **Smart Index Detection** - Automatically identifies missing indexes on foreign keys
- **Query Anti-pattern Detection** - Finds common performance anti-patterns
- **Execution Plan Analysis** - Detailed analysis of EXPLAIN output
- **Performance Scoring** - Overall database performance score (0-100)
- **Cross-Database Support** - Works with both PostgreSQL and MySQL

### Slash Commands (/ Commands)
Slash commands are used for managing database connections and application settings:

#### Database Connection Management
- `/add <connection_name>` - Add a new database connection
  - Example: `/add mydb`
  - Supports PostgreSQL and MySQL databases
- `/switch <connection_name>` - Switch to a specific database connection
  - Example: `/switch mydb`
- `/list` - Display all configured database connections and their status
- `/remove <connection_name>` - Remove a specific database connection
  - Example: `/remove mydb`

#### General Commands
- `/help` - Show help information for all available commands
- `/clear` - Clear screen and redisplay welcome message
- `/exit` or `/quit` - Exit the application

### @ Commands (Database Commands)
@ commands are used for quick database connection selection and query execution:

#### Connection Selection
- `@` - Display list of all available database connections
- `@<connection_name>` - Quickly switch to a specific database connection
  - Example: `@mydb` - Switch to connection named mydb

#### Database Queries
- `@<SQL_query>` - Execute SQL queries directly (processed through AI)
  - Example: `@show tables` - Display all tables in current database
  - Example: `@select * from users limit 10` - Query first 10 records from users table

#### Usage Examples
```
# View all connections
@

# Switch to production database
@prod_db

# Execute queries
@show tables
@SELECT COUNT(*) FROM users WHERE created_at > '2024-01-01'
```

## Security Features

- **SQL Injection Protection** - Automatically detect suspicious input
- **Dangerous Operation Alerts** - Operations like DROP, TRUNCATE require confirmation
- **Parameterized Queries** - Recommend safe query methods
- **Production Environment Protection** - Ask environment type before critical operations
