package microsoft

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// Handler provides HTTP handlers for Microsoft integration routes.
type Handler struct {
	provider *Provider
	outlook  *OutlookService
	onedrive *OneDriveService
	todo     *ToDoService
}

// NewHandler creates a new Microsoft integration handler.
func NewHandler(provider *Provider) *Handler {
	return &Handler{
		provider: provider,
		outlook:  NewOutlookService(provider),
		onedrive: NewOneDriveService(provider),
		todo:     NewToDoService(provider),
	}
}

// RegisterRoutes registers all Microsoft integration routes.
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	// OAuth routes
	r.GET("/auth", h.GetAuthURL)
	r.GET("/callback", h.HandleCallback)
	r.POST("/disconnect", h.Disconnect)
	r.GET("/status", h.GetStatus)

	// Calendar routes
	if h.provider.HasFeature("calendar") {
		calendar := r.Group("/calendar")
		{
			calendar.GET("/events", h.GetCalendarEvents)
			calendar.POST("/events", h.CreateCalendarEvent)
			calendar.POST("/sync", h.SyncCalendar)
		}
	}

	// Mail routes
	if h.provider.HasFeature("mail") {
		mail := r.Group("/mail")
		{
			mail.GET("/emails", h.GetEmails)
			mail.GET("/emails/:id", h.GetEmail)
			mail.POST("/send", h.SendEmail)
			mail.POST("/sync", h.SyncMail)
		}
	}

	// Files routes (OneDrive)
	if h.provider.HasFeature("files") {
		files := r.Group("/files")
		{
			files.GET("", h.GetFiles)
			files.GET("/:id", h.GetFile)
			files.POST("/sync", h.SyncFiles)
		}
	}

	// Tasks routes (Microsoft To Do)
	if h.provider.HasFeature("tasks") {
		tasks := r.Group("/tasks")
		{
			tasks.GET("/lists", h.GetTaskLists)
			tasks.GET("/:list_id/tasks", h.GetTasks)
			tasks.POST("/:list_id/tasks", h.CreateTask)
			tasks.POST("/:list_id/tasks/:task_id/complete", h.CompleteTask)
			tasks.POST("/sync", h.SyncTasks)
		}
	}
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

	// Get features from query params (optional)
	features := c.QueryArray("features")
	if len(features) == 0 {
		features = h.provider.Features()
	}

	authURL := h.provider.GetAuthURLWithFeatures(state, features)
	c.JSON(http.StatusOK, gin.H{
		"auth_url": authURL,
		"features": features,
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
		"account_email": token.AccountEmail,
		"scopes":        token.Scopes,
	})
}

// Disconnect disconnects the Microsoft integration.
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

// ============================================================================
// CALENDAR HANDLERS
// ============================================================================

// GetCalendarEvents returns calendar events.
func (h *Handler) GetCalendarEvents(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Parse date range
	startStr := c.Query("start")
	endStr := c.Query("end")

	start := time.Now().AddDate(0, 0, -7) // Default: 7 days ago
	end := time.Now().AddDate(0, 1, 0)    // Default: 1 month from now

	if startStr != "" {
		if t, err := time.Parse(time.RFC3339, startStr); err == nil {
			start = t
		}
	}
	if endStr != "" {
		if t, err := time.Parse(time.RFC3339, endStr); err == nil {
			end = t
		}
	}

	events, err := h.outlook.GetEvents(c.Request.Context(), userID, start, end)
	if err != nil {
		log.Printf("Failed to get events: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get events"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"events": events,
		"count":  len(events),
	})
}

// CreateCalendarEvent creates a new calendar event.
func (h *Handler) CreateCalendarEvent(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var event OutlookEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	created, err := h.outlook.CreateEvent(c.Request.Context(), userID, &event)
	if err != nil {
		log.Printf("Failed to create event: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create event"})
		return
	}

	c.JSON(http.StatusCreated, created)
}

