package handlers

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rhl/businessos-backend/internal/database/sqlc"
	"github.com/rhl/businessos-backend/internal/middleware"
	"github.com/rhl/businessos-backend/internal/utils"
)

// ============================================================================
// PIPELINES HANDLERS
// ============================================================================

// ListPipelines returns all pipelines for the current user
func (h *CRMHandler) ListPipelines(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}

	queries := sqlc.New(h.pool)
	pipelines, err := queries.ListPipelines(c.Request.Context(), user.ID)
	if err != nil {
		slog.Error("Failed to list pipelines", "error", err, "user_id", user.ID)
		utils.RespondInternalError(c, slog.Default(), "list pipelines", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pipelines": transformPipelines(pipelines),
		"count":     len(pipelines),
	})
}

// GetPipeline returns a single pipeline by ID
func (h *CRMHandler) GetPipeline(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondInvalidID(c, slog.Default(), "pipeline_id")
		return
	}

	queries := sqlc.New(h.pool)
	pipeline, err := queries.GetPipeline(c.Request.Context(), pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		utils.RespondNotFound(c, slog.Default(), "Pipeline")
		return
	}

	c.JSON(http.StatusOK, transformPipeline(pipeline))
}

// CreatePipelineRequest represents the request to create a pipeline
type CreatePipelineRequest struct {
	Name         string  `json:"name" binding:"required"`
	Description  *string `json:"description"`
	PipelineType *string `json:"pipeline_type"` // sales, hiring, projects, custom
	Currency     *string `json:"currency"`
	IsDefault    bool    `json:"is_default"`
	Color        *string `json:"color"`
	Icon         *string `json:"icon"`
}

// CreatePipeline creates a new pipeline
func (h *CRMHandler) CreatePipeline(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}

	var req CreatePipelineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, slog.Default(), err)
		return
	}

	queries := sqlc.New(h.pool)
	pipeline, err := queries.CreatePipeline(c.Request.Context(), sqlc.CreatePipelineParams{
		UserID:       user.ID,
		Name:         req.Name,
		Description:  req.Description,
		PipelineType: req.PipelineType,
		Currency:     req.Currency,
		IsDefault:    &req.IsDefault,
		Color:        req.Color,
		Icon:         req.Icon,
	})
	if err != nil {
		utils.RespondInternalError(c, slog.Default(), "create pipeline", nil)
		return
	}

	c.JSON(http.StatusCreated, transformPipeline(pipeline))
}

// UpdatePipelineRequest represents the request to update a pipeline
type UpdatePipelineRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description"`
	Currency    *string `json:"currency"`
	Color       *string `json:"color"`
	Icon        *string `json:"icon"`
	IsActive    bool    `json:"is_active"`
}

// UpdatePipeline updates an existing pipeline
func (h *CRMHandler) UpdatePipeline(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondInvalidID(c, slog.Default(), "pipeline_id")
		return
	}

	var req UpdatePipelineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, slog.Default(), err)
		return
	}

	queries := sqlc.New(h.pool)
	pipeline, err := queries.UpdatePipeline(c.Request.Context(), sqlc.UpdatePipelineParams{
		ID:          pgtype.UUID{Bytes: id, Valid: true},
		Name:        req.Name,
		Description: req.Description,
		Currency:    req.Currency,
		Color:       req.Color,
		Icon:        req.Icon,
		IsActive:    &req.IsActive,
	})
	if err != nil {
		utils.RespondInternalError(c, slog.Default(), "update pipeline", nil)
		return
	}

	c.JSON(http.StatusOK, transformPipeline(pipeline))
}

// DeletePipeline deletes a pipeline
func (h *CRMHandler) DeletePipeline(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondInvalidID(c, slog.Default(), "pipeline_id")
		return
	}

	queries := sqlc.New(h.pool)
	err = queries.DeletePipeline(c.Request.Context(), sqlc.DeletePipelineParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		utils.RespondInternalError(c, slog.Default(), "delete pipeline", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Pipeline deleted"})
}

// ============================================================================
// PIPELINE STAGES HANDLERS
// ============================================================================

// ListPipelineStages returns all stages for a pipeline
func (h *CRMHandler) ListPipelineStages(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}

	pipelineID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondInvalidID(c, slog.Default(), "pipeline_id")
		return
	}

	queries := sqlc.New(h.pool)
	stages, err := queries.ListPipelineStages(c.Request.Context(), pgtype.UUID{Bytes: pipelineID, Valid: true})
	if err != nil {
		utils.RespondInternalError(c, slog.Default(), "list stages", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stages": transformPipelineStages(stages),
		"count":  len(stages),
	})
}

