package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/middleware"
	"github.com/rhl/businessos-backend/internal/services"
)

// WebPushHandler handles Web Push API endpoints
type WebPushHandler struct {
	svc *services.WebPushService
}

// NewWebPushHandler creates a new Web Push handler
func NewWebPushHandler(svc *services.WebPushService) *WebPushHandler {
	return &WebPushHandler{svc: svc}
}

// GetVAPIDPublicKey returns the VAPID public key for frontend subscription
// GET /api/notifications/push/vapid-public-key
func (h *WebPushHandler) GetVAPIDPublicKey(c *gin.Context) {
	publicKey := h.svc.GetPublicKey()
	if publicKey == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "Web Push not configured",
			"enabled": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"public_key": publicKey,
		"enabled":    true,
	})
}

// POST /api/notifications/push/subscribe
func (h *WebPushHandler) Subscribe(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if !h.svc.IsEnabled() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Web Push not configured"})
		return
	}

	var input services.SubscriptionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if input.Endpoint == "" || input.P256dh == "" || input.Auth == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "endpoint, p256dh, and auth are required"})
		return
	}

	// Get user agent from request if not provided
	if input.UserAgent == "" {
		input.UserAgent = c.GetHeader("User-Agent")
	}

	if err := h.svc.Subscribe(c.Request.Context(), user.ID, input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save subscription"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// POST /api/notifications/push/unsubscribe
func (h *WebPushHandler) Unsubscribe(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		Endpoint string `json:"endpoint" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "endpoint is required"})
		return
	}

	if err := h.svc.Unsubscribe(c.Request.Context(), user.ID, req.Endpoint); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove subscription"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// TestPush sends a test push notification to the current user
// POST /api/notifications/push/test
func (h *WebPushHandler) TestPush(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if !h.svc.IsEnabled() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Web Push not configured"})
		return
	}

	payload := services.PushPayload{
		Title:    "Test Notification",
		Body:     "Web Push is working! 🎉",
		Icon:     "/icon-192.png",
		Tag:      "test",
		Priority: "normal",
		Data: map[string]interface{}{
			"type": "test",
		},
	}

	if err := h.svc.SendToUser(c.Request.Context(), user.ID, payload); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send test notification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Test notification sent"})
}
