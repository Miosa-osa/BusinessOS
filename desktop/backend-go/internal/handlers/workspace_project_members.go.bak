package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rhl/businessos-backend/internal/database/sqlc"
	"github.com/rhl/businessos-backend/internal/middleware"
)

// ============================================================================
// WORKSPACE PROJECT MEMBER HANDLERS
// ============================================================================

// ListWorkspaceProjectMembers returns all members assigned to a project within a workspace
func (h *Handlers) ListWorkspaceProjectMembers(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	projectID, err := uuid.Parse(c.Param("projectId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	queries := sqlc.New(h.pool)
	ctx := c.Request.Context()

	// Check workspace membership
	isMember, err := queries.CheckUserIsWorkspaceMember(ctx, sqlc.CheckUserIsWorkspaceMemberParams{
		WorkspaceID: pgtype.UUID{Bytes: workspaceID, Valid: true},
		UserID:      user.ID,
	})
	if err != nil || !isMember {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not a member of this workspace"})
		return
	}

	members, err := queries.ListWorkspaceProjectMembers(ctx, pgtype.UUID{Bytes: projectID, Valid: true})
	if err != nil {
		log.Printf("ListWorkspaceProjectMembers error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list project members"})
		return
	}

	result := make([]gin.H, len(members))
	for i, m := range members {
		result[i] = gin.H{
			"id":                 m.ID,
			"project_id":         m.ProjectID,
			"user_id":            m.UserID,
			"workspace_id":       m.WorkspaceID,
			"project_role":       m.ProjectRole,
			"assigned_by":        m.AssignedBy,
			"assigned_at":        m.AssignedAt,
			"notification_level": m.NotificationLevel,
			"created_at":         m.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"members": result,
		"total":   len(result),
	})
}

// AddProjectMemberRequest represents the request body for adding a member to a project
type AddProjectMemberRequest struct {
	UserID            string  `json:"user_id" binding:"required"`
	ProjectRole       *string `json:"project_role"`
	NotificationLevel *string `json:"notification_level"`
}

// AddWorkspaceProjectMember adds a workspace member to a project
func (h *Handlers) AddWorkspaceProjectMember(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	projectID, err := uuid.Parse(c.Param("projectId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var req AddProjectMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	queries := sqlc.New(h.pool)
	ctx := c.Request.Context()

	// Check if current user has permission (owner, admin, or manager)
	currentMember, err := queries.GetWorkspaceMember(ctx, sqlc.GetWorkspaceMemberParams{
		WorkspaceID: pgtype.UUID{Bytes: workspaceID, Valid: true},
		UserID:      user.ID,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusForbidden, gin.H{"error": "Not a member of this workspace"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check membership"})
		return
	}

	currentRoleName := ""
	if currentMember.RoleName != nil {
		currentRoleName = *currentMember.RoleName
	}
	if currentRoleName != "owner" && currentRoleName != "admin" && currentRoleName != "manager" {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to add project members"})
		return
	}

	// Check if target user is a workspace member
	isTargetMember, err := queries.CheckUserIsWorkspaceMember(ctx, sqlc.CheckUserIsWorkspaceMemberParams{
		WorkspaceID: pgtype.UUID{Bytes: workspaceID, Valid: true},
		UserID:      req.UserID,
	})
	if err != nil || !isTargetMember {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User is not a member of this workspace"})
		return
	}

	// Check if already a project member
	existingMember, err := queries.GetWorkspaceProjectMember(ctx, sqlc.GetWorkspaceProjectMemberParams{
		ProjectID: pgtype.UUID{Bytes: projectID, Valid: true},
		UserID:    req.UserID,
	})
	if err == nil && existingMember.ID.Valid {
		c.JSON(http.StatusConflict, gin.H{"error": "User is already a member of this project"})
		return
	}

	// Set default role if not provided
	projectRole := "contributor"
	if req.ProjectRole != nil {
		projectRole = *req.ProjectRole
	}

	// Validate project role
	validRoles := map[string]bool{
		"lead":        true,
		"contributor": true,
		"reviewer":    true,
		"viewer":      true,
	}
	if !validRoles[projectRole] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project role. Must be: lead, contributor, reviewer, or viewer"})
		return
	}

	// Set default notification level
	notificationLevel := "all"
	if req.NotificationLevel != nil {
		notificationLevel = *req.NotificationLevel
	}

	// Add project member
	member, err := queries.AddWorkspaceProjectMember(ctx, sqlc.AddWorkspaceProjectMemberParams{
		ProjectID:         pgtype.UUID{Bytes: projectID, Valid: true},
		UserID:            req.UserID,
		WorkspaceID:       pgtype.UUID{Bytes: workspaceID, Valid: true},
		ProjectRole:       &projectRole,
		AssignedBy:        &user.ID,
		NotificationLevel: &notificationLevel,
	})
	if err != nil {
		log.Printf("AddWorkspaceProjectMember error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add project member"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"member": gin.H{
			"id":                 member.ID,
			"project_id":         member.ProjectID,
			"user_id":            member.UserID,
			"workspace_id":       member.WorkspaceID,
			"project_role":       member.ProjectRole,
			"assigned_by":        member.AssignedBy,
			"assigned_at":        member.AssignedAt,
			"notification_level": member.NotificationLevel,
			"created_at":         member.CreatedAt,
		},
	})
}

// BulkAddProjectMembersRequest represents the request for bulk adding members
type BulkAddProjectMembersRequest struct {
	Members []AddProjectMemberRequest `json:"members" binding:"required,min=1"`
}

// BulkAddWorkspaceProjectMembers adds multiple workspace members to a project at once
func (h *Handlers) BulkAddWorkspaceProjectMembers(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	projectID, err := uuid.Parse(c.Param("projectId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var req BulkAddProjectMembersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	queries := sqlc.New(h.pool)
	ctx := c.Request.Context()

	// Check if current user has permission
	currentMember, err := queries.GetWorkspaceMember(ctx, sqlc.GetWorkspaceMemberParams{
		WorkspaceID: pgtype.UUID{Bytes: workspaceID, Valid: true},
		UserID:      user.ID,
	})
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not a member of this workspace"})
		return
	}

	currentRoleName := ""
	if currentMember.RoleName != nil {
		currentRoleName = *currentMember.RoleName
	}
	if currentRoleName != "owner" && currentRoleName != "admin" && currentRoleName != "manager" {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to add project members"})
		return
	}

	added := []gin.H{}
	failed := []gin.H{}

	for _, memberReq := range req.Members {
		// Check if user is workspace member
		isWorkspaceMember, _ := queries.CheckUserIsWorkspaceMember(ctx, sqlc.CheckUserIsWorkspaceMemberParams{
			WorkspaceID: pgtype.UUID{Bytes: workspaceID, Valid: true},
			UserID:      memberReq.UserID,
		})
		if !isWorkspaceMember {
			failed = append(failed, gin.H{"user_id": memberReq.UserID, "error": "Not a workspace member"})
			continue
		}

		// Check if already a project member
		existingMember, _ := queries.GetWorkspaceProjectMember(ctx, sqlc.GetWorkspaceProjectMemberParams{
			ProjectID: pgtype.UUID{Bytes: projectID, Valid: true},
			UserID:    memberReq.UserID,
		})
		if existingMember.ID.Valid {
			failed = append(failed, gin.H{"user_id": memberReq.UserID, "error": "Already a project member"})
			continue
		}

		projectRole := "contributor"
		if memberReq.ProjectRole != nil {
			projectRole = *memberReq.ProjectRole
		}

		notificationLevel := "all"
		if memberReq.NotificationLevel != nil {
			notificationLevel = *memberReq.NotificationLevel
		}

		member, err := queries.AddWorkspaceProjectMember(ctx, sqlc.AddWorkspaceProjectMemberParams{
			ProjectID:         pgtype.UUID{Bytes: projectID, Valid: true},
			UserID:            memberReq.UserID,
			WorkspaceID:       pgtype.UUID{Bytes: workspaceID, Valid: true},
			ProjectRole:       &projectRole,
			AssignedBy:        &user.ID,
			NotificationLevel: &notificationLevel,
		})
		if err != nil {
			failed = append(failed, gin.H{"user_id": memberReq.UserID, "error": "Failed to add"})
			continue
		}

		added = append(added, gin.H{
			"id":           member.ID,
			"user_id":      member.UserID,
			"project_role": member.ProjectRole,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"added":  added,
		"failed": failed,
	})
}

// UpdateProjectMemberRoleRequest represents the request for updating a project member's role
type UpdateProjectMemberRoleRequest struct {
	ProjectRole       *string `json:"project_role"`
	NotificationLevel *string `json:"notification_level"`
}

// UpdateWorkspaceProjectMemberRole updates a member's role in a project
func (h *Handlers) UpdateWorkspaceProjectMemberRole(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	projectID, err := uuid.Parse(c.Param("projectId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	targetUserID := c.Param("userId")
	if targetUserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	var req UpdateProjectMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	queries := sqlc.New(h.pool)
	ctx := c.Request.Context()

	// Check if current user has permission
	currentMember, err := queries.GetWorkspaceMember(ctx, sqlc.GetWorkspaceMemberParams{
		WorkspaceID: pgtype.UUID{Bytes: workspaceID, Valid: true},
		UserID:      user.ID,
	})
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not a member of this workspace"})
		return
	}

	currentRoleName := ""
	if currentMember.RoleName != nil {
		currentRoleName = *currentMember.RoleName
	}
	if currentRoleName != "owner" && currentRoleName != "admin" && currentRoleName != "manager" {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to update project members"})
		return
	}

	// Validate project role if provided
	if req.ProjectRole != nil {
		validRoles := map[string]bool{
			"lead": true, "contributor": true, "reviewer": true, "viewer": true,
		}
		if !validRoles[*req.ProjectRole] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project role"})
			return
		}
	}

	// Update member
	member, err := queries.UpdateWorkspaceProjectMemberRole(ctx, sqlc.UpdateWorkspaceProjectMemberRoleParams{
		ProjectID:         pgtype.UUID{Bytes: projectID, Valid: true},
		UserID:            targetUserID,
		ProjectRole:       req.ProjectRole,
		NotificationLevel: req.NotificationLevel,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project member not found"})
			return
		}
		log.Printf("UpdateWorkspaceProjectMember error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update project member"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"member": gin.H{
			"id":                 member.ID,
			"user_id":            member.UserID,
			"project_role":       member.ProjectRole,
			"notification_level": member.NotificationLevel,
			"updated_at":         member.UpdatedAt,
		},
	})
}

// RemoveWorkspaceProjectMember removes a member from a project
func (h *Handlers) RemoveWorkspaceProjectMember(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	projectID, err := uuid.Parse(c.Param("projectId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	targetUserID := c.Param("userId")
	if targetUserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	queries := sqlc.New(h.pool)
	ctx := c.Request.Context()

	// Check if current user has permission
	currentMember, err := queries.GetWorkspaceMember(ctx, sqlc.GetWorkspaceMemberParams{
		WorkspaceID: pgtype.UUID{Bytes: workspaceID, Valid: true},
		UserID:      user.ID,
	})
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not a member of this workspace"})
		return
	}

	currentRoleName := ""
	if currentMember.RoleName != nil {
		currentRoleName = *currentMember.RoleName
	}
	if currentRoleName != "owner" && currentRoleName != "admin" && currentRoleName != "manager" {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to remove project members"})
		return
	}

	// Remove member
	err = queries.RemoveWorkspaceProjectMember(ctx, sqlc.RemoveWorkspaceProjectMemberParams{
		ProjectID: pgtype.UUID{Bytes: projectID, Valid: true},
		UserID:    targetUserID,
	})
	if err != nil {
		log.Printf("RemoveWorkspaceProjectMember error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove project member"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Member removed from project"})
}

// GetUserWorkspaceProjects returns all projects the current user is assigned to in a workspace
func (h *Handlers) GetUserWorkspaceProjects(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	queries := sqlc.New(h.pool)
	ctx := c.Request.Context()

	// Check workspace membership
	isMember, err := queries.CheckUserIsWorkspaceMember(ctx, sqlc.CheckUserIsWorkspaceMemberParams{
		WorkspaceID: pgtype.UUID{Bytes: workspaceID, Valid: true},
		UserID:      user.ID,
	})
	if err != nil || !isMember {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not a member of this workspace"})
		return
	}

	// Get user's project assignments
	assignments, err := queries.ListUserWorkspaceProjectAssignments(ctx, sqlc.ListUserWorkspaceProjectAssignmentsParams{
		WorkspaceID: pgtype.UUID{Bytes: workspaceID, Valid: true},
		UserID:      user.ID,
	})
	if err != nil {
		log.Printf("ListUserWorkspaceProjectAssignments error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list project assignments"})
		return
	}

	result := make([]gin.H, len(assignments))
	for i, a := range assignments {
		result[i] = gin.H{
			"project_id":         a.ProjectID,
			"project_role":       a.ProjectRole,
			"assigned_at":        a.AssignedAt,
			"notification_level": a.NotificationLevel,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"assignments": result,
		"total":       len(result),
	})
}
