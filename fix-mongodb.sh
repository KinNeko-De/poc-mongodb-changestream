#!/bin/bash

# This script recreates the MongoDB keyfile and restarts the containers

echo "Stopping all MongoDB containers..."
docker-compose down

echo "Removing MongoDB volume..."
docker volume rm poc-mongodb-changestream_mongodb_data 2>/dev/null || true

echo "Creating a new MongoDB keyfile..."
echo "thisisasecurekeyfilewithatleast6characters" > docker/mongodb.key
chmod 400 docker/mongodb.key

echo "Starting MongoDB with Docker Compose..."
docker-compose up -d mongodb

echo "Waiting for MongoDB to start..."
sleep 15

echo "Starting MongoDB setup (replica set initialization)..."
docker-compose up -d mongodb-setup

echo "Done! MongoDB should be running with the new configuration."
echo "Check logs with: docker-compose logs mongodb"
