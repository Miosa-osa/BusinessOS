package services

import (
	"encoding/json"
	"log/slog"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// NewComplianceService tests
// ---------------------------------------------------------------------------

func TestNewComplianceService_DefaultState(t *testing.T) {
	logger := slog.Default()
	svc := NewComplianceService("http://localhost:9999", logger)
	assert.NotNil(t, svc)
	assert.Equal(t, 0.0, svc.status.OverallScore)
	assert.Len(t, svc.status.Domains, 3)
	assert.Empty(t, svc.status.Certificates)
}

func TestNewComplianceService_DomainsInitialized(t *testing.T) {
	logger := slog.Default()
	svc := NewComplianceService("http://localhost:9999", logger)

	for _, domain := range []string{"data_security", "process_integrity", "regulatory"} {
		d, ok := svc.status.Domains[domain]
		require.True(t, ok, "domain %s should exist", domain)
		assert.Equal(t, 0.0, d.Score)
		assert.Equal(t, 0, d.ChecksPassed)
		assert.Equal(t, 0, d.ChecksFailed)
	}
}

// ---------------------------------------------------------------------------
// GetStatus tests (without OSA running, returns cached)
// ---------------------------------------------------------------------------

func TestComplianceService_GetStatus_Cached(t *testing.T) {
	logger := slog.Default()
	svc := NewComplianceService("http://localhost:9999", logger)

	// Manually set a status to test cache behavior
	svc.mu.Lock()
	svc.status.OverallScore = 85.5
	svc.lastRefresh = time.Now()
	svc.mu.Unlock()

	status, err := svc.GetStatus(nil)
	require.NoError(t, err)
	assert.Equal(t, 85.5, status.OverallScore)
}

// ---------------------------------------------------------------------------
// VerifyAuditChain tests
// ---------------------------------------------------------------------------

func TestComplianceService_VerifyAuditChain_EmptyTrail(t *testing.T) {
	logger := slog.Default()
	svc := NewComplianceService("http://localhost:9999", logger)

	// Seed empty cache
	svc.mu.Lock()
	svc.auditCache["empty-session"] = []AuditEntry{}
	svc.mu.Unlock()

	result, err := svc.VerifyAuditChain(nil, "empty-session")
	require.NoError(t, err)
	assert.True(t, result.Verified)
	assert.Equal(t, 0, result.Entries)
	assert.Empty(t, result.Issues)
}

func TestComplianceService_VerifyAuditChain_CorruptedHash(t *testing.T) {
	logger := slog.Default()
	svc := NewComplianceService("http://localhost:9999", logger)

	entries := []AuditEntry{
		{
			ID:        "audit-1",
			SessionID: "sess-corrupt",
			Timestamp: time.Date(2026, 3, 15, 10, 0, 0, 0, time.UTC),
			Action:    "tool_call",
			Actor:     "agent-1",
			Hash:      "tampered_hash_00000000000000000000000000000000000000000000000000000000000001",
			PrevHash:  "",
		},
	}

	svc.mu.Lock()
	svc.auditCache["sess-corrupt"] = entries
	svc.mu.Unlock()

	result, err := svc.VerifyAuditChain(nil, "sess-corrupt")
	require.NoError(t, err)
	assert.False(t, result.Verified)
	assert.NotEmpty(t, result.Issues)
}

// ---------------------------------------------------------------------------
// CollectEvidence tests
// ---------------------------------------------------------------------------

func TestComplianceService_CollectEvidence_DataSecurity(t *testing.T) {
	logger := slog.Default()
	svc := NewComplianceService("http://localhost:9999", logger)

	result, err := svc.CollectEvidence(nil, EvidenceCollectRequest{
		Domain: "data_security",
		Period: "2026-Q1",
	})
	require.NoError(t, err)
	assert.Equal(t, "data_security", result.Domain)
	assert.Equal(t, "2026-Q1", result.Period)
	// data_security domain generates 2 synthetic evidence items
	assert.GreaterOrEqual(t, result.Collected, 2)
}

func TestComplianceService_CollectEvidence_ProcessIntegrity(t *testing.T) {
	logger := slog.Default()
	svc := NewComplianceService("http://localhost:9999", logger)

	result, err := svc.CollectEvidence(nil, EvidenceCollectRequest{
		Domain: "process_integrity",
		Period: "2026-Q2",
	})
	require.NoError(t, err)
	assert.Equal(t, "process_integrity", result.Domain)
	assert.GreaterOrEqual(t, result.Collected, 1)
}

func TestComplianceService_CollectEvidence_Regulatory(t *testing.T) {
	logger := slog.Default()
	svc := NewComplianceService("http://localhost:9999", logger)

	result, err := svc.CollectEvidence(nil, EvidenceCollectRequest{
		Domain: "regulatory",
		Period: "2026-Q3",
	})
	require.NoError(t, err)
	assert.Equal(t, "regulatory", result.Domain)
	assert.GreaterOrEqual(t, result.Collected, 1)
}

func TestComplianceService_CollectEvidence_UnknownDomain(t *testing.T) {
	logger := slog.Default()
	svc := NewComplianceService("http://localhost:9999", logger)

	result, err := svc.CollectEvidence(nil, EvidenceCollectRequest{
		Domain: "unknown_domain",
		Period: "2026-Q4",
	})
	require.NoError(t, err)
	// Unknown domain has no audit entries and no synthetic evidence
	assert.Equal(t, 0, result.Collected)
}

// ---------------------------------------------------------------------------
// GetGapAnalysis tests
// ---------------------------------------------------------------------------

func TestComplianceService_GetGapAnalysis_DefaultSOC2(t *testing.T) {
	logger := slog.Default()
	svc := NewComplianceService("http://localhost:9999", logger)

	result, err := svc.GetGapAnalysis(nil, "")
	require.NoError(t, err)
	assert.Equal(t, "SOC2", result.Framework)
	assert.Greater(t, len(result.Gaps), 0)
	assert.Greater(t, result.Score, 0.0)
	assert.LessOrEqual(t, result.Score, 1.0)
}

func TestComplianceService_GetGapAnalysis_AllFrameworks(t *testing.T) {
	logger := slog.Default()
	svc := NewComplianceService("http://localhost:9999", logger)

	frameworks := []string{"SOC2", "HIPAA", "GDPR", "SOX"}
	for _, fw := range frameworks {
		result, err := svc.GetGapAnalysis(nil, fw)
		require.NoError(t, err, "framework %s should succeed", fw)
		assert.Equal(t, fw, result.Framework)
		assert.GreaterOrEqual(t, result.Score, 0.0)
		assert.LessOrEqual(t, result.Score, 1.0)
		assert.False(t, result.AnalyzedAt.IsZero())
	}
}

func TestComplianceService_GetGapAnalysis_UnknownFramework(t *testing.T) {
	logger := slog.Default()
	svc := NewComplianceService("http://localhost:9999", logger)

	result, err := svc.GetGapAnalysis(nil, "UNKNOWN_FRAMEWORK")
	require.NoError(t, err)
	assert.Equal(t, "UNKNOWN_FRAMEWORK", result.Framework)
	assert.Empty(t, result.Gaps)
	assert.Equal(t, 1.0, result.Score) // No gaps = perfect score
}

// ---------------------------------------------------------------------------
// CreateRemediation tests
// ---------------------------------------------------------------------------

func TestComplianceService_CreateRemediation(t *testing.T) {
	logger := slog.Default()
	svc := NewComplianceService("http://localhost:9999", logger)

	task, err := svc.CreateRemediation(nil, RemediationRequest{
		GapID:    "soc2-cc6.1",
		Priority: "high",
		Assignee: "security-team",
		DueDate:  "2026-04-01",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, task.ID)
	assert.Equal(t, "soc2-cc6.1", task.GapID)
	assert.Equal(t, "high", task.Priority)
	assert.Equal(t, "security-team", task.Assignee)
	assert.Equal(t, "2026-04-01", task.DueDate)
	assert.Equal(t, "open", task.Status)
	assert.False(t, task.CreatedAt.IsZero())
}

func TestComplianceService_CreateRemediation_UniqueIDs(t *testing.T) {
	logger := slog.Default()
	svc := NewComplianceService("http://localhost:9999", logger)

	task1, _ := svc.CreateRemediation(nil, RemediationRequest{
		GapID: "gap-1", Priority: "low", Assignee: "a", DueDate: "2026-05-01",
	})
	task2, _ := svc.CreateRemediation(nil, RemediationRequest{
		GapID: "gap-2", Priority: "high", Assignee: "b", DueDate: "2026-06-01",
	})

	assert.NotEqual(t, task1.ID, task2.ID)
}

// ---------------------------------------------------------------------------
// computeEntryHash tests
// ---------------------------------------------------------------------------

func TestComputeEntryHash_SameInput(t *testing.T) {
	entry := AuditEntry{
		SessionID: "sess-1",
		Timestamp: time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC),
		Action:    "tool_call",
		Actor:     "agent-1",
	}
	h1 := computeEntryHash(entry, "")
	h2 := computeEntryHash(entry, "")
	assert.Equal(t, h1, h2)
}

func TestComputeEntryHash_DifferentPrevHash(t *testing.T) {
	entry := AuditEntry{
		SessionID: "sess-1",
		Timestamp: time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC),
		Action:    "tool_call",
		Actor:     "agent-1",
	}
	h1 := computeEntryHash(entry, "")
	h2 := computeEntryHash(entry, "prev-hash")
	assert.NotEqual(t, h1, h2)
}

