package main

import (
	"fmt"
	"pnle/lp"
)

func main() {
	fmt.Println("hello plne")
	problem := lp.LoadProblemFromFile("file3.txt")
	fmt.Println("Initial problem")
	problem.DisplaySimplexTableau()
	problem.TwoPhasedSimplexAlgorithm()
}
