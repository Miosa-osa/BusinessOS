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
// WORKSPACE ROLE HANDLERS
// ============================================================================

// ListWorkspaceRoles returns all roles for a workspace
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

	roles, err := queries.ListWorkspaceRoles(ctx, pgtype.UUID{Bytes: workspaceID, Valid: true})
	if err != nil {
		log.Printf("ListWorkspaceRoles error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list roles"})
		return
	}

	result := make([]gin.H, len(roles))
	for i, r := range roles {
		result[i] = gin.H{
			"id":              r.ID,
			"name":            r.Name,
			"display_name":    r.DisplayName,
			"description":     r.Description,
			"color":           r.Color,
			"icon":            r.Icon,
			"permissions":     r.Permissions,
			"is_default":      r.IsDefault,
			"is_system":       r.IsSystem,
			"hierarchy_level": r.HierarchyLevel,
			"member_count":    r.MemberCount,
			"created_at":      r.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"roles": result,
		"total": len(result),
	})
}

// GetWorkspaceRole returns a specific role
func (h *Handlers) GetWorkspaceRole(c *gin.Context) {
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

	roleID, err := uuid.Parse(c.Param("roleId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
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

	role, err := queries.GetWorkspaceRole(ctx, sqlc.GetWorkspaceRoleParams{
		ID:          pgtype.UUID{Bytes: roleID, Valid: true},
		WorkspaceID: pgtype.UUID{Bytes: workspaceID, Valid: true},
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"role": gin.H{
			"id":              role.ID,
			"name":            role.Name,
			"display_name":    role.DisplayName,
			"description":     role.Description,
			"color":           role.Color,
			"icon":            role.Icon,
			"permissions":     role.Permissions,
			"is_default":      role.IsDefault,
			"is_system":       role.IsSystem,
			"hierarchy_level": role.HierarchyLevel,
			"created_at":      role.CreatedAt,
		},
	})
}

// CreateWorkspaceRoleRequest represents the request body for creating a role
type CreateWorkspaceRoleRequest struct {
	Name           string  `json:"name" binding:"required,min=2,max=50"`
	DisplayName    *string `json:"display_name"`
	Description    *string `json:"description"`
	Color          *string `json:"color"`
	Icon           *string `json:"icon"`
	Permissions    []byte  `json:"permissions"`
	HierarchyLevel *int32  `json:"hierarchy_level"`
	IsDefault      *bool   `json:"is_default"`
}

// CreateWorkspaceRole creates a new custom role for a workspace
func (h *Handlers) CreateWorkspaceRole(c *gin.Context) {
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

	var req CreateWorkspaceRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

	currentRoleName := ""
	if currentMember.RoleName != nil {
		currentRoleName = *currentMember.RoleName
	}
	// Only owner and admin can create roles
	if currentRoleName != "owner" && currentRoleName != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to create roles"})
		return
	}

	// Check for reserved role names
	reservedNames := map[string]bool{
		"owner":   true,
		"admin":   true,
		"manager": true,
		"member":  true,
		"viewer":  true,
	}
	if reservedNames[req.Name] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot create a role with a reserved name"})
		return
	}

	// Check if role name already exists
	_, err = queries.GetWorkspaceRoleByName(ctx, sqlc.GetWorkspaceRoleByNameParams{
		WorkspaceID: pgtype.UUID{Bytes: workspaceID, Valid: true},
		Name:        req.Name,
	})
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "A role with this name already exists"})
		return
	}
	if err != pgx.ErrNoRows {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check role name"})
		return
	}

	// Set defaults
	displayName := req.Name
	if req.DisplayName != nil {
		displayName = *req.DisplayName
	}

	var hierarchyLevel *int32
	if req.HierarchyLevel != nil {
		hierarchyLevel = req.HierarchyLevel
	} else {
		defaultLevel := int32(10)
		hierarchyLevel = &defaultLevel
	}

	isDefault := false
	if req.IsDefault != nil {
		isDefault = *req.IsDefault
	}

	isSystem := false

	// Create role
	role, err := queries.CreateWorkspaceRole(ctx, sqlc.CreateWorkspaceRoleParams{
		WorkspaceID:    pgtype.UUID{Bytes: workspaceID, Valid: true},
		Name:           req.Name,
		DisplayName:    displayName,
		Description:    req.Description,
		Color:          req.Color,
		Icon:           req.Icon,
		Permissions:    req.Permissions,
		IsDefault:      &isDefault,
		IsSystem:       &isSystem,
		HierarchyLevel: hierarchyLevel,
	})
	if err != nil {
		log.Printf("CreateWorkspaceRole error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create role"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"role": gin.H{
			"id":              role.ID,
			"name":            role.Name,
			"display_name":    role.DisplayName,
			"description":     role.Description,
			"color":           role.Color,
			"icon":            role.Icon,
			"permissions":     role.Permissions,
			"is_default":      role.IsDefault,
			"is_system":       role.IsSystem,
			"hierarchy_level": role.HierarchyLevel,
			"created_at":      role.CreatedAt,
		},
	})
}

// UpdateWorkspaceRoleRequest represents the request body for updating a role
type UpdateWorkspaceRoleRequest struct {
	DisplayName    *string `json:"display_name"`
	Description    *string `json:"description"`
	Color          *string `json:"color"`
	Icon           *string `json:"icon"`
	Permissions    []byte  `json:"permissions"`
	HierarchyLevel *int32  `json:"hierarchy_level"`
	IsDefault      *bool   `json:"is_default"`
}

