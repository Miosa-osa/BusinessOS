//! Supervision tree architecture for fault tolerance (Joe Armstrong style).
//!
//! Implements Erlang/OTP-style supervision with:
//! - Process crash detection and recovery
//! - Configurable restart strategies (one-for-one, one-for-all, rest-for-one)
//! - Exponential backoff with max restart limits
//! - Timeout handling and graceful shutdown
//! - Memory bounds enforcement
//! - State recovery on restart

pub mod strategy;
pub mod supervisor;
pub mod worker;
pub mod error;

pub use strategy::{RestartStrategy, RestartPolicy};
pub use supervisor::{SupervisorConfig, SupervisorHandle};
pub use worker::{Worker, WorkerConfig, WorkerHandle, WorkerState};
pub use error::{SupervisionError, SupervisionResult};
