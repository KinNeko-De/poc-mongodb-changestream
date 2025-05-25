# MongoDB Change Stream PoC

This project demonstrates MongoDB change streams using Go applications with a focus on resilience and production-readiness.

## Overview

The project consists of two Go applications:

1. **Inserter** (`inserter/main.go`): Inserts random data into MongoDB using 2 parallel goroutines
2. **Watcher** (`watcher/main.go`): Watches for changes in the MongoDB collection using change streams with resilience features

## Prerequisites

- Go 1.21 or later
- Docker and Docker Compose (recommended)
- MongoDB running locally on port 27017 (if not using Docker)
- MongoDB should be configured as a replica set (required for change streams)

## Key Features

The watcher application includes several production-ready features:

1. **Resilient Connection Handling**:
   - Automatic reconnection with exponential backoff
   - Resume token persistence for continuing after restarts
   - Connection health monitoring
   
2. **Error Handling**:
   - Categorized error handling for different MongoDB error types
   - Special handling for resume token errors
   - Proper logging of errors with context
   
3. **Monitoring and Metrics**:
   - Tracks operation counts (inserts, updates, etc.)
   - Measures error rates and reconnection attempts
   - Provides health status information
   - Metrics persistence to file
   
4. **Configurability**:
   - Environment variable configuration
   - Reasonable defaults for all settings
   - Validation of configuration parameters

## Quick Start with Docker (Recommended)

The easiest way to get started is using the provided Docker Compose setup:

1. **Automated Setup** (Windows):
   ```bash
   ./setup.bat
   ```
   
   Or on Linux/Mac:
   ```bash
   chmod +x setup.sh
   ./setup.sh
   ```

2. **Manual Setup**:
   ```bash
   # Copy environment template
   cp .env.example .env
   
   # Start MongoDB with Docker Compose
   make docker-up
   # or: docker-compose up -d mongodb mongodb-setup
   
   # Download Go dependencies
   go mod tidy
   ```

3. **Run the applications**:
   ```bash
   # Terminal 1: Start the watcher
   go run watcher/*.go
   
   # Terminal 2: Start the inserter
   go run inserter/main.go
   ```

## Manual MongoDB Setup (Alternative)

If you prefer to install MongoDB manually, change streams require MongoDB to be running as a replica set:

1. Start MongoDB with replica set configuration:
   ```bash
   mongod --replSet rs0 --dbpath /path/to/your/data
   ```

2. Connect to MongoDB and initialize the replica set:
   ```bash
   mongosh
   ```
   ```javascript
   rs.initiate()
   ```

## Docker Commands

The project includes several useful Docker commands via Makefile:

```bash
make docker-up        # Start MongoDB in Docker
make docker-down      # Stop and remove containers
make docker-logs      # View MongoDB logs
make docker-build-apps # Build Go apps as Docker images
make docker-run-apps  # Run both apps in Docker
make docker-full      # Start everything in Docker
make dev-with-docker  # Start MongoDB in Docker, run Go apps locally
```

## Configuration

The watcher application supports extensive configuration via environment variables:

### MongoDB Connection
- `MONGO_URI`: MongoDB connection string (default: `mongodb://localhost:27017/?replicaSet=rs0`)
- `MONGO_DB`: Database name (default: `changestream_poc`)
- `MONGO_COLLECTION`: Collection name (default: `test_data`)
- `MONGO_CONNECT_TIMEOUT`: Connection timeout in seconds (default: `10s`)

### Change Stream Settings
- `RESUME_TOKEN_FILE`: File to store the resume token (default: `resume_token.json`)
- `BATCH_SIZE`: Number of changes to retrieve in a batch (default: `100`)
- `FULL_DOCUMENT_OPTION`: Full document option, either `updateLookup`, `whenAvailable`, or `required` (default: `updateLookup`)

### Reconnection Settings
- `INITIAL_BACKOFF`: Initial backoff duration (default: `1s`)
- `MAX_BACKOFF`: Maximum backoff duration (default: `1m`)
- `BACKOFF_MULTIPLIER`: Backoff multiplier for exponential backoff (default: `2.0`)
- `MAX_RECONNECT_TRIES`: Maximum reconnection attempts (default: `12`)

### Monitoring
- `METRICS_FILE`: File to store metrics (default: `watcher_metrics.json`)
- `LOG_FILE`: Log file path (default: `watcher.log`)
- `STATS_INTERVAL`: Interval to print statistics (default: `30s`)

