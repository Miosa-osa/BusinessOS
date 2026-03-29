package handlers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/database/sqlc"
	"github.com/rhl/businessos-backend/internal/middleware"
	"github.com/rhl/businessos-backend/internal/services"
)

// MobileHandler implements the Mobile API (/api/mobile/v1)
// Uses lean payloads, cursor pagination, and Unix timestamps
type MobileHandler struct {
	pool                *pgxpool.Pool
	queries             *sqlc.Queries
	notificationService *services.NotificationService
}

// NewMobileHandler creates a new mobile handler instance
func NewMobileHandler(pool *pgxpool.Pool, notificationService *services.NotificationService) *MobileHandler {
	return &MobileHandler{
		pool:                pool,
		queries:             sqlc.New(pool),
		notificationService: notificationService,
	}
}

func (h *MobileHandler) GetMe(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		MobileRespondUnauthorized(c)
		return
	}

	// Build user response
	userResp := MobileUserResponse{
		ID:    user.ID,
		Email: user.Email,
		Name:  user.Name,
	}

	// Get avatar URL if available (Image is *string)
	if user.Image != nil && *user.Image != "" {
		userResp.AvatarURL = user.Image
	}

	// Get timezone from notification preferences (defaults to UTC)
	timezone := "UTC"

	// Build preferences from notification_preferences table
	prefs := &MobilePreferencesResponse{
		NotificationsEnabled: true, // Default
	}

	// Load from notification_preferences table
	notifPrefs, err := h.queries.GetNotificationPreferencesByUser(c.Request.Context(), user.ID)
	if err == nil {
		// Extract timezone from quiet hours settings if available
		if notifPrefs.QuietHoursTimezone != nil && *notifPrefs.QuietHoursTimezone != "" {
			timezone = *notifPrefs.QuietHoursTimezone
		}


		// InAppEnabled is *bool, default to true if nil
		if notifPrefs.InAppEnabled != nil {
			prefs.NotificationsEnabled = *notifPrefs.InAppEnabled
		}
		// Handle quiet hours - QuietHoursEnabled is *bool
		if notifPrefs.QuietHoursEnabled != nil && *notifPrefs.QuietHoursEnabled {
			if notifPrefs.QuietHoursStart.Valid {
				startStr := notifPrefs.QuietHoursStart.Microseconds / 1000000 // Convert to seconds
				hours := startStr / 3600
				mins := (startStr % 3600) / 60
				timeStr := fmt.Sprintf("%02d:%02d", hours, mins)
				prefs.QuietHoursStart = &timeStr
			}
			if notifPrefs.QuietHoursEnd.Valid {
				endStr := notifPrefs.QuietHoursEnd.Microseconds / 1000000
				hours := endStr / 3600
				mins := (endStr % 3600) / 60
				timeStr := fmt.Sprintf("%02d:%02d", hours, mins)
				prefs.QuietHoursEnd = &timeStr
			}
		}
	}

	// Set timezone in user response
	userResp.Timezone = timezone

	// Build full response
	response := MobileMeResponse{
		User:        userResp,
		Workspace:   nil, // Single-tenant for now, skip workspace
		Preferences: prefs,
	}

	c.JSON(http.StatusOK, response)
}

