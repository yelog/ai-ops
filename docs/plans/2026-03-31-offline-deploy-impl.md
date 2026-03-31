---
render_with_liquid: false
---

# Offline Deploy Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement modular offline package management for fully air-gapped K8S v1.28.0 deployments, with CLI + Web UI dual entry points.

**Architecture:** New `internal/offline` module with PackageDB, Exporter, Importer services. Resource manifest defined as Go constants. CLI extends `cmd/cli`. Frontend adds Offline tab under Deploy section. Shell scripts handle image loading and dependency installation on target nodes.

**Tech Stack:**
- Backend: Go, Gin, SQLite, archive/tar, compress/gzip, crypto/sha256
- CLI: Go (flag package, consistent with existing cmd/ pattern)
- Frontend: React, TypeScript, Ant Design (Table, Card, Modal, Upload, Progress, Tabs, Collapse)
- Scripts: Bash (Ubuntu dpkg + CentOS rpm)

---

## Task 1: Create Offline Package Model and Database

**Files:**
- Create: `internal/offline/types.go`
- Create: `internal/offline/package_db.go`
- Modify: `internal/storage/sqlite/db.go:209-217`
- Create: `tests/offline/offline_test.go`

**Step 1: Define offline types**

Create file: `internal/offline/types.go`

```go
package offline

import "time"

type OfflinePackage struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Version      string    `json:"version"`
	OSList       string    `json:"os_list"`       // JSON ["ubuntu","centos"]
	Modules      string    `json:"modules"`       // JSON ["core","network"]
	Status       string    `json:"status"`        // pending, exporting, ready, failed
	Size         int64     `json:"size"`
	Checksum     string    `json:"checksum"`
	StoragePath  string    `json:"storage_path"`
	ErrorMessage string    `json:"error_message,omitempty"`
	CreatedBy    string    `json:"created_by"`
	CreatedAt    time.Time `json:"created_at"`
}

type ModuleInfo struct {
	Name          string   `json:"name"`
	Required      bool     `json:"required"`
	Description   string   `json:"description"`
	Images        []string `json:"images"`
	Binaries      []string `json:"binaries,omitempty"`
	EstimatedSize string   `json:"estimated_size"`
}

type ResourceManifest struct {
	Version string       `json:"version"`
	Modules []ModuleInfo `json:"modules"`
}

type ExportRequest struct {
	Name    string   `json:"name" binding:"required"`
	OSList  []string `json:"os_list" binding:"required"`
	Modules []string `json:"modules" binding:"required"`
}
```

**Step 2: Add offline_packages table to schema**

Modify file: `internal/storage/sqlite/db.go` — append before the closing backtick of `getSchema()`, after the last CREATE INDEX:

```go
CREATE TABLE IF NOT EXISTS offline_packages (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    version TEXT NOT NULL,
    os_list TEXT NOT NULL,
    modules TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    size INTEGER DEFAULT 0,
    checksum TEXT,
    storage_path TEXT,
    error_message TEXT,
    created_by TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (created_by) REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_offline_packages_status ON offline_packages(status);
```

**Step 3: Create PackageDB**

Create file: `internal/offline/package_db.go`

```go
package offline

import (
	"database/sql"
	"errors"
	"time"
)

type PackageDB struct {
	db *sql.DB
}

func NewPackageDB(db *sql.DB) *PackageDB {
	return &PackageDB{db: db}
}

func (r *PackageDB) Create(p *OfflinePackage) error {
	_, err := r.db.Exec(`
		INSERT INTO offline_packages (id, name, version, os_list, modules, status, size, checksum, storage_path, error_message, created_by, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, p.ID, p.Name, p.Version, p.OSList, p.Modules, p.Status, p.Size, p.Checksum, p.StoragePath, p.ErrorMessage, p.CreatedBy, time.Now())
	return err
}

