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

const tolerance = 1e-8

func isInteger(x float64) bool {
	return math.Abs(x-math.Round(x)) < tolerance
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
func isIntegerSolution(lp LinearProblem) bool {
	for _, value := range lp.OptimalVariableValues {
		if !isInteger(value) {
			return false
		}
	}
	return true
}
func (ilp *IntegerLineaProblem) Solve() *LinearProblem {
	problemQueue := utils.NewQueue[LinearProblem]()
	problemQueue.Enqueue(ilp.InitialProblem)
	iteration := 1
	isMaximization := ilp.InitialProblem.IsMaximization
	var bestSolution *LinearProblem
	bestValue := math.Inf(-1)
	if !ilp.InitialProblem.IsMaximization {
		bestValue = math.Inf(1)
	}
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
			//since we have negative optimal value for maximization and positive optimal value for minimization
			if isMaximization && solution.OptimalObjectiveFunctionValue*-1 < bestValue {
				continue
			} else if !isMaximization && solution.OptimalObjectiveFunctionValue*-1 > bestValue {
				continue
			}
			if isIntegerSolution(*solution) {
				bestSolution = solution
				bestValue = solution.OptimalObjectiveFunctionValue * -1
				continue
			} else {
				boundIndex := chooseBranchingVariable(solution)
				integerBound := int64(math.Floor(solution.OptimalVariableValues[boundIndex]))
				// construct the constraint
				boundConstraint := make([]float64, len(ilp.InitialProblem.ObjectiveFunction))
				boundConstraint[boundIndex] = 1
				// setup the lower bound value
				lowerBoundProblem := currentProblem.Clone()
				lowerBoundProblem.Constraints = append(lowerBoundProblem.Constraints, boundConstraint)
				lowerBoundProblem.ConstraintTypes = append(lowerBoundProblem.ConstraintTypes, "<=")
				lowerBoundProblem.Rhs = append(lowerBoundProblem.Rhs[:len(lowerBoundProblem.Rhs)-1],
					float64(integerBound),
					lowerBoundProblem.Rhs[len(lowerBoundProblem.Rhs)-1])
				lowerBoundProblem.InitialConstraintLength += 1
				problemQueue.Enqueue(*lowerBoundProblem)
				// setup the upper bound value
				upperBoundProblem := currentProblem.Clone()
				upperBoundProblem.Constraints = append(upperBoundProblem.Constraints, boundConstraint)
				upperBoundProblem.ConstraintTypes = append(upperBoundProblem.ConstraintTypes, ">=")
				upperBoundProblem.Rhs = append(upperBoundProblem.Rhs[:len(upperBoundProblem.Rhs)-1],
					float64(integerBound+1),
					upperBoundProblem.Rhs[len(upperBoundProblem.Rhs)-1])
				upperBoundProblem.InitialConstraintLength += 1
				problemQueue.Enqueue(*upperBoundProblem)
			}
		}
		iteration++
	}
	return bestSolution
}

func chooseBranchingVariable(lp *LinearProblem) int {
	bestScore := math.Inf(-1)
	var bestVarIndex int
	for i, value := range lp.OptimalVariableValues {
		if !isInteger(value) {
			score := evaluateBranchingCandidate(value)
			if score > bestScore {
				bestScore = score
				bestVarIndex = i
			}
		}
	}
	return bestVarIndex
}

func evaluateBranchingCandidate(value float64) float64 {
	fractionalPart := math.Abs(value - math.Round(value))
	return fractionalPart
}
