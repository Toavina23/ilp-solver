package lp

import (
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
)

type SimplexTableau struct {
	Phase         int8        `json:"phase"`
	Iteration     int32       `json:"iteration"`
	BaseVariables []string    `json:"baseVariables"`
	Headers       []string    `json:"headers"`
	Tableau       [][]float64 `json:"tableau"`
}
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
	HasSolution                   bool
	SolutionSteps                 []*SimplexTableau
}

func (lp *LinearProblem) CreateSolutionMarkdownExpression() string {
	var sb strings.Builder

	sb.WriteString("```math\n")
	for i, value := range lp.OptimalVariableValues {
		if value == 0 {
			continue
		}
		if i == len(lp.OptimalVariableValues)-1 {
			if value == 1 {
				sb.WriteString(fmt.Sprintf("x_%v = ", i+1))
			} else {
				sb.WriteString(fmt.Sprintf("%8.2fx_%v = ", value, i+1))
			}
		} else {
			if value == 1 {
				sb.WriteString(fmt.Sprintf("x_%v + ", i+1))
			} else {
				sb.WriteString(fmt.Sprintf("%8.2fx_%v + ", value, i+1))
			}
		}
	}
	sb.WriteString(fmt.Sprintf("%8.2f\n", lp.OptimalObjectiveFunctionValue))
	sb.WriteString("```")
	return sb.String()
}
func (lp *LinearProblem) CreateMarkdownExpression() string {
	var sb strings.Builder
	problemType := "Minimize"
	if lp.IsMaximization {
		problemType = "Maximize"
	}
	sb.WriteString("```math\n")
	sb.WriteString("\\begin{aligned}\n")
	sb.WriteString(fmt.Sprintf("&\\text{%s:}\\\\\n", problemType))
	objectiveFunction := ""
	for i, value := range lp.OriginalProblem.ObjectiveFunction {
		if value == 0 {
			continue
		}
		if i == len(lp.OriginalProblem.ObjectiveFunction)-1 {
			if value == 1 {
				objectiveFunction += fmt.Sprintf("x_%v", i+1)
			} else {
				objectiveFunction += fmt.Sprintf("%8.2fx_%v", value, i+1)
			}
		} else {
			if value == 1 {
				objectiveFunction += fmt.Sprintf("x_%v", i+1)
			} else {

				objectiveFunction += fmt.Sprintf("%8.2fx_%v + ", value, i+1)
			}
		}
	}
	sb.WriteString(fmt.Sprintf("&Z = %s \\\\[10pt]\n", objectiveFunction))
	sb.WriteString("&\\text{Subject to:} \\\\\n")
	sb.WriteString("&\\left\\{\n")
	sb.WriteString("\\begin{array}{l}\n")
	for i, constraint := range lp.OriginalProblem.Constraints {
		constraintRow := ""
		for j, value := range constraint {
			if value == 0 {
				continue
			}
			if j == len(constraint)-1 {
				if value == 1 {
					constraintRow += fmt.Sprintf("x_%v", j+1)
				} else {
					constraintRow += fmt.Sprintf("%8.2fx_%v", value, j+1)
				}
			} else {
				if value == 1 {
					constraintRow += fmt.Sprintf("x_%v + ", j+1)
				} else {

					constraintRow += fmt.Sprintf("%8.2fx_%v + ", value, j+1)
				}
			}
		}
		constraintRow = strings.Trim(constraintRow, "+ ")
		constraintType := "\\leq"
		switch lp.OriginalProblem.ConstraintTypes[i] {
		case ">=":
			constraintType = "\\geq"
		case "=":
			constraintType = "\\eq"
		}
		constraintRow += constraintType
		constraintRow += fmt.Sprintf("%8.2f", lp.OriginalProblem.Rhs[i])
		constraintRow += "\\\\\n"
		sb.WriteString(constraintRow)
	}
	sb.WriteString("\\end{array}\n")
	sb.WriteString("\\right.\n")
	sb.WriteString("\\end{aligned}\n")
	sb.WriteString("```")

	return sb.String()
}
func (lp *LinearProblem) Solve() *LinearProblem {
	fmt.Println("Starting Two-Phased Simplex Algorithm")

	feasibleSolution := lp.addConstraintVariables()
	fmt.Println("Initial Tableau for Phase 1:")
	feasibleSolution.SaveSimplexTableau(0, 0)
	feasibleSolution.DisplaySimplexTableau()

	// Phase 1
	phase1Solution := feasibleSolution.Phase1()
	if phase1Solution == nil {
		fmt.Println("No feasible solution found in Phase 1.")
		lp.HasSolution = false
		return nil
	}
	fmt.Println("Phase 1 Complete. Feasible solution found")

	// Phase 2
	optimalSolution := phase1Solution.Phase2()
	if optimalSolution == nil {
		fmt.Println("No optimal solution found in Phase 2.")
		lp.HasSolution = false
		return nil
	}
	fmt.Println("Optimal Solution:")
	optimalSolution.DisplaySimplexTableau()
	optimalSolution.SaveSolution()
	optimalSolution.HasSolution = true
	return optimalSolution
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
	iteration := 0
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
			return nil // Unbounded solution
		}
		lp.BaseVariable[pivotRow] = pivotColumn
		lp.pivot(pivotRow, pivotColumn)
		lp.SaveSimplexTableau(1, int32(iteration))
		lp.DisplaySimplexTableau()
		iteration++
	}
}

