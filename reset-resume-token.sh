#!/bin/bash
# reset-resume-token.sh - Script to reset the resume token for the watcher

# Default file path
RESUME_TOKEN_FILE="watcher/resume_token.json"

# Check if file exists
if [ -f "$RESUME_TOKEN_FILE" ]; then
    echo "Removing resume token file: $RESUME_TOKEN_FILE"
    rm -f "$RESUME_TOKEN_FILE"
    echo "Resume token has been reset. The watcher will start with a new change stream."
else
    echo "No resume token file found at $RESUME_TOKEN_FILE"
    echo "The watcher will start with a new change stream on next run."
fi
