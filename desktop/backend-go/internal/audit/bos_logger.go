// Package audit provides comprehensive audit logging for compliance.
//
// Implements hash-chain audit trail with GDPR/SOC2 compliance for all
// BOS ↔ BusinessOS operations.
package audit

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// AuditEvent represents a single audit log entry.
type AuditEvent struct {
	// Chain integrity
	EventID        uuid.UUID       `json:"event_id"`
	SequenceNumber int64           `json:"sequence_number"`
	EntryHash      string          `json:"entry_hash"`
	PreviousHash   string          `json:"previous_hash"`
	MerkleTreeHash *string         `json:"merkle_tree_hash,omitempty"`

	// Event metadata
	EventType      string    `json:"event_type"`
	EventCategory  string    `json:"event_category"`
	Timestamp      time.Time `json:"timestamp"`
	Severity       string    `json:"severity"`

	// Actor context
	UserID    *uuid.UUID `json:"user_id,omitempty"`
	SessionID *uuid.UUID `json:"session_id,omitempty"`
	IPAddress *string    `json:"ip_address,omitempty"`
	UserAgent *string    `json:"user_agent,omitempty"`

	// Resource context
	ResourceType *string   `json:"resource_type,omitempty"`
	ResourceID   *uuid.UUID `json:"resource_id,omitempty"`
	WorkspaceID  *uuid.UUID `json:"workspace_id,omitempty"`

	// Event payload
	Payload json.RawMessage `json:"payload"`

	// Compliance metadata
	GDPRClassification     *string `json:"gdpr_classification,omitempty"`
	DataSubjectsAffected   *int32  `json:"data_subjects_affected,omitempty"`
	PIIDetected            bool    `json:"pii_detected"`
	LegalHold              bool    `json:"legal_hold"`
	RetentionExpiresAt     *time.Time `json:"retention_expires_at,omitempty"`
	DeletionBlockedUntil   *time.Time `json:"deletion_blocked_until,omitempty"`
}

// AuditLogger provides audit event recording and verification.
type AuditLogger struct {
	conn *pgx.Conn
}

// NewAuditLogger creates a new audit logger.
func NewAuditLogger(conn *pgx.Conn) *AuditLogger {
	return &AuditLogger{conn: conn}
}

// LogModelDiscovered logs a process mining model discovery.
func (al *AuditLogger) LogModelDiscovered(
	ctx context.Context,
	userID uuid.UUID,
	logSource string,
	algorithm string,
	resultHash string,
	activitiesCount int32,
	durationMs int64,
) (uuid.UUID, error) {
	payload := map[string]interface{}{
		"log_source":      logSource,
		"algorithm":       algorithm,
		"result_hash":     resultHash,
		"activities_count": activitiesCount,
		"duration_ms":     durationMs,
	}

	return al.recordEvent(ctx, AuditEvent{
		EventType:              "model_discovered",
		EventCategory:          "ProcessMining",
		UserID:                 &userID,
		Severity:               "info",
		GDPRClassification:     stringPtr("processing_activity"),
		Payload:                toJSON(payload),
	})
}

// LogConformanceChecked logs a process conformance check.
func (al *AuditLogger) LogConformanceChecked(
	ctx context.Context,
	userID uuid.UUID,
	modelID uuid.UUID,
	logID uuid.UUID,
	fitness float64,
	precision float64,
	generalization float64,
	logEntriesTested int32,
) (uuid.UUID, error) {
	resourceType := "process_model"

	payload := map[string]interface{}{
		"model_id":        modelID,
		"log_id":          logID,
		"fitness":         fitness,
		"precision":       precision,
		"generalization":  generalization,
		"log_entries_tested": logEntriesTested,
	}

	return al.recordEvent(ctx, AuditEvent{
		EventType:              "conformance_checked",
		EventCategory:          "ProcessMining",
		UserID:                 &userID,
		ResourceType:           &resourceType,
		ResourceID:             &modelID,
		Severity:               "info",
		GDPRClassification:     stringPtr("processing_activity"),
		DataSubjectsAffected:   &logEntriesTested,
		Payload:                toJSON(payload),
	})
}

