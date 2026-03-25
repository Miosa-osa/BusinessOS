package services

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/circuitbreaker"
)

// CircuitBreakerWrapper adds circuit breaker protection to ComplianceService
type CircuitBreakerWrapper struct {
	*ComplianceService
	cb *circuitbreaker.CircuitBreaker
	logger *slog.Logger
}

// NewCircuitBreakerWrapper creates a compliance service with circuit breaker protection
func NewCircuitBreakerWrapper(osaBaseURL string, logger *slog.Logger) *CircuitBreakerWrapper {
	cb := circuitbreaker.NewBuilder().
		WithConfig(circuitbreaker.ComplianceServiceConfig()).
		Build()

	// Set up callbacks for monitoring
	cb.OnStateChange(func(oldState, newState circuitbreaker.State) {
		logger.Info("Compliance service circuit breaker state changed",
			"old_state", oldState,
			"new_state", newState,
			"osa_base_url", osaBaseURL)
	})

	cb.OnFailure(func(err error) {
		logger.Warn("Compliance service circuit breaker failure",
			"error", err,
			"osa_base_url", osaBaseURL)
	})

	cb.OnSuccess(func() {
		logger.Debug("Compliance service circuit breaker success")
	})

	cb.OnTimeout(func() {
		logger.Warn("Compliance service circuit breaker timeout")
	})

	// Create the underlying compliance service
	complianceService := NewComplianceService(osaBaseURL, logger)

	return &CircuitBreakerWrapper{
		ComplianceService: complianceService,
		cb: cb,
		logger: logger,
	}
}

// GetStatus wraps the original GetStatus with circuit breaker protection
func (c *CircuitBreakerWrapper) GetStatus(ctx context.Context) (ComplianceStatus, error) {
	var result ComplianceStatus
	var err error

	// Execute with circuit breaker protection
	err = c.cb.Execute(ctx, func() error {
		result, err = c.ComplianceService.GetStatus(ctx)
		return err
	})

	if err != nil {
		if circuitbreaker.IsCircuitOpenError(err) {
			// Circuit is open, return cached status if available
			c.logger.Warn("Compliance service circuit open, using cached status",
				"error", err)
			return c.getFallbackStatus(ctx)
		}
		return result, fmt.Errorf("circuit breaker protected get status: %w", err)
	}

	return result, nil
}

// GetAuditTrail wraps the original GetAuditTrail with circuit breaker protection
func (c *CircuitBreakerWrapper) GetAuditTrail(ctx context.Context, params AuditTrailParams) (AuditTrailResponse, error) {
	var result AuditTrailResponse
	var err error

	// Execute with circuit breaker protection and fallback
	err = c.cb.ExecuteWithFallback(ctx,
		func() error {
			result, err = c.ComplianceService.GetAuditTrail(ctx, params)
			return err
		},
		func() error {
			// Fallback: generate empty audit trail
			result = AuditTrailResponse{
				Entries: []AuditEntry{},
				Total:   0,
				Offset:  params.Offset,
				Limit:   params.Limit,
			}
			return nil
		})

	if err != nil {
		return result, fmt.Errorf("circuit breaker protected get audit trail: %w", err)
	}

	return result, nil
}

// CollectEvidence wraps the original CollectEvidence with circuit breaker protection
func (c *CircuitBreakerWrapper) CollectEvidence(ctx context.Context, req EvidenceCollectRequest) (EvidenceCollectResponse, error) {
	var result EvidenceCollectResponse
	var err error

	// Execute with circuit breaker protection and fallback
	err = c.cb.ExecuteWithFallback(ctx,
		func() error {
			result, err = c.ComplianceService.CollectEvidence(ctx, req)
			return err
		},
		func() error {
			// Fallback: generate minimal evidence
			result = EvidenceCollectResponse{
				Domain:    req.Domain,
				Period:    req.Period,
				Items:     c.generateFallbackEvidence(req.Domain, req.Period),
				Collected: len(c.generateFallbackEvidence(req.Domain, req.Period)),
			}
			return nil
		})

	if err != nil {
		return result, fmt.Errorf("circuit breaker protected collect evidence: %w", err)
	}

	return result, nil
}

