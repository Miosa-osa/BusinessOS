package handlers

import (
	"encoding/json"
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
// WORKSPACE PROFILE HANDLERS
// ============================================================================

// GetWorkspaceProfile returns the current user's profile in a workspace
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

	profile, err := queries.GetUserWorkspaceProfile(ctx, sqlc.GetUserWorkspaceProfileParams{
		WorkspaceID: pgtype.UUID{Bytes: workspaceID, Valid: true},
		UserID:      user.ID,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			// Return empty profile if none exists
			c.JSON(http.StatusOK, gin.H{
				"profile": gin.H{
					"workspace_id": workspaceID,
					"user_id":      user.ID,
					"display_name": nil,
					"title":        nil,
					"department":   nil,
					"avatar_url":   nil,
					"timezone":     nil,
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"profile": profileToJSON(profile),
	})
}

// GetMemberProfile returns another member's profile in a workspace
func (h *Handlers) GetMemberProfile(c *gin.Context) {
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

	// Check membership (current user)
	isMember, err := queries.CheckUserIsWorkspaceMember(ctx, sqlc.CheckUserIsWorkspaceMemberParams{
		WorkspaceID: pgtype.UUID{Bytes: workspaceID, Valid: true},
		UserID:      user.ID,
	})
	if err != nil || !isMember {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not a member of this workspace"})
		return
	}

	// Check if target user is a member
	isTargetMember, err := queries.CheckUserIsWorkspaceMember(ctx, sqlc.CheckUserIsWorkspaceMemberParams{
		WorkspaceID: pgtype.UUID{Bytes: workspaceID, Valid: true},
		UserID:      targetUserID,
	})
	if err != nil || !isTargetMember {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found in this workspace"})
		return
	}

	profile, err := queries.GetUserWorkspaceProfile(ctx, sqlc.GetUserWorkspaceProfileParams{
		WorkspaceID: pgtype.UUID{Bytes: workspaceID, Valid: true},
		UserID:      targetUserID,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			// Return basic info if no profile exists
			c.JSON(http.StatusOK, gin.H{
				"profile": gin.H{
					"workspace_id": workspaceID,
					"user_id":      targetUserID,
					"display_name": nil,
					"title":        nil,
					"department":   nil,
					"avatar_url":   nil,
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get profile"})
		return
	}

	// Return public profile info (exclude sensitive preferences)
	c.JSON(http.StatusOK, gin.H{
		"profile": gin.H{
			"id":              profile.ID,
			"workspace_id":    profile.WorkspaceID,
			"user_id":         profile.UserID,
			"display_name":    profile.DisplayName,
			"title":           profile.Title,
			"department":      profile.Department,
			"avatar_url":      profile.AvatarUrl,
			"timezone":        profile.Timezone,
			"expertise_areas": profile.ExpertiseAreas,
		},
	})
}

// UpdateWorkspaceProfileRequest represents the request body for updating a profile
type UpdateWorkspaceProfileRequest struct {
	DisplayName              *string  `json:"display_name"`
	Title                    *string  `json:"title"`
	Department               *string  `json:"department"`
	AvatarUrl                *string  `json:"avatar_url"`
	WorkEmail                *string  `json:"work_email"`
	Phone                    *string  `json:"phone"`
	Timezone                 *string  `json:"timezone"`
	WorkingHours             any      `json:"working_hours"`
	NotificationPreferences  any      `json:"notification_preferences"`
	PreferredOutputStyle     *string  `json:"preferred_output_style"`
	CommunicationPreferences any      `json:"communication_preferences"`
	ExpertiseAreas           []string `json:"expertise_areas"`
}

// UpdateWorkspaceProfile updates the current user's profile in a workspace
func (h *Handlers) UpdateWorkspaceProfile(c *gin.Context) {
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

	var req UpdateWorkspaceProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

	// Convert JSON fields to bytes
	var workingHoursBytes []byte
	if req.WorkingHours != nil {
		workingHoursBytes, _ = json.Marshal(req.WorkingHours)
	}

	var notificationPrefsBytes []byte
	if req.NotificationPreferences != nil {
		notificationPrefsBytes, _ = json.Marshal(req.NotificationPreferences)
	}

	// Try to get existing profile
	_, err = queries.GetUserWorkspaceProfile(ctx, sqlc.GetUserWorkspaceProfileParams{
		WorkspaceID: pgtype.UUID{Bytes: workspaceID, Valid: true},
		UserID:      user.ID,
	})

	var profile sqlc.UserWorkspaceProfile

	if err == pgx.ErrNoRows {
		// Create new profile
		profile, err = queries.CreateUserWorkspaceProfile(ctx, sqlc.CreateUserWorkspaceProfileParams{
			WorkspaceID:             pgtype.UUID{Bytes: workspaceID, Valid: true},
			UserID:                  user.ID,
			DisplayName:             req.DisplayName,
			Title:                   req.Title,
			Department:              req.Department,
			AvatarUrl:               req.AvatarUrl,
			WorkEmail:               req.WorkEmail,
			Phone:                   req.Phone,
			Timezone:                req.Timezone,
			WorkingHours:            workingHoursBytes,
			NotificationPreferences: notificationPrefsBytes,
			ExpertiseAreas:          req.ExpertiseAreas,
		})
		if err != nil {
			log.Printf("CreateUserWorkspaceProfile error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create profile"})
			return
		}
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get profile"})
		return
	} else {
		// Update existing profile
		profile, err = queries.UpdateUserWorkspaceProfile(ctx, sqlc.UpdateUserWorkspaceProfileParams{
			WorkspaceID:             pgtype.UUID{Bytes: workspaceID, Valid: true},
			UserID:                  user.ID,
			DisplayName:             req.DisplayName,
			Title:                   req.Title,
			Department:              req.Department,
			AvatarUrl:               req.AvatarUrl,
			WorkEmail:               req.WorkEmail,
			Phone:                   req.Phone,
			Timezone:                req.Timezone,
			WorkingHours:            workingHoursBytes,
			NotificationPreferences: notificationPrefsBytes,
			ExpertiseAreas:          req.ExpertiseAreas,
		})
		if err != nil {
			log.Printf("UpdateUserWorkspaceProfile error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"profile": profileToJSON(profile),
	})
}

// UpdateWorkspaceStatus is a quick endpoint to update user status
// Note: Status is managed at the workspace member level, not profile
func (h *Handlers) UpdateWorkspaceStatus(c *gin.Context) {
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

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate status
	validStatuses := map[string]bool{
		"active": true, "away": true, "busy": true, "offline": true,
	}
	if !validStatuses[req.Status] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status"})
		return
	}

	queries := sqlc.New(h.pool)
	ctx := c.Request.Context()

	// Update workspace member status
	err = queries.UpdateWorkspaceMemberStatus(ctx, sqlc.UpdateWorkspaceMemberStatusParams{
		WorkspaceID: pgtype.UUID{Bytes: workspaceID, Valid: true},
		UserID:      user.ID,
		Status:      &req.Status,
	})
	if err != nil {
		log.Printf("UpdateWorkspaceMemberStatus error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": req.Status})
}

// TouchLastActive updates the user's last active timestamp
func (h *Handlers) TouchLastActive(c *gin.Context) {
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

	// This is a best-effort operation - just return success
	// TODO: Add a proper query for this if needed
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// Helper function to convert profile to JSON response
func profileToJSON(p sqlc.UserWorkspaceProfile) gin.H {
	return gin.H{
		"id":                         p.ID,
		"workspace_id":               p.WorkspaceID,
		"user_id":                    p.UserID,
		"display_name":               p.DisplayName,
		"title":                      p.Title,
		"department":                 p.Department,
		"avatar_url":                 p.AvatarUrl,
		"work_email":                 p.WorkEmail,
		"phone":                      p.Phone,
		"timezone":                   p.Timezone,
		"working_hours":              p.WorkingHours,
		"notification_preferences":   p.NotificationPreferences,
		"preferred_output_style":     p.PreferredOutputStyle,
		"communication_preferences":  p.CommunicationPreferences,
		"expertise_areas":            p.ExpertiseAreas,
		"created_at":                 p.CreatedAt,
		"updated_at":                 p.UpdatedAt,
	}
}
