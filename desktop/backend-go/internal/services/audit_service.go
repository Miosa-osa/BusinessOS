// Package services provides audit service for compliance logging.
package services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rhl/businessos-backend/internal/audit"
)

// AuditService provides audit logging operations.
type AuditService struct {
	auditLogger *audit.AuditLogger
	pool        *pgxpool.Pool
}

// NewAuditService creates a new audit service.
func NewAuditService(pool *pgxpool.Pool) (*AuditService, error) {
	conn, err := pool.Acquire(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to acquire connection: %w", err)
	}
	defer conn.Release()

	return &AuditService{
		auditLogger: audit.NewAuditLogger(conn.Conn()),
		pool:        pool,
	}, nil
}

// LogProcessMiningDiscovery logs a process model discovery event.
func (as *AuditService) LogProcessMiningDiscovery(
	ctx context.Context,
	userID uuid.UUID,
	logSource string,
	algorithm string,
	resultHash string,
	activitiesCount int32,
	durationMs int64,
) (uuid.UUID, error) {
	if resultHash == "" {
		return uuid.UUID{}, fmt.Errorf("result_hash required")
	}

	return as.auditLogger.LogModelDiscovered(
		ctx, userID, logSource, algorithm, resultHash, activitiesCount, durationMs,
	)
}

// LogConformanceCheck logs a conformance checking event.
func (as *AuditService) LogConformanceCheck(
	ctx context.Context,
	userID uuid.UUID,
	modelID uuid.UUID,
	logID uuid.UUID,
	fitness float64,
	precision float64,
	generalization float64,
	logEntriesTested int32,
) (uuid.UUID, error) {
	if fitness < 0 || fitness > 1 {
		return uuid.UUID{}, fmt.Errorf("fitness must be between 0 and 1")
	}

	return as.auditLogger.LogConformanceChecked(
		ctx, userID, modelID, logID, fitness, precision, generalization, logEntriesTested,
	)
}

// LogStatisticsComputation logs a statistics computation event.
func (as *AuditService) LogStatisticsComputation(
	ctx context.Context,
	userID uuid.UUID,
	logID uuid.UUID,
	statisticType string,
	resultHash string,
	sampleSize int32,
) (uuid.UUID, error) {
	if statisticType == "" {
		return uuid.UUID{}, fmt.Errorf("statistic_type required")
	}

	return as.auditLogger.LogStatisticsComputed(
		ctx, userID, logID, statisticType, resultHash, sampleSize,
	)
}

// LogAccessChange logs access permission grant or revocation.
func (as *AuditService) LogAccessChange(
	ctx context.Context,
	adminID uuid.UUID,
	targetUserID uuid.UUID,
	resourceType string,
	resourceID uuid.UUID,
	permission string,
	granted bool,
) (uuid.UUID, error) {
	if resourceType == "" {
		return uuid.UUID{}, fmt.Errorf("resource_type required")
	}

	if granted {
		return as.auditLogger.LogAccessGranted(
			ctx, adminID, targetUserID, resourceType, resourceID, permission,
		)
	}

	return as.auditLogger.LogAccessRevoked(
		ctx, adminID, targetUserID, resourceID, "access revoked",
	)
}

// LogSecurityEvent logs a security-related event.
func (as *AuditService) LogSecurityEvent(
	ctx context.Context,
	eventType string,
	userID *uuid.UUID,
	details map[string]interface{},
) (uuid.UUID, error) {
	switch eventType {
	case "authentication_failure":
		return as.auditLogger.LogAuthenticationFailure(
			ctx,
			details["username_hash"].(string),
			details["ip_address"].(string),
			details["failure_reason"].(string),
		)
	case "privilege_escalation_attempt":
		return as.auditLogger.LogPrivilegeEscalationAttempt(
			ctx,
			*userID,
			details["attempted_role"].(string),
		)
	case "suspicious_activity":
		return as.auditLogger.LogSuspiciousActivity(
			ctx,
			*userID,
			details["activity_type"].(string),
			details["confidence_score"].(float64),
		)
	default:
		return uuid.UUID{}, fmt.Errorf("unknown security event type: %s", eventType)
	}
}

// QueryAuditLogs retrieves audit events with optional filters.
func (as *AuditService) QueryAuditLogs(
	ctx context.Context,
	userID *uuid.UUID,
	resourceID *uuid.UUID,
	eventType *string,
	fromDate *time.Time,
	toDate *time.Time,
	limit int,
) ([]audit.AuditEvent, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 10000 {
		limit = 10000 // Cap to prevent resource exhaustion
	}

	return as.auditLogger.QueryEvents(ctx, userID, resourceID, eventType, fromDate, toDate, limit)
}

