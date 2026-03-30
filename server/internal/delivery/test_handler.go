package delivery

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TestMessage struct {
	id      uuid.UUID `json:"id"`
	message string    `json:"message"`
	status  int       `json:"status"`
}

func (h *Handler) Ping(c *gin.Context) {
	c.JSON(200, gin.H{"message": "pong"})
}