// ListTasks returns paginated tasks
// Query: limit (1-50), cursor, status, due (today|week|overdue), fields
func (h *MobileHandler) ListTasks(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		MobileRespondUnauthorized(c)
		return
	}

	// Parse pagination
	limit := 20
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = ClampInt(parsed, 1, 50)
		}
	}

	// Parse cursor
	var cursorID uuid.UUID
	var cursorTime time.Time
	if cursor := c.Query("cursor"); cursor != "" {
		var err error
		cursorID, cursorTime, err = DecodeCursor(cursor)
		if err != nil {
			MobileRespondInvalidCursor(c)
			return
		}
	}

	// Parse filters
	statusFilter := c.Query("status")
	dueFilter := c.Query("due")

	// Parse field selection
	fields := ParseFieldsParam(c.Query("fields"))

	// Build query params
	params := sqlc.ListTasksForMobileParams{
		UserID:     user.ID,
		LimitCount: int32(limit + 1), // +1 to detect has_more
	}

	// Set cursor params if provided
	if !cursorTime.IsZero() {
		params.CursorUpdatedAt = pgtype.Timestamp{Time: cursorTime, Valid: true}
		params.CursorID = pgtype.UUID{Bytes: cursorID, Valid: true}
	}

	// Set status filter (note: sqlc uses Taskstatus not TaskStatus)
	if statusFilter != "" {
		params.Status = sqlc.NullTaskstatus{
			Taskstatus: sqlc.Taskstatus(statusFilter),
			Valid:      true,
		}
	}

	// Set due filter (DueFilter is *string)
	if dueFilter != "" {
		params.DueFilter = &dueFilter
	}

	// Execute query
	ctx := c.Request.Context()
	rows, err := h.queries.ListTasksForMobile(ctx, params)
	if err != nil {
		MobileRespondInternalError(c)
		return
	}

	// Check if there are more results
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit] // Trim to requested limit
	}

	// Transform to mobile response
	tasks := make([]interface{}, 0, len(rows))
	var lastTask sqlc.ListTasksForMobileRow

	for _, row := range rows {
		task := transformMobileTaskRow(row)

		// Apply field selection if specified
		if len(fields) > 0 {
			tasks = append(tasks, SelectFields(task, fields))
		} else {
			tasks = append(tasks, task)
		}

		lastTask = row
	}

	// Build cursor for next page
	nextCursor := ""
	if hasMore && len(rows) > 0 {
		nextCursor = EncodeCursor(lastTask.ID.Bytes, lastTask.UpdatedAt.Time)
	}

	// Build response
	response := MobileTaskListResponse{
		Tasks:   tasks,
		Cursor:  nextCursor,
		HasMore: hasMore,
	}

	c.JSON(http.StatusOK, response)
}

// transformMobileTaskRow converts a ListTasksForMobileRow to MobileTaskResponse
func transformMobileTaskRow(row sqlc.ListTasksForMobileRow) MobileTaskResponse {
	resp := MobileTaskResponse{
		ID:        row.ID.Bytes,
		Title:     row.Title,
		Status:    string(row.Status.Taskstatus),
		Priority:  string(row.Priority.Taskpriority),
		UpdatedAt: row.UpdatedAt.Time.Unix(),
	}

	// Handle nullable fields
	if row.DueDate.Valid {
		s := row.DueDate.Time.Format("2006-01-02")
		resp.DueDate = &s
	}

	if row.ProjectName != nil {
		resp.Project = row.ProjectName
	}

	if row.AssigneeName != nil {
		resp.Assignee = row.AssigneeName
	}

	return resp
}