func (lp *LinearProblem) Phase2() *LinearProblem {
	// Remove artificial variables and reset objective function
	iteration := 0
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
		lp.SaveSimplexTableau(2, int32(iteration))
		lp.DisplaySimplexTableau()
		iteration++
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
func CreateProblem(problemContent string) *LinearProblem {
	problemLines := strings.Split(problemContent, "\n")
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
func LoadProblemFromFile(filename string) *LinearProblem {
	file, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal("invalid filename provided")
	}
	content := string(file)
	return CreateProblem(content)
}
func (lp *LinearProblem) DisplaySimplexTableau() {
	lastSimplexTableau := lp.SolutionSteps[len(lp.SolutionSteps)-1]
	for _, header := range lastSimplexTableau.Headers {
		fmt.Printf("%-8s", header)
	}
	fmt.Printf("\n")
	for i, row := range lastSimplexTableau.Tableau {
		fmt.Printf("%-8s", lastSimplexTableau.BaseVariables[i])
		for _, value := range row {
			fmt.Printf("%-8.5f", value)
		}
		fmt.Printf("\n")
	}
}

func (lp *LinearProblem) Clone() *LinearProblem {
	clone := &LinearProblem{
		ObjectiveFunction:       make([]float64, len(lp.ObjectiveFunction)),
		Constraints:             make([][]float64, len(lp.Constraints)),
		ConstraintTypes:         make([]string, len(lp.ConstraintTypes)),
		Rhs:                     make([]float64, len(lp.Rhs)),
		IsMaximization:          lp.IsMaximization,
		SurplusVar:              lp.SurplusVar,
		ArtificialVars:          lp.ArtificialVars,
		InitialConstraintLength: lp.InitialConstraintLength,
		InitialObjectiveLength:  lp.InitialObjectiveLength,
		BaseVariable:            make([]int, len(lp.BaseVariable)),
		HasSolution:             lp.HasSolution,
	}

	// Deep copy slice fields
	copy(clone.ObjectiveFunction, lp.ObjectiveFunction)
	copy(clone.ConstraintTypes, lp.ConstraintTypes)
	copy(clone.Rhs, lp.Rhs)
	copy(clone.BaseVariable, lp.BaseVariable)

	// Deep copy 2D slice
	for i, constraint := range lp.Constraints {
		clone.Constraints[i] = make([]float64, len(constraint))
		copy(clone.Constraints[i], constraint)
	}

	// Deep copy OptimalVariableValues
	if lp.OptimalVariableValues != nil {
		clone.OptimalVariableValues = make([]float64, len(lp.OptimalVariableValues))
		copy(clone.OptimalVariableValues, lp.OptimalVariableValues)
	}

	// Copy OptimalObjectiveFunctionValue
	clone.OptimalObjectiveFunctionValue = lp.OptimalObjectiveFunctionValue

	// Handle pointer fields
	if lp.OriginalProblem != nil {
		clone.OriginalProblem = lp.OriginalProblem.Clone()
	}
	if lp.Phase1Problem != nil {
		clone.Phase1Problem = lp.Phase1Problem.Clone()
	}

	return clone
}
func (lp *LinearProblem) SaveSimplexTableau(phase int8, iteration int32) {
	headers := make([]string, len(lp.ObjectiveFunction))
	tableau := make([][]float64, len(lp.Constraints))
	for i := 0; i < len(lp.ObjectiveFunction); i++ {
		if i < lp.InitialObjectiveLength {
			headers[i] = fmt.Sprintf("x%v", i+1)
		} else if i >= lp.InitialObjectiveLength && i < lp.InitialObjectiveLength+lp.SurplusVar {
			headers[i] = fmt.Sprintf("s%v", i+1-(lp.SurplusVar-1))
		} else {
			headers[i] = fmt.Sprintf("a%v", i+1-(lp.ArtificialVars-1))
		}
	}
	for i, constraint := range lp.Constraints {
		tableau[i] = constraint
		tableau[i] = append(tableau[i], lp.Rhs[i])
	}
	var baseVariables []string
	for _, baseIndex := range lp.BaseVariable {
		if baseIndex < lp.InitialObjectiveLength {
			baseVariables = append(baseVariables, headers[baseIndex])
		} else {
			baseVariables = append(baseVariables, headers[baseIndex])
		}
	}
	headers = append([]string{
		"F",
	}, headers...)
	headers = append(headers, []string{
		"RHS",
	}...)
	baseVariables = append(baseVariables, "Z")
	tableau = append(tableau, lp.ObjectiveFunction)
	tableau[len(tableau)-1] = append(tableau[len(tableau)-1], lp.Rhs[len(lp.Rhs)-1])
	lp.SolutionSteps = append(lp.SolutionSteps, &SimplexTableau{
		BaseVariables: baseVariables,
		Headers:       headers,
		Tableau:       tableau,
		Phase:         phase,
		Iteration:     iteration,
	})
}
