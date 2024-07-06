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
		baseVariables := make([]string, len(solution.Constraints))
		baseVariables = append(baseVariables, "F")
		for _, baseIndex := range solution.BaseVariable {
			if baseIndex < solution.InitialObjectiveLength {
				baseVariables = append(baseVariables, fmt.Sprintf("x%v", baseIndex+1))
			} else {
				baseVariables = append(baseVariables, fmt.Sprintf("s%v", baseIndex+1))
			}
		}
		baseVariables = append(baseVariables, "Z")
		headers := make([]string, len(solution.ObjectiveFunction))
		tableau := make([][]float64, len(solution.Constraints))
		for i, constraint := range solution.Constraints {
			if i < solution.InitialObjectiveLength {
				headers = append(headers, fmt.Sprintf("x%v", i))
			} else if i >= solution.InitialObjectiveLength && i < solution.InitialObjectiveLength+solution.SurplusVar {
				headers = append(headers, fmt.Sprintf("s%v", i))
			} else {
				headers = append(headers, fmt.Sprintf("a%v", i))
			}
			tableau = append(tableau, constraint)
			tableau[i] = append(tableau[i], solution.Rhs[i])
		}
		tableau = append(tableau, solution.ObjectiveFunction)
		ctx.JSON(200, gin.H{
			"baseVariables": baseVariables,
			"headers":       headers,
			"tableau":       tableau,
		})
	})
	err := r.Run()
	if err != nil {
		log.Fatal(err)
	}
}
