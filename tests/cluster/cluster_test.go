package cluster

import (
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/your-org/ai-k8s-ops/internal/cluster"
	"github.com/your-org/ai-k8s-ops/internal/storage/sqlite"
)

func TestClusterModel(t *testing.T) {
	dbPath := "/tmp/test-cluster.db"
	defer os.Remove(dbPath)

	db, err := sqlite.Init(dbPath)
	if err != nil {
		t.Fatalf("Failed to init database: %v", err)
	}
	defer db.Close()

	clusterDB := cluster.NewClusterDB(db)

	c := &cluster.Cluster{
		ID:          "cluster-123",
		Name:        "test-cluster",
		Description: "Test cluster",
		Environment: "dev",
		Provider:    "bare-metal",
		Version:     "v1.28.0",
		APIServer:   "https://192.168.1.10:6443",
		Status:      "healthy",
	}

	err = clusterDB.Create(c)
	if err != nil {
		t.Fatalf("Failed to create cluster: %v", err)
	}

	retrieved, err := clusterDB.GetByID("cluster-123")
	if err != nil {
		t.Fatalf("Failed to get cluster: %v", err)
	}

	if retrieved.Name != "test-cluster" {
		t.Errorf("Expected name test-cluster, got %s", retrieved.Name)
	}

	clusters, err := clusterDB.List()
	if err != nil {
		t.Fatalf("Failed to list clusters: %v", err)
	}

	if len(clusters) != 1 {
		t.Errorf("Expected 1 cluster, got %d", len(clusters))
	}

	c.Status = "warning"
	err = clusterDB.Update(c)
	if err != nil {
		t.Fatalf("Failed to update cluster: %v", err)
	}

	updated, _ := clusterDB.GetByID("cluster-123")
	if updated.Status != "warning" {
		t.Errorf("Expected status warning, got %s", updated.Status)
	}

	err = clusterDB.Delete("cluster-123")
	if err != nil {
		t.Fatalf("Failed to delete cluster: %v", err)
	}

	_, err = clusterDB.GetByID("cluster-123")
	if err == nil {
		t.Error("Cluster should be deleted")
	}
}
