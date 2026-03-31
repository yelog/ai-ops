# Phase 1 Week 7-8: Deployment Feature Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement deployment template management, task creation, progress tracking, and real-time logs.

**Architecture:** Backend deployment module with database operations. Frontend pages for template management and task monitoring. WebSocket for real-time log streaming.

**Tech Stack:**
- Backend: Go, Gorilla WebSocket
- Frontend: React, XTerm.js for terminal
- Database: SQLite

---

## Task 1: Create Deployment Model and Database Operations

**Files:**
- Create: `internal/deploy/template.go`
- Create: `internal/deploy/template_db.go`
- Create: `internal/deploy/task.go`
- Create: `internal/deploy/task_db.go`
- Create: `tests/deploy/deploy_test.go`

**Step 1: Create deployment types**

Create file: `internal/deploy/template.go`

```go
package deploy

import "time"

type DeploymentTemplate struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Type        string    `json:"type"` // dev, test, staging, prod, custom
	Provider    string    `json:"provider"` // bare-metal, vm, cloud
	Config      string    `json:"config"` // JSON string
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
```

**Step 2: Create database operations**

Create file: `internal/deploy/template_db.go`

```go
package deploy

import (
	"database/sql"
	"errors"
	"time"
)

type TemplateDB struct {
	db *sql.DB
}

func NewTemplateDB(db *sql.DB) *TemplateDB {
	return &TemplateDB{db: db}
}

func (r *TemplateDB) Create(t *DeploymentTemplate) error {
	_, err := r.db.Exec(`
		INSERT INTO deployment_templates (id, name, description, type, provider, config, components, is_default, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, t.ID, t.Name, t.Description, t.Type, t.Provider, t.Config, t.Components, t.IsDefault, time.Now())
	return err
}

