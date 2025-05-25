package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Document represents the structure of documents we'll insert
type Document struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Name      string             `bson:"name"`
	Value     int                `bson:"value"`
	Timestamp time.Time          `bson:"timestamp"`
	Source    string             `bson:"source"`
}

const (
	defaultMongoURI = "mongodb://localhost:27017/?replicaSet=rs0"
	database        = "changestream_poc"
	collection      = "test_data"
	insertCount     = 100 // Number of documents each goroutine will insert
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

	coll := client.Database(database).Collection(collection)

	// Create a WaitGroup to wait for both goroutines to finish
	var wg sync.WaitGroup
	wg.Add(2)

	// Start first goroutine
	go func() {
		defer wg.Done()
		insertRandomData(coll, "goroutine-1", insertCount)
	}()

	// Start second goroutine
	go func() {
		defer wg.Done()
		insertRandomData(coll, "goroutine-2", insertCount)
	}()

	// Wait for both goroutines to complete
	wg.Wait()

	fmt.Println("All insertions completed!")
}

// insertRandomData inserts random documents into the MongoDB collection
func insertRandomData(coll *mongo.Collection, source string, count int) {
	fmt.Printf("Starting data insertion from %s\n", source)

	names := []string{"Alice", "Bob", "Charlie", "Diana", "Eve", "Frank", "Grace", "Henry", "Ivy", "Jack"}

	for i := 0; i < count; i++ {
		doc := Document{
			Name:      names[rand.Intn(len(names))],
			Value:     rand.Intn(1000),
			Timestamp: time.Now(),
			Source:    source,
		}

		result, err := coll.InsertOne(context.TODO(), doc)
		if err != nil {
			log.Printf("Error inserting document from %s: %v", source, err)
			continue
		}

		fmt.Printf("%s inserted document with ID: %v\n", source, result.InsertedID)

		// Add a small delay to make the insertions more observable
		time.Sleep(time.Millisecond * 100)
	}

	fmt.Printf("Finished data insertion from %s\n", source)
}
