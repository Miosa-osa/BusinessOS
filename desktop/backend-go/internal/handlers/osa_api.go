package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rhl/businessos-backend/internal/integrations/osa"
	"github.com/rhl/businessos-backend/internal/middleware"
)

// GenerateAppRequest matches frontend expectations
type GenerateAppRequest struct {
	Name        string                 `json:"name" binding:"required"`
	Description string                 `json:"description" binding:"required"`
	Type        string                 `json:"type"` // "full-stack", "module", "tool"
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	WorkspaceID string                 `json:"workspace_id"`
}

// HandleGenerateApp - POST /api/osa/generate
func (h *Handlers) HandleGenerateApp(c *gin.Context) {
	if h.osaClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "OSA integration not enabled",
		})
		return
	}

	var req GenerateAppRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Parse user ID
	userID, err := uuid.Parse(user.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Parse workspace ID
	var workspaceID uuid.UUID
	if req.WorkspaceID != "" {
		var parseErr error
		workspaceID, parseErr = uuid.Parse(req.WorkspaceID)
		if parseErr != nil {
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "App generation failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"app_id":       resp.AppID,
		"workspace_id": resp.WorkspaceID,
		"status":       resp.Status,
		"message":      "App generation started. Use /api/osa/status/:app_id to track progress.",
	})
}

// HandleGetAppStatus - GET /api/osa/status/:app_id
func (h *Handlers) HandleGetAppStatus(c *gin.Context) {
	if h.osaClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "OSA integration not enabled",
		})
		return
	}

	appID := c.Param("app_id")
	if appID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "app_id required"})
		return
	}

	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, err := uuid.Parse(user.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
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

// HandleListWorkspaces - GET /api/osa/workspaces
func (h *Handlers) HandleListWorkspaces(c *gin.Context) {
	if h.osaClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "OSA integration not enabled",
		})
		return
	}

	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, err := uuid.Parse(user.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
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

// HandleOSAHealth - GET /api/osa/health
func (h *Handlers) HandleOSAHealth(c *gin.Context) {
	log.Printf("DEBUG: HandleOSAHealth called")
	if h.osaClient == nil {
		log.Printf("DEBUG: OSA client is nil")
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"enabled": false,
			"status":  "disabled",
		})
		return
	}

	log.Printf("DEBUG: Calling OSA health check")
	health, err := h.osaClient.HealthCheck(c.Request.Context())
	if err != nil {
		log.Printf("DEBUG: OSA health check failed: %v", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"enabled": true,
			"status":  "unhealthy",
			"error":   err.Error(),
		})
		return
	}

	log.Printf("DEBUG: OSA health check succeeded: %+v", health)
	c.JSON(http.StatusOK, gin.H{
		"enabled": true,
		"status":  health.Status,
		"version": health.Version,
	})
}
