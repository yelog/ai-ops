package crypto

import (
	"testing"

	"github.com/your-org/ai-k8s-ops/pkg/crypto"
)

func TestHashPassword(t *testing.T) {
	password := "testPassword123"

	hash, err := crypto.HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	if hash == "" {
		t.Error("Hash should not be empty")
	}

	if hash == password {
		t.Error("Hash should not equal plain password")
	}
}

func TestCheckPassword(t *testing.T) {
	password := "testPassword123"

	hash, err := crypto.HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	if !crypto.CheckPassword(password, hash) {
		t.Error("Password check should succeed for correct password")
	}

	if crypto.CheckPassword("wrongPassword", hash) {
		t.Error("Password check should fail for wrong password")
	}
}

func TestCheckPasswordEmpty(t *testing.T) {
	hash, _ := crypto.HashPassword("password123")
	if crypto.CheckPassword("", hash) {
		t.Error("Empty password check should fail")
	}
}