func TestComputeEntryHash_DifferentDetails(t *testing.T) {
	entry1 := AuditEntry{
		SessionID: "sess-1",
		Timestamp: time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC),
		Action:    "tool_call",
		Actor:     "agent-1",
	}
	entry2 := AuditEntry{
		SessionID: "sess-1",
		Timestamp: time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC),
		Action:    "tool_call",
		Actor:     "agent-1",
		Details:   map[string]any{"key": "value"},
	}
	h1 := computeEntryHash(entry1, "")
	h2 := computeEntryHash(entry2, "")
	assert.NotEqual(t, h1, h2)
}

// ---------------------------------------------------------------------------
// computeMerkleRoot tests
// ---------------------------------------------------------------------------

func TestComputeMerkleRoot_Empty(t *testing.T) {
	root := computeMerkleRoot(nil)
	assert.Empty(t, root)
}

func TestComputeMerkleRoot_Single(t *testing.T) {
	entries := []AuditEntry{
		{Hash: "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"},
	}
	root := computeMerkleRoot(entries)
	assert.Equal(t, "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2", root)
}

// ---------------------------------------------------------------------------
// parsePeriod tests
// ---------------------------------------------------------------------------

func TestParsePeriod_Quarter(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Time
	}{
		{"2026-Q1", time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"2026-Q2", time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)},
		{"2026-Q3", time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)},
		{"2026-Q4", time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := parsePeriod(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestParsePeriod_Month(t *testing.T) {
	result := parsePeriod("2026-06")
	expected := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, expected, result)
}

func TestParsePeriod_Invalid(t *testing.T) {
	before := time.Now().AddDate(0, -2, 0)
	result := parsePeriod("invalid")
	after := time.Now()
	assert.True(t, result.After(before) && result.Before(after),
		"invalid period should default to ~1 month ago, got %v", result)
}

// ---------------------------------------------------------------------------
// computeGapScore tests
// ---------------------------------------------------------------------------

func TestComputeGapScore_Empty(t *testing.T) {
	assert.Equal(t, 1.0, computeGapScore(nil))
}

func TestComputeGapScore_AllResolved(t *testing.T) {
	gaps := []ComplianceGap{
		{Severity: "critical", Status: "resolved"},
		{Severity: "high", Status: "resolved"},
		{Severity: "medium", Status: "resolved"},
		{Severity: "low", Status: "resolved"},
	}
	assert.Equal(t, 1.0, computeGapScore(gaps))
}

func TestComputeGapScore_Bounded(t *testing.T) {
	// Even many critical open gaps should not go below 0
	gaps := []ComplianceGap{}
	for i := 0; i < 100; i++ {
		gaps = append(gaps, ComplianceGap{Severity: "critical", Status: "open"})
	}
	score := computeGapScore(gaps)
	assert.GreaterOrEqual(t, score, 0.0)
	assert.LessOrEqual(t, score, 1.0)
}

func TestComputeGapScore_Mixed(t *testing.T) {
	gaps := []ComplianceGap{
		{Severity: "critical", Status: "open"},
		{Severity: "high", Status: "resolved"},
		{Severity: "low", Status: "open"},
	}
	score := computeGapScore(gaps)
	assert.True(t, score > 0.0 && score < 1.0, "score should be between 0 and 1, got %f", score)
}

// ---------------------------------------------------------------------------
// generateDomainEvidence tests
// ---------------------------------------------------------------------------

func TestGenerateDomainEvidence_DataSecurity(t *testing.T) {
	items := generateDomainEvidence("data_security", "2026-Q1")
	assert.Len(t, items, 2)
	for _, item := range items {
		assert.Equal(t, "data_security", item.Domain)
		assert.Equal(t, "2026-Q1", item.Period)
		assert.Equal(t, "policy_check", item.Type)
	}
}

func TestGenerateDomainEvidence_ProcessIntegrity(t *testing.T) {
	items := generateDomainEvidence("process_integrity", "2026-Q2")
	assert.Len(t, items, 1)
	assert.Equal(t, "process_check", items[0].Type)
}

func TestGenerateDomainEvidence_Regulatory(t *testing.T) {
	items := generateDomainEvidence("regulatory", "2026-Q3")
	assert.Len(t, items, 1)
	assert.Equal(t, "regulatory_check", items[0].Type)
}

func TestGenerateDomainEvidence_Unknown(t *testing.T) {
	items := generateDomainEvidence("nonexistent", "2026-Q4")
	assert.Empty(t, items)
}

// ---------------------------------------------------------------------------
// JSON serialization tests for types
// ---------------------------------------------------------------------------

func TestAuditEntry_MarshalJSON(t *testing.T) {
	entry := AuditEntry{
		ID:        "audit-001",
		SessionID: "sess-1",
		Timestamp: time.Date(2026, 3, 15, 10, 30, 0, 0, time.UTC),
		Action:    "tool_call",
		Actor:     "agent-1",
		ToolName:  "http_post",
		Hash:      "abc123",
		PrevHash:  "",
	}

	data, err := json.Marshal(entry)
	require.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, "audit-001", decoded["id"])
	assert.Equal(t, "tool_call", decoded["action"])
}

