package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"pierodesenzi/goline/internal/config"
	apphttp "pierodesenzi/goline/internal/http"
	"pierodesenzi/goline/internal/tasks"
)

func main() {
	cfg := config.Load()

	// Set Gin mode (affects logging + debug behavior)
	gin.SetMode(cfg.GinMode)

	log.Printf("Starting GoLine (mode=%s)", cfg.GinMode)

	// Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
	})
	log.Printf("Redis configured (addr=%s)", cfg.RedisAddr)

	r := gin.New()

	// Middleware stack:
	// - Logger: logs incoming HTTP requests
	// - Recovery: prevents panics from crashing the server (returns 500 instead)
	// - ErrorLogger: custom middleware for application-level error handling
	r.Use(gin.Logger(), gin.Recovery(), apphttp.ErrorLogger())

	taskService := tasks.NewService(rdb)
	taskHandler := tasks.NewHandler(taskService)

	// Register HTTP routes
	apphttp.RegisterRoutes(r, taskHandler)

	addr := ":" + cfg.HTTPPort
	log.Printf("HTTP server listening on %s", addr)

	// Start HTTP server
	if err := r.Run(addr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}