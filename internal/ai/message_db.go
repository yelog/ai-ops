package ai

import (
	"database/sql"
	"time"
)

type MessageDB struct {
	db *sql.DB
}

func NewMessageDB(db *sql.DB) *MessageDB {
	return &MessageDB{db: db}
}

func (r *MessageDB) Create(m *Message) error {
	_, err := r.db.Exec(`
		INSERT INTO messages (id, conversation_id, role, content, tokens, metadata, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, m.ID, m.ConversationID, m.Role, m.Content, m.Tokens, m.Metadata, time.Now())
	return err
}

func (r *MessageDB) ListByConversation(conversationID string) ([]*Message, error) {
	rows, err := r.db.Query(`
		SELECT id, conversation_id, role, content, tokens, metadata, created_at
		FROM messages WHERE conversation_id = ? ORDER BY created_at ASC
	`, conversationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*Message
	for rows.Next() {
		m := &Message{}
		err := rows.Scan(&m.ID, &m.ConversationID, &m.Role, &m.Content, &m.Tokens, &m.Metadata, &m.CreatedAt)
		if err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, nil
}

func (r *MessageDB) DeleteByConversation(conversationID string) error {
	_, err := r.db.Exec(`DELETE FROM messages WHERE conversation_id = ?`, conversationID)
	return err
}
