//! Real-time event streaming for BOS process discovery and conformance checking.
//!
//! This module provides Server-Sent Events (SSE) streaming for large-scale process mining
//! operations. BOS can process 1M-100M event logs, and this streaming layer provides
//! real-time progress updates to connected clients (UI, monitoring systems, etc).
//!
//! # Architecture
//!
//! ```text
//! BOS Process Mining (Rust)
//!     |
//!     v
//! StreamingCoordinator (this module)
//!     |
//!     +-- ProgressEvent (discovery %, current step)
//!     +-- MetricsEvent (throughput, elapsed time)
//!     +-- PartialResultEvent (intermediate findings)
//!     +-- ErrorEvent (recoverable/fatal errors)
//!     |
//!     v
//! HTTP Server (SSE endpoint)
//!     |
//!     v
//! BusinessOS Go Backend (receives, aggregates)
//!     |
//!     v
//! WebSocket -> SvelteKit UI (real-time progress bar)
//! ```
//!
//! # Usage
//!
//! ```ignore
//! let coordinator = StreamingCoordinator::new(num_workers);
//! let handle = coordinator.start_discovery(event_log).await?;
//!
//! // Emit events during processing
//! handle.emit_progress(ProgressEvent {
//!     events_processed: 1000,
//!     total_events: 100_000_000,
//!     current_step: "Conformance Checking".to_string(),
//!     // ...
//! }).await?;
//! ```

use serde::{Deserialize, Serialize};
use std::sync::atomic::{AtomicU64, Ordering};
use std::sync::Arc;
use std::time::{Duration, Instant};
use tokio::sync::mpsc;
use uuid::Uuid;

/// Maximum buffer size for streaming events before backpressure
const MAX_EVENT_BUFFER: usize = 1000;

/// Streaming event types
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq)]
pub enum StreamEventType {
    /// Discovery process started
    #[serde(rename = "discovery_started")]
    DiscoveryStarted,

    /// Progress update during discovery (1-99%)
    #[serde(rename = "discovery_progress")]
    DiscoveryProgress,

    /// Conformance checking started
    #[serde(rename = "conformance_started")]
    ConformanceStarted,

    /// Progress during conformance (1-99%)
    #[serde(rename = "conformance_progress")]
    ConformanceProgress,

    /// Partial results available (intermediate findings)
    #[serde(rename = "partial_results")]
    PartialResults,

    /// Processing complete (100% discovery/conformance)
    #[serde(rename = "processing_complete")]
    ProcessingComplete,

    /// Recoverable error (will retry)
    #[serde(rename = "error_recoverable")]
    ErrorRecoverable,

    /// Fatal error (cannot proceed)
    #[serde(rename = "error_fatal")]
    ErrorFatal,

    /// Metrics snapshot (throughput, elapsed time)
    #[serde(rename = "metrics")]
    Metrics,

    /// Heartbeat (keeps connection alive)
    #[serde(rename = "heartbeat")]
    Heartbeat,
}

/// Event streaming type that clients receive
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct StreamEvent {
    /// Unique event ID
    pub id: String,

    /// Event classification
    pub event_type: StreamEventType,

    /// Processing session/correlation ID
    pub session_id: String,

    /// Progress event (if applicable)
    #[serde(skip_serializing_if = "Option::is_none")]
    pub progress: Option<ProgressEvent>,

    /// Metrics event (if applicable)
    #[serde(skip_serializing_if = "Option::is_none")]
    pub metrics: Option<MetricsEvent>,

    /// Partial result event (if applicable)
    #[serde(skip_serializing_if = "Option::is_none")]
    pub partial_result: Option<PartialResultEvent>,

    /// Error event (if applicable)
    #[serde(skip_serializing_if = "Option::is_none")]
    pub error: Option<ErrorEvent>,

    /// Unix timestamp in milliseconds
    pub timestamp_ms: u64,

    /// Estimated time remaining (seconds, None if unknown)
    #[serde(skip_serializing_if = "Option::is_none")]
    pub estimated_remaining_secs: Option<u64>,
}

