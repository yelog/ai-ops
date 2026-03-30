package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/your-org/ai-k8s-ops/internal/cluster"
)

type ClusterHandler struct {
	clusterDB *cluster.ClusterDB
}

func NewClusterHandler(clusterDB *cluster.ClusterDB) *ClusterHandler {
	return &ClusterHandler{clusterDB: clusterDB}
}

type CreateClusterRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Environment string `json:"environment" binding:"required"`
	Provider    string `json:"provider" binding:"required"`
	Version     string `json:"version"`
	APIServer   string `json:"api_server"`
	Kubeconfig  string `json:"kubeconfig"`
}

func (h *ClusterHandler) CreateCluster(c *gin.Context) {
	var req CreateClusterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if _, err := h.clusterDB.GetByName(req.Name); err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "cluster name already exists"})
		return
	}

	cl := &cluster.Cluster{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		Environment: req.Environment,
		Provider:    req.Provider,
		Version:     req.Version,
		APIServer:   req.APIServer,
		Kubeconfig:  req.Kubeconfig,
		Status:      "pending",
	}

	if err := h.clusterDB.Create(cl); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create cluster"})
		return
	}

	c.JSON(http.StatusCreated, cl)
}

func (h *ClusterHandler) ListClusters(c *gin.Context) {
	env := c.Query("environment")

	var clusters []*cluster.Cluster
	var err error

	if env != "" {
		clusters, err = h.clusterDB.ListByEnvironment(env)
	} else {
		clusters, err = h.clusterDB.List()
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list clusters"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"clusters": clusters})
}

func (h *ClusterHandler) GetCluster(c *gin.Context) {
	id := c.Param("id")

	cl, err := h.clusterDB.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "cluster not found"})
		return
	}

	c.JSON(http.StatusOK, cl)
}

func (h *ClusterHandler) UpdateCluster(c *gin.Context) {
	id := c.Param("id")

	var req CreateClusterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cl, err := h.clusterDB.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "cluster not found"})
		return
	}

	cl.Name = req.Name
	cl.Description = req.Description
	cl.Environment = req.Environment
	cl.Provider = req.Provider
	cl.Version = req.Version
	cl.APIServer = req.APIServer

	if err := h.clusterDB.Update(cl); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update cluster"})
		return
	}

	c.JSON(http.StatusOK, cl)
}

func (h *ClusterHandler) DeleteCluster(c *gin.Context) {
	id := c.Param("id")

	if err := h.clusterDB.Delete(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "cluster not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "cluster deleted"})
}
