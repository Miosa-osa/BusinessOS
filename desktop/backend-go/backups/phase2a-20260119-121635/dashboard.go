package handlers

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rhl/businessos-backend/internal/database/sqlc"
	"github.com/rhl/businessos-backend/internal/middleware"
	"github.com/rhl/businessos-backend/internal/services"
)

// GetDashboardSummary returns a summary of the user's data
func (h *Handlers) GetDashboardSummary(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	queries := sqlc.New(h.pool)

	// Get data for various entities
	projectRows, _ := queries.ListProjects(c.Request.Context(), sqlc.ListProjectsParams{UserID: user.ID})
	clients, _ := queries.ListClients(c.Request.Context(), sqlc.ListClientsParams{UserID: user.ID})
	contexts, _ := queries.ListContexts(c.Request.Context(), sqlc.ListContextsParams{UserID: user.ID})
	artifacts, _ := queries.ListArtifacts(c.Request.Context(), sqlc.ListArtifactsParams{UserID: user.ID})
	tasks, _ := queries.ListTasks(c.Request.Context(), sqlc.ListTasksParams{UserID: user.ID})

	// Get today's focus items
	today := time.Now()
	focusItems, _ := queries.ListFocusItems(c.Request.Context(), sqlc.ListFocusItemsParams{
		UserID:    user.ID,
		FocusDate: pgtype.Date{Time: today, Valid: true},
	})

	// Ensure arrays are not nil (return empty arrays instead)
	if projectRows == nil {
		projectRows = []sqlc.ListProjectsRow{}
	}
	if clients == nil {
		clients = []sqlc.Client{}
	}
	if contexts == nil {
		contexts = []sqlc.Context{}
	}
	if artifacts == nil {
		artifacts = []sqlc.Artifact{}
	}
	if tasks == nil {
		tasks = []sqlc.Task{}
	}
	if focusItems == nil {
		focusItems = []sqlc.FocusItem{}
	}

	c.JSON(http.StatusOK, gin.H{
		"projects":       TransformProjectRows(projectRows),
		"clients":        TransformClients(clients),
		"contexts":       TransformContexts(contexts),
		"artifacts":      TransformArtifacts(artifacts),
		"tasks":          TransformTasks(tasks),
		"focus_items":    TransformFocusItems(focusItems),
		"activities":     []interface{}{}, // Placeholder for activities
		"energy_level":   3,               // Default energy level (1-5 scale)
		"project_count":  len(projectRows),
		"client_count":   len(clients),
		"context_count":  len(contexts),
		"artifact_count": len(artifacts),
		"task_count":     len(tasks),
	})
}

