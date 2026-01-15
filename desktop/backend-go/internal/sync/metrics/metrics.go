package metrics

import (
	"sync"
	"time"
)

// Metrics tracks synchronization metrics for monitoring and alerting.
// This implements a basic in-memory metrics system that can be exposed
// via Prometheus or other monitoring systems.
//
// Note: For production, consider using prometheus/client_golang directly.
// This implementation provides a foundation that can be easily integrated.

var (
	instance *Metrics
	once     sync.Once
)

// Metrics holds synchronization metrics.
type Metrics struct {
	mu sync.RWMutex

	// OutboxPendingEvents tracks the number of events waiting to be processed
	OutboxPendingEvents int64

	// OutboxProcessingEvents tracks the number of events currently being processed
	OutboxProcessingEvents int64

	// OutboxCompletedEvents tracks the total number of successfully processed events
	OutboxCompletedEvents int64

	// OutboxFailedEvents tracks the total number of failed events
	OutboxFailedEvents int64

	// ProcessingDuration tracks event processing duration
	ProcessingDuration []time.Duration

	// ErrorsTotal tracks the total number of errors by type
	ErrorsTotal map[string]int64

	// ConflictsTotal tracks the total number of conflicts by type
	ConflictsTotal map[string]int64

	// LastProcessedAt tracks when the last event was successfully processed
	LastProcessedAt time.Time

	// LastErrorAt tracks when the last error occurred
	LastErrorAt time.Time
}

// GetMetrics returns the singleton metrics instance.
func GetMetrics() *Metrics {
	once.Do(func() {
		instance = &Metrics{
			ErrorsTotal:    make(map[string]int64),
			ConflictsTotal: make(map[string]int64),
		}
	})
	return instance
}

// IncrementPendingEvents increments the pending events counter.
func (m *Metrics) IncrementPendingEvents() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.OutboxPendingEvents++
}

// DecrementPendingEvents decrements the pending events counter.
func (m *Metrics) DecrementPendingEvents() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.OutboxPendingEvents > 0 {
		m.OutboxPendingEvents--
	}
}

// SetPendingEvents sets the pending events counter to a specific value.
// This is useful when polling the database for current pending count.
func (m *Metrics) SetPendingEvents(count int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.OutboxPendingEvents = count
}

// IncrementProcessingEvents increments the processing events counter.
func (m *Metrics) IncrementProcessingEvents() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.OutboxProcessingEvents++
}

// DecrementProcessingEvents decrements the processing events counter.
func (m *Metrics) DecrementProcessingEvents() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.OutboxProcessingEvents > 0 {
		m.OutboxProcessingEvents--
	}
}

// IncrementCompletedEvents increments the completed events counter.
func (m *Metrics) IncrementCompletedEvents() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.OutboxCompletedEvents++
	m.LastProcessedAt = time.Now()
}

// IncrementFailedEvents increments the failed events counter.
func (m *Metrics) IncrementFailedEvents() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.OutboxFailedEvents++
}

// RecordProcessingDuration records the duration of event processing.
// Durations are stored in a sliding window (last 1000 events).
func (m *Metrics) RecordProcessingDuration(duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ProcessingDuration = append(m.ProcessingDuration, duration)

	// Keep only the last 1000 durations
	if len(m.ProcessingDuration) > 1000 {
		m.ProcessingDuration = m.ProcessingDuration[len(m.ProcessingDuration)-1000:]
	}
}

// GetAverageProcessingDuration returns the average processing duration.
func (m *Metrics) GetAverageProcessingDuration() time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.ProcessingDuration) == 0 {
		return 0
	}

	var total time.Duration
	for _, d := range m.ProcessingDuration {
		total += d
	}

	return total / time.Duration(len(m.ProcessingDuration))
}

