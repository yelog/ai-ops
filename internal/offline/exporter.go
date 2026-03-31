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

// Export runs the full export pipeline
// If sync is true, runs synchronously (for CLI); otherwise runs in background (for API)
func (e *Exporter) Export(pkg *OfflinePackage, sync ...bool) {
	synchronous := len(sync) > 0 && sync[0]

	runExport := func() {
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
		log.Printf("Export completed for %s", pkg.ID)
	}

	if synchronous {
		runExport()
	} else {
		go runExport()
	}
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

	// Write manifest.json
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
		"install.sh":     installScript,
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
