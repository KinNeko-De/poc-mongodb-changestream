package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("Starting producer...")

	time.Sleep(20 * time.Second)

	fmt.Println("Shutting down producer...")
}