// GetPercentileProcessingDuration returns the Nth percentile processing duration.
// percentile should be between 0 and 100 (e.g., 50 for median, 95 for p95, 99 for p99).
func (m *Metrics) GetPercentileProcessingDuration(percentile int) time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.ProcessingDuration) == 0 {
		return 0
	}

	// Simple percentile calculation (can be optimized with proper sorting)
	sorted := make([]time.Duration, len(m.ProcessingDuration))
	copy(sorted, m.ProcessingDuration)

	// Bubble sort (acceptable for small datasets)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	index := (percentile * len(sorted)) / 100
	if index >= len(sorted) {
		index = len(sorted) - 1
	}

	return sorted[index]
}

// IncrementError increments the error counter for a specific error type.
func (m *Metrics) IncrementError(errorType string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ErrorsTotal[errorType]++
	m.LastErrorAt = time.Now()
}

// GetErrorCount returns the error count for a specific error type.
func (m *Metrics) GetErrorCount(errorType string) int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.ErrorsTotal[errorType]
}

// IncrementConflict increments the conflict counter for a specific conflict type.
func (m *Metrics) IncrementConflict(conflictType string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ConflictsTotal[conflictType]++
}

// GetConflictCount returns the conflict count for a specific conflict type.
func (m *Metrics) GetConflictCount(conflictType string) int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.ConflictsTotal[conflictType]
}

// GetSnapshot returns a snapshot of all metrics for reporting.
func (m *Metrics) GetSnapshot() Snapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Copy maps to avoid race conditions
	errorsTotal := make(map[string]int64, len(m.ErrorsTotal))
	for k, v := range m.ErrorsTotal {
		errorsTotal[k] = v
	}

	conflictsTotal := make(map[string]int64, len(m.ConflictsTotal))
	for k, v := range m.ConflictsTotal {
		conflictsTotal[k] = v
	}

	return Snapshot{
		OutboxPendingEvents:    m.OutboxPendingEvents,
		OutboxProcessingEvents: m.OutboxProcessingEvents,
		OutboxCompletedEvents:  m.OutboxCompletedEvents,
		OutboxFailedEvents:     m.OutboxFailedEvents,
		ErrorsTotal:            errorsTotal,
		ConflictsTotal:         conflictsTotal,
		AvgProcessingDuration:  m.GetAverageProcessingDuration(),
		P95ProcessingDuration:  m.GetPercentileProcessingDuration(95),
		P99ProcessingDuration:  m.GetPercentileProcessingDuration(99),
		LastProcessedAt:        m.LastProcessedAt,
		LastErrorAt:            m.LastErrorAt,
	}
}

// Reset resets all metrics to zero. Use with caution (mainly for testing).
func (m *Metrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.OutboxPendingEvents = 0
	m.OutboxProcessingEvents = 0
	m.OutboxCompletedEvents = 0
	m.OutboxFailedEvents = 0
	m.ProcessingDuration = nil
	m.ErrorsTotal = make(map[string]int64)
	m.ConflictsTotal = make(map[string]int64)
	m.LastProcessedAt = time.Time{}
	m.LastErrorAt = time.Time{}
}

// Snapshot represents a point-in-time snapshot of metrics.
type Snapshot struct {
	OutboxPendingEvents    int64             `json:"outbox_pending_events"`
	OutboxProcessingEvents int64             `json:"outbox_processing_events"`
	OutboxCompletedEvents  int64             `json:"outbox_completed_events"`
	OutboxFailedEvents     int64             `json:"outbox_failed_events"`
	ErrorsTotal            map[string]int64  `json:"errors_total"`
	ConflictsTotal         map[string]int64  `json:"conflicts_total"`
	AvgProcessingDuration  time.Duration     `json:"avg_processing_duration"`
	P95ProcessingDuration  time.Duration     `json:"p95_processing_duration"`
	P99ProcessingDuration  time.Duration     `json:"p99_processing_duration"`
	LastProcessedAt        time.Time         `json:"last_processed_at"`
	LastErrorAt            time.Time         `json:"last_error_at"`
}