func (r *PackageDB) GetByID(id string) (*OfflinePackage, error) {
	p := &OfflinePackage{}
	err := r.db.QueryRow(`
		SELECT id, name, version, os_list, modules, status, size, checksum, storage_path, error_message, created_by, created_at
		FROM offline_packages WHERE id = ?
	`, id).Scan(&p.ID, &p.Name, &p.Version, &p.OSList, &p.Modules, &p.Status, &p.Size, &p.Checksum, &p.StoragePath, &p.ErrorMessage, &p.CreatedBy, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *PackageDB) List() ([]*OfflinePackage, error) {
	rows, err := r.db.Query(`
		SELECT id, name, version, os_list, modules, status, size, checksum, storage_path, error_message, created_by, created_at
		FROM offline_packages ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var packages []*OfflinePackage
	for rows.Next() {
		p := &OfflinePackage{}
		err := rows.Scan(&p.ID, &p.Name, &p.Version, &p.OSList, &p.Modules, &p.Status, &p.Size, &p.Checksum, &p.StoragePath, &p.ErrorMessage, &p.CreatedBy, &p.CreatedAt)
		if err != nil {
			return nil, err
		}
		packages = append(packages, p)
	}
	return packages, nil
}

func (r *PackageDB) UpdateStatus(id string, status string, errMsg string) error {
	result, err := r.db.Exec(`
		UPDATE offline_packages SET status = ?, error_message = ? WHERE id = ?
	`, status, errMsg, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("package not found")
	}
	return nil
}

func (r *PackageDB) UpdateComplete(id string, size int64, checksum string, storagePath string) error {
	result, err := r.db.Exec(`
		UPDATE offline_packages SET status = 'ready', size = ?, checksum = ?, storage_path = ? WHERE id = ?
	`, size, checksum, storagePath, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("package not found")
	}
	return nil
}

func (r *PackageDB) Delete(id string) error {
	result, err := r.db.Exec(`DELETE FROM offline_packages WHERE id = ?`, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("package not found")
	}
	return nil
}
```

**Step 4: Write tests**

Create file: `tests/offline/offline_test.go`

```go
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
```

**Step 5: Run tests**

```bash
go test ./tests/offline/... -v
```

**Step 6: Commit**

```bash
git add internal/offline internal/storage/sqlite/db.go tests/offline
git commit -m "feat(offline): add offline package model and database operations"
```

---

## Task 2: Create Resource Manifest

**Files:**
- Create: `internal/offline/manifest.go`

**Step 1: Define resource manifest with all module definitions**

Create file: `internal/offline/manifest.go`

```go
package offline

const K8sVersion = "v1.28.0"

func GetResourceManifest() *ResourceManifest {
	return &ResourceManifest{
		Version: K8sVersion,
		Modules: []ModuleInfo{
			{
				Name:     "core",
				Required: true,
				Description: "K8S 核心组件",
				Images: []string{
					"registry.k8s.io/kube-apiserver:v1.28.0",
					"registry.k8s.io/kube-controller-manager:v1.28.0",
					"registry.k8s.io/kube-scheduler:v1.28.0",
					"registry.k8s.io/kube-proxy:v1.28.0",
					"registry.k8s.io/etcd:3.5.9-0",
					"registry.k8s.io/coredns/coredns:v1.10.1",
					"registry.k8s.io/pause:3.9",
				},
				Binaries: []string{
					"kubeadm",
					"kubelet",
					"kubectl",
					"crictl",
					"containerd",
				},
				EstimatedSize: "2.1GB",
			},
			{
				Name:     "network",
				Required: true,
				Description: "网络插件 (Calico)",
				Images: []string{
					"docker.io/calico/cni:v3.26.1",
					"docker.io/calico/node:v3.26.1",
					"docker.io/calico/kube-controllers:v3.26.1",
				},
				EstimatedSize: "320MB",
			},
			{
				Name:     "monitoring",
				Required: false,
				Description: "Prometheus + Grafana",
				Images: []string{
					"quay.io/prometheus/prometheus:v2.47.0",
					"grafana/grafana:10.1.0",
					"quay.io/prometheus/node-exporter:v1.6.1",
					"registry.k8s.io/kube-state-metrics/kube-state-metrics:v2.10.0",
				},
				EstimatedSize: "800MB",
			},
			{
				Name:     "logging",
				Required: false,
				Description: "Loki 日志系统",
				Images: []string{
					"grafana/loki:2.9.1",
					"grafana/promtail:2.9.1",
				},
				EstimatedSize: "200MB",
			},
			{
				Name:     "tracing",
				Required: false,
				Description: "Jaeger 追踪系统",
				Images: []string{
					"jaegertracing/all-in-one:1.49",
				},
				EstimatedSize: "150MB",
			},
		},
	}
}

// ValidModules returns list of valid module names
func ValidModules() []string {
	return []string{"core", "network", "monitoring", "logging", "tracing"}
}

// ValidOSList returns list of supported OS
func ValidOSList() []string {
	return []string{"ubuntu", "centos"}
}

// ValidateModules checks if all module names are valid
func ValidateModules(modules []string) bool {
	valid := make(map[string]bool)
	for _, m := range ValidModules() {
		valid[m] = true
	}
	for _, m := range modules {
		if !valid[m] {
			return false
		}
	}
	return true
}

// ValidateOSList checks if all OS names are valid
func ValidateOSList(osList []string) bool {
	valid := make(map[string]bool)
	for _, os := range ValidOSList() {
		valid[os] = true
	}
	for _, os := range osList {
		if !valid[os] {
			return false
		}
	}
	return true
}

// HasRequiredModules checks if required modules are present
func HasRequiredModules(modules []string) bool {
	moduleSet := make(map[string]bool)
	for _, m := range modules {
		moduleSet[m] = true
	}
	manifest := GetResourceManifest()
	for _, mod := range manifest.Modules {
		if mod.Required && !moduleSet[mod.Name] {
			return false
		}
	}
	return true
}
```

**Step 2: Commit**

```bash
git add internal/offline/manifest.go
git commit -m "feat(offline): define resource manifest with module definitions"
```

---

## Task 3: Create Exporter Service

**Files:**
- Create: `internal/offline/exporter.go`

**Step 1: Implement the export service**

Create file: `internal/offline/exporter.go`

```go
package offline

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Exporter struct {
	packageDB *PackageDB
	outputDir string
}

func NewExporter(packageDB *PackageDB, outputDir string) *Exporter {
	return &Exporter{packageDB: packageDB, outputDir: outputDir}
}

// Export runs the full export pipeline in background
func (e *Exporter) Export(pkg *OfflinePackage) {
	go func() {
		err := e.packageDB.UpdateStatus(pkg.ID, "exporting", "")
		if err != nil {
			log.Printf("Failed to update status for %s: %v", pkg.ID, err)
			return
		}

		if err := e.doExport(pkg); err != nil {
			log.Printf("Export failed for %s: %v", pkg.ID, err)
			e.packageDB.UpdateStatus(pkg.ID, "failed", err.Error())
			return
		}
	}()
}

func (e *Exporter) doExport(pkg *OfflinePackage) error {
	// Parse modules and OS list
	var modules []string
	if err := json.Unmarshal([]byte(pkg.Modules), &modules); err != nil {
		return fmt.Errorf("invalid modules JSON: %w", err)
	}
	var osList []string
	if err := json.Unmarshal([]byte(pkg.OSList), &osList); err != nil {
		return fmt.Errorf("invalid os_list JSON: %w", err)
	}

	// Create temp working directory
	workDir := filepath.Join(e.outputDir, "work", pkg.ID)
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return fmt.Errorf("failed to create work dir: %w", err)
	}
	defer os.RemoveAll(workDir)

	manifest := GetResourceManifest()

	// Write manifest.yaml
	if err := e.writeManifest(workDir, pkg, modules, osList); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	// Pull and save images for selected modules
	for _, modName := range modules {
		for _, mod := range manifest.Modules {
			if mod.Name == modName {
				if err := e.exportModuleImages(workDir, mod); err != nil {
					return fmt.Errorf("failed to export images for module %s: %w", modName, err)
				}
				break
			}
		}
	}

	// Download binaries for core module
	for _, modName := range modules {
		if modName == "core" {
			if err := e.exportCoreBinaries(workDir); err != nil {
				return fmt.Errorf("failed to export core binaries: %w", err)
			}
			break
		}
	}

	// Copy install scripts
	if err := e.copyScripts(workDir); err != nil {
		return fmt.Errorf("failed to copy scripts: %w", err)
	}

	// Create tar.gz archive
	archivePath := filepath.Join(e.outputDir, "packages", fmt.Sprintf("%s.tar.gz", pkg.Name))
	if err := os.MkdirAll(filepath.Dir(archivePath), 0755); err != nil {
		return fmt.Errorf("failed to create packages dir: %w", err)
	}

	if err := createTarGz(archivePath, workDir); err != nil {
		return fmt.Errorf("failed to create archive: %w", err)
	}

	// Calculate checksum and size
	checksum, size, err := checksumFile(archivePath)
	if err != nil {
		return fmt.Errorf("failed to checksum: %w", err)
	}

	return e.packageDB.UpdateComplete(pkg.ID, size, checksum, archivePath)
}

func (e *Exporter) writeManifest(workDir string, pkg *OfflinePackage, modules []string, osList []string) error {
	m := map[string]interface{}{
		"version": pkg.Version,
		"name":    pkg.Name,
		"os_list": osList,
		"modules": modules,
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(workDir, "manifest.json"), data, 0644)
}

func (e *Exporter) exportModuleImages(workDir string, mod ModuleInfo) error {
	imageDir := filepath.Join(workDir, mod.Name, "images")
	if err := os.MkdirAll(imageDir, 0755); err != nil {
		return err
	}

	for _, image := range mod.Images {
		log.Printf("Pulling image: %s", image)

		// docker pull
		pullCmd := exec.Command("docker", "pull", image)
		pullCmd.Stdout = os.Stdout
		pullCmd.Stderr = os.Stderr
		if err := pullCmd.Run(); err != nil {
			return fmt.Errorf("failed to pull image %s: %w", image, err)
		}

		// docker save to tar
		safeName := strings.ReplaceAll(image, "/", "_")
		safeName = strings.ReplaceAll(safeName, ":", "_")
		tarPath := filepath.Join(imageDir, safeName+".tar")

		saveCmd := exec.Command("docker", "save", "-o", tarPath, image)
		if err := saveCmd.Run(); err != nil {
			return fmt.Errorf("failed to save image %s: %w", image, err)
		}

		log.Printf("Saved image: %s -> %s", image, tarPath)
	}
	return nil
}

func (e *Exporter) exportCoreBinaries(workDir string) error {
	binDir := filepath.Join(workDir, "core", "binaries")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return err
	}

	// Define download URLs for K8S v1.28.0 binaries
	binaries := map[string]string{
		"kubeadm": fmt.Sprintf("https://dl.k8s.io/release/%s/bin/linux/amd64/kubeadm", K8sVersion),
		"kubelet": fmt.Sprintf("https://dl.k8s.io/release/%s/bin/linux/amd64/kubelet", K8sVersion),
		"kubectl": fmt.Sprintf("https://dl.k8s.io/release/%s/bin/linux/amd64/kubectl", K8sVersion),
	}

	for name, url := range binaries {
		log.Printf("Downloading binary: %s", name)
		outPath := filepath.Join(binDir, name)
		cmd := exec.Command("curl", "-L", "-o", outPath, url)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to download %s: %w", name, err)
		}
		os.Chmod(outPath, 0755)
	}

	return nil
}

func (e *Exporter) copyScripts(workDir string) error {
	scriptDir := filepath.Join(workDir, "scripts")
	if err := os.MkdirAll(scriptDir, 0755); err != nil {
		return err
	}

	// Embed install scripts content
	scripts := map[string]string{
		"install.sh":    installScript,
		"load-images.sh": loadImagesScript,
		"setup-deps.sh":  setupDepsScript,
	}

	for name, content := range scripts {
		path := filepath.Join(scriptDir, name)
		if err := os.WriteFile(path, []byte(content), 0755); err != nil {
			return fmt.Errorf("failed to write script %s: %w", name, err)
		}
	}
	return nil
}

// createTarGz creates a .tar.gz archive from a source directory
func createTarGz(archivePath string, sourceDir string) error {
	file, err := os.Create(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	gw := gzip.NewWriter(file)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = relPath

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(tw, f)
		return err
	})
}

// checksumFile calculates SHA256 checksum and file size
func checksumFile(path string) (string, int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", 0, err
	}
	defer f.Close()

	h := sha256.New()
	size, err := io.Copy(h, f)
	if err != nil {
		return "", 0, err
	}

	return "sha256:" + hex.EncodeToString(h.Sum(nil)), size, nil
}

// Embedded script contents
var installScript = `#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PKG_DIR="$(dirname "$SCRIPT_DIR")"

echo "=== AI-K8S-OPS Offline Installer ==="
echo "Package directory: $PKG_DIR"

# Detect OS
if [ -f /etc/os-release ]; then
    . /etc/os-release
    OS=$ID
else
    echo "ERROR: Cannot detect OS"
    exit 1
fi
echo "Detected OS: $OS"

# Step 1: Install system dependencies
echo "Step 1: Installing system dependencies..."
bash "$SCRIPT_DIR/setup-deps.sh" "$PKG_DIR" "$OS"

# Step 2: Load container images
echo "Step 2: Loading container images..."
bash "$SCRIPT_DIR/load-images.sh" "$PKG_DIR"

# Step 3: Install binaries
echo "Step 3: Installing K8S binaries..."
if [ -d "$PKG_DIR/core/binaries" ]; then
    cp "$PKG_DIR/core/binaries/kubeadm" /usr/local/bin/
    cp "$PKG_DIR/core/binaries/kubelet" /usr/local/bin/
    cp "$PKG_DIR/core/binaries/kubectl" /usr/local/bin/
    chmod +x /usr/local/bin/{kubeadm,kubelet,kubectl}
    echo "Binaries installed to /usr/local/bin/"
fi

echo "=== Offline installation complete ==="
echo "Next steps:"
echo "  kubeadm init --kubernetes-version=v1.28.0"
`

var loadImagesScript = `#!/bin/bash
set -e

PKG_DIR="${1:-.}"

echo "Loading container images..."

find "$PKG_DIR" -name "*.tar" -path "*/images/*" | while read tar_file; do
    echo "  Loading: $(basename $tar_file)"
    ctr -n k8s.io images import "$tar_file" 2>/dev/null || \
    nerdctl load -i "$tar_file" 2>/dev/null || \
    docker load -i "$tar_file" 2>/dev/null || \
    echo "  WARNING: Failed to load $tar_file"
done

echo "Image loading complete."
`

var setupDepsScript = `#!/bin/bash
set -e

PKG_DIR="${1:-.}"
OS="${2:-ubuntu}"

echo "Setting up system dependencies for $OS..."

case "$OS" in
    ubuntu|debian)
        if [ -d "$PKG_DIR/core/packages/ubuntu" ]; then
            dpkg -i "$PKG_DIR/core/packages/ubuntu/"*.deb 2>/dev/null || true
            echo "Ubuntu packages installed."
        fi
        ;;
    centos|rhel|rocky|alma)
        if [ -d "$PKG_DIR/core/packages/centos" ]; then
            rpm -ivh "$PKG_DIR/core/packages/centos/"*.rpm 2>/dev/null || true
            echo "CentOS packages installed."
        fi
        ;;
    *)
        echo "WARNING: Unsupported OS: $OS"
        ;;
