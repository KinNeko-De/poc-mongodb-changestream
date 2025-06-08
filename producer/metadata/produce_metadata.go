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

var minDelay = 2
var maxDelay = 5

func CreateJitteredDelay() time.Duration {
	jitter := time.Duration(rand.Intn(maxDelay-minDelay)+minDelay) * time.Second
	fmt.Printf("Jittered delay: %v\n", jitter)
	return jitter
}
