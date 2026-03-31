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
		ID:        uuid.New().String(),
		Name:      req.Name,
		Version:   offline.K8sVersion,
		OSList:    string(osListJSON),
		Modules:   string(modulesJSON),
		Status:    "pending",
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
