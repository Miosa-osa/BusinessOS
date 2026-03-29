package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rhl/businessos-backend/internal/database/sqlc"
	"github.com/rhl/businessos-backend/internal/middleware"
	"github.com/rhl/businessos-backend/internal/utils"
)

// ============================================================================
// CRM DEALS HANDLERS
// ============================================================================

// ListCRMDeals returns all CRM deals for the current user
func (h *CRMHandler) ListCRMDeals(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}

	queries := sqlc.New(h.pool)

	// Parse query params
	pipelineID := c.Query("pipeline_id")
	stageID := c.Query("stage_id")
	status := c.Query("status")
	ownerID := c.Query("owner_id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	deals, err := queries.ListCRMDeals(c.Request.Context(), sqlc.ListCRMDealsParams{
		UserID:     user.ID,
		PipelineID: crmToNullUUID(pipelineID),
		StageID:    crmToNullUUID(stageID),
		Status:     crmToNullString(status),
		OwnerID:    crmToNullString(ownerID),
		LimitVal:   int32(limit),
		OffsetVal:  int32(offset),
	})
	if err != nil {
		utils.RespondInternalError(c, slog.Default(), "list deals", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"deals": transformCRMDeals(deals),
		"count": len(deals),
	})
}

// GetCRMDeal returns a single CRM deal by ID
func (h *CRMHandler) GetCRMDeal(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondInvalidID(c, slog.Default(), "deal_id")
		return
	}

	queries := sqlc.New(h.pool)
	deal, err := queries.GetCRMDeal(c.Request.Context(), pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		utils.RespondNotFound(c, slog.Default(), "Deal")
		return
	}

	c.JSON(http.StatusOK, transformCRMDealRow(deal))
}

// CreateCRMDealRequest represents the request to create a CRM deal
type CreateCRMDealRequest struct {
	PipelineID        string                 `json:"pipeline_id" binding:"required"`
	StageID           string                 `json:"stage_id" binding:"required"`
	Name              string                 `json:"name" binding:"required"`
	Description       *string                `json:"description"`
	Amount            *float64               `json:"amount"`
	Currency          *string                `json:"currency"`
	Probability       *int32                 `json:"probability"`
	ExpectedCloseDate *string                `json:"expected_close_date"`
	OwnerID           *string                `json:"owner_id"`
	CompanyID         *string                `json:"company_id"`
	PrimaryContactID  *string                `json:"primary_contact_id"`
	Status            *string                `json:"status"`
	Priority          *string                `json:"priority"`
	LeadSource        *string                `json:"lead_source"`
	CustomFields      map[string]interface{} `json:"custom_fields"`
}

// CreateCRMDeal creates a new CRM deal
func (h *CRMHandler) CreateCRMDeal(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}

	var req CreateCRMDealRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, slog.Default(), err)
		return
	}

	pipelineID, err := uuid.Parse(req.PipelineID)
	if err != nil {
		utils.RespondInvalidID(c, slog.Default(), "pipeline_id")
		return
	}

	stageID, err := uuid.Parse(req.StageID)
	if err != nil {
		utils.RespondInvalidID(c, slog.Default(), "stage_id")
		return
	}

	queries := sqlc.New(h.pool)

	// Convert custom fields to JSON
	customFields, _ := json.Marshal(req.CustomFields)

	// Parse expected close date
	var expectedCloseDate pgtype.Date
	if req.ExpectedCloseDate != nil {
		if t, err := time.Parse("2006-01-02", *req.ExpectedCloseDate); err == nil {
			expectedCloseDate = pgtype.Date{Time: t, Valid: true}
		}
	}

	deal, err := queries.CreateCRMDeal(c.Request.Context(), sqlc.CreateCRMDealParams{
		UserID:            user.ID,
		PipelineID:        pgtype.UUID{Bytes: pipelineID, Valid: true},
		StageID:           pgtype.UUID{Bytes: stageID, Valid: true},
		Name:              req.Name,
		Description:       req.Description,
		Amount:            crmToNumeric(req.Amount),
		Currency:          req.Currency,
		Probability:       req.Probability,
		ExpectedCloseDate: expectedCloseDate,
		OwnerID:           req.OwnerID,
		CompanyID:         crmToNullUUID(crmPtrToString(req.CompanyID)),
		PrimaryContactID:  crmToNullUUID(crmPtrToString(req.PrimaryContactID)),
		Status:            req.Status,
		Priority:          req.Priority,
		LeadSource:        req.LeadSource,
		CustomFields:      customFields,
	})
	if err != nil {
		utils.RespondInternalError(c, slog.Default(), "create deal", err)
		return
	}

	c.JSON(http.StatusCreated, transformCRMDealBasic(deal))
}

// UpdateCRMDealRequest represents the request to update a CRM deal
type UpdateCRMDealRequest struct {
	Name              string                 `json:"name" binding:"required"`
	Description       *string                `json:"description"`
	Amount            *float64               `json:"amount"`
	Probability       *int32                 `json:"probability"`
	ExpectedCloseDate *string                `json:"expected_close_date"`
	OwnerID           *string                `json:"owner_id"`
	CompanyID         *string                `json:"company_id"`
	PrimaryContactID  *string                `json:"primary_contact_id"`
	Priority          *string                `json:"priority"`
	CustomFields      map[string]interface{} `json:"custom_fields"`
}

