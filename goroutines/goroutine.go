package main

import (
	"fmt"
	"time"
)

func process(item string) {
	for i := 1; i <= 5; i++ {
		time.Sleep(time.Second / 2)
		fmt.Println("Processed", i, item)
	}
}

func main() {
	process("order")
}