func (h *MobileHandler) GetTask(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		MobileRespondUnauthorized(c)
		return
	}

	// Parse task ID
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		MobileRespondValidationError(c, "id", "valid UUID")
		return
	}

	ctx := c.Request.Context()

	// Get task with details using sqlc
	row, err := h.queries.GetTaskForMobile(ctx, sqlc.GetTaskForMobileParams{
		ID:     pgtype.UUID{Bytes: taskID, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		MobileRespondNotFound(c, "Task")
		return
	}

	// Get subtask count
	subtaskCount, _ := h.queries.CountSubtasksForTask(ctx, pgtype.UUID{Bytes: taskID, Valid: true})

	// Get comment count
	commentsCount, _ := h.queries.CountCommentsByEntity(ctx, sqlc.CountCommentsByEntityParams{
		EntityType: "task",
		EntityID:   pgtype.UUID{Bytes: taskID, Valid: true},
	})

	// Transform to detail response
	response := transformMobileTaskDetailRow(row, int(subtaskCount), int(commentsCount))

	c.JSON(http.StatusOK, response)
}

// transformMobileTaskDetailRow converts GetTaskForMobileRow to MobileTaskDetailResponse
func transformMobileTaskDetailRow(row sqlc.GetTaskForMobileRow, subtaskCount int, commentsCount int) MobileTaskDetailResponse {
	resp := MobileTaskDetailResponse{
		ID:            row.ID.Bytes,
		Title:         row.Title,
		Status:        string(row.Status.Taskstatus),
		Priority:      string(row.Priority.Taskpriority),
		Tags:          []string{}, // Tags not yet implemented
		CommentsCount: commentsCount,
		SubtasksCount: subtaskCount,
		CreatedAt:     row.CreatedAt.Time.Unix(),
		UpdatedAt:     row.UpdatedAt.Time.Unix(),
	}

	// Handle nullable description
	if row.Description != nil {
		resp.Description = row.Description
	}

	// Handle nullable dates
	if row.DueDate.Valid {
		s := row.DueDate.Time.Format("2006-01-02")
		resp.DueDate = &s
	}

	if row.StartDate.Valid {
		s := row.StartDate.Time.Format("2006-01-02")
		resp.StartDate = &s
	}

	if row.CompletedAt.Valid {
		s := row.CompletedAt.Time.Format(time.RFC3339)
		resp.CompletedAt = &s
	}

	// Handle project reference
	if row.ProjectID.Valid && row.ProjectName != nil {
		resp.Project = &MobileProjectRefResponse{
			ID:   row.ProjectID.Bytes,
			Name: *row.ProjectName,
		}
	}

	// Handle assignee reference
	if row.AssigneeUuid.Valid && row.AssigneeName != nil {
		resp.Assignee = &MobileAssigneeResponse{
			ID:   uuid.UUID(row.AssigneeUuid.Bytes).String(),
			Name: *row.AssigneeName,
		}
		if row.AssigneeAvatar != nil {
			resp.Assignee.AvatarURL = row.AssigneeAvatar
		}
	}

	return resp
}

func (h *MobileHandler) QuickCreateTask(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		MobileRespondUnauthorized(c)
		return
	}

	var req MobileTaskQuickCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		MobileRespondValidationError(c, "body", "valid JSON with title field")
		return
	}

	// Build query params
	params := sqlc.QuickCreateTaskParams{
		UserID: user.ID,
		Title:  req.Title,
	}

	// Parse optional due date
	if req.DueDate != nil {
		if t, err := time.Parse("2006-01-02", *req.DueDate); err == nil {
			params.DueDate = pgtype.Timestamp{Time: t, Valid: true}
		} else {
			MobileRespondValidationError(c, "due_date", "YYYY-MM-DD format")
			return
		}
	}

	// Parse optional priority
	if req.Priority != nil {
		params.Priority = sqlc.NullTaskpriority{
			Taskpriority: sqlc.Taskpriority(*req.Priority),
			Valid:        true,
		}
	}

	// Create task
	ctx := c.Request.Context()
	task, err := h.queries.QuickCreateTask(ctx, params)
	if err != nil {
		MobileRespondInternalError(c)
		return
	}

	// Build response
	response := MobileTaskResponse{
		ID:        task.ID.Bytes,
		Title:     task.Title,
		Status:    string(task.Status.Taskstatus),
		Priority:  string(task.Priority.Taskpriority),
		UpdatedAt: task.CreatedAt.Time.Unix(),
	}

	if task.DueDate.Valid {
		s := task.DueDate.Time.Format("2006-01-02")
		response.DueDate = &s
	}

	c.JSON(http.StatusCreated, response)
}

