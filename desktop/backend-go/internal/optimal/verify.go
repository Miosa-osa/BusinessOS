package optimal

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"
)

// VerifyResult is the aggregate outcome of an L0 fidelity test run.
type VerifyResult struct {
	Sampled int            `json:"sampled"`
	Passed  int            `json:"passed"`
	Failed  int            `json:"failed"`
	Score   float64        `json:"score"` // 0.0–1.0
	Details []VerifyDetail `json:"details"`
}

// VerifyDetail is a single context sample with its fidelity score.
type VerifyDetail struct {
	Path   string  `json:"path"`
	Title  string  `json:"title"`
	L0     string  `json:"l0"`
	Score  float64 `json:"score"`
	Status string  `json:"status"` // "pass" or "fail"
}

// Verify samples sampleSize random contexts from the SQLite index and checks
// whether each context's L0 abstract faithfully represents the full content
// using Jaccard keyword similarity. Scores above 0.3 pass.
func Verify(dbPath string, sampleSize int) (*VerifyResult, error) {
	if sampleSize <= 0 {
		sampleSize = 10
	}

	db, err := openDB(dbPath)
	if err != nil {
		return nil, fmt.Errorf("optimal/verify: open db: %w", err)
	}
	defer db.Close()

	// Fetch all rows then sample client-side — SQLite's RANDOM() ORDER can be slow on FTS tables.
	rows, err := db.Query(`SELECT COALESCE(path,''), title, content FROM contexts WHERE content != '' LIMIT 500`)
	if err != nil {
		return nil, fmt.Errorf("optimal/verify: query contexts: %w", err)
	}
	defer rows.Close()

	type row struct {
		path, title, content string
	}
	var pool []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.path, &r.title, &r.content); err != nil {
			return nil, fmt.Errorf("optimal/verify: scan row: %w", err)
		}
		pool = append(pool, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("optimal/verify: iterate: %w", err)
	}

	if len(pool) == 0 {
		return &VerifyResult{Score: 1.0}, nil
	}

	// Reservoir sample without replacement.
	if sampleSize > len(pool) {
		sampleSize = len(pool)
	}
	indices := rand.Perm(len(pool))[:sampleSize]

	result := &VerifyResult{Sampled: sampleSize}
	for _, idx := range indices {
		r := pool[idx]
		l0 := L0Abstract(r.content)
		score := jaccardKeywords(l0, r.content)

		status := "pass"
		if score < 0.3 {
			status = "fail"
			result.Failed++
		} else {
			result.Passed++
		}

		result.Details = append(result.Details, VerifyDetail{
			Path:   r.path,
			Title:  r.title,
			L0:     l0,
			Score:  score,
			Status: status,
		})
	}

	if result.Sampled > 0 {
		result.Score = float64(result.Passed) / float64(result.Sampled)
	}
	return result, nil
}

// ── keyword helpers ───────────────────────────────────────────────────────────

var reWord = regexp.MustCompile(`[a-zA-Z]{3,}`)

// extractKeywords returns a deduplicated lowercase set of words ≥3 chars.
func extractKeywords(text string) map[string]struct{} {
	words := reWord.FindAllString(strings.ToLower(text), -1)
	set := make(map[string]struct{}, len(words))
	for _, w := range words {
		set[w] = struct{}{}
	}
	return set
}

// jaccardKeywords computes Jaccard similarity between keyword sets of a and b.
// Returns 1.0 when both are empty (identical empty sets).
func jaccardKeywords(a, b string) float64 {
	setA := extractKeywords(a)
	setB := extractKeywords(b)

	if len(setA) == 0 && len(setB) == 0 {
		return 1.0
	}

	var inter int
	for k := range setA {
		if _, ok := setB[k]; ok {
			inter++
		}
	}
	union := len(setA) + len(setB) - inter
	if union == 0 {
		return 1.0
	}
	return float64(inter) / float64(union)
}
