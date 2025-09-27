package main

import (
	"fmt"
	"sync"
	"time"
)

func process(item string, wg *sync.WaitGroup) {
	for i := 1; i <= 5; i++ {
		// time.Sleep(time.Second / 2)
		defer wg.Done()
		fmt.Println("Processed", i, item)
		// wg.Add(1)
	}
}

func main() {
	var wg sync.WaitGroup
	wg.Add(5)

	go process("order", &wg)

	wg.Wait()

	ch := make(chan int)

	i := 0
	go func() {
		for {
			time.Sleep(1 * time.Second)
			ch <- i
			i++
		}
	}()

	for {
		select {
		case a := <-ch:
			fmt.Println("a: ", a)
		}
	}

}
