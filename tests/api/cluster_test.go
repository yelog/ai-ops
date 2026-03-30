package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/your-org/ai-k8s-ops/internal/api"
	"github.com/your-org/ai-k8s-ops/internal/auth"
	"github.com/your-org/ai-k8s-ops/internal/cluster"
	"github.com/your-org/ai-k8s-ops/internal/storage/sqlite"
	"github.com/your-org/ai-k8s-ops/pkg/crypto"
)

func setupClusterTestServer(t *testing.T) (*gin.Engine, *auth.UserDB, *cluster.ClusterDB, func()) {
	gin.SetMode(gin.TestMode)

	dbPath := "/tmp/test-cluster-api-" + time.Now().Format("20060102150405.000") + ".db"
	db, err := sqlite.Init(dbPath)
	if err != nil {
		t.Fatalf("Failed to init database: %v", err)
	}

	userDB := auth.NewUserDB(db)
	clusterDB := cluster.NewClusterDB(db)
	router := api.NewRouterWithDB(db, "test-secret", 24*time.Hour)

	cleanup := func() {
		db.Close()
		os.Remove(dbPath)
	}

	return router, userDB, clusterDB, cleanup
}

func TestCreateCluster(t *testing.T) {
	router, userDB, _, cleanup := setupClusterTestServer(t)
	defer cleanup()

	hashedPassword, _ := crypto.HashPassword("password123")
	userDB.Create(&auth.User{
		ID:       uuid.New().String(),
		Username: "testuser",
		Email:    "test@example.com",
		Password: hashedPassword,
		Role:     "operator",
	})
	token, _ := auth.GenerateToken("user-123", "testuser", "operator", "test-secret", 24*time.Hour)

	reqBody := map[string]string{
		"name":        "prod-cluster",
		"description": "Production cluster",
		"environment": "prod",
		"provider":    "bare-metal",
		"version":     "v1.28.0",
		"api_server":  "https://192.168.1.10:6443",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/clusters", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["name"] != "prod-cluster" {
		t.Errorf("Expected name prod-cluster, got %v", response["name"])
	}
}

func TestListClusters(t *testing.T) {
	router, userDB, clusterDB, cleanup := setupClusterTestServer(t)
	defer cleanup()

	clusterDB.Create(&cluster.Cluster{
		ID:          uuid.New().String(),
		Name:        "test-cluster",
		Environment: "dev",
		Status:      "healthy",
	})

	hashedPassword, _ := crypto.HashPassword("password123")
	userDB.Create(&auth.User{
		ID:       "user-123",
		Username: "testuser",
		Email:    "test@example.com",
		Password: hashedPassword,
		Role:     "viewer",
	})
	token, _ := auth.GenerateToken("user-123", "testuser", "viewer", "test-secret", 24*time.Hour)

	req := httptest.NewRequest("GET", "/api/v1/clusters", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["clusters"] == nil {
		t.Error("Expected clusters in response")
	}
}

func TestGetCluster(t *testing.T) {
	router, userDB, clusterDB, cleanup := setupClusterTestServer(t)
	defer cleanup()

	clusterID := uuid.New().String()
	clusterDB.Create(&cluster.Cluster{
		ID:          clusterID,
		Name:        "test-cluster",
		Environment: "dev",
		Status:      "healthy",
	})

	hashedPassword, _ := crypto.HashPassword("password123")
	userDB.Create(&auth.User{
		ID:       "user-123",
		Username: "testuser",
		Email:    "test@example.com",
		Password: hashedPassword,
		Role:     "viewer",
	})
	token, _ := auth.GenerateToken("user-123", "testuser", "viewer", "test-secret", 24*time.Hour)

	req := httptest.NewRequest("GET", "/api/v1/clusters/"+clusterID, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["id"] != clusterID {
		t.Errorf("Expected cluster id %s, got %v", clusterID, response["id"])
	}
}
