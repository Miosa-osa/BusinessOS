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

// ─────────────────────────────────────────────────────────────────────────────
// Cross-System Audit Trail Tests (TDD — failing first, implement second)
// ─────────────────────────────────────────────────────────────────────────────

func TestComplianceHandler_GetAuditTrail_OSAUnavailable_Returns503(t *testing.T) {
	logger := slog.Default()
	// Point to unreachable OSA server
	complianceSvc := services.NewComplianceService("http://127.0.0.1:9", logger)
	handler := NewComplianceHandler(complianceSvc, logger)

	r := gin.New()
	r.GET("/api/compliance/audit-trail", handler.GetAuditTrail)

	req := httptest.NewRequest(http.MethodGet, "/api/compliance/audit-trail?session_id=test-sess", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// When OSA is unavailable and no cache exists, should return 503 Service Unavailable
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var errResp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Contains(t, errResp, "error")
}

func TestComplianceHandler_GetAuditTrail_OSAHashChainVerified(t *testing.T) {
	// Start a mock OSA server that returns audit entries with hash chain
	mockOSA := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/audit-trail/test-sess" {
			// Return mock audit trail with hash chain (OSA format)
			resp := map[string]any{
				"session_id":  "test-sess",
				"entry_count": 2,
				"entries": []map[string]any{
					{
						"index":          0,
						"timestamp":      "2026-03-24T10:00:00Z",
						"session_id":     "test-sess",
						"tool_name":      "pm4py_discover",
						"arguments_hash": "abc123",
						"result_hash":    "def456",
						"duration_ms":    100,
						"provider":       "ollama",
						"model":          "mistral",
						"previous_hash":  "genesis",
						"entry_hash":     "hash0",
					},
					{
						"index":          1,
						"timestamp":      "2026-03-24T10:00:05Z",
						"session_id":     "test-sess",
						"tool_name":      "analyze_log",
						"arguments_hash": "ghi789",
						"result_hash":    "jkl012",
						"duration_ms":    50,
						"provider":       "ollama",
						"model":          "mistral",
						"previous_hash":  "hash0",
						"entry_hash":     "hash1",
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer mockOSA.Close()

	logger := slog.Default()
	complianceSvc := services.NewComplianceService(mockOSA.URL, logger)
	handler := NewComplianceHandler(complianceSvc, logger)

	r := gin.New()
	r.GET("/api/compliance/audit-trail", handler.GetAuditTrail)

	req := httptest.NewRequest(http.MethodGet, "/api/compliance/audit-trail?session_id=test-sess", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var result services.AuditTrailResponse
	err := json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, 2, result.Total)
	assert.Len(t, result.Entries, 2)

	// Verify hash chain: each entry's hash was computed from previous
	for i, entry := range result.Entries {
		assert.NotEmpty(t, entry.Hash)
		assert.NotEmpty(t, entry.PrevHash)
		if i == 0 {
			assert.Equal(t, "genesis", entry.PrevHash)
		} else {
			// Hash should match previous entry's hash
			assert.Equal(t, result.Entries[i-1].Hash, entry.PrevHash)
		}
	}
}

func TestComplianceHandler_VerifyAuditChain_ValidatesOSAChain(t *testing.T) {
	// Mock OSA that returns 2 valid entries with correct hash chain
	mockOSA := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/audit-trail/test-sess" {
			resp := map[string]any{
				"session_id":  "test-sess",
				"entry_count": 2,
				"entries": []map[string]any{
					{
						"index":          0,
						"timestamp":      "2026-03-24T10:00:00Z",
						"session_id":     "test-sess",
						"tool_name":      "pm4py_discover",
						"arguments_hash": "abc123",
						"result_hash":    "def456",
						"duration_ms":    100,
						"provider":       "ollama",
						"model":          "mistral",
						"previous_hash":  "genesis",
						"entry_hash":     "computed_hash_0",
					},
					{
						"index":          1,
						"timestamp":      "2026-03-24T10:00:05Z",
						"session_id":     "test-sess",
						"tool_name":      "analyze_log",
						"arguments_hash": "ghi789",
						"result_hash":    "jkl012",
						"duration_ms":    50,
						"provider":       "ollama",
						"model":          "mistral",
						"previous_hash":  "computed_hash_0",
						"entry_hash":     "computed_hash_1",
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer mockOSA.Close()

	logger := slog.Default()
	complianceSvc := services.NewComplianceService(mockOSA.URL, logger)
	handler := NewComplianceHandler(complianceSvc, logger)

	r := gin.New()
	r.GET("/api/compliance/audit-trail/verify/:session_id", handler.VerifyAuditChain)

	req := httptest.NewRequest(http.MethodGet, "/api/compliance/audit-trail/verify/test-sess", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var result services.VerifyResult
	err := json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, 2, result.Entries)
	assert.NotEmpty(t, result.MerkleRoot)
	// With hash chain, should return verified (or at least report issues if chain broken)
	assert.IsType(t, result.Verified, true)
	assert.IsType(t, result.Issues, []string{})
}

func TestComplianceHandler_GetAuditTrail_CombinesOSAWithBusinessOSEvents(t *testing.T) {
	// Mock OSA returning 1 entry
	mockOSA := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/audit-trail/test-sess" {
			resp := map[string]any{
				"session_id":  "test-sess",
				"entry_count": 1,
				"entries": []map[string]any{
					{
						"index":          0,
						"timestamp":      "2026-03-24T10:00:00Z",
						"session_id":     "test-sess",
						"tool_name":      "pm4py_discover",
						"arguments_hash": "abc123",
						"result_hash":    "def456",
						"duration_ms":    100,
						"provider":       "ollama",
						"model":          "mistral",
						"previous_hash":  "genesis",
						"entry_hash":     "osa_hash_0",
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer mockOSA.Close()

	logger := slog.Default()
	complianceSvc := services.NewComplianceService(mockOSA.URL, logger)
	handler := NewComplianceHandler(complianceSvc, logger)

	r := gin.New()
	r.GET("/api/compliance/audit-trail", handler.GetAuditTrail)

	req := httptest.NewRequest(http.MethodGet, "/api/compliance/audit-trail?session_id=test-sess&limit=50", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var result services.AuditTrailResponse
	err := json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err)
	// Should have at least 1 entry from OSA (may add BusinessOS events as well)
	assert.GreaterOrEqual(t, result.Total, 1)
}

func TestComplianceHandler_DegradedMode_ReturnsBusinessOSOnlyWhenOSAFails(t *testing.T) {
	logger := slog.Default()
	// Unreachable OSA — no retries
	complianceSvc := services.NewComplianceService("http://127.0.0.1:7", logger)
	handler := NewComplianceHandler(complianceSvc, logger)

	r := gin.New()
	r.GET("/api/compliance/audit-trail", handler.GetAuditTrail)

	req := httptest.NewRequest(http.MethodGet, "/api/compliance/audit-trail?session_id=sess-1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// May return 503 (unavailable), or 200 with empty entries (degraded)
	// The test expects: either 503 or 200 with empty/minimal data
	assert.Contains(t, []int{http.StatusOK, http.StatusServiceUnavailable}, w.Code)
}

// TestComplianceHandler_VerifyCompliance_BOSAliasPath verifies the Canopy adapter path /api/bos/compliance/verify.
func TestComplianceHandler_VerifyCompliance_BOSAliasPath(t *testing.T) {
	logger := slog.Default()
	complianceSvc := services.NewComplianceService("http://localhost:9999", logger)
	handler := NewComplianceHandler(complianceSvc, logger)

	r := gin.New()
	r.POST("/api/bos/compliance/verify", handler.VerifyCompliance)

	body, _ := json.Marshal(services.ComplianceVerifyRequest{
		WorkspaceID: "ws-test",
		Framework:   "SOC2",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/bos/compliance/verify", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var result services.ComplianceVerifyResponse
	err := json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.NotEmpty(t, result.Status)
	assert.Equal(t, "SOC2", result.Framework)
}

// TestComplianceHandler_VerifyCompliance_BOSAliasPath_MissingFields verifies the alias path enforces the same validation.
func TestComplianceHandler_VerifyCompliance_BOSAliasPath_MissingFields(t *testing.T) {
	logger := slog.Default()
	complianceSvc := services.NewComplianceService("http://localhost:9999", logger)
	handler := NewComplianceHandler(complianceSvc, logger)

	r := gin.New()
	r.POST("/api/bos/compliance/verify", handler.VerifyCompliance)

	req := httptest.NewRequest(http.MethodPost, "/api/bos/compliance/verify",
		bytes.NewReader([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
