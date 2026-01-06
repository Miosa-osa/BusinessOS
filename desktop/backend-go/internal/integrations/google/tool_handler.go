// Package google provides HTTP handlers for individual Google tool integrations.
package google

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ToolHandler provides HTTP handlers for a specific Google tool.
type ToolHandler struct {
	pool     *pgxpool.Pool
	provider *ToolProvider
	calendar *CalendarService
	gmail    *GmailService
}

// NewToolHandler creates a new handler for a specific Google tool.
func NewToolHandler(pool *pgxpool.Pool, toolID string) (*ToolHandler, error) {
	provider, err := NewToolProvider(pool, toolID)
	if err != nil {
		return nil, err
	}

	h := &ToolHandler{
		pool:     pool,
		provider: provider,
	}

	// Initialize service-specific handlers based on tool type
	switch toolID {
	case "google_calendar":
		// Create a legacy provider wrapper for CalendarService
		legacyProvider := &Provider{pool: pool}
		legacyProvider.oauthConfig = provider.oauthConfig
		h.calendar = NewCalendarService(legacyProvider)
	case "google_gmail":
		legacyProvider := &Provider{pool: pool}
		legacyProvider.oauthConfig = provider.oauthConfig
		h.gmail = NewGmailService(legacyProvider)
	}

	return h, nil
}

// RegisterRoutes registers routes for this tool.
func (h *ToolHandler) RegisterRoutes(r *gin.RouterGroup) {
	// Common OAuth routes for all tools
	r.GET("/auth", h.GetAuthURL)
	r.GET("/callback", h.HandleCallback)
	r.POST("/disconnect", h.Disconnect)
	r.GET("/status", h.GetStatus)

	// Tool-specific routes
	switch h.provider.ID() {
	case "google_calendar":
		h.registerCalendarRoutes(r)
	case "google_gmail":
		h.registerGmailRoutes(r)
	}
}

func (h *ToolHandler) registerCalendarRoutes(r *gin.RouterGroup) {
	calendar := r.Group("/calendar")
	{
		calendar.GET("/events", h.GetCalendarEvents)
		calendar.POST("/events", h.CreateCalendarEvent)
		calendar.DELETE("/events/:id", h.DeleteCalendarEvent)
		calendar.POST("/sync", h.SyncCalendar)
	}
}

func (h *ToolHandler) registerGmailRoutes(r *gin.RouterGroup) {
	gmail := r.Group("/gmail")
	{
		gmail.GET("/emails", h.GetEmails)
		gmail.GET("/emails/:id", h.GetEmail)
		gmail.POST("/emails/send", h.SendEmail)
		gmail.POST("/sync", h.SyncGmail)
	}
}

// ============================================================================
// OAuth Handlers
// ============================================================================

// GetAuthURL returns the OAuth authorization URL for this specific tool.
func (h *ToolHandler) GetAuthURL(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Generate state with user ID and tool ID
	state := generateToolState(userID, h.provider.ID())
	authURL := h.provider.GetAuthURL(state)

	c.JSON(http.StatusOK, gin.H{
		"auth_url": authURL,
		"tool_id":  h.provider.ID(),
		"scopes":   h.provider.tool.Scopes,
	})
}

// HandleCallback handles the OAuth callback for this tool.
func (h *ToolHandler) HandleCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing authorization code"})
		return
	}

	// Extract user ID from state
	userID, toolID := extractToolState(state)
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state"})
		return
	}

	// Verify tool ID matches
	if toolID != "" && toolID != h.provider.ID() {
		log.Printf("Tool ID mismatch: expected %s, got %s", h.provider.ID(), toolID)
	}

	// Exchange code for tokens
	token, err := h.provider.ExchangeCode(c.Request.Context(), code)
	if err != nil {
		log.Printf("Failed to exchange code: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange code"})
		return
	}

	// Save tokens
	if err := h.provider.SaveToken(c.Request.Context(), userID, token); err != nil {
		log.Printf("Failed to save token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save token"})
		return
	}

	// Redirect to frontend success page
	// SECURITY: Validate redirect_uri to prevent open redirect attacks
	frontendURL := getSafeRedirectURL(c.Query("redirect_uri"), h.provider.ID())
	c.Redirect(http.StatusTemporaryRedirect, frontendURL)
}

// Disconnect removes the user's connection to this tool.
func (h *ToolHandler) Disconnect(c *gin.Context) {
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

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"tool_id": h.provider.ID(),
	})
}

// GetStatus returns the connection status for this tool.
func (h *ToolHandler) GetStatus(c *gin.Context) {
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

	c.JSON(http.StatusOK, gin.H{
		"tool_id":   h.provider.ID(),
		"tool_name": h.provider.Name(),
		"connected": status.Connected,
		"account":   status.AccountName,
		"scopes":    status.Scopes,
	})
}

// ============================================================================
// Calendar Handlers
// ============================================================================

