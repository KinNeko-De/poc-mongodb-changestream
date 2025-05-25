@echo off
REM This script sets up a simple MongoDB instance with replica set for change streams

echo Stopping any existing MongoDB containers...
docker stop mongodb-rs0 2>nul
docker rm mongodb-rs0 2>nul

echo Creating keyfile directory...
mkdir -p %USERPROFILE%\mongodb\keyfile

echo Creating MongoDB keyfile...
echo thisisasecurekeyfilewithatleast6characters > %USERPROFILE%\mongodb\keyfile\mongodb.key

echo Setting keyfile permissions...
icacls %USERPROFILE%\mongodb\keyfile\mongodb.key /inheritance:r
icacls %USERPROFILE%\mongodb\keyfile\mongodb.key /grant:r "%USERNAME%":F

echo Starting MongoDB with replica set...
docker run -d --name mongodb-rs0 -p 27017:27017 ^
  -v %USERPROFILE%\mongodb\data:/data/db ^
  -v %USERPROFILE%\mongodb\keyfile:/keyfile ^
  mongo:7.0 mongod --replSet rs0 --keyFile /keyfile/mongodb.key --bind_ip_all

echo Waiting for MongoDB to start...
timeout /t 10 /nobreak

echo Initializing replica set...
docker exec mongodb-rs0 mongosh --eval "rs.initiate({_id: 'rs0', members: [{_id: 0, host: 'localhost:27017'}]})"

echo MongoDB replica set is now running!
echo Connection string: mongodb://localhost:27017/?replicaSet=rs0
