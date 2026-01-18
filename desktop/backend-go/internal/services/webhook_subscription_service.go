// Package services provides business logic for BusinessOS.
package services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rhl/businessos-backend/internal/security"
)

// =============================================================================
// WEBHOOK SUBSCRIPTION SERVICE
// Manages webhook subscriptions for real-time sync
// =============================================================================

// WebhookSubscriptionService manages webhook subscriptions for live sync.
type WebhookSubscriptionService struct {
	pool       *pgxpool.Pool
	logger     *slog.Logger
	encryption *security.TokenEncryption
}

// NewWebhookSubscriptionService creates a new webhook subscription service.
func NewWebhookSubscriptionService(pool *pgxpool.Pool, logger *slog.Logger) *WebhookSubscriptionService {
	return &WebhookSubscriptionService{
		pool:       pool,
		logger:     logger,
		encryption: security.GetGlobalEncryption(),
	}
}

// =============================================================================
// TYPES
// =============================================================================

// WebhookSubscription represents an active webhook registration.
type WebhookSubscription struct {
	ID                     uuid.UUID  `json:"id"`
	UserID                 uuid.UUID  `json:"user_id"`
	WorkspaceID            *uuid.UUID `json:"workspace_id,omitempty"`
	Provider               string     `json:"provider"`
	ResourceType           string     `json:"resource_type,omitempty"`
	ExternalSubscriptionID string     `json:"external_subscription_id,omitempty"`
	WebhookURL             string     `json:"webhook_url,omitempty"`
	Events                 []string   `json:"events,omitempty"`
	ExpiresAt              *time.Time `json:"expires_at,omitempty"`
	Status                 string     `json:"status"` // active, expired, failed, paused
	LastEventAt            *time.Time `json:"last_event_at,omitempty"`
	EventCount             int        `json:"event_count"`
	ConsecutiveFailures    int        `json:"consecutive_failures"`
	LastError              string     `json:"last_error,omitempty"`
	LastErrorAt            *time.Time `json:"last_error_at,omitempty"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
}

// CreateSubscriptionInput is the input for creating a webhook subscription.
type CreateSubscriptionInput struct {
	UserID                 uuid.UUID
	WorkspaceID            *uuid.UUID
	Provider               string
	ResourceType           string
	ExternalSubscriptionID string
	WebhookURL             string
	WebhookSecret          string // Will be encrypted
	Events                 []string
	ExpiresAt              *time.Time
}

// =============================================================================
// SUBSCRIPTION MANAGEMENT
// =============================================================================

// CreateSubscription creates a new webhook subscription.
func (s *WebhookSubscriptionService) CreateSubscription(ctx context.Context, input CreateSubscriptionInput) (*WebhookSubscription, error) {
	// Encrypt the webhook secret if provided
	var encryptedSecret *string
	if input.WebhookSecret != "" && s.encryption != nil {
		encrypted, err := s.encryption.Encrypt(input.WebhookSecret)
		if err != nil {
			s.logger.Error("Failed to encrypt webhook secret",
				slog.String("provider", input.Provider),
				slog.Any("error", err),
			)
			return nil, fmt.Errorf("encrypt webhook secret: %w", err)
		}
		encryptedSecret = &encrypted
	}

	var sub WebhookSubscription
	err := s.pool.QueryRow(ctx, `
		INSERT INTO webhook_subscriptions (
			user_id, workspace_id, provider, resource_type,
			external_subscription_id, webhook_url, webhook_secret_encrypted,
			events, expires_at, status
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, 'active')
		ON CONFLICT (user_id, provider, resource_type) DO UPDATE SET
			external_subscription_id = EXCLUDED.external_subscription_id,
			webhook_url = EXCLUDED.webhook_url,
			webhook_secret_encrypted = EXCLUDED.webhook_secret_encrypted,
			events = EXCLUDED.events,
			expires_at = EXCLUDED.expires_at,
			status = 'active',
			consecutive_failures = 0,
			last_error = NULL,
			updated_at = NOW()
		RETURNING id, user_id, provider, resource_type, status, event_count, created_at, updated_at
	`,
		input.UserID, input.WorkspaceID, input.Provider, input.ResourceType,
		input.ExternalSubscriptionID, input.WebhookURL, encryptedSecret,
		input.Events, input.ExpiresAt,
	).Scan(
		&sub.ID, &sub.UserID, &sub.Provider, &sub.ResourceType,
		&sub.Status, &sub.EventCount, &sub.CreatedAt, &sub.UpdatedAt,
	)
	if err != nil {
		s.logger.Error("Failed to create webhook subscription",
			slog.String("provider", input.Provider),
			slog.Any("error", err),
		)
		return nil, fmt.Errorf("create webhook subscription: %w", err)
	}

	s.logger.Info("Webhook subscription created",
		slog.String("provider", input.Provider),
		slog.String("resource_type", input.ResourceType),
		slog.String("subscription_id", sub.ID.String()),
	)

	return &sub, nil
}

// GetSubscription retrieves a webhook subscription.
func (s *WebhookSubscriptionService) GetSubscription(ctx context.Context, userID uuid.UUID, provider, resourceType string) (*WebhookSubscription, error) {
	var sub WebhookSubscription
	err := s.pool.QueryRow(ctx, `
		SELECT id, user_id, workspace_id, provider, resource_type,
			   external_subscription_id, webhook_url, events,
			   expires_at, status, last_event_at, event_count,
			   consecutive_failures, last_error, last_error_at,
			   created_at, updated_at
		FROM webhook_subscriptions
		WHERE user_id = $1 AND provider = $2 AND resource_type = $3
	`, userID, provider, resourceType).Scan(
		&sub.ID, &sub.UserID, &sub.WorkspaceID, &sub.Provider, &sub.ResourceType,
		&sub.ExternalSubscriptionID, &sub.WebhookURL, &sub.Events,
		&sub.ExpiresAt, &sub.Status, &sub.LastEventAt, &sub.EventCount,
		&sub.ConsecutiveFailures, &sub.LastError, &sub.LastErrorAt,
		&sub.CreatedAt, &sub.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get webhook subscription: %w", err)
	}

	return &sub, nil
}

// ListSubscriptions lists all webhook subscriptions for a user.
func (s *WebhookSubscriptionService) ListSubscriptions(ctx context.Context, userID uuid.UUID) ([]WebhookSubscription, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, user_id, workspace_id, provider, resource_type,
			   external_subscription_id, webhook_url, events,
			   expires_at, status, last_event_at, event_count,
			   consecutive_failures, last_error, last_error_at,
			   created_at, updated_at
		FROM webhook_subscriptions
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list webhook subscriptions: %w", err)
	}
	defer rows.Close()

	var subscriptions []WebhookSubscription
	for rows.Next() {
		var sub WebhookSubscription
		err := rows.Scan(
			&sub.ID, &sub.UserID, &sub.WorkspaceID, &sub.Provider, &sub.ResourceType,
			&sub.ExternalSubscriptionID, &sub.WebhookURL, &sub.Events,
			&sub.ExpiresAt, &sub.Status, &sub.LastEventAt, &sub.EventCount,
			&sub.ConsecutiveFailures, &sub.LastError, &sub.LastErrorAt,
			&sub.CreatedAt, &sub.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan webhook subscription: %w", err)
		}
		subscriptions = append(subscriptions, sub)
	}

	return subscriptions, nil
}

// =============================================================================
// EVENT TRACKING
// =============================================================================

// RecordEvent records that a webhook event was received.
func (s *WebhookSubscriptionService) RecordEvent(ctx context.Context, userID uuid.UUID, provider, resourceType string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE webhook_subscriptions
		SET last_event_at = NOW(),
			event_count = event_count + 1,
			consecutive_failures = 0,
			updated_at = NOW()
		WHERE user_id = $1 AND provider = $2 AND resource_type = $3
	`, userID, provider, resourceType)
	if err != nil {
		return fmt.Errorf("record webhook event: %w", err)
	}
	return nil
}

