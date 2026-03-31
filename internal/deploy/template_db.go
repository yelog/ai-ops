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
