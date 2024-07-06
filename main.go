package main

import (
	"fmt"
	"pnle/lp"
)

func main() {
	fmt.Println("hello plne")
	problem := lp.LoadIntegerLinearProblemFromFile("file4.txt")
	problem.Solve()
}