// LogStatisticsComputed logs process statistics computation.
func (al *AuditLogger) LogStatisticsComputed(
	ctx context.Context,
	userID uuid.UUID,
	logID uuid.UUID,
	statisticType string,
	resultHash string,
	sampleSize int32,
) (uuid.UUID, error) {
	resourceType := "log"

	payload := map[string]interface{}{
		"statistic_type": statisticType,
		"result_hash":    resultHash,
		"sample_size":    sampleSize,
	}

	return al.recordEvent(ctx, AuditEvent{
		EventType:              "statistics_computed",
		EventCategory:          "ProcessMining",
		UserID:                 &userID,
		ResourceType:           &resourceType,
		ResourceID:             &logID,
		Severity:               "info",
		GDPRClassification:     stringPtr("analytics"),
		DataSubjectsAffected:   &sampleSize,
		Payload:                toJSON(payload),
	})
}

// LogAccessGranted logs access permission grant.
func (al *AuditLogger) LogAccessGranted(
	ctx context.Context,
	adminID uuid.UUID,
	targetUserID uuid.UUID,
	resourceType string,
	resourceID uuid.UUID,
	permission string,
) (uuid.UUID, error) {
	payload := map[string]interface{}{
		"target_user_id": targetUserID,
		"permission":     permission,
		"granted_by":     adminID,
	}

	return al.recordEvent(ctx, AuditEvent{
		EventType:              "access_granted",
		EventCategory:          "Compliance",
		UserID:                 &adminID,
		ResourceType:           &resourceType,
		ResourceID:             &resourceID,
		Severity:               "info",
		GDPRClassification:     stringPtr("access_control"),
		DataSubjectsAffected:   int32Ptr(1),
		Payload:                toJSON(payload),
	})
}

// LogAccessRevoked logs access permission revocation.
func (al *AuditLogger) LogAccessRevoked(
	ctx context.Context,
	adminID uuid.UUID,
	targetUserID uuid.UUID,
	resourceID uuid.UUID,
	revocationReason string,
) (uuid.UUID, error) {
	resourceType := "resource"

	payload := map[string]interface{}{
		"target_user_id": targetUserID,
		"revoked_by":     adminID,
		"reason":         revocationReason,
	}

	return al.recordEvent(ctx, AuditEvent{
		EventType:              "access_revoked",
		EventCategory:          "Compliance",
		UserID:                 &adminID,
		ResourceType:           &resourceType,
		ResourceID:             &resourceID,
		Severity:               "warning",
		GDPRClassification:     stringPtr("access_control"),
		DataSubjectsAffected:   int32Ptr(1),
		Payload:                toJSON(payload),
	})
}

// LogDataDeletion logs GDPR right-to-be-forgotten deletion.
func (al *AuditLogger) LogDataDeletion(
	ctx context.Context,
	adminID uuid.UUID,
	targetUserID uuid.UUID,
	resourceID uuid.UUID,
	deletionReason string,
	previousHash string,
) (uuid.UUID, error) {
	resourceType := "resource"

	payload := map[string]interface{}{
		"deletion_reason": deletionReason,
		"previous_hash":   previousHash,
	}

	// GDPR deletion should be marked critical
	event := AuditEvent{
		EventType:              "data_deletion",
		EventCategory:          "Compliance",
		UserID:                 &adminID,
		ResourceType:           &resourceType,
		ResourceID:             &resourceID,
		Severity:               "critical",
		GDPRClassification:     stringPtr("right_to_be_forgotten"),
		DataSubjectsAffected:   int32Ptr(1),
		PIIDetected:            true,
		Payload:                toJSON(payload),
	}

	eventID, err := al.recordEvent(ctx, event)
	if err != nil {
		return uuid.UUID{}, err
	}

	// Set deletion-related retention metadata
	now := time.Now()
	retentionExpires := now.AddDate(7, 0, 0) // 7 years per GDPR
	deletionBlockedUntil := now.AddDate(0, 0, 30) // 30 day grace period

	query := `
		UPDATE audit_events
		SET
			retention_expires_at = $1,
			deletion_blocked_until = $2
		WHERE event_id = $3
	`

	_, err = al.conn.Exec(ctx, query, retentionExpires, deletionBlockedUntil, eventID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to set retention metadata", "error", err, "event_id", eventID)
	}

	return eventID, nil
}

// LogAuthenticationFailure logs authentication failure (security event).
func (al *AuditLogger) LogAuthenticationFailure(
	ctx context.Context,
	usernameHash string,
	ipAddress string,
	failureReason string,
) (uuid.UUID, error) {
	payload := map[string]interface{}{
		"username_hash": usernameHash,
		"ip_address":    ipAddress,
		"failure_reason": failureReason,
	}

	return al.recordEvent(ctx, AuditEvent{
		EventType:      "authentication_failure",
		EventCategory:  "Security",
		IPAddress:      &ipAddress,
		Severity:       "warning",
		GDPRClassification: stringPtr("security_event"),
		Payload:        toJSON(payload),
	})
}

