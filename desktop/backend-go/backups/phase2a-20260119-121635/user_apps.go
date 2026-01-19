package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rhl/businessos-backend/internal/database/sqlc"
	"github.com/rhl/businessos-backend/internal/logging"
	"github.com/rhl/businessos-backend/internal/middleware"
	"github.com/rhl/businessos-backend/internal/utils"
	// NOTE: system package (native app detection) moved to feature/native-app-capture branch
)

// UserAppsHandler handles external web app management
type UserAppsHandler struct {
	queries        *sqlc.Queries
	faviconFetcher *utils.FaviconFetcher
}

// NewUserAppsHandler creates a new user apps handler
func NewUserAppsHandler(queries *sqlc.Queries) *UserAppsHandler {
	return &UserAppsHandler{
		queries:        queries,
		faviconFetcher: utils.NewFaviconFetcher(),
	}
}

// =============================================================================
// REQUEST/RESPONSE TYPES
// =============================================================================

type CreateUserAppRequest struct {
	WorkspaceID   string                 `json:"workspace_id" binding:"required"`
	Name          string                 `json:"name" binding:"required"`
	URL           string                 `json:"url" binding:"required"` // For native apps, this is the bundle ID
	Icon          string                 `json:"icon"`                   // Lucide icon name (deprecated, use logo_url)
	Color         string                 `json:"color"`                  // Hex color
	LogoURL       string                 `json:"logo_url"`               // URL to app logo/favicon (auto-fetched if not provided)
	Category      string                 `json:"category"`               // e.g., "productivity"
	Description   string                 `json:"description"`            // Optional
	IframeConfig  map[string]interface{} `json:"iframe_config"`          // Optional iframe settings
	OpenOnStartup bool                   `json:"open_on_startup"`        // Auto-open flag
	AppType       string                 `json:"app_type"`               // "web" or "native"
}

type UpdateUserAppRequest struct {
	Name          *string                `json:"name"`
	URL           *string                `json:"url"`
	Icon          *string                `json:"icon"`
	Color         *string                `json:"color"`
	LogoURL       *string                `json:"logo_url"` // URL to app logo/favicon
	Category      *string                `json:"category"`
	Description   *string                `json:"description"`
	PositionX     *int32                 `json:"position_x"`
	PositionY     *int32                 `json:"position_y"`
	PositionZ     *int32                 `json:"position_z"`
	IframeConfig  map[string]interface{} `json:"iframe_config"`
	IsActive      *bool                  `json:"is_active"`
	OpenOnStartup *bool                  `json:"open_on_startup"`
}

type UpdatePositionRequest struct {
	PositionX int32 `json:"position_x" binding:"required"`
	PositionY int32 `json:"position_y" binding:"required"`
	PositionZ int32 `json:"position_z" binding:"required"`
}

// =============================================================================
// HANDLERS
// =============================================================================

// ListUserApps godoc
// @Summary List all external apps for a workspace
// @Tags User Apps
// @Produce json
// @Param workspace_id query string true "Workspace ID"
// @Param include_inactive query bool false "Include inactive apps"
// @Success 200 {object} map[string]interface{}
// @Router /api/user-apps [get]
func (h *UserAppsHandler) ListUserApps(c *gin.Context) {
	workspaceID := c.Query("workspace_id")
	includeInactive := c.Query("include_inactive") == "true"

	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspace_id is required"})
		return
	}

	workspaceUUID, err := uuid.Parse(workspaceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace_id format"})
		return
	}

	// Convert uuid.UUID to pgtype.UUID
	pgWorkspaceID := pgtype.UUID{Bytes: workspaceUUID, Valid: true}

	var apps []sqlc.UserExternalApp
	if includeInactive {
		apps, err = h.queries.ListAllUserExternalApps(c.Request.Context(), pgWorkspaceID)
	} else {
		apps, err = h.queries.ListUserExternalApps(c.Request.Context(), pgWorkspaceID)
	}

	if err != nil {
		logging.Error("[UserApps] Failed to list apps: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch apps"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"apps": apps})
}

// GetUserApp godoc
// @Summary Get a specific external app
// @Tags User Apps
// @Produce json
// @Param id path string true "App ID"
// @Param workspace_id query string true "Workspace ID"
// @Success 200 {object} sqlc.UserExternalApp
// @Router /api/user-apps/{id} [get]
func (h *UserAppsHandler) GetUserApp(c *gin.Context) {
	appID := c.Param("id")
	workspaceID := c.Query("workspace_id")

	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspace_id is required"})
		return
	}

	appUUID, err := uuid.Parse(appID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid app ID format"})
		return
	}

	workspaceUUID, err := uuid.Parse(workspaceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace_id format"})
		return
	}

	// Convert to pgtype.UUID
	pgAppID := pgtype.UUID{Bytes: appUUID, Valid: true}
	pgWorkspaceID := pgtype.UUID{Bytes: workspaceUUID, Valid: true}

	app, err := h.queries.GetUserExternalApp(c.Request.Context(), sqlc.GetUserExternalAppParams{
		ID:          pgAppID,
		WorkspaceID: pgWorkspaceID,
	})

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "App not found"})
		return
	}

	if err != nil {
		logging.Error("[UserApps] Failed to get app: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch app"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"app": app})
}

