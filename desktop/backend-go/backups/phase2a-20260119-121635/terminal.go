package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/container"
	"github.com/rhl/businessos-backend/internal/middleware"
	"github.com/rhl/businessos-backend/internal/terminal"
)

// TerminalHandler handles terminal-related HTTP requests
type TerminalHandler struct {
	wsHandler    *terminal.WebSocketHandler
	manager      *terminal.Manager
	containerMgr *container.ContainerManager
	pubsub       *terminal.TerminalPubSub
}

// NewTerminalHandler creates a new terminal handler
// pubsub is optional - if nil, terminal works without horizontal scaling
func NewTerminalHandler(containerMgr *container.ContainerManager, pubsub *terminal.TerminalPubSub) *TerminalHandler {
	manager := terminal.NewManager(containerMgr)

	// Wire up pub/sub if available
	if pubsub != nil {
		manager.SetPubSub(pubsub)
		log.Printf("[Terminal] Pub/sub enabled for horizontal scaling (instance=%s)", pubsub.InstanceID())
	} else {
		log.Printf("[Terminal] Pub/sub disabled - single instance mode")
	}

	wsHandler := terminal.NewWebSocketHandler(manager)
	return &TerminalHandler{
		wsHandler:    wsHandler,
		manager:      manager,
		containerMgr: containerMgr,
		pubsub:       pubsub,
	}
}

// HandleWebSocket handles WebSocket terminal connections
// @Summary Connect to terminal via WebSocket
// @Description Establishes a WebSocket connection for real-time terminal I/O
// @Tags terminal
// @Produce json
// @Param cols query int false "Terminal columns" default(80)
// @Param rows query int false "Terminal rows" default(24)
// @Param shell query string false "Shell to use" default(zsh)
// @Param cwd query string false "Working directory"
// @Success 101 {string} string "Switching Protocols"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /api/terminal/ws [get]
func (h *TerminalHandler) HandleWebSocket(c *gin.Context) {
	log.Printf("[Terminal] HandleWebSocket called from %s", c.Request.RemoteAddr)
	log.Printf("[Terminal] Request headers: %v", c.Request.Header)

	// Get authenticated user from context (set by AuthMiddleware as "user")
	user := middleware.GetCurrentUser(c)
	if user == nil {
		log.Printf("[Terminal] No authenticated user found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	log.Printf("[Terminal] User authenticated: %s (%s)", user.Name, user.ID)

	// Upgrade to WebSocket and handle connection
	h.wsHandler.HandleConnection(c.Writer, c.Request, user.ID)
}

// ListSessions lists all active terminal sessions for the user
// @Summary List terminal sessions
// @Description Returns all active terminal sessions for the authenticated user
// @Tags terminal
// @Produce json
// @Success 200 {object} map[string]interface{} "Sessions list"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /api/terminal/sessions [get]
func (h *TerminalHandler) ListSessions(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	sessions := h.manager.GetUserSessions(user.ID)

	c.JSON(http.StatusOK, gin.H{
		"sessions": sessions,
		"count":    len(sessions),
	})
}

// CloseSession closes a specific terminal session
// @Summary Close terminal session
// @Description Closes and cleans up a terminal session
// @Tags terminal
// @Produce json
// @Param id path string true "Session ID"
// @Success 200 {object} map[string]string "Session closed"
// @Failure 404 {object} map[string]string "Session not found"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /api/terminal/sessions/{id} [delete]
func (h *TerminalHandler) CloseSession(c *gin.Context) {
	sessionID := c.Param("id")

	if err := h.manager.CloseSession(sessionID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Session closed"})
}

// GetManager returns the terminal manager (for cleanup)
func (h *TerminalHandler) GetManager() *terminal.Manager {
	return h.manager
}

// Shutdown gracefully shuts down the terminal handler
func (h *TerminalHandler) Shutdown() {
	if h.pubsub != nil {
		log.Printf("[Terminal] Closing pub/sub connections...")
		h.pubsub.Close()
	}
	h.manager.Shutdown()
}
