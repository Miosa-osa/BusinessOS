package optimal

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
)

// StoreEmbedding saves a float32 embedding vector for the given contextID.
// The embedding is serialised as JSON and stored in the embeddings column of
// the contexts table. The column is added if missing.
func StoreEmbedding(dbPath, contextID string, embedding []float32) error {
	if dbPath == "" {
		return fmt.Errorf("optimal/vector: dbPath must not be empty")
	}
	if contextID == "" {
		return fmt.Errorf("optimal/vector: contextID must not be empty")
	}
	if len(embedding) == 0 {
		return fmt.Errorf("optimal/vector: embedding must not be empty")
	}

	db, err := openDB(dbPath)
	if err != nil {
		return fmt.Errorf("optimal/vector: open db: %w", err)
	}
	defer db.Close()

	// Ensure the embeddings column exists — added lazily so we don't require a migration.
	_, _ = db.Exec(`ALTER TABLE contexts ADD COLUMN embeddings TEXT`)

	blob, err := json.Marshal(embedding)
	if err != nil {
		return fmt.Errorf("optimal/vector: marshal embedding: %w", err)
	}

	_, err = db.Exec(`UPDATE contexts SET embeddings = ? WHERE id = ?`, string(blob), contextID)
	if err != nil {
		return fmt.Errorf("optimal/vector: store embedding: %w", err)
	}
	return nil
}

// SearchByVector finds the top limit contexts whose stored embeddings are most
// similar to queryEmbedding using cosine similarity. This is an O(n) scan —
// acceptable for under 50 K contexts.
func SearchByVector(dbPath string, queryEmbedding []float32, limit int) ([]SearchResult, error) {
	if dbPath == "" {
		return nil, fmt.Errorf("optimal/vector: dbPath must not be empty")
	}
	if len(queryEmbedding) == 0 {
		return nil, fmt.Errorf("optimal/vector: queryEmbedding must not be empty")
	}
	if limit <= 0 {
		limit = 10
	}

	db, err := openDB(dbPath)
	if err != nil {
		return nil, fmt.Errorf("optimal/vector: open db: %w", err)
	}
	defer db.Close()

	rows, err := db.Query(`SELECT COALESCE(path,''), title, COALESCE(embeddings,'') FROM contexts WHERE embeddings != ''`)
	if err != nil {
		return nil, fmt.Errorf("optimal/vector: query embeddings: %w", err)
	}
	defer rows.Close()

	type scored struct {
		path  string
		title string
		score float64
	}
	var candidates []scored

	for rows.Next() {
		var path, title, blob string
		if err := rows.Scan(&path, &title, &blob); err != nil {
			continue
		}
		if blob == "" {
			continue
		}
		var vec []float32
		if err := json.Unmarshal([]byte(blob), &vec); err != nil {
			continue
		}
		sim := CosineSimilarity(queryEmbedding, vec)
		candidates = append(candidates, scored{path: path, title: title, score: sim})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("optimal/vector: iterate: %w", err)
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})
	if len(candidates) > limit {
		candidates = candidates[:limit]
	}

	results := make([]SearchResult, 0, len(candidates))
	for _, c := range candidates {
		results = append(results, SearchResult{
			Path:     c.path,
			Abstract: c.title,
			Score:    c.score,
		})
	}
	return results, nil
}

// CosineSimilarity returns the cosine similarity in [−1, 1] between vectors a
// and b. Returns 0 when either vector has zero magnitude or the lengths differ.
func CosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var dot, magA, magB float64
	for i := range a {
		fa, fb := float64(a[i]), float64(b[i])
		dot += fa * fb
		magA += fa * fa
		magB += fb * fb
	}
	denom := math.Sqrt(magA) * math.Sqrt(magB)
	if denom == 0 {
		return 0
	}
	return dot / denom
}
