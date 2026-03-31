package handlers

import (
	"context"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/your-org/ai-k8s-ops/internal/deployer"
)

type DeployerHandler struct {
	deployments sync.Map // map[string]*deployer.Deployer
}

func NewDeployerHandler() *DeployerHandler {
	return &DeployerHandler{}
}

type DeployRequest struct {
	Name              string `json:"name" binding:"required"`
	KubernetesVersion string `json:"kubernetes_version"`
	PodNetworkCIDR    string `json:"pod_network_cidr"`
	ServiceCIDR       string `json:"service_cidr"`
}

type DeployResponse struct {
	ID          string                `json:"id"`
	Name        string                `json:"name"`
	Status      *deployer.DeploymentStatus `json:"status"`
	JoinCommand string                `json:"join_command,omitempty"`
}

// DeployCluster starts a new cluster deployment
func (h *DeployerHandler) DeployCluster(c *gin.Context) {
	var req DeployRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config := deployer.DefaultConfig()
	if req.Name != "" {
		config.Name = req.Name
	}
	if req.KubernetesVersion != "" {
		config.KubernetesVersion = req.KubernetesVersion
	}
	if req.PodNetworkCIDR != "" {
		config.PodNetworkCIDR = req.PodNetworkCIDR
	}
	if req.ServiceCIDR != "" {
		config.ServiceCIDR = req.ServiceCIDR
	}

	id := uuid.New().String()
	d := deployer.NewDeployer(config)
	h.deployments.Store(id, d)

	// Start deployment in background
	go func() {
		ctx := context.Background()
		if err := d.Deploy(ctx); err != nil {
			// Error is already logged by deployer
			return
		}
	}()

	c.JSON(http.StatusCreated, DeployResponse{
		ID:     id,
		Name:   config.Name,
		Status: d.Status(),
	})
}

// GetDeploymentStatus returns deployment status
func (h *DeployerHandler) GetDeploymentStatus(c *gin.Context) {
	id := c.Param("id")

	val, ok := h.deployments.Load(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "deployment not found"})
		return
	}

	d := val.(*deployer.Deployer)
	status := d.Status()

	response := DeployResponse{
		ID:     id,
		Status: status,
	}

	// If deployment is ready, get join command
	if status.Phase == "ready" {
		if joinCmd, err := d.GetJoinCommand(); err == nil {
			response.JoinCommand = joinCmd
		}
	}

	c.JSON(http.StatusOK, response)
}

// ListDeployments lists all deployments
func (h *DeployerHandler) ListDeployments(c *gin.Context) {
	var deployments []DeployResponse

	h.deployments.Range(func(key, value interface{}) bool {
		d := value.(*deployer.Deployer)
		deployments = append(deployments, DeployResponse{
			ID:     key.(string),
			Status: d.Status(),
		})
		return true
	})

	c.JSON(http.StatusOK, gin.H{"deployments": deployments})
}

// DeployOfflinePackage deploys from an offline package
func (h *DeployerHandler) DeployOfflinePackage(c *gin.Context) {
	var req struct {
		PackagePath string `json:"package_path" binding:"required"`
		Name        string `json:"name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := uuid.New().String()

	go func() {
		ctx := context.Background()
		err := deployer.DeployOfflinePackage(ctx, req.PackagePath)
		if err != nil {
			// TODO: Store error in status
			return
		}
	}()

	c.JSON(http.StatusCreated, gin.H{
		"id":      id,
		"message": "Offline deployment started",
	})
}