esac

# Disable swap
swapoff -a || true
sed -i '/swap/d' /etc/fstab || true

# Load kernel modules
modprobe br_netfilter || true
modprobe overlay || true

# Set sysctl
cat > /etc/sysctl.d/k8s.conf <<SYSCTL
net.bridge.bridge-nf-call-iptables  = 1
net.bridge.bridge-nf-call-ip6tables = 1
net.ipv4.ip_forward                 = 1
SYSCTL
sysctl --system > /dev/null 2>&1

echo "System dependencies setup complete."
`
```

**Step 2: Commit**

```bash
git add internal/offline/exporter.go
git commit -m "feat(offline): implement exporter service with tar.gz packaging"
```

---

## Task 4: Create Importer Service

**Files:**
- Create: `internal/offline/importer.go`

**Step 1: Implement import service**

Create file: `internal/offline/importer.go`

```go
package offline

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Importer struct {
	packageDB *PackageDB
	storeDir  string
}

func NewImporter(packageDB *PackageDB, storeDir string) *Importer {
	return &Importer{packageDB: packageDB, storeDir: storeDir}
}

// Import validates and registers an uploaded offline package file
func (i *Importer) Import(filePath string, userID string) (*OfflinePackage, error) {
	// Verify file exists
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}

	// Calculate checksum
	checksum, err := calcChecksum(filePath)
	if err != nil {
		return nil, fmt.Errorf("checksum failed: %w", err)
	}

	// Read manifest from archive
	manifest, err := readManifestFromArchive(filePath)
	if err != nil {
		return nil, fmt.Errorf("invalid package: %w", err)
	}

	modulesJSON, _ := json.Marshal(manifest["modules"])
	osListJSON, _ := json.Marshal(manifest["os_list"])
	version, _ := manifest["version"].(string)
	name, _ := manifest["name"].(string)

	pkg := &OfflinePackage{
		ID:          fmt.Sprintf("pkg-%s", checksum[7:19]),
		Name:        name,
		Version:     version,
		OSList:      string(osListJSON),
		Modules:     string(modulesJSON),
		Status:      "ready",
		Size:        info.Size(),
		Checksum:    checksum,
		StoragePath: filePath,
		CreatedBy:   userID,
	}

	if err := i.packageDB.Create(pkg); err != nil {
		return nil, fmt.Errorf("failed to register package: %w", err)
	}

	return pkg, nil
}

// VerifyChecksum verifies the integrity of a package file
func VerifyChecksum(filePath string, expectedChecksum string) error {
	actual, err := calcChecksum(filePath)
	if err != nil {
		return err
	}
	if actual != expectedChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actual)
	}
	return nil
}

func calcChecksum(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return "sha256:" + hex.EncodeToString(h.Sum(nil)), nil
}

// readManifestFromArchive reads manifest.json from a .tar.gz file
func readManifestFromArchive(archivePath string) (map[string]interface{}, error) {
	f, err := os.Open(archivePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	gr, err := gzip.NewReader(f)
	if err != nil {
		return nil, fmt.Errorf("not a valid gzip file: %w", err)
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if filepath.Base(header.Name) == "manifest.json" {
			data, err := io.ReadAll(tr)
			if err != nil {
				return nil, err
			}
			var manifest map[string]interface{}
			if err := json.Unmarshal(data, &manifest); err != nil {
				return nil, fmt.Errorf("invalid manifest.json: %w", err)
			}
			return manifest, nil
		}
	}
	return nil, fmt.Errorf("manifest.json not found in archive")
}
```

