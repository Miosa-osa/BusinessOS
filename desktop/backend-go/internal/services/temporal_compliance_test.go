package services

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTemporalDailyDuration verifies 1 day = 86,400 seconds
func TestTemporalDailyDuration(t *testing.T) {
	t1 := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC)

	duration := t2.Sub(t1).Seconds()
	expected := 86400.0

	assert.Equal(t, expected, duration, "1 day should be exactly 86,400 seconds")
}

// TestTemporalQuarterlyDuration verifies 90 days = 7,776,000 seconds
func TestTemporalQuarterlyDuration(t *testing.T) {
	t1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC)

	duration := t2.Sub(t1).Seconds()
	expected := 90.0 * 86400.0 // 7,776,000 seconds

	assert.Equal(t, expected, duration, "90 days should be exactly 7,776,000 seconds")
}

// TestComplianceTemporalDailyVsQuarterlyDurationRatio verifies the ratio calculation
func TestComplianceTemporalDailyVsQuarterlyDurationRatio(t *testing.T) {
	// Create event timestamps for daily cadence (10 events, 1 day apart)
	baseTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	dailyTimestamps := make([]time.Time, 10)
	for i := 0; i < 10; i++ {
		dailyTimestamps[i] = baseTime.AddDate(0, 0, i)
	}

	// Create event timestamps for quarterly cadence (4 events, 90 days apart)
	quarterlyTimestamps := make([]time.Time, 4)
	for i := 0; i < 4; i++ {
		quarterlyTimestamps[i] = baseTime.AddDate(0, 0, i*90)
	}

	// Calculate intervals between consecutive timestamps
	dailyIntervals := calculateIntervals(dailyTimestamps)
	quarterlyIntervals := calculateIntervals(quarterlyTimestamps)

	// Calculate mean intervals
	meanDaily := calculateMean(dailyIntervals)
	meanQuarterly := calculateMean(quarterlyIntervals)

	// Calculate ratio
	ratio := meanQuarterly / meanDaily

	// Verify values
	assert.Equal(t, 86400.0, meanDaily, "Daily mean interval should be exactly 86,400 seconds")
	assert.Equal(t, 7776000.0, meanQuarterly, "Quarterly mean interval should be exactly 7,776,000 seconds")
	assert.Equal(t, 90.0, ratio, "Ratio should be exactly 90.0")

	// Verify consistency of intervals
	for _, interval := range dailyIntervals {
		assert.Equal(t, 86400.0, interval, "Each daily interval should be 86,400 seconds")
	}

	for _, interval := range quarterlyIntervals {
		assert.Equal(t, 7776000.0, interval, "Each quarterly interval should be 7,776,000 seconds")
	}
}

// TestTemporalCalculationPrecision tests numerical precision of temporal calculations
func TestTemporalCalculationPrecision(t *testing.T) {
	base := time.Date(2026, 2, 15, 12, 30, 45, 0, time.UTC)

	// Create 5 timestamps at exactly 86,400 second intervals
	timestamps := make([]time.Time, 5)
	for i := 0; i < 5; i++ {
		timestamps[i] = base.Add(time.Duration(i) * time.Second * 86400)
	}

	// Verify all consecutive intervals are exactly 86,400 seconds
	for i := 0; i < len(timestamps)-1; i++ {
		duration := timestamps[i+1].Sub(timestamps[i]).Seconds()
		assert.Equal(t, 86400.0, duration,
			"Interval %d should be 86,400 seconds, got %f", i, duration)
	}
}

// TestTemporalRatioCalculationDetailed tests detailed ratio calculation
func TestTemporalRatioCalculationDetailed(t *testing.T) {
	// Daily dataset: 6 events, 5 intervals
	dailyBase := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	dailyEvents := make([]time.Time, 6)
	for i := 0; i < 6; i++ {
		dailyEvents[i] = dailyBase.AddDate(0, 0, i)
	}

	dailyDurations := calculateIntervals(dailyEvents)
	sumDaily := 0.0
	for _, d := range dailyDurations {
		sumDaily += d
	}
	meanDaily := sumDaily / float64(len(dailyDurations))

	// Quarterly dataset: 4 events, 3 intervals (90 days each)
	quarterlyBase := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	quarterlyEvents := make([]time.Time, 4)
	for i := 0; i < 4; i++ {
		quarterlyEvents[i] = quarterlyBase.AddDate(0, 0, i*90)
	}

	quarterlyDurations := calculateIntervals(quarterlyEvents)
	sumQuarterly := 0.0
	for _, d := range quarterlyDurations {
		sumQuarterly += d
	}
	meanQuarterly := sumQuarterly / float64(len(quarterlyDurations))

	// Calculate ratio
	ratio := meanQuarterly / meanDaily

	// Verify
	assert.Equal(t, 86400.0, meanDaily)
	assert.Equal(t, 7776000.0, meanQuarterly)
	assert.Equal(t, 90.0, ratio)

	// Verify with tolerance for floating point comparison
	tolerance := 1e-9
	assert.InDelta(t, 90.0, ratio, tolerance)
}

