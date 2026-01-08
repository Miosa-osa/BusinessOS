package handlers

import (
	"encoding/json"
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

// ListProjects returns all projects for the current user
func (h *Handlers) ListProjects(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	queries := sqlc.New(h.pool)

	// Parse optional status filter
	var status sqlc.NullProjectstatus
	if s := c.Query("status"); s != "" {
		status = sqlc.NullProjectstatus{Projectstatus: stringToProjectStatus(s), Valid: true}
	}

	// Parse optional priority filter
	var priority sqlc.NullProjectpriority
	if p := c.Query("priority"); p != "" {
		priority = sqlc.NullProjectpriority{Projectpriority: stringToProjectPriority(p), Valid: true}
	}

	// Parse optional client_id filter
	var clientID pgtype.UUID
	if cid := c.Query("client_id"); cid != "" {
		if id, err := uuid.Parse(cid); err == nil {
			clientID = pgtype.UUID{Bytes: id, Valid: true}
		}
	}

	projects, err := queries.ListProjects(c.Request.Context(), sqlc.ListProjectsParams{
		UserID:   user.ID,
		Status:   status,
		Priority: priority,
		ClientID: clientID,
	})
	if err != nil {
		log.Printf("ListProjects error for user %s: %v", user.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list projects"})
		return
	}

	c.JSON(http.StatusOK, TransformProjectRows(projects))
}

// CreateProject creates a new project
func (h *Handlers) CreateProject(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var req struct {
		Name            string          `json:"name" binding:"required"`
		Description     *string         `json:"description"`
		Status          *string         `json:"status"`
		Priority        *string         `json:"priority"`
		ClientName      *string         `json:"client_name"`
		ClientID        *string         `json:"client_id"`
		ProjectType     *string         `json:"project_type"`
		ProjectMetadata json.RawMessage `json:"project_metadata"`
		StartDate       *string         `json:"start_date"`
		DueDate         *string         `json:"due_date"`
		Visibility      *string         `json:"visibility"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	queries := sqlc.New(h.pool)

	// Parse status with default
	status := sqlc.NullProjectstatus{
		Projectstatus: sqlc.ProjectstatusACTIVE, // Default to "active"
		Valid:         true,
	}
	if req.Status != nil {
		status.Projectstatus = stringToProjectStatus(*req.Status)
	}

	// Parse priority with default
	priority := sqlc.NullProjectpriority{
		Projectpriority: sqlc.ProjectpriorityMEDIUM, // Default to "medium"
		Valid:           true,
	}
	if req.Priority != nil {
		priority.Projectpriority = stringToProjectPriority(*req.Priority)
	}

	// Parse client_id
	var clientID pgtype.UUID
	if req.ClientID != nil {
		if id, err := uuid.Parse(*req.ClientID); err == nil {
			clientID = pgtype.UUID{Bytes: id, Valid: true}
		}
	}

	// Handle metadata
	metadata := []byte("{}")
	if req.ProjectMetadata != nil {
		metadata = req.ProjectMetadata
	}

	// Parse dates
	var startDate pgtype.Date
	if req.StartDate != nil {
		if t, err := time.Parse("2006-01-02", *req.StartDate); err == nil {
			startDate = pgtype.Date{Time: t, Valid: true}
		}
	}

	var dueDate pgtype.Date
	if req.DueDate != nil {
		if t, err := time.Parse("2006-01-02", *req.DueDate); err == nil {
			dueDate = pgtype.Date{Time: t, Valid: true}
		}
	}

	// Set owner_id to current user
	ownerID := user.ID

	project, err := queries.CreateProject(c.Request.Context(), sqlc.CreateProjectParams{
		UserID:          user.ID,
		Name:            req.Name,
		Description:     req.Description,
		Status:          status,
		Priority:        priority,
		ClientName:      req.ClientName,
		ClientID:        clientID,
		ProjectType:     req.ProjectType,
		ProjectMetadata: metadata,
		StartDate:       startDate,
		DueDate:         dueDate,
		Visibility:      req.Visibility,
		OwnerID:         &ownerID,
	})
	if err != nil {
		log.Printf("CreateProject error for user %s: %v", user.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create project"})
		return
	}

	c.JSON(http.StatusCreated, project)
}

// GetProject returns a single project with notes
func (h *Handlers) GetProject(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	queries := sqlc.New(h.pool)
	project, err := queries.GetProject(c.Request.Context(), sqlc.GetProjectParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	// Check if notes are requested
	if c.Query("include_notes") == "true" {
		notes, err := queries.GetProjectNotes(c.Request.Context(), pgtype.UUID{Bytes: id, Valid: true})
		if err == nil {
			c.JSON(http.StatusOK, gin.H{
				"project": project,
				"notes":   notes,
			})
			return
		}
	}

	// Get related artifacts
	if c.Query("include_artifacts") == "true" {
		artifacts, err := queries.ListArtifacts(c.Request.Context(), sqlc.ListArtifactsParams{
			UserID:    user.ID,
			ProjectID: pgtype.UUID{Bytes: id, Valid: true},
		})
		if err == nil {
			c.JSON(http.StatusOK, gin.H{
				"project":   project,
				"artifacts": artifacts,
			})
			return
		}
	}

	c.JSON(http.StatusOK, project)
}

// UpdateProject updates an existing project
func (h *Handlers) UpdateProject(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var req struct {
		Name            *string         `json:"name"`
		Description     *string         `json:"description"`
		Status          *string         `json:"status"`
		Priority        *string         `json:"priority"`
		ClientName      *string         `json:"client_name"`
		ClientID        *string         `json:"client_id"`
		ProjectType     *string         `json:"project_type"`
		ProjectMetadata json.RawMessage `json:"project_metadata"`
		StartDate       *string         `json:"start_date"`
		DueDate         *string         `json:"due_date"`
		Visibility      *string         `json:"visibility"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	queries := sqlc.New(h.pool)

	// Get existing project first
	existing, err := queries.GetProject(c.Request.Context(), sqlc.GetProjectParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	// Build update params with existing values as defaults
	name := existing.Name
	if req.Name != nil {
		name = *req.Name
	}

	description := existing.Description
	if req.Description != nil {
		description = req.Description
	}

	status := existing.Status
	if req.Status != nil {
		status = sqlc.NullProjectstatus{
			Projectstatus: stringToProjectStatus(*req.Status),
			Valid:         true,
		}
	}

	priority := existing.Priority
	if req.Priority != nil {
		priority = sqlc.NullProjectpriority{
			Projectpriority: stringToProjectPriority(*req.Priority),
			Valid:           true,
		}
	}

	clientName := existing.ClientName
	if req.ClientName != nil {
		clientName = req.ClientName
	}

	clientID := existing.ClientID
	if req.ClientID != nil {
		if cid, err := uuid.Parse(*req.ClientID); err == nil {
			clientID = pgtype.UUID{Bytes: cid, Valid: true}
		}
	}

	projectType := existing.ProjectType
	if req.ProjectType != nil {
		projectType = req.ProjectType
	}

	metadata := existing.ProjectMetadata
	if req.ProjectMetadata != nil {
		metadata = req.ProjectMetadata
	}

	startDate := existing.StartDate
	if req.StartDate != nil {
		if t, err := time.Parse("2006-01-02", *req.StartDate); err == nil {
			startDate = pgtype.Date{Time: t, Valid: true}
		}
	}

	dueDate := existing.DueDate
	if req.DueDate != nil {
		if t, err := time.Parse("2006-01-02", *req.DueDate); err == nil {
			dueDate = pgtype.Date{Time: t, Valid: true}
		}
	}

	visibility := existing.Visibility
	if req.Visibility != nil {
		visibility = req.Visibility
	}

	project, err := queries.UpdateProject(c.Request.Context(), sqlc.UpdateProjectParams{
		ID:              pgtype.UUID{Bytes: id, Valid: true},
		Name:            name,
		Description:     description,
		Status:          status,
		Priority:        priority,
		ClientName:      clientName,
		ClientID:        clientID,
		ProjectType:     projectType,
		ProjectMetadata: metadata,
		StartDate:       startDate,
		DueDate:         dueDate,
		Visibility:      visibility,
	})
	if err != nil {
		log.Printf("UpdateProject error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update project"})
		return
	}

	// Trigger notification if status changed
	if h.notificationTriggers != nil && req.Status != nil && existing.Status.Valid {
		oldStatus := string(existing.Status.Projectstatus)
		newStatus := *req.Status
		if oldStatus != newStatus {
			// Get project members to notify
			members, _ := queries.ListProjectMembers(c.Request.Context(), pgtype.UUID{Bytes: id, Valid: true})
			var memberIDs []string
			for _, m := range members {
				if m.TeamMemberID.Valid {
					memberIDs = append(memberIDs, uuid.UUID(m.TeamMemberID.Bytes).String())
				}
			}
			go h.notificationTriggers.OnProjectStatusChanged(c.Request.Context(), services.ProjectStatusChangedInput{
				ProjectID:   id,
				ProjectName: name,
				OldStatus:   oldStatus,
				NewStatus:   newStatus,
				ChangedByID: user.ID,
				ChangedBy:   user.Name,
				MemberIDs:   memberIDs,
			})
		}
	}

	c.JSON(http.StatusOK, project)
}

// DeleteProject deletes a project
func (h *Handlers) DeleteProject(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	queries := sqlc.New(h.pool)
	err = queries.DeleteProject(c.Request.Context(), sqlc.DeleteProjectParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete project"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Project deleted"})
}

// AddProjectNote adds a note to a project
func (h *Handlers) AddProjectNote(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var req struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	queries := sqlc.New(h.pool)

	// Verify project ownership
	_, err = queries.GetProject(c.Request.Context(), sqlc.GetProjectParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	note, err := queries.AddProjectNote(c.Request.Context(), sqlc.AddProjectNoteParams{
		ProjectID: pgtype.UUID{Bytes: id, Valid: true},
		Content:   req.Content,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add note"})
		return
	}

	c.JSON(http.StatusCreated, note)
}

// stringToProjectStatus converts a string to sqlc.Projectstatus
func stringToProjectStatus(s string) sqlc.Projectstatus {
	typeMap := map[string]sqlc.Projectstatus{
		"active":    sqlc.ProjectstatusACTIVE,
		"paused":    sqlc.ProjectstatusPAUSED,
		"completed": sqlc.ProjectstatusCOMPLETED,
		"archived":  sqlc.ProjectstatusARCHIVED,
	}
	if enum, ok := typeMap[strings.ToLower(s)]; ok {
		return enum
	}
	return sqlc.ProjectstatusACTIVE
}

// stringToProjectPriority converts a string to sqlc.Projectpriority
func stringToProjectPriority(p string) sqlc.Projectpriority {
	typeMap := map[string]sqlc.Projectpriority{
		"critical": sqlc.ProjectpriorityCRITICAL,
		"high":     sqlc.ProjectpriorityHIGH,
		"medium":   sqlc.ProjectpriorityMEDIUM,
		"low":      sqlc.ProjectpriorityLOW,
	}
	if enum, ok := typeMap[strings.ToLower(p)]; ok {
		return enum
	}
	return sqlc.ProjectpriorityMEDIUM
}

// TransformProjectRows transforms ListProjectsRow to a clean JSON response
func TransformProjectRows(rows []sqlc.ListProjectsRow) []map[string]interface{} {
	result := make([]map[string]interface{}, len(rows))
	for i, row := range rows {
		result[i] = map[string]interface{}{
			"id":                  projectUUIDToString(row.ID),
			"user_id":             row.UserID,
			"name":                row.Name,
			"description":         row.Description,
			"status":              row.Status.Projectstatus,
			"priority":            row.Priority.Projectpriority,
			"client_name":         row.ClientName,
			"client_id":           projectUUIDToString(row.ClientID),
			"client_company_name": row.ClientCompanyName,
			"project_type":        row.ProjectType,
			"project_metadata":    row.ProjectMetadata,
			"start_date":          dateToString(row.StartDate),
			"due_date":            dateToString(row.DueDate),
			"completed_at":        projectTimestamptzToString(row.CompletedAt),
			"visibility":          row.Visibility,
			"owner_id":            row.OwnerID,
			"created_at":          projectTimestampToString(row.CreatedAt),
			"updated_at":          projectTimestampToString(row.UpdatedAt),
		}
	}
	return result
}

// Helper functions for type conversion (project-specific)
func projectUUIDToString(u pgtype.UUID) *string {
	if !u.Valid {
		return nil
	}
	id := uuid.UUID(u.Bytes).String()
	return &id
}

func dateToString(d pgtype.Date) *string {
	if !d.Valid {
		return nil
	}
	s := d.Time.Format("2006-01-02")
	return &s
}

func projectTimestampToString(t pgtype.Timestamp) *string {
	if !t.Valid {
		return nil
	}
	s := t.Time.Format(time.RFC3339)
	return &s
}

func projectTimestamptzToString(t pgtype.Timestamptz) *string {
	if !t.Valid {
		return nil
	}
	s := t.Time.Format(time.RFC3339)
	return &s
}

// GetProjectStats returns project statistics for the user
func (h *Handlers) GetProjectStats(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	queries := sqlc.New(h.pool)
	stats, err := queries.GetProjectStats(c.Request.Context(), user.ID)
	if err != nil {
		log.Printf("GetProjectStats error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get project stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetOverdueProjects returns overdue projects for the user
func (h *Handlers) GetOverdueProjects(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	queries := sqlc.New(h.pool)
	projects, err := queries.GetOverdueProjects(c.Request.Context(), user.ID)
	if err != nil {
		log.Printf("GetOverdueProjects error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get overdue projects"})
		return
	}

	c.JSON(http.StatusOK, projects)
}

// GetUpcomingProjects returns projects due within 7 days
func (h *Handlers) GetUpcomingProjects(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	queries := sqlc.New(h.pool)
	projects, err := queries.GetUpcomingProjects(c.Request.Context(), user.ID)
	if err != nil {
		log.Printf("GetUpcomingProjects error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get upcoming projects"})
		return
	}

	c.JSON(http.StatusOK, projects)
}
