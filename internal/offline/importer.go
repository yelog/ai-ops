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
