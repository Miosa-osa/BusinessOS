package compliancealgo

import (
	"crypto/sha256"
	"encoding/hex"
	"math"
	"testing"
	"time"
)

// sha256Hex computes the SHA-256 hex digest of a string.
func sha256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

func TestComputeEntryHash_Deterministic(t *testing.T) {
	entry := AuditEntry{
		ID:        "audit-001",
		SessionID: "sess-abc",
		Timestamp: time.Date(2026, 3, 15, 10, 30, 0, 0, time.UTC),
		Action:    "tool_call",
		Actor:     "agent-1",
		ToolName:  "http_post",
		PrevHash:  "",
	}

	h1 := ComputeEntryHash(entry, "")
	h2 := ComputeEntryHash(entry, "")

	if h1 != h2 {
		t.Errorf("hash not deterministic: %s != %s", h1, h2)
	}
}

func TestComputeEntryHash_ChangesWithPrevHash(t *testing.T) {
	entry := AuditEntry{
		ID:        "audit-002",
		SessionID: "sess-abc",
		Timestamp: time.Date(2026, 3, 15, 10, 30, 0, 0, time.UTC),
		Action:    "tool_call",
		Actor:     "agent-1",
		ToolName:  "http_post",
	}

	h1 := ComputeEntryHash(entry, "")
	h2 := ComputeEntryHash(entry, "previous-hash-value")

	if h1 == h2 {
		t.Error("hash should change when prevHash differs")
	}
}

func TestComputeEntryHash_ChangesWithDetails(t *testing.T) {
	entry := AuditEntry{
		ID:        "audit-003",
		SessionID: "sess-abc",
		Timestamp: time.Date(2026, 3, 15, 10, 30, 0, 0, time.UTC),
		Action:    "tool_call",
		Actor:     "agent-1",
		ToolName:  "http_post",
	}

	h1 := ComputeEntryHash(entry, "")
	entry.Details = map[string]any{"key": "value"}
	h2 := ComputeEntryHash(entry, "")

	if h1 == h2 {
		t.Error("hash should change when details are added")
	}
}

func TestComputeEntryHash_ChangesWithDifferentFields(t *testing.T) {
	base := AuditEntry{
		SessionID: "sess-abc",
		Timestamp: time.Date(2026, 3, 15, 10, 30, 0, 0, time.UTC),
		Action:    "tool_call",
		Actor:     "agent-1",
		ToolName:  "http_post",
	}

	h1 := ComputeEntryHash(base, "")

	base.Action = "different_action"
	h2 := ComputeEntryHash(base, "")

	if h1 == h2 {
		t.Error("hash should change when action differs")
	}
}

func TestComputeEntryHash_Sha256Length(t *testing.T) {
	entry := AuditEntry{
		SessionID: "sess-test",
		Timestamp: time.Now().UTC(),
		Action:    "test",
		Actor:     "tester",
	}

	h := ComputeEntryHash(entry, "")
	if len(h) != 64 {
		t.Errorf("SHA-256 hex should be 64 chars, got %d", len(h))
	}
}

func TestComputeMerkleRoot_Empty(t *testing.T) {
	root := ComputeMerkleRoot(nil)
	if root != "" {
		t.Errorf("expected empty string for empty entries, got %s", root)
	}
}

