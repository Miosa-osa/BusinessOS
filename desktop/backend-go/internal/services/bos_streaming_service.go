package services

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

// BOSStreamEvent mirrors the Rust streaming event structure
type BOSStreamEvent struct {
	ID                     string                 `json:"id"`
	EventType              string                 `json:"event_type"`
	SessionID              string                 `json:"session_id"`
	Progress               *BOSProgressMetrics    `json:"progress,omitempty"`
	Metrics                *BOSAggregatedMetrics  `json:"metrics,omitempty"`
	Error                  *BOSErrorInfo          `json:"error,omitempty"`
	PartialResult          map[string]interface{} `json:"partial_result,omitempty"`
	TimestampMs            int64                  `json:"timestamp_ms"`
	EstimatedRemainingSecs *int64                 `json:"estimated_remaining_secs,omitempty"`
}

// BOSProgressMetrics contains real-time progress information
type BOSProgressMetrics struct {
	EventsProcessed int64   `json:"events_processed"`
	TotalEvents     *int64  `json:"total_events,omitempty"`
	PercentComplete int32   `json:"percent_complete"`
	CurrentStep     string  `json:"current_step"`
	ActiveWorkers   int32   `json:"active_workers"`
	ThroughputEPS   float64 `json:"throughput_eps"`
}

// BOSAggregatedMetrics contains aggregated processing metrics
type BOSAggregatedMetrics struct {
	ElapsedSecs          int64   `json:"elapsed_secs"`
	TotalProcessed       int64   `json:"total_processed"`
	AvgThroughputEPS     float64 `json:"avg_throughput_eps"`
	CurrentThroughputEPS float64 `json:"current_throughput_eps"`
	PeakThroughputEPS    float64 `json:"peak_throughput_eps"`
	VariantsFound        int64   `json:"variants_found"`
	ViolationsFound      int64   `json:"violations_found"`
}

// BOSErrorInfo contains error details with recovery information
type BOSErrorInfo struct {
	Code         string  `json:"code"`
	Message      string  `json:"message"`
	Recoverable  bool    `json:"recoverable"`
	RetryAttempt *int32  `json:"retry_attempt,omitempty"`
	MaxRetries   *int32  `json:"max_retries,omitempty"`
	Details      *string `json:"details,omitempty"`
}

// BOSStreamSubscriber represents a subscription to BOS events
type BOSStreamSubscriber struct {
	ID        string
	SessionID uuid.UUID
	UserID    uuid.UUID
	Events    chan *BOSStreamEvent
	ctx       context.Context
	cancelFn  context.CancelFunc
	done      chan struct{}
	closeOnce sync.Once
}

// Done returns a channel closed when subscriber is unsubscribed
func (s *BOSStreamSubscriber) Done() <-chan struct{} {
	return s.done
}

// BOSSessionMetrics tracks aggregated metrics for a session
type BOSSessionMetrics struct {
	SessionID      uuid.UUID
	UserID         uuid.UUID
	StartTime      time.Time
	LastUpdateTime time.Time
	Phase          string // discovery, conformance, complete

	// Progress tracking
	EventsProcessed atomic.Int64
	TotalEvents     *int64
	ProgressPercent atomic.Int32
	CurrentStep     string
	ActiveWorkers   atomic.Int32

	// Metrics tracking
	ElapsedSecs          atomic.Int64
	AvgThroughputEPS     atomic.Value // *float64
	CurrentThroughputEPS atomic.Value // *float64
	PeakThroughputEPS    atomic.Value // *float64
	VariantsFound        atomic.Int64
	ViolationsFound      atomic.Int64

	// State
	mu          sync.RWMutex
	IsCancelled bool
	IsComplete  bool
	LastError   *BOSErrorInfo
}

