package osa

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

// ResilientClient wraps the OSA client with circuit breaker, retry, and fallback
type ResilientClient struct {
	client              *Client
	circuitBreaker      *CircuitBreaker
	healthCheckCache    *HealthCheckCache
	fallbackClient      *FallbackClient
	requestQueue        *RequestQueue
	enableAutoRecovery  bool
}

// ResilientClientConfig holds configuration for the resilient client
type ResilientClientConfig struct {
	// OSA client configuration
	OSAConfig *Config

	// Circuit breaker configuration
	CircuitBreakerConfig *CircuitBreakerConfig

	// Fallback strategy
	FallbackStrategy FallbackStrategy

	// Cache TTL for responses
	CacheTTL time.Duration

	// Health check cache TTL
	HealthCheckCacheTTL time.Duration

	// Request queue size
	QueueSize int

	// Enable automatic recovery and retry
	EnableAutoRecovery bool
}

// DefaultResilientClientConfig returns sensible defaults
func DefaultResilientClientConfig() *ResilientClientConfig {
	return &ResilientClientConfig{
		OSAConfig:            DefaultConfig(),
		CircuitBreakerConfig: DefaultCircuitBreakerConfig(),
		FallbackStrategy:     FallbackStale,
		CacheTTL:             5 * time.Minute,
		HealthCheckCacheTTL:  30 * time.Second,
		QueueSize:            1000,
		EnableAutoRecovery:   true,
	}
}

// NewResilientClient creates a new resilient OSA client
func NewResilientClient(config *ResilientClientConfig) (*ResilientClient, error) {
	if config == nil {
		config = DefaultResilientClientConfig()
	}

	// Create base client
	client, err := NewClient(config.OSAConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create OSA client: %w", err)
	}

	// Create circuit breaker
	circuitBreaker := NewCircuitBreaker(config.CircuitBreakerConfig)

	// Create fallback client
	fallbackClient := NewFallbackClient(client, config.CacheTTL, config.FallbackStrategy)

	// Create request queue
	requestQueue := NewRequestQueue(config.QueueSize)

	// Create health check cache
	healthCheckCache := NewHealthCheckCache(config.HealthCheckCacheTTL, client.HealthCheck)

	resilientClient := &ResilientClient{
		client:             client,
		circuitBreaker:     circuitBreaker,
		healthCheckCache:   healthCheckCache,
		fallbackClient:     fallbackClient,
		requestQueue:       requestQueue,
		enableAutoRecovery: config.EnableAutoRecovery,
	}

	// Start auto-recovery if enabled
	if config.EnableAutoRecovery {
		go resilientClient.autoRecoveryLoop()
	}

	return resilientClient, nil
}

// GenerateApp generates an app with full resilience
func (r *ResilientClient) GenerateApp(ctx context.Context, req *AppGenerationRequest) (*AppGenerationResponse, error) {
	var resp *AppGenerationResponse
	var err error

	// Execute with circuit breaker
	err = r.circuitBreaker.Execute(ctx, func() error {
		// Execute with exponential backoff retry
		err = RetryWithBackoff(ctx, func() error {
			resp, err = r.client.GenerateApp(ctx, req)
			return err
		})
		return err
	})

	if err != nil {
		slog.Error("generate app failed after retries",
			"error", err,
			"circuit_state", r.circuitBreaker.State())

		// Try fallback
		resp, fallbackErr := r.fallbackClient.GenerateAppWithFallback(ctx, req)
		if fallbackErr != nil {
			// Queue the request if circuit is open
			if r.circuitBreaker.State() == StateOpen {
				queueID, queueErr := r.requestQueue.Enqueue("generate_app", req, req.UserID)
				if queueErr != nil {
					slog.Error("failed to queue request", "error", queueErr)
				} else {
					slog.Info("request queued for later processing", "queue_id", queueID)
				}
			}

			return nil, fmt.Errorf("all resilience strategies failed: primary=%w, fallback=%w", err, fallbackErr)
		}

		slog.Info("fallback successful for generate app")
		return resp, nil
	}

	// Cache successful response
	r.fallbackClient.cache.Set(
		r.fallbackClient.cacheKey("generate_app", req.UserID, req.WorkspaceID),
		resp,
	)

	return resp, nil
}

// GetAppStatus gets app status with full resilience
func (r *ResilientClient) GetAppStatus(ctx context.Context, appID string, userID uuid.UUID) (*AppStatusResponse, error) {
	var resp *AppStatusResponse
	var err error

	// Execute with circuit breaker
	err = r.circuitBreaker.Execute(ctx, func() error {
		// Execute with exponential backoff retry
		err = RetryWithBackoff(ctx, func() error {
			resp, err = r.client.GetAppStatus(ctx, appID, userID)
			return err
		})
		return err
	})

	if err != nil {
		slog.Error("get app status failed after retries",
			"app_id", appID,
			"error", err,
			"circuit_state", r.circuitBreaker.State())

		// Try fallback
		resp, fallbackErr := r.fallbackClient.GetAppStatusWithFallback(ctx, appID, userID)
		if fallbackErr != nil {
			return nil, fmt.Errorf("all resilience strategies failed: primary=%w, fallback=%w", err, fallbackErr)
		}

		slog.Info("fallback successful for get app status")
		return resp, nil
	}

	// Cache successful response
	cacheKey := r.fallbackClient.cacheKey("app_status", userID, uuid.Nil) + ":" + appID
	r.fallbackClient.cache.Set(cacheKey, resp)

	return resp, nil
}

