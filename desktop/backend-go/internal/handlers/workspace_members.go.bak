package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rhl/businessos-backend/internal/database/sqlc"
	"github.com/rhl/businessos-backend/internal/middleware"
	"github.com/rhl/businessos-backend/internal/services"
)

// ============================================================================
// WORKSPACE MEMBER HANDLERS
// ============================================================================

// ListWorkspaceMembers returns all members of a workspace
func (h *Handlers) ListWorkspaceMembers(c *gin.Context) {
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

	// Check membership
	isMember, err := queries.CheckUserIsWorkspaceMember(ctx, sqlc.CheckUserIsWorkspaceMemberParams{
		WorkspaceID: pgtype.UUID{Bytes: workspaceID, Valid: true},
		UserID:      user.ID,
	})
	if err != nil || !isMember {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not a member of this workspace"})
		return
	}

	// Get optional status filter
	status := c.Query("status")

	var members []sqlc.ListWorkspaceMembersRow
	if status != "" {
		membersByStatus, err := queries.ListWorkspaceMembersByStatus(ctx, sqlc.ListWorkspaceMembersByStatusParams{
			WorkspaceID: pgtype.UUID{Bytes: workspaceID, Valid: true},
			Status:      &status,
		})
		if err != nil {
			log.Printf("ListWorkspaceMembersByStatus error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list members"})
			return
		}
		// Convert to ListWorkspaceMembersRow type
		for _, m := range membersByStatus {
			members = append(members, sqlc.ListWorkspaceMembersRow{
				ID:              m.ID,
				WorkspaceID:     m.WorkspaceID,
				UserID:          m.UserID,
				RoleID:          m.RoleID,
				RoleName:        m.RoleName,
				Status:          m.Status,
				InvitedBy:       m.InvitedBy,
				InvitedAt:       m.InvitedAt,
				JoinedAt:        m.JoinedAt,
				CreatedAt:       m.CreatedAt,
				UpdatedAt:       m.UpdatedAt,
				RoleDisplayName: m.RoleDisplayName,
				RoleColor:       m.RoleColor,
				HierarchyLevel:  m.HierarchyLevel,
			})
		}
	} else {
		members, err = queries.ListWorkspaceMembers(ctx, pgtype.UUID{Bytes: workspaceID, Valid: true})
		if err != nil {
			log.Printf("ListWorkspaceMembers error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list members"})
			return
		}
	}

	result := make([]gin.H, len(members))
	for i, m := range members {
		result[i] = gin.H{
			"id":      m.ID,
			"user_id": m.UserID,
			"role": gin.H{
				"id":           m.RoleID,
				"name":         m.RoleName,
				"display_name": m.RoleDisplayName,
				"color":        m.RoleColor,
			},
			"status":     m.Status,
			"invited_by": m.InvitedBy,
			"invited_at": m.InvitedAt,
			"joined_at":  m.JoinedAt,
			"created_at": m.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"members": result,
		"total":   len(result),
	})
}

// UpdateMemberRoleRequest represents the request body for updating a member's role
type UpdateMemberRoleRequest struct {
	RoleID   *string `json:"role_id"`
	RoleName *string `json:"role_name"`
}