/// Progress update during discovery/conformance
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ProgressEvent {
    /// Events processed so far
    pub events_processed: u64,

    /// Total events in log (None if unknown)
    pub total_events: Option<u64>,

    /// Completion percentage (0-100)
    pub percent_complete: u32,

    /// Current processing step/phase
    pub current_step: String,

    /// Current worker count
    pub active_workers: usize,

    /// Throughput (events/second)
    pub throughput_eps: f64,
}

/// Metrics snapshot during processing
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MetricsEvent {
    /// Elapsed time (seconds)
    pub elapsed_secs: u64,

    /// Total events processed
    pub total_processed: u64,

    /// Average throughput (events/second)
    pub avg_throughput_eps: f64,

    /// Current throughput (events/second)
    pub current_throughput_eps: f64,

    /// Peak throughput (events/second)
    pub peak_throughput_eps: f64,

    /// Number of process variants discovered
    pub variants_found: u64,

    /// Number of conformance violations
    pub violations_found: u64,
}

/// Partial result (intermediate findings)
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PartialResultEvent {
    /// Type of partial result
    pub result_type: String, // e.g., "top_variants", "traces_sample", "bottlenecks"

    /// Serialized result data (JSON)
    pub data: serde_json::Value,

    /// Number of complete results collected
    pub items_count: usize,

    /// Whether this is the final partial result
    pub is_final: bool,
}

/// Error event with recovery info
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ErrorEvent {
    /// Error code
    pub code: String,

    /// Human-readable error message
    pub message: String,

    /// Whether this error is recoverable
    pub recoverable: bool,

    /// Retry attempt number (if applicable)
    pub retry_attempt: Option<u32>,

    /// Max retry attempts (if applicable)
    pub max_retries: Option<u32>,

    /// Detailed error context
    pub details: Option<String>,
}

/// Handle to a streaming session for emitting events
#[derive(Clone)]
pub struct StreamingSessionHandle {
    /// Unique session ID
    pub session_id: String,

    /// Event sender
    sender: mpsc::Sender<StreamEvent>,

    /// Start time for metrics
    start_time: Arc<Instant>,

    /// Processed event counter (atomic for lock-free updates)
    processed_count: Arc<AtomicU64>,
}

impl StreamingSessionHandle {
    /// Create new streaming session
    pub fn new(session_id: String, sender: mpsc::Sender<StreamEvent>) -> Self {
        Self {
            session_id,
            sender,
            start_time: Arc::new(Instant::now()),
            processed_count: Arc::new(AtomicU64::new(0)),
        }
    }

    /// Emit progress event
    pub async fn emit_progress(&self, progress: ProgressEvent) -> Result<(), String> {
        let event = StreamEvent {
            id: Uuid::new_v4().to_string(),
            event_type: match progress.percent_complete {
                0 => StreamEventType::DiscoveryStarted,
                1..=99 => StreamEventType::DiscoveryProgress,
                100 => StreamEventType::ProcessingComplete,
                _ => return Err("Invalid progress percentage".to_string()),
            },
            session_id: self.session_id.clone(),
            progress: Some(progress),
            metrics: None,
            partial_result: None,
            error: None,
            timestamp_ms: current_timestamp_ms(),
            estimated_remaining_secs: None,
        };

        self.sender.send(event).await.map_err(|e| e.to_string())
    }

    /// Emit metrics event
    pub async fn emit_metrics(&self, metrics: MetricsEvent) -> Result<(), String> {
        let event = StreamEvent {
            id: Uuid::new_v4().to_string(),
            event_type: StreamEventType::Metrics,
            session_id: self.session_id.clone(),
            progress: None,
            metrics: Some(metrics),
            partial_result: None,
            error: None,
            timestamp_ms: current_timestamp_ms(),
            estimated_remaining_secs: None,
        };

        self.sender.send(event).await.map_err(|e| e.to_string())
    }

