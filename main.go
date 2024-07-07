package main

import (
	"fmt"
	"log"
	"pnle/lp"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.Static("/assets", "./static")
	r.GET("/", func(ctx *gin.Context) {
		ctx.File("./static/index.html")
	})
	r.POST("/solve", func(ctx *gin.Context) {
		var requestBody struct {
			ProblemString string `json:"problemString" binding:"required"`
		}
		if err := ctx.Copy().ShouldBindJSON(&requestBody); err != nil {
			ctx.JSON(400, gin.H{
				"error": "Required parameter problemString not found",
			})
			return
		}
		problem := lp.CreateIntegerLinearProblem(requestBody.ProblemString)
		solution := problem.Solve()
		headers := make([]string, len(solution.ObjectiveFunction))
		tableau := make([][]float64, len(solution.Constraints))
		for i := 0; i < len(solution.ObjectiveFunction); i++ {
			if i < solution.InitialObjectiveLength {
				headers[i] = fmt.Sprintf("x%v", i+1)
			} else if i >= solution.InitialObjectiveLength && i < solution.InitialObjectiveLength+solution.SurplusVar {
				headers[i] = fmt.Sprintf("s%v", i+1-(solution.SurplusVar-1))
			} else {
				headers[i] = fmt.Sprintf("a%v", i+1-(solution.ArtificialVars-1))
			}
		}
		for i, constraint := range solution.Constraints {
			tableau[i] = constraint
			tableau[i] = append(tableau[i], solution.Rhs[i])
		}
		var baseVariables []string
		for _, baseIndex := range solution.BaseVariable {
			if baseIndex < solution.InitialObjectiveLength {
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
		tableau = append(tableau, solution.ObjectiveFunction)
		tableau[len(tableau)-1] = append(tableau[len(tableau)-1], solution.Rhs[len(solution.Rhs)-1])
		ctx.JSON(200, gin.H{
			"baseVariables":                 baseVariables,
			"headers":                       headers,
			"tableau":                       tableau,
			"optimalVariableValue":          solution.OptimalVariableValues,
			"optimalObjectiveFunctionValue": solution.OptimalObjectiveFunctionValue,
		})
	})
	err := r.Run()
	if err != nil {
		log.Fatal(err)
	}
}
