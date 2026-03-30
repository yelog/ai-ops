package main

import (
	"fmt"
	"log"

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

	router := api.NewRouter()
	address := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Server starting on %s", address)
	if err := router.Run(address); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
