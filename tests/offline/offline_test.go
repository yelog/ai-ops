package offline

import (
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/your-org/ai-k8s-ops/internal/offline"
	"github.com/your-org/ai-k8s-ops/internal/storage/sqlite"
)

func TestPackageDB_CreateAndGet(t *testing.T) {
	dbPath := "/tmp/test-offline.db"
	defer os.Remove(dbPath)

	db, err := sqlite.Init(dbPath)
	if err != nil {
		t.Fatalf("Failed to init database: %v", err)
	}
	defer db.Close()

	packageDB := offline.NewPackageDB(db)

	pkg := &offline.OfflinePackage{
		ID:      "pkg-001",
		Name:    "test-offline-pkg",
		Version: "v1.28.0",
		OSList:  `["ubuntu","centos"]`,
		Modules: `["core","network"]`,
		Status:  "pending",
	}

	err = packageDB.Create(pkg)
	if err != nil {
		t.Fatalf("Failed to create package: %v", err)
	}

	retrieved, err := packageDB.GetByID("pkg-001")
	if err != nil {
		t.Fatalf("Failed to get package: %v", err)
	}

	if retrieved.Name != "test-offline-pkg" {
		t.Errorf("Expected name test-offline-pkg, got %s", retrieved.Name)
	}
	if retrieved.Version != "v1.28.0" {
		t.Errorf("Expected version v1.28.0, got %s", retrieved.Version)
	}
}

func TestPackageDB_List(t *testing.T) {
	dbPath := "/tmp/test-offline-list.db"
	defer os.Remove(dbPath)

	db, err := sqlite.Init(dbPath)
	if err != nil {
		t.Fatalf("Failed to init database: %v", err)
	}
	defer db.Close()

	packageDB := offline.NewPackageDB(db)

	packageDB.Create(&offline.OfflinePackage{
		ID: "pkg-001", Name: "pkg1", Version: "v1.28.0",
		OSList: `["ubuntu"]`, Modules: `["core"]`, Status: "ready",
	})
	packageDB.Create(&offline.OfflinePackage{
		ID: "pkg-002", Name: "pkg2", Version: "v1.28.0",
		OSList: `["centos"]`, Modules: `["core","network"]`, Status: "pending",
	})

	packages, err := packageDB.List()
	if err != nil {
		t.Fatalf("Failed to list packages: %v", err)
	}
	if len(packages) != 2 {
		t.Errorf("Expected 2 packages, got %d", len(packages))
	}
}

func TestPackageDB_UpdateStatus(t *testing.T) {
	dbPath := "/tmp/test-offline-status.db"
	defer os.Remove(dbPath)

	db, err := sqlite.Init(dbPath)
	if err != nil {
		t.Fatalf("Failed to init database: %v", err)
	}
	defer db.Close()

	packageDB := offline.NewPackageDB(db)

	packageDB.Create(&offline.OfflinePackage{
		ID: "pkg-001", Name: "pkg1", Version: "v1.28.0",
		OSList: `["ubuntu"]`, Modules: `["core"]`, Status: "pending",
	})

	err = packageDB.UpdateStatus("pkg-001", "exporting", "")
	if err != nil {
		t.Fatalf("Failed to update status: %v", err)
	}

	pkg, _ := packageDB.GetByID("pkg-001")
	if pkg.Status != "exporting" {
		t.Errorf("Expected status exporting, got %s", pkg.Status)
	}
}

func TestPackageDB_UpdateComplete(t *testing.T) {
	dbPath := "/tmp/test-offline-complete.db"
	defer os.Remove(dbPath)

	db, err := sqlite.Init(dbPath)
	if err != nil {
		t.Fatalf("Failed to init database: %v", err)
	}
	defer db.Close()

	packageDB := offline.NewPackageDB(db)

	packageDB.Create(&offline.OfflinePackage{
		ID: "pkg-001", Name: "pkg1", Version: "v1.28.0",
		OSList: `["ubuntu"]`, Modules: `["core"]`, Status: "exporting",
	})

	err = packageDB.UpdateComplete("pkg-001", 2147483648, "sha256:abc123", "data/offline/packages/pkg-001.tar.gz")
	if err != nil {
		t.Fatalf("Failed to update complete: %v", err)
	}

	pkg, _ := packageDB.GetByID("pkg-001")
	if pkg.Status != "ready" {
		t.Errorf("Expected status ready, got %s", pkg.Status)
	}
	if pkg.Size != 2147483648 {
		t.Errorf("Expected size 2147483648, got %d", pkg.Size)
	}
}

func TestPackageDB_Delete(t *testing.T) {
	dbPath := "/tmp/test-offline-delete.db"
	defer os.Remove(dbPath)

	db, err := sqlite.Init(dbPath)
	if err != nil {
		t.Fatalf("Failed to init database: %v", err)
	}
	defer db.Close()

	packageDB := offline.NewPackageDB(db)

	packageDB.Create(&offline.OfflinePackage{
		ID: "pkg-001", Name: "pkg1", Version: "v1.28.0",
		OSList: `["ubuntu"]`, Modules: `["core"]`, Status: "ready",
	})

	err = packageDB.Delete("pkg-001")
	if err != nil {
		t.Fatalf("Failed to delete package: %v", err)
	}

	_, err = packageDB.GetByID("pkg-001")
	if err == nil {
		t.Error("Expected error getting deleted package")
	}
}