func (h *MobileHandler) UpdateTaskStatus(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		MobileRespondUnauthorized(c)
		return
	}

	// Parse task ID
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		MobileRespondValidationError(c, "id", "valid UUID")
		return
	}

	var req MobileTaskStatusUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		MobileRespondValidationError(c, "status", "one of: todo, in_progress, done, cancelled")
		return
	}

	// Update status with ownership check in query
	ctx := c.Request.Context()
	result, err := h.queries.UpdateTaskStatusMobile(ctx, sqlc.UpdateTaskStatusMobileParams{
		ID:     pgtype.UUID{Bytes: taskID, Valid: true},
		Status: sqlc.Taskstatus(req.Status),
		UserID: user.ID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			MobileRespondNotFound(c, "Task")
			return
		}
		MobileRespondInternalError(c)
		return
	}

	// Build response
	response := MobileTaskStatusResponse{
		ID:        result.ID.Bytes,
		Status:    string(result.Status.Taskstatus),
		UpdatedAt: result.UpdatedAt.Time.Unix(),
	}

	if result.CompletedAt.Valid {
		s := result.CompletedAt.Time.Format(time.RFC3339)
		response.CompletedAt = &s
	}

	c.JSON(http.StatusOK, response)
}

// ToggleTask toggles between todo/done (swipe gesture)
func (h *MobileHandler) ToggleTask(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		MobileRespondUnauthorized(c)
		return
	}

	// Parse task ID
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		MobileRespondValidationError(c, "id", "valid UUID")
		return
	}

	// Toggle status with ownership check in query
	ctx := c.Request.Context()
	result, err := h.queries.ToggleTaskStatusMobile(ctx, sqlc.ToggleTaskStatusMobileParams{
		ID:     pgtype.UUID{Bytes: taskID, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			MobileRespondNotFound(c, "Task")
			return
		}
		MobileRespondInternalError(c)
		return
	}

	// Build response
	response := MobileTaskStatusResponse{
		ID:        result.ID.Bytes,
		Status:    string(result.Status.Taskstatus),
		UpdatedAt: result.UpdatedAt.Time.Unix(),
	}

	if result.CompletedAt.Valid {
		s := result.CompletedAt.Time.Format(time.RFC3339)
		response.CompletedAt = &s
	}

	c.JSON(http.StatusOK, response)
}

func (h *MobileHandler) ListNotifications(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		MobileRespondUnauthorized(c)
		return
	}

	limitStr := c.DefaultQuery("limit", "20")
	limit, _ := strconv.Atoi(limitStr)
	limit = ClampInt(limit, 1, 50)
	unreadOnly := c.Query("unread") == "true"

	var cursorTime pgtype.Timestamptz
	var cursorID pgtype.UUID
	if cursor := c.Query("cursor"); cursor != "" {
		id, ts, err := DecodeCursor(cursor)
		if err == nil && id != uuid.Nil {
			cursorTime = pgtype.Timestamptz{Time: ts, Valid: true}
			cursorID = pgtype.UUID{Bytes: id, Valid: true}
		}
	}

	rows, err := h.queries.ListNotificationsForMobile(c.Request.Context(), sqlc.ListNotificationsForMobileParams{
		UserID:          user.ID,
		UnreadOnly:      &unreadOnly,
		CursorCreatedAt: cursorTime,
		CursorID:        cursorID,
		LimitCount:      int32(limit + 1),
	})
	if err != nil {
		MobileRespondError(c, http.StatusInternalServerError, ErrCodeInternal, "Failed to fetch notifications")
		return
	}

	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}

	notifications := make([]MobileNotificationResponse, 0, len(rows))
	for _, row := range rows {
		notifications = append(notifications, TransformToMobileNotification(row))
	}

	var nextCursor string
	if hasMore && len(rows) > 0 {
		last := rows[len(rows)-1]
		nextCursor = EncodeCursor(uuid.UUID(last.ID.Bytes), last.CreatedAt.Time)
	}

	unreadCount, _ := h.queries.GetUnreadNotificationCount(c.Request.Context(), user.ID)

	c.JSON(http.StatusOK, MobileNotificationListResponse{
		Notifications: notifications,
		Cursor:        nextCursor,
		HasMore:       hasMore,
		UnreadCount:   int(unreadCount),
	})
}

