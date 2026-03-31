package ai

import "time"

type Conversation struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	ClusterID string    `json:"cluster_id,omitempty"`
	Title     string    `json:"title"`
	Context   string    `json:"context"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Message struct {
	ID             string    `json:"id"`
	ConversationID string    `json:"conversation_id"`
	Role           string    `json:"role"`
	Content        string    `json:"content"`
	Tokens         int       `json:"tokens"`
	Metadata       string    `json:"metadata"`
	CreatedAt      time.Time `json:"created_at"`
}
