package config

import (
	"testing"

	"github.com/your-org/ai-k8s-ops/pkg/config"
)

func TestLoadConfig(t *testing.T) {
	cfg, err := config.Load("../../configs/config.yaml")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test server config
	if cfg.Server.Port != 8080 {
		t.Errorf("Expected server port 8080, got %d", cfg.Server.Port)
	}

	if cfg.Server.Mode != "development" {
		t.Errorf("Expected server mode development, got %s", cfg.Server.Mode)
	}

	// Test database config
	if cfg.Database.Type != "sqlite" {
		t.Errorf("Expected database type sqlite, got %s", cfg.Database.Type)
	}

	if cfg.Database.Path != "data/ai-k8s-ops.db" {
		t.Errorf("Expected database path data/ai-k8s-ops.db, got %s", cfg.Database.Path)
	}

	// Test auth config
	if cfg.Auth.JWTSecret == "" {
		t.Error("JWT secret should not be empty")
	}

	if cfg.Auth.JWTExpiryHours != 24 {
		t.Errorf("Expected JWT expiry 24 hours, got %d", cfg.Auth.JWTExpiryHours)
	}
}
