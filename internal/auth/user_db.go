package auth

import (
	"database/sql"
	"errors"
	"time"
)

type UserDB struct {
	db *sql.DB
}

func NewUserDB(db *sql.DB) *UserDB {
	return &UserDB{db: db}
}

func (r *UserDB) Create(user *User) error {
	_, err := r.db.Exec(`
		INSERT INTO users (id, username, email, password_hash, role, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, user.ID, user.Username, user.Email, user.Password, user.Role, time.Now())

	return err
}

func (r *UserDB) GetByID(id string) (*User, error) {
	user := &User{}
	var lastLogin sql.NullTime

	err := r.db.QueryRow(`
		SELECT id, username, email, password_hash, role, created_at, last_login
		FROM users WHERE id = ?
	`, id).Scan(&user.ID, &user.Username, &user.Email, &user.Password,
		&user.Role, &user.CreatedAt, &lastLogin)

	if err != nil {
		return nil, err
	}

	if lastLogin.Valid {
		user.LastLogin = &lastLogin.Time
	}

	return user, nil
}

func (r *UserDB) GetByUsername(username string) (*User, error) {
	user := &User{}
	var lastLogin sql.NullTime

	err := r.db.QueryRow(`
		SELECT id, username, email, password_hash, role, created_at, last_login
		FROM users WHERE username = ?
	`, username).Scan(&user.ID, &user.Username, &user.Email, &user.Password,
		&user.Role, &user.CreatedAt, &lastLogin)

	if err != nil {
		return nil, err
	}

	if lastLogin.Valid {
		user.LastLogin = &lastLogin.Time
	}

	return user, nil
}

func (r *UserDB) GetByEmail(email string) (*User, error) {
	user := &User{}
	var lastLogin sql.NullTime

	err := r.db.QueryRow(`
		SELECT id, username, email, password_hash, role, created_at, last_login
		FROM users WHERE email = ?
	`, email).Scan(&user.ID, &user.Username, &user.Email, &user.Password,
		&user.Role, &user.CreatedAt, &lastLogin)

	if err != nil {
		return nil, err
	}

	if lastLogin.Valid {
		user.LastLogin = &lastLogin.Time
	}

	return user, nil
}

func (r *UserDB) List() ([]*User, error) {
	rows, err := r.db.Query(`
		SELECT id, username, email, password_hash, role, created_at, last_login
		FROM users ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		user := &User{}
		var lastLogin sql.NullTime

		err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.Password,
			&user.Role, &user.CreatedAt, &lastLogin)
		if err != nil {
			return nil, err
		}

		if lastLogin.Valid {
			user.LastLogin = &lastLogin.Time
		}

		users = append(users, user)
	}

	return users, nil
}

func (r *UserDB) Update(user *User) error {
	result, err := r.db.Exec(`
		UPDATE users
		SET email = ?, role = ?, password_hash = ?
		WHERE id = ?
	`, user.Email, user.Role, user.Password, user.ID)

	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errors.New("user not found")
	}

	return nil
}

func (r *UserDB) Delete(id string) error {
	result, err := r.db.Exec(`DELETE FROM users WHERE id = ?`, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errors.New("user not found")
	}

	return nil
}

func (r *UserDB) UpdateLastLogin(id string) error {
	_, err := r.db.Exec(`
		UPDATE users SET last_login = ? WHERE id = ?
	`, time.Now(), id)

	return err
}
