package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/your-org/ai-k8s-ops/pkg/version"
)

func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
	})
}

func GetVersion(c *gin.Context) {
	c.JSON(http.StatusOK, version.GetInfo())
}
