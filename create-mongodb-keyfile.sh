#!/bin/bash

# This script creates a MongoDB keyfile with proper permissions for replica set authentication
# MongoDB requires the keyfile to be readable only by the owner (600 permissions)

KEYFILE="docker/mongodb.key"
TEMP_KEYFILE="docker/mongodb.key.tmp"

echo "Creating MongoDB keyfile for replica set authentication..."

# Generate a secure random key
openssl rand -base64 756 > "$TEMP_KEYFILE"

# Ensure proper permissions (600 - owner read/write only)
chmod 600 "$TEMP_KEYFILE"

# Move to final location
mv "$TEMP_KEYFILE" "$KEYFILE"

echo "MongoDB keyfile created successfully with proper permissions."
echo "Now you can run 'docker-compose up -d' to start MongoDB."
