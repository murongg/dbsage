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

## Quick Start

### Requirements
- Go 1.21+
- PostgreSQL database
- OpenAI API Key

### Installation

```bash
# 1. Enter project directory
cd go/

# 2. Initialize setup
make setup

# 3. Edit configuration file
nano .env
# Set your OPENAI_API_KEY and DATABASE_URL

# 4. Run the program
make run
```

### Configuration Example

```bash
# Environment variables
export OPENAI_API_KEY="your_openai_api_key_here"
export DATABASE_URL="postgres://username:password@localhost:5432/database?sslmode=disable"

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
```

## Built-in Commands

- `help` - Show feature list
- `clear` - Clear screen and redisplay welcome message
- `exit` / `quit` - Exit program

## Security Features

- **SQL Injection Protection** - Automatically detect suspicious input
- **Dangerous Operation Alerts** - Operations like DROP, TRUNCATE require confirmation
- **Parameterized Queries** - Recommend safe query methods
- **Production Environment Protection** - Ask environment type before critical operations
