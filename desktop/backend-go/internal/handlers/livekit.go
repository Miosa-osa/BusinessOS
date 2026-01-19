package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rhl/businessos-backend/internal/database/sqlc"
	"github.com/rhl/businessos-backend/internal/integrations/livekit"
)

// LiveKitHandler handles LiveKit voice agent operations
type LiveKitHandler struct {
	lkClient *livekit.ResilientClient
	queries  *sqlc.Queries
	logger   *slog.Logger
}

// NewLiveKitHandler creates a new LiveKit handler
func NewLiveKitHandler(lkClient *livekit.ResilientClient, queries *sqlc.Queries, logger *slog.Logger) *LiveKitHandler {
	if logger == nil {
		logger = slog.Default()
	}

	return &LiveKitHandler{
		lkClient: lkClient,
		queries:  queries,
		logger:   logger,
	}
}

// GenerateTokenRequest represents the request body for token generation
type GenerateTokenRequest struct {
	AgentRole string `json:"agent_role"` // "deepgram-agent" or "groq-agent"
	RoomName  string `json:"room_name"`  // Optional: will be auto-generated if empty
}

// GenerateTokenResponse represents the response for token generation
type GenerateTokenResponse struct {
	Token     string    `json:"token"`
	URL       string    `json:"url"`
	RoomName  string    `json:"room_name"`
	Identity  string    `json:"identity"`
	ExpiresAt time.Time `json:"expires_at"`
	SessionID string    `json:"session_id"`
}

