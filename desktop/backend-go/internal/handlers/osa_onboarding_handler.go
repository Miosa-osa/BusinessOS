package handlers

import (
	"log/slog"
	"net/http"

	"github.com/rhl/businessos-backend/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type OSAOnboardingHandler struct {
	service *services.OSAOnboardingService
}

func NewOSAOnboardingHandler(service *services.OSAOnboardingService) *OSAOnboardingHandler {
	return &OSAOnboardingHandler{
		service: service,
	}
}

// POST /api/osa-onboarding/analyze
// Analyzes user data and generates personalized insights
func (h *OSAOnboardingHandler) AnalyzeUser(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req struct {
		Email             string `json:"email" binding:"required"`
		GmailConnected    bool   `json:"gmail_connected"`
		CalendarConnected bool   `json:"calendar_connected"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	analysisReq := &services.OnboardingAnalysisRequest{
		UserID:            userUUID,
		Email:             req.Email,
		GmailConnected:    req.GmailConnected,
		CalendarConnected: req.CalendarConnected,
	}

	analysis, err := h.service.AnalyzeUser(c.Request.Context(), analysisReq)
	if err != nil {
		slog.Error("Failed to analyze user", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to analyze user data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"analysis": analysis,
	})
}

// POST /api/osa-onboarding/generate-apps
// Generates 4 personalized starter apps
func (h *OSAOnboardingHandler) GenerateStarterApps(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req struct {
		WorkspaceID string                       `json:"workspace_id" binding:"required"`
		Analysis    *services.UserAnalysisResult `json:"analysis" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	workspaceUUID, err := uuid.Parse(req.WorkspaceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	apps, err := h.service.GenerateStarterApps(c.Request.Context(), userUUID, workspaceUUID, req.Analysis)
	if err != nil {
		slog.Error("Failed to generate starter apps", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate apps"})
		return
	}

	// Save the profile
	err = h.service.SaveOnboardingProfile(c.Request.Context(), userUUID, workspaceUUID, req.Analysis, apps)
	if err != nil {
		slog.Warn("Failed to save onboarding profile", "error", err)
		// Don't fail the request, just log
	}

	// Check if all apps are ready
	allReady, _ := h.service.GetAppGenerationStatus(c.Request.Context(), userUUID, apps)

	c.JSON(http.StatusOK, gin.H{
		"starter_apps":    apps,
		"ready_to_launch": allReady,
	})
}

// GET /api/osa-onboarding/apps-status
// Checks the status of starter app generation
func (h *OSAOnboardingHandler) CheckAppsStatus(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	workspaceIDStr := c.Query("workspace_id")
	if workspaceIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspace_id required"})
		return
	}

	workspaceUUID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	// Get saved profile
	analysis, apps, err := h.service.GetOnboardingProfile(c.Request.Context(), workspaceUUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Onboarding profile not found"})
		return
	}

	userUUID, _ := uuid.Parse(userID.(string))

	// Check current status
	allReady, err := h.service.GetAppGenerationStatus(c.Request.Context(), userUUID, apps)
	if err != nil {
		slog.Error("Failed to check apps status", "error", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"analysis":        analysis,
		"starter_apps":    apps,
		"ready_to_launch": allReady,
	})
}

// GET /api/osa-onboarding/profile
// Gets the saved onboarding profile
func (h *OSAOnboardingHandler) GetProfile(c *gin.Context) {
	workspaceIDStr := c.Query("workspace_id")
	if workspaceIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspace_id required"})
		return
	}

	workspaceUUID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	analysis, apps, err := h.service.GetOnboardingProfile(c.Request.Context(), workspaceUUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Onboarding profile not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"analysis":     analysis,
		"starter_apps": apps,
	})
}

// RegisterOSAOnboardingRoutes registers OSA onboarding routes
func RegisterOSAOnboardingRoutes(router *gin.RouterGroup, handler *OSAOnboardingHandler) {
	router.POST("/analyze", handler.AnalyzeUser)
	router.POST("/generate-apps", handler.GenerateStarterApps)
	router.GET("/apps-status", handler.CheckAppsStatus)
	router.GET("/profile", handler.GetProfile)
}
