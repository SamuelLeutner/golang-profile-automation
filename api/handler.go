package api

import (
	"net/http"

	p "github.com/SamuelLeutner/golang-profile-automation/internal/services"
	"github.com/gin-gonic/gin"
)

func HandlePing(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}

func HandleProfile(c *gin.Context) {
	err := p.CreateProfile(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}
