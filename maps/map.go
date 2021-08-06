package main

import (
	"fmt"
	"strings"

	"golang.org/x/tour/wc"
)

func WordCount(s string) map[string]int {
	fmt.Println(s)
	var ss []string
	result := make(map[string]int)
	ss = strings.Fields(s)

	for i := 0; i < len(ss); i++ {
		result[ss[i]] = result[ss[i]] + 1
	}
	return result
}

func main() {
	wc.Test(WordCount)
}
