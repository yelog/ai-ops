package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/your-org/ai-k8s-ops/internal/ai"
	"github.com/your-org/ai-k8s-ops/internal/llm"
)

type AIHandler struct {
	conversationDB *ai.ConversationDB
	messageDB      *ai.MessageDB
	llmClient      *llm.Client
}

func NewAIHandler(conversationDB *ai.ConversationDB, messageDB *ai.MessageDB, llmClient *llm.Client) *AIHandler {
	return &AIHandler{
		conversationDB: conversationDB,
		messageDB:      messageDB,
		llmClient:      llmClient,
	}
}

func (h *AIHandler) CreateConversation(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req struct {
		Title     string `json:"title"`
		ClusterID string `json:"cluster_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conv := &ai.Conversation{
		ID:        uuid.New().String(),
		UserID:    userID.(string),
		ClusterID: req.ClusterID,
		Title:     req.Title,
	}

	if err := h.conversationDB.Create(conv); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create conversation"})
		return
	}

	c.JSON(http.StatusCreated, conv)
}

func (h *AIHandler) ListConversations(c *gin.Context) {
	userID, _ := c.Get("userID")

	conversations, err := h.conversationDB.ListByUser(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list conversations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"conversations": conversations})
}

func (h *AIHandler) GetConversation(c *gin.Context) {
	id := c.Param("id")

	conv, err := h.conversationDB.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "conversation not found"})
		return
	}

	messages, err := h.messageDB.ListByConversation(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get messages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"conversation": conv,
		"messages":     messages,
	})
}

func (h *AIHandler) DeleteConversation(c *gin.Context) {
	id := c.Param("id")

	if err := h.messageDB.DeleteByConversation(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete messages"})
		return
	}

	if err := h.conversationDB.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete conversation"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "conversation deleted"})
}

type ChatRequest struct {
	ConversationID string `json:"conversation_id"`
	Message        string `json:"message"`
}

func (h *AIHandler) Chat(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userMsg := &ai.Message{
		ID:             uuid.New().String(),
		ConversationID: req.ConversationID,
		Role:           "user",
		Content:        req.Message,
	}
	h.messageDB.Create(userMsg)

	messages, err := h.messageDB.ListByConversation(req.ConversationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get history"})
		return
	}

	var llmMessages []llm.Message
	llmMessages = append(llmMessages, llm.Message{
		Role:    "system",
		Content: ai.GetSystemPrompt(""),
	})
	for _, m := range messages {
		llmMessages = append(llmMessages, llm.Message{
			Role:    m.Role,
			Content: m.Content,
		})
	}

	response, err := h.llmClient.Chat(c.Request.Context(), llmMessages)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get AI response"})
		return
	}

	assistantMsg := &ai.Message{
		ID:             uuid.New().String(),
		ConversationID: req.ConversationID,
		Role:           "assistant",
		Content:        response,
	}
	h.messageDB.Create(assistantMsg)

	c.JSON(http.StatusOK, gin.H{
		"message": assistantMsg,
	})
}