func TestComplianceGap_MarshalJSON(t *testing.T) {
	gap := ComplianceGap{
		ID:          "soc2-cc6.1",
		Framework:   "SOC2",
		Control:     "CC6.1",
		Description: "Missing documentation",
		Severity:    "medium",
		Status:      "open",
	}

	data, err := json.Marshal(gap)
	require.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, "soc2-cc6.1", decoded["id"])
	assert.Equal(t, "medium", decoded["severity"])
	assert.Equal(t, "open", decoded["status"])
}

func TestRemediationTask_MarshalJSON(t *testing.T) {
	task := RemediationTask{
		ID:        "rem-001",
		GapID:     "gap-001",
		Priority:  "high",
		Assignee:  "team-a",
		DueDate:   "2026-04-01",
		Status:    "open",
		CreatedAt: time.Now(),
	}

	data, err := json.Marshal(task)
	require.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, "rem-001", decoded["id"])
	assert.Equal(t, "high", decoded["priority"])
}

// ---------------------------------------------------------------------------
// computeGaps tests
// ---------------------------------------------------------------------------

func TestComputeGaps_SOC2(t *testing.T) {
	logger := slog.Default()
	svc := NewComplianceService("http://localhost:9999", logger)

	gaps := svc.computeGaps("SOC2")
	assert.Len(t, gaps, 2)
	for _, gap := range gaps {
		assert.Equal(t, "SOC2", gap.Framework)
		assert.NotEmpty(t, gap.ID)
	}
}

