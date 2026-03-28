package handlers

import (
	"crypto/hmac"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/rhl/businessos-backend/internal/models"
)

// A2ARoutesHandler handles A2A protocol routes with audit trail logging.
// Supports PROV-O compliant audit chain signatures for governance tiers.
type A2ARoutesHandler struct {
	// AuditService handles hash-chain audit logging
	AuditService interface {
		LogA2ACall(agent, action, resourceType, resourceID string, snScore float64) (*models.AuditEntry, error)
		QueryAuditTrail(resourceType, resourceID string) ([]*models.AuditEntry, error)
	}
}

// NewA2ARoutesHandler creates a new A2A routes handler
func NewA2ARoutesHandler(auditService interface {
	LogA2ACall(agent, action, resourceType, resourceID string, snScore float64) (*models.AuditEntry, error)
	QueryAuditTrail(resourceType, resourceID string) ([]*models.AuditEntry, error)
}) *A2ARoutesHandler {
	return &A2ARoutesHandler{
		AuditService: auditService,
	}
}

// ============================================================================
// Request/Response Types
// ============================================================================

type DealCreateRequest struct {
	Name  string                 `json:"name" binding:"required"`
	Value float64                `json:"value,omitempty"`
	Extra map[string]interface{} `json:"extra,omitempty"`
}

type DealCreateResponse struct {
	Deal       map[string]interface{} `json:"deal"`
	AuditEntry *models.AuditEntry     `json:"audit_entry"`
}

type LeadUpdateRequest struct {
	LeadID string                 `json:"lead_id" binding:"required"`
	Status string                 `json:"status"`
	Extra  map[string]interface{} `json:"extra,omitempty"`
}

type LeadUpdateResponse struct {
	Lead       map[string]interface{} `json:"lead"`
	AuditEntry *models.AuditEntry     `json:"audit_entry"`
}

type TaskAssignRequest struct {
	Title    string                 `json:"title" binding:"required"`
	Assignee string                 `json:"assignee"`
	Extra    map[string]interface{} `json:"extra,omitempty"`
}

type TaskAssignResponse struct {
	Task       map[string]interface{} `json:"task"`
	AuditEntry *models.AuditEntry     `json:"audit_entry"`
}

type ProgressUpdateRequest struct {
	ProjectID string                 `json:"project_id" binding:"required"`
	Status    string                 `json:"status"`
	Percent   int                    `json:"percent"`
	Extra     map[string]interface{} `json:"extra,omitempty"`
}

type ProgressUpdateResponse struct {
	Status     string             `json:"status"`
	AuditEntry *models.AuditEntry `json:"audit_entry"`
}

type AuditQueryResponse struct {
	Entries []*models.AuditEntry `json:"entries"`
	Count   int                  `json:"count"`
}

// ============================================================================
// Middleware: Shared Secret Authentication
// ============================================================================

// requireSharedSecret middleware checks X-Shared-Secret header
func (h *A2ARoutesHandler) requireSharedSecret() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := h.checkAuth(c); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}
		c.Next()
	}
}

// ============================================================================
// Handler: Create Deal
// ============================================================================

// CreateDeal handles POST /api/integrations/a2a/crm/deals
func (h *A2ARoutesHandler) CreateDeal(c *gin.Context) {
	// Check auth
	if err := h.checkAuth(c); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var req DealCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	agentID := c.GetHeader("X-Agent-ID")
	if agentID == "" {
		agentID = "unknown-agent"
	}

	dealID := uuid.New().String()
	snScore := 0.9 // Default high confidence

	// Log to audit trail
	auditEntry, err := h.AuditService.LogA2ACall(
		agentID,
		"create",
		"deal",
		dealID,
		snScore,
	)
	if err != nil {
		slog.Error("failed to create audit entry", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "audit logging failed"})
		return
	}

	deal := map[string]interface{}{
		"id":   dealID,
		"name": req.Name,
	}
	if req.Value > 0 {
		deal["value"] = req.Value
	}
	if len(req.Extra) > 0 {
		for k, v := range req.Extra {
			deal[k] = v
		}
	}

	c.JSON(http.StatusCreated, DealCreateResponse{
		Deal:       deal,
		AuditEntry: auditEntry,
	})
}

// ============================================================================
// Handler: Update Lead
// ============================================================================

