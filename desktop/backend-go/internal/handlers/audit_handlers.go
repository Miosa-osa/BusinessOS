// Package handlers provides HTTP handlers for audit endpoints.
package handlers

import (
	"encoding/csv"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/rhl/businessos-backend/internal/services"
)

// AuditHandler handles audit log requests.
type AuditHandler struct {
	auditService *services.AuditService
}

// NewAuditHandler creates a new audit handler.
func NewAuditHandler(auditService *services.AuditService) *AuditHandler {
	return &AuditHandler{auditService: auditService}
}

// QueryLogsRequest represents a request to query audit logs.
type QueryLogsRequest struct {
	UserID    *uuid.UUID `form:"user_id"`
	EventType *string    `form:"event_type"`
	FromDate  *time.Time `form:"from_date"`
	ToDate    *time.Time `form:"to_date"`
	Limit     int        `form:"limit"`
}

// GetAuditLogs retrieves audit logs with optional filters.
// GET /api/audit/logs
func (h *AuditHandler) GetAuditLogs(c *gin.Context) {
	var req QueryLogsRequest

	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	if req.Limit == 0 {
		req.Limit = 100
	}

	events, err := h.auditService.QueryAuditLogs(
		c.Request.Context(),
		req.UserID,
		nil,
		req.EventType,
		req.FromDate,
		req.ToDate,
		req.Limit,
	)

	if err != nil {
		slog.ErrorContext(c.Request.Context(), "Failed to query audit logs", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query audit logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"events":      events,
		"total_count": len(events),
	})
}

// VerifyChainRequest represents a chain verification request.
type VerifyChainRequest struct {
	FromSequence int64 `json:"from_sequence"`
	ToSequence   int64 `json:"to_sequence"`
}