func (h *MobileHandler) GetNotificationCount(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		MobileRespondUnauthorized(c)
		return
	}

	count, err := h.queries.GetUnreadNotificationCount(c.Request.Context(), user.ID)
	if err != nil {
		MobileRespondError(c, http.StatusInternalServerError, ErrCodeInternal, "Failed to get count")
		return
	}

	c.JSON(http.StatusOK, MobileNotificationCountResponse{
		UnreadCount: int(count),
	})
}

func (h *MobileHandler) MarkNotificationsRead(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		MobileRespondUnauthorized(c)
		return
	}

	var req MobileNotificationReadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		MobileRespondError(c, http.StatusBadRequest, ErrCodeValidation, "Invalid request body")
		return
	}

	var markedCount int64
	var err error

	if req.All {
		markedCount, err = h.queries.MarkAllNotificationsAsRead(c.Request.Context(), user.ID)
		if err != nil {
			MobileRespondError(c, http.StatusInternalServerError, ErrCodeInternal, "Failed to mark notifications")
			return
		}
	} else if len(req.IDs) > 0 {
		ids := make([]pgtype.UUID, len(req.IDs))
		for i, id := range req.IDs {
			ids[i] = pgtype.UUID{Bytes: id, Valid: true}
		}
		err = h.queries.MarkNotificationsAsRead(c.Request.Context(), sqlc.MarkNotificationsAsReadParams{
			UserID: user.ID,
			Ids:    ids,
		})
		if err != nil {
			MobileRespondError(c, http.StatusInternalServerError, ErrCodeInternal, "Failed to mark notifications")
			return
		}
		markedCount = int64(len(req.IDs))
	}

	unreadCount, _ := h.queries.GetUnreadNotificationCount(c.Request.Context(), user.ID)

	c.JSON(http.StatusOK, MobileNotificationReadResponse{
		MarkedCount: int(markedCount),
		UnreadCount: int(unreadCount),
	})
}

func (h *MobileHandler) GetTodayLog(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		MobileRespondUnauthorized(c)
		return
	}

	row, err := h.queries.GetTodayDailyLogForMobile(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusOK, MobileDailyLogTodayResponse{
			Date:    time.Now().Format("2006-01-02"),
			Entries: []MobileDailyLogEntryResponse{},
		})
		return
	}

	c.JSON(http.StatusOK, MobileDailyLogTodayResponse{
		Date:        row.Date.Time.Format("2006-01-02"),
		Content:     row.Content,
		EnergyLevel: row.EnergyLevel,
	})
}

func (h *MobileHandler) GetLogHistory(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		MobileRespondUnauthorized(c)
		return
	}

	limitStr := c.DefaultQuery("limit", "14")
	limit, _ := strconv.Atoi(limitStr)
	limit = ClampInt(limit, 1, 30)

	var beforeDate pgtype.Date
	if before := c.Query("before"); before != "" {
		if t, err := time.Parse("2006-01-02", before); err == nil {
			beforeDate = pgtype.Date{Time: t, Valid: true}
		}
	}

	rows, err := h.queries.GetDailyLogHistoryForMobile(c.Request.Context(), sqlc.GetDailyLogHistoryForMobileParams{
		UserID:     user.ID,
		BeforeDate: beforeDate,
		LimitCount: int32(limit + 1),
	})
	if err != nil {
		MobileRespondError(c, http.StatusInternalServerError, ErrCodeInternal, "Failed to fetch history")
		return
	}

	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}

	logs := make([]MobileDailyLogHistoryItem, 0, len(rows))
	for _, row := range rows {
		logs = append(logs, TransformToDailyLogHistoryItem(row))
	}

	var nextBefore string
	if hasMore && len(rows) > 0 {
		nextBefore = rows[len(rows)-1].Date.Time.Format("2006-01-02")
	}

	c.JSON(http.StatusOK, MobileDailyLogHistoryResponse{
		Logs:    logs,
		HasMore: hasMore,
		Before:  nextBefore,
	})
}