// RecordError records a webhook processing error.
func (s *WebhookSubscriptionService) RecordError(ctx context.Context, userID uuid.UUID, provider, resourceType, errorMsg string) error {
	result, err := s.pool.Exec(ctx, `
		UPDATE webhook_subscriptions
		SET consecutive_failures = consecutive_failures + 1,
			last_error = $4,
			last_error_at = NOW(),
			status = CASE WHEN consecutive_failures >= 10 THEN 'failed' ELSE status END,
			updated_at = NOW()
		WHERE user_id = $1 AND provider = $2 AND resource_type = $3
	`, userID, provider, resourceType, errorMsg)
	if err != nil {
		return fmt.Errorf("record webhook error: %w", err)
	}

	// Check if we marked it as failed
	if result.RowsAffected() > 0 {
		s.logger.Warn("Webhook subscription error recorded",
			slog.String("provider", provider),
			slog.String("resource_type", resourceType),
			slog.String("error", errorMsg),
		)
	}

	return nil
}

// =============================================================================
// STATUS MANAGEMENT
// =============================================================================

// UpdateStatus updates the status of a webhook subscription.
func (s *WebhookSubscriptionService) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE webhook_subscriptions
		SET status = $2, updated_at = NOW()
		WHERE id = $1
	`, id, status)
	if err != nil {
		return fmt.Errorf("update webhook status: %w", err)
	}
	return nil
}

// PauseSubscription pauses a webhook subscription.
func (s *WebhookSubscriptionService) PauseSubscription(ctx context.Context, id uuid.UUID) error {
	return s.UpdateStatus(ctx, id, "paused")
}

// ResumeSubscription resumes a paused webhook subscription.
func (s *WebhookSubscriptionService) ResumeSubscription(ctx context.Context, id uuid.UUID) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE webhook_subscriptions
		SET status = 'active',
			consecutive_failures = 0,
			last_error = NULL,
			updated_at = NOW()
		WHERE id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("resume webhook subscription: %w", err)
	}
	return nil
}

// DeleteSubscription deletes a webhook subscription.
func (s *WebhookSubscriptionService) DeleteSubscription(ctx context.Context, id uuid.UUID) error {
	_, err := s.pool.Exec(ctx, `
		DELETE FROM webhook_subscriptions WHERE id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("delete webhook subscription: %w", err)
	}
	return nil
}

