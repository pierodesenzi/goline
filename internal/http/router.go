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

	tasksGroup := api.Group("/queue")
	tasksGroup.POST("/", taskHandler.Create)  // create queue
	tasksGroup.POST("/task", taskHandler.Enqueue)  // create task
}
