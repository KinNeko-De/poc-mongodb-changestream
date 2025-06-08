package metadata

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

func ProduceFileMetadata(ctx context.Context) error {
	fmt.Println("Producing file metadata...")
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

func CreateJitteredDelay() time.Duration {
	jitter := time.Duration(rand.Intn(3000)+2000) * time.Millisecond
	return jitter
}
