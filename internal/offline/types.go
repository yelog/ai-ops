package offline

import "time"

type OfflinePackage struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Version      string    `json:"version"`
	OSList       string    `json:"os_list"` // JSON ["ubuntu","centos"]
	Modules      string    `json:"modules"` // JSON ["core","network"]
	Status       string    `json:"status"`  // pending, exporting, ready, failed
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