// GenerateToken generates a LiveKit room token for authenticated users
// POST /api/livekit/token
func (h *LiveKitHandler) GenerateToken(c *gin.Context) {
	ctx := c.Request.Context()

	// Get user from context (set by auth middleware)
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID := userIDStr.(string)

	// Get user details from database
	user, err := h.queries.GetUserByID(ctx, userID)
	if err != nil {
		h.logger.Error("failed to get user", "error", err, "user_id", userID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user details"})
		return
	}

	// Get workspace ID (optional, from query or user default)
	workspaceIDStr := c.Query("workspace_id")
	var workspaceID uuid.UUID
	if workspaceIDStr != "" {
		workspaceID, err = uuid.Parse(workspaceIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workspace_id"})
			return
		}
	} else {
		// Use user's default workspace
		// For now, we'll generate a nil UUID - handler will handle it
		workspaceID = uuid.Nil
	}

	// Parse request body
	var req GenerateTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Use defaults if no body provided
		req.AgentRole = "deepgram-agent" // Default to Deepgram
	}

	// Validate agent role
	if req.AgentRole == "" {
		req.AgentRole = "deepgram-agent"
	}
	if req.AgentRole != "deepgram-agent" && req.AgentRole != "groq-agent" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid agent_role, must be 'deepgram-agent' or 'groq-agent'"})
		return
	}

	// Get user email (dereference pointer)
	userName := ""
	if user.Email != nil {
		userName = *user.Email
	}

	// Convert workspace UUID to string (workspaceIDStr already declared above)
	if workspaceID != uuid.Nil {
		workspaceIDStr = workspaceID.String()
	} else {
		workspaceIDStr = ""
	}

	tokenReq := &livekit.TokenRequest{
		UserID:      userID, // User ID as string
		UserName:    userName,
		WorkspaceID: workspaceIDStr, // Workspace ID as string
		RoomName:    req.RoomName,
		AgentRole:   req.AgentRole,
	}

	tokenResp, err := h.lkClient.GenerateRoomToken(ctx, tokenReq)
	if err != nil {
		h.logger.Error("failed to generate LiveKit token", "error", err, "user_id", userID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	// Create voice session in database
	_, err = h.queries.CreateVoiceSession(ctx, sqlc.CreateVoiceSessionParams{
		SessionID: tokenResp.SessionID,
		UserID:    userID,
		WorkspaceID: pgtype.UUID{
			Bytes: workspaceID,
			Valid: workspaceID != uuid.Nil,
		},
		AgentRole: &req.AgentRole,
	})
	if err != nil {
		h.logger.Error("failed to create voice session", "error", err, "session_id", tokenResp.SessionID)
		// Don't fail the request, token is still valid
	}

	h.logger.Info("generated LiveKit token",
		"user_id", userID,
		"session_id", tokenResp.SessionID,
		"room_name", tokenResp.RoomName,
		"agent_role", req.AgentRole,
	)

	c.JSON(http.StatusOK, GenerateTokenResponse{
		Token:     tokenResp.Token,
		URL:       tokenResp.URL,
		RoomName:  tokenResp.RoomName,
		Identity:  tokenResp.Identity,
		ExpiresAt: tokenResp.ExpiresAt,
		SessionID: tokenResp.SessionID,
	})
}

// GetSessionContext returns user context for Python voice agents
// GET /api/livekit/context/:session_id
// This endpoint is PUBLIC (no auth) - called by Python agents
func (h *LiveKitHandler) GetSessionContext(c *gin.Context) {
	ctx := c.Request.Context()
	sessionID := c.Param("session_id")

	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id is required"})
		return
	}

	// Get session from database
	session, err := h.queries.GetVoiceSessionBySessionID(ctx, sessionID)
	if err != nil {
		h.logger.Error("failed to get voice session", "error", err, "session_id", sessionID)
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	// Get user details
	user, err := h.queries.GetUserByID(ctx, session.UserID)
	if err != nil {
		h.logger.Error("failed to get user", "error", err, "user_id", session.UserID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user details"})
		return
	}

	// Get user facts for personalization
	userFacts, err := h.queries.GetUserFactsByUser(ctx, session.UserID)
	if err != nil {
		h.logger.Warn("failed to get user facts", "error", err, "user_id", session.UserID)
		userFacts = []sqlc.UserFact{} // Continue with empty facts
	}

	// Convert facts to preferences map
	preferences := make(map[string]string)
	for _, fact := range userFacts {
		preferences[fact.FactKey] = fact.FactValue
	}

	// TODO: Get conversation history from messages table
	conversationHistory := []livekit.ConversationMessage{}

	// TODO: Get RAG context from memories/documents
	ragContext := []livekit.RAGContextItem{}

	// Get user email
	userEmail := ""
	if user.Email != nil {
		userEmail = *user.Email
	}

	// Get workspace ID
	workspaceIDStr := ""
	if session.WorkspaceID.Valid {
		workspaceIDStr = fmt.Sprintf("%x-%x-%x-%x-%x",
			session.WorkspaceID.Bytes[0:4],
			session.WorkspaceID.Bytes[4:6],
			session.WorkspaceID.Bytes[6:8],
			session.WorkspaceID.Bytes[8:10],
			session.WorkspaceID.Bytes[10:16],
		)
	}

	// Build context response
	response := livekit.ContextResponse{
		UserID:              session.UserID,
		UserName:            userEmail,
		WorkspaceID:         workspaceIDStr,
		ConversationHistory: conversationHistory,
		RAGContext:          ragContext,
		UserPreferences:     preferences,
	}

	h.logger.Info("provided session context",
		"session_id", sessionID,
		"user_id", session.UserID,
		"preferences_count", len(preferences),
	)

	c.JSON(http.StatusOK, response)
}

// ListActiveSessions lists active voice sessions for the authenticated user
// GET /api/livekit/sessions
func (h *LiveKitHandler) ListActiveSessions(c *gin.Context) {
	ctx := c.Request.Context()

	// Get user from context
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID := userIDStr.(string)

	// List active sessions
	sessions, err := h.queries.ListActiveVoiceSessionsForUser(ctx, sqlc.ListActiveVoiceSessionsForUserParams{
		UserID: userID,
		Limit:  10, // Limit to 10 most recent
	})
	if err != nil {
		h.logger.Error("failed to list voice sessions", "error", err, "user_id", userID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list sessions"})
		return
	}

	// Convert to response format
	sessionStatuses := make([]livekit.SessionStatus, len(sessions))
	for i, session := range sessions {
		var endedAt *time.Time
		if session.EndedAt.Valid {
			endedAt = &session.EndedAt.Time
		}

		// Dereference pointers
		state := ""
		if session.State != nil {
			state = *session.State
		}
		agentRole := ""
		if session.AgentRole != nil {
			agentRole = *session.AgentRole
		}

		sessionStatuses[i] = livekit.SessionStatus{
			SessionID: session.SessionID,
			State:     state,
			RoomName:  fmt.Sprintf("ws_%s", session.SessionID[:8]), // Reconstruct room name
			AgentRole: agentRole,
			CreatedAt: session.CreatedAt.Time,
			EndedAt:   endedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"sessions": sessionStatuses,
		"total":    len(sessionStatuses),
	})
}

// DeleteRoom ends a voice session and cleans up the room
// DELETE /api/livekit/room/:room_name
func (h *LiveKitHandler) DeleteRoom(c *gin.Context) {
	ctx := c.Request.Context()
	roomName := c.Param("room_name")

	if roomName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "room_name is required"})
		return
	}

	// Get user from context
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID := userIDStr.(string)

	// Extract session ID from room name (format: ws_{workspace_id}_{user_id}_{timestamp})
	// For simplicity, we'll search by user_id and state
	sessions, err := h.queries.ListActiveVoiceSessionsForUser(ctx, sqlc.ListActiveVoiceSessionsForUserParams{
		UserID: userID,
		Limit:  100,
	})
	if err != nil {
		h.logger.Error("failed to list voice sessions", "error", err, "user_id", userID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to find session"})
		return
	}

	// Find matching session by room name pattern
	var sessionID string
	for _, session := range sessions {
		expectedRoomPrefix := fmt.Sprintf("ws_%s", session.SessionID[:8])
		if roomName == expectedRoomPrefix || roomName == session.SessionID {
			sessionID = session.SessionID
			break
		}
	}

	if sessionID == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	// Update session state to 'ended'
	endedState := "ended"
	err = h.queries.UpdateVoiceSessionState(ctx, sqlc.UpdateVoiceSessionStateParams{
		SessionID: sessionID,
		State:     &endedState,
	})
	if err != nil {
		h.logger.Error("failed to end voice session", "error", err, "session_id", sessionID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to end session"})
		return
	}

	h.logger.Info("ended voice session", "session_id", sessionID, "room_name", roomName)

	c.JSON(http.StatusOK, gin.H{
		"message":    "session ended",
		"session_id": sessionID,
	})
}