func TestComputeMerkleRoot_SingleEntry(t *testing.T) {
	entries := []AuditEntry{
		{Hash: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
	}
	root := ComputeMerkleRoot(entries)
	// Single entry: the loop (len > 1) doesn't execute, so the raw hash is returned.
	expected := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	if root != expected {
		t.Errorf("Merkle root mismatch:\n  got: %s\n  want: %s", root, expected)
	}
}

func TestComputeMerkleRoot_TwoEntries(t *testing.T) {
	h1 := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	h2 := "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	entries := []AuditEntry{{Hash: h1}, {Hash: h2}}
	root := ComputeMerkleRoot(entries)

	expected := sha256Hex(h1 + h2)
	if root != expected {
		t.Errorf("Merkle root mismatch:\n  got: %s\n  want: %s", root, expected)
	}
}

func TestComputeMerkleRoot_OddCount(t *testing.T) {
	h1 := "1111111111111111111111111111111111111111111111111111111111111111"
	h2 := "2222222222222222222222222222222222222222222222222222222222222222"
	h3 := "3333333333333333333333333333333333333333333333333333333333333333"
	entries := []AuditEntry{{Hash: h1}, {Hash: h2}, {Hash: h3}}

	root := ComputeMerkleRoot(entries)

	// First level: SHA256(h1+h2), SHA256(h3+h3) [odd duplicate]
	level1a := sha256Hex(h1 + h2)
	level1b := sha256Hex(h3 + h3)
	expected := sha256Hex(level1a + level1b)

	if root != expected {
		t.Errorf("Merkle root mismatch for odd count:\n  got: %s\n  want: %s", root, expected)
	}
}

func TestComputeMerkleRoot_Deterministic(t *testing.T) {
	entries := []AuditEntry{
		{Hash: "aaaa"},
		{Hash: "bbbb"},
		{Hash: "cccc"},
		{Hash: "dddd"},
	}

	r1 := ComputeMerkleRoot(entries)
	r2 := ComputeMerkleRoot(entries)

	if r1 != r2 {
		t.Error("Merkle root should be deterministic")
	}
}

func TestParsePeriod_Quarter(t *testing.T) {
	parsed := ParsePeriod("2026-Q1")
	expected := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	if !parsed.Equal(expected) {
		t.Errorf("Q1: got %v, want %v", parsed, expected)
	}

	parsed = ParsePeriod("2026-Q2")
	expected = time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	if !parsed.Equal(expected) {
		t.Errorf("Q2: got %v, want %v", parsed, expected)
	}

	parsed = ParsePeriod("2026-Q3")
	expected = time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	if !parsed.Equal(expected) {
		t.Errorf("Q3: got %v, want %v", parsed, expected)
	}

	parsed = ParsePeriod("2026-Q4")
	expected = time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	if !parsed.Equal(expected) {
		t.Errorf("Q4: got %v, want %v", parsed, expected)
	}
}

func TestParsePeriod_Month(t *testing.T) {
	parsed := ParsePeriod("2026-06")
	expected := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	if !parsed.Equal(expected) {
		t.Errorf("month: got %v, want %v", parsed, expected)
	}
}

func TestParsePeriod_Default(t *testing.T) {
	before := time.Now().AddDate(0, -1, -1)
	parsed := ParsePeriod("invalid-period")
	after := time.Now()

	if parsed.Before(before) || parsed.After(after) {
		t.Errorf("default should be ~1 month ago, got %v", parsed)
	}
}

func TestComputeGapScore_Empty(t *testing.T) {
	score := ComputeGapScore(nil)
	if score != 1.0 {
		t.Errorf("empty gaps should score 1.0, got %f", score)
	}
}

func TestComputeGapScore_AllResolved(t *testing.T) {
	gaps := []ComplianceGap{
		{Severity: "critical", Status: "resolved"},
		{Severity: "high", Status: "resolved"},
	}
	score := ComputeGapScore(gaps)
	if score != 1.0 {
		t.Errorf("all resolved should score 1.0, got %f", score)
	}
}

func TestComputeGapScore_AllOpen(t *testing.T) {
	gaps := []ComplianceGap{
		{Severity: "critical", Status: "open"},
	}
	score := ComputeGapScore(gaps)
	// penalty = 4.0, totalWeight*3 = 12.0, score = 1 - 4/12 = 0.667
	expected := 1.0 - (4.0 / (4.0 * 3.0))
	if math.Abs(score-expected) > 1e-9 {
		t.Errorf("1 critical open: got %f, want %f", score, expected)
	}
}

func TestComputeGapScore_MixedSeverity(t *testing.T) {
	gaps := []ComplianceGap{
		{Severity: "critical", Status: "open"},
		{Severity: "high", Status: "open"},
		{Severity: "medium", Status: "resolved"},
		{Severity: "low", Status: "open"},
	}
	// totalWeight = 4+3+2+1 = 10, penalty = 4+3+1 = 8 (medium resolved)
	// score = 1 - 8/30 = 0.733
	score := ComputeGapScore(gaps)
	expected := 1.0 - (8.0 / (10.0 * 3.0))
	if math.Abs(score-expected) > 1e-9 {
		t.Errorf("mixed: got %f, want %f", score, expected)
	}
}

func TestComputeGapScore_Bounded(t *testing.T) {
	// Many critical open gaps shouldn't go below 0
	gaps := []ComplianceGap{}
	for i := 0; i < 20; i++ {
		gaps = append(gaps, ComplianceGap{Severity: "critical", Status: "open"})
	}
	score := ComputeGapScore(gaps)
	if score < 0.0 || score > 1.0 {
		t.Errorf("score should be in [0,1], got %f", score)
	}
}

func TestComputeGapScore_LowSeverityOnly(t *testing.T) {
	gaps := []ComplianceGap{
		{Severity: "low", Status: "open"},
		{Severity: "low", Status: "open"},
	}
	// totalWeight = 2, penalty = 2, score = 1 - 2/6 = 0.667
	score := ComputeGapScore(gaps)
	expected := 1.0 - (2.0 / (2.0 * 3.0))
	if math.Abs(score-expected) > 1e-9 {
		t.Errorf("low severity: got %f, want %f", score, expected)
	}
}
