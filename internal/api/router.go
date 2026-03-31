package api

import (
	"database/sql"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/your-org/ai-k8s-ops/internal/api/handlers"
	"github.com/your-org/ai-k8s-ops/internal/api/middleware"
	"github.com/your-org/ai-k8s-ops/internal/auth"
	"github.com/your-org/ai-k8s-ops/internal/cluster"
	"github.com/your-org/ai-k8s-ops/internal/deploy"
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

			clusterDB := cluster.NewClusterDB(db)
			clusterHandler := handlers.NewClusterHandler(clusterDB)

			clusterGroup := v1.Group("/clusters")
			clusterGroup.Use(middleware.AuthMiddleware(jwtSecret))
			{
				clusterGroup.POST("", clusterHandler.CreateCluster)
				clusterGroup.GET("", clusterHandler.ListClusters)
				clusterGroup.GET("/:id", clusterHandler.GetCluster)
				clusterGroup.PUT("/:id", clusterHandler.UpdateCluster)
				clusterGroup.DELETE("/:id", clusterHandler.DeleteCluster)
			}

			deployHandler := handlers.NewDeployHandler(
				deploy.NewTemplateDB(db),
				deploy.NewTaskDB(db),
			)

			templateGroup := v1.Group("/deploy/templates")
			templateGroup.Use(middleware.AuthMiddleware(jwtSecret))
			{
				templateGroup.POST("", deployHandler.CreateTemplate)
				templateGroup.GET("", deployHandler.ListTemplates)
				templateGroup.GET("/:id", deployHandler.GetTemplate)
				templateGroup.PUT("/:id", deployHandler.UpdateTemplate)
				templateGroup.DELETE("/:id", deployHandler.DeleteTemplate)
			}

			taskGroup := v1.Group("/deploy/tasks")
			taskGroup.Use(middleware.AuthMiddleware(jwtSecret))
			{
				taskGroup.POST("", deployHandler.CreateTask)
				taskGroup.GET("", deployHandler.ListTasks)
				taskGroup.GET("/:id", deployHandler.GetTask)
			}
		}
	}

	return router
}
