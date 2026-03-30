package cluster

import (
	"database/sql"
	"errors"
	"time"
)

type ClusterDB struct {
	db *sql.DB
}

func NewClusterDB(db *sql.DB) *ClusterDB {
	return &ClusterDB{db: db}
}

func (r *ClusterDB) Create(c *Cluster) error {
	now := time.Now()
	_, err := r.db.Exec(`
		INSERT INTO clusters (id, name, description, environment, provider, version, api_server, kubeconfig, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, c.ID, c.Name, c.Description, c.Environment, c.Provider, c.Version,
		c.APIServer, c.Kubeconfig, c.Status, now, now)

	return err
}

func (r *ClusterDB) GetByID(id string) (*Cluster, error) {
	c := &Cluster{}
	err := r.db.QueryRow(`
		SELECT id, name, description, environment, provider, version, api_server, kubeconfig, status, created_at, updated_at
		FROM clusters WHERE id = ?
	`, id).Scan(&c.ID, &c.Name, &c.Description, &c.Environment, &c.Provider,
		&c.Version, &c.APIServer, &c.Kubeconfig, &c.Status, &c.CreatedAt, &c.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return c, nil
}

func (r *ClusterDB) GetByName(name string) (*Cluster, error) {
	c := &Cluster{}
	err := r.db.QueryRow(`
		SELECT id, name, description, environment, provider, version, api_server, kubeconfig, status, created_at, updated_at
		FROM clusters WHERE name = ?
	`, name).Scan(&c.ID, &c.Name, &c.Description, &c.Environment, &c.Provider,
		&c.Version, &c.APIServer, &c.Kubeconfig, &c.Status, &c.CreatedAt, &c.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return c, nil
}

func (r *ClusterDB) List() ([]*Cluster, error) {
	rows, err := r.db.Query(`
		SELECT id, name, description, environment, provider, version, api_server, kubeconfig, status, created_at, updated_at
		FROM clusters ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clusters []*Cluster
	for rows.Next() {
		c := &Cluster{}
		err := rows.Scan(&c.ID, &c.Name, &c.Description, &c.Environment, &c.Provider,
			&c.Version, &c.APIServer, &c.Kubeconfig, &c.Status, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, err
		}
		clusters = append(clusters, c)
	}

	return clusters, nil
}

func (r *ClusterDB) ListByEnvironment(env string) ([]*Cluster, error) {
	rows, err := r.db.Query(`
		SELECT id, name, description, environment, provider, version, api_server, kubeconfig, status, created_at, updated_at
		FROM clusters WHERE environment = ? ORDER BY created_at DESC
	`, env)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clusters []*Cluster
	for rows.Next() {
		c := &Cluster{}
		err := rows.Scan(&c.ID, &c.Name, &c.Description, &c.Environment, &c.Provider,
			&c.Version, &c.APIServer, &c.Kubeconfig, &c.Status, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, err
		}
		clusters = append(clusters, c)
	}

	return clusters, nil
}

func (r *ClusterDB) Update(c *Cluster) error {
	result, err := r.db.Exec(`
		UPDATE clusters
		SET name = ?, description = ?, environment = ?, provider = ?, version = ?,
		    api_server = ?, kubeconfig = ?, status = ?, updated_at = ?
		WHERE id = ?
	`, c.Name, c.Description, c.Environment, c.Provider, c.Version,
		c.APIServer, c.Kubeconfig, c.Status, time.Now(), c.ID)

	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errors.New("cluster not found")
	}

	return nil
}

func (r *ClusterDB) Delete(id string) error {
	result, err := r.db.Exec(`DELETE FROM clusters WHERE id = ?`, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errors.New("cluster not found")
	}

	return nil
}