// =============================================================================
// EXPIRATION MANAGEMENT
// =============================================================================

// GetExpiringSubscriptions returns subscriptions that will expire within the given duration.
func (s *WebhookSubscriptionService) GetExpiringSubscriptions(ctx context.Context, within time.Duration) ([]WebhookSubscription, error) {
	expiresBy := time.Now().Add(within)

	rows, err := s.pool.Query(ctx, `
		SELECT id, user_id, workspace_id, provider, resource_type,
			   external_subscription_id, webhook_url, events,
			   expires_at, status, event_count, created_at, updated_at
		FROM webhook_subscriptions
		WHERE status = 'active'
		  AND expires_at IS NOT NULL
		  AND expires_at <= $1
		ORDER BY expires_at ASC
	`, expiresBy)
	if err != nil {
		return nil, fmt.Errorf("get expiring subscriptions: %w", err)
	}
	defer rows.Close()

	var subscriptions []WebhookSubscription
	for rows.Next() {
		var sub WebhookSubscription
		err := rows.Scan(
			&sub.ID, &sub.UserID, &sub.WorkspaceID, &sub.Provider, &sub.ResourceType,
			&sub.ExternalSubscriptionID, &sub.WebhookURL, &sub.Events,
			&sub.ExpiresAt, &sub.Status, &sub.EventCount, &sub.CreatedAt, &sub.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan expiring subscription: %w", err)
		}
		subscriptions = append(subscriptions, sub)
	}

	return subscriptions, nil
}

// RenewSubscription updates the expiration time of a subscription.
func (s *WebhookSubscriptionService) RenewSubscription(ctx context.Context, id uuid.UUID, newExternalID string, newExpiresAt time.Time) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE webhook_subscriptions
		SET external_subscription_id = $2,
			expires_at = $3,
			status = 'active',
			updated_at = NOW()
		WHERE id = $1
	`, id, newExternalID, newExpiresAt)
	if err != nil {
		return fmt.Errorf("renew webhook subscription: %w", err)
	}

	s.logger.Info("Webhook subscription renewed",
		slog.String("subscription_id", id.String()),
		slog.Time("new_expires_at", newExpiresAt),
	)

	return nil
}

// MarkExpired marks all expired subscriptions as expired.
func (s *WebhookSubscriptionService) MarkExpired(ctx context.Context) (int, error) {
	result, err := s.pool.Exec(ctx, `
		UPDATE webhook_subscriptions
		SET status = 'expired', updated_at = NOW()
		WHERE status = 'active'
		  AND expires_at IS NOT NULL
		  AND expires_at < NOW()
	`)
	if err != nil {
		return 0, fmt.Errorf("mark expired subscriptions: %w", err)
	}

	count := int(result.RowsAffected())
	if count > 0 {
		s.logger.Info("Marked webhook subscriptions as expired", slog.Int("count", count))
	}

	return count, nil
}

// =============================================================================
// STATISTICS
// =============================================================================

// SubscriptionStats contains statistics about webhook subscriptions.
type SubscriptionStats struct {
	TotalSubscriptions  int            `json:"total_subscriptions"`
	ActiveSubscriptions int            `json:"active_subscriptions"`
	FailedSubscriptions int            `json:"failed_subscriptions"`
	TotalEventsReceived int            `json:"total_events_received"`
	ByProvider          map[string]int `json:"by_provider"`
}

// GetStats returns statistics about webhook subscriptions for a user.
func (s *WebhookSubscriptionService) GetStats(ctx context.Context, userID uuid.UUID) (*SubscriptionStats, error) {
	stats := &SubscriptionStats{
		ByProvider: make(map[string]int),
	}

	// Get counts
	err := s.pool.QueryRow(ctx, `
		SELECT
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE status = 'active') as active,
			COUNT(*) FILTER (WHERE status = 'failed') as failed,
			COALESCE(SUM(event_count), 0) as total_events
		FROM webhook_subscriptions
		WHERE user_id = $1
	`, userID).Scan(
		&stats.TotalSubscriptions,
		&stats.ActiveSubscriptions,
		&stats.FailedSubscriptions,
		&stats.TotalEventsReceived,
	)
	if err != nil {
		return nil, fmt.Errorf("get subscription stats: %w", err)
	}

	// Get counts by provider
	rows, err := s.pool.Query(ctx, `
		SELECT provider, COUNT(*) as count
		FROM webhook_subscriptions
		WHERE user_id = $1
		GROUP BY provider
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("get provider stats: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var provider string
		var count int
		if err := rows.Scan(&provider, &count); err != nil {
			return nil, fmt.Errorf("scan provider stat: %w", err)
		}
		stats.ByProvider[provider] = count
	}

	return stats, nil
}
