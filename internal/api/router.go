package api

import (
	"database/sql"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/your-org/ai-k8s-ops/internal/ai"
	"github.com/your-org/ai-k8s-ops/internal/api/handlers"
	"github.com/your-org/ai-k8s-ops/internal/api/middleware"
	"github.com/your-org/ai-k8s-ops/internal/auth"
	"github.com/your-org/ai-k8s-ops/internal/cluster"
	"github.com/your-org/ai-k8s-ops/internal/deploy"
	"github.com/your-org/ai-k8s-ops/internal/llm"
	"github.com/your-org/ai-k8s-ops/internal/offline"
)

func NewRouter() *gin.Engine {
	return NewRouterWithDB(nil, "dev-secret-key", 24*time.Hour, nil)
}

func NewRouterWithDB(db *sql.DB, jwtSecret string, jwtExpiry time.Duration, aiConfig *AIConfig) *gin.Engine {
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

			offlineDB := offline.NewPackageDB(db)
			offlineHandler := handlers.NewOfflineHandler(
				offlineDB,
				offline.NewExporter(offlineDB, "data/offline"),
				offline.NewImporter(offlineDB, "data/offline"),
			)

			offlineGroup := v1.Group("/offline")
			offlineGroup.Use(middleware.AuthMiddleware(jwtSecret))
			{
				offlineGroup.GET("/resources", offlineHandler.GetResources)
				offlineGroup.POST("/packages/export", offlineHandler.ExportPackage)
				offlineGroup.POST("/packages/import", offlineHandler.ImportPackage)
				offlineGroup.GET("/packages", offlineHandler.ListPackages)
				offlineGroup.GET("/packages/:id", offlineHandler.GetPackage)
				offlineGroup.DELETE("/packages/:id", offlineHandler.DeletePackage)
				offlineGroup.GET("/packages/:id/download", offlineHandler.DownloadPackage)
			}

			if aiConfig != nil && aiConfig.APIKey != "" {
				aiHandler := handlers.NewAIHandler(
					ai.NewConversationDB(db),
					ai.NewMessageDB(db),
					llm.NewClient(aiConfig.APIKey, aiConfig.Model),
				)

				aiGroup := v1.Group("/ai")
				aiGroup.Use(middleware.AuthMiddleware(jwtSecret))
				{
					aiGroup.POST("/conversations", aiHandler.CreateConversation)
					aiGroup.GET("/conversations", aiHandler.ListConversations)
					aiGroup.GET("/conversations/:id", aiHandler.GetConversation)
					aiGroup.DELETE("/conversations/:id", aiHandler.DeleteConversation)
					aiGroup.POST("/chat", aiHandler.Chat)
				}
			}
		}
	}

	return router
}

type AIConfig struct {
	APIKey string
	Model  string
}
