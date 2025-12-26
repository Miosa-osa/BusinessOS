package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/middleware"
	"github.com/rhl/businessos-backend/internal/services"
)


type SlackOAuthHandler struct {
	slackService *services.SlackService
}

// NewSlackOAuthHandler creates a new Slack OAuth handler
func NewSlackOAuthHandler(slackService *services.SlackService) *SlackOAuthHandler {
	return &SlackOAuthHandler{
		slackService: slackService,
	}
}

// initiates the Slack OAuth flow
func (h *SlackOAuthHandler) InitiateSlackAuth(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	// Generate random state for CSRF protection
	state := generateSlackRandomState()

	c.SetCookie("slack_oauth_state", state, 600, "/", "", false, true)
	c.SetCookie("slack_oauth_user", user.ID, 600, "/", "", false, true)

	authURL := h.slackService.GetAuthURL(state)

	c.JSON(http.StatusOK, gin.H{
		"auth_url": authURL,
	})
}

func (h *SlackOAuthHandler) HandleSlackCallback(c *gin.Context) {
	// Verify state
	state := c.Query("state")
	storedState, err := c.Cookie("slack_oauth_state")
	if err != nil || state != storedState {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state parameter"})
		return
	}

	userID, err := c.Cookie("slack_oauth_user")
	if err != nil || userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User session not found"})
		return
	}

	if errMsg := c.Query("error"); errMsg != "" {
		c.Redirect(http.StatusTemporaryRedirect, "/settings?slack_error="+errMsg)
		return
	}

	// Exchange code for tokens
	code := c.Query("code")
	response, err := h.slackService.ExchangeCode(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange code: " + err.Error()})
		return
	}

	// Save tokens to database
	if err := h.slackService.SaveToken(c.Request.Context(), userID, response); err != nil {
		// If token already exists, update it
		if err := h.slackService.UpdateToken(c.Request.Context(), userID, response); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save token"})
			return
		}
	}

	// Clear OAuth cookies
	c.SetCookie("slack_oauth_state", "", -1, "/", "", false, true)
	c.SetCookie("slack_oauth_user", "", -1, "/", "", false, true)

	// Redirect to settings page with success
	c.Redirect(http.StatusTemporaryRedirect, "/settings?slack_connected=true")
}

// GetSlackConnectionStatus returns the Slack connection status for a user
func (h *SlackOAuthHandler) GetSlackConnectionStatus(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	status, err := h.slackService.GetConnectionStatus(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"connected": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"connected":      true,
		"workspace_id":   status.WorkspaceID,
		"workspace_name": status.WorkspaceName,
		"connected_at":   status.CreatedAt,
	})
}

// DisconnectSlack disconnects the user's Slack workspace
func (h *SlackOAuthHandler) DisconnectSlack(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	// Delete OAuth tokens
	if err := h.slackService.DeleteToken(c.Request.Context(), user.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to disconnect Slack workspace"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Slack workspace disconnected"})
}

// Helper function for generating random state
func generateSlackRandomState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
