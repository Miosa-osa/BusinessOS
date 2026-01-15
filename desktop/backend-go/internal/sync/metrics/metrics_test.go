package metrics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetMetrics(t *testing.T) {
	m1 := GetMetrics()
	m2 := GetMetrics()

	// Should return the same instance (singleton)
	assert.Same(t, m1, m2)
}

func TestPendingEventsCounter(t *testing.T) {
	m := GetMetrics()
	m.Reset()

	m.IncrementPendingEvents()
	assert.Equal(t, int64(1), m.OutboxPendingEvents)

	m.IncrementPendingEvents()
	assert.Equal(t, int64(2), m.OutboxPendingEvents)

	m.DecrementPendingEvents()
	assert.Equal(t, int64(1), m.OutboxPendingEvents)

	m.SetPendingEvents(10)
	assert.Equal(t, int64(10), m.OutboxPendingEvents)
}

func TestProcessingEventsCounter(t *testing.T) {
	m := GetMetrics()
	m.Reset()

	m.IncrementProcessingEvents()
	assert.Equal(t, int64(1), m.OutboxProcessingEvents)

	m.DecrementProcessingEvents()
	assert.Equal(t, int64(0), m.OutboxProcessingEvents)

	// Should not go negative
	m.DecrementProcessingEvents()
	assert.Equal(t, int64(0), m.OutboxProcessingEvents)
}

func TestCompletedEventsCounter(t *testing.T) {
	m := GetMetrics()
	m.Reset()

	beforeTime := time.Now()
	m.IncrementCompletedEvents()
	afterTime := time.Now()

	assert.Equal(t, int64(1), m.OutboxCompletedEvents)
	assert.True(t, m.LastProcessedAt.After(beforeTime) || m.LastProcessedAt.Equal(beforeTime))
	assert.True(t, m.LastProcessedAt.Before(afterTime) || m.LastProcessedAt.Equal(afterTime))
}

func TestFailedEventsCounter(t *testing.T) {
	m := GetMetrics()
	m.Reset()

	m.IncrementFailedEvents()
	assert.Equal(t, int64(1), m.OutboxFailedEvents)

	m.IncrementFailedEvents()
	assert.Equal(t, int64(2), m.OutboxFailedEvents)
}

func TestProcessingDuration(t *testing.T) {
	m := GetMetrics()
	m.Reset()

	// Record some durations
	m.RecordProcessingDuration(100 * time.Millisecond)
	m.RecordProcessingDuration(200 * time.Millisecond)
	m.RecordProcessingDuration(300 * time.Millisecond)

	// Check average
	avg := m.GetAverageProcessingDuration()
	assert.Equal(t, 200*time.Millisecond, avg)

	// Check percentiles
	p50 := m.GetPercentileProcessingDuration(50)
	assert.Equal(t, 200*time.Millisecond, p50)

	p95 := m.GetPercentileProcessingDuration(95)
	assert.Equal(t, 300*time.Millisecond, p95)
}

func TestProcessingDurationSlidingWindow(t *testing.T) {
	m := GetMetrics()
	m.Reset()

	// Record more than 1000 durations
	for i := 0; i < 1500; i++ {
		m.RecordProcessingDuration(time.Duration(i) * time.Millisecond)
	}

	// Should keep only the last 1000
	assert.Equal(t, 1000, len(m.ProcessingDuration))
}

func TestErrorCounter(t *testing.T) {
	m := GetMetrics()
	m.Reset()

	beforeTime := time.Now()
	m.IncrementError("network_error")
	afterTime := time.Now()

	assert.Equal(t, int64(1), m.GetErrorCount("network_error"))
	assert.Equal(t, int64(0), m.GetErrorCount("database_error"))
	assert.True(t, m.LastErrorAt.After(beforeTime) || m.LastErrorAt.Equal(beforeTime))
	assert.True(t, m.LastErrorAt.Before(afterTime) || m.LastErrorAt.Equal(afterTime))

	m.IncrementError("network_error")
	assert.Equal(t, int64(2), m.GetErrorCount("network_error"))

	m.IncrementError("database_error")
	assert.Equal(t, int64(1), m.GetErrorCount("database_error"))
}

func TestConflictCounter(t *testing.T) {
	m := GetMetrics()
	m.Reset()

	m.IncrementConflict("concurrent")
	assert.Equal(t, int64(1), m.GetConflictCount("concurrent"))

	m.IncrementConflict("concurrent")
	assert.Equal(t, int64(2), m.GetConflictCount("concurrent"))

	m.IncrementConflict("stale")
	assert.Equal(t, int64(1), m.GetConflictCount("stale"))
}

func TestGetSnapshot(t *testing.T) {
	m := GetMetrics()
	m.Reset()

	// Set up some metrics
	m.SetPendingEvents(5)
	m.IncrementProcessingEvents()
	m.IncrementCompletedEvents()
	m.IncrementFailedEvents()
	m.RecordProcessingDuration(100 * time.Millisecond)
	m.RecordProcessingDuration(200 * time.Millisecond)
	m.IncrementError("network_error")
	m.IncrementConflict("concurrent")

	snapshot := m.GetSnapshot()

	assert.Equal(t, int64(5), snapshot.OutboxPendingEvents)
	assert.Equal(t, int64(1), snapshot.OutboxProcessingEvents)
	assert.Equal(t, int64(1), snapshot.OutboxCompletedEvents)
	assert.Equal(t, int64(1), snapshot.OutboxFailedEvents)
	assert.Equal(t, int64(1), snapshot.ErrorsTotal["network_error"])
	assert.Equal(t, int64(1), snapshot.ConflictsTotal["concurrent"])
	assert.Equal(t, 150*time.Millisecond, snapshot.AvgProcessingDuration)
}

func TestReset(t *testing.T) {
	m := GetMetrics()
	m.Reset()

	// Set up some metrics
	m.IncrementPendingEvents()
	m.IncrementCompletedEvents()
	m.IncrementError("test_error")

	// Reset
	m.Reset()

	// Verify all counters are reset
	assert.Equal(t, int64(0), m.OutboxPendingEvents)
	assert.Equal(t, int64(0), m.OutboxCompletedEvents)
	assert.Equal(t, int64(0), m.GetErrorCount("test_error"))
	assert.Empty(t, m.ProcessingDuration)
}
