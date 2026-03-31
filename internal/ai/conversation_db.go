package ai

import (
	"database/sql"
	"time"
)

type ConversationDB struct {
	db *sql.DB
}

func NewConversationDB(db *sql.DB) *ConversationDB {
	return &ConversationDB{db: db}
}

func (r *ConversationDB) Create(c *Conversation) error {
	now := time.Now()
	_, err := r.db.Exec(`
		INSERT INTO conversations (id, user_id, cluster_id, title, context, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, c.ID, c.UserID, c.ClusterID, c.Title, c.Context, now, now)
	return err
}

func (r *ConversationDB) GetByID(id string) (*Conversation, error) {
	c := &Conversation{}
	err := r.db.QueryRow(`
		SELECT id, user_id, cluster_id, title, context, created_at, updated_at
		FROM conversations WHERE id = ?
	`, id).Scan(&c.ID, &c.UserID, &c.ClusterID, &c.Title, &c.Context, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *ConversationDB) ListByUser(userID string) ([]*Conversation, error) {
	rows, err := r.db.Query(`
		SELECT id, user_id, cluster_id, title, context, created_at, updated_at
		FROM conversations WHERE user_id = ? ORDER BY updated_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conversations []*Conversation
	for rows.Next() {
		c := &Conversation{}
		err := rows.Scan(&c.ID, &c.UserID, &c.ClusterID, &c.Title, &c.Context, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, err
		}
		conversations = append(conversations, c)
	}
	return conversations, nil
}

func (r *ConversationDB) Update(c *Conversation) error {
	_, err := r.db.Exec(`
		UPDATE conversations SET title = ?, context = ?, updated_at = ? WHERE id = ?
	`, c.Title, c.Context, time.Now(), c.ID)
	return err
}

func (r *ConversationDB) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM conversations WHERE id = ?`, id)
	return err
}
