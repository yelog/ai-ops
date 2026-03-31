package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/your-org/ai-k8s-ops/internal/api"
	"github.com/your-org/ai-k8s-ops/internal/storage/sqlite"
	"github.com/your-org/ai-k8s-ops/pkg/config"
	"github.com/your-org/ai-k8s-ops/pkg/version"
)

func main() {
	log.Printf("AI-K8S-OPS Server v%s starting...", version.Version)

	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := sqlite.Init(cfg.Database.Path)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()
	log.Printf("Database initialized at %s", cfg.Database.Path)

	if cfg.Server.Mode == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	jwtExpiry := time.Duration(cfg.Auth.JWTExpiryHours) * time.Hour

	aiConfig := &api.AIConfig{
		APIKey: cfg.AI.APIKey,
		Model:  cfg.AI.Model,
	}

	router := api.NewRouterWithDB(db, cfg.Auth.JWTSecret, jwtExpiry, aiConfig)
	address := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Server starting on %s", address)
	log.Printf("Available endpoints:")
	log.Printf("  GET  /api/v1/system/health")
	log.Printf("  GET  /api/v1/system/version")
	log.Printf("  POST /api/v1/auth/register")
	log.Printf("  POST /api/v1/auth/login")
	log.Printf("  GET  /api/v1/auth/profile")
	log.Printf("  POST /api/v1/clusters")
	log.Printf("  GET  /api/v1/clusters")
	log.Printf("  GET  /api/v1/clusters/:id")
	log.Printf("  PUT  /api/v1/clusters/:id")
	log.Printf("  DELETE /api/v1/clusters/:id")
	log.Printf("  AI endpoints (if API key configured):")
	log.Printf("  POST /api/v1/ai/conversations")
	log.Printf("  GET  /api/v1/ai/conversations")
	log.Printf("  POST /api/v1/ai/chat")
	if err := router.Run(address); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