// ListFocusItems returns focus items for a specific date
func (h *Handlers) ListFocusItems(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	queries := sqlc.New(h.pool)

	// Parse date from query, default to today
	dateStr := c.Query("date")
	focusDate := time.Now()
	if dateStr != "" {
		if t, err := time.Parse("2006-01-02", dateStr); err == nil {
			focusDate = t
		}
	}

	items, err := queries.ListFocusItems(c.Request.Context(), sqlc.ListFocusItemsParams{
		UserID:    user.ID,
		FocusDate: pgtype.Date{Time: focusDate, Valid: true},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list focus items"})
		return
	}

	c.JSON(http.StatusOK, TransformFocusItems(items))
}

// CreateFocusItem creates a new focus item
func (h *Handlers) CreateFocusItem(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var req struct {
		Text      string  `json:"text" binding:"required"`
		FocusDate *string `json:"focus_date"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	queries := sqlc.New(h.pool)

	// Parse focus date, default to today
	focusDate := time.Now()
	if req.FocusDate != nil {
		if t, err := time.Parse("2006-01-02", *req.FocusDate); err == nil {
			focusDate = t
		}
	}

	item, err := queries.CreateFocusItem(c.Request.Context(), sqlc.CreateFocusItemParams{
		UserID:    user.ID,
		Text:      req.Text,
		FocusDate: pgtype.Timestamp{Time: focusDate, Valid: true},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create focus item"})
		return
	}

	c.JSON(http.StatusCreated, TransformFocusItem(item))
}

// UpdateFocusItem updates a focus item
func (h *Handlers) UpdateFocusItem(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid focus item ID"})
		return
	}

	var req struct {
		Text      string `json:"text" binding:"required"`
		Completed *bool  `json:"completed"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	queries := sqlc.New(h.pool)

	item, err := queries.UpdateFocusItem(c.Request.Context(), sqlc.UpdateFocusItemParams{
		ID:        pgtype.UUID{Bytes: id, Valid: true},
		Text:      req.Text,
		Completed: req.Completed,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update focus item"})
		return
	}

	c.JSON(http.StatusOK, TransformFocusItem(item))
}

// DeleteFocusItem deletes a focus item
func (h *Handlers) DeleteFocusItem(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid focus item ID"})
		return
	}

	queries := sqlc.New(h.pool)
	err = queries.DeleteFocusItem(c.Request.Context(), sqlc.DeleteFocusItemParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete focus item"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Focus item deleted"})
}

// ListTasks returns all tasks for the current user
func (h *Handlers) ListTasks(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	queries := sqlc.New(h.pool)

	// Parse optional filters
	statusQuery := c.Query("status")
	var status sqlc.NullTaskstatus
	if statusQuery != "" {
		status = sqlc.NullTaskstatus{Taskstatus: stringToTaskStatus(statusQuery), Valid: true}
	}

	priorityQuery := c.Query("priority")
	var priority sqlc.NullTaskpriority
	if priorityQuery != "" {
		priority = sqlc.NullTaskpriority{Taskpriority: stringToTaskPriority(priorityQuery), Valid: true}
	}

	var projectID pgtype.UUID
	if pid := c.Query("project_id"); pid != "" {
		if parsed, err := uuid.Parse(pid); err == nil {
			projectID = pgtype.UUID{Bytes: parsed, Valid: true}
		}
	}

	tasks, err := queries.ListTasks(c.Request.Context(), sqlc.ListTasksParams{
		UserID:    user.ID,
		Status:    status,
		Priority:  priority,
		ProjectID: projectID,
	})
	if err != nil {
		log.Printf("[ListTasks] Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list tasks: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, TransformTasks(tasks))
}

// CreateTask creates a new task
func (h *Handlers) CreateTask(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var req struct {
		Title       string  `json:"title" binding:"required"`
		Description *string `json:"description"`
		Status      *string `json:"status"`
		Priority    *string `json:"priority"`
		DueDate     *string `json:"due_date"`
		ProjectID   *string `json:"project_id"`
		AssigneeID  *string `json:"assignee_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	queries := sqlc.New(h.pool)

	// Parse optional fields
	var status sqlc.NullTaskstatus
	if req.Status != nil {
		status = sqlc.NullTaskstatus{
			Taskstatus: stringToTaskStatus(*req.Status),
			Valid:      true,
		}
	}

	var priority sqlc.NullTaskpriority
	if req.Priority != nil {
		priority = sqlc.NullTaskpriority{
			Taskpriority: stringToTaskPriority(*req.Priority),
			Valid:        true,
		}
	}

	var dueDate pgtype.Timestamp
	if req.DueDate != nil {
		if t, err := time.Parse(time.RFC3339, *req.DueDate); err == nil {
			dueDate = pgtype.Timestamp{Time: t, Valid: true}
		}
	}

	var projectID, assigneeID pgtype.UUID
	if req.ProjectID != nil {
		if parsed, err := uuid.Parse(*req.ProjectID); err == nil {
			projectID = pgtype.UUID{Bytes: parsed, Valid: true}
		}
	}
	if req.AssigneeID != nil {
		if parsed, err := uuid.Parse(*req.AssigneeID); err == nil {
			assigneeID = pgtype.UUID{Bytes: parsed, Valid: true}
		}
	}

	task, err := queries.CreateTask(c.Request.Context(), sqlc.CreateTaskParams{
		UserID:      user.ID,
		Title:       req.Title,
		Description: req.Description,
		Status:      status,
		Priority:    priority,
		DueDate:     dueDate,
		ProjectID:   projectID,
		AssigneeID:  assigneeID,
	})
	if err != nil {
		log.Printf("[CreateTask] Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task: " + err.Error()})
		return
	}

	// Trigger notification if task was assigned to someone else
	if h.notificationTriggers != nil && req.AssigneeID != nil && *req.AssigneeID != user.ID {
		taskID := uuid.UUID(task.ID.Bytes)
		var projID *uuid.UUID
		if task.ProjectID.Valid {
			id := uuid.UUID(task.ProjectID.Bytes)
			projID = &id
		}
		go h.notificationTriggers.OnTaskAssigned(c.Request.Context(), services.TaskAssignedInput{
			TaskID:       taskID,
			TaskTitle:    task.Title,
			AssigneeID:   *req.AssigneeID,
			AssignerID:   user.ID,
			AssignerName: user.Name,
			ProjectID:    projID,
		})
	}

	c.JSON(http.StatusCreated, TransformTask(task))
}

// UpdateTask updates a task
func (h *Handlers) UpdateTask(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	var req struct {
		Title       *string `json:"title"`
		Description *string `json:"description"`
		Status      *string `json:"status"`
		Priority    *string `json:"priority"`
		DueDate     *string `json:"due_date"`
		ProjectID   *string `json:"project_id"`
		AssigneeID  *string `json:"assignee_id"`
		Position    *int32  `json:"position"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	queries := sqlc.New(h.pool)

	// Get existing task (for comparison and ownership verification)
	existingTask, err := queries.GetTask(c.Request.Context(), sqlc.GetTaskParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	// Parse optional fields
	var status sqlc.NullTaskstatus
	if req.Status != nil {
		status = sqlc.NullTaskstatus{
			Taskstatus: stringToTaskStatus(*req.Status),
			Valid:      true,
		}
	}

	var priority sqlc.NullTaskpriority
	if req.Priority != nil {
		priority = sqlc.NullTaskpriority{
			Taskpriority: stringToTaskPriority(*req.Priority),
			Valid:        true,
		}
	}

	var dueDate pgtype.Timestamp
	if req.DueDate != nil {
		if t, err := time.Parse(time.RFC3339, *req.DueDate); err == nil {
			dueDate = pgtype.Timestamp{Time: t, Valid: true}
		}
	}

	var projectID, assigneeID pgtype.UUID
	if req.ProjectID != nil {
		if parsed, err := uuid.Parse(*req.ProjectID); err == nil {
			projectID = pgtype.UUID{Bytes: parsed, Valid: true}
		}
	}
	if req.AssigneeID != nil {
		if parsed, err := uuid.Parse(*req.AssigneeID); err == nil {
			assigneeID = pgtype.UUID{Bytes: parsed, Valid: true}
		}
	}

	task, err := queries.UpdateTask(c.Request.Context(), sqlc.UpdateTaskParams{
		ID:          pgtype.UUID{Bytes: id, Valid: true},
		Title:       req.Title,
		Description: req.Description,
		Status:      status,
		Priority:    priority,
		DueDate:     dueDate,
		ProjectID:   projectID,
		AssigneeID:  assigneeID,
		Position:    req.Position,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task"})
		return
	}

	// Trigger notifications for changes
	if h.notificationTriggers != nil {
		taskID := uuid.UUID(task.ID.Bytes)
		var projID *uuid.UUID
		if task.ProjectID.Valid {
			pid := uuid.UUID(task.ProjectID.Bytes)
			projID = &pid
		}

		// Check if assignee changed (new assignment)
		oldAssigneeID := ""
		if existingTask.AssigneeID.Valid {
			oldAssigneeID = uuid.UUID(existingTask.AssigneeID.Bytes).String()
		}
		newAssigneeID := ""
		if task.AssigneeID.Valid {
			newAssigneeID = uuid.UUID(task.AssigneeID.Bytes).String()
		}

		if newAssigneeID != "" && newAssigneeID != oldAssigneeID && newAssigneeID != user.ID {
			go h.notificationTriggers.OnTaskAssigned(c.Request.Context(), services.TaskAssignedInput{
				TaskID:       taskID,
				TaskTitle:    task.Title,
				AssigneeID:   newAssigneeID,
				AssignerID:   user.ID,
				AssignerName: user.Name,
				ProjectID:    projID,
			})
		}

		// Check if status changed to completed
		if req.Status != nil {
			newStatus := strings.ToLower(*req.Status)
			oldStatus := strings.ToLower(string(existingTask.Status.Taskstatus))
			if newStatus == "completed" && oldStatus != "completed" {
				go h.notificationTriggers.OnTaskCompleted(c.Request.Context(), services.TaskCompletedInput{
					TaskID:        taskID,
					TaskTitle:     task.Title,
					CompletedByID: user.ID,
					CompletedBy:   user.Name,
					OwnerID:       task.UserID,
					ProjectID:     projID,
				})
			} else if newStatus != oldStatus && oldAssigneeID != "" && oldAssigneeID != user.ID {
				// Status changed, notify assignee
				go h.notificationTriggers.OnTaskStatusChanged(c.Request.Context(), services.TaskStatusChangedInput{
					TaskID:      taskID,
					TaskTitle:   task.Title,
					OldStatus:   oldStatus,
					NewStatus:   newStatus,
					ChangedByID: user.ID,
					ChangedBy:   user.Name,
					AssigneeID:  oldAssigneeID,
				})
			}
		}
	}

	c.JSON(http.StatusOK, TransformTask(task))
}

// ToggleTask toggles the completion status of a task
func (h *Handlers) ToggleTask(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	queries := sqlc.New(h.pool)

	// Get existing task for ownership verification and to check old status
	existingTask, err := queries.GetTask(c.Request.Context(), sqlc.GetTaskParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	task, err := queries.ToggleTaskStatus(c.Request.Context(), pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to toggle task"})
		return
	}

	// Trigger completion notification if task was just completed
	if h.notificationTriggers != nil {
		oldStatus := strings.ToLower(string(existingTask.Status.Taskstatus))
		newStatus := strings.ToLower(string(task.Status.Taskstatus))

		if newStatus == "completed" && oldStatus != "completed" {
			taskID := uuid.UUID(task.ID.Bytes)
			var projID *uuid.UUID
			if task.ProjectID.Valid {
				pid := uuid.UUID(task.ProjectID.Bytes)
				projID = &pid
			}
			go h.notificationTriggers.OnTaskCompleted(c.Request.Context(), services.TaskCompletedInput{
				TaskID:        taskID,
				TaskTitle:     task.Title,
				CompletedByID: user.ID,
				CompletedBy:   user.Name,
				OwnerID:       task.UserID,
				ProjectID:     projID,
			})
		}
	}

	c.JSON(http.StatusOK, TransformTask(task))
}

// DeleteTask deletes a task
func (h *Handlers) DeleteTask(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	queries := sqlc.New(h.pool)
	err = queries.DeleteTask(c.Request.Context(), sqlc.DeleteTaskParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task deleted"})
}

// stringToTaskStatus converts a string to sqlc.Taskstatus
func stringToTaskStatus(s string) sqlc.Taskstatus {
	typeMap := map[string]sqlc.Taskstatus{
		"todo":        sqlc.TaskstatusTodo,
		"in_progress": sqlc.TaskstatusInProgress,
		"done":        sqlc.TaskstatusDone,
		"cancelled":   sqlc.TaskstatusCancelled,
	}
	if enum, ok := typeMap[strings.ToLower(s)]; ok {
		return enum
	}
	return sqlc.TaskstatusTodo
}

// stringToTaskPriority converts a string to sqlc.Taskpriority
func stringToTaskPriority(p string) sqlc.Taskpriority {
	typeMap := map[string]sqlc.Taskpriority{
		"critical": sqlc.TaskpriorityCritical,
		"high":     sqlc.TaskpriorityHigh,
		"medium":   sqlc.TaskpriorityMedium,
		"low":      sqlc.TaskpriorityLow,
	}
	if enum, ok := typeMap[strings.ToLower(p)]; ok {
		return enum
	}
	return sqlc.TaskpriorityMedium
}
