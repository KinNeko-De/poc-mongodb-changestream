#!/bin/bash
# health-check.sh - Simple script to check the health of the watcher
# Returns 0 if the watcher is healthy, 1 otherwise

# Path to the metrics file
METRICS_FILE="${METRICS_FILE:-watcher_metrics.json}"

# Check if the metrics file exists
if [ ! -f "$METRICS_FILE" ]; then
    echo "ERROR: Metrics file not found: $METRICS_FILE"
    exit 1
fi

# Parse the metrics file
HEALTH=$(grep -o '"health":"[^"]*"' "$METRICS_FILE" | cut -d'"' -f4)
LAST_EVENT=$(grep -o '"last_event":"[^"]*"' "$METRICS_FILE" | cut -d'"' -f4)

# Get current timestamp
NOW=$(date +%s)

# Parse the last event timestamp
if [[ -n "$LAST_EVENT" ]]; then
    # Convert ISO timestamp to epoch seconds for comparison
    LAST_EVENT_EPOCH=$(date -d "$LAST_EVENT" +%s 2>/dev/null || date -j -f "%Y-%m-%dT%H:%M:%SZ" "$LAST_EVENT" +%s 2>/dev/null)
    
    # Check if the timestamp is valid
    if [[ -n "$LAST_EVENT_EPOCH" ]]; then
        # Calculate the difference in seconds
        DIFF=$((NOW - LAST_EVENT_EPOCH))
        
        # If the last event was more than 5 minutes ago, consider it unhealthy
        if [[ $DIFF -gt 300 ]]; then
            echo "WARNING: No events processed in the last 5 minutes"
            exit 1
        fi
    else
        echo "ERROR: Could not parse last event timestamp: $LAST_EVENT"
    fi
fi

# Check the health status from the metrics file
if [[ "$HEALTH" != "true" ]]; then
    echo "ERROR: Watcher reports unhealthy status"
    exit 1
fi

# If we get here, the watcher is healthy
echo "OK: Watcher is healthy"
exit 0
