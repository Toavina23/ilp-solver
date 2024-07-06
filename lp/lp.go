package lp

import (
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
)

type LinearProblem struct {
	ObjectiveFunction             []float64
	Constraints                   [][]float64
	ConstraintTypes               []string
	Rhs                           []float64
	IsMaximization                bool
	SurplusVar                    int
	ArtificialVars                int
	InitialConstraintLength       int
	InitialObjectiveLength        int
	BaseVariable                  []int
	OriginalProblem               *LinearProblem
	Phase1Problem                 *LinearProblem
	OptimalVariableValues         []float64
	OptimalObjectiveFunctionValue float64
}

func (lp *LinearProblem) TwoPhasedSimplexAlgorithm() {
	fmt.Println("Starting Two-Phased Simplex Algorithm")

	feasibleSolution := lp.addConstraintVariables()
	fmt.Println("Initial Tableau for Phase 1:")
	feasibleSolution.DisplaySimplexTableau()

	// Phase 1
	phase1Solution := feasibleSolution.Phase1()
	if phase1Solution == nil {
		fmt.Println("No feasible solution found in Phase 1.")
		return
	}
	fmt.Println("Phase 1 Complete. Feasible solution found")

	// Phase 2
	optimalSolution := phase1Solution.Phase2()
	if optimalSolution == nil {
		fmt.Println("No optimal solution found in Phase 2.")
		return
	}
	fmt.Println("Optimal Solution:")
	optimalSolution.DisplaySimplexTableau()
	optimalSolution.SaveSolution()
}

func (lp *LinearProblem) SaveSolution() {
	for i := 0; i < len(lp.OriginalProblem.ObjectiveFunction); i++ {
		optimalVariableValue := 0.0
		for j, variableIndex := range lp.BaseVariable {
			// this means that xi is in the base variable, so it have solution different of 0
			if variableIndex == i {
				optimalVariableValue = lp.Rhs[j]
				break
			}
		}
		lp.OptimalVariableValues = append(lp.OptimalVariableValues, optimalVariableValue)
		fmt.Printf("x%v=%v\n", i+1, optimalVariableValue)
	}
	fmt.Printf("Z=%v\n", lp.Rhs[len(lp.Rhs)-1])
	lp.OptimalObjectiveFunctionValue = lp.Rhs[len(lp.Rhs)-1]
}
func (lp *LinearProblem) Phase1() *LinearProblem {
	for {
		pivotColumn := lp.findPivotColumn()
		if pivotColumn == -1 {
			// check if the optimal value is 0,
			// in that case the principal problem have a solution
			if lp.Rhs[len(lp.Rhs)-1] == 0 {
				return lp
			}
			return nil
		}
		pivotRow := lp.findPivotRow(pivotColumn)
		if pivotRow == -1 {
			log.Fatal("Unbounded solution")
			return nil // Unbounded solution
		}
		lp.BaseVariable[pivotRow] = pivotColumn
		lp.pivot(pivotRow, pivotColumn)
		lp.DisplaySimplexTableau()
	}
}

func (lp *LinearProblem) Phase2() *LinearProblem {
	// Remove artificial variables and reset objective function
	lp.removeArtificialVariables()
	for {
		pivotColumn := lp.findPivotColumn()
		if pivotColumn == -1 {
			return lp // Optimal solution found
		}

		pivotRow := lp.findPivotRow(pivotColumn)
		if pivotRow == -1 {
			log.Fatal("Unbounded solution")
			return nil // Unbounded solution
		}
		lp.BaseVariable[pivotRow] = pivotColumn
		lp.pivot(pivotRow, pivotColumn)

	}
}

func (lp *LinearProblem) findPivotColumn() int {
	pivotColumn := -1
	if lp.IsMaximization {
		minCoeff := 0.0
		for j := 0; j < len(lp.ObjectiveFunction); j++ {
			if lp.ObjectiveFunction[j] > minCoeff {
				minCoeff = lp.ObjectiveFunction[j]
				pivotColumn = j
			}
		}
	} else {
		maxCoeff := 0.0
		for j := 0; j < len(lp.ObjectiveFunction); j++ {
			if lp.ObjectiveFunction[j] < maxCoeff {
				maxCoeff = lp.ObjectiveFunction[j]
				pivotColumn = j
			}
		}
	}
	return pivotColumn
}

func (lp *LinearProblem) findPivotRow(pivotColumn int) int {
	minRatio := math.Inf(1)
	pivotRow := -1
	for i := 0; i < lp.InitialConstraintLength; i++ {
		if lp.Constraints[i][pivotColumn] > 0 {
			ratio := lp.Rhs[i] / lp.Constraints[i][pivotColumn]
			if ratio < minRatio {
				minRatio = ratio
				pivotRow = i
			}
		}
	}
	return pivotRow
}

