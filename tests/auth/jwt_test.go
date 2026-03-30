package auth

import (
	"testing"
	"time"

	"github.com/your-org/ai-k8s-ops/internal/auth"
)

func TestGenerateToken(t *testing.T) {
	secret := "test-secret-key"
	userID := "user-123"
	username := "testuser"
	role := "admin"

	token, err := auth.GenerateToken(userID, username, role, secret, 24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	if token == "" {
		t.Error("Token should not be empty")
	}
}

func TestValidateToken(t *testing.T) {
	secret := "test-secret-key"
	userID := "user-123"
	username := "testuser"
	role := "admin"

	token, err := auth.GenerateToken(userID, username, role, secret, 24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	claims, err := auth.ValidateToken(token, secret)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("Expected userID %s, got %s", userID, claims.UserID)
	}

	if claims.Username != username {
		t.Errorf("Expected username %s, got %s", username, claims.Username)
	}

	if claims.Role != role {
		t.Errorf("Expected role %s, got %s", role, claims.Role)
	}
}

func TestValidateTokenInvalid(t *testing.T) {
	secret := "test-secret-key"

	_, err := auth.ValidateToken("invalid-token", secret)
	if err == nil {
		t.Error("Should fail for invalid token")
	}
}

func TestValidateTokenWrongSecret(t *testing.T) {
	secret := "test-secret-key"
	wrongSecret := "wrong-secret"

	token, _ := auth.GenerateToken("user-123", "testuser", "admin", secret, 24*time.Hour)

	_, err := auth.ValidateToken(token, wrongSecret)
	if err == nil {
		t.Error("Should fail for wrong secret")
	}
}
