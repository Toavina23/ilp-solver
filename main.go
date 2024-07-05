package main

import (
	"fmt"
	"pnle/lp"
)

func main() {
	fmt.Println("hello plne")
	problem := lp.LoadProblemFromFile("file2.txt")
	fmt.Println(problem.ObjectiveFunction)
	fmt.Println(problem.IsMaximization)
	fmt.Println(problem.Constraints)
	fmt.Println(problem.ConstraintTypes)
	fmt.Println(problem.Rhs)
}
