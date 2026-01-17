package handlers

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/database/sqlc"
)

// UserContextResponse represents user context for voice agent
type UserContextResponse struct {
	Name           string   `json:"name"`
	Email          string   `json:"email,omitempty"`
	Workspace      string   `json:"workspace,omitempty"`
	RecentActivity string   `json:"recent_activity,omitempty"`
	Preferences    []string `json:"preferences,omitempty"`
}

// HandleVoiceUserContext returns user context for the Python LiveKit voice agent
// GET /api/voice/user-context/:user_id
// This endpoint is called by the Python voice agent to fetch user context
func (h *Handlers) HandleVoiceUserContext(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id required"})
		return
	}

	slog.Info("[VoiceAgent] Fetching user context", "user_id", userID)

	// Create queries object
	queries := sqlc.New(h.pool)

	// Query user from database
	user, err := queries.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		slog.Warn("[VoiceAgent] User not found", "user_id", userID, "error", err)
		// Return empty context rather than error - agent can work without it
		c.JSON(http.StatusOK, UserContextResponse{
			Name: "User",
		})
		return
	}

	// Build context response
	response := UserContextResponse{
		Name: "User",
	}

	// Set name if available
	if user.Name != nil {
		response.Name = *user.Name
	}

	// Set email if available
	if user.Email != nil {
		response.Email = *user.Email
	}

	// TODO: Fetch workspace info - GetUserByID doesn't return workspace
	// Would need a separate query or join to get workspace name
	response.Workspace = ""

	// TODO: Add recent activity from logs/tasks
	// For now, just provide basic info
	response.RecentActivity = "Active today"

	slog.Info("[VoiceAgent] User context fetched",
		"user_id", userID,
		"name", response.Name,
		"workspace", response.Workspace)

	c.JSON(http.StatusOK, response)
}
