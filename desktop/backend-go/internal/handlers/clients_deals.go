package handlers

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rhl/businessos-backend/internal/database/sqlc"
	"github.com/rhl/businessos-backend/internal/middleware"
	"github.com/rhl/businessos-backend/internal/utils"
)

// ListClientDeals returns all deals for a client.
//
// After the SQLC migration, ListClientDeals will query the deals table
// (JOINed with pipeline_stages) and return rows with Deal-compatible fields.
func (h *ClientHandler) ListClientDeals(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondInvalidID(c, slog.Default(), "client_id")
		return
	}

	queries := sqlc.New(h.pool)

	// Verify client ownership
	_, err = queries.GetClient(c.Request.Context(), sqlc.GetClientParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		utils.RespondNotFound(c, slog.Default(), "Client")
		return
	}

	deals, err := queries.ListClientDeals(c.Request.Context(), pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		utils.RespondInternalError(c, slog.Default(), "list deals", nil)
		return
	}

	c.JSON(http.StatusOK, transformClientDealsFromRows(deals))
}

// CreateClientDeal creates a new deal for a client by inserting into the deals
// table via the CreateClientDeal SQLC query. The SQL auto-resolves the
// user's default pipeline and matches the stage by name.
func (h *ClientHandler) CreateClientDeal(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}

	clientID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondInvalidID(c, slog.Default(), "client_id")
		return
	}

	var req struct {
		Name              string  `json:"name" binding:"required"`
		Value             float64 `json:"value"`
		Stage             *string `json:"stage"`
		Probability       *int32  `json:"probability"`
		ExpectedCloseDate *string `json:"expected_close_date"`
		Notes             *string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, slog.Default(), err)
		return
	}

	queries := sqlc.New(h.pool)

	// Verify client ownership
	_, err = queries.GetClient(c.Request.Context(), sqlc.GetClientParams{
		ID:     pgtype.UUID{Bytes: clientID, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		utils.RespondNotFound(c, slog.Default(), "Client")
		return
	}

	// Parse expected close date
	var expectedCloseDate pgtype.Date
	if req.ExpectedCloseDate != nil {
		if t, err := time.Parse("2006-01-02", *req.ExpectedCloseDate); err == nil {
			expectedCloseDate = pgtype.Date{Time: t, Valid: true}
		}
	}

	// Convert value to pgtype.Numeric (maps to "amount" in the deals table)
	amount := crmToNumeric(&req.Value)

	// Insert into the deals table using CreateClientDeal (SQLC-generated query
	// that auto-resolves pipeline_id and stage_id from the client's owner and
	// the provided stage name).
	stageName := ""
	if req.Stage != nil {
		stageName = *req.Stage
	}
	deal, err := queries.CreateClientDeal(c.Request.Context(), sqlc.CreateClientDealParams{
		ClientID:          pgtype.UUID{Bytes: clientID, Valid: true},
		Name:              req.Name,
		Amount:            amount,
		Name_2:            stageName,
		Probability:       req.Probability,
		ExpectedCloseDate: expectedCloseDate,
		Description:       req.Notes,
	})
	if err != nil {
		slog.Error("failed to create client deal", "error", err)
		utils.RespondInternalError(c, slog.Default(), "create deal", nil)
		return
	}

	c.JSON(http.StatusCreated, transformClientDealFromDeal(deal, clientID))
}

// UpdateClientDeal updates a client deal in the deals table via the UpdateClientDeal
// SQLC query. Stage changes require a separate call to UpdateCRMDealStage.
func (h *ClientHandler) UpdateClientDeal(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}

	clientID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondInvalidID(c, slog.Default(), "client_id")
		return
	}

	dealID, err := uuid.Parse(c.Param("dealId"))
	if err != nil {
		utils.RespondInvalidID(c, slog.Default(), "deal_id")
		return
	}

	var req struct {
		Name              string  `json:"name" binding:"required"`
		Value             float64 `json:"value"`
		Stage             *string `json:"stage"`
		Probability       *int32  `json:"probability"`
		ExpectedCloseDate *string `json:"expected_close_date"`
		Notes             *string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, slog.Default(), err)
		return
	}

	queries := sqlc.New(h.pool)

	// Verify client ownership
	_, err = queries.GetClient(c.Request.Context(), sqlc.GetClientParams{
		ID:     pgtype.UUID{Bytes: clientID, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		utils.RespondNotFound(c, slog.Default(), "Client")
		return
	}

	// Parse expected close date
	var expectedCloseDate pgtype.Date
	if req.ExpectedCloseDate != nil {
		if t, err := time.Parse("2006-01-02", *req.ExpectedCloseDate); err == nil {
			expectedCloseDate = pgtype.Date{Time: t, Valid: true}
		}
	}

	// Convert value to pgtype.Numeric (maps to "amount" in the deals table)
	amount := crmToNumeric(&req.Value)

	// Update the deals table using UpdateClientDeal (SQLC-generated query).
	// Note: stage changes require a separate call to UpdateCRMDealStage because
	// UpdateClientDeal does not update the stage_id column.
	var status *string
	deal, err := queries.UpdateClientDeal(c.Request.Context(), sqlc.UpdateClientDealParams{
		ID:                pgtype.UUID{Bytes: dealID, Valid: true},
		Name:              req.Name,
		Amount:            amount,
		Status:            status,
		Probability:       req.Probability,
		ExpectedCloseDate: expectedCloseDate,
		Description:       req.Notes,
	})
	if err != nil {
		slog.Error("failed to update client deal", "error", err)
		utils.RespondInternalError(c, slog.Default(), "update deal", nil)
		return
	}

	// If stage was provided, move the deal to the new stage
	if req.Stage != nil && *req.Stage != "" {
		pipeline, pErr := queries.GetDefaultPipeline(c.Request.Context(), user.ID)
		if pErr != nil {
			slog.Error("no default pipeline found for user", "user_id", user.ID, "error", pErr)
			// Return the deal as-is without stage update
			c.JSON(http.StatusOK, transformClientDealFromDeal(deal, clientID))
			return
		}

		stageID, sErr := resolveStageID(c, queries, pipeline.ID, req.Stage)
		if sErr != nil {
			slog.Error("failed to resolve stage", "error", sErr)
			c.JSON(http.StatusOK, transformClientDealFromDeal(deal, clientID))
			return
		}

		movedDeal, mErr := queries.UpdateCRMDealStage(c.Request.Context(), sqlc.UpdateCRMDealStageParams{
			ID:      pgtype.UUID{Bytes: dealID, Valid: true},
			StageID: stageID,
		})
		if mErr != nil {
			slog.Error("failed to move deal stage", "error", mErr)
		} else {
			deal = movedDeal
		}
	}

	c.JSON(http.StatusOK, transformClientDealFromDeal(deal, clientID))
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// resolveStageID finds a pipeline stage by name within the given pipeline.
// If stageName is nil or empty, returns the first stage (lowest position).
func resolveStageID(c *gin.Context, queries *sqlc.Queries, pipelineID pgtype.UUID, stageName *string) (pgtype.UUID, error) {
	stages, err := queries.ListPipelineStages(c.Request.Context(), pipelineID)
	if err != nil {
		slog.Error("failed to list pipeline stages", "error", err)
		return pgtype.UUID{}, err
	}

	if len(stages) == 0 {
		slog.Error("pipeline has no stages", "pipeline_id", pipelineID)
		return pgtype.UUID{}, err
	}

	// If no stage specified, use the first stage (lowest position)
	if stageName == nil || *stageName == "" {
		return stages[0].ID, nil
	}

	// Map legacy stage names to pipeline stage display names
	name := mapLegacyStageName(*stageName)

	// Find matching stage by name (case-insensitive)
	for _, stage := range stages {
		if strings.EqualFold(stage.Name, name) {
			return stage.ID, nil
		}
	}

	// Fallback: return the first stage if no match found
	slog.Warn("stage name not found in pipeline, using first stage",
		"stage_name", *stageName, "pipeline_id", pipelineID)
	return stages[0].ID, nil
}

// mapLegacyStageName converts legacy deal stage enum values (qualification,
// proposal, negotiation, closed_won, closed_lost) to pipeline stage display
// names used in the CRM pipeline system.
func mapLegacyStageName(legacy string) string {
	mapping := map[string]string{
		"qualification": "Qualification",
		"proposal":      "Proposal",
		"negotiation":   "Negotiation",
		"closed_won":    "Closed Won",
		"closed_lost":   "Closed Lost",
	}
	if name, ok := mapping[strings.ToLower(legacy)]; ok {
		return name
	}
	return legacy
}

// ============================================================================
// TRANSFORM FUNCTIONS
// ============================================================================

// transformClientDealFromDeal transforms a sqlc.Deal into the client deal JSON
// response shape, mapping deals table fields to the frontend format.
func transformClientDealFromDeal(d sqlc.Deal, clientID uuid.UUID) gin.H {
	return gin.H{
		"id":                  crmUuidToString(d.ID),
		"client_id":           clientID.String(),
		"name":                d.Name,
		"value":               crmNumericToFloat(d.Amount),
		"stage_id":            crmUuidToString(d.StageID),
		"probability":         d.Probability,
		"expected_close_date": crmDateToString(d.ExpectedCloseDate),
		"notes":               d.Description,
		"status":              d.Status,
		"priority":            d.Priority,
		"pipeline_id":         crmUuidToString(d.PipelineID),
		"created_at":          d.CreatedAt.Time.Format(time.RFC3339),
		"updated_at":          d.UpdatedAt.Time.Format(time.RFC3339),
		"closed_at":           crmDateToString(d.ActualCloseDate),
	}
}

// transformClientDealFromRow transforms a ListClientDealsRow into the frontend
// JSON shape, mapping deals table fields back to the legacy response format.
func transformClientDealFromRow(d sqlc.ListClientDealsRow) gin.H {
	return gin.H{
		"id":                  crmUuidToString(d.ID),
		"client_id":           crmUuidToString(d.ClientID),
		"name":                d.Name,
		"value":               crmNumericToFloat(d.Amount),
		"stage":               strings.ToLower(strings.ReplaceAll(d.StageName, " ", "_")),
		"stage_id":            crmUuidToString(d.StageID),
		"probability":         d.Probability,
		"expected_close_date": crmDateToString(d.ExpectedCloseDate),
		"notes":               d.Description,
		"status":              d.Status,
		"priority":            d.Priority,
		"pipeline_id":         crmUuidToString(d.PipelineID),
		"created_at":          d.CreatedAt.Time.Format(time.RFC3339),
		"updated_at":          d.UpdatedAt.Time.Format(time.RFC3339),
		"closed_at":           crmDateToString(d.ActualCloseDate),
	}
}

// transformClientDealsFromRows transforms a slice of ListClientDealsRow into
// JSON response objects.
func transformClientDealsFromRows(deals []sqlc.ListClientDealsRow) []gin.H {
	result := make([]gin.H, len(deals))
	for i, d := range deals {
		result[i] = transformClientDealFromRow(d)
	}
	return result
}
