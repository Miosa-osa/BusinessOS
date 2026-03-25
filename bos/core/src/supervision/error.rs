//! Error types for supervision tree operations.

use thiserror::Error;

/// Result type for supervision operations.
pub type SupervisionResult<T> = Result<T, SupervisionError>;

/// Errors that can occur in supervision tree operations.
#[derive(Error, Debug, Clone)]
pub enum SupervisionError {
    #[error("Worker crashed: {reason}")]
    WorkerCrashed { reason: String },

    #[error("Supervisor crashed: {reason}")]
    SupervisorCrashed { reason: String },

    #[error("Restart limit exceeded: max {max_restarts} restarts in {window_secs}s")]
    RestartLimitExceeded {
        max_restarts: usize,
        window_secs: u64,
    },

    #[error("Worker timeout: {name} exceeded {timeout_secs}s")]
    WorkerTimeout {
        name: String,
        timeout_secs: u64,
    },

    #[error("Memory limit exceeded: {current_bytes} > {limit_bytes}")]
    MemoryLimitExceeded {
        current_bytes: usize,
        limit_bytes: usize,
    },

    #[error("Worker not found: {name}")]
    WorkerNotFound { name: String },

    #[error("Invalid configuration: {message}")]
    InvalidConfiguration { message: String },

    #[error("Channel error: {reason}")]
    ChannelError { reason: String },

    #[error("State recovery failed: {reason}")]
    StateRecoveryFailed { reason: String },

    #[error("Shutdown error: {reason}")]
    ShutdownError { reason: String },
}