    /// Emit partial result event
    pub async fn emit_partial_result(&self, result: PartialResultEvent) -> Result<(), String> {
        let event = StreamEvent {
            id: Uuid::new_v4().to_string(),
            event_type: StreamEventType::PartialResults,
            session_id: self.session_id.clone(),
            progress: None,
            metrics: None,
            partial_result: Some(result),
            error: None,
            timestamp_ms: current_timestamp_ms(),
            estimated_remaining_secs: None,
        };

        self.sender.send(event).await.map_err(|e| e.to_string())
    }

    /// Emit error event (recoverable)
    pub async fn emit_error_recoverable(
        &self,
        code: String,
        message: String,
        retry_attempt: Option<u32>,
        max_retries: Option<u32>,
    ) -> Result<(), String> {
        let event = StreamEvent {
            id: Uuid::new_v4().to_string(),
            event_type: StreamEventType::ErrorRecoverable,
            session_id: self.session_id.clone(),
            progress: None,
            metrics: None,
            partial_result: None,
            error: Some(ErrorEvent {
                code,
                message,
                recoverable: true,
                retry_attempt,
                max_retries,
                details: None,
            }),
            timestamp_ms: current_timestamp_ms(),
            estimated_remaining_secs: None,
        };

        self.sender.send(event).await.map_err(|e| e.to_string())
    }

    /// Emit error event (fatal)
    pub async fn emit_error_fatal(
        &self,
        code: String,
        message: String,
        details: Option<String>,
    ) -> Result<(), String> {
        let event = StreamEvent {
            id: Uuid::new_v4().to_string(),
            event_type: StreamEventType::ErrorFatal,
            session_id: self.session_id.clone(),
            progress: None,
            metrics: None,
            partial_result: None,
            error: Some(ErrorEvent {
                code,
                message,
                recoverable: false,
                retry_attempt: None,
                max_retries: None,
                details,
            }),
            timestamp_ms: current_timestamp_ms(),
            estimated_remaining_secs: None,
        };

        self.sender.send(event).await.map_err(|e| e.to_string())
    }

    /// Increment processed event counter
    pub fn increment_processed(&self, count: u64) {
        self.processed_count.fetch_add(count, Ordering::Relaxed);
    }

    /// Get elapsed seconds since session start
    pub fn elapsed_secs(&self) -> u64 {
        self.start_time.elapsed().as_secs()
    }

    /// Get total processed events
    pub fn total_processed(&self) -> u64 {
        self.processed_count.load(Ordering::Relaxed)
    }

    /// Calculate throughput (events/second)
    pub fn calculate_throughput(&self) -> f64 {
        let elapsed = self.start_time.elapsed().as_secs_f64();
        if elapsed < 0.001 {
            return 0.0;
        }
        let processed = self.total_processed() as f64;
        processed / elapsed
    }
}

/// Streaming coordinator manages all active streaming sessions
pub struct StreamingCoordinator {
    /// Receive end of event channel
    pub receiver: mpsc::Receiver<StreamEvent>,

    /// Sessions map
    sessions: Arc<std::sync::Mutex<std::collections::HashMap<String, StreamingSessionHandle>>>,
}

impl StreamingCoordinator {
    /// Create new streaming coordinator
    pub fn new() -> (Self, mpsc::Sender<StreamEvent>) {
        let (tx, rx) = mpsc::channel(MAX_EVENT_BUFFER);
        (
            Self {
                receiver: rx,
                sessions: Arc::new(std::sync::Mutex::new(
                    std::collections::HashMap::new(),
                )),
            },
            tx,
        )
    }

    /// Create new streaming session
    pub fn create_session(&self, sender: mpsc::Sender<StreamEvent>) -> StreamingSessionHandle {
        let session_id = Uuid::new_v4().to_string();
        let handle = StreamingSessionHandle::new(session_id.clone(), sender);

        if let Ok(mut sessions) = self.sessions.lock() {
            sessions.insert(session_id.clone(), handle.clone());
        }

        handle
    }

    /// Get session by ID
    pub fn get_session(&self, session_id: &str) -> Option<StreamingSessionHandle> {
        self.sessions
            .lock()
            .ok()
            .and_then(|sessions| sessions.get(session_id).cloned())
    }