// CreateUserApp godoc
// @Summary Create a new external app
// @Tags User Apps
// @Accept json
// @Produce json
// @Param app body CreateUserAppRequest true "App data"
// @Success 201 {object} sqlc.UserExternalApp
// @Router /api/user-apps [post]
func (h *UserAppsHandler) CreateUserApp(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID := user.ID

	var req CreateUserAppRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logging.Error("[UserApps] JSON binding error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	workspaceUUID, err := uuid.Parse(req.WorkspaceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace_id format"})
		return
	}

	// Set defaults
	if req.Icon == "" {
		req.Icon = "app-window"
	}
	if req.Color == "" {
		req.Color = "#6366F1"
	}
	if req.Category == "" {
		req.Category = "productivity"
	}
	if req.AppType == "" {
		req.AppType = "web"
	}

	// Marshal iframe_config to JSON
	iframeConfigJSON, err := json.Marshal(req.IframeConfig)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid iframe_config"})
		return
	}

	// Convert to pgtype.UUID
	pgWorkspaceID := pgtype.UUID{Bytes: workspaceUUID, Valid: true}

	// Prepare nullable fields as pointers
	var category, description *string
	if req.Category != "" {
		category = &req.Category
	}
	if req.Description != "" {
		description = &req.Description
	}
	openOnStartup := &req.OpenOnStartup

	// Resolve logo URL BEFORE creating the app (web apps only)
	// NOTE: Native app logo detection moved to feature/native-app-capture branch
	logoURL := req.LogoURL
	if logoURL == "" && req.AppType != "native" {
		// For web apps, fetch favicon from URL using Google's Favicon API
		fetchedLogoURL, err := h.faviconFetcher.FetchFaviconURL(req.URL)
		if err != nil {
			logging.Warn("[UserApps] Failed to fetch favicon for %s: %v", req.URL, err)
			// Continue without logo - not critical
		} else {
			logoURL = fetchedLogoURL
		}
	}

	// Prepare logo_url as pointer for nullable field
	var logoURLPtr *string
	if logoURL != "" {
		logoURLPtr = &logoURL
	}

	// Create app with logo_url included
	app, err := h.queries.CreateUserExternalApp(c.Request.Context(), sqlc.CreateUserExternalAppParams{
		UserID:        userID,
		WorkspaceID:   pgWorkspaceID,
		Name:          req.Name,
		Url:           req.URL,
		Icon:          req.Icon,
		Color:         req.Color,
		LogoUrl:       logoURLPtr,
		Category:      category,
		Description:   description,
		IframeConfig:  iframeConfigJSON,
		OpenOnStartup: openOnStartup,
		AppType:       req.AppType,
	})

	if err != nil {
		logging.Error("[UserApps] Failed to create app: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create app"})
		return
	}

	logging.Info("[UserApps] Created app: %s (%s) with logo: %v for workspace %s", app.Name, app.ID, logoURL != "", workspaceUUID)
	c.JSON(http.StatusCreated, gin.H{"app": app})
}

// UpdateUserApp godoc
// @Summary Update an existing external app
// @Tags User Apps
// @Accept json
// @Produce json
// @Param id path string true "App ID"
// @Param workspace_id query string true "Workspace ID"
// @Param app body UpdateUserAppRequest true "App data"
// @Success 200 {object} sqlc.UserExternalApp
// @Router /api/user-apps/{id} [put]
func (h *UserAppsHandler) UpdateUserApp(c *gin.Context) {
	appID := c.Param("id")
	workspaceID := c.Query("workspace_id")

	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspace_id is required"})
		return
	}

	appUUID, err := uuid.Parse(appID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid app ID format"})
		return
	}

	workspaceUUID, err := uuid.Parse(workspaceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace_id format"})
		return
	}

	var req UpdateUserAppRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert to pgtype.UUID
	pgAppID := pgtype.UUID{Bytes: appUUID, Valid: true}
	pgWorkspaceID := pgtype.UUID{Bytes: workspaceUUID, Valid: true}

	// Build update params
	params := sqlc.UpdateUserExternalAppParams{
		ID:          pgAppID,
		WorkspaceID: pgWorkspaceID,
		// All other fields are already pointers in the request struct
		Name:          req.Name,
		Url:           req.URL,
		Icon:          req.Icon,
		Color:         req.Color,
		LogoUrl:       req.LogoURL,
		Category:      req.Category,
		Description:   req.Description,
		PositionX:     req.PositionX,
		PositionY:     req.PositionY,
		PositionZ:     req.PositionZ,
		IsActive:      req.IsActive,
		OpenOnStartup: req.OpenOnStartup,
	}

	// Handle iframe_config separately
	if req.IframeConfig != nil {
		iframeConfigJSON, err := json.Marshal(req.IframeConfig)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid iframe_config"})
			return
		}
		params.IframeConfig = iframeConfigJSON
	}

	// Update app
	app, err := h.queries.UpdateUserExternalApp(c.Request.Context(), params)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "App not found"})
		return
	}

	if err != nil {
		logging.Error("[UserApps] Failed to update app: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update app"})
		return
	}

	logging.Info("[UserApps] Updated app: %s (%s)", app.Name, app.ID)
	c.JSON(http.StatusOK, gin.H{"app": app})
}

