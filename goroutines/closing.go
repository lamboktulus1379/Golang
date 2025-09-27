package main

import (
	"fmt"
	"sync"
)

func worker(id int, ch chan int, wg *sync.WaitGroup) {
	defer wg.Done()
	for i := 0; i < 5; i++ {
		ch <- i * id
	}
}

func main() {
	ch := make(chan int)
	var wg sync.WaitGroup

	// Start multiple worker goroutines
	for i := 1; i <= 3; i++ {
		wg.Add(1)
		go worker(i, ch, &wg)
	}

	// Close channel in a separate goroutine after workers finish
	go func() {
		wg.Wait()
		close(ch)
	}()

	// Receive values from the channel
	for val := range ch {
		fmt.Println(val)
	}
}