// UpdateWorkspaceMemberRole updates a member's role in a workspace
func (h *Handlers) UpdateWorkspaceMemberRole(c *gin.Context) {
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

	targetUserID := c.Param("userId")
	if targetUserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	var req UpdateMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.RoleID == nil && req.RoleName == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Either role_id or role_name is required"})
		return
	}

	queries := sqlc.New(h.pool)
	ctx := c.Request.Context()

	// Check if current user has permission to manage roles
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

	// Check permission - need team.manage_roles or be owner/admin
	currentRoleName := ""
	if currentMember.RoleName != nil {
		currentRoleName = *currentMember.RoleName
	}
	if currentRoleName != "owner" && currentRoleName != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to manage roles"})
		return
	}

	// Check if target user exists in workspace
	targetMember, err := queries.GetWorkspaceMember(ctx, sqlc.GetWorkspaceMemberParams{
		WorkspaceID: pgtype.UUID{Bytes: workspaceID, Valid: true},
		UserID:      targetUserID,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "User is not a member of this workspace"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get member"})
		return
	}

	// Cannot change owner's role
	targetRoleName := ""
	if targetMember.RoleName != nil {
		targetRoleName = *targetMember.RoleName
	}
	if targetRoleName == "owner" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot change the owner's role"})
		return
	}

	// Get the new role
	var newRole sqlc.WorkspaceRole
	if req.RoleID != nil {
		roleID, err := uuid.Parse(*req.RoleID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
			return
		}
		newRole, err = queries.GetWorkspaceRole(ctx, sqlc.GetWorkspaceRoleParams{
			ID:          pgtype.UUID{Bytes: roleID, Valid: true},
			WorkspaceID: pgtype.UUID{Bytes: workspaceID, Valid: true},
		})
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
			return
		}
	} else {
		newRole, err = queries.GetWorkspaceRoleByName(ctx, sqlc.GetWorkspaceRoleByNameParams{
			WorkspaceID: pgtype.UUID{Bytes: workspaceID, Valid: true},
			Name:        *req.RoleName,
		})
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
			return
		}
	}

	// Cannot assign owner role
	if newRole.Name == "owner" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot assign owner role"})
		return
	}

	// Cannot assign role higher than your own (unless you're owner)
	if currentRoleName != "owner" && newRole.HierarchyLevel != nil && currentMember.HierarchyLevel != nil {
		if *newRole.HierarchyLevel > *currentMember.HierarchyLevel {
			c.JSON(http.StatusForbidden, gin.H{"error": "Cannot assign a role higher than your own"})
			return
		}
	}

	// Update the member's role
	updatedMember, err := queries.UpdateWorkspaceMemberRole(ctx, sqlc.UpdateWorkspaceMemberRoleParams{
		WorkspaceID: pgtype.UUID{Bytes: workspaceID, Valid: true},
		UserID:      targetUserID,
		RoleID:      newRole.ID,
		RoleName:    &newRole.Name,
	})
	if err != nil {
		log.Printf("UpdateWorkspaceMemberRole error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update member role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"member": gin.H{
			"user_id":   updatedMember.UserID,
			"role_name": updatedMember.RoleName,
			"role_id":   updatedMember.RoleID,
		},
		"message": "Role updated successfully",
	})
}

// RemoveWorkspaceMember removes a member from a workspace
func (h *Handlers) RemoveWorkspaceMember(c *gin.Context) {
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

	targetUserID := c.Param("userId")
	if targetUserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	queries := sqlc.New(h.pool)
	ctx := c.Request.Context()

	// Self-removal is always allowed
	isSelfRemoval := targetUserID == user.ID

	if !isSelfRemoval {
		// Check if current user has permission to remove members
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
		if currentRoleName != "owner" && currentRoleName != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to remove members"})
			return
		}
	}

	// Check if target is owner
	isOwner, err := queries.CheckUserIsWorkspaceOwner(ctx, sqlc.CheckUserIsWorkspaceOwnerParams{
		ID:      pgtype.UUID{Bytes: workspaceID, Valid: true},
		OwnerID: targetUserID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check ownership"})
		return
	}
	if isOwner {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot remove the workspace owner"})
		return
	}

	// Remove the member
	err = queries.DeleteWorkspaceMember(ctx, sqlc.DeleteWorkspaceMemberParams{
		WorkspaceID: pgtype.UUID{Bytes: workspaceID, Valid: true},
		UserID:      targetUserID,
	})
	if err != nil {
		log.Printf("DeleteWorkspaceMember error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove member"})
		return
	}

	// Also remove from all project members in this workspace
	_, err = h.pool.Exec(ctx, "DELETE FROM workspace_project_members WHERE workspace_id = $1 AND user_id = $2", workspaceID, targetUserID)
	if err != nil {
		log.Printf("Delete project members error: %v", err)
		// Non-fatal, continue
	}

	c.JSON(http.StatusOK, gin.H{"message": "Member removed successfully"})
}

