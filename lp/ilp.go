package lp

import (
	"fmt"
	"log"
	"math"
	"pnle/utils"
)

type IntegerLineaProblem struct {
	InitialProblem                LinearProblem
	HasSolution                   bool
	OptimalVariableValues         []float64
	OptimalObjectiveFunctionValue float64
}

func isInteger(f float64) bool {
	residual := math.Mod(f, 1.0)
	return residual == 0
}
func LoadIntegerLinearProblemFromFile(filename string) *IntegerLineaProblem {
	problem := LoadProblemFromFile(filename)
	return &IntegerLineaProblem{
		InitialProblem: *problem,
	}
}
func CreateIntegerLinearProblem(content string) *IntegerLineaProblem {
	problem := CreateProblem(content)
	return &IntegerLineaProblem{
		InitialProblem: *problem,
	}
}

func (ilp *IntegerLineaProblem) Solve() *LinearProblem {
	problemQueue := utils.NewQueue[LinearProblem]()
	problemQueue.Enqueue(ilp.InitialProblem)
	iteration := 1
	for {
		fmt.Printf("Iteration no: %v\n", iteration)
		currentProblem, containsElement := problemQueue.Dequeue()
		if !containsElement {
			break
		}
		solution := currentProblem.Solve()
		if solution == nil && iteration == 1 {
			log.Fatal("This integer linear problem have no solution")
		}
		if solution != nil {
			if isInteger(solution.OptimalObjectiveFunctionValue) {
				fmt.Println("Integer solution found")
				for i, value := range solution.OptimalVariableValues {
					fmt.Printf("x%v=%v\n", i+1, value)
				}
				fmt.Printf("Z=%v\n", solution.OptimalObjectiveFunctionValue)
				return solution
			} else {
				for i, boundFloatValue := range solution.OptimalVariableValues {
					integerBound := int64(math.Floor(boundFloatValue))
					// construct the constraint
					boundConstraint := make([]float64, len(ilp.InitialProblem.ObjectiveFunction))
					boundConstraint[i] = 1
					// setup the lower bound value
					lowerBoundProblem := ilp.InitialProblem.Clone()
					lowerBoundProblem.Constraints = append(lowerBoundProblem.Constraints, boundConstraint)
					lowerBoundProblem.ConstraintTypes = append(lowerBoundProblem.ConstraintTypes, "<=")
					lowerBoundProblem.Rhs = append(lowerBoundProblem.Rhs[:len(lowerBoundProblem.Rhs)-1],
						float64(integerBound),
						lowerBoundProblem.Rhs[len(lowerBoundProblem.Rhs)-1])
					lowerBoundProblem.InitialConstraintLength += 1
					problemQueue.Enqueue(*lowerBoundProblem)
					// setup the upper bound value
					upperBoundProblem := ilp.InitialProblem.Clone()
					upperBoundProblem.Constraints = append(upperBoundProblem.Constraints, boundConstraint)
					upperBoundProblem.ConstraintTypes = append(upperBoundProblem.ConstraintTypes, ">=")
					upperBoundProblem.Rhs = append(upperBoundProblem.Rhs[:len(upperBoundProblem.Rhs)-1],
						float64(integerBound+1),
						upperBoundProblem.Rhs[len(upperBoundProblem.Rhs)-1])
					upperBoundProblem.InitialConstraintLength += 1
					problemQueue.Enqueue(*upperBoundProblem)
				}
			}
		}
		iteration++
	}
	return nil
}
