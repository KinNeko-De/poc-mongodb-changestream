@echo off
REM This script creates a MongoDB keyfile with proper permissions for replica set authentication

set KEYFILE=docker\mongodb.key
set TEMP_KEYFILE=docker\mongodb.key.tmp

echo Creating MongoDB keyfile for replica set authentication...

REM Generate a random keyfile
REM On Windows, we need a different approach than openssl
echo mongodb-keyfile-random-content-%RANDOM%%RANDOM%%RANDOM%%RANDOM% > %TEMP_KEYFILE%

REM Copy to final location
copy /Y %TEMP_KEYFILE% %KEYFILE% > nul
del %TEMP_KEYFILE%

echo MongoDB keyfile created successfully.
echo Now you can run 'docker-compose up -d' to start MongoDB.
