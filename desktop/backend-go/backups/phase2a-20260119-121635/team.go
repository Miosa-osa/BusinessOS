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

// ListTeamMembers returns all team members for the current user
func (h *Handlers) ListTeamMembers(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	queries := sqlc.New(h.pool)
	members, err := queries.ListTeamMembers(c.Request.Context(), user.ID)
	if err != nil {
		log.Printf("ListTeamMembers error for user %s: %v", user.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list team members", "details": err.Error()})
		return
	}

	log.Printf("ListTeamMembers: found %d members for user %s", len(members), user.ID)
	c.JSON(http.StatusOK, TransformTeamMemberListRows(members))
}

// CreateTeamMember creates a new team member
func (h *Handlers) CreateTeamMember(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var req struct {
		Name       string   `json:"name" binding:"required"`
		Email      string   `json:"email" binding:"required"`
		Role       string   `json:"role" binding:"required"`
		AvatarUrl  *string  `json:"avatar_url"`
		Status     *string  `json:"status"`
		Capacity   *int32   `json:"capacity"`
		ManagerID  *string  `json:"manager_id"`
		Skills     []string `json:"skills"`
		HourlyRate *float64 `json:"hourly_rate"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	queries := sqlc.New(h.pool)

	// Parse optional fields
	var status sqlc.NullMemberstatus
	if req.Status != nil {
		status = sqlc.NullMemberstatus{
			Memberstatus: stringToMemberStatus(*req.Status),
			Valid:        true,
		}
	}

	var managerID pgtype.UUID
	if req.ManagerID != nil {
		if parsed, err := uuid.Parse(*req.ManagerID); err == nil {
			managerID = pgtype.UUID{Bytes: parsed, Valid: true}
		}
	}

	// Handle skills array as JSONB
	skills := []byte("[]")
	if req.Skills != nil && len(req.Skills) > 0 {
		if skillsJSON, err := json.Marshal(req.Skills); err == nil {
			skills = skillsJSON
		}
	}

	// Handle hourly rate
	var hourlyRate pgtype.Numeric
	if req.HourlyRate != nil {
		hourlyRate = pgtype.Numeric{Valid: true}
		hourlyRate.Scan(*req.HourlyRate)
	}

	member, err := queries.CreateTeamMember(c.Request.Context(), sqlc.CreateTeamMemberParams{
		UserID:     user.ID,
		Name:       req.Name,
		Email:      req.Email,
		Role:       req.Role,
		AvatarUrl:  req.AvatarUrl,
		Status:     status,
		Capacity:   req.Capacity,
		ManagerID:  managerID,
		Skills:     skills,
		HourlyRate: hourlyRate,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create team member"})
		return
	}

	c.JSON(http.StatusCreated, member)
}

// GetTeamMember returns a single team member
func (h *Handlers) GetTeamMember(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team member ID"})
		return
	}

	queries := sqlc.New(h.pool)
	member, err := queries.GetTeamMember(c.Request.Context(), sqlc.GetTeamMemberParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team member not found"})
		return
	}

	// Check if activities are requested
	if c.Query("include_activities") == "true" {
		limit := int32(20)
		activities, err := queries.GetTeamMemberActivities(c.Request.Context(), sqlc.GetTeamMemberActivitiesParams{
			MemberID: pgtype.UUID{Bytes: id, Valid: true},
			Limit:    limit,
		})
		if err == nil {
			c.JSON(http.StatusOK, gin.H{
				"member":     member,
				"activities": activities,
			})
			return
		}
	}

	c.JSON(http.StatusOK, member)
}

// UpdateTeamMember updates a team member
func (h *Handlers) UpdateTeamMember(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team member ID"})
		return
	}

	var req struct {
		Name       *string  `json:"name"`
		Email      *string  `json:"email"`
		Role       *string  `json:"role"`
		AvatarUrl  *string  `json:"avatar_url"`
		Status     *string  `json:"status"`
		Capacity   *int32   `json:"capacity"`
		ManagerID  *string  `json:"manager_id"`
		Skills     []string `json:"skills"`
		HourlyRate *float64 `json:"hourly_rate"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	queries := sqlc.New(h.pool)

	// Get existing member
	existing, err := queries.GetTeamMember(c.Request.Context(), sqlc.GetTeamMemberParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team member not found"})
		return
	}

	// Build update params with existing values as defaults
	name := existing.Name
	if req.Name != nil {
		name = *req.Name
	}

	email := existing.Email
	if req.Email != nil {
		email = *req.Email
	}

	role := existing.Role
	if req.Role != nil {
		role = *req.Role
	}

	avatarUrl := existing.AvatarUrl
	if req.AvatarUrl != nil {
		avatarUrl = req.AvatarUrl
	}

	status := existing.Status
	if req.Status != nil {
		status = sqlc.NullMemberstatus{
			Memberstatus: stringToMemberStatus(*req.Status),
			Valid:        true,
		}
	}

	capacity := existing.Capacity
	if req.Capacity != nil {
		capacity = req.Capacity
	}

	managerID := existing.ManagerID
	if req.ManagerID != nil {
		if parsed, err := uuid.Parse(*req.ManagerID); err == nil {
			managerID = pgtype.UUID{Bytes: parsed, Valid: true}
		}
	}

	skills := existing.Skills
	if req.Skills != nil {
		if skillsJSON, err := json.Marshal(req.Skills); err == nil {
			skills = skillsJSON
		}
	}

	hourlyRate := existing.HourlyRate
	if req.HourlyRate != nil {
		hourlyRate = pgtype.Numeric{Valid: true}
		hourlyRate.Scan(*req.HourlyRate)
	}

	member, err := queries.UpdateTeamMember(c.Request.Context(), sqlc.UpdateTeamMemberParams{
		ID:         pgtype.UUID{Bytes: id, Valid: true},
		Name:       name,
		Email:      email,
		Role:       role,
		AvatarUrl:  avatarUrl,
		Status:     status,
		Capacity:   capacity,
		ManagerID:  managerID,
		Skills:     skills,
		HourlyRate: hourlyRate,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update team member"})
		return
	}

	c.JSON(http.StatusOK, member)
}

// UpdateTeamMemberStatus updates a team member's status
func (h *Handlers) UpdateTeamMemberStatus(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team member ID"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	queries := sqlc.New(h.pool)

	// Verify ownership
	_, err = queries.GetTeamMember(c.Request.Context(), sqlc.GetTeamMemberParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team member not found"})
		return
	}

	member, err := queries.UpdateTeamMemberStatus(c.Request.Context(), sqlc.UpdateTeamMemberStatusParams{
		ID: pgtype.UUID{Bytes: id, Valid: true},
		Status: sqlc.NullMemberstatus{
			Memberstatus: stringToMemberStatus(req.Status),
			Valid:        true,
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update team member status"})
		return
	}

	c.JSON(http.StatusOK, member)
}

// UpdateTeamMemberCapacity updates a team member's capacity
func (h *Handlers) UpdateTeamMemberCapacity(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team member ID"})
		return
	}

	var req struct {
		Capacity int32 `json:"capacity" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	queries := sqlc.New(h.pool)

	// Verify ownership
	_, err = queries.GetTeamMember(c.Request.Context(), sqlc.GetTeamMemberParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team member not found"})
		return
	}

	member, err := queries.UpdateTeamMemberCapacity(c.Request.Context(), sqlc.UpdateTeamMemberCapacityParams{
		ID:       pgtype.UUID{Bytes: id, Valid: true},
		Capacity: &req.Capacity,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update team member capacity"})
		return
	}

	c.JSON(http.StatusOK, member)
}

// AddTeamMemberActivity adds an activity to a team member
func (h *Handlers) AddTeamMemberActivity(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team member ID"})
		return
	}

	var req struct {
		ActivityType string `json:"activity_type" binding:"required"`
		Description  string `json:"description" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	queries := sqlc.New(h.pool)

	// Verify ownership
	_, err = queries.GetTeamMember(c.Request.Context(), sqlc.GetTeamMemberParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team member not found"})
		return
	}

	activity, err := queries.CreateTeamMemberActivity(c.Request.Context(), sqlc.CreateTeamMemberActivityParams{
		MemberID:     pgtype.UUID{Bytes: id, Valid: true},
		ActivityType: req.ActivityType,
		Description:  req.Description,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add activity"})
		return
	}

	c.JSON(http.StatusCreated, activity)
}

// GetTeamMemberActivities returns activities for a team member
func (h *Handlers) GetTeamMemberActivities(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team member ID"})
		return
	}

	queries := sqlc.New(h.pool)

	// Verify ownership
	_, err = queries.GetTeamMember(c.Request.Context(), sqlc.GetTeamMemberParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team member not found"})
		return
	}

	limit := int32(50)
	activities, err := queries.GetTeamMemberActivities(c.Request.Context(), sqlc.GetTeamMemberActivitiesParams{
		MemberID: pgtype.UUID{Bytes: id, Valid: true},
		Limit:    limit,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get activities"})
		return
	}

	c.JSON(http.StatusOK, activities)
}

// DeleteTeamMember deletes a team member
func (h *Handlers) DeleteTeamMember(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team member ID"})
		return
	}

	queries := sqlc.New(h.pool)
	err = queries.DeleteTeamMember(c.Request.Context(), sqlc.DeleteTeamMemberParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete team member"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Team member deleted"})
}

// stringToMemberStatus converts a string to sqlc.Memberstatus
func stringToMemberStatus(s string) sqlc.Memberstatus {
	typeMap := map[string]sqlc.Memberstatus{
		"available":  sqlc.MemberstatusAVAILABLE,
		"busy":       sqlc.MemberstatusBUSY,
		"overloaded": sqlc.MemberstatusOVERLOADED,
		"ooo":        sqlc.MemberstatusOOO,
	}
	if enum, ok := typeMap[strings.ToLower(s)]; ok {
		return enum
	}
	return sqlc.MemberstatusAVAILABLE
}