**Step 2: Commit**

```bash
git add internal/offline/importer.go
git commit -m "feat(offline): implement importer service with checksum verification"
```

---

## Task 5: Create Offline API Handlers

**Files:**
- Create: `internal/api/handlers/offline.go`
- Modify: `internal/api/router.go`

**Step 1: Create offline handlers**

Create file: `internal/api/handlers/offline.go`

```go
package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/your-org/ai-k8s-ops/internal/offline"
)

type OfflineHandler struct {
	packageDB *offline.PackageDB
	exporter  *offline.Exporter
	importer  *offline.Importer
}

func NewOfflineHandler(packageDB *offline.PackageDB, exporter *offline.Exporter, importer *offline.Importer) *OfflineHandler {
	return &OfflineHandler{packageDB: packageDB, exporter: exporter, importer: importer}
}

func (h *OfflineHandler) GetResources(c *gin.Context) {
	manifest := offline.GetResourceManifest()
	c.JSON(http.StatusOK, manifest)
}

func (h *OfflineHandler) ExportPackage(c *gin.Context) {
	var req offline.ExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate modules
	if !offline.ValidateModules(req.Modules) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid module name"})
		return
	}
	if !offline.HasRequiredModules(req.Modules) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing required modules: core, network"})
		return
	}
	if !offline.ValidateOSList(req.OSList) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid OS name, must be ubuntu or centos"})
		return
	}

	modulesJSON, _ := json.Marshal(req.Modules)
	osListJSON, _ := json.Marshal(req.OSList)

	userID, _ := c.Get("userID")

	pkg := &offline.OfflinePackage{
		ID:      uuid.New().String(),
		Name:    req.Name,
		Version: offline.K8sVersion,
		OSList:  string(osListJSON),
		Modules: string(modulesJSON),
		Status:  "pending",
		CreatedBy: userID.(string),
	}

	if err := h.packageDB.Create(pkg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create export task"})
		return
	}

	// Start async export
	h.exporter.Export(pkg)

	c.JSON(http.StatusCreated, pkg)
}

func (h *OfflineHandler) ImportPackage(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no file uploaded"})
		return
	}
	defer file.Close()

	// Save uploaded file
	uploadDir := "data/offline/uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create upload directory"})
		return
	}

	filePath := filepath.Join(uploadDir, header.Filename)
	out, err := os.Create(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to write file"})
		return
	}

	userID, _ := c.Get("userID")
	pkg, err := h.importer.Import(filePath, userID.(string))
	if err != nil {
		os.Remove(filePath)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("import failed: %v", err)})
		return
	}

	c.JSON(http.StatusCreated, pkg)
}

func (h *OfflineHandler) ListPackages(c *gin.Context) {
	packages, err := h.packageDB.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list packages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"packages": packages})
}

func (h *OfflineHandler) GetPackage(c *gin.Context) {
	id := c.Param("id")

	pkg, err := h.packageDB.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "package not found"})
		return
	}

	c.JSON(http.StatusOK, pkg)
}

func (h *OfflineHandler) DeletePackage(c *gin.Context) {
	id := c.Param("id")

	pkg, err := h.packageDB.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "package not found"})
		return
	}

	// Remove file from disk
	if pkg.StoragePath != "" {
		os.Remove(pkg.StoragePath)
	}

	if err := h.packageDB.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete package"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "package deleted"})
}

func (h *OfflineHandler) DownloadPackage(c *gin.Context) {
	id := c.Param("id")

	pkg, err := h.packageDB.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "package not found"})
		return
	}

	if pkg.Status != "ready" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "package is not ready for download"})
		return
	}

	if _, err := os.Stat(pkg.StoragePath); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "package file not found on disk"})
		return
	}

	fileName := filepath.Base(pkg.StoragePath)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	c.File(pkg.StoragePath)
}
```

**Step 2: Update router to add offline routes**

Modify file: `internal/api/router.go` — add the offline routes after the deploy task routes block (after line 83):

Add import:
```go
"github.com/your-org/ai-k8s-ops/internal/offline"
```

Add route group:
```go
			offlineDB := offline.NewPackageDB(db)
			offlineHandler := handlers.NewOfflineHandler(
				offlineDB,
				offline.NewExporter(offlineDB, "data/offline"),
				offline.NewImporter(offlineDB, "data/offline"),
			)

			offlineGroup := v1.Group("/offline")
			offlineGroup.Use(middleware.AuthMiddleware(jwtSecret))
			{
				offlineGroup.GET("/resources", offlineHandler.GetResources)
				offlineGroup.POST("/packages/export", offlineHandler.ExportPackage)
				offlineGroup.POST("/packages/import", offlineHandler.ImportPackage)
				offlineGroup.GET("/packages", offlineHandler.ListPackages)
				offlineGroup.GET("/packages/:id", offlineHandler.GetPackage)
				offlineGroup.DELETE("/packages/:id", offlineHandler.DeletePackage)
				offlineGroup.GET("/packages/:id/download", offlineHandler.DownloadPackage)
			}
```

**Step 3: Run backend tests**

```bash
go test ./... -v -short
```

**Step 4: Commit**

```bash
git add internal/api/handlers/offline.go internal/api/router.go
git commit -m "feat(offline): add offline API handlers and routes"
```

---

