#!/bin/bash

# Setup script for MongoDB Change Stream PoC
set -e

echo "🚀 Setting up MongoDB Change Stream PoC..."

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed. Please install Docker first."
    exit 1
fi

# Check if Docker Compose is installed
if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose is not installed. Please install Docker Compose first."
    exit 1
fi

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go first."
    exit 1
fi

echo "✅ Prerequisites check passed!"

# Create .env file if it doesn't exist
if [ ! -f .env ]; then
    echo "📝 Creating .env file..."
    cp .env.example .env
    echo "✅ .env file created from template"
else
    echo "ℹ️  .env file already exists"
fi

# Download Go dependencies
echo "📦 Downloading Go dependencies..."
go mod tidy

# Start MongoDB
echo "🗄️  Starting MongoDB with Docker Compose..."
docker-compose up -d mongodb mongodb-setup

# Wait for MongoDB to be ready
echo "⏳ Waiting for MongoDB to be ready..."
sleep 15

# Check if MongoDB is ready
echo "🔍 Checking MongoDB connection..."
if docker-compose exec mongodb mongosh --eval "db.adminCommand('ping')" > /dev/null 2>&1; then
    echo "✅ MongoDB is ready!"
else
    echo "❌ MongoDB is not ready. Please check Docker logs with: docker-compose logs mongodb"
    exit 1
fi

echo ""
echo "🎉 Setup completed successfully!"
echo ""
echo "📋 Next steps:"
echo "   1. Start the watcher: go run watcher/main.go"
echo "   2. In another terminal, start the inserter: go run inserter/main.go"
echo "   3. Or use VS Code launch configuration to run both"
echo ""
echo "🔧 Useful commands:"
echo "   - View MongoDB logs: docker-compose logs -f mongodb"
echo "   - Stop MongoDB: docker-compose down"
echo "   - Connect to MongoDB: mongosh 'mongodb://admin:password123@localhost:27017/changestream_poc?authSource=admin&replicaSet=rs0'"
echo ""