// LogPrivilegeEscalationAttempt logs attempted privilege escalation.
func (al *AuditLogger) LogPrivilegeEscalationAttempt(
	ctx context.Context,
	userID uuid.UUID,
	attemptedRole string,
) (uuid.UUID, error) {
	payload := map[string]interface{}{
		"attempted_role": attemptedRole,
		"action_taken":   "blocked",
	}

	return al.recordEvent(ctx, AuditEvent{
		EventType:      "privilege_escalation_attempt",
		EventCategory:  "Security",
		UserID:         &userID,
		Severity:       "critical",
		GDPRClassification: stringPtr("security_event"),
		Payload:        toJSON(payload),
	})
}

// LogSuspiciousActivity logs detected suspicious behavior.
func (al *AuditLogger) LogSuspiciousActivity(
	ctx context.Context,
	userID uuid.UUID,
	activityType string,
	confidenceScore float64,
) (uuid.UUID, error) {
	payload := map[string]interface{}{
		"activity_type":    activityType,
		"confidence_score": confidenceScore,
	}

	return al.recordEvent(ctx, AuditEvent{
		EventType:      "suspicious_activity_detected",
		EventCategory:  "Security",
		UserID:         &userID,
		Severity:       "warning",
		GDPRClassification: stringPtr("security_event"),
		Payload:        toJSON(payload),
	})
}

// recordEvent stores an audit event in PostgreSQL with hash chain integrity.
func (al *AuditLogger) recordEvent(ctx context.Context, event AuditEvent) (uuid.UUID, error) {
	// Assign event ID
	event.EventID = uuid.New()
	event.Timestamp = time.Now().UTC()

	// Get sequence number and previous hash
	var sequenceNumber int64
	var previousHash string

	err := al.conn.QueryRow(ctx, `
		SELECT COALESCE(MAX(sequence_number), 0) + 1,
		       COALESCE((SELECT entry_hash FROM audit_events ORDER BY sequence_number DESC LIMIT 1), '0')
		FROM audit_events
	`).Scan(&sequenceNumber, &previousHash)

	if err != nil && err != pgx.ErrNoRows {
		return uuid.UUID{}, fmt.Errorf("failed to get sequence: %w", err)
	}

	event.SequenceNumber = sequenceNumber
	event.PreviousHash = previousHash

	// Compute entry hash
	event.EntryHash = computeEntryHash(sequenceNumber, &event)

	// Set default retention
	now := time.Now()
	retentionExpires := now.AddDate(7, 0, 0) // 7 years

	// Insert into PostgreSQL
	query := `
		INSERT INTO audit_events (
			event_id, sequence_number, entry_hash, previous_hash,
			event_type, event_category, created_at, severity,
			user_id, session_id, ip_address, user_agent,
			resource_type, resource_id, workspace_id,
			payload,
			legal_hold, retention_expires_at,
			gdpr_classification, data_subjects_affected, pii_detected
		) VALUES (
			$1, $2, $3, $4,
			$5, $6, $7, $8,
			$9, $10, $11, $12,
			$13, $14, $15,
			$16,
			$17, $18,
			$19, $20, $21
		)
	`

	_, err = al.conn.Exec(ctx, query,
		event.EventID,
		event.SequenceNumber,
		event.EntryHash,
		event.PreviousHash,
		event.EventType,
		event.EventCategory,
		event.Timestamp,
		event.Severity,
		event.UserID,
		event.SessionID,
		event.IPAddress,
		event.UserAgent,
		event.ResourceType,
		event.ResourceID,
		event.WorkspaceID,
		event.Payload,
		event.LegalHold,
		retentionExpires,
		event.GDPRClassification,
		event.DataSubjectsAffected,
		event.PIIDetected,
	)

	if err != nil {
		return uuid.UUID{}, fmt.Errorf("failed to insert audit event: %w", err)
	}

	slog.InfoContext(ctx, "Audit event recorded",
		"event_id", event.EventID,
		"event_type", event.EventType,
		"user_id", event.UserID,
		"sequence", sequenceNumber,
	)

	return event.EventID, nil
}