// GetCalendarEvents returns calendar events.
func (h *ToolHandler) GetCalendarEvents(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Check if connected
	if !h.provider.IsConnected(c.Request.Context(), userID) {
		c.JSON(http.StatusOK, gin.H{
			"events":    []interface{}{},
			"count":     0,
			"connected": false,
			"message":   "Google Calendar not connected",
		})
		return
	}

	// Parse date range
	start := time.Now().AddDate(0, 0, -7)
	end := time.Now().AddDate(0, 1, 0)

	if startStr := c.Query("start"); startStr != "" {
		if t, err := time.Parse(time.RFC3339, startStr); err == nil {
			start = t
		}
	}
	if endStr := c.Query("end"); endStr != "" {
		if t, err := time.Parse(time.RFC3339, endStr); err == nil {
			end = t
		}
	}

	events, err := h.calendar.GetEvents(c.Request.Context(), userID, start, end)
	if err != nil {
		log.Printf("Failed to get events: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get events"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"events":    events,
		"count":     len(events),
		"connected": true,
	})
}

// CreateCalendarEvent creates a new calendar event.
func (h *ToolHandler) CreateCalendarEvent(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if !h.provider.IsConnected(c.Request.Context(), userID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Google Calendar not connected"})
		return
	}

	var event CalendarEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	created, err := h.calendar.CreateEvent(c.Request.Context(), userID, &event)
	if err != nil {
		log.Printf("Failed to create event: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create event"})
		return
	}

	c.JSON(http.StatusCreated, created)
}

// DeleteCalendarEvent deletes a calendar event.
func (h *ToolHandler) DeleteCalendarEvent(c *gin.Context) {
	userID := c.GetString("user_id")
	eventID := c.Param("id")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if err := h.calendar.DeleteEvent(c.Request.Context(), userID, eventID); err != nil {
		log.Printf("Failed to delete event: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// SyncCalendar syncs calendar events from Google.
func (h *ToolHandler) SyncCalendar(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if !h.provider.IsConnected(c.Request.Context(), userID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Google Calendar not connected"})
		return
	}

	timeMin := time.Now().AddDate(0, -1, 0)
	timeMax := time.Now().AddDate(0, 3, 0)

	result, err := h.calendar.SyncEvents(c.Request.Context(), userID, timeMin, timeMax)
	if err != nil {
		log.Printf("Failed to sync calendar: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sync calendar"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ============================================================================
// Gmail Handlers
// ============================================================================

// GetEmails returns emails from a folder.
func (h *ToolHandler) GetEmails(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if !h.provider.IsConnected(c.Request.Context(), userID) {
		c.JSON(http.StatusOK, gin.H{
			"emails":    []interface{}{},
			"count":     0,
			"connected": false,
			"message":   "Gmail not connected",
		})
		return
	}

	// Would call h.gmail.GetEmails(...)
	c.JSON(http.StatusOK, gin.H{
		"emails":    []interface{}{},
		"count":     0,
		"connected": true,
	})
}

// GetEmail returns a single email.
func (h *ToolHandler) GetEmail(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

// SendEmail sends an email.
func (h *ToolHandler) SendEmail(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

// SyncGmail syncs emails from Gmail.
func (h *ToolHandler) SyncGmail(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

// ============================================================================
// State Helpers
// ============================================================================

func generateToolState(userID, toolID string) string {
	data := map[string]string{
		"user_id":   userID,
		"tool_id":   toolID,
		"timestamp": time.Now().Format(time.RFC3339),
	}
	b, _ := json.Marshal(data)
	return string(b)
}

func extractToolState(state string) (userID, toolID string) {
	var data map[string]string
	if err := json.Unmarshal([]byte(state), &data); err != nil {
		return "", ""
	}
	return data["user_id"], data["tool_id"]
}

// getSafeRedirectURL validates and returns a safe redirect URL.
// SECURITY: Prevents open redirect attacks by only allowing known-safe origins.
func getSafeRedirectURL(requestedURL string, toolID string) string {
	// Default safe URL
	defaultURL := "http://localhost:5173/integrations?connected=" + toolID

	// Get allowed origins from environment
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}

	// If no redirect requested, use default
	if requestedURL == "" {
		return defaultURL
	}

	// Parse the requested URL
	parsed, err := url.Parse(requestedURL)
	if err != nil {
		log.Printf("Invalid redirect URL: %v", err)
		return defaultURL
	}

	// Allowed origins whitelist
	allowedOrigins := []string{
		"http://localhost:5173",
		"http://localhost:3000",
		"https://localhost:5173",
		"https://localhost:3000",
	}

	// Add configured frontend URL if set
	if frontendURL != "" {
		allowedOrigins = append(allowedOrigins, frontendURL)
	}

	// Add production URLs if configured
	if prodURL := os.Getenv("PRODUCTION_FRONTEND_URL"); prodURL != "" {
		allowedOrigins = append(allowedOrigins, prodURL)
	}

	// Check if the origin is allowed
	requestedOrigin := parsed.Scheme + "://" + parsed.Host
	for _, allowed := range allowedOrigins {
		if strings.HasPrefix(allowed, requestedOrigin) || requestedOrigin == allowed {
			// Origin is allowed, return the full URL
			return requestedURL
		}
	}

	// Origin not allowed - log and return default
	log.Printf("Blocked redirect to untrusted origin: %s (allowed: %v)", requestedOrigin, allowedOrigins)
	return defaultURL
}

// Note: decodeJSON is defined in helpers.go
