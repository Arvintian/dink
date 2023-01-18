package server

import (
	"context"
	"dink/pkg/k8s"
	"dink/pkg/server/handlers"

	"github.com/gin-gonic/gin"
)

func CreateHTTPRouter(ctx context.Context, client k8s.Interface) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	containers := router.Group("/containers")
	{
		containers.POST("/:namespace/:name", handlers.CreateContainer)
		containers.PUT("/:namespace/:name/start", handlers.StartContainer)
		containers.PUT("/:namespace/:name/stop", handlers.StopContainer)
		containers.DELETE("/:namespace/:name", handlers.RemoveContainer)
	}
	return router
}
