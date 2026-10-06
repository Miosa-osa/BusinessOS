package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rhl/businessos-backend/internal/database/sqlc"
	"github.com/rhl/businessos-backend/internal/middleware"
)

// SaveClientMessage persists turns produced by the user's local desktop runtimes.
func (h *ChatHandler) SaveClientMessage(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid conversation ID"})
		return
	}
	var request struct {
		Role        string `json:"role"`
		Content     string `json:"content"`
		Runtime     string `json:"runtime"`
		WorkspaceID string `json:"workspace_id"`
		Model       string `json:"model"`
	}
	if c.ShouldBindJSON(&request) != nil || strings.TrimSpace(request.Content) == "" || len(request.Content) > 1000000 || (request.Role != "user" && request.Role != "assistant") {
		c.JSON(400, gin.H{"error": "Invalid message"})
		return
	}
	switch request.Runtime {
	case "claude", "codex", "hermes", "ollama":
	default:
		c.JSON(400, gin.H{"error": "Invalid local runtime"})
		return
	}
	queries := sqlc.New(h.pool)
	convID := pgtype.UUID{Bytes: id, Valid: true}
	if _, err = queries.GetConversation(c.Request.Context(), sqlc.GetConversationParams{ID: convID, UserID: user.ID}); err != nil {
		c.JSON(404, gin.H{"error": "Conversation not found"})
		return
	}
	metadata, _ := json.Marshal(map[string]string{"runtime": request.Runtime, "workspace_id": request.WorkspaceID, "model": request.Model})
	message, err := queries.CreateMessage(c.Request.Context(), sqlc.CreateMessageParams{ConversationID: convID, Role: sqlc.Messagerole(strings.ToUpper(request.Role)), Content: request.Content, MessageMetadata: metadata})
	if err != nil {
		c.JSON(500, gin.H{"error": "Could not save message"})
		return
	}
	_, _ = h.pool.Exec(c.Request.Context(), "UPDATE conversations SET updated_at=NOW() WHERE id=$1 AND user_id=$2", id, user.ID)
	c.JSON(http.StatusCreated, TransformMessage(message))
}