// SyncCalendar syncs calendar events from Outlook.
func (h *Handler) SyncCalendar(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Parse sync range
	timeMin := time.Now().AddDate(0, -1, 0) // 1 month ago
	timeMax := time.Now().AddDate(0, 3, 0)  // 3 months from now

	result, err := h.outlook.SyncEvents(c.Request.Context(), userID, timeMin, timeMax)
	if err != nil {
		log.Printf("Failed to sync calendar: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sync calendar"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ============================================================================
// MAIL HANDLERS
// ============================================================================

// GetEmails returns emails from Outlook.
func (h *Handler) GetEmails(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	folderID := c.Query("folder_id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	messages, err := h.outlook.GetMessages(c.Request.Context(), userID, folderID, limit, offset)
	if err != nil {
		log.Printf("Failed to get emails: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get emails"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"emails": messages,
		"count":  len(messages),
	})
}

// GetEmail returns a single email.
func (h *Handler) GetEmail(c *gin.Context) {
	userID := c.GetString("user_id")
	emailID := c.Param("id")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Get from database by ID
	var message OutlookMessage
	err := h.provider.Pool().QueryRow(c.Request.Context(), `
		SELECT id, user_id, message_id, conversation_id, subject, body_preview, importance,
			from_email, from_name, is_read, is_draft, has_attachments, folder_id,
			received_datetime, sent_datetime, synced_at
		FROM microsoft_mail_messages
		WHERE user_id = $1 AND message_id = $2
	`, userID, emailID).Scan(
		&message.ID, &message.UserID, &message.MessageID, &message.ConversationID,
		&message.Subject, &message.BodyPreview, &message.Importance,
		&message.FromEmail, &message.FromName, &message.IsRead, &message.IsDraft,
		&message.HasAttachments, &message.FolderID,
		&message.ReceivedDateTime, &message.SentDateTime, &message.SyncedAt,
	)

	if err != nil {
		log.Printf("Failed to get email: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get email"})
		return
	}

	c.JSON(http.StatusOK, message)
}

// SendEmail sends a new email via Outlook.
func (h *Handler) SendEmail(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req struct {
		To      []string `json:"to" binding:"required"`
		Subject string   `json:"subject" binding:"required"`
		Body    string   `json:"body" binding:"required"`
		IsHTML  bool     `json:"is_html"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.outlook.SendMessage(c.Request.Context(), userID, req.To, req.Subject, req.Body, req.IsHTML); err != nil {
		log.Printf("Failed to send email: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send email"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// SyncMail syncs emails from Outlook.
func (h *Handler) SyncMail(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	maxResults := 100
	if mr := c.Query("max_results"); mr != "" {
		if n, err := strconv.Atoi(mr); err == nil {
			maxResults = n
		}
	}

	result, err := h.outlook.SyncMessages(c.Request.Context(), userID, maxResults)
	if err != nil {
		log.Printf("Failed to sync mail: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sync mail"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ============================================================================
// FILES HANDLERS (OneDrive)
// ============================================================================

// GetFiles returns OneDrive files.
func (h *Handler) GetFiles(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	parentID := c.Query("parent_id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	files, err := h.onedrive.GetFiles(c.Request.Context(), userID, parentID, limit, offset)
	if err != nil {
		log.Printf("Failed to get files: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get files"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"files": files,
		"count": len(files),
	})
}

// GetFile returns a single OneDrive file.
func (h *Handler) GetFile(c *gin.Context) {
	userID := c.GetString("user_id")
	fileID := c.Param("id")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var file OneDriveFile
	err := h.provider.Pool().QueryRow(c.Request.Context(), `
		SELECT id, user_id, item_id, name, description, mime_type, size_bytes,
			web_url, is_folder, folder_child_count, synced_at
		FROM microsoft_onedrive_files
		WHERE user_id = $1 AND item_id = $2
	`, userID, fileID).Scan(
		&file.ID, &file.UserID, &file.ItemID, &file.Name, &file.Description,
		&file.MimeType, &file.SizeBytes, &file.WebURL, &file.IsFolder,
		&file.FolderChildCount, &file.SyncedAt,
	)

	if err != nil {
		log.Printf("Failed to get file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get file"})
		return
	}

	c.JSON(http.StatusOK, file)
}

// SyncFiles syncs files from OneDrive.
func (h *Handler) SyncFiles(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	maxResults := 100
	if mr := c.Query("max_results"); mr != "" {
		if n, err := strconv.Atoi(mr); err == nil {
			maxResults = n
		}
	}

	result, err := h.onedrive.SyncFiles(c.Request.Context(), userID, maxResults)
	if err != nil {
		log.Printf("Failed to sync files: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sync files"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ============================================================================
// TASKS HANDLERS (Microsoft To Do)
// ============================================================================

// GetTaskLists returns Microsoft To Do task lists.
func (h *Handler) GetTaskLists(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	lists, err := h.todo.GetLists(c.Request.Context(), userID)
	if err != nil {
		log.Printf("Failed to get task lists: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get task lists"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"lists": lists,
		"count": len(lists),
	})
}

// GetTasks returns tasks from a specific list.
func (h *Handler) GetTasks(c *gin.Context) {
	userID := c.GetString("user_id")
	listID := c.Param("list_id")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	includeCompleted := c.Query("include_completed") == "true"
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	tasks, err := h.todo.GetTasks(c.Request.Context(), userID, listID, includeCompleted, limit, offset)
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

// CreateTask creates a new task in a Microsoft To Do list.
func (h *Handler) CreateTask(c *gin.Context) {
	userID := c.GetString("user_id")
	listID := c.Param("list_id")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var task ToDoTask
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	created, err := h.todo.CreateTask(c.Request.Context(), userID, listID, &task)
	if err != nil {
		log.Printf("Failed to create task: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
		return
	}

	c.JSON(http.StatusCreated, created)
}

// CompleteTask marks a task as completed.
func (h *Handler) CompleteTask(c *gin.Context) {
	userID := c.GetString("user_id")
	listID := c.Param("list_id")
	taskID := c.Param("task_id")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if err := h.todo.CompleteTask(c.Request.Context(), userID, listID, taskID); err != nil {
		log.Printf("Failed to complete task: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// SyncTasks syncs all tasks from all lists.
func (h *Handler) SyncTasks(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	result, err := h.todo.SyncAllTasks(c.Request.Context(), userID)
	if err != nil {
		log.Printf("Failed to sync tasks: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sync tasks"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

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
