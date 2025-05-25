@echo off
REM This script starts a simple MongoDB replica set without authentication

echo Stopping existing MongoDB containers...
docker-compose down

echo Starting simple MongoDB replica set...
docker-compose -f docker-compose-simple.yml up -d

echo Waiting for MongoDB to start...
timeout /t 15 /nobreak

echo MongoDB replica set should now be running!
echo Connection string: mongodb://localhost:27017/?replicaSet=rs0
echo You can run your applications now.