## Task 6: Extend CLI with Offline Commands

**Files:**
- Modify: `cmd/cli/main.go`

**Step 1: Implement CLI offline commands**

Replace file: `cmd/cli/main.go`

```go
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/your-org/ai-k8s-ops/internal/offline"
	"github.com/your-org/ai-k8s-ops/internal/storage/sqlite"
	"github.com/your-org/ai-k8s-ops/pkg/version"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(0)
	}

	switch os.Args[1] {
	case "version":
		fmt.Printf("AI-K8S-OPS CLI v%s\n", version.Version)
	case "offline":
		if len(os.Args) < 3 {
			printOfflineUsage()
			os.Exit(1)
		}
		handleOffline(os.Args[2])
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Printf("AI-K8S-OPS CLI v%s\n\n", version.Version)
	fmt.Println("Usage: ai-k8s-ops <command> [subcommand] [flags]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  version    Show version info")
	fmt.Println("  offline    Manage offline packages")
}

func printOfflineUsage() {
	fmt.Println("Usage: ai-k8s-ops offline <subcommand> [flags]")
	fmt.Println()
	fmt.Println("Subcommands:")
	fmt.Println("  export     Export offline package (requires network)")
	fmt.Println("  import     Import offline package")
	fmt.Println("  list       List offline packages")
	fmt.Println("  inspect    Inspect offline package file")
}

func handleOffline(subcmd string) {
	switch subcmd {
	case "export":
		handleExport()
	case "import":
		handleImport()
	case "list":
		handleList()
	case "inspect":
		handleInspect()
	default:
		fmt.Printf("Unknown offline subcommand: %s\n", subcmd)
		printOfflineUsage()
		os.Exit(1)
	}
}

func handleExport() {
	fs := flag.NewFlagSet("export", flag.ExitOnError)
	name := fs.String("name", "", "Package name (required)")
	osFlag := fs.String("os", "ubuntu,centos", "Target OS list (comma-separated)")
	modules := fs.String("modules", "core,network", "Modules to include (comma-separated)")
	output := fs.String("output", "data/offline", "Output directory")
	dbPath := fs.String("db", "data/ai-k8s-ops.db", "Database path")
	fs.Parse(os.Args[3:])

	if *name == "" {
		fmt.Println("Error: --name is required")
		fs.PrintDefaults()
		os.Exit(1)
	}

	osList := strings.Split(*osFlag, ",")
	moduleList := strings.Split(*modules, ",")

	if !offline.ValidateModules(moduleList) {
		fmt.Printf("Error: invalid module. Valid modules: %s\n", strings.Join(offline.ValidModules(), ", "))
		os.Exit(1)
	}
	if !offline.HasRequiredModules(moduleList) {
		fmt.Println("Error: core and network modules are required")
		os.Exit(1)
	}
	if !offline.ValidateOSList(osList) {
		fmt.Printf("Error: invalid OS. Valid OS: %s\n", strings.Join(offline.ValidOSList(), ", "))
		os.Exit(1)
	}

	db, err := sqlite.Init(*dbPath)
	if err != nil {
		fmt.Printf("Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	packageDB := offline.NewPackageDB(db)
	exporter := offline.NewExporter(packageDB, *output)

	modulesJSON, _ := json.Marshal(moduleList)
	osListJSON, _ := json.Marshal(osList)

	pkg := &offline.OfflinePackage{
		ID:      fmt.Sprintf("cli-%d", os.Getpid()),
		Name:    *name,
		Version: offline.K8sVersion,
		OSList:  string(osListJSON),
		Modules: string(modulesJSON),
		Status:  "pending",
		CreatedBy: "cli",
	}

	if err := packageDB.Create(pkg); err != nil {
		fmt.Printf("Error creating package record: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Exporting offline package: %s\n", *name)
	fmt.Printf("  Version: %s\n", offline.K8sVersion)
	fmt.Printf("  OS: %s\n", *osFlag)
	fmt.Printf("  Modules: %s\n", *modules)
	fmt.Printf("  Output: %s\n", *output)
	fmt.Println()

	// Run export synchronously for CLI
	exporter.Export(pkg)

	fmt.Println("Export started. Check status with: ai-k8s-ops offline list")
}

func handleImport() {
	fs := flag.NewFlagSet("import", flag.ExitOnError)
	file := fs.String("file", "", "Package file path (required)")
	dbPath := fs.String("db", "data/ai-k8s-ops.db", "Database path")
	fs.Parse(os.Args[3:])

	if *file == "" {
		fmt.Println("Error: --file is required")
		fs.PrintDefaults()
		os.Exit(1)
	}

	db, err := sqlite.Init(*dbPath)
	if err != nil {
		fmt.Printf("Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	packageDB := offline.NewPackageDB(db)
	importer := offline.NewImporter(packageDB, "data/offline")

	fmt.Printf("Importing offline package: %s\n", *file)

	pkg, err := importer.Import(*file, "cli")
	if err != nil {
		fmt.Printf("Import failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Import successful!\n")
	fmt.Printf("  ID: %s\n", pkg.ID)
	fmt.Printf("  Name: %s\n", pkg.Name)
	fmt.Printf("  Version: %s\n", pkg.Version)
	fmt.Printf("  Checksum: %s\n", pkg.Checksum)
}

func handleList() {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	dbPath := fs.String("db", "data/ai-k8s-ops.db", "Database path")
	fs.Parse(os.Args[3:])

	db, err := sqlite.Init(*dbPath)
	if err != nil {
		fmt.Printf("Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	packageDB := offline.NewPackageDB(db)
	packages, err := packageDB.List()
	if err != nil {
		fmt.Printf("Error listing packages: %v\n", err)
		os.Exit(1)
	}

	if len(packages) == 0 {
		fmt.Println("No offline packages found.")
		return
	}

	fmt.Printf("%-36s %-20s %-10s %-12s %-10s\n", "ID", "NAME", "VERSION", "STATUS", "SIZE")
	fmt.Println(strings.Repeat("-", 90))
	for _, p := range packages {
		size := formatSize(p.Size)
		fmt.Printf("%-36s %-20s %-10s %-12s %-10s\n", p.ID, p.Name, p.Version, p.Status, size)
	}
}

func handleInspect() {
	fs := flag.NewFlagSet("inspect", flag.ExitOnError)
	file := fs.String("file", "", "Package file path (required)")
	fs.Parse(os.Args[3:])

	if *file == "" {
		fmt.Println("Error: --file is required")
		fs.PrintDefaults()
		os.Exit(1)
	}

	info, err := os.Stat(*file)
	if err != nil {
		fmt.Printf("Error: file not found: %s\n", *file)
		os.Exit(1)
	}

	fmt.Printf("File: %s\n", *file)
	fmt.Printf("Size: %s\n", formatSize(info.Size()))
	fmt.Println("(Full inspection requires extracting manifest - not yet implemented)")
}

func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1fGB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1fMB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1fKB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%dB", bytes)
	}
}
```

**Step 2: Build CLI**

```bash
go build -o bin/ai-k8s-ops cmd/cli/main.go
```

**Step 3: Commit**

