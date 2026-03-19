package main

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	apphttp "pierodesenzi/goline/internal/http"
	"pierodesenzi/goline/internal/tasks"
)

func main() {
	// Initialize Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery(), apphttp.ErrorLogger())

	// Inject dependency
	taskService := tasks.NewService(rdb)
	taskHandler := tasks.NewHandler(taskService)

	apphttp.RegisterRoutes(r, taskHandler)

	r.Run(":8080")
}