package storage

import (
	"os"
	"testing"

	"github.com/your-org/ai-k8s-ops/internal/storage/sqlite"
)

func TestDatabaseInit(t *testing.T) {
	dbPath := "/tmp/test-ai-k8s-ops.db"
	defer os.Remove(dbPath)

	db, err := sqlite.Init(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Verify tables exist
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		t.Fatalf("users table does not exist: %v", err)
	}

	err = db.QueryRow("SELECT COUNT(*) FROM clusters").Scan(&count)
	if err != nil {
		t.Fatalf("clusters table does not exist: %v", err)
	}
}