// ============================================================================
// WORKSPACE INVITATION HANDLERS
// ============================================================================

// InviteMemberRequest represents the request body for inviting a member
type InviteMemberRequest struct {
	Email    string  `json:"email" binding:"required,email"`
	RoleID   *string `json:"role_id"`
	RoleName *string `json:"role_name"`
}

// InviteWorkspaceMember sends a magic link invitation to join the workspace
func (h *Handlers) InviteWorkspaceMember(c *gin.Context) {
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

	var req InviteMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	queries := sqlc.New(h.pool)
	ctx := c.Request.Context()

	// Check if current user has permission to invite
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
	// Check team.invite permission - owner, admin, and manager can invite
	if currentRoleName != "owner" && currentRoleName != "admin" && currentRoleName != "manager" {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to invite members"})
		return
	}

	// Check workspace member limit
	workspace, err := queries.GetWorkspaceByID(ctx, pgtype.UUID{Bytes: workspaceID, Valid: true})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Workspace not found"})
		return
	}

	memberCount, _ := queries.CountWorkspaceMembers(ctx, pgtype.UUID{Bytes: workspaceID, Valid: true})
	if workspace.MaxMembers != nil && memberCount >= int64(*workspace.MaxMembers) {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Workspace member limit reached",
			"message": "Upgrade your plan to add more members",
		})
		return
	}

	// Check if already a member (by email in user table)
	// For now, we'll skip this check since we don't have access to the user table

	// Check if pending invitation already exists
	pendingExists, err := queries.CheckPendingInvitationExists(ctx, sqlc.CheckPendingInvitationExistsParams{
		WorkspaceID: pgtype.UUID{Bytes: workspaceID, Valid: true},
		Email:       req.Email,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check existing invitations"})
		return
	}
	if pendingExists {
		c.JSON(http.StatusConflict, gin.H{
			"error":   "pending_invitation",
			"message": "A pending invitation already exists for this email",
		})
		return
	}

	// Get the role to assign
	var role sqlc.WorkspaceRole
	if req.RoleID != nil {
		roleID, err := uuid.Parse(*req.RoleID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
			return
		}
		role, err = queries.GetWorkspaceRole(ctx, sqlc.GetWorkspaceRoleParams{
			ID:          pgtype.UUID{Bytes: roleID, Valid: true},
			WorkspaceID: pgtype.UUID{Bytes: workspaceID, Valid: true},
		})
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
			return
		}
	} else if req.RoleName != nil {
		role, err = queries.GetWorkspaceRoleByName(ctx, sqlc.GetWorkspaceRoleByNameParams{
			WorkspaceID: pgtype.UUID{Bytes: workspaceID, Valid: true},
			Name:        *req.RoleName,
		})
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
			return
		}
	} else {
		// Use default role
		role, err = queries.GetDefaultWorkspaceRole(ctx, pgtype.UUID{Bytes: workspaceID, Valid: true})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get default role"})
			return
		}
	}

	// Cannot invite as owner
	if role.Name == "owner" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot invite as owner"})
		return
	}

	// Generate secure token
	token, err := generateSecureToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate invitation token"})
		return
	}

	// Set expiration (7 days)
	expiresAt := time.Now().AddDate(0, 0, 7)

	// Get inviter name (use email as fallback)
	inviterName := user.Email

	// Create invitation
	invitation, err := queries.CreateWorkspaceInvitation(ctx, sqlc.CreateWorkspaceInvitationParams{
		WorkspaceID:   pgtype.UUID{Bytes: workspaceID, Valid: true},
		Email:         req.Email,
		Token:         token,
		RoleID:        role.ID,
		RoleName:      role.Name,
		InvitedByID:   user.ID,
		InvitedByName: &inviterName,
		ExpiresAt:     pgtype.Timestamptz{Time: expiresAt, Valid: true},
	})
	if err != nil {
		log.Printf("CreateWorkspaceInvitation error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create invitation"})
		return
	}

	// Send invitation email if email service is configured
	if h.emailService != nil && h.emailService.IsEnabled() {
		// Get workspace logo URL if available
		workspaceLogo := ""
		if workspace.LogoUrl != nil {
			workspaceLogo = *workspace.LogoUrl
		}

		emailData := services.InvitationEmailData{
			RecipientEmail: req.Email,
			WorkspaceName:  workspace.Name,
			WorkspaceLogo:  workspaceLogo,
			InviterName:    inviterName,
			RoleName:       role.Name,
			Token:          token,
			ExpiresIn:      "7 days",
		}

		if err := h.emailService.SendInvitationEmail(ctx, emailData); err != nil {
			// Log but don't fail - invitation is created
			log.Printf("Failed to send invitation email to %s: %v", req.Email, err)
		} else {
			log.Printf("Invitation email sent to %s for workspace %s", req.Email, workspace.Name)
		}
	} else {
		log.Printf("Email service not configured - invitation created for %s with token: %s", req.Email, token)
	}

	c.JSON(http.StatusCreated, gin.H{
		"invitation": gin.H{
			"id":           invitation.ID,
			"email":        invitation.Email,
			"role_name":    invitation.RoleName,
			"status":       invitation.Status,
			"expires_at":   invitation.ExpiresAt,
			"invited_by":   inviterName,
			"workspace_id": invitation.WorkspaceID,
		},
		"message": "Invitation sent successfully",
	})
}

