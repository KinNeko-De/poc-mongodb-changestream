package metadata

import (
	"context"
	"fmt"
	"time"
)

func MiningFileMetadata(ctx context.Context) error {
	fmt.Println("Mining file metadata...")

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Context cancelled, mining metadata stopped")
			return ctx.Err()
		case <-time.After(5 * time.Second):
			fmt.Println("No file metadata to mine at the moment, waiting...")
		}
	}
}
