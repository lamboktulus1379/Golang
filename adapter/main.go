package main

import "fmt"

type RoundHole struct {
	Radius int
}

func (roundHole RoundHole) GetRadius() int {
	return roundHole.Radius
}

func (roundHole RoundHole) IsFits(roundPeg RoundPeg) bool {
	return roundHole.GetRadius() >= roundPeg.GetRadius()
}

type RoundPeg struct {
	Radius int
}

func (roundPeg RoundPeg) GetRadius() int {
	return roundPeg.Radius
}

type SquarePeg struct {
	Width int
}

func (squarePeg SquarePeg) GetWidth() int {
	return squarePeg.Width
}

type SquarePegAdapter struct {
	*RoundPeg
}

func main() {
	fmt.Println("Adapter Design Pattern")

	hole := RoundHole{5}
	rPeg := RoundPeg{5}

	fit := hole.IsFits(rPeg)
	fmt.Println(fit)

	smallSquarePeg := SquarePeg{5}
	largeSquarePeg := SquarePeg{10}

	// hole.IsFits(smallSquarePeg)

	smallSquarePegAdapter := SquarePegAdapter{
		RoundPeg: &RoundPeg{
			Radius: smallSquarePeg.GetWidth(),
		},
	}
	largeSqurePegAdapter := SquarePegAdapter{
		RoundPeg: &RoundPeg{
			Radius: largeSquarePeg.GetWidth(),
		},
	}

	fmt.Println(hole.IsFits(*smallSquarePegAdapter.RoundPeg))
	fmt.Println(hole.IsFits(*largeSqurePegAdapter.RoundPeg))
}
