# MongoDB Change Stream PoC

This project demonstrates MongoDB change streams using Go applications.

## Overview

The project consists of two Go applications:

1. **Inserter** (`inserter/main.go`): Inserts random data into MongoDB using 2 parallel goroutines
2. **Watcher** (`watcher/main.go`): Watches for changes in the MongoDB collection using change streams

## Prerequisites

- Go 1.21 or later
- Docker and Docker Compose (recommended)
- MongoDB running locally on port 27017 (if not using Docker)
- MongoDB should be configured as a replica set (required for change streams)

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
   go run watcher/main.go
   
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

Both applications support configuration via environment variables:

- `MONGO_URI`: MongoDB connection string (default: `mongodb://localhost:27017`)

For Docker setup, the connection string is:
```
mongodb://admin:password123@localhost:27017/changestream_poc?authSource=admin&replicaSet=rs0
```

## Usage

### Running the Watcher

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
