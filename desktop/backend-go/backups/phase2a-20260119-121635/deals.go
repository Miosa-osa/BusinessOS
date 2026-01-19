package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rhl/businessos-backend/internal/database/sqlc"
	"github.com/rhl/businessos-backend/internal/middleware"
)

var _ = uuid.UUID{} // keep uuid import for parsing

// ListDeals returns all deals for the current user
func (h *Handlers) ListDeals(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	queries := sqlc.New(h.pool)

	// Get optional stage filter
	stageFilter := c.Query("stage")

	stage := stringToDealStage(stageFilter)
	deals, err := queries.ListDeals(c.Request.Context(), sqlc.ListDealsParams{
		UserID: user.ID,
		Stage:  sqlc.NullDealstage{Dealstage: stage, Valid: stageFilter != ""},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list deals"})
		return
	}

	c.JSON(http.StatusOK, TransformDealListRows(deals))
}

// UpdateDealStage updates the stage of a deal
func (h *Handlers) UpdateDealStage(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid deal ID"})
		return
	}

	var req struct {
		Stage string `json:"stage" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	queries := sqlc.New(h.pool)

	stage := stringToDealStage(req.Stage)
	deal, err := queries.UpdateDealStage(c.Request.Context(), sqlc.UpdateDealStageParams{
		ID:    pgtype.UUID{Bytes: id, Valid: true},
		Stage: sqlc.NullDealstage{Dealstage: stage, Valid: true},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update deal stage"})
		return
	}

	c.JSON(http.StatusOK, TransformDeal(deal))
}
