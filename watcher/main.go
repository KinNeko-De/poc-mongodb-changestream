package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	defaultMongoURI = "mongodb://localhost:27017/?replicaSet=rs0"
	database        = "changestream_poc"
	collection      = "test_data"
)

// getMongoURI returns the MongoDB URI from environment variable or default
func getMongoURI() string {
	if uri := os.Getenv("MONGO_URI"); uri != "" {
		return uri
	}
	return defaultMongoURI
}

func main() {
	// Connect to MongoDB - use hardcoded URI directly without environment variables
	mongoURI := "mongodb://localhost:27017/?replicaSet=rs0"
	fmt.Printf("Connecting to MongoDB at: %s\n", mongoURI)

	clientOptions := options.Client().ApplyURI(mongoURI)
	// Force the MongoDB driver to only use the address we specify
	clientOptions.SetHosts([]string{"localhost:27017"})
	// Don't set direct mode as it's incompatible with replica sets

	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer client.Disconnect(context.TODO())

	// Test the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal("Failed to ping MongoDB:", err)
	}

	fmt.Println("Connected to MongoDB successfully!")
	fmt.Println("Starting change stream watcher...")

	// Get the collection
	coll := client.Database(database).Collection(collection)

	// Create a change stream
	pipeline := mongo.Pipeline{}
	changeStream, err := coll.Watch(context.TODO(), pipeline)
	if err != nil {
		log.Fatal("Failed to create change stream:", err)
	}
	defer changeStream.Close(context.TODO())

	// Create a channel to handle OS signals for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Create a context for the change stream operations
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start watching for changes in a separate goroutine
	go func() {
		fmt.Println("Watching for changes...")
		for changeStream.Next(ctx) {
			var changeEvent bson.M
			if err := changeStream.Decode(&changeEvent); err != nil {
				log.Printf("Error decoding change event: %v", err)
				continue
			}

			// Process the change event
			processChangeEvent(changeEvent)
		}

		if err := changeStream.Err(); err != nil {
			log.Printf("Change stream error: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-sigChan
	fmt.Println("\nReceived interrupt signal, shutting down gracefully...")
	cancel()
}

// processChangeEvent handles incoming change stream events
func processChangeEvent(event bson.M) {
	operationType, ok := event["operationType"].(string)
	if !ok {
		log.Println("Could not determine operation type")
		return
	}

	fmt.Printf("Change detected: %s\n", operationType)

	switch operationType {
	case "insert":
		handleInsert(event)
	case "update":
		handleUpdate(event)
	case "replace":
		handleReplace(event)
	case "delete":
		handleDelete(event)
	default:
		fmt.Printf("Unhandled operation type: %s\n", operationType)
	}
}

// handleInsert processes insert operations
func handleInsert(event bson.M) {
	fullDocument, ok := event["fullDocument"].(bson.M)
	if !ok {
		log.Println("Could not extract full document from insert event")
		return
	}

	fmt.Printf("Document inserted: %+v\n", fullDocument)

	// TODO: Add custom logic for handling insert operations
	// This is where you can add specific business logic when a document is inserted
	// Examples:
	// - Send notifications
	// - Update caches
	// - Trigger workflows
	// - Log to external systems
	// - Update metrics
}

// handleUpdate processes update operations
func handleUpdate(event bson.M) {
	documentKey, ok := event["documentKey"].(bson.M)
	if !ok {
		log.Println("Could not extract document key from update event")
		return
	}

	fmt.Printf("Document updated with key: %+v\n", documentKey)

	// Check if we have the full document after update
	if fullDocument, exists := event["fullDocument"]; exists {
		fmt.Printf("Updated document: %+v\n", fullDocument)
	}

	// TODO: Add custom logic for handling update operations
	// This is where you can add specific business logic when a document is updated
}

// handleReplace processes replace operations
func handleReplace(event bson.M) {
	documentKey, ok := event["documentKey"].(bson.M)
	if !ok {
		log.Println("Could not extract document key from replace event")
		return
	}

	fmt.Printf("Document replaced with key: %+v\n", documentKey)

	if fullDocument, exists := event["fullDocument"]; exists {
		fmt.Printf("Replacement document: %+v\n", fullDocument)
	}

	// TODO: Add custom logic for handling replace operations
	// This is where you can add specific business logic when a document is replaced
}

// handleDelete processes delete operations
func handleDelete(event bson.M) {
	documentKey, ok := event["documentKey"].(bson.M)
	if !ok {
		log.Println("Could not extract document key from delete event")
		return
	}

	fmt.Printf("Document deleted with key: %+v\n", documentKey)

	// TODO: Add custom logic for handling delete operations
	// This is where you can add specific business logic when a document is deleted
}
