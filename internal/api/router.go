package api

import (
	"database/sql"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/your-org/ai-k8s-ops/internal/api/handlers"
	"github.com/your-org/ai-k8s-ops/internal/api/middleware"
	"github.com/your-org/ai-k8s-ops/internal/auth"
)

func NewRouter() *gin.Engine {
	return NewRouterWithDB(nil, "dev-secret-key", 24*time.Hour)
}

func NewRouterWithDB(db *sql.DB, jwtSecret string, jwtExpiry time.Duration) *gin.Engine {
	router := gin.Default()

	v1 := router.Group("/api/v1")
	{
		system := v1.Group("/system")
		{
			system.GET("/health", handlers.HealthCheck)
			system.GET("/version", handlers.GetVersion)
		}

		if db != nil {
			userDB := auth.NewUserDB(db)
			authHandler := handlers.NewAuthHandler(userDB, jwtSecret, jwtExpiry)

			authGroup := v1.Group("/auth")
			{
				authGroup.POST("/register", authHandler.Register)
				authGroup.POST("/login", authHandler.Login)

				protected := authGroup.Group("")
				protected.Use(middleware.AuthMiddleware(jwtSecret))
				{
					protected.GET("/profile", authHandler.GetProfile)
				}
			}
		}
	}

	return router
}
