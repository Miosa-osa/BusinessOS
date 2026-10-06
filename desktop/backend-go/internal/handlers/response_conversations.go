package handlers

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/rhl/businessos-backend/internal/database/sqlc"
)

// Message response transformation
type MessageResponse struct {
	Metadata       json.RawMessage `json:"metadata,omitempty"`
	ID             string          `json:"id"`
	ConversationID string          `json:"conversation_id"`
	Role           string          `json:"role"`
	Content        string          `json:"content"`
	CreatedAt      string          `json:"created_at"`
}

func TransformMessage(m sqlc.Message, location ...*time.Location) MessageResponse {
	return MessageResponse{
		ID:             pgtypeUUIDToStringRequired(m.ID),
		Metadata:       m.MessageMetadata,
		ConversationID: pgtypeUUIDToStringRequired(m.ConversationID),
		Role:           strings.ToLower(string(m.Role)),
		Content:        m.Content,
		CreatedAt:      formatConversationTime(m.CreatedAt.Time, location...),
	}
}

func TransformMessages(messages []sqlc.Message, location ...*time.Location) []MessageResponse {
	result := make([]MessageResponse, len(messages))
	for i, m := range messages {
		result[i] = TransformMessage(m, location...)
	}
	return result
}

// Conversation response transformation
type ConversationResponse struct {
	ID           string  `json:"id"`
	UserID       string  `json:"user_id"`
	Title        string  `json:"title"`
	ContextID    *string `json:"context_id"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
	MessageCount int64   `json:"message_count"`
}

func TransformConversation(c sqlc.Conversation, location ...*time.Location) ConversationResponse {
	title := "New Conversation"
	if c.Title != nil {
		title = *c.Title
	}

	return ConversationResponse{
		ID:        pgtypeUUIDToStringRequired(c.ID),
		UserID:    c.UserID,
		Title:     title,
		ContextID: pgtypeUUIDToString(c.ContextID),
		CreatedAt: formatConversationTime(c.CreatedAt.Time, location...),
		UpdatedAt: formatConversationTime(c.UpdatedAt.Time, location...),
	}
}

func TransformConversationListRow(c sqlc.ListConversationsRow, location ...*time.Location) ConversationResponse {
	title := "New Conversation"
	if c.Title != nil {
		title = *c.Title
	}

	return ConversationResponse{
		ID:           pgtypeUUIDToStringRequired(c.ID),
		UserID:       c.UserID,
		Title:        title,
		ContextID:    pgtypeUUIDToString(c.ContextID),
		CreatedAt:    formatConversationTime(c.CreatedAt.Time, location...),
		UpdatedAt:    formatConversationTime(c.UpdatedAt.Time, location...),
		MessageCount: c.MessageCount,
	}
}

func TransformConversationListRows(conversations []sqlc.ListConversationsRow, location ...*time.Location) []ConversationResponse {
	result := make([]ConversationResponse, len(conversations))
	for i, c := range conversations {
		result[i] = TransformConversationListRow(c, location...)
	}
	return result
}

func TransformConversationByContextRow(c sqlc.ListConversationsByContextRow, location ...*time.Location) ConversationResponse {
	title := "New Conversation"
	if c.Title != nil {
		title = *c.Title
	}

	return ConversationResponse{
		ID:           pgtypeUUIDToStringRequired(c.ID),
		UserID:       c.UserID,
		Title:        title,
		ContextID:    pgtypeUUIDToString(c.ContextID),
		CreatedAt:    formatConversationTime(c.CreatedAt.Time, location...),
		UpdatedAt:    formatConversationTime(c.UpdatedAt.Time, location...),
		MessageCount: c.MessageCount,
	}
}

func TransformConversationsByContextRows(conversations []sqlc.ListConversationsByContextRow, location ...*time.Location) []ConversationResponse {
	result := make([]ConversationResponse, len(conversations))
	for i, c := range conversations {
		result[i] = TransformConversationByContextRow(c, location...)
	}
	return result
}

// These columns are timestamp without time zone. PostgreSQL writes NOW() in
// the connection time zone, while pgx decodes the wall clock as UTC by default.
func formatConversationTime(value time.Time, locations ...*time.Location) string {
	if len(locations) > 0 && locations[0] != nil {
		value = time.Date(value.Year(), value.Month(), value.Day(), value.Hour(), value.Minute(), value.Second(), value.Nanosecond(), locations[0])
	}
	return value.Format(time.RFC3339)
}
