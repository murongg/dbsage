# DBSage - Database AI Assistant

An intelligent database management tool that brings AI-powered assistance to your database operations.

## Features

- **🧠 AI-Powered**: Convert natural language queries into optimized SQL
- **🛡️ Safety First**: Built-in protection against dangerous operations 
- **🔌 Multi-Database**: Support for PostgreSQL, MySQL, and SQLite
- **💻 Cross-Platform**: Works on Linux, macOS, and Windows

## Installation

### Quick Install
```bash
# Linux / macOS
curl -fsSL https://raw.githubusercontent.com/murongg/dbsage/main/install.sh | bash

# Windows: Download install.bat and run as administrator
```

### Manual Install
```bash
git clone https://github.com/murongg/dbsage.git
cd dbsage/go
make setup
make run
```

## Quick Start

1. **Configure API Key**
   ```bash
   nano ~/.dbsage/config.env
   # Set: OPENAI_API_KEY=your_api_key_here
   ```

2. **Launch DBSage**
   ```bash
   dbsage
   ```

3. **Add Database**
   ```bash
   /add test connection
   # Follow prompts to enter database details
   ```

4. **Start Querying**
   ```bash
   "Show me all users created this month"
   "Find slow queries in the database"
   "Suggest indexes for the orders table"
   ```

## Commands

```bash
# Connection Management
/add test connection   # Add database connection
/switch production     # Switch database
/list                  # Show all connections

# Quick Access
@                      # Show connections
@production           # Quick switch
help                  # Show help
exit                  # Exit
```

## Configuration

```bash
# ~/.dbsage/config.env
OPENAI_API_KEY=your_openai_api_key_here
OPENAI_BASE_URL=https://api.openai.com/v1

# Optional database connection
DATABASE_URL=postgres://user:pass@localhost:5432/db
```

## Examples

```
You: "Query the first 10 users"
DBSage: SELECT * FROM users LIMIT 10;

You: "This query is slow: SELECT * FROM orders WHERE created_at > '2024-01-01'"
DBSage: 🚀 Optimization suggestions:
        1. Add index: CREATE INDEX idx_orders_created_at ON orders (created_at);
        2. Avoid SELECT *, specify required columns
        3. Consider date partitioning for large tables

You: "DELETE FROM users WHERE active = false"
DBSage: ⚠️ DANGEROUS OPERATION DETECTED ⚠️
        This will DELETE 1,247 records. Please confirm.
```

## Links

- **Documentation**: [GitHub Repository](https://github.com/murongg/dbsage)
- **Issues**: [GitHub Issues](https://github.com/murongg/dbsage/issues)
- **Releases**: [GitHub Releases](https://github.com/murongg/dbsage/releases)
