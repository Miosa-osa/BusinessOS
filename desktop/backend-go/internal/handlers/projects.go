package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rhl/businessos-backend/internal/database/sqlc"
	"github.com/rhl/businessos-backend/internal/middleware"
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
	var status sqlc.Projectstatus
	if s := c.Query("status"); s != "" {
		status = stringToProjectStatus(s)
	}

	projects, err := queries.ListProjects(c.Request.Context(), sqlc.ListProjectsParams{
		UserID: user.ID,
		Status: sqlc.NullProjectstatus{Projectstatus: status, Valid: status != ""},
	})
	if err != nil {
		log.Printf("ListProjects error for user %s: %v", user.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list projects"})
		return
	}

	c.JSON(http.StatusOK, TransformProjects(projects))
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
		ProjectType     *string         `json:"project_type"`
		ProjectMetadata json.RawMessage `json:"project_metadata"`
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

	// Handle project_type with default
	projectType := "general"
	if req.ProjectType != nil {
		projectType = *req.ProjectType
	}

	// Handle metadata
	metadata := []byte("{}")
	if req.ProjectMetadata != nil {
		metadata = req.ProjectMetadata
	}

	project, err := queries.CreateProject(c.Request.Context(), sqlc.CreateProjectParams{
		UserID:          user.ID,
		Name:            req.Name,
		Description:     req.Description,
		Status:          status,
		Priority:        priority,
		ClientName:      req.ClientName,
		ProjectType:     &projectType,
		ProjectMetadata: metadata,
	})
	if err != nil {
		log.Printf("CreateProject error for user %s: %v", user.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create project", "details": err.Error()})
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
		ProjectType     *string         `json:"project_type"`
		ProjectMetadata json.RawMessage `json:"project_metadata"`
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

	projectType := existing.ProjectType
	if req.ProjectType != nil {
		projectType = req.ProjectType
	}

	metadata := existing.ProjectMetadata
	if req.ProjectMetadata != nil {
		metadata = req.ProjectMetadata
	}

	project, err := queries.UpdateProject(c.Request.Context(), sqlc.UpdateProjectParams{
		ID:              pgtype.UUID{Bytes: id, Valid: true},
		Name:            name,
		Description:     description,
		Status:          status,
		Priority:        priority,
		ClientName:      clientName,
		ProjectType:     projectType,
		ProjectMetadata: metadata,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update project"})
		return
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
