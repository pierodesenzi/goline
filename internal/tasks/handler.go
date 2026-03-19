package tasks

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(s *Service) *Handler {
	return &Handler{service: s}
}

type CreateQueueRequest struct {
	Name string `json:"name" binding:"required"`
}

type CreateTaskRequest struct {
	Queue string `json:"queue" binding:"required"`
	Params map[string]interface{} `json:"params" binding:"required"`
}

func (h *Handler) Enqueue(c *gin.Context) {
	var req CreateTaskRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err)
		return
	}

	task, err := h.service.Enqueue(req.Queue, req.Params)
	if err != nil {
		InternalError(c, err)
		return
	}

	status, ok := task["status"].(string)
	if !ok {
		c.JSON(400, gin.H{"error": "status must be a string"})
    	return
	}

	if task["status"] != "enqueued" {
		QueueDoesNotExist(c, status)
		return
	}

	c.JSON(http.StatusCreated, task)
	
}
func (h *Handler) Create(c *gin.Context) {
	var req CreateQueueRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err)
		return
	}

	task, err := h.service.Create(req.Name)
	if err != nil {
		InternalError(c, err)
		return
	}

	c.JSON(http.StatusCreated, task)
}
