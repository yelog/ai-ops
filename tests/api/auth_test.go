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

	"github.com/your-org/ai-k8s-ops/internal/api"
	"github.com/your-org/ai-k8s-ops/internal/auth"
	"github.com/your-org/ai-k8s-ops/internal/storage/sqlite"
	"github.com/your-org/ai-k8s-ops/pkg/crypto"
)

func setupTestServer(t *testing.T) (*gin.Engine, *auth.UserDB, func()) {
	gin.SetMode(gin.TestMode)

	dbPath := "/tmp/test-auth-" + time.Now().Format("20060102150405") + ".db"
	db, err := sqlite.Init(dbPath)
	if err != nil {
		t.Fatalf("Failed to init database: %v", err)
	}

	userDB := auth.NewUserDB(db)
	router := api.NewRouterWithDB(db, "test-secret-key", 24*time.Hour, nil)

	cleanup := func() {
		db.Close()
		os.Remove(dbPath)
	}

	return router, userDB, cleanup
}

func TestRegisterHandler(t *testing.T) {
	router, _, cleanup := setupTestServer(t)
	defer cleanup()

	reqBody := map[string]string{
		"username": "testuser",
		"email":    "test@example.com",
		"password": "password123",
		"role":     "viewer",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["username"] != "testuser" {
		t.Errorf("Expected username testuser, got %v", response["username"])
	}
}

func TestLoginHandler(t *testing.T) {
	router, userDB, cleanup := setupTestServer(t)
	defer cleanup()

	hashedPassword, _ := crypto.HashPassword("password123")
	user := &auth.User{
		ID:       "user-123",
		Username: "testuser",
		Email:    "test@example.com",
		Password: hashedPassword,
		Role:     "viewer",
	}
	userDB.Create(user)

	reqBody := map[string]string{
		"username": "testuser",
		"password": "password123",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["token"] == nil {
		t.Error("Expected token in response")
	}
}

func TestLoginHandlerInvalidCredentials(t *testing.T) {
	router, userDB, cleanup := setupTestServer(t)
	defer cleanup()

	hashedPassword, _ := crypto.HashPassword("password123")
	user := &auth.User{
		ID:       "user-123",
		Username: "testuser",
		Email:    "test@example.com",
		Password: hashedPassword,
		Role:     "viewer",
	}
	userDB.Create(user)

	reqBody := map[string]string{
		"username": "testuser",
		"password": "wrongpassword",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestGetProfileHandler(t *testing.T) {
	router, userDB, cleanup := setupTestServer(t)
	defer cleanup()

	hashedPassword, _ := crypto.HashPassword("password123")
	user := &auth.User{
		ID:       "user-123",
		Username: "testuser",
		Email:    "test@example.com",
		Password: hashedPassword,
		Role:     "viewer",
	}
	userDB.Create(user)

	token, _ := auth.GenerateToken("user-123", "testuser", "viewer", "test-secret-key", 24*time.Hour)

	req := httptest.NewRequest("GET", "/api/v1/auth/profile", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}
