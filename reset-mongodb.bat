@echo off
REM Script to completely reset the MongoDB Docker setup

echo Stopping and removing all Docker containers...
docker-compose down

echo Removing MongoDB data volume...
docker volume rm poc-mongodb-changestream_mongodb_data 2>nul

echo Generating new MongoDB keyfile...
REM Use PowerShell to generate a secure random key
powershell -Command "$randomKey = [System.Convert]::ToBase64String([System.Security.Cryptography.RandomNumberGenerator]::Create().GetBytes(756)); Set-Content -Path 'docker\mongodb.key' -Value $randomKey"

echo Starting MongoDB with Docker Compose...
docker-compose up -d mongodb mongodb-setup

echo Waiting for MongoDB to initialize...
timeout /t 10 /nobreak > nul

echo Done! MongoDB should be running with a fresh configuration.
echo Check container status with: docker-compose ps
echo Check logs with: docker-compose logs -f mongodb

echo Waiting for MongoDB to initialize...
timeout /t 15 /nobreak > nul

echo Checking MongoDB status...
docker-compose logs mongodb

echo Reset complete! MongoDB should now be running with replica set configuration.
echo Connection string: mongodb://admin:password123@localhost:27017/changestream_poc?authSource=admin^&replicaSet=rs0
