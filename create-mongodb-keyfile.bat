@echo off
REM This script creates a MongoDB keyfile with proper permissions for replica set authentication

set KEYFILE=docker\mongodb.key

echo Creating MongoDB keyfile for replica set authentication...

REM Make sure the docker directory exists
if not exist docker mkdir docker

REM Generate a random string for the keyfile (using PowerShell)
powershell -Command "$randomKey = [System.Convert]::ToBase64String([System.Security.Cryptography.RandomNumberGenerator]::Create().GetBytes(756)); Set-Content -Path '%KEYFILE%' -Value $randomKey"

echo MongoDB keyfile created successfully.
echo Now you can run 'docker-compose up -d' to start MongoDB.