// ListWorkspaceInvitations returns all invitations for a workspace
func (h *Handlers) ListWorkspaceInvitations(c *gin.Context) {
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

	// Check membership
	isMember, err := queries.CheckUserIsWorkspaceMember(ctx, sqlc.CheckUserIsWorkspaceMemberParams{
		WorkspaceID: pgtype.UUID{Bytes: workspaceID, Valid: true},
		UserID:      user.ID,
	})
	if err != nil || !isMember {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not a member of this workspace"})
		return
	}

	status := c.Query("status")

	var invitations []sqlc.WorkspaceInvitation
	if status != "" {
		invitations, err = queries.ListWorkspaceInvitationsByStatus(ctx, sqlc.ListWorkspaceInvitationsByStatusParams{
			WorkspaceID: pgtype.UUID{Bytes: workspaceID, Valid: true},
			Status:      status,
		})
	} else {
		invitations, err = queries.ListWorkspaceInvitations(ctx, pgtype.UUID{Bytes: workspaceID, Valid: true})
	}
	if err != nil {
		log.Printf("ListWorkspaceInvitations error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list invitations"})
		return
	}

	result := make([]gin.H, len(invitations))
	for i, inv := range invitations {
		result[i] = gin.H{
			"id":             inv.ID,
			"email":          inv.Email,
			"role_name":      inv.RoleName,
			"status":         inv.Status,
			"invited_by_id":  inv.InvitedByID,
			"invited_by_name": inv.InvitedByName,
			"expires_at":     inv.ExpiresAt,
			"accepted_at":    inv.AcceptedAt,
			"created_at":     inv.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"invitations": result,
		"total":       len(result),
	})
}

