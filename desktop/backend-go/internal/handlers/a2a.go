package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/idempotency"
	"github.com/rhl/businessos-backend/internal/services"
	"github.com/rhl/businessos-backend/internal/utils"
)

// ---------------------------------------------------------------------------
// Request / Response Types
// ---------------------------------------------------------------------------

// discoverAgentRequest is the body for POST /api/integrations/a2a/agents/discover
type discoverAgentRequest struct {
	AgentURL string `json:"agent_url" binding:"required"`
}

// callAgentRequest is the body for POST /api/integrations/a2a/agents/call
type callAgentRequest struct {
	AgentURL string `json:"agent_url" binding:"required"`
	Message  string `json:"message" binding:"required"`
}

// executeAgentToolRequest is the body for POST /api/integrations/a2a/agents/tools/:name
type executeAgentToolRequest struct {
	AgentURL string         `json:"agent_url" binding:"required"`
	Args     map[string]any `json:"args"`
}

// getAgentToolsRequest is the query for GET /api/integrations/a2a/agents/tools
type getAgentToolsRequest struct {
	AgentURL string `form:"agent_url" binding:"required"`
}

// disconnectAgentRequest is the query/body for DELETE /api/integrations/a2a/agents
type disconnectAgentRequest struct {
	AgentURL string `json:"agent_url" binding:"required"`
}

// ---------------------------------------------------------------------------
// Handler
// ---------------------------------------------------------------------------

// A2AHandler handles A2A agent communication endpoints.
type A2AHandler struct {
	a2aClient      *services.A2AClient
	idempotencyKey *idempotency.Store
}

// NewA2AHandler creates a new A2A handler.
// Starts a background goroutine to clean up expired idempotency entries hourly.
func NewA2AHandler(a2aClient *services.A2AClient) *A2AHandler {
	store := idempotency.New()
	h := &A2AHandler{
		a2aClient:      a2aClient,
		idempotencyKey: store,
	}
	// WvdA: bounded cleanup — stops when process exits; ticker prevents unbounded memory growth
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			if n := store.Cleanup(); n > 0 {
				slog.Info("idempotency cache cleanup", "deleted_keys", n)
			}
		}
	}()
	return h
}

// DiscoverAgent handles POST /api/integrations/a2a/agents/discover
// Fetches the agent card from a remote A2A agent and caches the connection.
func (h *A2AHandler) DiscoverAgent(c *gin.Context) {
	var req discoverAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, slog.Default(), err)
		return
	}

	if err := validateAgentURL(req.AgentURL); err != nil {
		utils.RespondBadRequest(c, slog.Default(), err.Error())
		return
	}

	card, err := h.a2aClient.DiscoverAgent(c.Request.Context(), req.AgentURL)
	if err != nil {
		slog.Error("A2A agent discovery failed", "url", req.AgentURL, "error", err)
		utils.RespondBadRequest(c, slog.Default(), "Agent discovery failed: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"agent_card": card,
	})
}

// CallAgent handles POST /api/integrations/a2a/agents/call
// Sends a message to a remote A2A agent and returns the task result.
// Supports Idempotency-Key header for idempotent replays.
func (h *A2AHandler) CallAgent(c *gin.Context) {
	// Check idempotency cache first
	if found, status, body := h.checkIdempotency(c); found {
		var result interface{}
		json.Unmarshal([]byte(body), &result)
		c.JSON(status, result)
		return
	}

	var req callAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, slog.Default(), err)
		return
	}

	if err := validateAgentURL(req.AgentURL); err != nil {
		utils.RespondBadRequest(c, slog.Default(), err.Error())
		return
	}

	task, err := h.a2aClient.CallAgent(c.Request.Context(), req.AgentURL, req.Message)
	if err != nil {
		slog.Error("A2A agent call failed", "url", req.AgentURL, "error", err)
		utils.RespondBadRequest(c, slog.Default(), "Agent call failed: "+err.Error())
		return
	}

	response := gin.H{
		"task": task,
	}

	// Cache idempotent response
	h.cacheIdempotencyResponse(c, http.StatusOK, response)

	c.JSON(http.StatusOK, response)
}

// GetAgentTools handles GET /api/integrations/a2a/agents/tools
// Retrieves the list of tools exposed by a remote A2A agent.
func (h *A2AHandler) GetAgentTools(c *gin.Context) {
	var req getAgentToolsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.RespondInvalidRequest(c, slog.Default(), err)
		return
	}

	if err := validateAgentURL(req.AgentURL); err != nil {
		utils.RespondBadRequest(c, slog.Default(), err.Error())
		return
	}

	tools, err := h.a2aClient.GetAgentTools(c.Request.Context(), req.AgentURL)
	if err != nil {
		slog.Error("A2A agent tools discovery failed", "url", req.AgentURL, "error", err)
		utils.RespondBadRequest(c, slog.Default(), "Failed to discover agent tools: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tools": tools,
	})
}