// CreatePipelineStageRequest represents the request to create a stage
type CreatePipelineStageRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description"`
	Position    int32   `json:"position"`
	Probability *int32  `json:"probability"`
	StageType   *string `json:"stage_type"` // open, won, lost
	RottingDays *int32  `json:"rotting_days"`
	Color       *string `json:"color"`
}

// CreatePipelineStage creates a new pipeline stage
func (h *CRMHandler) CreatePipelineStage(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}

	pipelineID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondInvalidID(c, slog.Default(), "pipeline_id")
		return
	}

	var req CreatePipelineStageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, slog.Default(), err)
		return
	}

	queries := sqlc.New(h.pool)
	stage, err := queries.CreatePipelineStage(c.Request.Context(), sqlc.CreatePipelineStageParams{
		PipelineID:  pgtype.UUID{Bytes: pipelineID, Valid: true},
		Name:        req.Name,
		Description: req.Description,
		Position:    req.Position,
		Probability: req.Probability,
		StageType:   req.StageType,
		RottingDays: req.RottingDays,
		Color:       req.Color,
	})
	if err != nil {
		utils.RespondInternalError(c, slog.Default(), "create stage", nil)
		return
	}

	c.JSON(http.StatusCreated, transformPipelineStage(stage))
}

// UpdatePipelineStageRequest represents the request to update a stage
type UpdatePipelineStageRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description"`
	Probability *int32  `json:"probability"`
	RottingDays *int32  `json:"rotting_days"`
	Color       *string `json:"color"`
}

// UpdatePipelineStage updates an existing stage
func (h *CRMHandler) UpdatePipelineStage(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}

	stageID, err := uuid.Parse(c.Param("stageId"))
	if err != nil {
		utils.RespondInvalidID(c, slog.Default(), "stage_id")
		return
	}

	var req UpdatePipelineStageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, slog.Default(), err)
		return
	}

	queries := sqlc.New(h.pool)
	stage, err := queries.UpdatePipelineStage(c.Request.Context(), sqlc.UpdatePipelineStageParams{
		ID:          pgtype.UUID{Bytes: stageID, Valid: true},
		Name:        req.Name,
		Description: req.Description,
		Probability: req.Probability,
		RottingDays: req.RottingDays,
		Color:       req.Color,
	})
	if err != nil {
		utils.RespondInternalError(c, slog.Default(), "update stage", nil)
		return
	}

	c.JSON(http.StatusOK, transformPipelineStage(stage))
}

// ReorderPipelineStages reorders stages in a pipeline
func (h *CRMHandler) ReorderPipelineStages(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}

	var req struct {
		StageOrders []struct {
			ID       string `json:"id"`
			Position int32  `json:"position"`
		} `json:"stage_orders"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, slog.Default(), err)
		return
	}

	queries := sqlc.New(h.pool)
	for _, order := range req.StageOrders {
		stageID, err := uuid.Parse(order.ID)
		if err != nil {
			continue
		}
		queries.UpdateStagePosition(c.Request.Context(), sqlc.UpdateStagePositionParams{
			ID:       pgtype.UUID{Bytes: stageID, Valid: true},
			Position: order.Position,
		})
	}

	c.JSON(http.StatusOK, gin.H{"message": "Stages reordered"})
}

// DeletePipelineStage deletes a stage
func (h *CRMHandler) DeletePipelineStage(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}

	stageID, err := uuid.Parse(c.Param("stageId"))
	if err != nil {
		utils.RespondInvalidID(c, slog.Default(), "stage_id")
		return
	}

	queries := sqlc.New(h.pool)
	err = queries.DeletePipelineStage(c.Request.Context(), pgtype.UUID{Bytes: stageID, Valid: true})
	if err != nil {
		utils.RespondInternalError(c, slog.Default(), "delete stage", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Stage deleted"})
}
