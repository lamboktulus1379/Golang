package main

import (
	"fmt"
	"sync"
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

}