// Orchestrate orchestrates with full resilience
func (r *ResilientClient) Orchestrate(ctx context.Context, req *OrchestrateRequest) (*OrchestrateResponse, error) {
	var resp *OrchestrateResponse
	var err error

	// Execute with circuit breaker
	err = r.circuitBreaker.Execute(ctx, func() error {
		// Execute with exponential backoff retry
		err = RetryWithBackoff(ctx, func() error {
			resp, err = r.client.Orchestrate(ctx, req)
			return err
		})
		return err
	})

	if err != nil {
		slog.Error("orchestrate failed after retries",
			"error", err,
			"circuit_state", r.circuitBreaker.State())

		// Try fallback
		resp, fallbackErr := r.fallbackClient.OrchestrateWithFallback(ctx, req)
		if fallbackErr != nil {
			// Queue the request if circuit is open
			if r.circuitBreaker.State() == StateOpen {
				queueID, queueErr := r.requestQueue.Enqueue("orchestrate", req, req.UserID)
				if queueErr != nil {
					slog.Error("failed to queue request", "error", queueErr)
				} else {
					slog.Info("request queued for later processing", "queue_id", queueID)
				}
			}

			return nil, fmt.Errorf("all resilience strategies failed: primary=%w, fallback=%w", err, fallbackErr)
		}

		slog.Info("fallback successful for orchestrate")
		return resp, nil
	}

	// Cache successful response
	r.fallbackClient.cache.Set(
		r.fallbackClient.cacheKey("orchestrate", req.UserID, req.WorkspaceID),
		resp,
	)

	return resp, nil
}

// GetWorkspaces gets workspaces with full resilience
func (r *ResilientClient) GetWorkspaces(ctx context.Context, userID uuid.UUID) (*WorkspacesResponse, error) {
	var resp *WorkspacesResponse
	var err error

	// Execute with circuit breaker
	err = r.circuitBreaker.Execute(ctx, func() error {
		// Execute with exponential backoff retry
		err = RetryWithBackoff(ctx, func() error {
			resp, err = r.client.GetWorkspaces(ctx, userID)
			return err
		})
		return err
	})

	if err != nil {
		slog.Error("get workspaces failed after retries",
			"error", err,
			"circuit_state", r.circuitBreaker.State())

		// Try fallback
		resp, fallbackErr := r.fallbackClient.GetWorkspacesWithFallback(ctx, userID)
		if fallbackErr != nil {
			return nil, fmt.Errorf("all resilience strategies failed: primary=%w, fallback=%w", err, fallbackErr)
		}

		slog.Info("fallback successful for get workspaces")
		return resp, nil
	}

	// Cache successful response
	r.fallbackClient.cache.Set(
		r.fallbackClient.cacheKey("workspaces", userID, uuid.Nil),
		resp,
	)

	return resp, nil
}

// HealthCheck performs a health check with caching
func (r *ResilientClient) HealthCheck(ctx context.Context) (*HealthResponse, error) {
	return r.healthCheckCache.Check(ctx)
}

// autoRecoveryLoop attempts to process queued requests when circuit recovers
func (r *ResilientClient) autoRecoveryLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// Only process queue if circuit is closed
		if r.circuitBreaker.State() != StateClosed {
			continue
		}

		// Check if service is healthy
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_, err := r.HealthCheck(ctx)
		cancel()

		if err != nil {
			continue
		}

		// Process queued requests
		for r.requestQueue.Size() > 0 {
			req, ok := r.requestQueue.Dequeue()
			if !ok {
				break
			}

			slog.Info("processing queued request",
				"request_id", req.ID,
				"operation", req.Operation,
				"queued_at", req.QueuedAt)

			// TODO: Implement actual request processing
			// This would unmarshal the payload and call the appropriate method
		}
	}
}

// Metrics returns circuit breaker metrics
func (r *ResilientClient) Metrics() CircuitMetrics {
	return r.circuitBreaker.Metrics()
}

// State returns the current circuit breaker state
func (r *ResilientClient) State() CircuitState {
	return r.circuitBreaker.State()
}

// QueueSize returns the current request queue size
func (r *ResilientClient) QueueSize() int {
	return r.requestQueue.Size()
}

// InvalidateHealthCache invalidates the health check cache
func (r *ResilientClient) InvalidateHealthCache() {
	r.healthCheckCache.Invalidate()
}

// InvalidateCache invalidates all response caches
func (r *ResilientClient) InvalidateCache() {
	r.fallbackClient.cache.Clear()
}

// ResetCircuitBreaker resets the circuit breaker to closed state
func (r *ResilientClient) ResetCircuitBreaker() {
	r.circuitBreaker.Reset()
	slog.Info("circuit breaker manually reset")
}

// Close closes the client and cleans up resources
func (r *ResilientClient) Close() error {
	return r.client.Close()
}
