package tasks

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func BadRequest(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, gin.H{
		"error":   "invalid_request",
		"message": err.Error(),
	})
}

func InternalError(c *gin.Context, err error) {
	c.Error(err)

	c.JSON(http.StatusInternalServerError, gin.H{
		"error": "internal_error",
	})
}
