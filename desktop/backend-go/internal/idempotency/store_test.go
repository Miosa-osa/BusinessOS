package idempotency

import (
	"encoding/json"
	"sync"
	"testing"
	"time"
)

func TestStoreAndRetrieve(t *testing.T) {
	store := New()
	key := "test-key-123"
	status := 200
	body := []byte(`{"result":"success"}`)

	// Store the response
	err := store.Store(key, status, body)
	if err != nil {
		t.Fatalf("Store failed: %v", err)
	}

	// Retrieve it
	entry := store.Get(key)
	if entry == nil {
		t.Fatal("Get returned nil for stored key")
	}

	if entry.Status != status {
		t.Errorf("Expected status %d, got %d", status, entry.Status)
	}

	if entry.Body != string(body) {
		t.Errorf("Expected body %s, got %s", body, entry.Body)
	}
}

func TestStoreEmptyKey(t *testing.T) {
	store := New()

	err := store.Store("", 200, []byte("test"))
	if err != ErrEmptyKey {
		t.Errorf("Expected ErrEmptyKey, got %v", err)
	}
}

func TestGetNonExistent(t *testing.T) {
	store := New()

	entry := store.Get("non-existent-key")
	if entry != nil {
		t.Error("Get should return nil for non-existent key")
	}
}

func TestGetEmptyKey(t *testing.T) {
	store := New()

	entry := store.Get("")
	if entry != nil {
		t.Error("Get should return nil for empty key")
	}
}

func TestExpiration(t *testing.T) {
	store := New()
	key := "expiring-key"
	status := 200
	body := []byte(`{"result":"will expire"}`)

	// Store response
	err := store.Store(key, status, body)
	if err != nil {
		t.Fatalf("Store failed: %v", err)
	}

	// Verify it's retrievable
	entry := store.Get(key)
	if entry == nil {
		t.Fatal("Newly stored key should be retrievable")
	}

	// Manually expire the entry (for testing)
	store.mu.Lock()
	stored := store.cache[key]
	stored.ExpiresAt = time.Now().Add(-1 * time.Second)
	store.mu.Unlock()

	// Should now return nil
	entry = store.Get(key)
	if entry != nil {
		t.Error("Get should return nil for expired key")
	}
}

func TestDelete(t *testing.T) {
	store := New()
	key := "delete-test"
	status := 200
	body := []byte("test data")

	// Store and verify
	store.Store(key, status, body)
	if store.Get(key) == nil {
		t.Fatal("Key should exist after store")
	}

	// Delete
	store.Delete(key)

	// Verify it's deleted
	if store.Get(key) != nil {
		t.Error("Get should return nil after delete")
	}
}

func TestDeleteEmptyKey(t *testing.T) {
	store := New()

	// Should not panic
	store.Delete("")
}

func TestCleanup(t *testing.T) {
	store := New()

	// Store some entries
	for i := 0; i < 5; i++ {
		key := "key-" + string(rune(i))
		store.Store(key, 200, []byte("data"))
	}

	// Expire all entries
	store.mu.Lock()
	now := time.Now()
	for _, entry := range store.cache {
		entry.ExpiresAt = now.Add(-1 * time.Second)
	}
	store.mu.Unlock()

	// Run cleanup
	deleted := store.Cleanup()
	if deleted != 5 {
		t.Errorf("Expected 5 deleted entries, got %d", deleted)
	}

	// Verify cache is empty
	store.mu.RLock()
	size := len(store.cache)
	store.mu.RUnlock()

	if size != 0 {
		t.Errorf("Expected empty cache after cleanup, got %d entries", size)
	}
}

func TestStats(t *testing.T) {
	store := New()

	// Add some entries
	for i := 0; i < 3; i++ {
		key := "stat-key-" + string(rune(i))
		store.Store(key, 200, []byte("data"))
	}

	stats := store.Stats()

	if stats["total_keys"] != 3 {
		t.Errorf("Expected 3 keys in stats, got %v", stats["total_keys"])
	}
}

