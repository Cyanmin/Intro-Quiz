package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HelloHandler returns a greeting message.
// @Summary      Say hello
// @Description  Responds with a simple greeting.
// @Tags         example
// @Produce      json
// @Success      200 {object} map[string]string
// @Router       /api/hello [get]
func HelloHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "hello"})
}
