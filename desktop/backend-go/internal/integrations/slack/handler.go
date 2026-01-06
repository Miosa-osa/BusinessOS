package slack

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// Handler provides HTTP handlers for Slack integration routes.
type Handler struct {
	provider *Provider
	channels *ChannelService
	messages *MessageService
}

// NewHandler creates a new Slack integration handler.
func NewHandler(provider *Provider) *Handler {
	return &Handler{
		provider: provider,
		channels: NewChannelService(provider),
		messages: NewMessageService(provider),
	}
}

// RegisterRoutes registers all Slack integration routes.
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	// OAuth routes
	r.GET("/auth", h.GetAuthURL)
	r.GET("/callback", h.HandleCallback)
	r.POST("/disconnect", h.Disconnect)
	r.GET("/status", h.GetStatus)

	// Channel routes
	channels := r.Group("/channels")
	{
		channels.GET("", h.GetChannels)
		channels.POST("/sync", h.SyncChannels)
	}

	// Message routes
	messages := r.Group("/messages")
	{
		messages.GET("/:channel_id", h.GetMessages)
		messages.POST("/:channel_id", h.SendMessage)
		messages.POST("/:channel_id/sync", h.SyncMessages)
	}
}

// GetAuthURL returns the OAuth authorization URL.
func (h *Handler) GetAuthURL(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	state := generateState(userID)
	authURL := h.provider.GetAuthURL(state)

	c.JSON(http.StatusOK, gin.H{
		"auth_url": authURL,
	})
}

// HandleCallback handles the OAuth callback.
func (h *Handler) HandleCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing authorization code"})
		return
	}

	userID := extractUserIDFromState(state)
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state"})
		return
	}

	token, err := h.provider.ExchangeCode(c.Request.Context(), code)
	if err != nil {
		log.Printf("Failed to exchange code: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange code"})
		return
	}

	if err := h.provider.SaveToken(c.Request.Context(), userID, token); err != nil {
		log.Printf("Failed to save token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"team_name":    token.AccountName,
		"scopes":       token.Scopes,
	})
}

// Disconnect disconnects the Slack integration.
func (h *Handler) Disconnect(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if err := h.provider.Disconnect(c.Request.Context(), userID); err != nil {
		log.Printf("Failed to disconnect: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to disconnect"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// GetStatus returns the connection status.
func (h *Handler) GetStatus(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	status, err := h.provider.GetConnectionStatus(c.Request.Context(), userID)
	if err != nil {
		log.Printf("Failed to get status: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get status"})
		return
	}

	c.JSON(http.StatusOK, status)
}

// GetChannels returns the user's Slack channels.
func (h *Handler) GetChannels(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	includePrivate := c.Query("private") == "true"
	includeDMs := c.Query("dms") == "true"

	channels, err := h.channels.GetChannels(c.Request.Context(), userID, includePrivate, includeDMs)
	if err != nil {
		log.Printf("Failed to get channels: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get channels"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"channels": channels,
		"count":    len(channels),
	})
}

// SyncChannels syncs channels from Slack.
func (h *Handler) SyncChannels(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	result, err := h.channels.SyncChannels(c.Request.Context(), userID)
	if err != nil {
		log.Printf("Failed to sync channels: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sync channels"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetMessages returns messages for a channel.
func (h *Handler) GetMessages(c *gin.Context) {
	userID := c.GetString("user_id")
	channelID := c.Param("channel_id")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	messages, err := h.messages.GetMessages(c.Request.Context(), userID, channelID, limit, offset)
	if err != nil {
		log.Printf("Failed to get messages: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get messages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
		"count":    len(messages),
	})
}

// SendMessage sends a message to a channel.
func (h *Handler) SendMessage(c *gin.Context) {
	userID := c.GetString("user_id")
	channelID := c.Param("channel_id")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req struct {
		Text string `json:"text" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.messages.SendMessage(c.Request.Context(), userID, channelID, req.Text); err != nil {
		log.Printf("Failed to send message: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send message"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// SyncMessages syncs messages for a channel from Slack.
func (h *Handler) SyncMessages(c *gin.Context) {
	userID := c.GetString("user_id")
	channelID := c.Param("channel_id")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))

	result, err := h.messages.SyncMessages(c.Request.Context(), userID, channelID, limit)
	if err != nil {
		log.Printf("Failed to sync messages: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sync messages"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// Helper functions

func generateState(userID string) string {
	data := map[string]string{
		"user_id":   userID,
		"timestamp": time.Now().Format(time.RFC3339),
	}
	b, _ := json.Marshal(data)
	return string(b)
}

func extractUserIDFromState(state string) string {
	var data map[string]string
	if err := json.Unmarshal([]byte(state), &data); err != nil {
		return ""
	}
	return data["user_id"]
}
