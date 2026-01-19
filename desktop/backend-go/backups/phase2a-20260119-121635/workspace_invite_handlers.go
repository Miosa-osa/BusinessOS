package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rhl/businessos-backend/internal/middleware"
)

// =====================================================================
// WORKSPACE INVITE HANDLERS
// =====================================================================

// CreateInviteRequest represents the request to create an invitation
type CreateInviteRequest struct {
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role" binding:"required,oneof=owner admin manager member viewer guest"`
}

// CreateWorkspaceInvite creates a new workspace invitation
// POST /api/workspaces/:id/invites
// Required permission: invite_members (manager+)
func (h *Handlers) CreateWorkspaceInvite(c *gin.Context) {
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

	var req CreateInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create invitation
	invite, err := h.inviteService.CreateInvite(
		c.Request.Context(),
		workspaceID,
		req.Email,
		req.Role,
		user.ID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Log the action
	if h.auditService != nil {
		h.auditService.LogAction(
			c.Request.Context(),
			workspaceID,
			user.ID,
			"invite_member",
			"invite",
			nil,
			map[string]interface{}{
				"invite_id": invite.ID.String(),
				"email":     req.Email,
				"role":      req.Role,
			},
			getIPAddress(c),
			getUserAgent(c),
		)
	}

	c.JSON(http.StatusCreated, invite)
}

// ListWorkspaceInvites lists all invitations for a workspace
// GET /api/workspaces/:id/invites
// Required permission: manage_members (admin+)
func (h *Handlers) ListWorkspaceInvites(c *gin.Context) {
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

	invites, err := h.inviteService.ListWorkspaceInvites(c.Request.Context(), workspaceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"invites": invites})
}

// AcceptInviteRequest represents the request to accept an invitation
type AcceptInviteRequest struct {
	Token string `json:"token" binding:"required"`
}

// AcceptWorkspaceInvite accepts a workspace invitation
// POST /api/workspaces/invites/accept
// No workspace permission required (public endpoint for invited users)
func (h *Handlers) AcceptWorkspaceInvite(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var req AcceptInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get invite to log workspace ID
	invite, err := h.inviteService.GetInviteByToken(c.Request.Context(), req.Token)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invalid or expired invitation"})
		return
	}

	// Accept invitation
	err = h.inviteService.AcceptInvite(c.Request.Context(), req.Token, user.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Log the action
	if h.auditService != nil {
		h.auditService.LogAction(
			c.Request.Context(),
			invite.WorkspaceID,
			user.ID,
			"accept_invite",
			"invite",
			nil,
			map[string]interface{}{
				"invite_id": invite.ID.String(),
				"email":     invite.Email,
				"role":      invite.Role,
			},
			getIPAddress(c),
			getUserAgent(c),
		)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Invitation accepted successfully"})
}

// RevokeWorkspaceInvite revokes a pending invitation
// DELETE /api/workspaces/:id/invites/:inviteId
// Required permission: manage_members (admin+)
func (h *Handlers) RevokeWorkspaceInvite(c *gin.Context) {
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

	inviteID, err := uuid.Parse(c.Param("inviteId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid invite ID"})
		return
	}

	err = h.inviteService.RevokeInvite(c.Request.Context(), inviteID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Log the action
	if h.auditService != nil {
		h.auditService.LogAction(
			c.Request.Context(),
			workspaceID,
			user.ID,
			"revoke_invite",
			"invite",
			nil,
			map[string]interface{}{
				"invite_id": inviteID.String(),
			},
			getIPAddress(c),
			getUserAgent(c),
		)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Invitation revoked successfully"})
}

// Helper functions

func getIPAddress(c *gin.Context) *string {
	ip := c.ClientIP()
	return &ip
}

func getUserAgent(c *gin.Context) *string {
	ua := c.GetHeader("User-Agent")
	if ua == "" {
		return nil
	}
	return &ua
}
