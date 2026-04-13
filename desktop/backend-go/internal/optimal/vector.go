package optimal

import (
	"encoding/binary"
	"fmt"
	"math"
	"sort"
)

// StoreEmbedding saves a float32 embedding vector for the given contextID.
// The embedding is serialised as a raw BLOB (IEEE 754 little-endian float32)
// and stored in the vectors table, which is separate from contexts.
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

	// Ensure vectors table exists — matches the real Elixir engine schema.
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS vectors (
			context_id TEXT PRIMARY KEY,
			embedding  BLOB    NOT NULL,
			model      TEXT    NOT NULL DEFAULT 'nomic-embed-text',
			dimensions INTEGER NOT NULL DEFAULT 768,
			created_at TEXT    NOT NULL DEFAULT (datetime('now'))
		)`)
	if err != nil {
		return fmt.Errorf("optimal/vector: ensure table: %w", err)
	}

	blob := float32sToBlob(embedding)

	_, err = db.Exec(`
		INSERT INTO vectors (context_id, embedding, dimensions)
		VALUES (?, ?, ?)
		ON CONFLICT(context_id) DO UPDATE SET embedding = excluded.embedding,
		                                      dimensions = excluded.dimensions`,
		contextID, blob, len(embedding),
	)
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

	// Join vectors with contexts to get path and title for result display.
	rows, err := db.Query(`
		SELECT v.context_id, COALESCE(c.path, ''), COALESCE(c.title, v.context_id), v.embedding
		FROM vectors v
		LEFT JOIN contexts c ON c.id = v.context_id
		WHERE v.embedding != ''`)
	if err != nil {
		return nil, fmt.Errorf("optimal/vector: query vectors: %w", err)
	}
	defer rows.Close()

	type scored struct {
		path  string
		title string
		score float64
	}
	var candidates []scored

	for rows.Next() {
		var contextID, path, title string
		var blob []byte
		if err := rows.Scan(&contextID, &path, &title, &blob); err != nil {
			continue
		}
		if len(blob) == 0 {
			continue
		}
		vec := blobToFloat32s(blob)
		if len(vec) == 0 {
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

// float32sToBlob serialises a []float32 as a raw BLOB using IEEE 754
// little-endian encoding. 4 bytes per element.
func float32sToBlob(vecs []float32) []byte {
	buf := make([]byte, len(vecs)*4)
	for i, v := range vecs {
		bits := math.Float32bits(v)
		binary.LittleEndian.PutUint32(buf[i*4:], bits)
	}
	return buf
}

// blobToFloat32s deserialises a raw BLOB back to []float32.
// Returns nil when the blob length is not a multiple of 4.
func blobToFloat32s(blob []byte) []float32 {
	if len(blob)%4 != 0 {
		return nil
	}
	vecs := make([]float32, len(blob)/4)
	for i := range vecs {
		bits := binary.LittleEndian.Uint32(blob[i*4:])
		vecs[i] = math.Float32frombits(bits)
	}
	return vecs
}