// UpdateCRMDeal updates an existing CRM deal
func (h *CRMHandler) UpdateCRMDeal(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondInvalidID(c, slog.Default(), "deal_id")
		return
	}

	var req UpdateCRMDealRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, slog.Default(), err)
		return
	}

	queries := sqlc.New(h.pool)

	// Convert custom fields to JSON
	customFields, _ := json.Marshal(req.CustomFields)

	// Parse expected close date
	var expectedCloseDate pgtype.Date
	if req.ExpectedCloseDate != nil {
		if t, err := time.Parse("2006-01-02", *req.ExpectedCloseDate); err == nil {
			expectedCloseDate = pgtype.Date{Time: t, Valid: true}
		}
	}

	deal, err := queries.UpdateCRMDeal(c.Request.Context(), sqlc.UpdateCRMDealParams{
		ID:                pgtype.UUID{Bytes: id, Valid: true},
		Name:              req.Name,
		Description:       req.Description,
		Amount:            crmToNumeric(req.Amount),
		Probability:       req.Probability,
		ExpectedCloseDate: expectedCloseDate,
		OwnerID:           req.OwnerID,
		CompanyID:         crmToNullUUID(crmPtrToString(req.CompanyID)),
		PrimaryContactID:  crmToNullUUID(crmPtrToString(req.PrimaryContactID)),
		Priority:          req.Priority,
		CustomFields:      customFields,
	})
	if err != nil {
		utils.RespondInternalError(c, slog.Default(), "update deal", nil)
		return
	}

	c.JSON(http.StatusOK, transformCRMDealBasic(deal))
}

// MoveCRMDealStage moves a deal to a different stage
func (h *CRMHandler) MoveCRMDealStage(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondInvalidID(c, slog.Default(), "deal_id")
		return
	}

	var req struct {
		StageID string `json:"stage_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, slog.Default(), err)
		return
	}

	stageID, err := uuid.Parse(req.StageID)
	if err != nil {
		utils.RespondInvalidID(c, slog.Default(), "stage_id")
		return
	}

	queries := sqlc.New(h.pool)
	deal, err := queries.UpdateCRMDealStage(c.Request.Context(), sqlc.UpdateCRMDealStageParams{
		ID:      pgtype.UUID{Bytes: id, Valid: true},
		StageID: pgtype.UUID{Bytes: stageID, Valid: true},
	})
	if err != nil {
		utils.RespondInternalError(c, slog.Default(), "move deal", nil)
		return
	}

	c.JSON(http.StatusOK, transformCRMDealBasic(deal))
}

// UpdateCRMDealStatus updates the status of a deal (open, won, lost)
func (h *CRMHandler) UpdateCRMDealStatus(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondInvalidID(c, slog.Default(), "deal_id")
		return
	}

	var req struct {
		Status     string  `json:"status" binding:"required"`
		LostReason *string `json:"lost_reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, slog.Default(), err)
		return
	}

	queries := sqlc.New(h.pool)
	deal, err := queries.UpdateCRMDealStatus(c.Request.Context(), sqlc.UpdateCRMDealStatusParams{
		ID:         pgtype.UUID{Bytes: id, Valid: true},
		Status:     &req.Status,
		LostReason: req.LostReason,
	})
	if err != nil {
		utils.RespondInternalError(c, slog.Default(), "update deal status", nil)
		return
	}

	c.JSON(http.StatusOK, transformCRMDealBasic(deal))
}

// DeleteCRMDeal deletes a CRM deal
func (h *CRMHandler) DeleteCRMDeal(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondInvalidID(c, slog.Default(), "deal_id")
		return
	}

	queries := sqlc.New(h.pool)
	err = queries.DeleteCRMDeal(c.Request.Context(), sqlc.DeleteCRMDealParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		utils.RespondInternalError(c, slog.Default(), "delete deal", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Deal deleted"})
}

// GetCRMDealStats returns deal statistics for the current user
func (h *CRMHandler) GetCRMDealStats(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}

	pipelineID := c.Query("pipeline_id")

	queries := sqlc.New(h.pool)
	stats, err := queries.GetCRMDealStats(c.Request.Context(), sqlc.GetCRMDealStatsParams{
		UserID:     user.ID,
		PipelineID: crmToNullUUID(pipelineID),
	})
	if err != nil {
		utils.RespondInternalError(c, slog.Default(), "get deal stats", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total_deals": stats.TotalDeals,
		"open_deals":  stats.OpenDeals,
		"won_deals":   stats.WonDeals,
		"lost_deals":  stats.LostDeals,
		"open_value":  stats.OpenValue,
		"won_value":   stats.WonValue,
		"lost_value":  stats.LostValue,
	})
}

// ============================================================================
// LEGACY DEALS HANDLERS (simple deal stage tracking, pre-CRM pipeline)
// ============================================================================

// ListDeals returns all deals for the current user
func (h *CRMHandler) ListDeals(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.RespondUnauthorized(c, slog.Default())
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
		utils.RespondInternalError(c, slog.Default(), "list deals", err)
		return
	}

	c.JSON(http.StatusOK, TransformDealListRows(deals))
}

// UpdateDealStage updates the stage of a deal
func (h *CRMHandler) UpdateDealStage(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondInvalidID(c, slog.Default(), "deal ID")
		return
	}

	var req struct {
		Stage string `json:"stage" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, slog.Default(), err)
		return
	}

	queries := sqlc.New(h.pool)

	stage := stringToDealStage(req.Stage)
	deal, err := queries.UpdateDealStage(c.Request.Context(), sqlc.UpdateDealStageParams{
		ID:    pgtype.UUID{Bytes: id, Valid: true},
		Stage: sqlc.NullDealstage{Dealstage: stage, Valid: true},
	})
	if err != nil {
		utils.RespondInternalError(c, slog.Default(), "update deal stage", err)
		return
	}

	c.JSON(http.StatusOK, TransformDeal(deal))
}