func (h *MobileHandler) DeltaSync(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		MobileRespondUnauthorized(c)
		return
	}

	sinceStr := c.Query("since")
	if sinceStr == "" {
		MobileRespondError(c, http.StatusBadRequest, ErrCodeValidation, "since parameter required")
		return
	}

	sinceUnix, err := strconv.ParseInt(sinceStr, 10, 64)
	if err != nil {
		MobileRespondError(c, http.StatusBadRequest, ErrCodeValidation, "Invalid since timestamp")
		return
	}
	since := time.Unix(sinceUnix, 0)

	limitStr := c.DefaultQuery("limit", "100")
	limit, _ := strconv.Atoi(limitStr)
	limit = ClampInt(limit, 1, 500)

	ctx := c.Request.Context()
	sincePg := pgtype.Timestamp{Time: since, Valid: true}
	sincePgTz := pgtype.Timestamptz{Time: since, Valid: true}

	taskRows, _ := h.queries.GetTaskChangesSince(ctx, sqlc.GetTaskChangesSinceParams{
		UserID:     user.ID,
		UpdatedAt:  sincePg,
		LimitCount: int32(limit),
	})

	notifRows, _ := h.queries.GetNotificationChangesSince(ctx, sqlc.GetNotificationChangesSinceParams{
		UserID:     user.ID,
		CreatedAt:  sincePgTz,
		LimitCount: int32(limit),
	})

	tasks := make([]MobileTaskResponse, 0, len(taskRows))
	var latestTaskTime int64
	for _, row := range taskRows {
		tasks = append(tasks, TransformSyncTaskRow(row))
		if row.UpdatedAt.Valid && row.UpdatedAt.Time.Unix() > latestTaskTime {
			latestTaskTime = row.UpdatedAt.Time.Unix()
		}
	}

	notifications := make([]MobileNotificationResponse, 0, len(notifRows))
	var latestNotifTime int64
	for _, row := range notifRows {
		notifications = append(notifications, TransformSyncNotificationRow(row))
		if row.CreatedAt.Valid && row.CreatedAt.Time.Unix() > latestNotifTime {
			latestNotifTime = row.CreatedAt.Time.Unix()
		}
	}

	serverTime := time.Now().Unix()
	hasMore := len(taskRows) >= limit || len(notifRows) >= limit

	c.JSON(http.StatusOK, MobileSyncResponse{
		Tasks:         tasks,
		Notifications: notifications,
		ServerTime:    serverTime,
		HasMore:       hasMore,
	})
}

func (h *MobileHandler) ListChatThreads(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		MobileRespondUnauthorized(c)
		return
	}

	limitStr := c.DefaultQuery("limit", "20")
	limit, _ := strconv.Atoi(limitStr)
	limit = ClampInt(limit, 1, 50)

	rows, err := h.queries.ListConversationsForMobile(c.Request.Context(), sqlc.ListConversationsForMobileParams{
		UserID:     user.ID,
		LimitCount: int32(limit),
	})
	if err != nil {
		MobileRespondError(c, http.StatusInternalServerError, ErrCodeInternal, "Failed to fetch threads")
		return
	}

	threads := make([]MobileChatThreadResponse, 0, len(rows))
	for _, row := range rows {
		threads = append(threads, TransformToChatThread(row))
	}

	c.JSON(http.StatusOK, MobileChatThreadListResponse{
		Threads: threads,
	})
}

