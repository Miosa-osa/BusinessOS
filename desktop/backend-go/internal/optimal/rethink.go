package optimal

import (
	"fmt"
	"strings"
)

// RethinkReport is the synthesis output produced when enough evidence has
// accumulated around a topic to justify a knowledge update.
type RethinkReport struct {
	Topic        string         `json:"topic"`
	Observations []Observation  `json:"observations"`
	SearchHits   []SearchResult `json:"search_hits"`
	Synthesis    string         `json:"synthesis"`
	Confidence   float64        `json:"confidence"`
}

// confidenceThreshold is the minimum confidence required to generate a synthesis.
// Mirrors the Elixir engine's >= 1.5 rule.
const confidenceThreshold = 1.5

// Rethink synthesizes observations and search hits into a knowledge update when
// enough evidence has accumulated for topic.
//
// Algorithm:
//  1. Load all observations whose content or category contains topic.
//  2. Search SQLite for related contexts matching topic.
//  3. Sum evidence: each observation = 0.5, each search hit = 0.5.
//  4. If total confidence >= 1.5, generate a synthesis summary.
//  5. Return the full report regardless (caller inspects Confidence to decide).
func Rethink(dbPath, nodesRoot, topic string) (*RethinkReport, error) {
	if topic == "" {
		return nil, fmt.Errorf("optimal/rethink: topic must not be empty")
	}

	report := &RethinkReport{Topic: topic}

	// Step 1: observations matching topic.
	allObs, err := ListObservations(dbPath, "")
	if err != nil {
		// Non-fatal: continue with empty observations.
		allObs = nil
	}
	lower := strings.ToLower(topic)
	for _, o := range allObs {
		if strings.Contains(strings.ToLower(o.Content), lower) ||
			strings.Contains(strings.ToLower(o.Category), lower) {
			report.Observations = append(report.Observations, o)
		}
	}

	// Step 2: search for related contexts.
	hits, err := SearchContexts(dbPath, topic, 10)
	if err != nil {
		// Non-fatal: continue with empty hits.
		hits = nil
	}
	report.SearchHits = hits

	// Step 3: accumulate confidence.
	conf := float64(len(report.Observations))*0.5 + float64(len(report.SearchHits))*0.5

	// Step 4: synthesise when threshold reached.
	if conf >= confidenceThreshold {
		report.Synthesis = buildSynthesis(topic, report.Observations, report.SearchHits)
	}
	report.Confidence = conf

	return report, nil
}

// buildSynthesis produces a structured synthesis paragraph from the collected
// evidence. It is intentionally lightweight — the agent is expected to further
// process this text, not treat it as a final document.
func buildSynthesis(topic string, obs []Observation, hits []SearchResult) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("## Synthesis: %s\n\n", topic))
	sb.WriteString(fmt.Sprintf("Evidence base: %d observations, %d context hits.\n\n", len(obs), len(hits)))

	if len(obs) > 0 {
		sb.WriteString("### Patterns from observations\n")
		for _, o := range obs {
			sb.WriteString(fmt.Sprintf("- [%s/%s] %s\n", o.Category, o.Source, o.Content))
		}
		sb.WriteString("\n")
	}

	if len(hits) > 0 {
		sb.WriteString("### Related contexts\n")
		for _, h := range hits {
			sb.WriteString(fmt.Sprintf("- %s — %s\n", h.Path, truncate(h.Abstract, 120)))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("### Recommended action\n")
	sb.WriteString(fmt.Sprintf(
		"Update context.md for nodes related to '%s'. Review patterns above and consolidate into standing rules.\n",
		topic,
	))

	return sb.String()
}

// truncate shortens s to at most n runes, appending "..." if truncated.
func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "..."
}
