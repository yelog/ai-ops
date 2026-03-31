package offline

import (
	"database/sql"
	"errors"
	"time"
)

type PackageDB struct {
	db *sql.DB
}

func NewPackageDB(db *sql.DB) *PackageDB {
	return &PackageDB{db: db}
}

func (r *PackageDB) Create(p *OfflinePackage) error {
	_, err := r.db.Exec(`
		INSERT INTO offline_packages (id, name, version, os_list, modules, status, size, checksum, storage_path, error_message, created_by, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, p.ID, p.Name, p.Version, p.OSList, p.Modules, p.Status, p.Size, p.Checksum, p.StoragePath, p.ErrorMessage, p.CreatedBy, time.Now())
	return err
}

func (r *PackageDB) GetByID(id string) (*OfflinePackage, error) {
	p := &OfflinePackage{}
	err := r.db.QueryRow(`
		SELECT id, name, version, os_list, modules, status, size, checksum, storage_path, error_message, created_by, created_at
		FROM offline_packages WHERE id = ?
	`, id).Scan(&p.ID, &p.Name, &p.Version, &p.OSList, &p.Modules, &p.Status, &p.Size, &p.Checksum, &p.StoragePath, &p.ErrorMessage, &p.CreatedBy, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *PackageDB) List() ([]*OfflinePackage, error) {
	rows, err := r.db.Query(`
		SELECT id, name, version, os_list, modules, status, size, checksum, storage_path, error_message, created_by, created_at
		FROM offline_packages ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var packages []*OfflinePackage
	for rows.Next() {
		p := &OfflinePackage{}
		err := rows.Scan(&p.ID, &p.Name, &p.Version, &p.OSList, &p.Modules, &p.Status, &p.Size, &p.Checksum, &p.StoragePath, &p.ErrorMessage, &p.CreatedBy, &p.CreatedAt)
		if err != nil {
			return nil, err
		}
		packages = append(packages, p)
	}
	return packages, nil
}

func (r *PackageDB) UpdateStatus(id string, status string, errMsg string) error {
	result, err := r.db.Exec(`
		UPDATE offline_packages SET status = ?, error_message = ? WHERE id = ?
	`, status, errMsg, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("package not found")
	}
	return nil
}

func (r *PackageDB) UpdateComplete(id string, size int64, checksum string, storagePath string) error {
	result, err := r.db.Exec(`
		UPDATE offline_packages SET status = 'ready', size = ?, checksum = ?, storage_path = ? WHERE id = ?
	`, size, checksum, storagePath, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("package not found")
	}
	return nil
}

func (r *PackageDB) Delete(id string) error {
	result, err := r.db.Exec(`DELETE FROM offline_packages WHERE id = ?`, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("package not found")
	}
	return nil
}
