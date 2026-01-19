package livekit

import (
	"time"
)

// TokenRequest represents a request to generate a LiveKit room token
type TokenRequest struct {
	UserID      string `json:"user_id"` // User ID as string (not UUID)
	UserName    string `json:"user_name"`
	WorkspaceID string `json:"workspace_id"`         // Workspace ID as string (not UUID)
	RoomName    string `json:"room_name,omitempty"`  // Optional: auto-generated if empty
	AgentRole   string `json:"agent_role,omitempty"` // "deepgram-agent" or "groq-agent"
	Metadata    string `json:"metadata,omitempty"`   // Custom metadata for the participant
}

// TokenResponse represents a LiveKit room token response
type TokenResponse struct {
	Token     string    `json:"token"`
	URL       string    `json:"url"`
	RoomName  string    `json:"room_name"`
	Identity  string    `json:"identity"`
	ExpiresAt time.Time `json:"expires_at"`
	SessionID string    `json:"session_id"` // Corresponding voice_sessions.session_id
}

// ContextRequest represents a request for user context from Python agents
type ContextRequest struct {
	SessionID string `json:"session_id"`
	UserID    string `json:"user_id,omitempty"`
}

// ConversationMessage represents a message in conversation history
type ConversationMessage struct {
	Role      string    `json:"role"` // "user" or "assistant"
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// RAGContextItem represents a context item from RAG retrieval
type RAGContextItem struct {
	Content   string    `json:"content"`
	Source    string    `json:"source"`
	Relevance float64   `json:"relevance"`
	Timestamp time.Time `json:"timestamp,omitempty"`
}

// ContextResponse represents the context provided to Python agents
type ContextResponse struct {
	UserID              string                 `json:"user_id"`
	UserName            string                 `json:"user_name"`
	WorkspaceID         string                 `json:"workspace_id"`
	ConversationHistory []ConversationMessage  `json:"conversation_history"`
	RAGContext          []RAGContextItem       `json:"rag_context"`
	UserPreferences     map[string]string      `json:"user_preferences"`
	Metadata            map[string]interface{} `json:"metadata,omitempty"`
}

// SessionStatus represents the status of a voice session
type SessionStatus struct {
	SessionID string     `json:"session_id"`
	State     string     `json:"state"` // "active", "ended"
	RoomName  string     `json:"room_name"`
	AgentRole string     `json:"agent_role"`
	CreatedAt time.Time  `json:"created_at"`
	EndedAt   *time.Time `json:"ended_at,omitempty"`
}

// ErrorResponse represents an error from LiveKit or our service
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}
