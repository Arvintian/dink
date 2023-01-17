package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateContainer(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{})
}
