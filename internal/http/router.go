package http

import (
	"net/http"

	"pierodesenzi/goline/internal/tasks"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, taskHandler *tasks.Handler) {
	api := r.Group("/api")

	// healthcheck
	api.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	tasksGroup := api.Group("/tasks")
	{
		tasksGroup.POST("", taskHandler.Create)
	}
}
