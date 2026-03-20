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
	Queue string					`json:"queue" binding:"required"`
	Function string					`json:"function" binding:"required"`
	Params map[string]interface{}	`json:"params" binding:"required"`
}


func (h *Handler) Enqueue(c *gin.Context) {
	var req CreateTaskRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err)
		return
	}

	task, err := h.service.Enqueue(req.Queue, req.Function, req.Params)
	if err != nil {
		InternalError(c, err)
		return
	}

	if task.Status == "QUEUE_DOES_NOT_EXIST" {
		QueueDoesNotExist(c, task.Status)
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

	queue, err := h.service.Create(req.Name)
	if err != nil {
		InternalError(c, err)
		return
	}

	if queue.Status == "ALREADY_EXISTS" {
		c.JSON(http.StatusConflict, queue)
		return
	}

	c.JSON(http.StatusCreated, queue)
}
