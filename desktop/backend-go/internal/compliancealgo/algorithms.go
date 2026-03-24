// Package compliancealgo tests the pure algorithmic functions from
// internal/services/compliance_service.go in isolation, without
// requiring the miosa-sdk-go dependency that blocks the parent module.
package compliancealgo

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// AuditEntry mirrors the compliance_service.AuditEntry struct.
type AuditEntry struct {
	ID        string         `json:"id"`
	SessionID string         `json:"session_id"`
	Timestamp time.Time      `json:"timestamp"`
	Action    string         `json:"action"`
	Actor     string         `json:"actor"`
	ToolName  string         `json:"tool_name,omitempty"`
	Details   map[string]any `json:"details,omitempty"`
	Hash      string         `json:"hash"`
	PrevHash  string         `json:"prev_hash"`
}

// ComplianceGap mirrors the compliance_service.ComplianceGap struct.
type ComplianceGap struct {
	ID          string `json:"id"`
	Framework   string `json:"framework"`
	Control     string `json:"control"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
	Status      string `json:"status"`
}

// ComputeEntryHash computes the SHA-256 hash for an audit entry.
// This is the exact algorithm from compliance_service.go.
func ComputeEntryHash(entry AuditEntry, prevHash string) string {
	data := fmt.Sprintf("%s|%s|%s|%s|%s|%s",
		entry.SessionID,
		entry.Timestamp.UTC().Format(time.RFC3339Nano),
		entry.Action,
		entry.Actor,
		entry.ToolName,
		prevHash,
	)
	if entry.Details != nil {
		detailsJSON, _ := json.Marshal(entry.Details)
		data += "|" + string(detailsJSON)
	}
	h := sha256.Sum256([]byte(data))
	return hex.EncodeToString(h[:])
}

// ComputeMerkleRoot builds a Merkle tree from audit entry hashes.
// This is the exact algorithm from compliance_service.go.
func ComputeMerkleRoot(entries []AuditEntry) string {
	if len(entries) == 0 {
		return ""
	}

	hashes := make([]string, len(entries))
	for i, e := range entries {
		hashes[i] = e.Hash
	}

	for len(hashes) > 1 {
		var next []string
		for i := 0; i < len(hashes); i += 2 {
			if i+1 < len(hashes) {
				combined := hashes[i] + hashes[i+1]
				h := sha256.Sum256([]byte(combined))
				next = append(next, hex.EncodeToString(h[:]))
			} else {
				combined := hashes[i] + hashes[i]
				h := sha256.Sum256([]byte(combined))
				next = append(next, hex.EncodeToString(h[:]))
			}
		}
		hashes = next
	}

	return hashes[0]
}

// ParsePeriod parses compliance period strings into start times.
// Supports formats: "2026-Q1", "2026-01", and defaults to 1 month ago.
func ParsePeriod(period string) time.Time {
	now := time.Now()
	switch {
	case len(period) == 7 && period[5] == 'Q':
		quarter := period[6] - '1'
		year, _ := strconv.Atoi(period[:4])
		return time.Date(year, time.Month(int(quarter)*3+1), 1, 0, 0, 0, 0, time.UTC)
	case len(period) == 7:
		t, _ := time.Parse("2006-01", period)
		return t
	default:
		return now.AddDate(0, -1, 0)
	}
}

// ComputeGapScore computes a compliance score (0.0-1.0) from gaps.
// This is the exact algorithm from compliance_service.go.
func ComputeGapScore(gaps []ComplianceGap) float64 {
	if len(gaps) == 0 {
		return 1.0
	}

	totalWeight := 0.0
	penaltyWeight := 0.0
	severityWeights := map[string]float64{
		"critical": 4.0,
		"high":     3.0,
		"medium":   2.0,
		"low":      1.0,
	}

	for _, gap := range gaps {
		w := severityWeights[gap.Severity]
		totalWeight += w
		if gap.Status != "resolved" {
			penaltyWeight += w
		}
	}

	return 1.0 - (penaltyWeight / (totalWeight * 3.0))
}
