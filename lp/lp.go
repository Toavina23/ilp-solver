package lp

import (
	"log"
	"os"
	"strconv"
	"strings"
)

type LinearProblem struct {
	ObjectiveFunction []float64
	Constraints       [][]float64
	ConstraintTypes   []string
	Rhs               []float64
	IsMaximization    bool
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
	for i, constraintType := range constraintTypes {
		if constraintType == "=" {
			// adding >=
			constraints = append(constraints, constraints[i])
			rhs = append(rhs, rhs[i])
			constraintTypes = append(constraintTypes, ">=")
			// adding <=
			constraints = append(constraints, constraints[i])
			rhs = append(rhs, rhs[i])
			constraintTypes = append(constraintTypes, "<=")
			// removing
			constraints = append(constraints[:i], constraints[i+1:]...)
			rhs = append(rhs[:i], rhs[i+1:]...)
			constraintTypes = append(constraintTypes[:i], constraintTypes[i+1:]...)
		}
	}
	return &LinearProblem{
		ObjectiveFunction: objectiveFunction,
		IsMaximization:    isMaximization,
		Constraints:       constraints,
		ConstraintTypes:   constraintTypes,
		Rhs:               rhs,
	}
}
