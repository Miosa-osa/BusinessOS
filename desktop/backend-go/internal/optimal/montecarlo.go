package optimal

import (
	"math"
	"math/rand/v2"
)

// MonteCarloResult holds the output of N Monte Carlo simulations over an
// impact report. BestCase/WorstCase/Expected are keyed by affected node name
// and hold the probability (0–1) that the node is impacted in that scenario.
type MonteCarloResult struct {
	Simulations int                `json:"simulations"`
	BestCase    map[string]float64 `json:"best_case"`
	WorstCase   map[string]float64 `json:"worst_case"`
	Expected    map[string]float64 `json:"expected"`
	Confidence  float64            `json:"confidence"` // 0–1
}

// Sample runs n Monte Carlo simulations over a SimulationResult.
//
// For each simulation and each affected node the node is considered "triggered"
// with probability equal to its Impact score. After all simulations:
//   - Expected  = fraction of runs the node was triggered
//   - BestCase  = Expected − 1σ  (floored at 0)
//   - WorstCase = Expected + 1σ  (capped at 1)
//   - Confidence = 1 − (mean std-dev across all nodes), so wider variance → lower confidence
//
// When n <= 0 it defaults to 1000 (matching the Elixir engine default).
// When result is nil or has no affected nodes an empty result is returned.
func Sample(result *SimulationResult, simulations int) *MonteCarloResult {
	if simulations <= 0 {
		simulations = 1000
	}
	if result == nil || len(result.AffectedNodes) == 0 {
		return &MonteCarloResult{
			Simulations: simulations,
			BestCase:    map[string]float64{},
			WorstCase:   map[string]float64{},
			Expected:    map[string]float64{},
			Confidence:  1.0,
		}
	}

	nodes := result.AffectedNodes
	// hitCount[i] tracks how many simulations triggered node i.
	hitCount := make([]int, len(nodes))

	for s := 0; s < simulations; s++ {
		for i, node := range nodes {
			if rand.Float64() < node.Impact {
				hitCount[i]++
			}
		}
	}

	best := make(map[string]float64, len(nodes))
	worst := make(map[string]float64, len(nodes))
	expected := make(map[string]float64, len(nodes))

	n := float64(simulations)
	var totalStddev float64

	for i, node := range nodes {
		p := float64(hitCount[i]) / n
		// Binomial standard deviation: σ = sqrt(p*(1-p)/n)
		stddev := math.Sqrt(p * (1 - p) / n)

		expected[node.Name] = p
		best[node.Name] = math.Max(0, p-stddev)
		worst[node.Name] = math.Min(1, p+stddev)
		totalStddev += stddev
	}

	// Confidence: inverse of mean normalised stddev.
	// Mean stddev is bounded by 0.5 (maximum for p=0.5), so we scale to [0,1].
	meanStddev := totalStddev / float64(len(nodes))
	confidence := 1.0 - (meanStddev / 0.5)
	if confidence < 0 {
		confidence = 0
	}
	if confidence > 1 {
		confidence = 1
	}

	return &MonteCarloResult{
		Simulations: simulations,
		BestCase:    best,
		WorstCase:   worst,
		Expected:    expected,
		Confidence:  confidence,
	}
}
