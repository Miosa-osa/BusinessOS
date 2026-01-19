package handlers

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/rhl/businessos-backend/internal/middleware"
	"github.com/rhl/businessos-backend/internal/windowcapture"
)

// WindowCaptureHandler handles window capture WebSocket connections
type WindowCaptureHandler struct {
	logger *slog.Logger
}

// NewWindowCaptureHandler creates a new window capture handler
func NewWindowCaptureHandler() *WindowCaptureHandler {
	return &WindowCaptureHandler{
		logger: slog.Default().With("component", "window_capture"),
	}
}

// WebSocket upgrader for window capture
var captureUpgrader = websocket.Upgrader{
	ReadBufferSize:  8192,
	WriteBufferSize: 65536, // Larger buffer for frame data
	CheckOrigin: func(r *http.Request) bool {
		// Use same origin check as terminal
		origin := r.Header.Get("Origin")
		if origin == "" {
			return false
		}
		// Allow localhost for development (5173, 5174, 5175, etc)
		allowedOrigins := []string{
			"http://localhost:5173",
			"http://localhost:5174",
			"http://localhost:5175",
			"http://localhost:3000",
			"https://localhost:5173",
			"https://localhost:5174",
			"https://app.businessos.com",
		}
		for _, allowed := range allowedOrigins {
			if origin == allowed {
				return true
			}
		}
		return false
	},
}

// HandleCapture handles the WebSocket connection for window capture streaming
// @Summary Stream window capture via WebSocket
// @Description Establishes a WebSocket connection for real-time window capture streaming
// @Tags window-capture
// @Produce json
// @Success 101 {string} string "Switching Protocols"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /api/window-capture/stream [get]
func (h *WindowCaptureHandler) HandleCapture(c *gin.Context) {
	// Get authenticated user
	user := middleware.GetCurrentUser(c)
	if user == nil {
		h.logger.Warn("window capture connection denied - no authenticated user")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	h.logger.Info("window capture connection request",
		"user_id", user.ID,
		"remote_addr", c.Request.RemoteAddr,
	)

	// Upgrade to WebSocket
	conn, err := captureUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("failed to upgrade WebSocket connection",
			"error", err,
			"user_id", user.ID,
		)
		return
	}

	// Create and run stream session
	session := windowcapture.NewStreamSession(conn, h.logger)
	session.Run()
}

// CheckPermission checks if screen capture permission is granted
// @Summary Check screen capture permission
// @Description Returns whether the app has screen capture permission on macOS
// @Tags window-capture
// @Produce json
// @Success 200 {object} map[string]bool "Permission status"
// @Router /api/window-capture/permission [get]
func (h *WindowCaptureHandler) CheckPermission(c *gin.Context) {
	granted := windowcapture.HasScreenCapturePermission()
	c.JSON(http.StatusOK, gin.H{
		"granted": granted,
	})
}

// RequestPermission triggers the screen capture permission dialog
// @Summary Request screen capture permission
// @Description Triggers the macOS screen capture permission dialog
// @Tags window-capture
// @Produce json
// @Success 200 {object} map[string]string "Request sent"
// @Router /api/window-capture/permission [post]
func (h *WindowCaptureHandler) RequestPermission(c *gin.Context) {
	windowcapture.RequestScreenCapturePermission()
	c.JSON(http.StatusOK, gin.H{
		"message": "Permission request triggered",
	})
}

// ListWindows lists windows for a specific bundle ID
// @Summary List windows for app
// @Description Returns all windows for a specific app bundle ID
// @Tags window-capture
// @Produce json
// @Param bundle_id query string true "App Bundle ID"
// @Success 200 {object} map[string]interface{} "Windows list"
// @Failure 400 {object} map[string]string "Bad request"
// @Router /api/window-capture/windows [get]
func (h *WindowCaptureHandler) ListWindows(c *gin.Context) {
	bundleID := c.Query("bundle_id")
	if bundleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bundle_id is required"})
		return
	}

	windows, err := windowcapture.GetWindowsForBundleID(bundleID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"windows": []interface{}{},
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"windows": windows,
		"count":   len(windows),
	})
}