// DeleteUserApp godoc
// @Summary Delete an external app
// @Tags User Apps
// @Param id path string true "App ID"
// @Param workspace_id query string true "Workspace ID"
// @Success 204
// @Router /api/user-apps/{id} [delete]
func (h *UserAppsHandler) DeleteUserApp(c *gin.Context) {
	appID := c.Param("id")
	workspaceID := c.Query("workspace_id")

	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspace_id is required"})
		return
	}

	appUUID, err := uuid.Parse(appID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid app ID format"})
		return
	}

	workspaceUUID, err := uuid.Parse(workspaceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace_id format"})
		return
	}

	// Convert to pgtype.UUID
	pgAppID := pgtype.UUID{Bytes: appUUID, Valid: true}
	pgWorkspaceID := pgtype.UUID{Bytes: workspaceUUID, Valid: true}

	err = h.queries.DeleteUserExternalApp(c.Request.Context(), sqlc.DeleteUserExternalAppParams{
		ID:          pgAppID,
		WorkspaceID: pgWorkspaceID,
	})

	if err != nil {
		logging.Error("[UserApps] Failed to delete app: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete app"})
		return
	}

	logging.Info("[UserApps] Deleted app: %s", appID)
	c.JSON(http.StatusNoContent, nil)
}

// UpdateAppPosition godoc
// @Summary Update app position in 3D desktop
// @Tags User Apps
// @Accept json
// @Produce json
// @Param id path string true "App ID"
// @Param position body UpdatePositionRequest true "Position data"
// @Success 204
// @Router /api/user-apps/{id}/position [put]
func (h *UserAppsHandler) UpdateAppPosition(c *gin.Context) {
	appID := c.Param("id")

	appUUID, err := uuid.Parse(appID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid app ID format"})
		return
	}

	var req UpdatePositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert to pgtype.UUID
	pgAppID := pgtype.UUID{Bytes: appUUID, Valid: true}

	// Convert to pointers for SQLC
	posX := &req.PositionX
	posY := &req.PositionY
	posZ := &req.PositionZ

	err = h.queries.UpdateAppPosition(c.Request.Context(), sqlc.UpdateAppPositionParams{
		ID:        pgAppID,
		PositionX: posX,
		PositionY: posY,
		PositionZ: posZ,
	})

	if err != nil {
		logging.Error("[UserApps] Failed to update position: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update position"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// RecordAppOpened godoc
// @Summary Record that an app was opened (updates last_opened_at)
// @Tags User Apps
// @Param id path string true "App ID"
// @Success 204
// @Router /api/user-apps/{id}/open [post]
func (h *UserAppsHandler) RecordAppOpened(c *gin.Context) {
	appID := c.Param("id")

	appUUID, err := uuid.Parse(appID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid app ID format"})
		return
	}

	// Convert to pgtype.UUID
	pgAppID := pgtype.UUID{Bytes: appUUID, Valid: true}

	err = h.queries.RecordAppOpened(c.Request.Context(), pgAppID)
	if err != nil {
		logging.Error("[UserApps] Failed to record app opened: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record usage"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// GetStartupApps godoc
// @Summary Get all apps configured to open on startup
// @Tags User Apps
// @Produce json
// @Param workspace_id query string true "Workspace ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/user-apps/startup [get]
func (h *UserAppsHandler) GetStartupApps(c *gin.Context) {
	workspaceID := c.Query("workspace_id")

	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspace_id is required"})
		return
	}

	workspaceUUID, err := uuid.Parse(workspaceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace_id format"})
		return
	}

	// Convert to pgtype.UUID
	pgWorkspaceID := pgtype.UUID{Bytes: workspaceUUID, Valid: true}

	apps, err := h.queries.GetStartupApps(c.Request.Context(), pgWorkspaceID)
	if err != nil {
		logging.Error("[UserApps] Failed to get startup apps: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch startup apps"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"apps": apps})
}
