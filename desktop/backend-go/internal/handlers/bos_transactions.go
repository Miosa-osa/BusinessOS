package handlers

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/transactions"
	"github.com/rhl/businessos-backend/internal/utils"
)

// BOSTransactionHandler implements 2-phase commit endpoints for BusinessOS
type BOSTransactionHandler struct {
	coordinator *transactions.TransactionCoordinator
	logger      *slog.Logger
}

// NewBOSTransactionHandler creates a new handler for 2PC transactions
func NewBOSTransactionHandler(db *pgxpool.Pool, logger *slog.Logger) *BOSTransactionHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return &BOSTransactionHandler{
		coordinator: transactions.NewTransactionCoordinator(db, logger),
		logger:      logger,
	}
}

// RegisterRoutes registers all 2PC transaction routes
func (h *BOSTransactionHandler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/bos/tx")
	{
		group.POST("/prepare", h.Prepare)
		group.POST("/commit", h.Commit)
		group.POST("/abort", h.Abort)
		group.GET("/status/:xid", h.GetStatus)
	}
}

// ─── Request/Response Types ───────────────────────────────────────────────────

// PrepareRequestPayload represents an incoming prepare request
type PrepareRequestPayload struct {
	TransactionID string                            `json:"transaction_id" binding:"required"`
	Algorithm     string                            `json:"algorithm" binding:"required"`
	LogData       transactions.LogData              `json:"log_data" binding:"required"`
	Parameters    transactions.AlgorithmParams      `json:"parameters" binding:"required"`
	TimeoutMS     int64                             `json:"timeout_ms,omitempty"`
}

// PrepareResponsePayload is the HTTP response for prepare
type PrepareResponsePayload struct {
	TransactionID string                  `json:"transaction_id"`
	Status        string                  `json:"status"`
	Vote          string                  `json:"vote,omitempty"`
	Version       int                     `json:"version"`
	Model         *transactions.ModelInfo `json:"model,omitempty"`
	Error         *transactions.ParticipantErr `json:"error,omitempty"`
	Timestamp     time.Time               `json:"timestamp"`
}

// CommitRequestPayload represents an incoming commit request
type CommitRequestPayload struct {
	TransactionID string `json:"transaction_id" binding:"required"`
}

// CommitResponsePayload is the HTTP response for commit
type CommitResponsePayload struct {
	TransactionID string    `json:"transaction_id"`
	Status        string    `json:"status"`
	Version       int       `json:"version"`
	Timestamp     time.Time `json:"timestamp"`
}

// AbortRequestPayload represents an incoming abort request
type AbortRequestPayload struct {
	TransactionID string `json:"transaction_id" binding:"required"`
	Reason        string `json:"reason,omitempty"`
}

// AbortResponsePayload is the HTTP response for abort
type AbortResponsePayload struct {
	TransactionID string    `json:"transaction_id"`
	Status        string    `json:"status"`
	Version       int       `json:"version"`
	Timestamp     time.Time `json:"timestamp"`
}

// StatusResponsePayload is the HTTP response for status queries
type StatusResponsePayload struct {
	TransactionID string    `json:"transaction_id"`
	Status        string    `json:"status"`
	StartedAt     time.Time `json:"started_at"`
	Timestamp     time.Time `json:"timestamp"`
}

// ─── Prepare Endpoint ──────────────────────────────────────────────────────────

// Prepare handles POST /api/bos/tx/prepare
// Request: {xid, algorithm, log_id}
// Response: {xid, status: "prepared", version: N}
//
// @Summary Prepare a 2-phase commit transaction
// @Description Initiates prepare phase with the participant. Returns transaction ID and prepared status.
// @Tags transactions
// @Accept json
// @Produce json
// @Param request body PrepareRequestPayload true "Prepare request with transaction data"
// @Success 200 {object} PrepareResponsePayload "Transaction prepared"
// @Failure 400 {object} utils.ErrorResponse "Invalid request"
// @Failure 500 {object} utils.ErrorResponse "Server error"
// @Router /api/bos/tx/prepare [post]
func (h *BOSTransactionHandler) Prepare(c *gin.Context) {
	requestID := uuid.New().String()
	c.Header("X-Request-ID", requestID)

	var req PrepareRequestPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WarnContext(c.Request.Context(), "invalid prepare request",
			"error", err, "request_id", requestID)
		utils.RespondInvalidRequest(c, h.logger, err)
		return
	}

	ctx := c.Request.Context()
	txID := transactions.TransactionID(req.TransactionID)

	// Begin transaction in coordinator
	if _, err := h.coordinator.BeginTransaction(ctx, "bos-process-model", req.Algorithm); err != nil {
		h.logger.ErrorContext(ctx, "failed to begin transaction",
			"error", err, "request_id", requestID)
		utils.RespondInternalError(c, h.logger, "transaction begin", err)
		return
	}

	// Prepare request
	prepareReq := &transactions.PrepareRequest{
		TransactionID: txID,
		LogData:       req.LogData,
		Algorithm:     req.Algorithm,
		Parameters:    req.Parameters,
		TimeoutMS:     req.TimeoutMS,
	}

	prepareResp, err := h.coordinator.Prepare(ctx, txID, prepareReq)
	if err != nil {
		h.logger.ErrorContext(ctx, "prepare phase failed",
			"error", err, "tx_id", txID, "request_id", requestID)
		utils.RespondInternalError(c, h.logger, "prepare phase", err)
		return
	}

	// Build response
	resp := PrepareResponsePayload{
		TransactionID: string(prepareResp.TransactionID),
		Status:        "prepared",
		Vote:          string(prepareResp.Vote),
		Version:       1, // Version from discovery round
		Model:         prepareResp.Model,
		Error:         prepareResp.Error,
		Timestamp:     time.Now().UTC(),
	}

	h.logger.InfoContext(ctx, "prepare succeeded",
		"tx_id", txID, "vote", prepareResp.Vote, "request_id", requestID)

	c.Header("X-Request-ID", requestID)
	c.JSON(http.StatusOK, resp)
}

