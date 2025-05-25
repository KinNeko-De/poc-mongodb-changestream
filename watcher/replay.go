// replay.go - Utility to replay all documents in a collection
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// CommandFlags struct in flags.go should be updated to include:
// ReplayAll bool

// ReplayAllDocuments fetches all documents in the collection and processes them as insert events
func ReplayAllDocuments(ctx context.Context, config *Config, metrics *Metrics) error {
	log.Println("Starting to replay all documents in the collection...")

	// Connect to MongoDB
	clientOptions := options.Client().ApplyURI(config.MongoURI)
	clientOptions.SetHosts([]string{"localhost:27017"})
	clientOptions.SetConnectTimeout(config.ConnectTimeout)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB for replay: %w", err)
	}
	defer client.Disconnect(ctx)

	// Test the connection
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return fmt.Errorf("failed to ping MongoDB for replay: %w", err)
	}

	// Get the collection
	coll := client.Database(config.Database).Collection(config.Collection)

	// Create a cursor to iterate through all documents
	findOptions := options.Find()
	cursor, err := coll.Find(ctx, bson.D{}, findOptions)
	if err != nil {
		return fmt.Errorf("failed to query documents for replay: %w", err)
	}
	defer cursor.Close(ctx)

	// Process each document as if it was an insert
	documentCount := 0
	startTime := time.Now()

	for cursor.Next(ctx) {
		var document bson.M
		if err := cursor.Decode(&document); err != nil {
			log.Printf("Error decoding document during replay: %v", err)
			metrics.IncrementProcessingError()
			continue
		}

		// Create a simulated change event
		changeEvent := bson.M{
			"operationType": "insert",
			"fullDocument":  document,
			"ns": bson.M{
				"db":   config.Database,
				"coll": config.Collection,
			},
			"documentKey":       bson.M{"_id": document["_id"]},
			"clusterTime":       time.Now(),
			"wallTime":          time.Now(),
			"_replaySimulation": true, // Mark as simulated for custom handling
		}

		// Process the simulated event
		processChangeEvent(changeEvent, metrics)
		documentCount++

		// Add a small delay to prevent overwhelming the system
		if documentCount%100 == 0 {
			log.Printf("Replayed %d documents so far...", documentCount)
			time.Sleep(10 * time.Millisecond)
		}
	}

	if err := cursor.Err(); err != nil {
		return fmt.Errorf("cursor error during replay: %w", err)
	}

	duration := time.Since(startTime)
	log.Printf("Replay completed: %d documents processed in %v (%.1f docs/sec)",
		documentCount, duration, float64(documentCount)/duration.Seconds())

	return nil
}