// QueryEvents retrieves audit events with optional filters.
func (al *AuditLogger) QueryEvents(
	ctx context.Context,
	userID *uuid.UUID,
	resourceID *uuid.UUID,
	eventType *string,
	fromDate *time.Time,
	toDate *time.Time,
	limit int,
) ([]AuditEvent, error) {
	query := `
		SELECT
			event_id, sequence_number, entry_hash, previous_hash,
			event_type, event_category, created_at, severity,
			user_id, session_id, ip_address, user_agent,
			resource_type, resource_id, workspace_id,
			payload,
			legal_hold, retention_expires_at, deletion_blocked_until,
			gdpr_classification, data_subjects_affected, pii_detected
		FROM audit_events
		WHERE 1=1
	`

	args := []interface{}{}
	argIndex := 1

	if userID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argIndex)
		args = append(args, userID)
		argIndex++
	}

	if resourceID != nil {
		query += fmt.Sprintf(" AND resource_id = $%d", argIndex)
		args = append(args, resourceID)
		argIndex++
	}

	if eventType != nil {
		query += fmt.Sprintf(" AND event_type = $%d", argIndex)
		args = append(args, eventType)
		argIndex++
	}

	if fromDate != nil {
		query += fmt.Sprintf(" AND created_at >= $%d", argIndex)
		args = append(args, fromDate)
		argIndex++
	}

	if toDate != nil {
		query += fmt.Sprintf(" AND created_at <= $%d", argIndex)
		args = append(args, toDate)
		argIndex++
	}

	query += ` ORDER BY created_at DESC`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, limit)
	}

	rows, err := al.conn.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit events: %w", err)
	}
	defer rows.Close()

	var events []AuditEvent
	for rows.Next() {
		var e AuditEvent
		var createdAt time.Time
		var deletionBlockedUntil pgtype.Timestamp

		err := rows.Scan(
			&e.EventID, &e.SequenceNumber, &e.EntryHash, &e.PreviousHash,
			&e.EventType, &e.EventCategory, &createdAt, &e.Severity,
			&e.UserID, &e.SessionID, &e.IPAddress, &e.UserAgent,
			&e.ResourceType, &e.ResourceID, &e.WorkspaceID,
			&e.Payload,
			&e.LegalHold, &e.RetentionExpiresAt, &deletionBlockedUntil,
			&e.GDPRClassification, &e.DataSubjectsAffected, &e.PIIDetected,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit event: %w", err)
		}

		e.Timestamp = createdAt

		if deletionBlockedUntil.Valid {
			e.DeletionBlockedUntil = &deletionBlockedUntil.Time
		}

		events = append(events, e)
	}

	return events, rows.Err()
}

// VerifyChainIntegrity verifies hash chain integrity for a range of events.
func (al *AuditLogger) VerifyChainIntegrity(
	ctx context.Context,
	fromSeq int64,
	toSeq int64,
) (bool, []string, error) {
	query := `
		SELECT event_id, sequence_number, entry_hash, previous_hash, payload
		FROM audit_events
		WHERE sequence_number >= $1 AND sequence_number <= $2
		ORDER BY sequence_number ASC
	`

	rows, err := al.conn.Query(ctx, query, fromSeq, toSeq)
	if err != nil {
		return false, nil, fmt.Errorf("failed to fetch chain entries: %w", err)
	}
	defer rows.Close()

	var issues []string
	var previousHash string
	isValid := true

	for rows.Next() {
		var eventID uuid.UUID
		var seqNum int64
		var entryHash string
		var nextPrevHash string
		var payload json.RawMessage

		err := rows.Scan(&eventID, &seqNum, &entryHash, &nextPrevHash, &payload)
		if err != nil {
			return false, nil, fmt.Errorf("failed to scan entry: %w", err)
		}

		// Verify link to previous
		if seqNum > fromSeq && nextPrevHash != previousHash {
			isValid = false
			issues = append(issues, fmt.Sprintf(
				"Chain break at sequence %d: expected %s, got %s",
				seqNum, previousHash, nextPrevHash,
			))
		}

		previousHash = entryHash
	}

	return isValid, issues, rows.Err()
}

// Helper functions

func computeEntryHash(seq int64, event *AuditEvent) string {
	h := sha256.New()
	data := fmt.Sprintf("%d|%s|%s|%s|%s",
		seq, event.Timestamp.String(), event.EventType,
		event.PreviousHash, string(event.Payload))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func toJSON(v interface{}) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		return json.RawMessage(`{}`)
	}
	return b
}

func stringPtr(s string) *string {
	return &s
}

func int32Ptr(i int32) *int32 {
	return &i
}
