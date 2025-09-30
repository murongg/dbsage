# DBSage - Database AI Assistant

An intelligent database management tool built with Go, focused on AI-driven operations and analysis for PostgreSQL, MySQL, and SQLite databases.

> **Note**: This is currently an MVP version with features under active development.

## Core Features

- **Natural Language Queries** - Describe your needs in natural language, automatically generates SQL
- **Smart Tool Selection** - Automatically chooses appropriate database tools
- **Performance Analysis** - EXPLAIN query analysis and slow query detection
- **Index Suggestions** - AI-powered index optimization recommendations
- **Safety Checks** - Automatically identifies dangerous operations and provides warnings
- **Multi-step Analysis** - Supports complex recursive task processing

## Installation Guide

### System Requirements
- **Operating System**: Linux, macOS, Windows
- **API Requirements**: OpenAI API Key

### One-Click Installation

#### Linux / macOS
```bash
# One-click installation
curl -fsSL https://raw.githubusercontent.com/murongg/dbsage/main/install.sh | bash

# Or download and run
wget https://raw.githubusercontent.com/murongg/dbsage/main/install.sh
chmod +x install.sh
./install.sh
```

#### Windows
```powershell
# Download and run installation script
# 1. Download install.bat script
# 2. Right-click "Run as administrator"
```

#### Manual Installation
```bash
# 1. Clone repository
git clone https://github.com/murongg/dbsage.git
cd dbsage/go

# 2. Install dependencies and build
make setup

# 3. Configure environment variables
nano .env
# Set OPENAI_API_KEY and OPENAI_BASE_URL

# 4. Run the program
make run
```

### Configuration Example

```bash
# Configuration file ~/.dbsage/config.env
OPENAI_API_KEY=your_openai_api_key_here
OPENAI_BASE_URL=https://api.openai.com/v1

# Database connection (optional, can also be added at runtime)
DATABASE_URL=postgres://username:password@localhost:5432/database?sslmode=disable

# Other configurations
LOG_LEVEL=info
MAX_CONNECTIONS=10
TIMEOUT=30s
```

## Quick Start

After installation, follow these steps to get started:

1. **Configure API Key**
   ```bash
   # Edit configuration file
   nano ~/.dbsage/config.env
   # Set OPENAI_API_KEY=your_api_key_here
   ```

2. **Launch DBSage**
   ```bash
   dbsage
   ```

3. **Add Database Connection**
   ```bash
   # Run in DBSage
   /add mydb
   # Follow prompts to enter database connection information
   ```

## Usage Examples

```
You: Query the first 10 records from the users table
DBSage: SELECT * FROM users LIMIT 10;

You: Analyze database performance bottlenecks
DBSage: Analyzing database performance, getting slow query list...

You: Suggest indexes for users table
DBSage: Index suggestions:
1. CREATE INDEX idx_users_email ON users (email);
2. CREATE INDEX idx_users_created_at ON users (created_at);

You: Optimize this slow query: SELECT * FROM orders WHERE created_at > '2024-01-01'
DBSage: Optimization suggestions:
1. Avoid SELECT *, specify required columns
2. Add index on created_at column
3. Consider partitioning for large tables
```

## Command Reference

### Connection Management Commands
- `/add <connection_name>` - Add database connection
- `/switch <connection_name>` - Switch database connection
- `/list` - Display all database connections
- `/remove <connection_name>` - Remove database connection

### Quick Commands
- `@` - Display all connections
- `@<connection_name>` - Quick switch connection
- `@<SQL_query>` - Execute SQL query directly

### Basic Commands
- `help` - Show help information
- `clear` - Clear screen
- `exit` / `quit` - Exit program

## Key Features

### 🔍 Intelligent Analysis
- Natural language to SQL conversion
- Query performance analysis
- Index suggestions
- Performance bottleneck detection

### 🛡️ Security Features
- SQL injection detection
- Dangerous operation confirmation
- Parameterized query suggestions
- Production environment protection

### 🚀 Supported Databases
- PostgreSQL
- MySQL
- SQLite

## Contributing & Support

- **Documentation**: [https://github.com/murongg/dbsage](https://github.com/murongg/dbsage)
- **Issue Reports**: [GitHub Issues](https://github.com/murongg/dbsage/issues)
- **Configuration Directory**: `~/.dbsage/`