```bash
git add cmd/cli/main.go
git commit -m "feat(offline): implement CLI offline export/import/list/inspect commands"
```

---

## Task 7: Create Frontend Types and Service

**Files:**
- Create: `frontend/src/types/offline.ts`
- Create: `frontend/src/services/offline.service.ts`

**Step 1: Create TypeScript types**

Create file: `frontend/src/types/offline.ts`

```typescript
export interface OfflinePackage {
  id: string
  name: string
  version: string
  os_list: string      // JSON string
  modules: string      // JSON string
  status: 'pending' | 'exporting' | 'ready' | 'failed'
  size: number
  checksum: string
  storage_path: string
  error_message?: string
  created_by: string
  created_at: string
}

export interface ModuleInfo {
  name: string
  required: boolean
  description: string
  images: string[]
  binaries?: string[]
  estimated_size: string
}

export interface ResourceManifest {
  version: string
  modules: ModuleInfo[]
}

export interface ExportRequest {
  name: string
  os_list: string[]
  modules: string[]
}

export interface PackageListResponse {
  packages: OfflinePackage[]
}
```

**Step 2: Create offline service**

Create file: `frontend/src/services/offline.service.ts`

```typescript
import api from './api'
import type {
  OfflinePackage,
  ResourceManifest,
  ExportRequest,
  PackageListResponse,
} from '@/types/offline'

export const offlineService = {
  async getResources(): Promise<ResourceManifest> {
    const response = await api.get<ResourceManifest>('/api/v1/offline/resources')
    return response.data
  },

  async exportPackage(data: ExportRequest): Promise<OfflinePackage> {
    const response = await api.post<OfflinePackage>('/api/v1/offline/packages/export', data)
    return response.data
  },

  async importPackage(file: File): Promise<OfflinePackage> {
    const formData = new FormData()
    formData.append('file', file)
    const response = await api.post<OfflinePackage>('/api/v1/offline/packages/import', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    return response.data
  },

  async listPackages(): Promise<OfflinePackage[]> {
    const response = await api.get<PackageListResponse>('/api/v1/offline/packages')
    return response.data.packages
  },

  async getPackage(id: string): Promise<OfflinePackage> {
    const response = await api.get<OfflinePackage>(`/api/v1/offline/packages/${id}`)
    return response.data
  },

  async deletePackage(id: string): Promise<void> {
    await api.delete(`/api/v1/offline/packages/${id}`)
  },

  getDownloadUrl(id: string): string {
    const baseUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080'
    return `${baseUrl}/api/v1/offline/packages/${id}/download`
  },
}
```

**Step 3: Commit**

```bash
git add frontend/src/types/offline.ts frontend/src/services/offline.service.ts
git commit -m "feat(offline): create frontend types and API service"
```

---

## Task 8: Create Offline Management Page

**Files:**
- Create: `frontend/src/pages/Deploy/Offline/index.tsx`

**Step 1: Create offline management page with Tabs**

Create file: `frontend/src/pages/Deploy/Offline/index.tsx`