func (h *MobileHandler) GetChatHistory(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		MobileRespondUnauthorized(c)
		return
	}

	conversationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		MobileRespondError(c, http.StatusBadRequest, ErrCodeValidation, "Invalid conversation ID")
		return
	}

	limitStr := c.DefaultQuery("limit", "30")
	limit, _ := strconv.Atoi(limitStr)
	limit = ClampInt(limit, 1, 100)

	var cursorTime pgtype.Timestamp
	var cursorID pgtype.UUID
	if cursor := c.Query("cursor"); cursor != "" {
		id, ts, err := DecodeCursor(cursor)
		if err == nil && id != uuid.Nil {
			cursorTime = pgtype.Timestamp{Time: ts, Valid: true}
			cursorID = pgtype.UUID{Bytes: id, Valid: true}
		}
	}

	rows, err := h.queries.GetMessagesForMobile(c.Request.Context(), sqlc.GetMessagesForMobileParams{
		ConversationID:  pgtype.UUID{Bytes: conversationID, Valid: true},
		CursorCreatedAt: cursorTime,
		CursorID:        cursorID,
		LimitCount:      int32(limit + 1),
	})
	if err != nil {
		MobileRespondError(c, http.StatusInternalServerError, ErrCodeInternal, "Failed to fetch messages")
		return
	}

	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}

	messages := make([]MobileChatMessageResponse, 0, len(rows))
	for _, row := range rows {
		messages = append(messages, TransformToChatMessage(row))
	}

	var nextCursor string
	if hasMore && len(rows) > 0 {
		last := rows[len(rows)-1]
		nextCursor = EncodeCursor(uuid.UUID(last.ID.Bytes), last.CreatedAt.Time)
	}

	c.JSON(http.StatusOK, MobileChatHistoryResponse{
		Messages: messages,
		Cursor:   nextCursor,
		HasMore:  hasMore,
	})
}

func (h *MobileHandler) SendChatMessage(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		MobileRespondUnauthorized(c)
		return
	}

	var req MobileChatMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		MobileRespondError(c, http.StatusBadRequest, ErrCodeValidation, "Invalid request")
		return
	}

	ctx := c.Request.Context()

	_, err := h.queries.CreateMessage(ctx, sqlc.CreateMessageParams{
		ConversationID:  pgtype.UUID{Bytes: req.ConversationID, Valid: true},
		Role:            sqlc.MessageroleUSER,
		Content:         req.Content,
		MessageMetadata: nil,
	})
	if err != nil {
		MobileRespondError(c, http.StatusInternalServerError, ErrCodeInternal, "Failed to save message")
		return
	}

	// For now, return acknowledgment - streaming response would be separate
	c.JSON(http.StatusOK, gin.H{
		"status":  "received",
		"message": "Message saved. AI response will be streamed separately.",
	})
}

func (h *MobileHandler) RegisterPushDevice(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		MobileRespondUnauthorized(c)
		return
	}

	var req MobilePushRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		MobileRespondError(c, http.StatusBadRequest, ErrCodeValidation, "Invalid request")
		return
	}

	deviceID := middleware.GetDeviceID(c)
	if deviceID == "" {
		deviceID = req.DeviceID
	}

	_, err := h.queries.RegisterPushDevice(c.Request.Context(), sqlc.RegisterPushDeviceParams{
		UserID:      user.ID,
		DeviceID:    deviceID,
		Platform:    req.Platform,
		PushToken:   req.PushToken,
		AppVersion:  &req.AppVersion,
		OsVersion:   &req.OsVersion,
		DeviceModel: &req.DeviceModel,
	})
	if err != nil {
		MobileRespondError(c, http.StatusInternalServerError, ErrCodeInternal, "Failed to register device")
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "registered"})
}

func (h *MobileHandler) UnregisterPushDevice(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		MobileRespondUnauthorized(c)
		return
	}

	deviceID := middleware.GetDeviceID(c)
	if deviceID == "" {
		deviceID = c.Query("device_id")
	}

	if deviceID == "" {
		MobileRespondError(c, http.StatusBadRequest, ErrCodeValidation, "Device ID required")
		return
	}

	err := h.queries.UnregisterPushDevice(c.Request.Context(), sqlc.UnregisterPushDeviceParams{
		UserID:   user.ID,
		DeviceID: deviceID,
	})
	if err != nil {
		MobileRespondError(c, http.StatusInternalServerError, ErrCodeInternal, "Failed to unregister")
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "unregistered"})
}

func (h *MobileHandler) verifyTaskOwnership(ctx context.Context, taskID uuid.UUID, userID string) bool {
	_, err := h.queries.GetTask(ctx, sqlc.GetTaskParams{
		ID:     pgtype.UUID{Bytes: taskID, Valid: true},
		UserID: userID,
	})
	return err == nil
}
