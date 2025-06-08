package metadata

import (
	"context"
	"fmt"
	"time"
)

func ProduceFileMetadata(ctx context.Context) error {
	fmt.Println("Producing file metadata...")
	select {
	case <-ctx.Done():
		fmt.Println("Context cancelled, producing metadata stopped")
		return ctx.Err()
	case <-time.After(20 * time.Second):
		fmt.Println("File metadata produced")
		return nil // 5 seconds passed, action completed
	}

	return nil
}