// ExecuteAgentTool handles POST /api/integrations/a2a/agents/tools/:name
// Invokes a specific tool on a remote A2A agent.
// Supports Idempotency-Key header for idempotent replays.
func (h *A2AHandler) ExecuteAgentTool(c *gin.Context) {
	// Check idempotency cache first
	if found, status, body := h.checkIdempotency(c); found {
		var result interface{}
		json.Unmarshal([]byte(body), &result)
		c.JSON(status, result)
		return
	}

	toolName := c.Param("name")
	if toolName == "" {
		utils.RespondBadRequest(c, slog.Default(), "Tool name is required")
		return
	}

	var req executeAgentToolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, slog.Default(), err)
		return
	}

	if err := validateAgentURL(req.AgentURL); err != nil {
		utils.RespondBadRequest(c, slog.Default(), err.Error())
		return
	}

	args := req.Args
	if args == nil {
		args = make(map[string]any)
	}

	result, err := h.a2aClient.ExecuteAgentTool(c.Request.Context(), req.AgentURL, toolName, args)
	if err != nil {
		slog.Error("A2A tool execution failed", "tool", toolName, "url", req.AgentURL, "error", err)
		utils.RespondBadRequest(c, slog.Default(), "Tool execution failed: "+err.Error())
		return
	}

	response := gin.H{
		"result": result,
	}

	// Cache idempotent response
	h.cacheIdempotencyResponse(c, http.StatusOK, response)

	c.JSON(http.StatusOK, response)
}

// ListConnectedAgents handles GET /api/integrations/a2a/agents
// Returns all cached A2A agent connections.
func (h *A2AHandler) ListConnectedAgents(c *gin.Context) {
	agents := h.a2aClient.ListConnectedAgents()

	type agentSummary struct {
		URL      string                `json:"url"`
		Name     string                `json:"name"`
		Version  string                `json:"version"`
		Skills   []services.AgentSkill `json:"skills,omitempty"`
		LastSeen string                `json:"last_seen"`
	}

	summaries := make([]agentSummary, 0, len(agents))
	for _, conn := range agents {
		summary := agentSummary{
			URL:      conn.URL,
			LastSeen: conn.LastSeen.Format("2006-01-02T15:04:05Z"),
		}
		if conn.Card != nil {
			summary.Name = conn.Card.Name
			summary.Version = conn.Card.Version
			summary.Skills = conn.Card.Skills
		}
		summaries = append(summaries, summary)
	}

	c.JSON(http.StatusOK, gin.H{
		"agents": summaries,
	})
}

// DisconnectAgent handles DELETE /api/integrations/a2a/agents
// Removes a cached A2A agent connection.
func (h *A2AHandler) DisconnectAgent(c *gin.Context) {
	var req disconnectAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, slog.Default(), err)
		return
	}

	if err := h.a2aClient.DisconnectAgent(req.AgentURL); err != nil {
		utils.RespondNotFound(c, slog.Default(), "A2A agent connection")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// validateAgentURL performs basic URL format validation at the handler level.
// The service layer performs full SSRF protection via ValidateA2AAgentURL.
func validateAgentURL(rawURL string) error {
	if rawURL == "" {
		return fmt.Errorf("agent_url is required")
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid agent_url format")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("agent_url must use http or https")
	}
	return nil
}

// checkIdempotency checks if a request has been processed before using the Idempotency-Key header.
// If found in cache, returns (true, status, body). Otherwise, returns (false, 0, nil).
func (h *A2AHandler) checkIdempotency(c *gin.Context) (bool, int, string) {
	idempKey := c.GetHeader("Idempotency-Key")
	if idempKey == "" {
		return false, 0, ""
	}

	entry := h.idempotencyKey.Get(idempKey)
	if entry == nil {
		return false, 0, ""
	}

	slog.Debug("idempotency cache hit", "key", idempKey, "status", entry.Status)
	return true, entry.Status, entry.Body
}

// cacheIdempotencyResponse stores a response for idempotent replayability.
func (h *A2AHandler) cacheIdempotencyResponse(c *gin.Context, status int, body interface{}) {
	idempKey := c.GetHeader("Idempotency-Key")
	if idempKey == "" {
		return
	}

	bodyBytes, _ := json.Marshal(body)
	_ = h.idempotencyKey.Store(idempKey, status, bodyBytes)
}
