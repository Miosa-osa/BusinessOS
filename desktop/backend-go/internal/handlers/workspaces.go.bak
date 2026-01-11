package handlers

import (
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rhl/businessos-backend/internal/database/sqlc"
	"github.com/rhl/businessos-backend/internal/middleware"
)

// ============================================================================
// WORKSPACE HANDLERS
// ============================================================================

// CreateWorkspaceRequest represents the request body for creating a workspace
type CreateWorkspaceRequest struct {
	Name        string  `json:"name" binding:"required,min=2,max=255"`
	Slug        *string `json:"slug"`
	Description *string `json:"description"`
	LogoURL     *string `json:"logo_url"`
}

// UpdateWorkspaceRequest represents the request body for updating a workspace
type UpdateWorkspaceRequest struct {
	Name        *string                `json:"name"`
	Slug        *string                `json:"slug"`
	Description *string                `json:"description"`
	LogoURL     *string                `json:"logo_url"`
	Settings    map[string]interface{} `json:"settings"`
}

// CreateWorkspace creates a new workspace and adds the creator as owner
func (h *Handlers) CreateWorkspace(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var req CreateWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	queries := sqlc.New(h.pool)
	ctx := c.Request.Context()

	// Generate slug if not provided
	slug := generateSlug(req.Name)
	if req.Slug != nil && *req.Slug != "" {
		slug = generateSlug(*req.Slug)
	}

	// Check slug uniqueness and append number if needed
	baseSlug := slug
	counter := 2
	for {
		exists, err := queries.CheckSlugExists(ctx, slug)
		if err != nil {
			log.Printf("CheckSlugExists error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check slug"})
			return
		}
		if !exists {
			break
		}
		slug = baseSlug + "-" + string(rune('0'+counter))
		counter++
		if counter > 100 {
			c.JSON(http.StatusConflict, gin.H{"error": "Could not generate unique slug"})
			return
		}
	}

	// Start transaction
	tx, err := h.pool.Begin(ctx)
	if err != nil {
		log.Printf("Begin transaction error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}
	defer tx.Rollback(ctx)

	qtx := queries.WithTx(tx)

	// Create workspace
	workspace, err := qtx.CreateWorkspace(ctx, sqlc.CreateWorkspaceParams{
		Name:        req.Name,
		Slug:        slug,
		Description: req.Description,
		LogoUrl:     req.LogoURL,
		PlanType:    stringPtr("free"),
		OwnerID:     user.ID,
	})
	if err != nil {
		log.Printf("CreateWorkspace error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create workspace"})
		return
	}

	// Seed default roles using the function we created in the migration
	_, err = tx.Exec(ctx, "SELECT seed_workspace_default_roles($1)", workspace.ID)
	if err != nil {
		log.Printf("seed_workspace_default_roles error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to seed default roles"})
		return
	}

	// Get the owner role
	ownerRole, err := qtx.GetWorkspaceRoleByName(ctx, sqlc.GetWorkspaceRoleByNameParams{
		WorkspaceID: workspace.ID,
		Name:        "owner",
	})
	if err != nil {
		log.Printf("GetWorkspaceRoleByName error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get owner role"})
		return
	}

	// Add creator as owner member
	now := time.Now()
	member, err := qtx.CreateWorkspaceMember(ctx, sqlc.CreateWorkspaceMemberParams{
		WorkspaceID: workspace.ID,
		UserID:      user.ID,
		RoleID:      ownerRole.ID,
		RoleName:    stringPtr("owner"),
		Status:      stringPtr("active"),
		JoinedAt:    pgtype.Timestamptz{Time: now, Valid: true},
	})
	if err != nil {
		log.Printf("CreateWorkspaceMember error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add owner member"})
		return
	}

	// Commit transaction
	if err = tx.Commit(ctx); err != nil {
		log.Printf("Commit transaction error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"workspace": transformWorkspace(workspace),
		"member": gin.H{
			"role_name": member.RoleName,
			"joined_at": member.JoinedAt,
		},
	})
}

// ListWorkspaces returns all workspaces the current user is a member of
func (h *Handlers) ListWorkspaces(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	queries := sqlc.New(h.pool)
	workspaces, err := queries.ListUserWorkspaces(c.Request.Context(), user.ID)
	if err != nil {
		log.Printf("ListUserWorkspaces error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list workspaces"})
		return
	}

	result := make([]gin.H, len(workspaces))
	for i, w := range workspaces {
		result[i] = gin.H{
			"id":            w.ID,
			"name":          w.Name,
			"slug":          w.Slug,
			"description":   w.Description,
			"logo_url":      w.LogoUrl,
			"plan_type":     w.PlanType,
			"role":          w.RoleName,
			"member_count":  w.MemberCount,
			"member_status": w.MemberStatus,
			"joined_at":     w.JoinedAt,
			"created_at":    w.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{"workspaces": result})
}

