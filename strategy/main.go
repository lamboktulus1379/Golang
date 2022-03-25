package main

import "fmt"

type Strategy interface {
	Execute(a int, b int) int
}

type AddStrategy struct{}

func (add AddStrategy) Execute(a int, b int) int {
	return a + b
}

type MultiplyStrategy struct{}

func (mult MultiplyStrategy) Execute(a int, b int) int {
	return a * b
}

type RealContext struct {
	Strategy Strategy
}

func (r *RealContext) SetStrategy(strategy Strategy) {
	r.Strategy = strategy
}

func (r *RealContext) ExecuteStrategy(a int, b int) int {
	return r.Strategy.Execute(a, b)
}

func main() {
	r := RealContext{AddStrategy{}}
	result := r.ExecuteStrategy(4, 3)

	fmt.Println("Result add: ", result)

	m := RealContext{MultiplyStrategy{}}
	mResult := m.ExecuteStrategy(4, 3)

	fmt.Println("Result multiply: ", mResult)
}
