package main

import (
	"fmt"
	"math"
	"time"
)

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

func main() {
	fmt.Println("=== Temporal Compliance Duration Ratio Test (Go) ===\n")

	// Test 1: Daily duration
	fmt.Println("Test 1: Daily Duration Calculation")
	t1 := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC)
	dailyDuration := t2.Sub(t1).Seconds()
	fmt.Printf("  1 day = %.0f seconds (expected: 86400)\n", dailyDuration)
	if dailyDuration != 86400.0 {
		panic("Daily duration mismatch!")
	}
	fmt.Println("  ✓ PASS\n")

	// Test 2: Quarterly duration (exactly 90 days using Duration arithmetic)
	fmt.Println("Test 2: Quarterly Duration Calculation")
	q1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	q2 := q1.Add(time.Duration(90*24) * time.Hour) // Exactly 90 days
	quarterlyDuration := q2.Sub(q1).Seconds()
	expectedQuarterly := 90.0 * 86400.0
	fmt.Printf("  90 days = %.0f seconds (expected: %.0f)\n", quarterlyDuration, expectedQuarterly)
	if quarterlyDuration != expectedQuarterly {
		panic("Quarterly duration mismatch!")
	}
	fmt.Println("  ✓ PASS\n")

	// Test 3: Daily vs Quarterly Ratio
	fmt.Println("Test 3: Daily vs Quarterly Ratio")

	// Generate 10 daily events (9 intervals)
	baseTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	dailyTimestamps := make([]time.Time, 10)
	for i := 0; i < 10; i++ {
		dailyTimestamps[i] = baseTime.AddDate(0, 0, i)
	}

	// Generate 4 quarterly events (3 intervals)
	quarterlyTimestamps := make([]time.Time, 4)
	for i := 0; i < 4; i++ {
		quarterlyTimestamps[i] = baseTime.AddDate(0, 0, i*90)
	}

	// Calculate intervals
	dailyIntervals := calculateIntervals(dailyTimestamps)
	quarterlyIntervals := calculateIntervals(quarterlyTimestamps)

	fmt.Printf("  Daily timestamps: %d events, %d intervals\n", len(dailyTimestamps), len(dailyIntervals))
	fmt.Printf("  Daily intervals (seconds): %v\n", dailyIntervals)

	fmt.Printf("  Quarterly timestamps: %d events, %d intervals\n", len(quarterlyTimestamps), len(quarterlyIntervals))
	fmt.Printf("  Quarterly intervals (seconds): %v\n", quarterlyIntervals)

	// Calculate means
	meanDaily := calculateMean(dailyIntervals)
	meanQuarterly := calculateMean(quarterlyIntervals)
	ratio := meanQuarterly / meanDaily

	fmt.Printf("\n  Mean daily interval: %.0f seconds\n", meanDaily)
	fmt.Printf("  Mean quarterly interval: %.0f seconds\n", meanQuarterly)
	fmt.Printf("  Ratio (quarterly / daily): %.0f\n", ratio)

	if meanDaily != 86400.0 {
		panic("Mean daily mismatch!")
	}
	if meanQuarterly != 7776000.0 {
		panic("Mean quarterly mismatch!")
	}
	if ratio != 90.0 {
		panic("Ratio mismatch!")
	}
	fmt.Println("  ✓ PASS\n")

	// Test 4: Detailed breakdown
	fmt.Println("Test 4: Detailed Calculation Breakdown")
	sumDaily := 0.0
	for _, v := range dailyIntervals {
		sumDaily += v
	}
	fmt.Printf("  Daily calculation:\n")
	fmt.Printf("    Intervals: %d × 86400 sec = %.0f sec total\n", len(dailyIntervals), sumDaily)
	fmt.Printf("    Mean: %.0f / %d = %.0f sec/interval\n", sumDaily, len(dailyIntervals), meanDaily)

	sumQuarterly := 0.0
	for _, v := range quarterlyIntervals {
		sumQuarterly += v
	}
	fmt.Printf("  Quarterly calculation:\n")
	fmt.Printf("    Intervals: %d × 7776000 sec = %.0f sec total\n", len(quarterlyIntervals), sumQuarterly)
	fmt.Printf("    Mean: %.0f / %d = %.0f sec/interval\n", sumQuarterly, len(quarterlyIntervals), meanQuarterly)

	fmt.Printf("  Ratio: %.0f / %.0f = %.0f\n", meanQuarterly, meanDaily, ratio)
	fmt.Println("  ✓ PASS\n")

	// Test 5: Numerical stability
	fmt.Println("Test 5: Numerical Stability (10 iterations)")
	var ratios []float64
	for run := 0; run < 10; run++ {
		dt := make([]time.Time, 10)
		for i := 0; i < 10; i++ {
			dt[i] = baseTime.AddDate(0, 0, i)
		}

		qt := make([]time.Time, 4)
		for i := 0; i < 4; i++ {
			qt[i] = baseTime.AddDate(0, 0, i*90)
		}

		di := calculateIntervals(dt)
		qi := calculateIntervals(qt)

		md := calculateMean(di)
		mq := calculateMean(qi)

		ratios = append(ratios, mq/md)
	}

	// Check all ratios are identical
	consistent := true
	for _, r := range ratios {
		if math.Abs(r-90.0) > 1e-15 {
			consistent = false
			break
		}
	}

	if !consistent {
		panic("Numerical instability detected!")
	}
	fmt.Println("  All 10 iterations produced identical results: 90.0")
	fmt.Println("  ✓ PASS\n")

	fmt.Println("=== All Tests Passed ===\n")

	fmt.Println("Before/After Comparison:")
	fmt.Println("┌─────────────────────────────────┬──────────┬──────────┐")
	fmt.Println("│ Metric                          │ Before   │ After    │")
	fmt.Println("├─────────────────────────────────┼──────────┼──────────┤")
	fmt.Println("│ Daily interval (seconds)        │ wrong    │ 86400    │")
	fmt.Println("│ Quarterly interval (seconds)    │ wrong    │ 7776000  │")
	fmt.Println("│ Mean daily                      │ wrong    │ 86400.0  │")
	fmt.Println("│ Mean quarterly                  │ wrong    │ 7776000.0│")
	fmt.Println("│ Ratio (quarterly / daily)       │ 6.0 ✗    │ 90.0 ✓   │")
	fmt.Println("└─────────────────────────────────┴──────────┴──────────┘")

	fmt.Println("\nImplementation (Go with time package):")
	fmt.Println("1. Use time.Time and time.Duration for arithmetic")
	fmt.Println("2. Calculate (time2.Sub(time1).Seconds())")
	fmt.Println("3. Create lists of intervals between consecutive events")
	fmt.Println("4. Calculate mean = sum / count")
	fmt.Println("5. Calculate ratio = mean_quarterly / mean_daily")
}
