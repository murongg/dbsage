# DBSage - Database Sage

A professional database management AI expert tool, developed in Go, focused on PostgreSQL database operations and AI intelligent analysis. Just as Claude Code CLI focuses on programming, DBSage focuses on database management.

## 🚀 Features

### 🔍 Database Operations
- **SQL Execution** - Safely execute SQL queries
- **Table Management** - View all tables and table structures
- **Performance Analysis** - EXPLAIN query analysis and slow query detection
- **Index Management** - View and analyze table indexes
- **Data Statistics** - Get table statistics and data analysis

### 🧙‍♂️ AI Intelligence
- **Natural Language Queries** - Describe needs in natural language, DBSage automatically generates SQL
- **Smart Tool Calling** - DBSage automatically selects appropriate database tools
- **Recursive Analysis** - Supports multi-step complex task processing
- **Safety Checks** - Automatically identifies dangerous operations and alerts

### 🎨 User Interface
- **Colorful Terminal** - Beautiful command line interface
- **Real-time Feedback** - Loading animations and status indicators
- **Command Support** - Built-in commands like help, clear, exit
- **History** - Maintains conversation context

## 📋 System Requirements

- Go 1.21+
- PostgreSQL database
- OpenAI API Key

## 🛠️ Quick Start

### Method 1: Using Makefile (Recommended)

```bash
# 1. Enter the Go project directory
cd go/

# 2. View available commands
make help

# 3. Initialize setup
make setup

# 4. Edit configuration file
nano .env
# Set your OPENAI_API_KEY and DATABASE_URL

# 5. Run the program
make run
```

### Method 2: Using Scripts

```bash
# 1. Enter the Go project directory
cd go/

# 2. Run setup script
./setup.sh

# 3. Edit configuration file
nano .env
# Set your OPENAI_API_KEY and DATABASE_URL

# 4. Start the program
./run.sh
```

### Method 3: Manual Setup

```bash
# 1. Enter the Go project directory
cd go/

# 2. Copy configuration file
cp config.example .env

# 3. Edit configuration file
nano .env
# Set your OPENAI_API_KEY and DATABASE_URL

# 4. Start the program
./run.sh
```

### Method 4: Environment Variables

```bash
# Set environment variables
export OPENAI_API_KEY="your_openai_api_key_here"
export OPENAI_BASE_URL="https://api.openai.com/v1"  # Optional
export DATABASE_URL="postgres://username:password@localhost:5432/database?sslmode=disable"

# Run directly
go run cmd/dbsage/main.go
```

### 🔧 Other Running Methods

```bash
# Build binary
go build -o dbsage cmd/dbsage/main.go
./dbsage

# Or run source code directly
go run cmd/dbsage/main.go
```

## 🔧 Available Tools

| Tool Name | Description |
|---------|---------|
| `execute_sql` | Execute SQL queries |
| `get_all_tables` | Get all tables list |
| `get_table_schema` | Get table structure details |
| `explain_query` | Analyze query performance |
| `get_table_indexes` | Get table index information |
| `get_table_stats` | Get table statistics data |
| `find_duplicate_data` | Find duplicate records |
| `get_slow_queries` | Get slow queries list |
| `get_database_size` | Get database size |
| `get_table_sizes` | Get table size information |
| `get_active_connections` | Get active connections |

## 💡 Usage Examples

```
👤 You: Help me query the first 10 records from the users table
🧙‍♂️ DBSage: I'll query the first 10 records from the users table for you.

🔍 Execute SQL: SELECT * FROM users LIMIT 10;

Query results: ...

👤 You: Analyze the performance bottlenecks of this database
🧙‍♂️ DBSage: I'll help you analyze database performance, starting with slow queries...

🐌 Get slow queries list
📏 Get all table sizes
🔗 Get active connection information

Based on the analysis results, the following performance issues were found: ...
```

## 🎯 Built-in Commands

- `help` - Show feature list
- `clear` - Clear screen and redisplay welcome message
- `exit` / `quit` - Exit program

## 🔒 Security Features

- **SQL Injection Protection** - Automatically detect suspicious input
- **Dangerous Operation Alerts** - Operations like DROP, TRUNCATE require confirmation
- **Parameterized Queries** - Recommend safe query methods
- **Production Environment Protection** - Ask environment type before critical operations

## 📁 Project Structure

```
go/
├── cmd/
│   └── dbsage/
│       └── main.go          # Main program entry
├── pkg/
│   └── database/            # Public database package
│       ├── connection_manager.go
│       ├── service.go
│       ├── tools.go
│       └── url_parser.go
├── internal/
│   ├── ui/                  # User interface components
│   │   ├── interface.go
│   │   ├── state.go
│   │   ├── commands.go
│   │   └── *.go
│   └── ai/                  # AI client components
│       ├── client.go
│       ├── prompts.go
│       └── tools.go
├── configs/
│   └── config.example       # Configuration example
├── scripts/
│   └── run.sh              # Startup script
├── go.mod                   # Go module definition
└── README.md               # Documentation
```

## 🐛 Troubleshooting

### Common Issues

1. **Database Connection Failed**
   - Check if DATABASE_URL format is correct
   - Confirm database service is running
   - Verify username and password are correct

2. **OpenAI API Call Failed**
   - Check if OPENAI_API_KEY is set
   - Verify API key is valid
   - Check network connection is normal

3. **Permission Error**
   - Ensure database user has sufficient permissions
   - Check table and view access permissions

## 📝 Development Notes

This project follows standard Go project structure:
- `cmd/` - Executable programs
- `pkg/` - Importable packages
- `internal/` - Internal packages (if needed)

Main dependencies:
- `github.com/lib/pq` - PostgreSQL driver
- `github.com/sashabaranov/go-openai` - OpenAI client
- `github.com/charmbracelet/bubbletea` - Terminal UI framework

## 📄 License

[Add your license information]

## 🤝 Contributing

Welcome to submit Issues and Pull Requests!
