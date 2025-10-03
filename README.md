# DBSage - Database AI Assistant

An intelligent database management tool that brings AI-powered assistance to your database operations.

## Features

- **🧠 AI-Powered**: Convert natural language queries into optimized SQL
- **🛡️ Safety First**: Built-in protection against dangerous operations 
- **🔌 Multi-Database**: Support for PostgreSQL, MySQL, and SQLite
- **💻 Cross-Platform**: Works on Linux, macOS, and Windows

## Installation

### Quick Install

#### Global Installation (Recommended)
```bash
# Linux / macOS - Install to /usr/local/bin (requires sudo)
curl -fsSL https://raw.githubusercontent.com/murongg/dbsage/main/install.sh | sudo bash

# Windows: Download install.bat and run as administrator
```

#### Local Installation
```bash
# Linux / macOS - Install to current directory (no sudo required)
curl -fsSL https://raw.githubusercontent.com/murongg/dbsage/main/install.sh | bash -s -- --local

# Run with ./dbsage in the installation directory
```

#### Installation Options
```bash
# Force reinstallation
curl -fsSL https://raw.githubusercontent.com/murongg/dbsage/main/install.sh | sudo bash -s -- --force

# Install to custom directory
curl -fsSL https://raw.githubusercontent.com/murongg/dbsage/main/install.sh | sudo bash -s -- --dir /opt/bin

# Show help
curl -fsSL https://raw.githubusercontent.com/murongg/dbsage/main/install.sh | bash -s -- --help
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
# Command Line Options
dbsage --version       # Show version information
dbsage -v             # Show version information (short)
dbsage --help         # Show usage help

# Connection Management
/add test connection   # Add database connection
/switch production     # Switch database
/list                  # Show all connections
/remove test          # Remove connection

# General Commands
/help                 # Show available commands
/clear                # Clear screen
/exit or /quit        # Exit application

# Quick Access
@                      # Show connections
@production           # Quick switch to connection
@<query>              # Execute database query directly
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