func (r *TemplateDB) GetByID(id string) (*DeploymentTemplate, error) {
	t := &DeploymentTemplate{}
	err := r.db.QueryRow(`
		SELECT id, name, description, type, provider, config, components, is_default, created_at
		FROM deployment_templates WHERE id = ?
	`, id).Scan(&t.ID, &t.Name, &t.Description, &t.Type, &t.Provider, &t.Config, &t.Components, &t.IsDefault, &t.CreatedAt)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (r *TemplateDB) List() ([]*DeploymentTemplate, error) {
	rows, err := r.db.Query(`
		SELECT id, name, description, type, provider, config, components, is_default, created_at
		FROM deployment_templates ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []*DeploymentTemplate
	for rows.Next() {
		t := &DeploymentTemplate{}
		err := rows.Scan(&t.ID, &t.Name, &t.Description, &t.Type, &t.Provider, &t.Config, &t.Components, &t.IsDefault, &t.CreatedAt)
		if err != nil {
			return nil, err
		}
		templates = append(templates, t)
	}
	return templates, nil
}

func (r *TemplateDB) Update(t *DeploymentTemplate) error {
	result, err := r.db.Exec(`
		UPDATE deployment_templates
		SET name = ?, description = ?, type = ?, provider = ?, config = ?, components = ?, is_default = ?
		WHERE id = ?
	`, t.Name, t.Description, t.Type, t.Provider, t.Config, t.Components, t.IsDefault, t.ID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("template not found")
	}
	return nil
}

func (r *TemplateDB) Delete(id string) error {
	result, err := r.db.Exec(`DELETE FROM deployment_templates WHERE id = ?`, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("template not found")
	}
	return nil
}
```

Create file: `internal/deploy/task_db.go`

```go
package deploy

import (
	"database/sql"
	"errors"
	"time"
)

type TaskDB struct {
	db *sql.DB
}

func NewTaskDB(db *sql.DB) *TaskDB {
	return &TaskDB{db: db}
}

func (r *TaskDB) Create(t *DeploymentTask) error {
	_, err := r.db.Exec(`
		INSERT INTO deployments (id, cluster_id, template_id, status, current_step, progress, error_message, created_by, started_at, finished_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, t.ID, t.ClusterID, t.TemplateID, t.Status, t.CurrentStep, t.Progress, t.ErrorMessage, t.CreatedBy, t.StartedAt, t.FinishedAt)
	return err
}

func (r *TaskDB) GetByID(id string) (*DeploymentTask, error) {
	t := &DeploymentTask{}
	err := r.db.QueryRow(`
		SELECT id, cluster_id, template_id, status, current_step, progress, error_message, created_by, started_at, finished_at
		FROM deployments WHERE id = ?
	`, id).Scan(&t.ID, &t.ClusterID, &t.TemplateID, &t.Status, &t.CurrentStep, &t.Progress, &t.ErrorMessage, &t.CreatedBy, &t.StartedAt, &t.FinishedAt)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (r *TaskDB) List() ([]*DeploymentTask, error) {
	rows, err := r.db.Query(`
		SELECT id, cluster_id, template_id, status, current_step, progress, error_message, created_by, started_at, finished_at
		FROM deployments ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*DeploymentTask
	for rows.Next() {
		t := &DeploymentTask{}
		err := rows.Scan(&t.ID, &t.ClusterID, &t.TemplateID, &t.Status, &t.CurrentStep, &t.Progress, &t.ErrorMessage, &t.CreatedBy, &t.StartedAt, &t.FinishedAt)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func (r *TaskDB) Update(t *DeploymentTask) error {
	result, err := r.db.Exec(`
		UPDATE deployments
		SET status = ?, current_step = ?, progress = ?, error_message = ?, started_at = ?, finished_at = ?
		WHERE id = ?
	`, t.Status, t.CurrentStep, t.Progress, t.ErrorMessage, t.StartedAt, t.FinishedAt, t.ID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("task not found")
	}
	return nil
}

func (r *TaskDB) UpdateProgress(id string, step string, progress int) error {
	_, err := r.db.Exec(`
		UPDATE deployments SET current_step = ?, progress = ? WHERE id = ?
	`, step, progress, id)
	return err
}

func (r *TaskDB) UpdateStatus(id string, status string, errMsg string) error {
	_, err := r.db.Exec(`
		UPDATE deployments SET status = ?, error_message = ? WHERE id = ?
	`, status, errMsg, id)
	return err
}
```

**Step 3: Write tests**

Create file: `tests/deploy/deploy_test.go`

```go
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
	
	// Test update progress
	err = taskDB.UpdateProgress("task-123", "Installing components", 50)
	if err != nil {
		t.Fatalf("Failed to update progress: %v", err)
	}
	
	updated, _ := taskDB.GetByID("task-123")
	if updated.Progress != 50 {
		t.Errorf("Expected progress 50, got %d", updated.Progress)
	}
}
```

**Step 4: Run tests**

```bash
go test ./tests/deploy/... -v
```

**Step 5: Commit**

```bash
git add internal/deploy tests/deploy
git commit -m "feat: implement deployment models and database operations"
```

---

## Task 2: Create Deployment API Handlers

**Files:**
- Create: `internal/api/handlers/deploy.go`
- Update: `internal/api/router.go`

**Step 1: Create deploy handlers**

Create file: `internal/api/handlers/deploy.go`

```go
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

// Template handlers

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

func (h *DeployHandler) DeleteTemplate(c *gin.Context) {
	id := c.Param("id")
	
	if err := h.templateDB.Delete(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "template deleted"})
}

// Task handlers

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
```

**Step 2: Update router**

Update file: `internal/api/router.go` to add deploy routes:

```go
// Add import
import "github.com/your-org/ai-k8s-ops/internal/deploy"

// Add in NewRouterWithDB function after cluster routes:
deployHandler := handlers.NewDeployHandler(
	deploy.NewTemplateDB(db),
	deploy.NewTaskDB(db),
)

// Template routes
templateGroup := v1.Group("/deploy/templates")
templateGroup.Use(middleware.AuthMiddleware(jwtSecret))
{
	templateGroup.POST("", deployHandler.CreateTemplate)
	templateGroup.GET("", deployHandler.ListTemplates)
	templateGroup.GET("/:id", deployHandler.GetTemplate)
	templateGroup.PUT("/:id", deployHandler.UpdateTemplate)
	templateGroup.DELETE("/:id", deployHandler.DeleteTemplate)
}

// Task routes
taskGroup := v1.Group("/deploy/tasks")
taskGroup.Use(middleware.AuthMiddleware(jwtSecret))
{
	taskGroup.POST("", deployHandler.CreateTask)
	taskGroup.GET("", deployHandler.ListTasks)
	taskGroup.GET("/:id", deployHandler.GetTask)
}
```

**Step 3: Commit**

```bash
git add internal/api
git commit -m "feat: implement deployment API handlers"
```

---

## Task 3: Create Deployment Frontend Types and Services

**Files:**
- Create: `frontend/src/types/deploy.ts`
- Create: `frontend/src/services/deploy.service.ts`

**Step 1: Create TypeScript types**

Create file: `frontend/src/types/deploy.ts`

```typescript
export interface DeploymentTemplate {
  id: string
  name: string
  description: string
  type: 'dev' | 'test' | 'staging' | 'prod' | 'custom'
  provider: 'bare-metal' | 'vm' | 'cloud'
  config: string
  components: string
  is_default: boolean
  created_at: string
}

export interface DeploymentTask {
  id: string
  cluster_id: string
  template_id: string
  status: 'pending' | 'running' | 'success' | 'failed' | 'rollback'
  current_step: string
  progress: number
  error_message?: string
  created_by: string
  started_at?: string
  finished_at?: string
}

export interface CreateTemplateRequest {
  name: string
  description?: string
  type: string
  provider: string
  config?: string
  components?: string
  is_default?: boolean
}

export interface CreateTaskRequest {
  cluster_id: string
  template_id: string
}

export interface TemplateListResponse {
  templates: DeploymentTemplate[]
}

export interface TaskListResponse {
  tasks: DeploymentTask[]
}
```

**Step 2: Create deploy service**

Create file: `frontend/src/services/deploy.service.ts`

```typescript
import api from './api'
import type {
  DeploymentTemplate,
  DeploymentTask,
  CreateTemplateRequest,
  CreateTaskRequest,
  TemplateListResponse,
  TaskListResponse,
} from '@/types/deploy'

export const deployService = {
  // Templates
  async listTemplates(): Promise<DeploymentTemplate[]> {
    const response = await api.get<TemplateListResponse>('/api/v1/deploy/templates')
    return response.data.templates
  },

  async getTemplate(id: string): Promise<DeploymentTemplate> {
    const response = await api.get<DeploymentTemplate>(`/api/v1/deploy/templates/${id}`)
    return response.data
  },

  async createTemplate(data: CreateTemplateRequest): Promise<DeploymentTemplate> {
    const response = await api.post<DeploymentTemplate>('/api/v1/deploy/templates', data)
    return response.data
  },

  async deleteTemplate(id: string): Promise<void> {
    await api.delete(`/api/v1/deploy/templates/${id}`)
  },

  // Tasks
  async listTasks(): Promise<DeploymentTask[]> {
    const response = await api.get<TaskListResponse>('/api/v1/deploy/tasks')
    return response.data.tasks
  },

  async getTask(id: string): Promise<DeploymentTask> {
    const response = await api.get<DeploymentTask>(`/api/v1/deploy/tasks/${id}`)
    return response.data
  },

  async createTask(data: CreateTaskRequest): Promise<DeploymentTask> {
    const response = await api.post<DeploymentTask>('/api/v1/deploy/tasks', data)
    return response.data
  },
}
```

**Step 3: Commit**

```bash
git add frontend/src/types frontend/src/services
git commit -m "feat: create deployment frontend types and services"
```

---

## Task 4: Create Deployment Template List Page

**Files:**
- Create: `frontend/src/pages/Deploy/Templates/index.tsx`

**Step 1: Create template list page**

Create file: `frontend/src/pages/Deploy/Templates/index.tsx`

```tsx
import { useState, useEffect } from 'react'
import { Table, Card, Button, Tag, Space, Popconfirm, message } from 'antd'
import { PlusOutlined, ReloadOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import { deployService } from '@/services/deploy.service'
import type { DeploymentTemplate } from '@/types/deploy'

const typeColors = {
  dev: 'blue',
  test: 'cyan',
  staging: 'purple',
  prod: 'red',
  custom: 'default',
}

export default function TemplateListPage() {
  const [templates, setTemplates] = useState<DeploymentTemplate[]>([])
  const [loading, setLoading] = useState(false)
  const navigate = useNavigate()

  const fetchTemplates = async () => {
    setLoading(true)
    try {
      const data = await deployService.listTemplates()
      setTemplates(data)
    } catch (error: any) {
      message.error('获取模板列表失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchTemplates()
  }, [])

  const handleDelete = async (id: string) => {
    try {
      await deployService.deleteTemplate(id)
      message.success('删除成功')
      fetchTemplates()
    } catch (error: any) {
      message.error('删除失败')
    }
  }

  const columns = [
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: '描述',
      dataIndex: 'description',
      key: 'description',
    },
    {
      title: '类型',
      dataIndex: 'type',
      key: 'type',
      render: (type: string) => <Tag color={typeColors[type as keyof typeof typeColors]}>{type}</Tag>,
    },
    {
      title: '提供商',
      dataIndex: 'provider',
      key: 'provider',
    },
    {
      title: '默认',
      dataIndex: 'is_default',
      key: 'is_default',
      render: (isDefault: boolean) => isDefault ? <Tag color="green">是</Tag> : <Tag>否</Tag>,
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (date: string) => new Date(date).toLocaleString(),
    },
    {
      title: '操作',
      key: 'action',
      render: (_: any, record: DeploymentTemplate) => (
        <Space>
          <Button type="link" onClick={() => navigate(`/deploy/templates/${record.id}`)}>
            查看
          </Button>
          <Popconfirm
            title="确定删除此模板？"
            onConfirm={() => handleDelete(record.id)}
            okText="确定"
            cancelText="取消"
          >
            <Button type="link" danger>
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  return (
    <Card
      title="部署模板"
      extra={
        <Space>
          <Button icon={<ReloadOutlined />} onClick={fetchTemplates}>
            刷新
          </Button>
          <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/deploy/templates/create')}>
            创建模板
          </Button>
        </Space>
      }
    >
      <Table
        columns={columns}
        dataSource={templates}
        rowKey="id"
        loading={loading}
        pagination={{ pageSize: 10 }}
      />
    </Card>
  )
}
```

**Step 2: Commit**

```bash
git add frontend/src/pages/Deploy
git commit -m "feat: create deployment template list page"
```

---

## Task 5: Create Deployment Task List Page

**Files:**
- Create: `frontend/src/pages/Deploy/Tasks/index.tsx`

**Step 1: Create task list page**

Create file: `frontend/src/pages/Deploy/Tasks/index.tsx`

```tsx
import { useState, useEffect } from 'react'
import { Table, Card, Button, Tag, Space, Progress, message } from 'antd'
import { ReloadOutlined, EyeOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import { deployService } from '@/services/deploy.service'
import type { DeploymentTask } from '@/types/deploy'

const statusColors = {
  pending: 'default',
  running: 'processing',
  success: 'success',
  failed: 'error',
  rollback: 'warning',
}

const statusTexts = {
  pending: '待执行',
  running: '执行中',
  success: '成功',
  failed: '失败',
  rollback: '已回滚',
}

export default function TaskListPage() {
  const [tasks, setTasks] = useState<DeploymentTask[]>([])
  const [loading, setLoading] = useState(false)
  const navigate = useNavigate()

  const fetchTasks = async () => {
    setLoading(true)
    try {
      const data = await deployService.listTasks()
      setTasks(data)
    } catch (error: any) {
      message.error('获取任务列表失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchTasks()
    // Auto refresh every 5 seconds
    const interval = setInterval(fetchTasks, 5000)
    return () => clearInterval(interval)
  }, [])

  const columns = [
    {
      title: '任务ID',
      dataIndex: 'id',
      key: 'id',
      width: 280,
      render: (id: string) => <span style={{ fontFamily: 'monospace' }}>{id.substring(0, 8)}...</span>,
    },
    {
      title: '集群ID',
      dataIndex: 'cluster_id',
      key: 'cluster_id',
      width: 280,
      render: (id: string) => <span style={{ fontFamily: 'monospace' }}>{id.substring(0, 8)}...</span>,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => <Tag color={statusColors[status as keyof typeof statusColors]}>{statusTexts[status as keyof typeof statusTexts]}</Tag>,
    },
    {
      title: '进度',
      dataIndex: 'progress',
      key: 'progress',
      width: 200,
      render: (progress: number, record: DeploymentTask) => (
        <Progress percent={progress} size="small" status={record.status === 'failed' ? 'exception' : 'active'} />
      ),
    },
    {
      title: '当前步骤',
      dataIndex: 'current_step',
      key: 'current_step',
      render: (step: string) => step || '-',
    },
    {
      title: '开始时间',
      dataIndex: 'started_at',
      key: 'started_at',
      render: (date: string) => date ? new Date(date).toLocaleString() : '-',
    },
    {
      title: '操作',
      key: 'action',
      render: (_: any, record: DeploymentTask) => (
        <Space>
          <Button type="link" icon={<EyeOutlined />} onClick={() => navigate(`/deploy/tasks/${record.id}`)}>
            详情
          </Button>
        </Space>
      ),
    },
  ]

  return (
    <Card
      title="部署任务"
      extra={
        <Button icon={<ReloadOutlined />} onClick={fetchTasks}>
          刷新
        </Button>
      }
    >
      <Table
        columns={columns}
        dataSource={tasks}
        rowKey="id"
        loading={loading}
        pagination={{ pageSize: 10 }}
      />
    </Card>
  )
}
```

**Step 2: Commit**

```bash
git add frontend/src/pages/Deploy
git commit -m "feat: create deployment task list page with auto-refresh"
```

---

## Task 6: Update Router with Deploy Routes

**Files:**
- Modify: `frontend/src/App.tsx`
- Create: `frontend/src/pages/Deploy/index.tsx`

**Step 1: Create deploy index page**

Create file: `frontend/src/pages/Deploy/index.tsx`

```tsx
import { Card, Row, Col, Button } from 'antd'
import { FileTextOutlined, PlayCircleOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'

export default function DeployPage() {
  const navigate = useNavigate()

  return (
    <div>
      <h2 style={{ marginBottom: 24 }}>部署中心</h2>
      
      <Row gutter={[16, 16]}>
        <Col xs={24} md={12}>
          <Card
            hoverable
            onClick={() => navigate('/deploy/templates')}
          >
            <div style={{ textAlign: 'center', padding: '20px 0' }}>
              <FileTextOutlined style={{ fontSize: 48, color: '#1890ff' }} />
              <h3 style={{ marginTop: 16 }}>部署模板</h3>
              <p style={{ color: '#666' }}>管理和配置部署模板</p>
              <Button type="primary">查看模板</Button>
            </div>
          </Card>
        </Col>
        
        <Col xs={24} md={12}>
          <Card
            hoverable
            onClick={() => navigate('/deploy/tasks')}
          >
            <div style={{ textAlign: 'center', padding: '20px 0' }}>
              <PlayCircleOutlined style={{ fontSize: 48, color: '#52c41a' }} />
              <h3 style={{ marginTop: 16 }}>部署任务</h3>
              <p style={{ color: '#666' }}>查看和管理部署任务</p>
              <Button type="primary">查看任务</Button>
            </div>
          </Card>
        </Col>
      </Row>
    </div>
  )
}
```

**Step 2: Update App.tsx routes**

Update file: `frontend/src/App.tsx` to add deploy routes:

```tsx
// Add import
import DeployPage from '@/pages/Deploy'
import TemplateListPage from '@/pages/Deploy/Templates'
import TaskListPage from '@/pages/Deploy/Tasks'

// Add routes inside the protected route:
<Route path="deploy" element={<DeployPage />}>
  <Route index element={<DeployPage />} />
</Route>
<Route path="deploy/templates" element={<TemplateListPage />} />
<Route path="deploy/tasks" element={<TaskListPage />} />
```

**Step 3: Test the application**

```bash
cd frontend
npm run dev
```

**Step 4: Commit**

```bash
git add frontend/src/App.tsx frontend/src/pages/Deploy
git commit -m "feat: integrate deployment pages into router"
```

---

## Task 7: Seed Default Deployment Templates

**Files:**
- Create: `cmd/seed/main.go`

**Step 1: Create seed script**

Create file: `cmd/seed/main.go`

```go
package main

import (
	"database/sql"
	"log"
	
	_ "github.com/mattn/go-sqlite3"
	"github.com/your-org/ai-k8s-ops/internal/deploy"
)

func main() {
	db, err := sql.Open("sqlite3", "data/ai-k8s-ops.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()
	
	templateDB := deploy.NewTemplateDB(db)
	
	templates := []*deploy.DeploymentTemplate{
		{
			ID:          "dev-template",
			Name:        "开发环境模板",
			Description: "单节点 K8S 集群，适合开发测试",
			Type:        "dev",
			Provider:    "bare-metal",
			Config:      `{"nodes": 1, "version": "v1.28.0", "network": "calico"}`,
			Components:  `["prometheus", "grafana"]`,
			IsDefault:   true,
		},
		{
			ID:          "test-template",
			Name:        "测试环境模板",
			Description: "3节点 K8S 集群，包含完整监控栈",
			Type:        "test",
			Provider:    "bare-metal",
			Config:      `{"nodes": 3, "version": "v1.28.0", "network": "calico"}`,
			Components:  `["prometheus", "grafana", "loki", "jaeger"]`,
			IsDefault:   false,
		},
		{
			ID:          "prod-template",
			Name:        "生产环境模板",
			Description: "高可用 K8S 集群，适合生产环境",
			Type:        "prod",
			Provider:    "bare-metal",
			Config:      `{"masters": 3, "workers": 3, "version": "v1.28.0", "network": "calico", "ha": true}`,
			Components:  `["prometheus", "grafana", "loki", "jaeger", "alertmanager"]`,
			IsDefault:   false,
		},
	}
	
	for _, t := range templates {
		err := templateDB.Create(t)
		if err != nil {
			log.Printf("Failed to create template %s: %v", t.Name, err)
		} else {
			log.Printf("Created template: %s", t.Name)
		}
	}
	
	log.Println("Seed completed!")
}
```

**Step 2: Run seed script**

```bash
go run cmd/seed/main.go
```

**Step 3: Commit**

```bash
git add cmd/seed
git commit -m "feat: add seed script for default deployment templates"
```

---

## Summary

This plan implements the deployment feature foundation:

✅ Deployment models and database operations
✅ Deployment API handlers (templates and tasks)
✅ Frontend types and services
✅ Template list page with CRUD
✅ Task list page with auto-refresh
✅ Router integration
✅ Default template seeding

**Next Phase**: Week 9-10 will implement AI interaction and basic monitoring.

---

**Plan complete and saved to `docs/plans/2026-03-30-phase1-deploy.md`**