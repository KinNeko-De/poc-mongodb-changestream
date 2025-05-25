#!/bin/bash

# Script to completely reset the MongoDB Docker setup
echo "Stopping and removing all Docker containers..."
docker-compose down

echo "Removing MongoDB data volume..."
docker volume rm poc-mongodb-changestream_mongodb_data || true

echo "Generating new MongoDB keyfile..."
openssl rand -base64 756 > docker/mongodb.key
chmod 600 docker/mongodb.key  # Change to 600 for proper permissions

echo "Starting MongoDB with Docker Compose..."
docker-compose up -d mongodb mongodb-setup

echo "Waiting for MongoDB to initialize..."
sleep 10

echo "Done! MongoDB should be running with a fresh configuration."
echo "Check container status with: docker-compose ps"
echo "Check logs with: docker-compose logs -f mongodb"

echo "Waiting for MongoDB to initialize..."
sleep 15

echo "Checking MongoDB status..."
docker-compose logs mongodb | tail -n 20

echo "Reset complete! MongoDB should now be running with replica set configuration."
echo "Connection string: mongodb://admin:password123@localhost:27017/changestream_poc?authSource=admin&replicaSet=rs0"
