package chaos

import (
	"fmt"
	"math"
	"time"
)

// RecoveryMetrics tracks recovery time metrics across multiple test runs
type RecoveryMetrics struct {
	DetectionTimes []time.Duration
	RecoveryTimes  []time.Duration
	TotalIncidents int
}

// NewRecoveryMetrics creates a new metrics tracker
func NewRecoveryMetrics() *RecoveryMetrics {
	return &RecoveryMetrics{
		DetectionTimes: make([]time.Duration, 0),
		RecoveryTimes:  make([]time.Duration, 0),
	}
}

// RecordDetection records a detection time
func (rm *RecoveryMetrics) RecordDetection(duration time.Duration) {
	rm.DetectionTimes = append(rm.DetectionTimes, duration)
	rm.TotalIncidents++
}

// RecordRecovery records a recovery time
func (rm *RecoveryMetrics) RecordRecovery(duration time.Duration) {
	rm.RecoveryTimes = append(rm.RecoveryTimes, duration)
}

// GetMeanDetectionTime calculates the average detection time
func (rm *RecoveryMetrics) GetMeanDetectionTime() time.Duration {
	if len(rm.DetectionTimes) == 0 {
		return 0
	}

	var total time.Duration
	for _, d := range rm.DetectionTimes {
		total += d
	}
	return total / time.Duration(len(rm.DetectionTimes))
}

// GetMeanRecoveryTime calculates the average recovery time (MTTR)
func (rm *RecoveryMetrics) GetMeanRecoveryTime() time.Duration {
	if len(rm.RecoveryTimes) == 0 {
		return 0
	}

	var total time.Duration
	for _, d := range rm.RecoveryTimes {
		total += d
	}
	return total / time.Duration(len(rm.RecoveryTimes))
}

// GetMaxDetectionTime returns the maximum detection time
func (rm *RecoveryMetrics) GetMaxDetectionTime() time.Duration {
	if len(rm.DetectionTimes) == 0 {
		return 0
	}

	max := rm.DetectionTimes[0]
	for _, d := range rm.DetectionTimes {
		if d > max {
			max = d
		}
	}
	return max
}

// GetMaxRecoveryTime returns the maximum recovery time
func (rm *RecoveryMetrics) GetMaxRecoveryTime() time.Duration {
	if len(rm.RecoveryTimes) == 0 {
		return 0
	}

	max := rm.RecoveryTimes[0]
	for _, d := range rm.RecoveryTimes {
		if d > max {
			max = d
		}
	}
	return max
}

// GetStdDevDetectionTime calculates the standard deviation of detection times
func (rm *RecoveryMetrics) GetStdDevDetectionTime() time.Duration {
	if len(rm.DetectionTimes) < 2 {
		return 0
	}

	mean := rm.GetMeanDetectionTime()
	var sumSquares float64

	for _, d := range rm.DetectionTimes {
		diff := float64(d - mean)
		sumSquares += diff * diff
	}

	variance := sumSquares / float64(len(rm.DetectionTimes))
	stdDev := math.Sqrt(variance)

	return time.Duration(stdDev)
}

// GetStdDevRecoveryTime calculates the standard deviation of recovery times
func (rm *RecoveryMetrics) GetStdDevRecoveryTime() time.Duration {
	if len(rm.RecoveryTimes) < 2 {
		return 0
	}

	mean := rm.GetMeanRecoveryTime()
	var sumSquares float64

	for _, d := range rm.RecoveryTimes {
		diff := float64(d - mean)
		sumSquares += diff * diff
	}

	variance := sumSquares / float64(len(rm.RecoveryTimes))
	stdDev := math.Sqrt(variance)

	return time.Duration(stdDev)
}

// GetPercentileDetectionTime calculates the percentile of detection times
func (rm *RecoveryMetrics) GetPercentileDetectionTime(percentile float64) time.Duration {
	if len(rm.DetectionTimes) == 0 {
		return 0
	}

	sorted := make([]time.Duration, len(rm.DetectionTimes))
	copy(sorted, rm.DetectionTimes)

	// Simple bubble sort (good enough for test metrics)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	index := int(float64(len(sorted)) * percentile)
	if index >= len(sorted) {
		index = len(sorted) - 1
	}

	return sorted[index]
}

// GetPercentileRecoveryTime calculates the percentile of recovery times
func (rm *RecoveryMetrics) GetPercentileRecoveryTime(percentile float64) time.Duration {
	if len(rm.RecoveryTimes) == 0 {
		return 0
	}

	sorted := make([]time.Duration, len(rm.RecoveryTimes))
	copy(sorted, rm.RecoveryTimes)

	// Simple bubble sort
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	index := int(float64(len(sorted)) * percentile)
	if index >= len(sorted) {
		index = len(sorted) - 1
	}

	return sorted[index]
}

// String returns a formatted summary of metrics
func (rm *RecoveryMetrics) String() string {
	return fmt.Sprintf(
		`=== Recovery Metrics Summary ===
Total Incidents: %d

Detection Times:
  Mean:   %v
  Max:    %v
  StdDev: %v
  P95:    %v
  P99:    %v

Recovery Times (MTTR):
  Mean:   %v
  Max:    %v
  StdDev: %v
  P95:    %v
  P99:    %v
`,
		rm.TotalIncidents,
		rm.GetMeanDetectionTime(),
		rm.GetMaxDetectionTime(),
		rm.GetStdDevDetectionTime(),
		rm.GetPercentileDetectionTime(0.95),
		rm.GetPercentileDetectionTime(0.99),
		rm.GetMeanRecoveryTime(),
		rm.GetMaxRecoveryTime(),
		rm.GetStdDevRecoveryTime(),
		rm.GetPercentileRecoveryTime(0.95),
		rm.GetPercentileRecoveryTime(0.99),
	)
}

// PrintMetrics outputs metrics to stdout
func (rm *RecoveryMetrics) PrintMetrics() {
	fmt.Println(rm.String())
}

// ExportMetrics returns metrics as a map for JSON serialization
func (rm *RecoveryMetrics) ExportMetrics() map[string]interface{} {
	return map[string]interface{}{
		"total_incidents": rm.TotalIncidents,
		"detection": map[string]interface{}{
			"mean_ms":    rm.GetMeanDetectionTime().Milliseconds(),
			"max_ms":     rm.GetMaxDetectionTime().Milliseconds(),
			"stddev_ms":  rm.GetStdDevDetectionTime().Milliseconds(),
			"p95_ms":     rm.GetPercentileDetectionTime(0.95).Milliseconds(),
			"p99_ms":     rm.GetPercentileDetectionTime(0.99).Milliseconds(),
			"samples_ms": rm.durationSliceToMillis(rm.DetectionTimes),
		},
		"recovery": map[string]interface{}{
			"mean_ms":    rm.GetMeanRecoveryTime().Milliseconds(),
			"max_ms":     rm.GetMaxRecoveryTime().Milliseconds(),
			"stddev_ms":  rm.GetStdDevRecoveryTime().Milliseconds(),
			"p95_ms":     rm.GetPercentileRecoveryTime(0.95).Milliseconds(),
			"p99_ms":     rm.GetPercentileRecoveryTime(0.99).Milliseconds(),
			"samples_ms": rm.durationSliceToMillis(rm.RecoveryTimes),
		},
	}
}

func (rm *RecoveryMetrics) durationSliceToMillis(durations []time.Duration) []int64 {
	result := make([]int64, len(durations))
	for i, d := range durations {
		result[i] = d.Milliseconds()
	}
	return result
}
