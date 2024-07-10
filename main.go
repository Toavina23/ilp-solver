package main

import (
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
		ctx.JSON(200, gin.H{
			"solutionProblemString": solution.CreateMarkdownExpression(),
			"tableaux":              solution.SolutionSteps,
			"solutionString":        solution.CreateSolutionMarkdownExpression(),
		})
	})
	err := r.Run()
	if err != nil {
		log.Fatal(err)
	}
}
