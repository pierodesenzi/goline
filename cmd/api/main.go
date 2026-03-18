package main

import (
	"github.com/gin-gonic/gin"

	apphttp "pierodesenzi/goline/internal/http"
	"pierodesenzi/goline/internal/tasks"
)

func main() {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery(), apphttp.ErrorLogger())

	taskHandler := tasks.NewHandler(tasks.NewService())

	apphttp.RegisterRoutes(r, taskHandler)

	r.Run(":8080")
}
