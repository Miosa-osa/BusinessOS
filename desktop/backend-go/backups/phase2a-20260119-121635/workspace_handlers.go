package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rhl/businessos-backend/internal/middleware"
	"github.com/rhl/businessos-backend/internal/services"
)

// =====================================================================
// WORKSPACE CRUD HANDLERS
// =====================================================================

// CreateWorkspace creates a new workspace
// POST /api/workspaces
func (h *Handlers) CreateWorkspace(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var req services.CreateWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	workspace, err := h.workspaceService.CreateWorkspace(c.Request.Context(), req, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, workspace)
}

// ListWorkspaces lists all workspaces the user is a member of
// GET /api/workspaces
func (h *Handlers) ListWorkspaces(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	workspaces, err := h.workspaceService.ListUserWorkspaces(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"workspaces": workspaces})
}

// GetWorkspace gets a workspace by ID
// GET /api/workspaces/:id
func (h *Handlers) GetWorkspace(c *gin.Context) {
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

	// Verify user is a member
	_, err = h.workspaceService.GetUserRole(c.Request.Context(), workspaceID, user.ID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not a member of this workspace"})
		return
	}

	workspace, err := h.workspaceService.GetWorkspace(c.Request.Context(), workspaceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, workspace)
}

// UpdateWorkspace updates a workspace
// PUT /api/workspaces/:id
func (h *Handlers) UpdateWorkspace(c *gin.Context) {
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

	// Check if user has permission (admin or owner)
	role, err := h.workspaceService.GetUserRole(c.Request.Context(), workspaceID, user.ID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not a member of this workspace"})
		return
	}
	if role != "owner" && role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only owners and admins can update workspace"})
		return
	}

	var req services.UpdateWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	workspace, err := h.workspaceService.UpdateWorkspace(c.Request.Context(), workspaceID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, workspace)
}

// DeleteWorkspace deletes a workspace
// DELETE /api/workspaces/:id
func (h *Handlers) DeleteWorkspace(c *gin.Context) {
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

	err = h.workspaceService.DeleteWorkspace(c.Request.Context(), workspaceID, user.ID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Workspace deleted successfully"})
}

// =====================================================================
// MEMBER MANAGEMENT HANDLERS
// =====================================================================

// ListWorkspaceMembers lists all members of a workspace
// GET /api/workspaces/:id/members
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

	// Verify user is a member
	_, err = h.workspaceService.GetUserRole(c.Request.Context(), workspaceID, user.ID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not a member of this workspace"})
		return
	}

	members, err := h.workspaceService.ListMembers(c.Request.Context(), workspaceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"members": members})
}

// AddWorkspaceMember adds a member to the workspace
// POST /api/workspaces/:id/members/invite
func (h *Handlers) AddWorkspaceMember(c *gin.Context) {
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

	// Check if user has permission to invite (manager, admin, or owner)
	role, err := h.workspaceService.GetUserRole(c.Request.Context(), workspaceID, user.ID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not a member of this workspace"})
		return
	}
	if role != "owner" && role != "admin" && role != "manager" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only owners, admins, and managers can invite members"})
		return
	}

	var req services.AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	member, err := h.workspaceService.AddMember(c.Request.Context(), workspaceID, req, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, member)
}

// UpdateWorkspaceMemberRole updates a member's role
// PUT /api/workspaces/:id/members/:userId
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

	// Check if user has permission (admin or owner)
	role, err := h.workspaceService.GetUserRole(c.Request.Context(), workspaceID, user.ID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not a member of this workspace"})
		return
	}
	if role != "owner" && role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only owners and admins can update member roles"})
		return
	}

	var req struct {
		Role string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	member, err := h.workspaceService.UpdateMemberRole(c.Request.Context(), workspaceID, targetUserID, req.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, member)
}

// RemoveWorkspaceMember removes a member from the workspace
// DELETE /api/workspaces/:id/members/:userId
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

	// Check if user has permission (admin or owner)
	role, err := h.workspaceService.GetUserRole(c.Request.Context(), workspaceID, user.ID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not a member of this workspace"})
		return
	}
	if role != "owner" && role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only owners and admins can remove members"})
		return
	}

	err = h.workspaceService.RemoveMember(c.Request.Context(), workspaceID, targetUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Member removed successfully"})
}

// =====================================================================
// ROLE MANAGEMENT HANDLERS
// =====================================================================

// ListWorkspaceRoles lists all roles in a workspace
// GET /api/workspaces/:id/roles
func (h *Handlers) ListWorkspaceRoles(c *gin.Context) {
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

	// Verify user is a member
	_, err = h.workspaceService.GetUserRole(c.Request.Context(), workspaceID, user.ID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not a member of this workspace"})
		return
	}

	roles, err := h.workspaceService.ListRoles(c.Request.Context(), workspaceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"roles": roles})
}

// =====================================================================
// USER PROFILE & ROLE CONTEXT HANDLERS
// =====================================================================

// GetWorkspaceProfile gets the current user's profile in the workspace
// GET /api/workspaces/:id/profile
func (h *Handlers) GetWorkspaceProfile(c *gin.Context) {
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

	// Get user's membership in this workspace
	members, err := h.workspaceService.ListMembers(c.Request.Context(), workspaceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Find current user's membership
	var userMember *services.WorkspaceMember
	for i := range members {
		if members[i].UserID == user.ID {
			userMember = &members[i]
			break
		}
	}

	if userMember == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not a member of this workspace"})
		return
	}

	// Return profile (using member info)
	profile := gin.H{
		"user_id":      userMember.UserID,
		"workspace_id": userMember.WorkspaceID.String(),
		"role":         userMember.Role,
		"status":       userMember.Status,
		"joined_at":    userMember.JoinedAt,
		"created_at":   userMember.CreatedAt,
	}

	c.JSON(http.StatusOK, profile)
}

// UpdateWorkspaceProfile updates the current user's profile in the workspace
// PUT /api/workspaces/:id/profile
func (h *Handlers) UpdateWorkspaceProfile(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	_, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	// For now, profile updates are not implemented
	// This would typically update user_workspace_profiles table
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Profile updates not yet implemented"})
}

// GetUserRoleContext gets the current user's role context (permissions) in the workspace
// GET /api/workspaces/:id/role-context
func (h *Handlers) GetUserRoleContext(c *gin.Context) {
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

	// Get role context using roleContextService if available
	if h.roleContextService != nil {
		roleCtx, err := h.roleContextService.GetUserRoleContext(c.Request.Context(), user.ID, workspaceID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, roleCtx)
		return
	}

	// Fallback: Build role context manually from workspace service
	// Get user's role name
	roleName, err := h.workspaceService.GetUserRole(c.Request.Context(), workspaceID, user.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not a member of this workspace"})
		return
	}

	workspace, err := h.workspaceService.GetWorkspace(c.Request.Context(), workspaceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get all roles and find the user's role
	roles, err := h.workspaceService.ListRoles(c.Request.Context(), workspaceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var userRole *services.WorkspaceRole
	for i := range roles {
		if roles[i].Name == roleName {
			userRole = &roles[i]
			break
		}
	}

	if userRole == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
		return
	}

	// Build role context response
	roleContext := gin.H{
		"user_id":           user.ID,
		"workspace_id":      workspaceID.String(),
		"workspace_name":    workspace.Name,
		"role_name":         userRole.Name,
		"role_display_name": userRole.DisplayName,
		"hierarchy_level":   userRole.HierarchyLevel,
		"permissions":       userRole.Permissions,
	}

	c.JSON(http.StatusOK, roleContext)
}
