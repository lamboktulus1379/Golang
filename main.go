package main

import (
	"errors"
	"fmt"
	"runtime"
	"time"
)

var t time.Time = time.Now()

func do(job string) error {
	fmt.Println("doing job", job, t.Second())

	time.Sleep(5 * time.Second)

	return errors.New("something went wrong!")
}

func main() {
	fmt.Println("Hello world!")

	jobs := []string{"one", "two", "three"}
	errc := make(chan error)

	for _, job := range jobs {

		go func(job string) {
			t = t.Add(1 * time.Second)
			fmt.Println("Second: ", t.Second(), runtime.NumGoroutine())
			errc <- do(job)
		}(job)
	}

	for _ = range jobs {
		if err := <-errc; err != nil {
			fmt.Println(err)
		}
	}

	fmt.Println("Goroutine Number: ", runtime.NumGoroutine())
}