// TestTemporalComplianceMultipleTraces tests ratio with multiple event traces
func TestTemporalComplianceMultipleTraces(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	// Daily trace: 8 events
	dailyTrace := make([]time.Time, 8)
	for i := 0; i < 8; i++ {
		dailyTrace[i] = base.AddDate(0, 0, i)
	}

	// Quarterly trace: 3 events
	quarterlyTrace := make([]time.Time, 3)
	for i := 0; i < 3; i++ {
		quarterlyTrace[i] = base.AddDate(0, 0, i*90)
	}

	// Calculate
	dailyIntervals := calculateIntervals(dailyTrace)
	quarterlyIntervals := calculateIntervals(quarterlyTrace)

	meanDaily := calculateMean(dailyIntervals)
	meanQuarterly := calculateMean(quarterlyIntervals)

	ratio := meanQuarterly / meanDaily

	// Verify
	require.Equal(t, len(dailyIntervals), 7)
	require.Equal(t, len(quarterlyIntervals), 2)

	assert.InDelta(t, 86400.0, meanDaily, 1e-9)
	assert.InDelta(t, 7776000.0, meanQuarterly, 1e-9)
	assert.InDelta(t, 90.0, ratio, 1e-9)
}

// TestTemporalEdgeCases tests edge cases in temporal calculations
func TestTemporalEdgeCases(t *testing.T) {
	// Single interval (minimum case for mean calculation)
	t1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)

	intervals := calculateIntervals([]time.Time{t1, t2})
	require.Len(t, intervals, 1)
	assert.Equal(t, 86400.0, intervals[0])

	mean := calculateMean(intervals)
	assert.Equal(t, 86400.0, mean)
}

// Helper function: calculate intervals in seconds between consecutive timestamps
func calculateIntervals(timestamps []time.Time) []float64 {
	if len(timestamps) < 2 {
		return []float64{}
	}

	intervals := make([]float64, len(timestamps)-1)
	for i := 0; i < len(timestamps)-1; i++ {
		intervals[i] = timestamps[i+1].Sub(timestamps[i]).Seconds()
	}
	return intervals
}

// Helper function: calculate mean of a slice of floats
func calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}

	sum := 0.0
	for _, v := range values {
		sum += v
	}

	return sum / float64(len(values))
}

// BenchmarkTemporalRatioCalculation benchmarks the ratio calculation
func BenchmarkTemporalRatioCalculation(b *testing.B) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	// Pre-generate timestamps
	dailyTimestamps := make([]time.Time, 100)
	for i := 0; i < 100; i++ {
		dailyTimestamps[i] = base.AddDate(0, 0, i)
	}

	quarterlyTimestamps := make([]time.Time, 10)
	for i := 0; i < 10; i++ {
		quarterlyTimestamps[i] = base.AddDate(0, 0, i*90)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dailyIntervals := calculateIntervals(dailyTimestamps)
		quarterlyIntervals := calculateIntervals(quarterlyTimestamps)

		meanDaily := calculateMean(dailyIntervals)
		meanQuarterly := calculateMean(quarterlyIntervals)

		_ = meanQuarterly / meanDaily
	}
}

// TestTemporalComplianceBoundaryValues tests boundary conditions
func TestTemporalComplianceBoundaryValues(t *testing.T) {
	// Exactly 86,400 seconds apart
	t1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	t2 := t1.Add(86400 * time.Second)

	diff := t2.Sub(t1).Seconds()
	assert.Equal(t, 86400.0, diff)

	// Exactly 7,776,000 seconds apart (90 days)
	t3 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	t4 := t3.Add(7776000 * time.Second)

	diff2 := t4.Sub(t3).Seconds()
	assert.Equal(t, 7776000.0, diff2)

	// Verify ratio
	ratio := diff2 / diff
	assert.Equal(t, 90.0, ratio)
}

// TestTemporalComplianceNumericalStability tests numerical stability in calculations
func TestTemporalComplianceNumericalStability(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	// Run calculation multiple times with same data
	var ratios []float64
	for run := 0; run < 10; run++ {
		dailyTimestamps := make([]time.Time, 10)
		for i := 0; i < 10; i++ {
			dailyTimestamps[i] = base.AddDate(0, 0, i)
		}

		quarterlyTimestamps := make([]time.Time, 4)
		for i := 0; i < 4; i++ {
			quarterlyTimestamps[i] = base.AddDate(0, 0, i*90)
		}

		dailyIntervals := calculateIntervals(dailyTimestamps)
		quarterlyIntervals := calculateIntervals(quarterlyTimestamps)

		meanDaily := calculateMean(dailyIntervals)
		meanQuarterly := calculateMean(quarterlyIntervals)

		ratios = append(ratios, meanQuarterly/meanDaily)
	}

	// All ratios should be identical
	for _, r := range ratios {
		assert.InDelta(t, 90.0, r, 1e-15)
	}

	// Verify no variance due to floating point errors
	variance := 0.0
	mean := 90.0
	for _, r := range ratios {
		variance += math.Pow(r-mean, 2)
	}
	variance /= float64(len(ratios))

	// Variance should be zero (or negligibly small)
	assert.InDelta(t, 0.0, variance, 1e-30)
}