// GetWorkspace returns details of a specific workspace
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

	queries := sqlc.New(h.pool)
	ctx := c.Request.Context()

	// Check membership
	member, err := queries.GetWorkspaceMember(ctx, sqlc.GetWorkspaceMemberParams{
		WorkspaceID: pgtype.UUID{Bytes: workspaceID, Valid: true},
		UserID:      user.ID,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusForbidden, gin.H{"error": "Not a member of this workspace"})
			return
		}
		log.Printf("GetWorkspaceMember error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check membership"})
		return
	}

	workspace, err := queries.GetWorkspaceByID(ctx, pgtype.UUID{Bytes: workspaceID, Valid: true})
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Workspace not found"})
			return
		}
		log.Printf("GetWorkspaceByID error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get workspace"})
		return
	}

	memberCount, _ := queries.CountWorkspaceMembers(ctx, pgtype.UUID{Bytes: workspaceID, Valid: true})

	c.JSON(http.StatusOK, gin.H{
		"id":           workspace.ID,
		"name":         workspace.Name,
		"slug":         workspace.Slug,
		"description":  workspace.Description,
		"logo_url":     workspace.LogoUrl,
		"plan_type":    workspace.PlanType,
		"max_members":  workspace.MaxMembers,
		"max_projects": workspace.MaxProjects,
		"settings":     workspace.Settings,
		"owner_id":     workspace.OwnerID,
		"member_count": memberCount,
		"my_role": gin.H{
			"name":            member.RoleName,
			"display_name":    member.RoleDisplayName,
			"color":           member.RoleColor,
			"hierarchy_level": member.HierarchyLevel,
			"permissions":     member.Permissions,
		},
		"created_at": workspace.CreatedAt,
		"updated_at": workspace.UpdatedAt,
	})
}

// UpdateWorkspace updates a workspace's details
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

	var req UpdateWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	queries := sqlc.New(h.pool)
	ctx := c.Request.Context()

	// Check permission (owner or admin)
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

	roleName := ""
	if member.RoleName != nil {
		roleName = *member.RoleName
	}
	if roleName != "owner" && roleName != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only owners and admins can update workspace settings"})
		return
	}

	// If slug is being changed, check uniqueness
	if req.Slug != nil && *req.Slug != "" {
		newSlug := generateSlug(*req.Slug)
		exists, err := queries.CheckSlugExistsExcluding(ctx, sqlc.CheckSlugExistsExcludingParams{
			Slug: newSlug,
			ID:   pgtype.UUID{Bytes: workspaceID, Valid: true},
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check slug"})
			return
		}
		if exists {
			c.JSON(http.StatusConflict, gin.H{"error": "Slug is already taken"})
			return
		}
		req.Slug = &newSlug
	}

	workspace, err := queries.UpdateWorkspace(ctx, sqlc.UpdateWorkspaceParams{
		ID:          pgtype.UUID{Bytes: workspaceID, Valid: true},
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		LogoUrl:     req.LogoURL,
	})
	if err != nil {
		log.Printf("UpdateWorkspace error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update workspace"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"workspace": transformWorkspace(workspace)})
}

// DeleteWorkspace deletes a workspace (owner only)
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

	queries := sqlc.New(h.pool)
	ctx := c.Request.Context()

	// Check if user is owner
	isOwner, err := queries.CheckUserIsWorkspaceOwner(ctx, sqlc.CheckUserIsWorkspaceOwnerParams{
		ID:      pgtype.UUID{Bytes: workspaceID, Valid: true},
		OwnerID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check ownership"})
		return
	}
	if !isOwner {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the workspace owner can delete it"})
		return
	}

	// Optional confirmation check
	confirm := c.Query("confirm")
	workspace, _ := queries.GetWorkspaceByID(ctx, pgtype.UUID{Bytes: workspaceID, Valid: true})
	if confirm != workspace.Slug {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Please confirm deletion by adding ?confirm=" + workspace.Slug,
			"message": "This action cannot be undone",
		})
		return
	}

	err = queries.DeleteWorkspace(ctx, pgtype.UUID{Bytes: workspaceID, Valid: true})
	if err != nil {
		log.Printf("DeleteWorkspace error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete workspace"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Workspace deleted successfully"})
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// generateSlug creates a URL-safe slug from a name
func generateSlug(name string) string {
	// Convert to lowercase
	slug := strings.ToLower(name)
	// Replace spaces with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")
	// Remove special characters
	reg := regexp.MustCompile("[^a-z0-9-]")
	slug = reg.ReplaceAllString(slug, "")
	// Remove multiple consecutive hyphens
	reg = regexp.MustCompile("-+")
	slug = reg.ReplaceAllString(slug, "-")
	// Trim hyphens from start and end
	slug = strings.Trim(slug, "-")
	// Limit length
	if len(slug) > 100 {
		slug = slug[:100]
	}
	return slug
}

// transformWorkspace converts a workspace row to a response object
func transformWorkspace(w sqlc.Workspace) gin.H {
	return gin.H{
		"id":           w.ID,
		"name":         w.Name,
		"slug":         w.Slug,
		"description":  w.Description,
		"logo_url":     w.LogoUrl,
		"plan_type":    w.PlanType,
		"max_members":  w.MaxMembers,
		"max_projects": w.MaxProjects,
		"settings":     w.Settings,
		"owner_id":     w.OwnerID,
		"created_at":   w.CreatedAt,
		"updated_at":   w.UpdatedAt,
	}
}