// ─── Commit Endpoint ───────────────────────────────────────────────────────────

// Commit handles POST /api/bos/tx/commit
// Request: {xid}
// Response: {xid, status: "committed"}
//
// @Summary Commit a prepared 2-phase commit transaction
// @Description Finalizes a transaction that passed the prepare phase. Persists model to database.
// @Tags transactions
// @Accept json
// @Produce json
// @Param request body CommitRequestPayload true "Commit request with transaction ID"
// @Success 200 {object} CommitResponsePayload "Transaction committed"
// @Failure 400 {object} utils.ErrorResponse "Invalid request"
// @Failure 404 {object} utils.ErrorResponse "Transaction not found"
// @Failure 500 {object} utils.ErrorResponse "Server error"
// @Router /api/bos/tx/commit [post]
func (h *BOSTransactionHandler) Commit(c *gin.Context) {
	requestID := uuid.New().String()
	c.Header("X-Request-ID", requestID)

	var req CommitRequestPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WarnContext(c.Request.Context(), "invalid commit request",
			"error", err, "request_id", requestID)
		utils.RespondInvalidRequest(c, h.logger, err)
		return
	}

	ctx := c.Request.Context()
	txID := transactions.TransactionID(req.TransactionID)

	// Commit the transaction
	if err := h.coordinator.Commit(ctx, txID); err != nil {
		h.logger.ErrorContext(ctx, "commit phase failed",
			"error", err, "tx_id", txID, "request_id", requestID)
		utils.RespondInternalError(c, h.logger, "commit phase", err)
		return
	}

	// Build response
	resp := CommitResponsePayload{
		TransactionID: string(txID),
		Status:        "committed",
		Version:       1,
		Timestamp:     time.Now().UTC(),
	}

	h.logger.InfoContext(ctx, "commit succeeded",
		"tx_id", txID, "request_id", requestID)

	c.Header("X-Request-ID", requestID)
	c.JSON(http.StatusOK, resp)
}

// ─── Abort Endpoint ────────────────────────────────────────────────────────────

// Abort handles POST /api/bos/tx/abort
// Request: {xid}
// Response: {xid, status: "aborted"}
//
// @Summary Abort a 2-phase commit transaction
// @Description Rolls back a transaction that failed in the prepare phase or was explicitly aborted.
// @Tags transactions
// @Accept json
// @Produce json
// @Param request body AbortRequestPayload true "Abort request with transaction ID"
// @Success 200 {object} AbortResponsePayload "Transaction aborted"
// @Failure 400 {object} utils.ErrorResponse "Invalid request"
// @Failure 500 {object} utils.ErrorResponse "Server error"
// @Router /api/bos/tx/abort [post]
func (h *BOSTransactionHandler) Abort(c *gin.Context) {
	requestID := uuid.New().String()
	c.Header("X-Request-ID", requestID)

	var req AbortRequestPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WarnContext(c.Request.Context(), "invalid abort request",
			"error", err, "request_id", requestID)
		utils.RespondInvalidRequest(c, h.logger, err)
		return
	}

	ctx := c.Request.Context()
	txID := transactions.TransactionID(req.TransactionID)

	reason := req.Reason
	if reason == "" {
		reason = "client_abort"
	}

	// Abort the transaction
	if err := h.coordinator.Abort(ctx, txID, reason); err != nil {
		h.logger.ErrorContext(ctx, "abort phase failed",
			"error", err, "tx_id", txID, "reason", reason, "request_id", requestID)
		utils.RespondInternalError(c, h.logger, "abort phase", err)
		return
	}

	// Build response
	resp := AbortResponsePayload{
		TransactionID: string(txID),
		Status:        "aborted",
		Version:       1,
		Timestamp:     time.Now().UTC(),
	}

	h.logger.InfoContext(ctx, "abort succeeded",
		"tx_id", txID, "reason", reason, "request_id", requestID)

	c.Header("X-Request-ID", requestID)
	c.JSON(http.StatusOK, resp)
}

// ─── Status Endpoint ───────────────────────────────────────────────────────────

// GetStatus handles GET /api/bos/tx/status/:xid
// Response: {xid, status, started_at}
//
// @Summary Get transaction status
// @Description Returns the current state of a transaction by its ID.
// @Tags transactions
// @Produce json
// @Param xid path string true "Transaction ID"
// @Success 200 {object} StatusResponsePayload "Transaction status"
// @Failure 404 {object} utils.ErrorResponse "Transaction not found"
// @Failure 500 {object} utils.ErrorResponse "Server error"
// @Router /api/bos/tx/status/{xid} [get]
func (h *BOSTransactionHandler) GetStatus(c *gin.Context) {
	xid := c.Param("xid")
	if xid == "" {
		utils.RespondInvalidRequest(c, h.logger, nil)
		return
	}

	ctx := c.Request.Context()
	txID := transactions.TransactionID(xid)

	record, err := h.coordinator.GetStatus(ctx, txID)
	if err != nil {
		h.logger.WarnContext(ctx, "transaction not found",
			"tx_id", txID, "error", err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "transaction_not_found",
			"tx_id": xid,
		})
		return
	}

	resp := StatusResponsePayload{
		TransactionID: string(record.ID),
		Status:        string(record.State),
		StartedAt:     record.StartedAt,
		Timestamp:     time.Now().UTC(),
	}

	c.JSON(http.StatusOK, resp)
}
