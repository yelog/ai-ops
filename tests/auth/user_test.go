package auth

import (
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/your-org/ai-k8s-ops/internal/auth"
	"github.com/your-org/ai-k8s-ops/internal/storage/sqlite"
)

func TestUserModel(t *testing.T) {
	dbPath := "/tmp/test-user.db"
	defer os.Remove(dbPath)

	db, err := sqlite.Init(dbPath)
	if err != nil {
		t.Fatalf("Failed to init database: %v", err)
	}
	defer db.Close()

	userDB := auth.NewUserDB(db)

	user := &auth.User{
		ID:       "user-123",
		Username: "testuser",
		Email:    "test@example.com",
		Password: "hashedPassword",
		Role:     "viewer",
	}

	err = userDB.Create(user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	retrieved, err := userDB.GetByUsername("testuser")
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}

	if retrieved.ID != user.ID {
		t.Errorf("Expected ID %s, got %s", user.ID, retrieved.ID)
	}

	if retrieved.Email != user.Email {
		t.Errorf("Expected email %s, got %s", user.Email, retrieved.Email)
	}

	byID, err := userDB.GetByID("user-123")
	if err != nil {
		t.Fatalf("Failed to get user by ID: %v", err)
	}

	if byID.Username != "testuser" {
		t.Errorf("Expected username testuser, got %s", byID.Username)
	}

	duplicate := &auth.User{
		ID:       "user-456",
		Username: "testuser",
		Email:    "another@example.com",
		Password: "hashedPassword",
		Role:     "viewer",
	}

	err = userDB.Create(duplicate)
	if err == nil {
		t.Error("Should fail for duplicate username")
	}

	users, err := userDB.List()
	if err != nil {
		t.Fatalf("Failed to list users: %v", err)
	}

	if len(users) != 1 {
		t.Errorf("Expected 1 user, got %d", len(users))
	}
}

func TestUpdateUser(t *testing.T) {
	dbPath := "/tmp/test-user-update.db"
	defer os.Remove(dbPath)

	db, err := sqlite.Init(dbPath)
	if err != nil {
		t.Fatalf("Failed to init database: %v", err)
	}
	defer db.Close()

	userDB := auth.NewUserDB(db)

	user := &auth.User{
		ID:       "user-123",
		Username: "testuser",
		Email:    "test@example.com",
		Password: "hashedPassword",
		Role:     "viewer",
	}

	userDB.Create(user)

	user.Email = "newemail@example.com"
	user.Role = "operator"

	err = userDB.Update(user)
	if err != nil {
		t.Fatalf("Failed to update user: %v", err)
	}

	updated, _ := userDB.GetByID("user-123")
	if updated.Email != "newemail@example.com" {
		t.Errorf("Expected email newemail@example.com, got %s", updated.Email)
	}

	if updated.Role != "operator" {
		t.Errorf("Expected role operator, got %s", updated.Role)
	}
}

func TestDeleteUser(t *testing.T) {
	dbPath := "/tmp/test-user-delete.db"
	defer os.Remove(dbPath)

	db, err := sqlite.Init(dbPath)
	if err != nil {
		t.Fatalf("Failed to init database: %v", err)
	}
	defer db.Close()

	userDB := auth.NewUserDB(db)

	user := &auth.User{
		ID:       "user-123",
		Username: "testuser",
		Email:    "test@example.com",
		Password: "hashedPassword",
		Role:     "viewer",
	}

	userDB.Create(user)

	err = userDB.Delete("user-123")
	if err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}

	_, err = userDB.GetByID("user-123")
	if err == nil {
		t.Error("User should be deleted")
	}
}