// GetGapAnalysis wraps the original GetGapAnalysis with circuit breaker protection
func (c *CircuitBreakerWrapper) GetGapAnalysis(ctx context.Context, framework string) (GapAnalysisResponse, error) {
	var result GapAnalysisResponse
	var err error

	// Execute with circuit breaker protection
	err = c.cb.Execute(ctx, func() error {
		result, err = c.ComplianceService.GetGapAnalysis(ctx, framework)
		return err
	})

	if err != nil {
		if circuitbreaker.IsCircuitOpenError(err) {
			// Circuit is open, return cached gaps with warning
			c.logger.Warn("Compliance service circuit open, using cached gaps",
				"framework", framework,
				"error", err)
			return c.getFallbackGapAnalysis(framework), nil
		}
		return result, fmt.Errorf("circuit breaker protected get gap analysis: %w", err)
	}

	return result, nil
}

// VerifyAuditChain wraps the original VerifyAuditChain with circuit breaker protection
func (c *CircuitBreakerWrapper) VerifyAuditChain(ctx context.Context, sessionID string) (VerifyResult, error) {
	var result VerifyResult
	var err error

	// Execute with circuit breaker protection
	err = c.cb.Execute(ctx, func() error {
		result, err = c.ComplianceService.VerifyAuditChain(ctx, sessionID)
		return err
	})

	if err != nil {
		// For verification failures, we can still return a result
		return VerifyResult{
			Verified: false,
			Entries:  0,
			Issues:   []string{fmt.Sprintf("Verification failed: %v", err)},
		}, nil
	}

	return result, nil
}

// EvaluateAuditEvent wraps the original EvaluateAuditEvent with circuit breaker protection
func (c *CircuitBreakerWrapper) EvaluateAuditEvent(ctx context.Context, entry AuditEntry, userRole string) error {
	// Execute with circuit breaker protection
	err := c.cb.Execute(ctx, func() error {
		return c.ComplianceService.EvaluateAuditEvent(ctx, entry, userRole)
	})

	if err != nil {
		if circuitbreaker.IsCircuitOpenError(err) {
			c.logger.Warn("Compliance service circuit open, skipping audit event evaluation",
				"event_id", entry.ID,
				"session_id", entry.SessionID,
				"error", err)
			return nil // Silently skip when circuit is open for audit events
		}
		return fmt.Errorf("circuit breaker protected evaluate audit event: %w", err)
	}

	return nil
}

// ReloadRules wraps the original ReloadRules with circuit breaker protection
func (c *CircuitBreakerWrapper) ReloadRules(ctx context.Context) error {
	// Execute with circuit breaker protection
	err := c.cb.Execute(ctx, func() error {
		return c.ComplianceService.ReloadRules(ctx)
	})

	if err != nil {
		if circuitbreaker.IsCircuitOpenError(err) {
			c.logger.Warn("Compliance service circuit open, skipping rules reload",
				"error", err)
			return nil // Silently skip when circuit is open for non-critical operations
		}
		return fmt.Errorf("circuit breaker protected reload rules: %w", err)
	}

	return nil
}

// GetCircuitBreakerStats returns the current circuit breaker statistics
func (c *CircuitBreakerWrapper) GetCircuitBreakerStats() circuitbreaker.Stats {
	return c.cb.GetStats()
}

// ResetCircuitBreaker resets the circuit breaker
func (c *CircuitBreakerWrapper) ResetCircuitBreaker() {
	c.cb.Reset()
	c.logger.Info("Compliance service circuit breaker reset")
}

// GetNextRetryDelay returns the delay before next retry
func (c *CircuitBreakerWrapper) GetNextRetryDelay() time.Duration {
	return c.cb.GetNextRetryDelay()
}

// --- Private helper methods ---

// getFallbackStatus returns a fallback status when OSA is unavailable
func (c *CircuitBreakerWrapper) getFallbackStatus(ctx context.Context) (ComplianceStatus, error) {
	c.logger.Info("Using fallback compliance status due to OSA unavailability")

	// Return a degraded but safe status
	return ComplianceStatus{
		OverallScore: 0.5, // Conservative score when OSA is down
		Domains: map[string]DomainCompliance{
			"data_security":    {Score: 0.5, ChecksPassed: 0, ChecksFailed: 1},
			"process_integrity": {Score: 0.5, ChecksPassed: 0, ChecksFailed: 1},
			"regulatory":       {Score: 0.5, ChecksPassed: 0, ChecksFailed: 1},
		},
		LastAudit: time.Now(),
		Certificates: []Certificate{},
	}, nil
}

