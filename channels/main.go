package main

import (
	"fmt"
)

func timesThree(arr []int, ch chan int) {
	minusCh := make(chan int, 3)
	for _, elem := range arr {
		value := elem * 3
		if value%2 == 0 {
			go minusThree(value, minusCh)
			value = <-minusCh
		}
		ch <- value
	}
}

func minusThree(number int, ch chan int) {
	ch <- number - 3
	fmt.Println("The functions continues after returning the result")
}

func main() {
	fmt.Println("We are executing a goroutine")
	arr := []int{2, 3, 4}
	ch := make(chan int, len(arr))
	go timesThree(arr, ch)
	for i := 0; i < len(arr); i++ {
		fmt.Printf("Result: %v \n", <-ch)
	}
}
