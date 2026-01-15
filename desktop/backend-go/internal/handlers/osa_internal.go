package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rhl/businessos-backend/internal/integrations/osa"
)

// Internal OSA API endpoints for terminal containers
// These endpoints accept X-User-ID header for authentication from trusted internal sources

// HandleInternalGenerateApp - POST /api/internal/osa/generate
// Internal endpoint for terminal containers to generate apps
func (h *Handlers) HandleInternalGenerateApp(c *gin.Context) {
	if h.osaClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "OSA integration not enabled",
		})
		return
	}

	// Get user ID from header (set by container environment)
	userIDStr := c.GetHeader("X-User-ID")
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "X-User-ID header required"})
		return
	}

	// User ID is a string (e.g., "eIEMREsumBSwllpZvqh_gw"), not a UUID
	// We'll use it as-is for logging, and generate a UUID for OSA calls
	var userID uuid.UUID

	// Try to parse as UUID first (for backward compatibility)
	parsedUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		// Not a UUID, generate a deterministic UUID from the string ID
		// Using UUID v5 (SHA-1 hash) with a namespace
		namespace := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8") // DNS namespace
		userID = uuid.NewSHA1(namespace, []byte(userIDStr))
		log.Printf("[OSA Internal] Converted string user ID %s to UUID %s", userIDStr, userID)
	} else {
		userID = parsedUUID
	}

	var req GenerateAppRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse workspace ID
	var workspaceID uuid.UUID
	if req.WorkspaceID != "" {
		workspaceID, err = uuid.Parse(req.WorkspaceID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace_id"})
			return
		}
	} else {
		workspaceID = uuid.New() // TODO: Get user's default workspace
	}

	// Call OSA client
	osaReq := &osa.AppGenerationRequest{
		WorkspaceID: workspaceID,
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		Parameters:  req.Parameters,
	}

	// Default type if not specified
	if osaReq.Type == "" {
		osaReq.Type = "full-stack"
	}

	resp, err := h.osaClient.GenerateApp(c.Request.Context(), osaReq)
	if err != nil {
		log.Printf("[OSA Internal] App generation failed for user %s: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "App generation failed",
			"details": err.Error(),
		})
		return
	}

	log.Printf("[OSA Internal] App generation started for user %s: app_id=%s", userID, resp.AppID)
	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"app_id":       resp.AppID,
		"appId":        resp.AppID, // Include both formats for compatibility
		"workspace_id": resp.WorkspaceID,
		"status":       resp.Status,
		"message":      "App generation started",
	})
}

// HandleInternalGetAppStatus - GET /api/internal/osa/status/:app_id
func (h *Handlers) HandleInternalGetAppStatus(c *gin.Context) {
	if h.osaClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "OSA integration not enabled",
		})
		return
	}

	// Get user ID from header
	userIDStr := c.GetHeader("X-User-ID")
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "X-User-ID header required"})
		return
	}

	// Convert string user ID to UUID (same logic as generate endpoint)
	var userID uuid.UUID
	parsedUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		namespace := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
		userID = uuid.NewSHA1(namespace, []byte(userIDStr))
	} else {
		userID = parsedUUID
	}

	appID := c.Param("app_id")
	if appID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "app_id required"})
		return
	}

	status, err := h.osaClient.GetAppStatus(c.Request.Context(), appID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "App not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, status)
}

// HandleInternalListWorkspaces - GET /api/internal/osa/workspaces
func (h *Handlers) HandleInternalListWorkspaces(c *gin.Context) {
	if h.osaClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "OSA integration not enabled",
		})
		return
	}

	// Get user ID from header
	userIDStr := c.GetHeader("X-User-ID")
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "X-User-ID header required"})
		return
	}

	// Convert string user ID to UUID (same logic as generate endpoint)
	var userID uuid.UUID
	parsedUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		namespace := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
		userID = uuid.NewSHA1(namespace, []byte(userIDStr))
	} else {
		userID = parsedUUID
	}

	workspaces, err := h.osaClient.GetWorkspaces(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch workspaces",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"workspaces": workspaces,
	})
}

// HandleInternalOSAHealth - GET /api/internal/osa/health
func (h *Handlers) HandleInternalOSAHealth(c *gin.Context) {
	if h.osaClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"enabled": false,
			"status":  "disabled",
		})
		return
	}

	health, err := h.osaClient.HealthCheck(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"enabled": true,
			"status":  "unhealthy",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"enabled": true,
		"status":  health.Status,
		"version": health.Version,
	})
}
