//! Worker process abstraction for supervision tree.
//!
//! A Worker represents a supervised process with lifecycle management,
//! state recovery, and timeout handling.

use std::sync::Arc;
use std::time::Duration;
use tokio::sync::{mpsc, oneshot};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

use super::error::{SupervisionError, SupervisionResult};

/// Unique identifier for a worker.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub struct WorkerId(Uuid);

impl WorkerId {
    /// Generate a new worker ID.
    pub fn new() -> Self {
        Self(Uuid::new_v4())
    }
}

impl Default for WorkerId {
    fn default() -> Self {
        Self::new()
    }
}

/// State of a worker process.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum WorkerState {
    /// Worker is starting up
    Starting,
    /// Worker is running normally
    Running,
    /// Worker is shutting down gracefully
    Shutting,
    /// Worker has stopped
    Stopped,
    /// Worker crashed
    Crashed,
    /// Worker is restarting
    Restarting,
}

/// Configuration for a worker process.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct WorkerConfig {
    /// Human-readable name for the worker
    pub name: String,

    /// Timeout for worker operations (milliseconds)
    pub timeout_ms: u64,

    /// Maximum memory per process (bytes)
    pub memory_limit_bytes: usize,

    /// Whether to recover state on restart
    pub recover_state: bool,

    /// Max reachability graph markings
    pub max_markings: usize,
}

impl WorkerConfig {
    /// Create a new worker configuration.
    pub fn new(name: impl Into<String>, timeout_ms: u64) -> Self {
        Self {
            name: name.into(),
            timeout_ms,
            memory_limit_bytes: 2 * 1024 * 1024 * 1024, // 2GB default
            recover_state: true,
            max_markings: 100_000,
        }
    }

    /// Discovery worker: 5 minute timeout
    pub fn discovery() -> Self {
        Self {
            name: "DiscoveryWorker".to_string(),
            timeout_ms: 5 * 60 * 1000,
            memory_limit_bytes: 2 * 1024 * 1024 * 1024,
            recover_state: true,
            max_markings: 100_000,
        }
    }

    /// Conformance worker: 30 second timeout
    pub fn conformance() -> Self {
        Self {
            name: "ConformanceWorker".to_string(),
            timeout_ms: 30 * 1000,
            memory_limit_bytes: 2 * 1024 * 1024 * 1024,
            recover_state: true,
            max_markings: 100_000,
        }
    }

    /// I/O worker: 10 second timeout
    pub fn io() -> Self {
        Self {
            name: "IOWorker".to_string(),
            timeout_ms: 10 * 1000,
            memory_limit_bytes: 2 * 1024 * 1024 * 1024,
            recover_state: true,
            max_markings: 100_000,
        }
    }
}

/// Message sent to a worker.
#[derive(Debug, Clone)]
pub enum WorkerMessage {
    /// Execute work with the given payload
    Work(Vec<u8>),
    /// Shut down gracefully
    Shutdown,
}

/// Worker handle for supervision.
#[derive(Clone)]
pub struct WorkerHandle {
    /// Unique worker ID
    id: WorkerId,
    /// Worker configuration
    config: WorkerConfig,
    /// Current state
    state: Arc<tokio::sync::Mutex<WorkerState>>,
    /// Channel to send work to the worker
    tx: mpsc::Sender<WorkerMessage>,
    /// Last known state for recovery
    last_state: Arc<tokio::sync::Mutex<Option<Vec<u8>>>>,
}

impl WorkerHandle {
    /// Create a new worker with the given configuration.
    pub fn new(config: WorkerConfig) -> Self {
        let (tx, _rx) = mpsc::channel(100);
        Self {
            id: WorkerId::new(),
            config,
            state: Arc::new(tokio::sync::Mutex::new(WorkerState::Starting)),
            tx,
            last_state: Arc::new(tokio::sync::Mutex::new(None)),
        }
    }

    /// Get the worker ID.
    pub fn id(&self) -> WorkerId {
        self.id
    }

    /// Get the worker name.
    pub fn name(&self) -> &str {
        &self.config.name
    }

    /// Get the current state.
    pub async fn state(&self) -> WorkerState {
        *self.state.lock().await
    }

    /// Set the worker state.
    pub async fn set_state(&self, new_state: WorkerState) {
        *self.state.lock().await = new_state;
    }

    /// Get the timeout as a Duration.
    pub fn timeout(&self) -> Duration {
        Duration::from_millis(self.config.timeout_ms)
    }

    /// Save worker state for recovery.
    pub async fn save_state(&self, state: Vec<u8>) -> SupervisionResult<()> {
        *self.last_state.lock().await = Some(state);
        Ok(())
    }

    /// Restore the last saved state.
    pub async fn restore_state(&self) -> SupervisionResult<Option<Vec<u8>>> {
        Ok(self.last_state.lock().await.clone())
    }

    /// Clear saved state.
    pub async fn clear_state(&self) {
        *self.last_state.lock().await = None;
    }