### General
- `DEBUG`: Enable debug mode (default: `false`)

For Docker setup, the connection string is:
```
mongodb://admin:password123@localhost:27017/changestream_poc?authSource=admin&replicaSet=rs0
```

## Usage

### Running the Watcher

```bash
# Run with default settings
go run watcher/*.go

# Run with custom settings
MONGO_URI="mongodb://user:pass@localhost:27017/?replicaSet=rs0" \
RESUME_TOKEN_FILE="custom_token.json" \
METRICS_FILE="custom_metrics.json" \
go run watcher/*.go

# Reset the resume token (start fresh)
go run watcher/*.go --reset-token

# Start from a specific point in time
go run watcher/*.go --start-from=2023-05-01T00:00:00Z

# Show help
go run watcher/*.go --help
```

### Reset Resume Token

To re-process events from MongoDB, you need to reset the resume token:

1. Using the command-line flag:
   ```bash
   go run watcher/*.go --reset-token
   ```

2. Using the reset script:
   ```bash
   # On Linux/macOS
   ./reset-resume-token.sh
   
   # On Windows
   reset-resume-token.bat
   ```

3. Using the make target:
   ```bash
   make reset-token
   ```

4. Manually deleting the file:
   ```bash
   rm -f watcher/resume_token.json
   ```

> **Important Note about Change Streams:** MongoDB only retains change stream events 
> in the oplog for a limited time (typically a few hours to days). Once events expire
> from the oplog, you cannot retrieve them through a change stream, even after resetting
> the resume token. To process existing documents, use the `--replay-all` flag.

### Replaying All Documents

If you need to process all documents in a collection (not just recent changes):

```bash
# Process all existing documents as "insert" events, then watch for new changes
go run watcher/*.go --replay-all

# You can combine flags
go run watcher/*.go --reset-token --replay-all
```

This will fetch all documents in the collection and process them as if they were
newly inserted, then continue watching for actual changes.

### Running the Inserter

```bash
# Run with default settings
go run inserter/main.go
```

## Health Monitoring

The project includes health check scripts for monitoring the watcher:

- `health-check.sh` for Linux/macOS
- `health-check.bat` for Windows

These scripts check if the watcher is healthy by inspecting its metrics file.

## Extending the Applications

### Adding Custom Business Logic

The watcher application includes placeholder functions for custom business logic in the event handlers:

- `handleInsert`: For insert operations
- `handleUpdate`: For update operations
- `handleReplace`: For replace operations
- `handleDelete`: For delete operations

These functions can be extended to implement application-specific behavior such as:
- Sending notifications
- Updating caches
- Triggering workflows
- Logging to external systems
- Updating metrics

Start the watcher first to capture all change events:

```bash
go run watcher/main.go
```

The watcher will:
- Connect to MongoDB
- Create a change stream on the `changestream_poc.test_data` collection
- Listen for insert, update, replace, and delete operations
- Log all change events to the console
- Run until interrupted (Ctrl+C)

### Running the Inserter

In a separate terminal, run the inserter:

```bash
go run inserter/main.go
```

The inserter will:
- Connect to MongoDB
- Start 2 goroutines that run in parallel
- Each goroutine inserts 100 random documents
- Documents include random names, values, timestamps, and source identification

## Project Structure

```
.
├── LICENSE
├── README.md
├── go.mod
├── inserter/
│   └── main.go     # Application that inserts random data
└── watcher/
    └── main.go     # Application that watches change streams
```

## Configuration

Both applications use the following default configuration:
- MongoDB URI: `mongodb://localhost:27017`
- Database: `changestream_poc`
- Collection: `test_data`

You can modify these constants in the respective main.go files.

## Document Structure

The inserted documents have the following structure:

```go
type Document struct {
    ID        primitive.ObjectID `bson:"_id,omitempty"`
    Name      string             `bson:"name"`
    Value     int                `bson:"value"`
    Timestamp time.Time          `bson:"timestamp"`
    Source    string             `bson:"source"`
}
```

## TODO Items

The watcher application includes TODO comments in the event handlers where you can add custom business logic:

- Handle insert operations
- Handle update operations  
- Handle replace operations
- Handle delete operations

## License

MIT License - see LICENSE file for details.
