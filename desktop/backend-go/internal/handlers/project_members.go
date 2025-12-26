package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rhl/businessos-backend/internal/database/sqlc"
	"github.com/rhl/businessos-backend/internal/middleware"
)

// ListProjectMembers returns all members assigned to a project
func (h *Handlers) ListProjectMembers(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	queries := sqlc.New(h.pool)

	// Verify project ownership
	_, err = queries.GetProject(c.Request.Context(), sqlc.GetProjectParams{
		ID:     pgtype.UUID{Bytes: projectID, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	members, err := queries.ListProjectMembers(c.Request.Context(), pgtype.UUID{Bytes: projectID, Valid: true})
	if err != nil {
		log.Printf("ListProjectMembers error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list project members"})
		return
	}

	c.JSON(http.StatusOK, members)
}

// AddProjectMember adds a team member to a project
func (h *Handlers) AddProjectMember(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var req struct {
		TeamMemberID string `json:"team_member_id" binding:"required"`
		Role         string `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	teamMemberID, err := uuid.Parse(req.TeamMemberID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team member ID"})
		return
	}

	queries := sqlc.New(h.pool)

	// Verify project ownership
	_, err = queries.GetProject(c.Request.Context(), sqlc.GetProjectParams{
		ID:     pgtype.UUID{Bytes: projectID, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	// Parse role
	role := sqlc.NullProjectrole{Projectrole: sqlc.ProjectroleMember, Valid: true}
	if req.Role != "" {
		role = sqlc.NullProjectrole{Projectrole: stringToProjectRole(req.Role), Valid: true}
	}

	member, err := queries.AddProjectMember(c.Request.Context(), sqlc.AddProjectMemberParams{
		ProjectID:    pgtype.UUID{Bytes: projectID, Valid: true},
		UserID:       user.ID,
		TeamMemberID: pgtype.UUID{Bytes: teamMemberID, Valid: true},
		Role:         role,
		AssignedBy:   &user.ID,
	})
	if err != nil {
		log.Printf("AddProjectMember error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add project member"})
		return
	}

	c.JSON(http.StatusCreated, member)
}

// UpdateProjectMemberRole updates a member's role in a project
func (h *Handlers) UpdateProjectMemberRole(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	memberID, err := uuid.Parse(c.Param("memberId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid member ID"})
		return
	}

	var req struct {
		Role string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	queries := sqlc.New(h.pool)

	member, err := queries.UpdateProjectMemberRole(c.Request.Context(), sqlc.UpdateProjectMemberRoleParams{
		ID:   pgtype.UUID{Bytes: memberID, Valid: true},
		Role: sqlc.NullProjectrole{Projectrole: stringToProjectRole(req.Role), Valid: true},
	})
	if err != nil {
		log.Printf("UpdateProjectMemberRole error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update member role"})
		return
	}

	c.JSON(http.StatusOK, member)
}

// RemoveProjectMember removes a member from a project
func (h *Handlers) RemoveProjectMember(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	teamMemberID, err := uuid.Parse(c.Param("memberId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid member ID"})
		return
	}

	queries := sqlc.New(h.pool)

	// Verify project ownership
	_, err = queries.GetProject(c.Request.Context(), sqlc.GetProjectParams{
		ID:     pgtype.UUID{Bytes: projectID, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	err = queries.RemoveProjectMemberByTeamMember(c.Request.Context(), sqlc.RemoveProjectMemberByTeamMemberParams{
		ProjectID:    pgtype.UUID{Bytes: projectID, Valid: true},
		TeamMemberID: pgtype.UUID{Bytes: teamMemberID, Valid: true},
	})
	if err != nil {
		log.Printf("RemoveProjectMember error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove project member"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Member removed from project"})
}

// stringToProjectRole converts a string to sqlc.Projectrole
func stringToProjectRole(s string) sqlc.Projectrole {
	switch s {
	case "owner":
		return sqlc.ProjectroleOwner
	case "admin":
		return sqlc.ProjectroleAdmin
	case "viewer":
		return sqlc.ProjectroleViewer
	default:
		return sqlc.ProjectroleMember
	}
}
