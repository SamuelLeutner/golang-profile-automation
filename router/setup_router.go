package router

import (
	h "github.com/SamuelLeutner/golang-profile-automation/api"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()
	api := r.Group("/api")

	api.GET("/ping", h.HandlePing)
	api.POST("/profile", h.HandleProfile)

	return r
}