func TestConcurrentAccess(t *testing.T) {
	store := New()
	numGoroutines := 100
	keysPerGoroutine := 10

	var wg sync.WaitGroup

	// Concurrent stores
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < keysPerGoroutine; j++ {
				key := "concurrent-" + string(rune(id*keysPerGoroutine+j))
				status := 200
				body := []byte(`{"id":"` + string(rune(id)) + `"}`)
				store.Store(key, status, body)
			}
		}(i)
	}

	wg.Wait()

	// Verify all keys are stored
	stats := store.Stats()
	expectedKeys := numGoroutines * keysPerGoroutine
	if stats["total_keys"] != expectedKeys {
		t.Errorf("Expected %d keys, got %v", expectedKeys, stats["total_keys"])
	}

	// Concurrent gets
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < keysPerGoroutine; j++ {
				key := "concurrent-" + string(rune(id*keysPerGoroutine+j))
				entry := store.Get(key)
				if entry == nil {
					t.Errorf("Failed to get key: %s", key)
				}
			}
		}(i)
	}

	wg.Wait()

	// Concurrent deletes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < keysPerGoroutine; j++ {
				key := "concurrent-" + string(rune(id*keysPerGoroutine+j))
				store.Delete(key)
			}
		}(i)
	}

	wg.Wait()

	// Verify all keys are deleted
	stats = store.Stats()
	if stats["total_keys"] != 0 {
		t.Errorf("Expected 0 keys after concurrent deletes, got %v", stats["total_keys"])
	}
}

func TestMarshalJSON(t *testing.T) {
	entry := &Entry{
		Status:    200,
		Body:      `{"result":"test"}`,
		StoredAt:  time.Date(2026, 3, 24, 10, 0, 0, 0, time.UTC),
		ExpiresAt: time.Date(2026, 3, 25, 10, 0, 0, 0, time.UTC),
	}

	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	// Verify it's valid JSON
	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Unmarshaling failed: %v", err)
	}

	if result["status"] != float64(200) {
		t.Errorf("Expected status 200, got %v", result["status"])
	}

	if _, ok := result["stored_at"]; !ok {
		t.Error("stored_at should be in JSON")
	}

	if _, ok := result["expires_at"]; !ok {
		t.Error("expires_at should be in JSON")
	}
}

func TestOverwriteKey(t *testing.T) {
	store := New()
	key := "overwrite-test"

	// Store first value
	store.Store(key, 200, []byte("first"))
	entry1 := store.Get(key)

	// Overwrite with second value
	store.Store(key, 201, []byte("second"))
	entry2 := store.Get(key)

	if entry1.Body == entry2.Body {
		t.Error("Entry should have been overwritten")
	}

	if entry2.Status != 201 {
		t.Errorf("Expected status 201, got %d", entry2.Status)
	}

	if entry2.Body != "second" {
		t.Errorf("Expected body 'second', got %s", entry2.Body)
	}
}

func TestCleanupPartial(t *testing.T) {
	store := New()

	// Store some entries
	for i := 0; i < 10; i++ {
		key := "cleanup-" + string(rune(i))
		store.Store(key, 200, []byte("data"))
	}

	// Expire only half of them
	store.mu.Lock()
	now := time.Now()
	count := 0
	for _, entry := range store.cache {
		if count < 5 {
			entry.ExpiresAt = now.Add(-1 * time.Second)
		}
		count++
	}
	store.mu.Unlock()

	// Run cleanup
	deleted := store.Cleanup()
	if deleted != 5 {
		t.Errorf("Expected 5 deleted entries, got %d", deleted)
	}

	// Verify 5 entries remain
	stats := store.Stats()
	if stats["total_keys"] != 5 {
		t.Errorf("Expected 5 remaining entries, got %v", stats["total_keys"])
	}
}
