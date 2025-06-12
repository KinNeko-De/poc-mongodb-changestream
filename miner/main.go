package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/kinneko-de/poc-mongodb-changestream/miner/metadata"
)

func main() {
	fmt.Println("Starting miner...")

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
	defer cancel()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := metadata.MiningFileMetadata(ctx)
		if err != nil && !os.IsTimeout(err) && err != context.Canceled && err != context.DeadlineExceeded {
			fmt.Printf("Error producing file metadata: %v\n", err)
		}
	}()

	wg.Wait()
	fmt.Println("Shutting down miner...")
}
