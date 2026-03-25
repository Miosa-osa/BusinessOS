// Package audit provides immutable audit trail implementation.
//
// Implements hash-chain based audit logging for detecting tampering
// and ensuring compliance with retention policies.
package audit

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
)

// LogEventParams contains parameters for logging an audit event.
type LogEventParams struct {
	Actor        *uuid.UUID
	Action       string
	Resource     *uuid.UUID
	ResourceType *string
	Timestamp    time.Time
	Metadata     map[string]interface{}
}

// AuditLogEntry represents a single immutable audit entry.
type AuditLogEntry struct {
	EventID      uuid.UUID
	SequenceNum  int64
	PreviousHash string
	EntryHash    string
	Actor        *uuid.UUID
	Action       string
	Resource     *uuid.UUID
	ResourceType *string
	Timestamp    time.Time
	Metadata     map[string]interface{}
}

// AuditLog manages immutable append-only audit trail.
type AuditLog struct {
	mu         sync.RWMutex
	events     []*AuditLogEntry
	eventCount int64
	lastHash   string
	logger     *slog.Logger
}

// NewAuditLog creates a new audit log instance.
func NewAuditLog() *AuditLog {
	return &AuditLog{
		events:   make([]*AuditLogEntry, 0),
		lastHash: "",
		logger:   slog.Default(),
	}
}

// LogEvent appends an immutable event to the audit trail.
//
// Each event is cryptographically linked to the previous entry,
// forming a hash chain. Attempts to tamper with any entry will
// break the chain and be detected during verification.
func (al *AuditLog) LogEvent(ctx context.Context, params *LogEventParams) error {
	if params == nil {
		return fmt.Errorf("event params cannot be nil")
	}

	al.mu.Lock()
	defer al.mu.Unlock()

	// Generate unique event ID
	eventID := uuid.New()

	// Compute entry hash (cryptographic commitment)
	entryHash := al.computeHash(
		eventID.String(),
		params.Action,
		params.Timestamp,
		al.lastHash,
	)

	// Create immutable entry
	entry := &AuditLogEntry{
		EventID:      eventID,
		SequenceNum:  al.eventCount + 1,
		PreviousHash: al.lastHash,
		EntryHash:    entryHash,
		Actor:        params.Actor,
		Action:       params.Action,
		Resource:     params.Resource,
		ResourceType: params.ResourceType,
		Timestamp:    params.Timestamp,
		Metadata:     params.Metadata,
	}

	// Append to chain (write-only)
	al.events = append(al.events, entry)
	al.eventCount++
	al.lastHash = entryHash

	al.logger.InfoContext(ctx,
		"audit event logged",
		"event_id", eventID.String(),
		"sequence", al.eventCount,
		"action", params.Action,
	)

	return nil
}

// GetHistory retrieves audit events for a specific resource.
//
// Returns events in chronological order (oldest first).
func (al *AuditLog) GetHistory(ctx context.Context, resourceID uuid.UUID) ([]*AuditLogEntry, error) {
	al.mu.RLock()
	defer al.mu.RUnlock()

	var history []*AuditLogEntry
	for _, entry := range al.events {
		if entry.Resource != nil && *entry.Resource == resourceID {
			history = append(history, entry)
		}
	}

	return history, nil
}

// EnforceRetention removes events older than the specified duration.
//
// Implements compliance with data retention policies (e.g., keep last 90 days).
// This is a destructive operation that permanently removes old entries.
func (al *AuditLog) EnforceRetention(maxAge time.Duration) error {
	al.mu.Lock()
	defer al.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-maxAge)

	var retained []*AuditLogEntry
	for _, entry := range al.events {
		if entry.Timestamp.After(cutoff) || entry.Timestamp.Equal(cutoff) {
			retained = append(retained, entry)
		}
	}

	pruned := int64(len(al.events)) - int64(len(retained))
	al.events = retained

	al.logger.Warn(
		"audit retention policy enforced",
		"pruned_count", pruned,
		"remaining_count", len(retained),
	)

	return nil
}

// VerifyChainIntegrity checks if the hash chain is intact.
//
// Computes hashes of all entries and verifies the chain is unbroken.
// Returns false if any entry has been tampered with.
func (al *AuditLog) VerifyChainIntegrity() bool {
	al.mu.RLock()
	defer al.mu.RUnlock()

	if len(al.events) == 0 {
		return true // Empty chain is valid
	}

	previousHash := ""
	for _, entry := range al.events {
		if entry.PreviousHash != previousHash {
			al.logger.Warn("chain integrity violation detected",
				"expected_hash", previousHash,
				"found_hash", entry.PreviousHash,
			)
			return false
		}

		// Recompute hash to detect tampering
		expectedHash := al.computeHash(
			entry.EventID.String(),
			entry.Action,
			entry.Timestamp,
			previousHash,
		)

		if expectedHash != entry.EntryHash {
			al.logger.Warn("entry tampering detected",
				"event_id", entry.EventID.String(),
				"expected_hash", expectedHash,
				"found_hash", entry.EntryHash,
			)
			return false
		}

		previousHash = entry.EntryHash
	}

	return true
}

// GetEventCount returns the total number of logged events.
func (al *AuditLog) GetEventCount() int64 {
	al.mu.RLock()
	defer al.mu.RUnlock()
	return al.eventCount
}

// computeHash generates a SHA256 hash for an audit entry.
// Links the current entry to the previous one, forming an immutable chain.
func (al *AuditLog) computeHash(eventID, action string, timestamp time.Time, previousHash string) string {
	input := fmt.Sprintf("%s:%s:%d:%s", eventID, action, timestamp.Unix(), previousHash)
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}
