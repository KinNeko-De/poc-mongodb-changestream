package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/kinneko-de/poc-mongodb-changestream/producer/metadata"
)

func main() {
	fmt.Println("Starting producer...")

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
	defer cancel()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		metadata.ProduceFileMetadata(ctx)
	}()

	wg.Wait()
	fmt.Println("Shutting down producer...")
}
