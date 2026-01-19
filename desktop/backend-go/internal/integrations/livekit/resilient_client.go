package livekit

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// ResilientClient wraps the LiveKit client with error tracking and monitoring
// Note: Unlike OSA, LiveKit token generation is local (JWT signing) so we don't need
// complex circuit breaker logic. This provides a consistent interface and monitoring.
type ResilientClient struct {
	client  *Client
	logger  *slog.Logger
	metrics *Metrics
	mu      sync.RWMutex
}

// Metrics tracks LiveKit client performance
type Metrics struct {
	mu                 sync.RWMutex
	totalRequests      uint64
	successfulRequests uint64
	failedRequests     uint64
	lastError          error
	lastErrorTime      time.Time
}

// ResilientClientConfig holds configuration for the resilient client
type ResilientClientConfig struct {
	// LiveKit client configuration
	LiveKitConfig *Config

	// Logger
	Logger *slog.Logger
}

// NewResilientClient creates a new resilient LiveKit client
func NewResilientClient(config *ResilientClientConfig) (*ResilientClient, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if config.LiveKitConfig == nil {
		return nil, fmt.Errorf("LiveKit config cannot be nil")
	}

	logger := config.Logger
	if logger == nil {
		logger = slog.Default()
	}

	// Create base client
	client, err := NewClient(config.LiveKitConfig, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create LiveKit client: %w", err)
	}

	return &ResilientClient{
		client:  client,
		logger:  logger,
		metrics: &Metrics{},
	}, nil
}

// GenerateRoomToken generates a room token with error tracking
func (rc *ResilientClient) GenerateRoomToken(ctx context.Context, req *TokenRequest) (*TokenResponse, error) {
	rc.metrics.mu.Lock()
	rc.metrics.totalRequests++
	rc.metrics.mu.Unlock()

	resp, err := rc.client.GenerateRoomToken(ctx, req)

	rc.metrics.mu.Lock()
	if err != nil {
		rc.metrics.failedRequests++
		rc.metrics.lastError = err
		rc.metrics.lastErrorTime = time.Now()
		rc.logger.Error("failed to generate room token",
			"error", err,
			"user_id", req.UserID,
			"workspace_id", req.WorkspaceID,
		)
	} else {
		rc.metrics.successfulRequests++
	}
	rc.metrics.mu.Unlock()

	return resp, err
}

// ValidateToken validates a token with error tracking
func (rc *ResilientClient) ValidateToken(ctx context.Context, token string) (bool, error) {
	rc.metrics.mu.Lock()
	rc.metrics.totalRequests++
	rc.metrics.mu.Unlock()

	valid, err := rc.client.ValidateToken(ctx, token)

	rc.metrics.mu.Lock()
	if err != nil {
		rc.metrics.failedRequests++
		rc.metrics.lastError = err
		rc.metrics.lastErrorTime = time.Now()
	} else {
		rc.metrics.successfulRequests++
	}
	rc.metrics.mu.Unlock()

	return valid, err
}

// GetMetrics returns a snapshot of the metrics
func (rc *ResilientClient) GetMetrics() MetricsSnapshot {
	rc.metrics.mu.RLock()
	defer rc.metrics.mu.RUnlock()

	return MetricsSnapshot{
		TotalRequests:      rc.metrics.totalRequests,
		SuccessfulRequests: rc.metrics.successfulRequests,
		FailedRequests:     rc.metrics.failedRequests,
		LastError:          rc.metrics.lastError,
		LastErrorTime:      rc.metrics.lastErrorTime,
		SuccessRate:        rc.calculateSuccessRate(),
	}
}

// MetricsSnapshot is a point-in-time snapshot of metrics
type MetricsSnapshot struct {
	TotalRequests      uint64
	SuccessfulRequests uint64
	FailedRequests     uint64
	LastError          error
	LastErrorTime      time.Time
	SuccessRate        float64
}

// calculateSuccessRate calculates the success rate (must be called with lock held)
func (rc *ResilientClient) calculateSuccessRate() float64 {
	if rc.metrics.totalRequests == 0 {
		return 0.0
	}
	return float64(rc.metrics.successfulRequests) / float64(rc.metrics.totalRequests) * 100.0
}

// Config returns the underlying client configuration
func (rc *ResilientClient) Config() *Config {
	return rc.client.Config()
}

// HealthCheck performs a basic health check
func (rc *ResilientClient) HealthCheck(ctx context.Context) error {
	// For LiveKit, health check is just verifying the configuration is valid
	return rc.client.config.Validate()
}
