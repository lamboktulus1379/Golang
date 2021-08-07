package main

import (
	"fmt"
	"math"
)

type ErrNegativeSqrt float64

func (e ErrNegativeSqrt) Error() string {
	return fmt.Sprint("cannot Sqrt negative number: ", float64(e))
}

func Sqrt(x float64) (float64, error) {
	errNeg := ErrNegativeSqrt(x)

	if x < 0 {
		return x, errNeg
	}
	return math.Sqrt(x), nil
}

func main() {
	fmt.Println(Sqrt(2))
	fmt.Println(Sqrt(-2))
}
