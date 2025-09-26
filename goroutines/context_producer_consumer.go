package main

import (
	"context"
	"fmt"
	"time"
)

func producer(ctx context.Context, ch chan int) {
	defer close(ch) // Close the channel when the producer finishes
	for i := 0; ; i++ {
		select {
		case <-ctx.Done(): // Stop producing if context is canceled
			fmt.Println("Producer stopping...")
			return
		case ch <- i:
			time.Sleep(500 * time.Millisecond) // Simulate work
		}
	}
}

func main() {
	ch := make(chan int)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	go producer(ctx, ch)

	for val := range ch {
		fmt.Println("Received:", val)
	}
	fmt.Println("Main completed")
}