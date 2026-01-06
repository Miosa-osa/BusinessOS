package clickup

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Handler provides HTTP handlers for ClickUp integration routes.
type Handler struct {
	provider *Provider
}

// NewHandler creates a new ClickUp integration handler.
func NewHandler(provider *Provider) *Handler {
	return &Handler{
		provider: provider,
	}
}

// RegisterRoutes registers all ClickUp integration routes.
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	// OAuth routes
	r.GET("/auth", h.GetAuthURL)
	r.GET("/callback", h.HandleCallback)
	r.POST("/disconnect", h.Disconnect)
	r.GET("/status", h.GetStatus)

	// Workspace routes
	r.GET("/workspaces", h.GetWorkspaces)
	r.POST("/workspaces/sync", h.SyncWorkspaces)

	// Space routes
	r.GET("/workspaces/:workspace_id/spaces", h.GetSpaces)

	// List routes
	r.GET("/spaces/:space_id/lists", h.GetListsFromSpace)
	r.GET("/folders/:folder_id/lists", h.GetListsFromFolder)

	// Task routes
	r.GET("/lists/:list_id/tasks", h.GetTasks)
	r.GET("/tasks/:id", h.GetTask)
	r.POST("/lists/:list_id/tasks", h.CreateTask)
	r.PUT("/tasks/:id", h.UpdateTask)
	r.POST("/tasks/sync", h.SyncTasks)
}

// GetAuthURL returns the OAuth authorization URL.
func (h *Handler) GetAuthURL(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Generate state with user ID for callback
	state := generateState(userID)

	authURL := h.provider.GetAuthURL(state)
	c.JSON(http.StatusOK, gin.H{
		"auth_url": authURL,
	})
}

// HandleCallback handles the OAuth callback.
func (h *Handler) HandleCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing authorization code"})
		return
	}

	// Extract user ID from state
	userID := extractUserIDFromState(state)
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state"})
		return
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

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"account_name":  token.AccountName,
		"account_email": token.AccountEmail,
		"account_id":    token.AccountID,
	})
}

// Disconnect disconnects the ClickUp integration.
func (h *Handler) Disconnect(c *gin.Context) {
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

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// GetStatus returns the connection status.
func (h *Handler) GetStatus(c *gin.Context) {
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

	c.JSON(http.StatusOK, status)
}

// Workspace Handlers

// GetWorkspaces retrieves all workspaces for the user.
func (h *Handler) GetWorkspaces(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	workspaces, err := h.provider.GetWorkspaces(c.Request.Context(), userID)
	if err != nil {
		log.Printf("Failed to get workspaces: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get workspaces"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"workspaces": workspaces,
		"count":      len(workspaces),
	})
}

// SyncWorkspaces syncs workspaces from ClickUp.
func (h *Handler) SyncWorkspaces(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	workspaces, err := h.provider.GetWorkspaces(c.Request.Context(), userID)
	if err != nil {
		log.Printf("Failed to sync workspaces: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sync workspaces"})
		return
	}

	// TODO: Save workspaces to database if needed

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"synced":     len(workspaces),
		"workspaces": workspaces,
	})
}

// Space Handlers

// GetSpaces retrieves all spaces for a workspace.
func (h *Handler) GetSpaces(c *gin.Context) {
	userID := c.GetString("user_id")
	workspaceID := c.Param("workspace_id")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing workspace_id"})
		return
	}

	spaces, err := h.provider.GetSpaces(c.Request.Context(), userID, workspaceID)
	if err != nil {
		log.Printf("Failed to get spaces: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get spaces"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"spaces": spaces,
		"count":  len(spaces),
	})
}

// List Handlers

// GetListsFromSpace retrieves all folderless lists from a space.
func (h *Handler) GetListsFromSpace(c *gin.Context) {
	userID := c.GetString("user_id")
	spaceID := c.Param("space_id")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if spaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing space_id"})
		return
	}

	lists, err := h.provider.GetListsFromSpace(c.Request.Context(), userID, spaceID)
	if err != nil {
		log.Printf("Failed to get lists from space: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get lists"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"lists": lists,
		"count": len(lists),
	})
}

// GetListsFromFolder retrieves all lists from a folder.
func (h *Handler) GetListsFromFolder(c *gin.Context) {
	userID := c.GetString("user_id")
	folderID := c.Param("folder_id")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if folderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing folder_id"})
		return
	}

	lists, err := h.provider.GetListsFromFolder(c.Request.Context(), userID, folderID)
	if err != nil {
		log.Printf("Failed to get lists from folder: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get lists"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"lists": lists,
		"count": len(lists),
	})
}

// Task Handlers

// GetTasks retrieves all tasks from a list.
func (h *Handler) GetTasks(c *gin.Context) {
	userID := c.GetString("user_id")
	listID := c.Param("list_id")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if listID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing list_id"})
		return
	}

	tasks, err := h.provider.GetTasks(c.Request.Context(), userID, listID)
	if err != nil {
		log.Printf("Failed to get tasks: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tasks"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks": tasks,
		"count": len(tasks),
	})
}

// GetTask retrieves a specific task by ID.
func (h *Handler) GetTask(c *gin.Context) {
	userID := c.GetString("user_id")
	taskID := c.Param("id")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing task id"})
		return
	}

	task, err := h.provider.GetTask(c.Request.Context(), userID, taskID)
	if err != nil {
		log.Printf("Failed to get task: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get task"})
		return
	}

	c.JSON(http.StatusOK, task)
}

// CreateTask creates a new task in a list.
func (h *Handler) CreateTask(c *gin.Context) {
	userID := c.GetString("user_id")
	listID := c.Param("list_id")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if listID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing list_id"})
		return
	}

	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	task, err := h.provider.CreateTask(c.Request.Context(), userID, listID, req)
	if err != nil {
		log.Printf("Failed to create task: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
		return
	}

	c.JSON(http.StatusCreated, task)
}

// UpdateTask updates an existing task.
func (h *Handler) UpdateTask(c *gin.Context) {
	userID := c.GetString("user_id")
	taskID := c.Param("id")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing task id"})
		return
	}

	var req UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	task, err := h.provider.UpdateTask(c.Request.Context(), userID, taskID, req)
	if err != nil {
		log.Printf("Failed to update task: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task"})
		return
	}

	c.JSON(http.StatusOK, task)
}

// SyncTasks syncs tasks from ClickUp.
func (h *Handler) SyncTasks(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Get list_id from query or body
	listID := c.Query("list_id")
	if listID == "" {
		var body map[string]string
		if err := c.ShouldBindJSON(&body); err == nil {
			listID = body["list_id"]
		}
	}

	if listID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing list_id parameter"})
		return
	}

	tasks, err := h.provider.GetTasks(c.Request.Context(), userID, listID)
	if err != nil {
		log.Printf("Failed to sync tasks: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sync tasks"})
		return
	}

	// TODO: Save tasks to database if needed

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"synced":  len(tasks),
		"tasks":   tasks,
	})
}

// Helper functions

func generateState(userID string) string {
	// Simple state with user ID - in production, use JWT or encrypted token
	data := map[string]string{
		"user_id":   userID,
		"timestamp": time.Now().Format(time.RFC3339),
	}
	b, _ := json.Marshal(data)
	return string(b)
}

func extractUserIDFromState(state string) string {
	var data map[string]string
	if err := json.Unmarshal([]byte(state), &data); err != nil {
		return ""
	}
	return data["user_id"]
}
