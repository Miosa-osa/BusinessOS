package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rhl/businessos-backend/internal/database/sqlc"
	"github.com/rhl/businessos-backend/internal/middleware"
	"github.com/rhl/businessos-backend/internal/services"
)

// ============================================================================
// DASHBOARD CRUD OPERATIONS
// ============================================================================

func (h *Handlers) ListUserDashboards(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	queries := sqlc.New(h.pool)
	dashboards, err := queries.ListUserDashboards(c.Request.Context(), user.ID)
	if err != nil {
		log.Printf("ListUserDashboards error for user %s: %v", user.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list dashboards"})
		return
	}

	c.JSON(http.StatusOK, transformDashboards(dashboards))
}

func (h *Handlers) GetUserDashboard(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	dashboardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid dashboard ID"})
		return
	}

	queries := sqlc.New(h.pool)
	dashboard, err := queries.GetDashboard(c.Request.Context(), sqlc.GetDashboardParams{
		ID:     pgtype.UUID{Bytes: dashboardID, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Dashboard not found"})
		return
	}

	c.JSON(http.StatusOK, transformDashboard(dashboard))
}

func (h *Handlers) CreateUserDashboard(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var req struct {
		Name        string          `json:"name" binding:"required"`
		Description *string         `json:"description"`
		Layout      json.RawMessage `json:"layout"`
		Visibility  *string         `json:"visibility"`
		WorkspaceID *string         `json:"workspace_id"`
		CreatedVia  *string         `json:"created_via"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Default layout to empty array if not provided
	layout := req.Layout
	if layout == nil {
		layout = []byte("[]")
	}

	// Parse workspace ID if provided
	var workspaceID pgtype.UUID
	if req.WorkspaceID != nil {
		if id, err := uuid.Parse(*req.WorkspaceID); err == nil {
			workspaceID = pgtype.UUID{Bytes: id, Valid: true}
		}
	}

	// Default visibility
	visibility := "private"
	if req.Visibility != nil {
		visibility = *req.Visibility
	}

	// Default created_via
	createdVia := "manual"
	if req.CreatedVia != nil {
		createdVia = *req.CreatedVia
	}

	queries := sqlc.New(h.pool)
	dashboard, err := queries.CreateDashboard(c.Request.Context(), sqlc.CreateDashboardParams{
		UserID:      user.ID,
		WorkspaceID: workspaceID,
		Name:        req.Name,
		Description: req.Description,
		Layout:      layout,
		Visibility:  &visibility,
		CreatedVia:  &createdVia,
	})
	if err != nil {
		log.Printf("CreateUserDashboard error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create dashboard"})
		return
	}

	c.JSON(http.StatusCreated, transformDashboard(dashboard))
}

func (h *Handlers) UpdateUserDashboard(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	dashboardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid dashboard ID"})
		return
	}

	var req struct {
		Name        *string         `json:"name"`
		Description *string         `json:"description"`
		Layout      json.RawMessage `json:"layout"`
		Visibility  *string         `json:"visibility"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	queries := sqlc.New(h.pool)
	dashboard, err := queries.UpdateDashboard(c.Request.Context(), sqlc.UpdateDashboardParams{
		ID:          pgtype.UUID{Bytes: dashboardID, Valid: true},
		UserID:      user.ID,
		Name:        req.Name,
		Description: req.Description,
		Layout:      req.Layout,
		Visibility:  req.Visibility,
	})
	if err != nil {
		log.Printf("UpdateUserDashboard error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update dashboard"})
		return
	}

	c.JSON(http.StatusOK, transformDashboard(dashboard))
}

// UpdateDashboardLayout updates only the layout of a dashboard (used by agent)
func (h *Handlers) UpdateDashboardLayout(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	dashboardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid dashboard ID"})
		return
	}

	var req struct {
		Layout json.RawMessage `json:"layout" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	queries := sqlc.New(h.pool)
	dashboard, err := queries.UpdateDashboardLayout(c.Request.Context(), sqlc.UpdateDashboardLayoutParams{
		ID:     pgtype.UUID{Bytes: dashboardID, Valid: true},
		UserID: user.ID,
		Layout: req.Layout,
	})
	if err != nil {
		log.Printf("UpdateDashboardLayout error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update layout"})
		return
	}

	// Broadcast SSE event for real-time sync across tabs/devices
	if h.notificationService != nil {
		h.notificationService.SSE().SendToUser(user.ID, services.SSEEvent{
			Type: "dashboard.updated",
			Data: map[string]interface{}{
				"dashboard_id": dashboardID.String(),
			},
		})
	}

	c.JSON(http.StatusOK, transformDashboard(dashboard))
}

func (h *Handlers) DeleteUserDashboard(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	dashboardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid dashboard ID"})
		return
	}

	queries := sqlc.New(h.pool)
	err = queries.DeleteDashboard(c.Request.Context(), sqlc.DeleteDashboardParams{
		ID:     pgtype.UUID{Bytes: dashboardID, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		log.Printf("DeleteUserDashboard error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete dashboard"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Dashboard deleted"})
}

// DuplicateUserDashboard creates a copy of a dashboard
func (h *Handlers) DuplicateUserDashboard(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	dashboardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid dashboard ID"})
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	// Name is optional, will default to "Copy of X"
	_ = c.ShouldBindJSON(&req)

	// Get original dashboard to build name if not provided
	queries := sqlc.New(h.pool)
	original, err := queries.GetDashboardByID(c.Request.Context(), pgtype.UUID{Bytes: dashboardID, Valid: true})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Dashboard not found"})
		return
	}

	name := req.Name
	if name == "" {
		name = "Copy of " + original.Name
	}

	dashboard, err := queries.DuplicateDashboard(c.Request.Context(), sqlc.DuplicateDashboardParams{
		ID:     pgtype.UUID{Bytes: dashboardID, Valid: true},
		UserID: user.ID,
		Name:   name,
	})
	if err != nil {
		log.Printf("DuplicateUserDashboard error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to duplicate dashboard"})
		return
	}

	c.JSON(http.StatusCreated, transformDashboard(sqlc.UserDashboard(dashboard)))
}

func (h *Handlers) SetDefaultUserDashboard(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	dashboardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid dashboard ID"})
		return
	}

	queries := sqlc.New(h.pool)

	// First clear existing default
	err = queries.ClearDefaultDashboard(c.Request.Context(), user.ID)
	if err != nil {
		log.Printf("ClearDefaultDashboard error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update default dashboard"})
		return
	}

	// Set new default
	err = queries.SetDefaultDashboard(c.Request.Context(), sqlc.SetDefaultDashboardParams{
		ID:     pgtype.UUID{Bytes: dashboardID, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		log.Printf("SetDefaultDashboard error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set default dashboard"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Default dashboard updated"})
}

// ============================================================================
// SHARING HANDLERS
// ============================================================================

// ShareUserDashboard updates sharing settings and generates a share token if needed
func (h *Handlers) ShareUserDashboard(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	dashboardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid dashboard ID"})
		return
	}

	var req struct {
		Visibility string `json:"visibility" binding:"required"` // private, workspace, public_link
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate visibility value
	if req.Visibility != "private" && req.Visibility != "workspace" && req.Visibility != "public_link" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid visibility value"})
		return
	}

	// Generate share token for public links
	var shareToken *string
	if req.Visibility == "public_link" {
		token := generateShareToken()
		shareToken = &token
	}

	queries := sqlc.New(h.pool)
	dashboard, err := queries.UpdateShareToken(c.Request.Context(), sqlc.UpdateShareTokenParams{
		ID:         pgtype.UUID{Bytes: dashboardID, Valid: true},
		UserID:     user.ID,
		Visibility: &req.Visibility,
		ShareToken: shareToken,
	})
	if err != nil {
		log.Printf("ShareUserDashboard error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update sharing settings"})
		return
	}

	c.JSON(http.StatusOK, transformDashboard(dashboard))
}

func (h *Handlers) GetSharedDashboard(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Share token required"})
		return
	}

	queries := sqlc.New(h.pool)
	dashboard, err := queries.GetDashboardByShareToken(c.Request.Context(), &token)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Dashboard not found or not publicly shared"})
		return
	}

	c.JSON(http.StatusOK, transformDashboard(dashboard))
}

// ============================================================================
// WIDGET TYPE HANDLERS
// ============================================================================

// ListWidgetTypes returns all available widget types
func (h *Handlers) ListWidgetTypes(c *gin.Context) {
	queries := sqlc.New(h.pool)
	widgets, err := queries.ListWidgetTypes(c.Request.Context())
	if err != nil {
		log.Printf("ListWidgetTypes error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list widget types"})
		return
	}

	c.JSON(http.StatusOK, transformWidgetTypes(widgets))
}

// GetWidgetSchema returns the config schema for a specific widget type
func (h *Handlers) GetWidgetSchema(c *gin.Context) {
	widgetType := c.Param("type")
	if widgetType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Widget type required"})
		return
	}

	queries := sqlc.New(h.pool)
	widget, err := queries.GetWidgetTypeByName(c.Request.Context(), widgetType)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Widget type not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"widget_type":    widget.WidgetType,
		"name":           widget.Name,
		"description":    widget.Description,
		"category":       widget.Category,
		"config_schema":  json.RawMessage(widget.ConfigSchema),
		"default_config": json.RawMessage(widget.DefaultConfig),
		"default_size":   json.RawMessage(widget.DefaultSize),
		"min_size":       json.RawMessage(widget.MinSize),
		"sse_events":     widget.SseEvents,
	})
}

// ============================================================================
// TEMPLATE HANDLERS
// ============================================================================

func (h *Handlers) ListDashboardTemplates(c *gin.Context) {
	queries := sqlc.New(h.pool)
	templates, err := queries.ListDashboardTemplates(c.Request.Context())
	if err != nil {
		log.Printf("ListDashboardTemplates error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list templates"})
		return
	}

	c.JSON(http.StatusOK, transformTemplates(templates))
}

func (h *Handlers) CreateDashboardFromTemplate(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	templateID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	var req struct {
		Name        string  `json:"name"`
		WorkspaceID *string `json:"workspace_id"`
	}
	_ = c.ShouldBindJSON(&req)

	var workspaceID pgtype.UUID
	if req.WorkspaceID != nil {
		if id, err := uuid.Parse(*req.WorkspaceID); err == nil {
			workspaceID = pgtype.UUID{Bytes: id, Valid: true}
		}
	}

	queries := sqlc.New(h.pool)
	dashboard, err := queries.CreateDashboardFromTemplate(c.Request.Context(), sqlc.CreateDashboardFromTemplateParams{
		ID:          pgtype.UUID{Bytes: templateID, Valid: true},
		UserID:      user.ID,
		WorkspaceID: workspaceID,
		Column4:     req.Name,
	})
	if err != nil {
		log.Printf("CreateDashboardFromTemplate error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create dashboard from template"})
		return
	}

	c.JSON(http.StatusCreated, transformDashboard(sqlc.UserDashboard(dashboard)))
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

func generateShareToken() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func transformDashboard(d sqlc.UserDashboard) gin.H {
	result := gin.H{
		"id":          dashboardUuidToString(d.ID),
		"user_id":     d.UserID,
		"name":        d.Name,
		"description": d.Description,
		"is_default":  d.IsDefault,
		"layout":      json.RawMessage(d.Layout),
		"visibility":  d.Visibility,
		"share_token": d.ShareToken,
		"created_via": d.CreatedVia,
		"created_at":  d.CreatedAt.Time,
		"updated_at":  d.UpdatedAt.Time,
	}
	if d.WorkspaceID.Valid {
		result["workspace_id"] = dashboardUuidToString(d.WorkspaceID)
	}
	return result
}

func transformDashboards(dashboards []sqlc.UserDashboard) []gin.H {
	result := make([]gin.H, len(dashboards))
	for i, d := range dashboards {
		result[i] = transformDashboard(d)
	}
	return result
}

func transformWidgetTypes(widgets []sqlc.DashboardWidget) []gin.H {
	result := make([]gin.H, len(widgets))
	for i, w := range widgets {
		result[i] = gin.H{
			"widget_type":    w.WidgetType,
			"name":           w.Name,
			"description":    w.Description,
			"category":       w.Category,
			"config_schema":  json.RawMessage(w.ConfigSchema),
			"default_config": json.RawMessage(w.DefaultConfig),
			"default_size":   json.RawMessage(w.DefaultSize),
			"min_size":       json.RawMessage(w.MinSize),
			"sse_events":     w.SseEvents,
			"is_enabled":     w.IsEnabled,
		}
	}
	return result
}

func transformTemplates(templates []sqlc.DashboardTemplate) []gin.H {
	result := make([]gin.H, len(templates))
	for i, t := range templates {
		result[i] = gin.H{
			"id":            dashboardUuidToString(t.ID),
			"name":          t.Name,
			"description":   t.Description,
			"category":      t.Category,
			"layout":        json.RawMessage(t.Layout),
			"thumbnail_url": t.ThumbnailUrl,
			"is_default":    t.IsDefault,
			"sort_order":    t.SortOrder,
		}
	}
	return result
}

func dashboardUuidToString(u pgtype.UUID) string {
	if !u.Valid {
		return ""
	}
	return uuid.UUID(u.Bytes).String()
}