// BOSStreamingService manages BOS streaming subscriptions and metrics aggregation
type BOSStreamingService struct {
	// Subscribers
	subscribers map[string]*BOSStreamSubscriber
	subMu       sync.RWMutex

	// Sessions with metrics
	sessions map[uuid.UUID]*BOSSessionMetrics
	sessMu   sync.RWMutex

	// WebSocket broadcast channels (for forwarding to UI)
	wsChannels map[uuid.UUID][]chan *BOSStreamEvent
	wsMu       sync.RWMutex

	logger *slog.Logger

	// Metrics
	totalEventsReceived atomic.Int64
	totalEventsDropped  atomic.Int64
}

// NewBOSStreamingService creates a new BOS streaming service
func NewBOSStreamingService(logger *slog.Logger) *BOSStreamingService {
	if logger == nil {
		logger = slog.Default()
	}

	return &BOSStreamingService{
		subscribers: make(map[string]*BOSStreamSubscriber),
		sessions:    make(map[uuid.UUID]*BOSSessionMetrics),
		wsChannels:  make(map[uuid.UUID][]chan *BOSStreamEvent),
		logger:      logger.With("component", "bos_streaming"),
	}
}

// Subscribe creates a new subscription for BOS session events
func (s *BOSStreamingService) Subscribe(
	ctx context.Context,
	userID uuid.UUID,
	sessionID uuid.UUID,
) *BOSStreamSubscriber {
	// Create cancellable context
	subCtx, cancel := context.WithCancel(ctx)

	subscriber := &BOSStreamSubscriber{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		UserID:    userID,
		Events:    make(chan *BOSStreamEvent, 100), // Buffered channel
		ctx:       subCtx,
		cancelFn:  cancel,
		done:      make(chan struct{}),
	}

	s.subMu.Lock()
	s.subscribers[subscriber.ID] = subscriber
	subCount := len(s.subscribers)
	s.subMu.Unlock()

	// Get or create session
	s.sessMu.Lock()
	metrics, exists := s.sessions[sessionID]
	if !exists {
		metrics = &BOSSessionMetrics{
			SessionID:      sessionID,
			UserID:         userID,
			StartTime:      time.Now(),
			LastUpdateTime: time.Now(),
			Phase:          "discovery",
		}
		s.sessions[sessionID] = metrics
	}
	s.sessMu.Unlock()

	s.logger.Info("BOS subscriber connected",
		"subscriber_id", subscriber.ID,
		"session_id", sessionID,
		"user_id", userID,
		"total_subscribers", subCount,
	)

	// Start cleanup on context done
	go s.cleanupOnContextDone(subscriber)

	return subscriber
}

// Unsubscribe removes a subscription
func (s *BOSStreamingService) Unsubscribe(subscriberID string) {
	s.subMu.Lock()
	subscriber, exists := s.subscribers[subscriberID]
	delete(s.subscribers, subscriberID)
	s.subMu.Unlock()

	if exists {
		subscriber.cancelFn()
		subscriber.closeOnce.Do(func() {
			close(subscriber.done)
			close(subscriber.Events)
		})

		s.logger.Info("BOS subscriber disconnected",
			"subscriber_id", subscriberID,
			"session_id", subscriber.SessionID,
		)
	}
}

// PublishEvent publishes an event to all subscribers for a session
func (s *BOSStreamingService) PublishEvent(event *BOSStreamEvent) {
	s.totalEventsReceived.Add(1)

	sessionID, err := uuid.Parse(event.SessionID)
	if err != nil {
		s.logger.Error("invalid session ID in event", "error", err)
		return
	}

	// Update session metrics
	s.updateSessionMetrics(sessionID, event)

	// Deliver to all subscribers for this session
	s.subMu.RLock()
	delivered := 0
	dropped := 0

	for _, subscriber := range s.subscribers {
		if subscriber.SessionID == sessionID {
			select {
			case subscriber.Events <- event:
				delivered++
			default:
				// Buffer full, drop event
				dropped++
				s.totalEventsDropped.Add(1)
			}
		}
	}
	s.subMu.RUnlock()

	if dropped > 0 {
		s.logger.Warn("dropped events due to buffer full",
			"dropped", dropped,
			"delivered", delivered,
			"session_id", sessionID,
		)
	}
}

