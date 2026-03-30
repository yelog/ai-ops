package api

import (
	"github.com/gin-gonic/gin"
	"github.com/your-org/ai-k8s-ops/internal/api/handlers"
)

func NewRouter() *gin.Engine {
	router := gin.Default()

	v1 := router.Group("/api/v1")
	{
		system := v1.Group("/system")
		{
			system.GET("/health", handlers.HealthCheck)
			system.GET("/version", handlers.GetVersion)
		}
	}

	return router
}
