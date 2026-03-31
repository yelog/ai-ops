package deploy

import (
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/your-org/ai-k8s-ops/internal/deploy"
	"github.com/your-org/ai-k8s-ops/internal/storage/sqlite"
)

func TestTemplateModel(t *testing.T) {
	dbPath := "/tmp/test-deploy.db"
	defer os.Remove(dbPath)

	db, err := sqlite.Init(dbPath)
	if err != nil {
		t.Fatalf("Failed to init database: %v", err)
	}
	defer db.Close()

	templateDB := deploy.NewTemplateDB(db)

	template := &deploy.DeploymentTemplate{
		ID:          "template-123",
		Name:        "dev-template",
		Description: "Development environment",
		Type:        "dev",
		Provider:    "bare-metal",
		Config:      "{}",
		Components:  "[]",
		IsDefault:   true,
	}

	err = templateDB.Create(template)
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	retrieved, err := templateDB.GetByID("template-123")
	if err != nil {
		t.Fatalf("Failed to get template: %v", err)
	}

	if retrieved.Name != "dev-template" {
		t.Errorf("Expected name dev-template, got %s", retrieved.Name)
	}

	templates, err := templateDB.List()
	if err != nil {
		t.Fatalf("Failed to list templates: %v", err)
	}

	if len(templates) != 1 {
		t.Errorf("Expected 1 template, got %d", len(templates))
	}

	template.Description = "Updated description"
	err = templateDB.Update(template)
	if err != nil {
		t.Fatalf("Failed to update template: %v", err)
	}

	updated, _ := templateDB.GetByID("template-123")
	if updated.Description != "Updated description" {
		t.Errorf("Expected updated description, got %s", updated.Description)
	}

	err = templateDB.Delete("template-123")
	if err != nil {
		t.Fatalf("Failed to delete template: %v", err)
	}

	_, err = templateDB.GetByID("template-123")
	if err == nil {
		t.Error("Template should be deleted")
	}
}

func TestTaskModel(t *testing.T) {
	dbPath := "/tmp/test-deploy-task.db"
	defer os.Remove(dbPath)

	db, err := sqlite.Init(dbPath)
	if err != nil {
		t.Fatalf("Failed to init database: %v", err)
	}
	defer db.Close()

	taskDB := deploy.NewTaskDB(db)

	task := &deploy.DeploymentTask{
		ID:         "task-123",
		ClusterID:  "cluster-123",
		TemplateID: "template-123",
		Status:     "pending",
		Progress:   0,
		CreatedBy:  "user-123",
	}

	err = taskDB.Create(task)
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	retrieved, err := taskDB.GetByID("task-123")
	if err != nil {
		t.Fatalf("Failed to get task: %v", err)
	}

	if retrieved.Status != "pending" {
		t.Errorf("Expected status pending, got %s", retrieved.Status)
	}

	err = taskDB.UpdateProgress("task-123", "Installing components", 50)
	if err != nil {
		t.Fatalf("Failed to update progress: %v", err)
	}

	updated, _ := taskDB.GetByID("task-123")
	if updated.Progress != 50 {
		t.Errorf("Expected progress 50, got %d", updated.Progress)
	}

	if updated.CurrentStep != "Installing components" {
		t.Errorf("Expected step 'Installing components', got %s", updated.CurrentStep)
	}

	err = taskDB.UpdateStatus("task-123", "running", "")
	if err != nil {
		t.Fatalf("Failed to update status: %v", err)
	}

	updated2, _ := taskDB.GetByID("task-123")
	if updated2.Status != "running" {
		t.Errorf("Expected status running, got %s", updated2.Status)
	}

	tasks, err := taskDB.List()
	if err != nil {
		t.Fatalf("Failed to list tasks: %v", err)
	}

	if len(tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(tasks))
	}
}