// UpdateLead handles POST /api/integrations/a2a/crm/leads
func (h *A2ARoutesHandler) UpdateLead(c *gin.Context) {
	if err := h.checkAuth(c); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var req LeadUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	agentID := c.GetHeader("X-Agent-ID")
	if agentID == "" {
		agentID = "unknown-agent"
	}

	snScore := 0.85

	auditEntry, err := h.AuditService.LogA2ACall(
		agentID,
		"update",
		"lead",
		req.LeadID,
		snScore,
	)
	if err != nil {
		slog.Error("failed to create audit entry", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "audit logging failed"})
		return
	}

	lead := map[string]interface{}{
		"id": req.LeadID,
	}
	if req.Status != "" {
		lead["status"] = req.Status
	}
	if len(req.Extra) > 0 {
		for k, v := range req.Extra {
			lead[k] = v
		}
	}

	c.JSON(http.StatusOK, LeadUpdateResponse{
		Lead:       lead,
		AuditEntry: auditEntry,
	})
}

// ============================================================================
// Handler: Assign Task
// ============================================================================

// AssignTask handles POST /api/integrations/a2a/projects/tasks
func (h *A2ARoutesHandler) AssignTask(c *gin.Context) {
	if err := h.checkAuth(c); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var req TaskAssignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	agentID := c.GetHeader("X-Agent-ID")
	if agentID == "" {
		agentID = "unknown-agent"
	}

	taskID := uuid.New().String()
	snScore := 0.9

	auditEntry, err := h.AuditService.LogA2ACall(
		agentID,
		"assign",
		"task",
		taskID,
		snScore,
	)
	if err != nil {
		slog.Error("failed to create audit entry", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "audit logging failed"})
		return
	}

	task := map[string]interface{}{
		"id":    taskID,
		"title": req.Title,
	}
	if req.Assignee != "" {
		task["assignee"] = req.Assignee
	}
	if len(req.Extra) > 0 {
		for k, v := range req.Extra {
			task[k] = v
		}
	}

	c.JSON(http.StatusCreated, TaskAssignResponse{
		Task:       task,
		AuditEntry: auditEntry,
	})
}

// ============================================================================
// Handler: Update Progress
// ============================================================================

// UpdateProgress handles POST /api/integrations/a2a/projects/progress
func (h *A2ARoutesHandler) UpdateProgress(c *gin.Context) {
	if err := h.checkAuth(c); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var req ProgressUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	agentID := c.GetHeader("X-Agent-ID")
	if agentID == "" {
		agentID = "unknown-agent"
	}

	snScore := 0.88

	auditEntry, err := h.AuditService.LogA2ACall(
		agentID,
		"update_progress",
		"project",
		req.ProjectID,
		snScore,
	)
	if err != nil {
		slog.Error("failed to create audit entry", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "audit logging failed"})
		return
	}

	c.JSON(http.StatusOK, ProgressUpdateResponse{
		Status:     "updated",
		AuditEntry: auditEntry,
	})
}

// ============================================================================
// Handler: Query Audit Trail
// ============================================================================

// QueryAudit handles GET /api/integrations/a2a/audit/query
func (h *A2ARoutesHandler) QueryAudit(c *gin.Context) {
	if err := h.checkAuth(c); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	resourceType := c.Query("resource_type")
	resourceID := c.Query("resource_id")

	if resourceType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "resource_type required"})
		return
	}

	entries, err := h.AuditService.QueryAuditTrail(resourceType, resourceID)
	if err != nil {
		slog.Error("failed to query audit trail", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "audit query failed"})
		return
	}

	if entries == nil {
		entries = make([]*models.AuditEntry, 0)
	}

	c.JSON(http.StatusOK, AuditQueryResponse{
		Entries: entries,
		Count:   len(entries),
	})
}

// ============================================================================
// Helper Methods
// ============================================================================

// checkAuth verifies shared secret authentication using constant-time comparison.
// Reads expected secret from BOS_A2A_SHARED_SECRET env var.
// In development (env var unset), logs a warning and allows any non-empty secret.
func (h *A2ARoutesHandler) checkAuth(c *gin.Context) error {
	secret := c.GetHeader("X-Shared-Secret")
	if secret == "" {
		return fmt.Errorf("X-Shared-Secret header required")
	}
	expected := os.Getenv("BOS_A2A_SHARED_SECRET")
	if expected == "" {
		// No secret configured — allow in dev mode, warn loudly
		slog.Warn("BOS_A2A_SHARED_SECRET not set; skipping secret validation (dev mode only)")
		return nil
	}
	// Constant-time comparison to prevent timing attacks
	if !hmac.Equal([]byte(secret), []byte(expected)) {
		return fmt.Errorf("invalid shared secret")
	}
	return nil
}

// ErrMissingSecret is the error message for missing secret
const ErrMissingSecret = "X-Shared-Secret header required"
