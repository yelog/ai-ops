package deploy

import (
	"database/sql"
	"errors"
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
		FROM deployments ORDER BY started_at DESC
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
