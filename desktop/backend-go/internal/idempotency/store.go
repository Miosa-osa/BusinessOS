package idempotency

import (
	"encoding/json"
	"sync"
	"time"
)

// Entry represents a cached idempotent response.
type Entry struct {
	Status    int       `json:"status"`
	Body      string    `json:"body"`
	StoredAt  time.Time `json:"stored_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// Store manages idempotency keys with 24-hour TTL.
type Store struct {
	mu    sync.RWMutex
	cache map[string]*Entry
}

const (
	ttl = 24 * time.Hour
)

// New creates a new idempotency store.
func New() *Store {
	return &Store{
		cache: make(map[string]*Entry),
	}
}

// Store caches a response by idempotency key.
func (s *Store) Store(key string, status int, body []byte) error {
	if key == "" {
		return ErrEmptyKey
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	entry := &Entry{
		Status:    status,
		Body:      string(body),
		StoredAt:  now,
		ExpiresAt: now.Add(ttl),
	}

	s.cache[key] = entry
	return nil
}

// Get retrieves a cached response by idempotency key.
// Returns nil if not found or expired.
func (s *Store) Get(key string) *Entry {
	if key == "" {
		return nil
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, exists := s.cache[key]
	if !exists {
		return nil
	}

	// Check expiration
	if time.Now().After(entry.ExpiresAt) {
		return nil
	}

	return entry
}

// Delete removes an idempotency key.
func (s *Store) Delete(key string) {
	if key == "" {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.cache, key)
}

// Cleanup removes all expired entries from the cache.
// Returns the number of keys deleted.
func (s *Store) Cleanup() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	count := 0

	for key, entry := range s.cache {
		if now.After(entry.ExpiresAt) {
			delete(s.cache, key)
			count++
		}
	}

	return count
}

// Stats returns statistics about the store.
func (s *Store) Stats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return map[string]interface{}{
		"total_keys": len(s.cache),
	}
}

// MarshalJSON serializes an Entry to JSON.
func (e *Entry) MarshalJSON() ([]byte, error) {
	type Alias Entry
	return json.Marshal(&struct {
		*Alias
		StoredAt  int64 `json:"stored_at"`
		ExpiresAt int64 `json:"expires_at"`
	}{
		Alias:     (*Alias)(e),
		StoredAt:  e.StoredAt.Unix(),
		ExpiresAt: e.ExpiresAt.Unix(),
	})
}