```tsx
import { useState, useEffect } from 'react'
import {
  Card, Tabs, Table, Button, Tag, Space, Progress, Modal,
  Form, Input, Checkbox, Upload, Collapse, message, Popconfirm,
} from 'antd'
import {
  ExportOutlined, ImportOutlined, ReloadOutlined,
  DownloadOutlined, DeleteOutlined, InboxOutlined,
} from '@ant-design/icons'
import { offlineService } from '@/services/offline.service'
import type { OfflinePackage, ResourceManifest } from '@/types/offline'

const { TabPane } = Tabs
const { Dragger } = Upload

const statusConfig: Record<string, { color: string; text: string }> = {
  pending: { color: 'default', text: '等待中' },
  exporting: { color: 'processing', text: '导出中' },
  ready: { color: 'success', text: '就绪' },
  failed: { color: 'error', text: '失败' },
}

function formatSize(bytes: number): string {
  if (bytes >= 1073741824) return (bytes / 1073741824).toFixed(1) + ' GB'
  if (bytes >= 1048576) return (bytes / 1048576).toFixed(1) + ' MB'
  if (bytes >= 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return bytes + ' B'
}

export default function OfflinePage() {
  const [packages, setPackages] = useState<OfflinePackage[]>([])
  const [manifest, setManifest] = useState<ResourceManifest | null>(null)
  const [loading, setLoading] = useState(false)
  const [exportModalOpen, setExportModalOpen] = useState(false)
  const [importModalOpen, setImportModalOpen] = useState(false)
  const [exportForm] = Form.useForm()

  const fetchPackages = async () => {
    setLoading(true)
    try {
      const data = await offlineService.listPackages()
      setPackages(data || [])
    } catch {
      message.error('获取离线包列表失败')
    } finally {
      setLoading(false)
    }
  }

  const fetchManifest = async () => {
    try {
      const data = await offlineService.getResources()
      setManifest(data)
    } catch {
      message.error('获取资源清单失败')
    }
  }

  useEffect(() => {
    fetchPackages()
    fetchManifest()
    const interval = setInterval(fetchPackages, 5000)
    return () => clearInterval(interval)
  }, [])

  const handleExport = async (values: any) => {
    try {
      await offlineService.exportPackage({
        name: values.name,
        os_list: values.os_list,
        modules: values.modules,
      })
      message.success('导出任务已创建')
      setExportModalOpen(false)
      exportForm.resetFields()
      fetchPackages()
    } catch {
      message.error('创建导出任务失败')
    }
  }

  const handleImport = async (file: File) => {
    try {
      await offlineService.importPackage(file)
      message.success('导入成功')
      setImportModalOpen(false)
      fetchPackages()
    } catch {
      message.error('导入失败')
    }
    return false
  }

  const handleDelete = async (id: string) => {
    try {
      await offlineService.deletePackage(id)
      message.success('删除成功')
      fetchPackages()
    } catch {
      message.error('删除失败')
    }
  }

  const handleDownload = (id: string) => {
    const token = localStorage.getItem('token')
    const url = offlineService.getDownloadUrl(id)
    const a = document.createElement('a')
    a.href = `${url}?token=${token}`
    a.click()
  }

  const columns = [
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
      width: 180,
    },
    {
      title: '版本',
      dataIndex: 'version',
      key: 'version',
      width: 100,
    },
    {
      title: '模块',
      dataIndex: 'modules',
      key: 'modules',
      render: (val: string) => {
        try {
          const modules: string[] = JSON.parse(val)
          return modules.map(m => <Tag key={m}>{m}</Tag>)
        } catch {
          return val
        }
      },
    },
    {
      title: 'OS',
      dataIndex: 'os_list',
      key: 'os_list',
      render: (val: string) => {
        try {
          const osList: string[] = JSON.parse(val)
          return osList.map(os => <Tag key={os} color="blue">{os}</Tag>)
        } catch {
          return val
        }
      },
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 120,
      render: (status: string) => {
        const config = statusConfig[status] || { color: 'default', text: status }
        return <Tag color={config.color}>{config.text}</Tag>
      },
    },
    {
      title: '大小',
      dataIndex: 'size',
      key: 'size',
      width: 100,
      render: (size: number) => size > 0 ? formatSize(size) : '-',
    },
    {
      title: '操作',
      key: 'action',
      width: 150,
      render: (_: any, record: OfflinePackage) => (
        <Space>
          {record.status === 'ready' && (
            <Button
              type="link"
              icon={<DownloadOutlined />}
              onClick={() => handleDownload(record.id)}
            >
              下载
            </Button>
          )}
          <Popconfirm
            title="确定删除此离线包？"
            onConfirm={() => handleDelete(record.id)}
            okText="确定"
            cancelText="取消"
          >
            <Button type="link" danger icon={<DeleteOutlined />}>
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  const moduleOptions = manifest?.modules.map(m => ({
    label: `${m.name} - ${m.description} (~${m.estimated_size})`,
    value: m.name,
    disabled: m.required,
  })) || []

  return (
    <Card
      title="离线管理"
      extra={
        <Space>
          <Button icon={<ReloadOutlined />} onClick={fetchPackages}>
            刷新
          </Button>
          <Button icon={<ImportOutlined />} onClick={() => setImportModalOpen(true)}>
            导入离线包
          </Button>
          <Button type="primary" icon={<ExportOutlined />} onClick={() => setExportModalOpen(true)}>
            导出离线包
          </Button>
        </Space>
      }
    >
      <Tabs defaultActiveKey="packages">
        <TabPane tab="离线包管理" key="packages">
          <Table
            columns={columns}
            dataSource={packages}
            rowKey="id"
            loading={loading}
            pagination={{ pageSize: 10 }}
            expandable={{
              expandedRowRender: (record: OfflinePackage) => (
                <div>
                  {record.status === 'exporting' && (
                    <Progress percent={30} status="active" style={{ maxWidth: 400 }} />
                  )}
                  {record.error_message && (
                    <p style={{ color: '#ff4d4f' }}>错误: {record.error_message}</p>
                  )}
                  {record.checksum && <p>校验和: <code>{record.checksum}</code></p>}
                </div>
              ),
            }}
          />
        </TabPane>

        <TabPane tab="资源清单" key="resources">
          {manifest && (
            <div>
              <p style={{ marginBottom: 16 }}>
                <strong>K8S 版本:</strong> {manifest.version}
              </p>
              <Collapse>
                {manifest.modules.map(mod => (
                  <Collapse.Panel
                    key={mod.name}
                    header={
                      <span>
                        <Tag color={mod.required ? 'red' : 'blue'}>
                          {mod.required ? '必选' : '可选'}
                        </Tag>
                        {mod.name} - {mod.description} ({mod.estimated_size})
                      </span>
                    }
                  >
                    <h4>容器镜像</h4>
                    <Table
                      size="small"
                      pagination={false}
                      dataSource={mod.images.map((img, i) => ({ key: i, type: '镜像', name: img }))}
                      columns={[
                        { title: '类型', dataIndex: 'type', width: 80 },
                        { title: '名称', dataIndex: 'name' },
                      ]}
                    />
                    {mod.binaries && mod.binaries.length > 0 && (
                      <>
                        <h4 style={{ marginTop: 16 }}>二进制文件</h4>
                        <Table
                          size="small"
                          pagination={false}
                          dataSource={mod.binaries.map((b, i) => ({ key: i, type: '二进制', name: b }))}
                          columns={[
                            { title: '类型', dataIndex: 'type', width: 80 },
                            { title: '名称', dataIndex: 'name' },
                          ]}
                        />
                      </>
                    )}
                  </Collapse.Panel>
                ))}
              </Collapse>
            </div>
          )}
        </TabPane>
      </Tabs>

      {/* Export Modal */}
      <Modal
        title="导出离线包"
        open={exportModalOpen}
        onCancel={() => setExportModalOpen(false)}
        footer={null}
      >
        <Form
          form={exportForm}
          layout="vertical"
          onFinish={handleExport}
          initialValues={{
            os_list: ['ubuntu', 'centos'],
            modules: ['core', 'network'],
          }}
        >
          <Form.Item
            name="name"
            label="包名称"
            rules={[{ required: true, message: '请输入包名称' }]}
          >
            <Input placeholder="例: prod-offline-pkg" />
          </Form.Item>

          <Form.Item name="os_list" label="目标 OS">
            <Checkbox.Group>
              <Checkbox value="ubuntu">Ubuntu 20.04/22.04</Checkbox>
              <Checkbox value="centos">CentOS 7/8</Checkbox>
            </Checkbox.Group>
          </Form.Item>

          <Form.Item name="modules" label="选择模块">
            <Checkbox.Group options={moduleOptions} />
          </Form.Item>

          <Form.Item>
            <Space>
              <Button onClick={() => setExportModalOpen(false)}>取消</Button>
              <Button type="primary" htmlType="submit">开始导出</Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      {/* Import Modal */}
      <Modal
        title="导入离线包"
        open={importModalOpen}
        onCancel={() => setImportModalOpen(false)}
        footer={null}
      >
        <Dragger
          accept=".tar.gz,.tgz"
          maxCount={1}
          beforeUpload={(file) => {
            handleImport(file)
            return false
          }}
        >
          <p className="ant-upload-drag-icon">
            <InboxOutlined />
          </p>
          <p className="ant-upload-text">将 .tar.gz 文件拖到此处</p>
          <p className="ant-upload-hint">或点击选择文件</p>
        </Dragger>
      </Modal>
    </Card>
  )
}
```

**Step 2: Commit**

```bash
git add frontend/src/pages/Deploy/Offline
git commit -m "feat(offline): create offline management page with export/import/resource tabs"
```

---

## Task 9: Integrate Offline Page into Router and Menu

**Files:**
- Modify: `frontend/src/App.tsx`
- Modify: `frontend/src/components/layouts/MainLayout.tsx`

**Step 1: Add offline route to App.tsx**

Add import at top of `frontend/src/App.tsx`:

```tsx
import OfflinePage from '@/pages/Deploy/Offline'
```

Add route after the deploy/tasks route:

```tsx
<Route path="deploy/offline" element={<OfflinePage />} />
```

**Step 2: Update MainLayout menu**

In `frontend/src/components/layouts/MainLayout.tsx`, change the deploy menu item to have children by replacing the simple deploy entry with a submenu:

Replace:
```tsx
  { key: '/deploy', icon: <CloudUploadOutlined />, label: '部署中心' },
```

With:
```tsx
  {
    key: '/deploy',
    icon: <CloudUploadOutlined />,
    label: '部署中心',
    children: [
      { key: '/deploy', label: '概览' },
      { key: '/deploy/templates', label: '部署模板' },
      { key: '/deploy/tasks', label: '部署任务' },
      { key: '/deploy/offline', label: '离线管理' },
    ],
  },
```

**Step 3: Verify frontend compiles**

```bash
cd frontend && npm run build
```

**Step 4: Commit**

```bash
git add frontend/src/App.tsx frontend/src/components/layouts/MainLayout.tsx
git commit -m "feat(offline): integrate offline page into router and sidebar menu"
```

---

