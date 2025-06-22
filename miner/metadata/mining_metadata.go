package metadata

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/kinneko-de/poc-mongodb-changestream/golang/store_file/v1"
)

const ResumeTokenDirectory = "app/data"
const ResumeTokenFile = "resume_token.bin"

var (
	client              *mongo.Client
	ResumeTokenFilePath = filepath.Join(ResumeTokenDirectory, ResumeTokenFile)
)

func MiningFileMetadata(ctx context.Context) error {
	fmt.Println("Mining file metadata...")

	if err := initializeMongoClient(ctx); err != nil {
		return err
	}
	defer disconnectMongoClient()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Context cancelled, mining metadata stopped")
			return ctx.Err()
		default:
			return WatchChengeStream(ctx)
		}
	}
}

func WatchChengeStream(ctx context.Context) error {
	fmt.Println("Watching change stream for file metadata...")
	err := EnsureREsumeTokenDirectoryExists()
	if err != nil {
		return fmt.Errorf("failed to create resume token directory: %v", err)
	}

	resumeToken, err := FetchResumeToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch resume token: %v", err)
	}

	collection := client.Database("store_file").Collection("file")
	changeStreamOptions := options.ChangeStream()

	if resumeToken != nil {
		fmt.Println("Resuming change stream from previous token")
		changeStreamOptions = changeStreamOptions.SetResumeAfter(resumeToken)
	} else {
		// Fetch everything that is still retained in the oplog
		// Do not use this in production
		changeStreamOptions = changeStreamOptions.SetStartAtOperationTime(&primitive.Timestamp{T: 1})
	}

	changeStream, err := collection.Watch(ctx, mongo.Pipeline{}, changeStreamOptions)
	if err != nil {
		return fmt.Errorf("failed to watch change stream: %v", err)
	}
	defer changeStream.Close(ctx)

	for changeStream.Next(ctx) {
		var change bson.M
		if err := changeStream.Decode(&change); err != nil {
			return fmt.Errorf("failed to decode change stream event: %v", err)
		}
		fmt.Printf("Change detected: %v\n", change)

		// only works for insert change events
		if fullDoc, ok := change["fullDocument"].(bson.M); ok {
			// Process the full document as needed
			fmt.Printf("Full document: %v\n", fullDoc)

			// Extract fields and create protobuf message
			fileMsg := &pb.FileStored{}

			// Extract FileId (string)
			if fileId, ok := fullDoc["FileId"].(string); ok {
				fileMsg.SetFileId(fileId)
			}

			// Extract CreatedAt (time)
			if createdAt, ok := fullDoc["CreatedAt"].(primitive.DateTime); ok {
				fileMsg.SetCreatedAt(timestamppb.New(createdAt.Time()))
			}

			// Extract Size (int64)
			if size, ok := fullDoc["Size"].(int64); ok {
				fileMsg.SetSize(size)
			}

			// Extract MediaType (string)
			if mediaType, ok := fullDoc["MediaType"].(string); ok {
				fileMsg.SetMediaType(mediaType)
			}

			// Extract Extension (string)
			if extension, ok := fullDoc["Extension"].(string); ok {
				fileMsg.SetExtension(extension)
			}

			// Serialize to JSON (simplified for demo)
			fmt.Printf("Protobuf message: FileId=%s, Size=%d, MediaType=%s, Extension=%s\n",
				fileMsg.GetFileId(), fileMsg.GetSize(), fileMsg.GetMediaType(), fileMsg.GetExtension())
		}

		resumeToken := changeStream.ResumeToken()
		StoreResumeToken(ctx, resumeToken)
	}

	if err := changeStream.Err(); err != nil {
		return fmt.Errorf("error in change stream: %v", err)
	}

	return nil
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

func EnsureREsumeTokenDirectoryExists() any {
	return os.MkdirAll(ResumeTokenDirectory, 0755)
}

func StoreResumeToken(ctx context.Context, token bson.Raw) error {
	return os.WriteFile(ResumeTokenFilePath, token, 0644)
}

func FetchResumeToken(ctx context.Context) (bson.Raw, error) {
	data, err := os.ReadFile(ResumeTokenFilePath)
	if err != nil {
		// If file doesn't exist, return nil (no previous token)
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	return bson.Raw(data), nil
}
