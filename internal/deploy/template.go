package deploy

import "time"

type DeploymentTemplate struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Type        string    `json:"type"`       // dev, test, staging, prod, custom
	Provider    string    `json:"provider"`   // bare-metal, vm, cloud
	Config      string    `json:"config"`     // JSON string
	Components  string    `json:"components"` // JSON string
	IsDefault   bool      `json:"is_default"`
	CreatedAt   time.Time `json:"created_at"`
}

type DeploymentTask struct {
	ID           string     `json:"id"`
	ClusterID    string     `json:"cluster_id"`
	TemplateID   string     `json:"template_id"`
	Status       string     `json:"status"` // pending, running, success, failed, rollback
	CurrentStep  string     `json:"current_step"`
	Progress     int        `json:"progress"` // 0-100
	ErrorMessage string     `json:"error_message,omitempty"`
	CreatedBy    string     `json:"created_by"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	FinishedAt   *time.Time `json:"finished_at,omitempty"`
}