    /// Send a message to the worker (with timeout).
    pub async fn send_work(&self, payload: Vec<u8>) -> SupervisionResult<()> {
        let timeout = self.timeout();

        tokio::time::timeout(timeout, self.tx.send(WorkerMessage::Work(payload)))
            .await
            .map_err(|_| SupervisionError::WorkerTimeout {
                name: self.config.name.clone(),
                timeout_secs: self.config.timeout_ms / 1000,
            })?
            .map_err(|e| SupervisionError::ChannelError {
                reason: e.to_string(),
            })
    }

    /// Gracefully shut down the worker.
    pub async fn shutdown(&self) -> SupervisionResult<()> {
        self.set_state(WorkerState::Shutting).await;

        let timeout = Duration::from_secs(5);
        tokio::time::timeout(timeout, self.tx.send(WorkerMessage::Shutdown))
            .await
            .map_err(|_| SupervisionError::ShutdownError {
                reason: "Shutdown timeout".to_string(),
            })?
            .map_err(|e| SupervisionError::ShutdownError {
                reason: e.to_string(),
            })?;

        self.set_state(WorkerState::Stopped).await;
        Ok(())
    }
}

/// Worker task that runs in a supervised context.
#[derive(Debug, Clone)]
pub struct Worker {
    /// Worker configuration
    pub config: WorkerConfig,
    /// Receiver for worker messages
    rx: Arc<tokio::sync::Mutex<Option<mpsc::Receiver<WorkerMessage>>>>,
}

impl Worker {
    /// Create a new worker (called by supervisor).
    pub fn new(config: WorkerConfig) -> Self {
        let (_, rx) = mpsc::channel(100);
        Self {
            config,
            rx: Arc::new(tokio::sync::Mutex::new(Some(rx))),
        }
    }

    /// Get a mutable reference to the message receiver.
    pub async fn get_receiver(&self) -> Option<mpsc::Receiver<WorkerMessage>> {
        self.rx.lock().await.take()
    }

    /// Run the worker with a given work function.
    /// The work function receives the message payload and should return a result.
    pub async fn run<F, Fut>(&self, mut work_fn: F) -> SupervisionResult<()>
    where
        F: FnMut(Vec<u8>) -> Fut,
        Fut: std::future::Future<Output = SupervisionResult<()>>,
    {
        let mut receiver = self
            .get_receiver()
            .await
            .ok_or_else(|| SupervisionError::ChannelError {
                reason: "Failed to get message receiver".to_string(),
            })?;

        loop {
            match tokio::time::timeout(
                Duration::from_millis(self.config.timeout_ms),
                receiver.recv(),
            )
            .await
            {
                Ok(Some(WorkerMessage::Work(payload))) => {
                    work_fn(payload).await?;
                }
                Ok(Some(WorkerMessage::Shutdown)) => {
                    break;
                }
                Ok(None) => {
                    break;
                }
                Err(_) => {
                    return Err(SupervisionError::WorkerTimeout {
                        name: self.config.name.clone(),
                        timeout_secs: self.config.timeout_ms / 1000,
                    });
                }
            }
        }

        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_worker_id_unique() {
        let id1 = WorkerId::new();
        let id2 = WorkerId::new();
        assert_ne!(id1, id2);
    }

    #[test]
    fn test_worker_config_discovery() {
        let config = WorkerConfig::discovery();
        assert_eq!(config.name, "DiscoveryWorker");
        assert_eq!(config.timeout_ms, 5 * 60 * 1000);
    }

    #[test]
    fn test_worker_config_conformance() {
        let config = WorkerConfig::conformance();
        assert_eq!(config.name, "ConformanceWorker");
        assert_eq!(config.timeout_ms, 30 * 1000);
    }

    #[test]
    fn test_worker_config_io() {
        let config = WorkerConfig::io();
        assert_eq!(config.name, "IOWorker");
        assert_eq!(config.timeout_ms, 10 * 1000);
    }

    #[tokio::test]
    async fn test_worker_handle_state_transitions() {
        let config = WorkerConfig::new("test", 1000);
        let handle = WorkerHandle::new(config);

        assert_eq!(handle.state().await, WorkerState::Starting);

        handle.set_state(WorkerState::Running).await;
        assert_eq!(handle.state().await, WorkerState::Running);

        handle.set_state(WorkerState::Crashed).await;
        assert_eq!(handle.state().await, WorkerState::Crashed);
    }

    #[tokio::test]
    async fn test_worker_handle_state_recovery() {
        let config = WorkerConfig::new("test", 1000);
        let handle = WorkerHandle::new(config);

        let state_data = vec![1, 2, 3, 4, 5];
        handle.save_state(state_data.clone()).await.unwrap();

        let restored = handle.restore_state().await.unwrap();
        assert_eq!(restored, Some(state_data));

        handle.clear_state().await;
        let cleared = handle.restore_state().await.unwrap();
        assert_eq!(cleared, None);
    }

    #[tokio::test]
    async fn test_worker_timeout() {
        let config = WorkerConfig::new("test", 100); // 100ms timeout
        let handle = WorkerHandle::new(config);

        handle.set_state(WorkerState::Running).await;

        // Try to send work, but the receiver is not actually listening
        // This should timeout after 100ms
        let result = handle.send_work(vec![1, 2, 3]).await;
        assert!(result.is_err());
    }
}
