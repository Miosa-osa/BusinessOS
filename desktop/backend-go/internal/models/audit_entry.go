package models

import "time"

// AuditEntry represents a single entry in the audit trail with hash-chain integrity.
// Each entry cryptographically signs the previous entry + current data for tamper evidence.
type AuditEntry struct {
	// ID is a unique identifier for this audit entry
	ID string `json:"id"`

	// Timestamp is when the action occurred (UTC)
	Timestamp time.Time `json:"timestamp"`

	// Agent identifies which agent performed the action (e.g., "osa-healing-agent")
	Agent string `json:"agent"`

	// Action describes what was done (e.g., "create_deal", "repair_process")
	Action string `json:"action"`

	// ResourceType is the type of resource affected (e.g., "deal", "task", "project")
	ResourceType string `json:"resource_type"`

	// ResourceID is the unique ID of the resource affected
	ResourceID string `json:"resource_id"`

	// PreviousHash is the DataHash of the prior entry in the chain.
	// Empty string for the first entry.
	PreviousHash string `json:"previous_hash"`

	// DataHash is SHA256(agent + action + resource_type + resource_id + timestamp).
	// Used to create an immutable fingerprint of this entry's data.
	DataHash string `json:"data_hash"`

	// Signature is HMAC-SHA256(previous_hash + data_hash, secret).
	// Provides tamper detection: modifying any field invalidates this signature.
	Signature string `json:"signature"`

	// SNScore is the Signal-to-Noise ratio (0.0-1.0) from the agent's confidence.
	// Determines governance tier (auto >0.8, human 0.7-0.8, board <0.7).
	SNScore float64 `json:"sn_score"`

	// GovernanceTier indicates approval level:
	// - "auto": no approval needed (S/N > 0.8)
	// - "human": manager approval required (0.7 <= S/N <= 0.8)
	// - "board": C-level approval required (S/N < 0.7)
	GovernanceTier string `json:"governance_tier"`

	// Result captures the outcome: "success", "failed", "escalated"
	Result string `json:"result"`
}

// GovernanceTierInfo provides metadata about governance approval requirements.
type GovernanceTierInfo struct {
	SNScore    float64
	Tier       string
	RequiresApproval bool
	ApprovalRole     string
}

// DetermineGovernanceTier returns the governance tier based on S/N score.
// - score > 0.8: "auto" (no approval needed)
// - 0.7 <= score <= 0.8: "human" (manager approval)
// - score < 0.7: "board" (C-level approval)
func DetermineGovernanceTier(snScore float64) GovernanceTierInfo {
	if snScore > 0.8 {
		return GovernanceTierInfo{
			SNScore:    snScore,
			Tier:       "auto",
			RequiresApproval: false,
		}
	} else if snScore >= 0.7 {
		return GovernanceTierInfo{
			SNScore:    snScore,
			Tier:       "human",
			RequiresApproval: true,
			ApprovalRole:     "manager",
		}
	}
	return GovernanceTierInfo{
		SNScore:    snScore,
		Tier:       "board",
		RequiresApproval: true,
		ApprovalRole:     "ceo",
	}
}