func TestComputeGaps_HIPAA(t *testing.T) {
	logger := slog.Default()
	svc := NewComplianceService("http://localhost:9999", logger)

	gaps := svc.computeGaps("HIPAA")
	assert.Len(t, gaps, 3)
}

func TestComputeGaps_GDPR(t *testing.T) {
	logger := slog.Default()
	svc := NewComplianceService("http://localhost:9999", logger)

	gaps := svc.computeGaps("GDPR")
	assert.Len(t, gaps, 2)
}

func TestComputeGaps_SOX(t *testing.T) {
	logger := slog.Default()
	svc := NewComplianceService("http://localhost:9999", logger)

	gaps := svc.computeGaps("SOX")
	assert.Len(t, gaps, 2)
}

func TestComputeGaps_Unknown(t *testing.T) {
	logger := slog.Default()
	svc := NewComplianceService("http://localhost:9999", logger)

	gaps := svc.computeGaps("UNKNOWN")
	assert.Empty(t, gaps)
}

// ---------------------------------------------------------------------------
// Edge case: Score calculation precision
// ---------------------------------------------------------------------------

func TestComputeGapScore_Precision(t *testing.T) {
	gaps := []ComplianceGap{
		{Severity: "critical", Status: "open"},
		{Severity: "high", Status: "open"},
		{Severity: "medium", Status: "open"},
		{Severity: "low", Status: "open"},
	}
	score := computeGapScore(gaps)
	// totalWeight = 4+3+2+1 = 10, penalty = 10, score = 1 - 10/30 = 0.6667
	expected := 1.0 - (10.0 / 30.0)
	assert.True(t, math.Abs(score-expected) < 1e-9,
		"expected %f, got %f", expected, score)
}
