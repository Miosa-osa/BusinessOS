package services

import "sync"

type SSEEvent struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// SSEBroadcaster manages SSE connections per user
type SSEBroadcaster struct {
	mu      sync.RWMutex
	clients map[string]map[chan SSEEvent]struct{} 
}

// NewSSEBroadcaster creates a new broadcaster
func NewSSEBroadcaster() *SSEBroadcaster {
	return &SSEBroadcaster{
		clients: make(map[string]map[chan SSEEvent]struct{}),
	}
}

// Subscribe creates a channel for a user
func (b *SSEBroadcaster) Subscribe(userID string) chan SSEEvent {
	b.mu.Lock()
	defer b.mu.Unlock()

	ch := make(chan SSEEvent, 16)
	if b.clients[userID] == nil {
		b.clients[userID] = make(map[chan SSEEvent]struct{})
	}
	b.clients[userID][ch] = struct{}{}
	return ch
}

func (b *SSEBroadcaster) Unsubscribe(userID string, ch chan SSEEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if channels, ok := b.clients[userID]; ok {
		delete(channels, ch)
		close(ch)
		if len(channels) == 0 {
			delete(b.clients, userID)
		}
	}
}

// SendToUser sends an event to all connections for a user
func (b *SSEBroadcaster) SendToUser(userID string, event SSEEvent) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if channels, ok := b.clients[userID]; ok {
		for ch := range channels {
			select {
			case ch <- event:
			default:
			}
		}
	}
}

// SendToAll broadcasts to all connected users
func (b *SSEBroadcaster) SendToAll(event SSEEvent) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, channels := range b.clients {
		for ch := range channels {
			select {
			case ch <- event:
			default:
			}
		}
	}
}

// IsConnected checks if a user has active connections
func (b *SSEBroadcaster) IsConnected(userID string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	channels, ok := b.clients[userID]
	return ok && len(channels) > 0
}

// ConnectedCount returns number of connected users
func (b *SSEBroadcaster) ConnectedCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.clients)
}
