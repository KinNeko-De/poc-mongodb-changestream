@echo off
REM Setup script for MongoDB Change Stream PoC on Windows
setlocal enabledelayedexpansion

echo 🚀 Setting up MongoDB Change Stream PoC...

REM Check if Docker is installed
docker --version >nul 2>&1
if errorlevel 1 (
    echo ❌ Docker is not installed. Please install Docker first.
    exit /b 1
)

REM Check if Docker Compose is installed
docker-compose --version >nul 2>&1
if errorlevel 1 (
    echo ❌ Docker Compose is not installed. Please install Docker Compose first.
    exit /b 1
)

REM Check if Go is installed
go version >nul 2>&1
if errorlevel 1 (
    echo ❌ Go is not installed. Please install Go first.
    exit /b 1
)

echo ✅ Prerequisites check passed!

REM Create .env file if it doesn't exist
if not exist .env (
    echo 📝 Creating .env file...
    copy .env.example .env >nul
    echo ✅ .env file created from template
) else (
    echo ℹ️  .env file already exists
)

REM Download Go dependencies
echo 📦 Downloading Go dependencies...
go mod tidy

REM Start MongoDB
echo 🗄️  Starting MongoDB with Docker Compose...
docker-compose up -d mongodb mongodb-setup

REM Wait for MongoDB to be ready
echo ⏳ Waiting for MongoDB to be ready...
timeout /t 15 /nobreak >nul

REM Check if MongoDB is ready
echo 🔍 Checking MongoDB connection...
docker-compose exec mongodb mongosh --eval "db.adminCommand('ping')" >nul 2>&1
if errorlevel 1 (
    echo ❌ MongoDB is not ready. Please check Docker logs with: docker-compose logs mongodb
    exit /b 1
)

echo ✅ MongoDB is ready!
echo.
echo 🎉 Setup completed successfully!
echo.
echo 📋 Next steps:
echo    1. Start the watcher: go run watcher/main.go
echo    2. In another terminal, start the inserter: go run inserter/main.go
echo    3. Or use VS Code launch configuration to run both
echo.
echo 🔧 Useful commands:
echo    - View MongoDB logs: docker-compose logs -f mongodb
echo    - Stop MongoDB: docker-compose down
echo    - Connect to MongoDB: mongosh "mongodb://admin:password123@localhost:27017/changestream_poc?authSource=admin&replicaSet=rs0"
echo.
