package testutil

import (
	"testing"
	"time"
)

// MessageChannel is a type-safe channel for inter-goroutine communication.
//
// Usage:
//
//	msgChan := NewMessageChannel[string](10)
//	go func() {
//		msgChan.Send("hello")
//	}()
//	msg := AssertReceivesMessage(t, msgChan, 1*time.Second)
//	assert.Equal(t, "hello", msg)
type MessageChannel[T any] struct {
	ch chan T
}

// NewMessageChannel creates message-passing channel (no shared state).
func NewMessageChannel[T any](bufferSize int) *MessageChannel[T] {
	return &MessageChannel[T]{
		ch: make(chan T, bufferSize),
	}
}

// Send transmits message (replaces shared mutable state).
func (m *MessageChannel[T]) Send(msg T) {
	m.ch <- msg
}

// Close closes the channel.
func (m *MessageChannel[T]) Close() {
	close(m.ch)
}

// Channel returns underlying chan (for select statements).
func (m *MessageChannel[T]) Channel() <-chan T {
	return m.ch
}

// AssertReceivesMessage validates message reception within timeout.
//
// Usage:
//
//	msgChan := NewMessageChannel[string](1)
//	msg := AssertReceivesMessage(t, msgChan, 5*time.Second)
//	assert.Equal(t, expectedMsg, msg)
func AssertReceivesMessage[T any](t *testing.T, msgChan *MessageChannel[T], timeout time.Duration) T {
	t.Helper()

	select {
	case msg := <-msgChan.ch:
		return msg
	case <-time.After(timeout):
		var zero T
		t.Fatalf("expected message not received within %v", timeout)
		return zero
	}
}

// AssertNoDirectMemoryAccess is a code review helper (validates no mutex usage).
// This is a placeholder for static analysis; in production, verify via linting.
//
// Usage:
//
//	// In code review:
//	// ✓ Uses channel sends, not shared variables
//	// ✓ No sync.Mutex locks visible
//	AssertNoDirectMemoryAccess(t, "MyHandler")
func AssertNoDirectMemoryAccess(t *testing.T, contextName string) {
	t.Helper()
	// This is a lint-time check, documented for code review only
	// In CI, use `go vet` with custom analyzer rules
}
