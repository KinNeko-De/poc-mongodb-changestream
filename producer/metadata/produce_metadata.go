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
	defer disconnectMongoClient()

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
		clientOptions := options.Client().
			ApplyURI("mongodb://localhost:27017/?replicaSet=rs0")
		client, err = mongo.Connect(ctx, clientOptions)
		if err != nil {
			return fmt.Errorf("failed to connect to MongoDB: %v", err)
		}
		fmt.Println("MongoDB client initialized")

		if err := client.Ping(ctx, nil); err != nil {
			return fmt.Errorf("failed to ping MongoDB: %v", err)
		}
		fmt.Println("MongoDB ping successful")
	}
	return nil
}

func disconnectMongoClient() {
	if client != nil {
		// Create a new context with timeout for disconnect operation, the application context might be cancelled
		disconnectCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := client.Disconnect(disconnectCtx); err != nil {
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
