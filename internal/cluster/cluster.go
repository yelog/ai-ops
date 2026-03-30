package cluster

import "time"

type Cluster struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Environment string    `json:"environment"`
	Provider    string    `json:"provider"`
	Version     string    `json:"version"`
	APIServer   string    `json:"api_server"`
	Kubeconfig  string    `json:"-"` // Encrypted, never expose in JSON
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ClusterRepository interface {
	Create(cluster *Cluster) error
	GetByID(id string) (*Cluster, error)
	GetByName(name string) (*Cluster, error)
	List() ([]*Cluster, error)
	ListByEnvironment(env string) ([]*Cluster, error)
	Update(cluster *Cluster) error
	Delete(id string) error
}
