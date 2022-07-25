package main

import "fmt"

func Print[T any](s []T) {
	for _, v := range s {
		fmt.Print(v)
	}
}

func main() {
	fmt.Println("[INFO]")
	Print([]string{"Hello, ", "playground\n"})
	Print([]int{1, 2, 3})
}
