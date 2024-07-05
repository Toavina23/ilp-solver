package lp

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

type LinearProblem struct {
	ObjectiveFunction       []float64
	Constraints             [][]float64
	ConstraintTypes         []string
	Rhs                     []float64
	IsMaximization          bool
	SurplusVar              int
	ArtificialVars          int
	InitialConstraintLength int
	InitialObjectiveLength  int
}

func (lp *LinearProblem) TwoPhasedSimplexAlgorithm() {
	feasibleSolution := lp.addConstraintVariables()
	fmt.Println("Feasible solution")
	feasibleSolution.DisplaySimplexTableau()
}
func (lp *LinearProblem) addConstraintVariables() *LinearProblem {
	n := len(lp.ObjectiveFunction)
	m := len(lp.Constraints)
	// add the variables
	additionalVars := 0
	surplusVars := 0
	artificialVars := 0
	for _, constraintType := range lp.ConstraintTypes {
		if constraintType == "<=" || constraintType == "=" {
			additionalVars += 1
			surplusVars++
		} else if constraintType == ">=" {
			additionalVars += 2
			surplusVars++
			artificialVars++
		} else if constraintType == "=" {
			artificialVars++
		}
	}
	newObjectiveFunction := make([]float64, n+additionalVars)
	// copy(newObjectiveFunction, lp.ObjectiveFunction)
	newConstraints := make([][]float64, m)
	newConstraintTypes := make([]string, m)
	slackIndex := n
	artificialIndex := n + surplusVars
	for i := range newConstraints {
		newConstraints[i] = make([]float64, n+additionalVars)
		copy(newConstraints[i], lp.Constraints[i])
		switch lp.ConstraintTypes[i] {
		case "<=":
			newConstraints[i][slackIndex] = 1
			slackIndex++
			newConstraintTypes[i] = "="
		case ">=":
			newConstraints[i][slackIndex] = -1
			newConstraints[i][artificialIndex] = 1
			slackIndex++
			artificialIndex++
			newConstraintTypes[i] = "="
		case "=":
			newConstraints[i][artificialIndex] = 1
			artificialIndex++
			newConstraintTypes[i] = "="
		}
	}
	// resetting artificial var index to use it in the objective function computation
	artificialIndex = n + surplusVars
	newRhs := make([]float64, m+1)
	copy(newRhs, lp.Rhs)
	objectiveFunctionRhsValue := 0.0
	for i, newConstraint := range newConstraints {
		for j, value := range newConstraint {
			// artificial variables must have 0 value for the objective function
			// since we are minimizing the sum of the artificial variables to find
			// our feasible solution for the main problem
			if j >= artificialIndex {
				newObjectiveFunction[j] += 0
			} else {
				newObjectiveFunction[j] += value
			}
		}
		objectiveFunctionRhsValue += newRhs[i]
	}
	for i := range newObjectiveFunction {
		newObjectiveFunction[i] *= -1
	}
	newRhs[m] = objectiveFunctionRhsValue
	return &LinearProblem{
		ObjectiveFunction: newObjectiveFunction,
		Rhs:               newRhs,
		Constraints:       newConstraints,
		ConstraintTypes:   newConstraintTypes,
		//always minimization for phase 1
		IsMaximization:          false,
		ArtificialVars:          artificialVars,
		SurplusVar:              surplusVars,
		InitialConstraintLength: lp.InitialConstraintLength,
		InitialObjectiveLength:  lp.InitialObjectiveLength,
	}
}

// phase 1 is for determining the first feasible solution
func (lp *LinearProblem) Phase1() {
	// create the simplex tableau
}
func LoadProblemFromFile(filename string) *LinearProblem {
	file, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal("invalid filename provided")
	}
	content := string(file)
	problemLines := strings.Split(content, "\n")
	if len(problemLines) == 0 {
		log.Fatal("empty file provided")
	}
	objectiveLine := problemLines[0]
	problemLines = problemLines[1:]

	isMaximization := strings.HasPrefix(objectiveLine, "max")
	objectiveFunctionValues := strings.Split(objectiveLine, " ")[1:]
	var objectiveFunction []float64
	for _, value := range objectiveFunctionValues {
		floatValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			log.Fatal("Invalid float value provided in the objective function")
		}
		objectiveFunction = append(objectiveFunction, floatValue)
	}

	var constraints [][]float64
	var constraintTypes []string
	var rhs []float64
	for i, problemLine := range problemLines {
		var constraintRow []float64
		strValues := strings.Split(problemLine, " ")
		for j, value := range strValues {
			if value == ">=" || value == "<=" || value == "=" {
				constraintTypes = append(constraintTypes, value)
			} else {
				floatValue, err := strconv.ParseFloat(value, 64)
				if err != nil {
					log.Fatalf("Invalid float value provided in the constraint function %v", i)
				}
				if j == len(strValues)-1 {
					rhs = append(rhs, floatValue)
				} else {
					constraintRow = append(constraintRow, floatValue)
				}
			}
		}
		constraints = append(constraints, constraintRow)
	}
	//result of the objective function
	rhs = append(rhs, 0)
	return &LinearProblem{
		ObjectiveFunction:       objectiveFunction,
		IsMaximization:          isMaximization,
		Constraints:             constraints,
		ConstraintTypes:         constraintTypes,
		Rhs:                     rhs,
		ArtificialVars:          0,
		SurplusVar:              0,
		InitialConstraintLength: len(constraints),
		InitialObjectiveLength:  len(objectiveFunction),
	}
}
func (lp *LinearProblem) DisplaySimplexTableau() {
	fmt.Println("Simplex Tableau:")

	// Calculate the number of variables and constraints
	numOrigVars := lp.InitialObjectiveLength
	numConstraints := lp.InitialConstraintLength
	totalVars := len(lp.ObjectiveFunction)

	// Print header
	fmt.Printf("%-6s", "Basic")
	for i := 0; i < numOrigVars; i++ {
		fmt.Printf("%-8s", fmt.Sprintf("x%d", i+1))
	}
	slackSurplusCount := 0
	artificialCount := 0
	for i := numOrigVars; i < totalVars; i++ {
		if slackSurplusCount < lp.SurplusVar {
			fmt.Printf("%-8s", fmt.Sprintf("s%d", slackSurplusCount+1))
			slackSurplusCount++
		} else {
			fmt.Printf("%-8s", fmt.Sprintf("a%d", artificialCount+1))
			artificialCount++
		}
	}
	fmt.Printf("%-8s\n", "RHS")

	// Print constraints
	for i := 0; i < numConstraints; i++ {
		fmt.Printf("%-6s", fmt.Sprintf("s%d", i+1))
		for j := 0; j < totalVars; j++ {
			fmt.Printf("%-8.2f", lp.Constraints[i][j])
		}
		fmt.Printf("%-8.2f\n", lp.Rhs[i])
	}

	// Print objective function
	if lp.IsMaximization {
		fmt.Printf("%-6s", "Z")
	} else {
		fmt.Printf("%-6s", "-Z")
	}
	for _, coef := range lp.ObjectiveFunction {
		if lp.IsMaximization {
			fmt.Printf("%-8.2f", -coef)
		} else {
			fmt.Printf("%-8.2f", coef)
		}
	}
	fmt.Printf("%-8.2f\n", lp.Rhs[len(lp.Rhs)-1])
}