func (lp *LinearProblem) pivot(pivotRow, pivotColumn int) {
	pivotElement := lp.Constraints[pivotRow][pivotColumn]

	// Update pivot row
	for j := 0; j < len(lp.Constraints[pivotRow]); j++ {
		lp.Constraints[pivotRow][j] /= pivotElement
	}
	lp.Rhs[pivotRow] /= pivotElement

	// Update other rows
	for i := 0; i < len(lp.Constraints); i++ {
		if i != pivotRow {
			factor := lp.Constraints[i][pivotColumn]
			for j := 0; j < len(lp.Constraints[i]); j++ {
				lp.Constraints[i][j] -= factor * lp.Constraints[pivotRow][j]
			}
			lp.Rhs[i] -= factor * lp.Rhs[pivotRow]
		}
	}

	// Update objective function
	factor := lp.ObjectiveFunction[pivotColumn]
	for j := 0; j < len(lp.ObjectiveFunction); j++ {
		lp.ObjectiveFunction[j] -= factor * lp.Constraints[pivotRow][j]
	}
	lp.Rhs[len(lp.Rhs)-1] -= factor * lp.Rhs[pivotRow]
}

func (lp *LinearProblem) removeArtificialVariables() {
	newVarCount := lp.InitialObjectiveLength + lp.SurplusVar
	lp.ObjectiveFunction = lp.ObjectiveFunction[:newVarCount]
	for i := range lp.Constraints {
		lp.Constraints[i] = lp.Constraints[i][:newVarCount]
	}
	lp.ArtificialVars = 0
	// recompute the objective function
	objectiveFunctionRhsValue := 0.0
	newObjectiveFunction := make([]float64, len(lp.ObjectiveFunction))
	copy(newObjectiveFunction, lp.OriginalProblem.ObjectiveFunction)
	// this will maintain the problem as a minimization problem
	for i := range newObjectiveFunction {
		newObjectiveFunction[i] *= -1
	}
	coeffs := make([]float64, len(lp.ObjectiveFunction))
	copy(coeffs, lp.OriginalProblem.ObjectiveFunction)
	for i, constraint := range lp.Constraints {
		for j, value := range constraint {
			newObjectiveFunction[j] += value * coeffs[lp.BaseVariable[i]]
		}
		objectiveFunctionRhsValue += lp.Rhs[i] * coeffs[lp.BaseVariable[i]]
	}
	lp.ObjectiveFunction = newObjectiveFunction
	lp.Rhs[len(lp.Rhs)-1] = objectiveFunctionRhsValue
}

func (lp *LinearProblem) addConstraintVariables() *LinearProblem {
	n := len(lp.ObjectiveFunction)
	m := len(lp.Constraints)
	// add the variables
	additionalVars := 0
	surplusVars := 0
	artificialVars := 0
	for _, constraintType := range lp.ConstraintTypes {
		if constraintType == "<=" {
			additionalVars += 1
			surplusVars++
		} else if constraintType == ">=" {
			additionalVars += 2
			surplusVars++
			artificialVars++
		} else if constraintType == "=" {
			additionalVars += 1
			artificialVars++
		}
	}
	coeffs := make([]float64, n+additionalVars)
	baseVariables := make([]int, m)
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
			baseVariables[i] = slackIndex
			slackIndex++
			newConstraintTypes[i] = "="
		case ">=":
			newConstraints[i][slackIndex] = -1
			newConstraints[i][artificialIndex] = 1
			baseVariables[i] = artificialIndex
			coeffs[artificialIndex] = -1
			slackIndex++
			artificialIndex++
			newConstraintTypes[i] = "="
		case "=":
			newConstraints[i][artificialIndex] = 1
			baseVariables[i] = artificialIndex
			coeffs[artificialIndex] = -1
			artificialIndex++
			newConstraintTypes[i] = "="
		}
	}
	// compute the objective function
	newObjectiveFunction := make([]float64, n+additionalVars)
	newRhs := make([]float64, m+1)
	copy(newRhs, lp.Rhs)
	objectiveFunctionRhsValue := 0.0
	for i, constraint := range newConstraints {
		for j, value := range constraint {
			newObjectiveFunction[j] += value * coeffs[baseVariables[i]]
		}
		objectiveFunctionRhsValue += newRhs[i] * coeffs[baseVariables[i]]
	}
	// resetting artificial var index to use it in the objective function computation
	if artificialVars > 0 {
		newRhs[m] = objectiveFunctionRhsValue
	} else {
		newRhs[m] = 0
	}
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
		BaseVariable:            baseVariables,
		OriginalProblem:         lp,
	}
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
			fmt.Printf("%-8.2f", coef)
		} else {
			fmt.Printf("%-8.2f", coef)
		}
	}
	fmt.Printf("%-8.2f\n", lp.Rhs[len(lp.Rhs)-1])
}
