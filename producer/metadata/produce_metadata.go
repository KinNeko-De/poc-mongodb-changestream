package metadata

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	minDelay = 2
	maxDelay = 5
	client   *mongo.Client
)

func ProduceFileMetadata(ctx context.Context) error {
	fmt.Println("Producing file metadata...")

	if err := initializeMongoClient(ctx); err != nil {
		return err
	}
	defer disconnectMongoClient(ctx) // Ensure the client is disconnected when the function exits

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Context cancelled, producing metadata stopped")
			return ctx.Err()
		case <-time.After(CreateJitteredDelay()):
			fmt.Println("File metadata produced")
		}
	}
}

func initializeMongoClient(ctx context.Context) error {
	if client == nil {
		var err error
		client, err = mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
		if err != nil {
			return fmt.Errorf("failed to connect to MongoDB: %v", err)
		}
		fmt.Println("MongoDB client initialized")

		// Ping the database to ensure it is reachable
		if err := client.Ping(ctx, nil); err != nil {
			return fmt.Errorf("failed to ping MongoDB: %v", err)
		}
		fmt.Println("MongoDB ping successful")
	}
	return nil
}

func disconnectMongoClient(ctx context.Context) {
	if client != nil {
		if err := client.Disconnect(ctx); err != nil {
			fmt.Printf("Failed to disconnect MongoDB client: %v\n", err)
		} else {
			fmt.Println("MongoDB client disconnected")
		}
	}
}

func CreateJitteredDelay() time.Duration {
	jitter := time.Duration(rand.Intn(maxDelay-minDelay)+minDelay) * time.Second
	fmt.Printf("Jittered delay: %v\n", jitter)
	return jitter
}
