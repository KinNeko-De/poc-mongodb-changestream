@echo off
REM Script to properly set up MongoDB for change streams with correct replica set configuration

echo Stopping and removing existing containers...
docker-compose down -v

echo Starting MongoDB with simple configuration...
docker-compose -f docker-compose-simple.yml up -d mongodb

echo Waiting for MongoDB to start...
timeout /t 10 /nobreak

echo Initializing replica set with proper localhost configuration...
docker exec -it poc-mongodb mongosh --eval "rs.initiate({ _id: 'rs0', members: [{ _id: 0, host: 'localhost:27017' }] })"

echo Creating database and collection...
docker exec -it poc-mongodb mongosh --eval "use changestream_poc; db.createCollection('test_data')"

echo MongoDB is now ready with proper configuration!
echo You can now run the watcher and inserter applications.