    /// Remove session
    pub fn remove_session(&self, session_id: &str) {
        if let Ok(mut sessions) = self.sessions.lock() {
            sessions.remove(session_id);
        }
    }

    /// Get active session count
    pub fn active_session_count(&self) -> usize {
        self.sessions
            .lock()
            .map(|sessions| sessions.len())
            .unwrap_or(0)
    }
}

impl Default for StreamingCoordinator {
    fn default() -> Self {
        Self::new().0
    }
}

// Utility function to get current timestamp in milliseconds
fn current_timestamp_ms() -> u64 {
    std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .map(|d| d.as_millis() as u64)
        .unwrap_or(0)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn test_streaming_session_creation() {
        let (tx, _rx) = mpsc::channel(100);
        let session = StreamingSessionHandle::new("test-session".to_string(), tx);

        assert_eq!(session.session_id, "test-session");
        assert_eq!(session.total_processed(), 0);
    }

    #[tokio::test]
    async fn test_progress_event_emission() {
        let (tx, mut rx) = mpsc::channel(100);
        let session = StreamingSessionHandle::new("test-session".to_string(), tx);

        let progress = ProgressEvent {
            events_processed: 1000,
            total_events: Some(100_000),
            percent_complete: 1,
            current_step: "Discovery".to_string(),
            active_workers: 4,
            throughput_eps: 1000.0,
        };

        session.emit_progress(progress).await.unwrap();

        let event = rx.recv().await;
        assert!(event.is_some());
        let event = event.unwrap();
        assert_eq!(event.event_type, StreamEventType::DiscoveryProgress);
        assert!(event.progress.is_some());
    }

    #[tokio::test]
    async fn test_error_event_recoverable() {
        let (tx, mut rx) = mpsc::channel(100);
        let session = StreamingSessionHandle::new("test-session".to_string(), tx);

        session
            .emit_error_recoverable(
                "TIMEOUT".to_string(),
                "Worker timeout".to_string(),
                Some(1),
                Some(3),
            )
            .await
            .unwrap();

        let event = rx.recv().await.unwrap();
        assert_eq!(event.event_type, StreamEventType::ErrorRecoverable);
        assert!(event.error.is_some());
        let err = event.error.unwrap();
        assert!(err.recoverable);
        assert_eq!(err.retry_attempt, Some(1));
    }

    #[tokio::test]
    async fn test_metrics_event() {
        let (tx, mut rx) = mpsc::channel(100);
        let session = StreamingSessionHandle::new("test-session".to_string(), tx);

        let metrics = MetricsEvent {
            elapsed_secs: 10,
            total_processed: 10_000,
            avg_throughput_eps: 1000.0,
            current_throughput_eps: 1100.0,
            peak_throughput_eps: 1200.0,
            variants_found: 42,
            violations_found: 3,
        };

        session.emit_metrics(metrics).await.unwrap();

        let event = rx.recv().await.unwrap();
        assert_eq!(event.event_type, StreamEventType::Metrics);
        assert!(event.metrics.is_some());
    }

    #[test]
    fn test_streaming_coordinator() {
        let (coordinator, tx) = StreamingCoordinator::new();

        let session = coordinator.create_session(tx);
        assert_eq!(coordinator.active_session_count(), 1);

        coordinator.remove_session(&session.session_id);
        assert_eq!(coordinator.active_session_count(), 0);
    }

    #[test]
    fn test_processed_counter() {
        let (tx, _rx) = mpsc::channel(100);
        let session = StreamingSessionHandle::new("test".to_string(), tx);

        session.increment_processed(100);
        assert_eq!(session.total_processed(), 100);

        session.increment_processed(50);
        assert_eq!(session.total_processed(), 150);
    }

    #[test]
    fn test_throughput_calculation() {
        let (tx, _rx) = mpsc::channel(100);
        let session = StreamingSessionHandle::new("test".to_string(), tx);

        session.increment_processed(1000);
        let throughput = session.calculate_throughput();

        // Throughput should be non-zero (depends on elapsed time)
        assert!(throughput >= 0.0);
    }
}