// UpdateWorkspaceRole updates a custom role
func (h *Handlers) UpdateWorkspaceRole(c *gin.Context) {
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

	roleID, err := uuid.Parse(c.Param("roleId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
		return
	}

	var req UpdateWorkspaceRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
	if currentRoleName != "owner" && currentRoleName != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to update roles"})
		return
	}

	// Get existing role to check if system role
	existingRole, err := queries.GetWorkspaceRole(ctx, sqlc.GetWorkspaceRoleParams{
		ID:          pgtype.UUID{Bytes: roleID, Valid: true},
		WorkspaceID: pgtype.UUID{Bytes: workspaceID, Valid: true},
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get role"})
		return
	}

	// Cannot edit system roles (query already prevents this, but give better error)
	if existingRole.IsSystem != nil && *existingRole.IsSystem {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot modify system roles"})
		return
	}

	// Update role
	role, err := queries.UpdateWorkspaceRole(ctx, sqlc.UpdateWorkspaceRoleParams{
		ID:             pgtype.UUID{Bytes: roleID, Valid: true},
		WorkspaceID:    pgtype.UUID{Bytes: workspaceID, Valid: true},
		DisplayName:    req.DisplayName,
		Description:    req.Description,
		Color:          req.Color,
		Icon:           req.Icon,
		Permissions:    req.Permissions,
		HierarchyLevel: req.HierarchyLevel,
		IsDefault:      req.IsDefault,
	})
	if err != nil {
		log.Printf("UpdateWorkspaceRole error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"role": gin.H{
			"id":              role.ID,
			"name":            role.Name,
			"display_name":    role.DisplayName,
			"description":     role.Description,
			"color":           role.Color,
			"icon":            role.Icon,
			"permissions":     role.Permissions,
			"is_default":      role.IsDefault,
			"is_system":       role.IsSystem,
			"hierarchy_level": role.HierarchyLevel,
			"updated_at":      role.UpdatedAt,
		},
	})
}

// DeleteWorkspaceRole deletes a custom role
func (h *Handlers) DeleteWorkspaceRole(c *gin.Context) {
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

	roleID, err := uuid.Parse(c.Param("roleId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
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
	if currentRoleName != "owner" && currentRoleName != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to delete roles"})
		return
	}

	// Get the role to check if it's a system role
	role, err := queries.GetWorkspaceRole(ctx, sqlc.GetWorkspaceRoleParams{
		ID:          pgtype.UUID{Bytes: roleID, Valid: true},
		WorkspaceID: pgtype.UUID{Bytes: workspaceID, Valid: true},
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get role"})
		return
	}

	// Cannot delete system roles
	if role.IsSystem != nil && *role.IsSystem {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot delete system roles"})
		return
	}

	// Delete role (query already checks is_system = FALSE)
	err = queries.DeleteWorkspaceRole(ctx, sqlc.DeleteWorkspaceRoleParams{
		ID:          pgtype.UUID{Bytes: roleID, Valid: true},
		WorkspaceID: pgtype.UUID{Bytes: workspaceID, Valid: true},
	})
	if err != nil {
		log.Printf("DeleteWorkspaceRole error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role deleted"})
}

// AssignRoleRequest represents the request for assigning a role to a member
type AssignRoleRequest struct {
	RoleID string `json:"role_id" binding:"required"`
}

// AssignWorkspaceRole assigns a role to a workspace member
func (h *Handlers) AssignWorkspaceRole(c *gin.Context) {
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

	var req AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
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
	if currentRoleName != "owner" && currentRoleName != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to assign roles"})
		return
	}

	// Get the role to assign
	role, err := queries.GetWorkspaceRole(ctx, sqlc.GetWorkspaceRoleParams{
		ID:          pgtype.UUID{Bytes: roleID, Valid: true},
		WorkspaceID: pgtype.UUID{Bytes: workspaceID, Valid: true},
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get role"})
		return
	}

	// Cannot assign owner role unless you're owner
	if role.Name == "owner" && currentRoleName != "owner" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only owners can assign the owner role"})
		return
	}

	// Update member's role
	member, err := queries.UpdateWorkspaceMemberRole(ctx, sqlc.UpdateWorkspaceMemberRoleParams{
		WorkspaceID: pgtype.UUID{Bytes: workspaceID, Valid: true},
		UserID:      targetUserID,
		RoleID:      pgtype.UUID{Bytes: roleID, Valid: true},
		RoleName:    &role.Name,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Member not found"})
			return
		}
		log.Printf("UpdateWorkspaceMemberRole error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"member": gin.H{
			"user_id":   member.UserID,
			"role_id":   member.RoleID,
			"role_name": member.RoleName,
		},
	})
}

// GetRolePermissions returns permissions for a specific role
func (h *Handlers) GetRolePermissions(c *gin.Context) {
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

	roleID, err := uuid.Parse(c.Param("roleId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
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

	permissions, err := queries.GetWorkspaceRolePermissions(ctx, sqlc.GetWorkspaceRolePermissionsParams{
		ID:          pgtype.UUID{Bytes: roleID, Valid: true},
		WorkspaceID: pgtype.UUID{Bytes: workspaceID, Valid: true},
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
			return
		}
		log.Printf("GetWorkspaceRolePermissions error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get role permissions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"permissions": permissions})
}

// GetCurrentUserPermissions returns the current user's permissions in a workspace
func (h *Handlers) GetCurrentUserPermissions(c *gin.Context) {
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

	// Get member to check role
	member, err := queries.GetWorkspaceMember(ctx, sqlc.GetWorkspaceMemberParams{
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

	permissions, err := queries.GetUserWorkspacePermissions(ctx, sqlc.GetUserWorkspacePermissionsParams{
		WorkspaceID: pgtype.UUID{Bytes: workspaceID, Valid: true},
		UserID:      user.ID,
	})
	if err != nil {
		log.Printf("GetUserWorkspacePermissions error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get permissions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"role_name":   member.RoleName,
		"permissions": permissions,
	})
}