## Task 10: Create Offline Install Scripts

**Files:**
- Create: `scripts/offline/install.sh`
- Create: `scripts/offline/load-images.sh`
- Create: `scripts/offline/setup-deps.sh`

**Step 1: Create standalone install scripts**

These are the same scripts embedded in Go but available as standalone files for manual use.

Create file: `scripts/offline/install.sh`

```bash
#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PKG_DIR="$(dirname "$SCRIPT_DIR")"

echo "=== AI-K8S-OPS Offline Installer ==="
echo "Package directory: $PKG_DIR"

# Detect OS
if [ -f /etc/os-release ]; then
    . /etc/os-release
    OS=$ID
else
    echo "ERROR: Cannot detect OS"
    exit 1
fi
echo "Detected OS: $OS"

# Preflight checks
echo ""
echo "=== Preflight Checks ==="

# Check disk space (need at least 5GB free)
FREE_SPACE=$(df -BG "$PKG_DIR" | tail -1 | awk '{print $4}' | tr -d 'G')
if [ "$FREE_SPACE" -lt 5 ]; then
    echo "ERROR: Need at least 5GB free disk space, only ${FREE_SPACE}GB available"
    exit 1
fi
echo "Disk space: ${FREE_SPACE}GB available"

# Check if running as root
if [ "$(id -u)" -ne 0 ]; then
    echo "ERROR: Must run as root"
    exit 1
fi
echo "Running as root: OK"

# Step 1: Install system dependencies
echo ""
echo "=== Step 1: Installing system dependencies ==="
bash "$SCRIPT_DIR/setup-deps.sh" "$PKG_DIR" "$OS"

# Step 2: Load container images
echo ""
echo "=== Step 2: Loading container images ==="
bash "$SCRIPT_DIR/load-images.sh" "$PKG_DIR"

# Step 3: Install binaries
echo ""
echo "=== Step 3: Installing K8S binaries ==="
if [ -d "$PKG_DIR/core/binaries" ]; then
    cp "$PKG_DIR/core/binaries/kubeadm" /usr/local/bin/
    cp "$PKG_DIR/core/binaries/kubelet" /usr/local/bin/
    cp "$PKG_DIR/core/binaries/kubectl" /usr/local/bin/
    chmod +x /usr/local/bin/{kubeadm,kubelet,kubectl}
    echo "Binaries installed to /usr/local/bin/"
    
    kubeadm version --short 2>/dev/null && echo "kubeadm: OK"
    kubectl version --client --short 2>/dev/null && echo "kubectl: OK"
fi

echo ""
echo "=== Offline installation complete ==="
echo ""
echo "Next steps:"
echo "  1. Initialize cluster:  kubeadm init --kubernetes-version=v1.28.0"
echo "  2. Set up kubeconfig:   mkdir -p ~/.kube && cp /etc/kubernetes/admin.conf ~/.kube/config"
echo "  3. Install network:     kubectl apply -f calico.yaml"
echo "  4. Join worker nodes:   kubeadm join ..."
```

Create file: `scripts/offline/load-images.sh`

```bash
#!/bin/bash
set -e

PKG_DIR="${1:-.}"

echo "Loading container images..."

LOADED=0
FAILED=0

find "$PKG_DIR" -name "*.tar" -path "*/images/*" | sort | while read tar_file; do
    name=$(basename "$tar_file")
    echo -n "  Loading: $name ... "
    
    if ctr -n k8s.io images import "$tar_file" 2>/dev/null; then
        echo "OK (containerd)"
        LOADED=$((LOADED+1))
    elif nerdctl load -i "$tar_file" 2>/dev/null; then
        echo "OK (nerdctl)"
        LOADED=$((LOADED+1))
    elif docker load -i "$tar_file" 2>/dev/null; then
        echo "OK (docker)"
        LOADED=$((LOADED+1))
    else
        echo "FAILED"
        FAILED=$((FAILED+1))
    fi
done

echo ""
echo "Image loading complete."
```

Create file: `scripts/offline/setup-deps.sh`

```bash
#!/bin/bash
set -e

PKG_DIR="${1:-.}"
OS="${2:-ubuntu}"

echo "Setting up system dependencies for $OS..."

case "$OS" in
    ubuntu|debian)
        if [ -d "$PKG_DIR/core/packages/ubuntu" ]; then
            echo "Installing Ubuntu/Debian packages..."
            dpkg -i "$PKG_DIR/core/packages/ubuntu/"*.deb 2>/dev/null || true
            echo "Ubuntu packages installed."
        else
            echo "No Ubuntu packages found in $PKG_DIR/core/packages/ubuntu"
        fi
        ;;
    centos|rhel|rocky|alma)
        if [ -d "$PKG_DIR/core/packages/centos" ]; then
            echo "Installing CentOS/RHEL packages..."
            rpm -ivh "$PKG_DIR/core/packages/centos/"*.rpm 2>/dev/null || true
            echo "CentOS packages installed."
        else
            echo "No CentOS packages found in $PKG_DIR/core/packages/centos"
        fi
        ;;
    *)
        echo "WARNING: Unsupported OS: $OS"
        echo "Supported: ubuntu, debian, centos, rhel, rocky, alma"
        ;;
esac

# Disable swap
echo "Disabling swap..."
swapoff -a || true
sed -i '/swap/d' /etc/fstab || true

# Load kernel modules
echo "Loading kernel modules..."
modprobe br_netfilter || true
modprobe overlay || true

cat > /etc/modules-load.d/k8s.conf <<EOF
br_netfilter
overlay
EOF

# Set sysctl parameters
echo "Configuring sysctl..."
cat > /etc/sysctl.d/k8s.conf <<EOF
net.bridge.bridge-nf-call-iptables  = 1
net.bridge.bridge-nf-call-ip6tables = 1
net.ipv4.ip_forward                 = 1
EOF
sysctl --system > /dev/null 2>&1

echo "System dependencies setup complete."
```

**Step 2: Commit**

```bash
git add scripts/offline
git commit -m "feat(offline): add standalone offline install scripts"
```

---

## Task 11: Create Data Directories

**Files:**
- Create: `data/offline/packages/.gitkeep`
- Create: `data/offline/uploads/.gitkeep`

**Step 1: Create directory structure**

```bash
mkdir -p data/offline/packages data/offline/uploads
touch data/offline/packages/.gitkeep data/offline/uploads/.gitkeep
```

**Step 2: Commit**

```bash
git add data/offline
git commit -m "chore: add offline data directories"
```

---

## Summary

This plan implements the offline deployment feature in 11 tasks:

1. Offline package model + DB operations + tests
2. Resource manifest (module definitions, validation)
3. Exporter service (image pull, binary download, tar.gz packaging)
4. Importer service (checksum verification, manifest parsing)
5. API handlers + router integration
6. CLI offline commands (export/import/list/inspect)
7. Frontend types + API service
8. Offline management page (package list, export modal, import modal, resource view)
9. Router + menu integration
10. Standalone install scripts
11. Data directories

**Plan complete and saved to `docs/plans/2026-03-31-offline-deploy-impl.md`**