// BroadcastEvent forwards an event to WebSocket clients
func (s *BOSStreamingService) BroadcastEvent(
	subscriberID string,
	sessionID uuid.UUID,
	event *BOSStreamEvent,
) {
	s.wsMu.RLock()
	channels, exists := s.wsChannels[sessionID]
	s.wsMu.RUnlock()

	if !exists {
		return
	}

	for _, ch := range channels {
		select {
		case ch <- event:
		default:
			// WebSocket channel full, skip
		}
	}
}

// RegisterWebSocketClient registers a WebSocket channel for broadcasting
func (s *BOSStreamingService) RegisterWebSocketClient(
	sessionID uuid.UUID,
	ch chan *BOSStreamEvent,
) {
	s.wsMu.Lock()
	s.wsChannels[sessionID] = append(s.wsChannels[sessionID], ch)
	s.wsMu.Unlock()

	s.logger.Debug("WebSocket client registered", "session_id", sessionID)
}

// UnregisterWebSocketClient removes a WebSocket channel
func (s *BOSStreamingService) UnregisterWebSocketClient(
	sessionID uuid.UUID,
	ch chan *BOSStreamEvent,
) {
	s.wsMu.Lock()
	channels := s.wsChannels[sessionID]

	// Remove channel from list
	for i, c := range channels {
		if c == ch {
			channels = append(channels[:i], channels[i+1:]...)
			break
		}
	}

	if len(channels) == 0 {
		delete(s.wsChannels, sessionID)
	} else {
		s.wsChannels[sessionID] = channels
	}
	s.wsMu.Unlock()
}

// GetAggregatedMetrics returns current aggregated metrics for a session
func (s *BOSStreamingService) GetAggregatedMetrics(sessionID uuid.UUID) *BOSAggregatedMetrics {
	s.sessMu.RLock()
	metrics, exists := s.sessions[sessionID]
	s.sessMu.RUnlock()

	if !exists {
		return nil
	}

	metrics.mu.RLock()
	defer metrics.mu.RUnlock()

	return &BOSAggregatedMetrics{
		ElapsedSecs:          metrics.ElapsedSecs.Load(),
		TotalProcessed:       metrics.EventsProcessed.Load(),
		AvgThroughputEPS:     loadFloat64(&metrics.AvgThroughputEPS),
		CurrentThroughputEPS: loadFloat64(&metrics.CurrentThroughputEPS),
		PeakThroughputEPS:    loadFloat64(&metrics.PeakThroughputEPS),
		VariantsFound:        metrics.VariantsFound.Load(),
		ViolationsFound:      metrics.ViolationsFound.Load(),
	}
}

// GetAllActiveSessions returns list of all active sessions
func (s *BOSStreamingService) GetAllActiveSessions() []map[string]interface{} {
	s.sessMu.RLock()
	defer s.sessMu.RUnlock()

	var sessions []map[string]interface{}
	for sessionID, metrics := range s.sessions {
		metrics.mu.RLock()
		sessions = append(sessions, map[string]interface{}{
			"session_id":       sessionID.String(),
			"user_id":          metrics.UserID.String(),
			"start_time":       metrics.StartTime,
			"phase":            metrics.Phase,
			"progress_pct":     metrics.ProgressPercent.Load(),
			"events_processed": metrics.EventsProcessed.Load(),
			"is_cancelled":     metrics.IsCancelled,
			"is_complete":      metrics.IsComplete,
		})
		metrics.mu.RUnlock()
	}

	return sessions
}

// CancelSession cancels a BOS processing session
func (s *BOSStreamingService) CancelSession(sessionID uuid.UUID) bool {
	s.sessMu.RLock()
	metrics, exists := s.sessions[sessionID]
	s.sessMu.RUnlock()

	if !exists {
		return false
	}

	metrics.mu.Lock()
	if metrics.IsCancelled {
		metrics.mu.Unlock()
		return false
	}
	metrics.IsCancelled = true
	metrics.mu.Unlock()

	return true
}