// RevokeWorkspaceInvitation cancels a pending invitation
func (h *Handlers) RevokeWorkspaceInvitation(c *gin.Context) {
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

	invitationID, err := uuid.Parse(c.Param("invitationId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid invitation ID"})
		return
	}

	queries := sqlc.New(h.pool)
	ctx := c.Request.Context()

	// Check permission
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
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to revoke invitations"})
		return
	}

	err = queries.RevokeInvitation(ctx, sqlc.RevokeInvitationParams{
		ID:          pgtype.UUID{Bytes: invitationID, Valid: true},
		WorkspaceID: pgtype.UUID{Bytes: workspaceID, Valid: true},
	})
	if err != nil {
		log.Printf("RevokeInvitation error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke invitation"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Invitation revoked successfully"})
}

// ResendInvitation generates a new token and resends the invitation email
func (h *Handlers) ResendInvitation(c *gin.Context) {
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

	invitationID, err := uuid.Parse(c.Param("invitationId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid invitation ID"})
		return
	}

	queries := sqlc.New(h.pool)
	ctx := c.Request.Context()

	// Check permission
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
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to resend invitations"})
		return
	}

	// Get the invitation to verify it exists and is pending
	invitation, err := queries.GetWorkspaceInvitation(ctx, sqlc.GetWorkspaceInvitationParams{
		ID:          pgtype.UUID{Bytes: invitationID, Valid: true},
		WorkspaceID: pgtype.UUID{Bytes: workspaceID, Valid: true},
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invitation not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get invitation"})
		return
	}

	if invitation.Status != "pending" {
		c.JSON(http.StatusConflict, gin.H{
			"error":   "already_processed",
			"message": "This invitation has already been " + invitation.Status,
		})
		return
	}

	// Generate new token
	newToken, err := generateSecureToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate new token"})
		return
	}

	// Set new expiration (7 days from now)
	newExpiresAt := time.Now().AddDate(0, 0, 7)

	// Update invitation with new token and expiration
	updatedInvitation, err := queries.UpdateInvitationToken(ctx, sqlc.UpdateInvitationTokenParams{
		ID:        pgtype.UUID{Bytes: invitationID, Valid: true},
		Token:     newToken,
		ExpiresAt: pgtype.Timestamptz{Time: newExpiresAt, Valid: true},
	})
	if err != nil {
		log.Printf("ResendInvitation update error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update invitation"})
		return
	}

	// Get workspace info for email
	workspace, err := queries.GetWorkspaceByID(ctx, pgtype.UUID{Bytes: workspaceID, Valid: true})
	if err != nil {
		log.Printf("GetWorkspace error: %v", err)
		// Continue anyway - we can still send the email
	}

	// Send new invitation email
	if h.emailService != nil && h.emailService.IsEnabled() {
		workspaceLogo := ""
		workspaceName := "Workspace"
		if workspace.LogoUrl != nil {
			workspaceLogo = *workspace.LogoUrl
		}
		workspaceName = workspace.Name

		inviterName := user.Email
		if updatedInvitation.InvitedByName != nil {
			inviterName = *updatedInvitation.InvitedByName
		}

		emailData := services.InvitationEmailData{
			RecipientEmail: updatedInvitation.Email,
			WorkspaceName:  workspaceName,
			WorkspaceLogo:  workspaceLogo,
			InviterName:    inviterName,
			RoleName:       updatedInvitation.RoleName,
			Token:          newToken,
			ExpiresIn:      "7 days",
		}

		if err := h.emailService.SendInvitationEmail(ctx, emailData); err != nil {
			log.Printf("Failed to resend invitation email to %s: %v", updatedInvitation.Email, err)
			// Don't fail - the token was updated
		} else {
			log.Printf("Invitation email resent to %s", updatedInvitation.Email)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Invitation resent successfully",
		"expires_at": newExpiresAt,
	})
}

