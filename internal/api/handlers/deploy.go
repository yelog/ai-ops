package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/your-org/ai-k8s-ops/internal/deploy"
)

type DeployHandler struct {
	templateDB *deploy.TemplateDB
	taskDB     *deploy.TaskDB
}

func NewDeployHandler(templateDB *deploy.TemplateDB, taskDB *deploy.TaskDB) *DeployHandler {
	return &DeployHandler{templateDB: templateDB, taskDB: taskDB}
}

func (h *DeployHandler) CreateTemplate(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		Type        string `json:"type" binding:"required"`
		Provider    string `json:"provider" binding:"required"`
		Config      string `json:"config"`
		Components  string `json:"components"`
		IsDefault   bool   `json:"is_default"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	template := &deploy.DeploymentTemplate{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		Provider:    req.Provider,
		Config:      req.Config,
		Components:  req.Components,
		IsDefault:   req.IsDefault,
	}

	if err := h.templateDB.Create(template); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create template"})
		return
	}

	c.JSON(http.StatusCreated, template)
}

func (h *DeployHandler) ListTemplates(c *gin.Context) {
	templates, err := h.templateDB.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list templates"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"templates": templates})
}

func (h *DeployHandler) GetTemplate(c *gin.Context) {
	id := c.Param("id")

	template, err := h.templateDB.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
		return
	}

	c.JSON(http.StatusOK, template)
}

func (h *DeployHandler) UpdateTemplate(c *gin.Context) {
	id := c.Param("id")

	template, err := h.templateDB.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
		return
	}

	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		Type        string `json:"type" binding:"required"`
		Provider    string `json:"provider" binding:"required"`
		Config      string `json:"config"`
		Components  string `json:"components"`
		IsDefault   bool   `json:"is_default"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	template.Name = req.Name
	template.Description = req.Description
	template.Type = req.Type
	template.Provider = req.Provider
	template.Config = req.Config
	template.Components = req.Components
	template.IsDefault = req.IsDefault

	if err := h.templateDB.Update(template); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update template"})
		return
	}

	c.JSON(http.StatusOK, template)
}

func (h *DeployHandler) DeleteTemplate(c *gin.Context) {
	id := c.Param("id")

	if err := h.templateDB.Delete(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "template deleted"})
}

func (h *DeployHandler) CreateTask(c *gin.Context) {
	var req struct {
		ClusterID  string `json:"cluster_id" binding:"required"`
		TemplateID string `json:"template_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("userID")
	now := time.Now()

	task := &deploy.DeploymentTask{
		ID:         uuid.New().String(),
		ClusterID:  req.ClusterID,
		TemplateID: req.TemplateID,
		Status:     "pending",
		Progress:   0,
		CreatedBy:  userID.(string),
		StartedAt:  &now,
	}

	if err := h.taskDB.Create(task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create task"})
		return
	}

	c.JSON(http.StatusCreated, task)
}

func (h *DeployHandler) ListTasks(c *gin.Context) {
	tasks, err := h.taskDB.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list tasks"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tasks": tasks})
}

func (h *DeployHandler) GetTask(c *gin.Context) {
	id := c.Param("id")

	task, err := h.taskDB.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	c.JSON(http.StatusOK, task)
}