// VerifyChainIntegrity verifies the hash chain integrity.
// POST /api/audit/verify
func (h *AuditHandler) VerifyChainIntegrity(c *gin.Context) {
	var req VerifyChainRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	if req.FromSequence < 0 || req.ToSequence < req.FromSequence {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sequence range"})
		return
	}

	startTime := time.Now()
	isValid, issues, err := h.auditService.VerifyAuditChain(
		c.Request.Context(),
		req.FromSequence,
		req.ToSequence,
	)

	if err != nil {
		slog.ErrorContext(c.Request.Context(), "Failed to verify chain", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify chain"})
		return
	}

	verificationTimeMs := time.Since(startTime).Milliseconds()

	c.JSON(http.StatusOK, gin.H{
		"is_valid":            isValid,
		"issues":              issues,
		"verified_entries":    req.ToSequence - req.FromSequence + 1,
		"verification_time_ms": verificationTimeMs,
	})
}

// ExportAuditLogsRequest represents an export request.
type ExportAuditLogsRequest struct {
	Format   string     `form:"format"`
	FromDate *time.Time `form:"from_date"`
	ToDate   *time.Time `form:"to_date"`
}

// ExportAuditLogs exports audit logs as CSV or JSON.
// GET /api/audit/export
func (h *AuditHandler) ExportAuditLogs(c *gin.Context) {
	var req ExportAuditLogsRequest

	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	if req.Format == "" {
		req.Format = "csv"
	}

	fromDate := time.Now().AddDate(0, 0, -30) // Default: last 30 days
	if req.FromDate != nil {
		fromDate = *req.FromDate
	}

	toDate := time.Now()
	if req.ToDate != nil {
		toDate = *req.ToDate
	}

	events, err := h.auditService.QueryAuditLogs(
		c.Request.Context(),
		nil,
		nil,
		nil,
		&fromDate,
		&toDate,
		10000,
	)

	if err != nil {
		slog.ErrorContext(c.Request.Context(), "Failed to export audit logs", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to export logs"})
		return
	}

	if req.Format == "csv" {
		c.Header("Content-Type", "text/csv")
		c.Header("Content-Disposition", 
			fmt.Sprintf("attachment; filename=\"audit_%s.csv\"", time.Now().Format("2006-01-02")))

		writer := csv.NewWriter(c.Writer)
		defer writer.Flush()

		// Write header
		headers := []string{
			"event_id", "timestamp", "event_type", "user_id", "severity",
			"resource_type", "resource_id", "pii_detected", "legal_hold",
		}
		if err := writer.Write(headers); err != nil {
			slog.ErrorContext(c.Request.Context(), "Failed to write CSV header", "error", err)
			return
		}

		// Write events
		for _, event := range events {
			userID := ""
			if event.UserID != nil {
				userID = event.UserID.String()
			}
			resType := ""
			if event.ResourceType != nil {
				resType = *event.ResourceType
			}
			resID := ""
			if event.ResourceID != nil {
				resID = event.ResourceID.String()
			}

			row := []string{
				event.EventID.String(),
				event.Timestamp.Format(time.RFC3339),
				event.EventType,
				userID,
				event.Severity,
				resType,
				resID,
				strconv.FormatBool(event.PIIDetected),
				strconv.FormatBool(event.LegalHold),
			}
			if err := writer.Write(row); err != nil {
				slog.ErrorContext(c.Request.Context(), "Failed to write CSV row", "error", err)
				return
			}
		}
		return
	}

	// JSON export
	c.JSON(http.StatusOK, gin.H{
		"events":      events,
		"total_count": len(events),
		"exported_at": time.Now().UTC().Format(time.RFC3339),
	})
}

// RetentionRequest represents a retention policy request.
type RetentionRequest struct {
	EventIDs []uuid.UUID `json:"event_ids"`
	Action   string      `json:"action"` // "legal_hold_apply" or "legal_hold_lift"
	Reason   string      `json:"reason"`
}

// UpdateRetention applies or lifts legal hold on events.
// PUT /api/audit/retention
func (h *AuditHandler) UpdateRetention(c *gin.Context) {
	var req RetentionRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	if len(req.EventIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "At least one event_id required"})
		return
	}

	var err error
	if req.Action == "legal_hold_apply" {
		err = h.auditService.ApplyLegalHold(c.Request.Context(), req.EventIDs, req.Reason)
	} else if req.Action == "legal_hold_lift" {
		err = h.auditService.LiftLegalHold(c.Request.Context(), req.EventIDs, req.Reason)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid action"})
		return
	}

	if err != nil {
		slog.ErrorContext(c.Request.Context(), "Failed to update retention", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update retention"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"updated_events":    len(req.EventIDs),
		"action":            req.Action,
		"legal_hold_applied": req.Action == "legal_hold_apply",
	})
}

// GetComplianceReport retrieves a compliance summary report.
// GET /api/audit/compliance-report
func (h *AuditHandler) GetComplianceReport(c *gin.Context) {
	fromDateStr := c.DefaultQuery("from_date", time.Now().AddDate(-1, 0, 0).Format("2006-01-02"))
	toDateStr := c.DefaultQuery("to_date", time.Now().Format("2006-01-02"))

	fromDate, err := time.Parse("2006-01-02", fromDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid from_date format"})
		return
	}

	toDate, err := time.Parse("2006-01-02", toDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid to_date format"})
		return
	}

	report, err := h.auditService.GetComplianceReport(c.Request.Context(), fromDate, toDate)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "Failed to generate compliance report", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate report"})
		return
	}

	c.JSON(http.StatusOK, report)
}

// PurgeExpiredEvents triggers purge of expired audit events.
// POST /api/audit/purge-expired
func (h *AuditHandler) PurgeExpiredEvents(c *gin.Context) {
	purgedCount, err := h.auditService.PurgeExpiredEvents(c.Request.Context())
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "Failed to purge expired events", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to purge events"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"purged_count": purgedCount,
		"timestamp":    time.Now().UTC().Format(time.RFC3339),
	})
}