// VerifyInvitation checks if a magic link token is valid (public endpoint)
func (h *Handlers) VerifyInvitation(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token is required"})
		return
	}

	queries := sqlc.New(h.pool)
	ctx := c.Request.Context()

	invitation, err := queries.GetWorkspaceInvitationByToken(ctx, token)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"valid":   false,
				"error":   "invitation_not_found",
				"message": "This invitation link is invalid or has been revoked.",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify invitation"})
		return
	}

	// Check status
	switch invitation.Status {
	case "accepted":
		c.JSON(http.StatusGone, gin.H{
			"valid":   false,
			"error":   "invitation_used",
			"message": "This invitation has already been used.",
		})
		return
	case "revoked":
		c.JSON(http.StatusGone, gin.H{
			"valid":   false,
			"error":   "invitation_revoked",
			"message": "This invitation has been revoked.",
		})
		return
	case "expired":
		c.JSON(http.StatusGone, gin.H{
			"valid":   false,
			"error":   "invitation_expired",
			"message": "This invitation has expired. Please request a new one.",
		})
		return
	}

	// Check expiration
	if time.Now().After(invitation.ExpiresAt.Time) {
		// Update status to expired
		queries.UpdateInvitationStatus(ctx, sqlc.UpdateInvitationStatusParams{
			ID:     invitation.ID,
			Status: "expired",
		})
		c.JSON(http.StatusGone, gin.H{
			"valid":   false,
			"error":   "invitation_expired",
			"message": "This invitation has expired. Please request a new one.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":       true,
		"email":       invitation.Email,
		"role_name":   invitation.RoleName,
		"invited_by":  invitation.InvitedByName,
		"expires_at":  invitation.ExpiresAt,
		"workspace": gin.H{
			"id":       invitation.WorkspaceID,
			"name":     invitation.WorkspaceName,
			"slug":     invitation.WorkspaceSlug,
			"logo_url": invitation.WorkspaceLogo,
		},
	})
}

// AcceptInvitation accepts a magic link invitation
func (h *Handlers) AcceptInvitation(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token is required"})
		return
	}

	queries := sqlc.New(h.pool)
	ctx := c.Request.Context()

	// Start transaction
	tx, err := h.pool.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}
	defer tx.Rollback(ctx)

	qtx := queries.WithTx(tx)

	invitation, err := qtx.GetWorkspaceInvitationByToken(ctx, token)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invitation not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get invitation"})
		return
	}

	// Verify status
	if invitation.Status != "pending" {
		c.JSON(http.StatusGone, gin.H{
			"error":   "invitation_" + invitation.Status,
			"message": "This invitation is no longer valid",
		})
		return
	}

	// Check expiration
	if time.Now().After(invitation.ExpiresAt.Time) {
		c.JSON(http.StatusGone, gin.H{
			"error":   "invitation_expired",
			"message": "This invitation has expired",
		})
		return
	}

	// Add user to workspace
	now := time.Now()
	_, err = qtx.CreateWorkspaceMember(ctx, sqlc.CreateWorkspaceMemberParams{
		WorkspaceID: invitation.WorkspaceID,
		UserID:      user.ID,
		RoleID:      invitation.RoleID,
		RoleName:    &invitation.RoleName,
		Status:      stringPtr("active"),
		InvitedBy:   &invitation.InvitedByID,
		InvitedAt:   invitation.CreatedAt,
		JoinedAt:    pgtype.Timestamptz{Time: now, Valid: true},
	})
	if err != nil {
		log.Printf("CreateWorkspaceMember error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add member"})
		return
	}

	// Mark invitation as accepted
	err = qtx.AcceptInvitation(ctx, sqlc.AcceptInvitationParams{
		ID:               invitation.ID,
		AcceptedByUserID: &user.ID,
	})
	if err != nil {
		log.Printf("AcceptInvitation error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to accept invitation"})
		return
	}

	// Commit transaction
	if err = tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully joined workspace",
		"workspace": gin.H{
			"id":   invitation.WorkspaceID,
			"name": invitation.WorkspaceName,
			"slug": invitation.WorkspaceSlug,
		},
		"role": invitation.RoleName,
	})
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// generateSecureToken generates a cryptographically secure random token
func generateSecureToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
