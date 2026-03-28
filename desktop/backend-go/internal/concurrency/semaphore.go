// Package concurrency provides admission control primitives for BusinessOS.
// Implements global concurrency limiting with metrics integration.
package concurrency

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rhl/businessos-backend/internal/observability"
)

// Semaphore provides global concurrency control with metrics.
// WvdA compliant: all Acquire() calls have 5s timeout.
type Semaphore struct {
	mu              sync.Mutex
	slots           chan struct{} // Buffered channel acts as semaphore slots
	maxSlots        int32
	acquireTimeout  time.Duration
	telemetry       *observability.Telemetry

	// Metrics (atomic for lock-free reads)
	totalRequests   atomic.Int64
	totalAcquired   atomic.Int64
	totalRejected   atomic.Int64
	totalReleased   atomic.Int64
}

// New creates a new Semaphore with the specified maximum concurrent slots.
// Default timeout is 5 seconds (WvdA compliant).
func New(maxSlots int, tel *observability.Telemetry) *Semaphore {
	if maxSlots <= 0 {
		maxSlots = 200 // Default to 200 slots if invalid
		slog.Warn("Semaphore: invalid maxSlots, using default", "maxSlots", maxSlots, "default", 200)
	}

	sem := &Semaphore{
		slots:          make(chan struct{}, maxSlots),
		maxSlots:       int32(maxSlots),
		acquireTimeout: 5 * time.Second, // WvdA compliant: bounded wait
		telemetry:      tel,
	}

	// Pre-fill slots
	for i := 0; i < maxSlots; i++ {
		sem.slots <- struct{}{}
	}

	slog.Info("Semaphore: initialized", "max_slots", maxSlots, "timeout", sem.acquireTimeout)
	return sem
}

// Acquire attempts to acquire a semaphore slot.
// Returns error if timeout is exceeded (5s default).
// WvdA compliant: bounded wait with explicit timeout.
func (s *Semaphore) Acquire(ctx context.Context) error {
	s.totalRequests.Add(1)

	// Record attempt
	if s.telemetry != nil {
		s.telemetry.RecordMetric("semaphore.acquire.attempts", 1, map[string]string{
			"max_slots": fmt.Sprint(s.maxSlots),
		}, "counter")
	}

	// WvdA compliance: bounded wait with timeout
	ctx, cancel := context.WithTimeout(ctx, s.acquireTimeout)
	defer cancel()

	select {
	case <-s.slots:
		// Slot acquired
		s.totalAcquired.Add(1)

		// Record success
		if s.telemetry != nil {
			s.telemetry.RecordMetric("semaphore.acquire.success", 1, nil, "counter")
		}

		slog.Debug("Semaphore: slot acquired", "available", s.Available())
		return nil

	case <-ctx.Done():
		// Timeout exceeded
		s.totalRejected.Add(1)

		// Record rejection
		if s.telemetry != nil {
			s.telemetry.RecordMetric("semaphore.acquire.rejected", 1, map[string]string{
				"reason": "timeout",
			}, "counter")
		}

		slog.Warn("Semaphore: acquire timeout", "timeout", s.acquireTimeout, "available", s.Available())
		return ctx.Err()
	}
}

// Release returns a semaphore slot back to the pool.
// Must be called after successful Acquire() (typically in defer).
func (s *Semaphore) Release() {
	select {
	case s.slots <- struct{}{}:
		s.totalReleased.Add(1)

		// Record release
		if s.telemetry != nil {
			s.telemetry.RecordMetric("semaphore.release", 1, nil, "counter")
		}

		slog.Debug("Semaphore: slot released", "available", s.Available())
	default:
		// Channel full: release without slot (should not happen in correct usage)
		slog.Error("Semaphore: release without acquire detected", "available", s.Available())
	}
}

// Available returns the number of currently available slots.
func (s *Semaphore) Available() int {
	return len(s.slots)
}

// Utilization returns the current utilization as a percentage (0-100).
func (s *Semaphore) Utilization() float64 {
	used := s.maxSlots - int32(len(s.slots))
	return (float64(used) / float64(s.maxSlots)) * 100.0
}

// RejectionRate returns the rejection rate as a percentage (0-100).
func (s *Semaphore) RejectionRate() float64 {
	total := s.totalRequests.Load()
	if total == 0 {
		return 0.0
	}
	rejected := s.totalRejected.Load()
	return (float64(rejected) / float64(total)) * 100.0
}

// RecordMetricsSnapshot records current metrics to telemetry.
// Call this periodically (e.g., every 10s) for monitoring dashboards.
func (s *Semaphore) RecordMetricsSnapshot() {
	if s.telemetry == nil {
		return
	}

	s.telemetry.RecordMetric("semaphore.utilization_percent", s.Utilization(), nil, "gauge")
	s.telemetry.RecordMetric("semaphore.rejection_rate_percent", s.RejectionRate(), nil, "gauge")
	s.telemetry.RecordMetric("semaphore.available_slots", float64(s.Available()), nil, "gauge")
	s.telemetry.RecordMetric("semaphore.total_requests", float64(s.totalRequests.Load()), nil, "counter")
}

// Stats returns current statistics for health checks.
type Stats struct {
	MaxSlots      int32   `json:"max_slots"`
	Available     int     `json:"available"`
	Utilization   float64 `json:"utilization_percent"`
	TotalRequests int64   `json:"total_requests"`
	TotalAcquired int64   `json:"total_acquired"`
	TotalRejected int64   `json:"total_rejected"`
	TotalReleased int64   `json:"total_released"`
	RejectionRate float64 `json:"rejection_rate_percent"`
}

// GetStats returns a snapshot of current statistics.
func (s *Semaphore) GetStats() Stats {
	return Stats{
		MaxSlots:      s.maxSlots,
		Available:     s.Available(),
		Utilization:   s.Utilization(),
		TotalRequests: s.totalRequests.Load(),
		TotalAcquired: s.totalAcquired.Load(),
		TotalRejected: s.totalRejected.Load(),
		TotalReleased: s.totalReleased.Load(),
		RejectionRate: s.RejectionRate(),
	}
}