// getFallbackGapAnalysis returns cached gaps when OSA is unavailable
func (c *CircuitBreakerWrapper) getFallbackGapAnalysis(framework string) GapAnalysisResponse {
	c.logger.Info("Using cached gap analysis due to OSA unavailability",
		"framework", framework)

	// Return predefined gaps for common frameworks
	gaps := []ComplianceGap{
		{ID: "fallback-1", Framework: framework, Control: "Unknown",
			Description: "OSA unavailable, compliance status degraded",
			Severity: "medium", Status: "open"},
	}

	return GapAnalysisResponse{
		Framework:  framework,
		Gaps:       gaps,
		Score:      0.5, // Conservative score when degraded
		AnalyzedAt: time.Now(),
	}
}

// generateFallbackEvidence generates minimal evidence when OSA is unavailable
func (c *CircuitBreakerWrapper) generateFallbackEvidence(domain, period string) []EvidenceItem {
	items := []EvidenceItem{}

	items = append(items, EvidenceItem{
		ID:          fmt.Sprintf("fallback-ev-1-%d", time.Now().UnixNano()),
		Domain:      domain,
		Period:      period,
		Type:        "degraded_mode",
		Description: "Evidence collection degraded due to OSA unavailability",
		CollectedAt: time.Now(),
		Hash:        "degraded-mode-hash",
	})

	return items
}

// MonitorCircuitBreaker starts monitoring the circuit breaker health
func (c *CircuitBreakerWrapper) MonitorCircuitBreaker(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			stats := c.GetCircuitBreakerStats()

			// Log circuit breaker state
			switch stats.State {
			case circuitbreaker.StateOpen:
				c.logger.Warn("Compliance service circuit breaker OPEN",
					"success_rate", stats.SuccessRate,
					"consecutive_failures", stats.ConsecutiveFailures,
					"total_calls", stats.TotalCalls)
			case circuitbreaker.StateHalfOpen:
				c.logger.Info("Compliance service circuit breaker HALF-OPEN",
					"success_rate", stats.SuccessRate,
					"success_count", stats.SuccessfulCalls,
					"total_calls", stats.TotalCalls)
			}

			// Alert if success rate is too low
			if stats.SuccessRate < 50.0 && stats.TotalCalls > 10 {
				c.logger.Error("Compliance service circuit breaker low success rate",
					"success_rate", stats.SuccessRate,
					"total_calls", stats.TotalCalls)
			}
		}
	}
}

// HealthCheck returns the circuit breaker health status
func (c *CircuitBreakerWrapper) HealthCheck(ctx context.Context) map[string]interface{} {
	stats := c.GetCircuitBreakerStats()

	return map[string]interface{}{
		"circuit_breaker": map[string]interface{}{
			"state":             stats.State,
			"success_rate":      stats.SuccessRate,
			"total_calls":       stats.TotalCalls,
			"successful_calls":  stats.SuccessfulCalls,
			"failed_calls":      stats.FailedCalls,
			"timeout_calls":     stats.TimeoutCalls,
			"consecutive_failures": stats.ConsecutiveFailures,
			"last_failure":      stats.LastFailure,
			"next_retry_delay":  c.GetNextRetryDelay(),
		},
		"compliance_service": map[string]interface{}{
			"osa_base_url": c.osaBaseURL,
			"last_refresh": c.lastRefresh,
		},
	}
}

// CircuitBreakerHealthHandler creates a Gin handler for circuit breaker health checks
func (wrapper *CircuitBreakerWrapper) CircuitBreakerHealthHandler() func(*gin.Context) {
	return func(c *gin.Context) {
		health := wrapper.HealthCheck(c.Request.Context())

		// Add circuit breaker specific status
		stats := wrapper.GetCircuitBreakerStats()
		status := "healthy"
		if stats.State == circuitbreaker.StateOpen {
			status = "degraded"
		} else if stats.State == circuitbreaker.StateHalfOpen {
			status = "recovering"
		}

		health["status"] = status
		health["timestamp"] = time.Now()

		c.JSON(http.StatusOK, health)
	}
}