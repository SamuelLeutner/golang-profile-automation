package api

import (
	"net/http"

	profile "github.com/SamuelLeutner/golang-profile-automation/internal/services"
	"github.com/gin-gonic/gin"
)

func HandlePing(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}

func HandleProfile(c *gin.Context) {
	p, err := profile.CreatePerfil(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, p)
}