// updateSessionMetrics updates metrics based on received event
func (s *BOSStreamingService) updateSessionMetrics(sessionID uuid.UUID, event *BOSStreamEvent) {
	s.sessMu.RLock()
	metrics, exists := s.sessions[sessionID]
	s.sessMu.RUnlock()

	if !exists {
		return
	}

	metrics.mu.Lock()
	defer metrics.mu.Unlock()

	metrics.LastUpdateTime = time.Now()

	if event.Progress != nil {
		metrics.EventsProcessed.Store(event.Progress.EventsProcessed)
		metrics.TotalEvents = event.Progress.TotalEvents
		metrics.ProgressPercent.Store(event.Progress.PercentComplete)
		metrics.CurrentStep = event.Progress.CurrentStep
		metrics.ActiveWorkers.Store(event.Progress.ActiveWorkers)
	}

	if event.Metrics != nil {
		metrics.ElapsedSecs.Store(event.Metrics.ElapsedSecs)
		storeFloat64(&metrics.AvgThroughputEPS, event.Metrics.AvgThroughputEPS)
		storeFloat64(&metrics.CurrentThroughputEPS, event.Metrics.CurrentThroughputEPS)
		storeFloat64(&metrics.PeakThroughputEPS, event.Metrics.PeakThroughputEPS)
		metrics.VariantsFound.Store(event.Metrics.VariantsFound)
		metrics.ViolationsFound.Store(event.Metrics.ViolationsFound)
		// Aggregate TotalProcessed from metrics events
		if event.Metrics.TotalProcessed > 0 {
			metrics.EventsProcessed.Store(event.Metrics.TotalProcessed)
		}
	}

	if event.Error != nil {
		metrics.LastError = event.Error
		if !event.Error.Recoverable {
			metrics.IsCancelled = true
		}
	}

	// Update phase based on event type
	switch event.EventType {
	case "discovery_started":
		metrics.Phase = "discovery"
	case "conformance_started":
		metrics.Phase = "conformance"
	case "processing_complete":
		metrics.Phase = "complete"
		metrics.IsComplete = true
	}
}

// cleanupOnContextDone cleans up subscriber when context is cancelled
func (s *BOSStreamingService) cleanupOnContextDone(subscriber *BOSStreamSubscriber) {
	<-subscriber.ctx.Done()
	s.Unsubscribe(subscriber.ID)
}

// Helper functions for atomic float64 storage using atomic.Value

func storeFloat64(av *atomic.Value, f float64) {
	av.Store(f)
}

func loadFloat64(av *atomic.Value) float64 {
	val := av.Load()
	if val == nil {
		return 0.0
	}
	if f, ok := val.(float64); ok {
		return f
	}
	return 0.0
}

// GetSessionMetrics returns session metrics if it exists
func (s *BOSStreamingService) GetSessionMetrics(sessionID uuid.UUID) *BOSSessionMetrics {
	s.sessMu.RLock()
	defer s.sessMu.RUnlock()

	return s.sessions[sessionID]
}

// GetActiveSubscriberCount returns number of active subscribers
func (s *BOSStreamingService) GetActiveSubscriberCount() int {
	s.subMu.RLock()
	defer s.subMu.RUnlock()

	return len(s.subscribers)
}

// GetTotalMetrics returns overall service metrics
func (s *BOSStreamingService) GetTotalMetrics() map[string]interface{} {
	return map[string]interface{}{
		"total_events_received": s.totalEventsReceived.Load(),
		"total_events_dropped":  s.totalEventsDropped.Load(),
		"active_subscribers":    s.GetActiveSubscriberCount(),
		"active_sessions":       len(s.GetAllActiveSessions()),
	}
}
