package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rhl/businessos-backend/internal/services"
	"log/slog"
)

func TestComplianceHandler_GetComplianceStatus(t *testing.T) {
	logger := slog.Default()
	complianceSvc := services.NewComplianceService("http://localhost:9999", logger)

	handler := NewComplianceHandler(complianceSvc, logger)

	r := gin.New()
	r.GET("/api/compliance/status", handler.GetComplianceStatus)

	req := httptest.NewRequest(http.MethodGet, "/api/compliance/status", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Service tries OSA (which won't be running), falls back to cached status
	assert.Equal(t, http.StatusOK, w.Code)

	var status services.ComplianceStatus
	err := json.Unmarshal(w.Body.Bytes(), &status)
	require.NoError(t, err)
	assert.Contains(t, status.Domains, "data_security")
	assert.Contains(t, status.Domains, "process_integrity")
	assert.Contains(t, status.Domains, "regulatory")
}

func TestComplianceHandler_GetAuditTrail_MissingSessionID(t *testing.T) {
	logger := slog.Default()
	complianceSvc := services.NewComplianceService("http://localhost:9999", logger)
	handler := NewComplianceHandler(complianceSvc, logger)

	r := gin.New()
	r.GET("/api/compliance/audit-trail", handler.GetAuditTrail)

	req := httptest.NewRequest(http.MethodGet, "/api/compliance/audit-trail", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestComplianceHandler_GetAuditTrail_WithSessionID(t *testing.T) {
	logger := slog.Default()
	complianceSvc := services.NewComplianceService("http://localhost:9999", logger)
	handler := NewComplianceHandler(complianceSvc, logger)

	r := gin.New()
	r.GET("/api/compliance/audit-trail", handler.GetAuditTrail)

	req := httptest.NewRequest(http.MethodGet, "/api/compliance/audit-trail?session_id=test-session", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Service tries OSA (fails), returns empty result
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestComplianceHandler_GetAuditTrail_WithLimit(t *testing.T) {
	logger := slog.Default()
	complianceSvc := services.NewComplianceService("http://localhost:9999", logger)
	handler := NewComplianceHandler(complianceSvc, logger)

	r := gin.New()
	r.GET("/api/compliance/audit-trail", handler.GetAuditTrail)

	req := httptest.NewRequest(http.MethodGet, "/api/compliance/audit-trail?session_id=sess-1&limit=10&offset=5", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var result services.AuditTrailResponse
	err := json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, 10, result.Limit)
	assert.Equal(t, 5, result.Offset)
}

func TestComplianceHandler_VerifyAuditChain_EmptySession(t *testing.T) {
	logger := slog.Default()
	complianceSvc := services.NewComplianceService("http://localhost:9999", logger)
	handler := NewComplianceHandler(complianceSvc, logger)

	r := gin.New()
	r.GET("/api/compliance/audit-trail/verify/:session_id", handler.VerifyAuditChain)

	req := httptest.NewRequest(http.MethodGet, "/api/compliance/audit-trail/verify/empty-session", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var result services.VerifyResult
	err := json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err)
	// Empty audit trail is considered verified
	assert.True(t, result.Verified)
}

func TestComplianceHandler_CollectEvidence_MissingBody(t *testing.T) {
	logger := slog.Default()
	complianceSvc := services.NewComplianceService("http://localhost:9999", logger)
	handler := NewComplianceHandler(complianceSvc, logger)

	r := gin.New()
	r.POST("/api/compliance/evidence/collect", handler.CollectEvidence)

	req := httptest.NewRequest(http.MethodPost, "/api/compliance/evidence/collect", bytes.NewReader([]byte("{}")))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestComplianceHandler_CollectEvidence_Success(t *testing.T) {
	logger := slog.Default()
	complianceSvc := services.NewComplianceService("http://localhost:9999", logger)
	handler := NewComplianceHandler(complianceSvc, logger)

	r := gin.New()
	r.POST("/api/compliance/evidence/collect", handler.CollectEvidence)

	body, _ := json.Marshal(services.EvidenceCollectRequest{
		Domain: "data_security",
		Period: "2026-Q1",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/compliance/evidence/collect", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var result services.EvidenceCollectResponse
	err := json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, "data_security", result.Domain)
	assert.Equal(t, "2026-Q1", result.Period)
	assert.GreaterOrEqual(t, result.Collected, 2) // At least domain-specific evidence
}

func TestComplianceHandler_GetGapAnalysis_Default(t *testing.T) {
	logger := slog.Default()
	complianceSvc := services.NewComplianceService("http://localhost:9999", logger)
	handler := NewComplianceHandler(complianceSvc, logger)

	r := gin.New()
	r.GET("/api/compliance/gap-analysis", handler.GetGapAnalysis)

	req := httptest.NewRequest(http.MethodGet, "/api/compliance/gap-analysis", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var result services.GapAnalysisResponse
	err := json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, "SOC2", result.Framework) // Defaults to SOC2
	assert.NotEmpty(t, result.Gaps)
}

func TestComplianceHandler_GetGapAnalysis_HIPAA(t *testing.T) {
	logger := slog.Default()
	complianceSvc := services.NewComplianceService("http://localhost:9999", logger)
	handler := NewComplianceHandler(complianceSvc, logger)

	r := gin.New()
	r.GET("/api/compliance/gap-analysis", handler.GetGapAnalysis)

	req := httptest.NewRequest(http.MethodGet, "/api/compliance/gap-analysis?framework=HIPAA", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var result services.GapAnalysisResponse
	err := json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, "HIPAA", result.Framework)
	assert.Greater(t, len(result.Gaps), 0)
}

func TestComplianceHandler_CreateRemediation_Success(t *testing.T) {
	logger := slog.Default()
	complianceSvc := services.NewComplianceService("http://localhost:9999", logger)
	handler := NewComplianceHandler(complianceSvc, logger)

	r := gin.New()
	r.POST("/api/compliance/remediation", handler.CreateRemediation)

	body, _ := json.Marshal(services.RemediationRequest{
		GapID:    "soc2-cc6.1",
		Priority: "high",
		Assignee: "security-team",
		DueDate:  "2026-04-01",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/compliance/remediation", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var task services.RemediationTask
	err := json.Unmarshal(w.Body.Bytes(), &task)
	require.NoError(t, err)
	assert.Equal(t, "soc2-cc6.1", task.GapID)
	assert.Equal(t, "high", task.Priority)
	assert.Equal(t, "security-team", task.Assignee)
	assert.Equal(t, "open", task.Status)
}

func TestComplianceHandler_CreateRemediation_MissingBody(t *testing.T) {
	logger := slog.Default()
	complianceSvc := services.NewComplianceService("http://localhost:9999", logger)
	handler := NewComplianceHandler(complianceSvc, logger)

	r := gin.New()
	r.POST("/api/compliance/remediation", handler.CreateRemediation)

	req := httptest.NewRequest(http.MethodPost, "/api/compliance/remediation", bytes.NewReader([]byte("{}")))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