// VerifyAuditChain verifies hash chain integrity for a sequence range.
func (as *AuditService) VerifyAuditChain(
	ctx context.Context,
	fromSeq int64,
	toSeq int64,
) (bool, []string, error) {
	if fromSeq < 0 || toSeq < fromSeq {
		return false, nil, fmt.Errorf("invalid sequence range")
	}

	return as.auditLogger.VerifyChainIntegrity(ctx, fromSeq, toSeq)
}

// ApplyLegalHold applies legal hold to audit events.
func (as *AuditService) ApplyLegalHold(
	ctx context.Context,
	eventIDs []uuid.UUID,
	reason string,
) error {
	if len(eventIDs) == 0 {
		return fmt.Errorf("at least one event_id required")
	}

	query := `
		UPDATE audit_events
		SET legal_hold = TRUE
		WHERE event_id = ANY($1)
	`

	_, err := as.pool.Exec(ctx, query, eventIDs)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to apply legal hold", "error", err)
		return fmt.Errorf("failed to apply legal hold: %w", err)
	}

	slog.InfoContext(ctx, "Legal hold applied", "event_count", len(eventIDs), "reason", reason)
	return nil
}

// LiftLegalHold removes legal hold from audit events.
func (as *AuditService) LiftLegalHold(
	ctx context.Context,
	eventIDs []uuid.UUID,
	reason string,
) error {
	if len(eventIDs) == 0 {
		return fmt.Errorf("at least one event_id required")
	}

	query := `
		UPDATE audit_events
		SET legal_hold = FALSE
		WHERE event_id = ANY($1)
	`

	_, err := as.pool.Exec(ctx, query, eventIDs)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to lift legal hold", "error", err)
		return fmt.Errorf("failed to lift legal hold: %w", err)
	}

	slog.InfoContext(ctx, "Legal hold lifted", "event_count", len(eventIDs), "reason", reason)
	return nil
}

// GetComplianceReport generates a compliance report for a date range.
func (as *AuditService) GetComplianceReport(
	ctx context.Context,
	fromDate time.Time,
	toDate time.Time,
) (map[string]interface{}, error) {
	query := `
		SELECT
			COUNT(*) as total_events,
			COUNT(DISTINCT user_id) as unique_users,
			COUNT(CASE WHEN pii_detected THEN 1 END) as events_with_pii,
			COUNT(CASE WHEN legal_hold THEN 1 END) as events_under_legal_hold,
			COUNT(CASE WHEN severity = 'critical' THEN 1 END) as critical_events,
			COUNT(CASE WHEN event_category = 'Security' THEN 1 END) as security_events
		FROM audit_events
		WHERE created_at >= $1 AND created_at <= $2
	`

	var (
		totalEvents        int
		uniqueUsers        int
		eventsWithPII      int
		eventsLegalHold    int
		criticalEvents     int
		securityEvents     int
	)

	err := as.pool.QueryRow(ctx, query, fromDate, toDate).Scan(
		&totalEvents, &uniqueUsers, &eventsWithPII, &eventsLegalHold,
		&criticalEvents, &securityEvents,
	)

	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("failed to generate report: %w", err)
	}

	report := map[string]interface{}{
		"period": map[string]string{
			"from": fromDate.Format(time.RFC3339),
			"to":   toDate.Format(time.RFC3339),
		},
		"summary": map[string]int{
			"total_events":           totalEvents,
			"unique_users":           uniqueUsers,
			"events_with_pii":        eventsWithPII,
			"events_under_legal_hold": eventsLegalHold,
			"critical_events":        criticalEvents,
			"security_events":        securityEvents,
		},
		"generated_at": time.Now().UTC().Format(time.RFC3339),
	}

	return report, nil
}

// PurgeExpiredEvents purges audit events that have exceeded their retention period.
func (as *AuditService) PurgeExpiredEvents(ctx context.Context) (int64, error) {
	query := `
		DELETE FROM audit_events
		WHERE retention_expires_at < NOW()
			AND legal_hold = FALSE
			AND event_type != 'data_deletion'
	`

	result, err := as.pool.Exec(ctx, query)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to purge expired events", "error", err)
		return 0, fmt.Errorf("failed to purge events: %w", err)
	}

	purgedCount := result.RowsAffected()
	slog.InfoContext(ctx, "Expired audit events purged", "count", purgedCount)

	return purgedCount, nil
}
