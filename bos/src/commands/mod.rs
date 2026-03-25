//! BOS Command Library — Complete command definitions and handlers.
//!
//! Provides 18+ commands for process mining, discovery, conformance checking,
//! statistics extraction, and ontology operations. All commands integrate with
//! the BusinessOS HTTP gateway for execution.
//!
//! ## Command Categories
//!
//! - **Process Discovery**: discover, discover_batch, list_models, validate_model
//! - **Conformance & Quality**: conform, check_conformance, statistics, quality_check
//! - **Analysis**: fingerprint, variability, org_evolution, variant_analysis
//! - **Export & Import**: export_petri_net, export_log, import_log, export_model
//! - **Ontology**: construct, execute, validate, compile
//! - **Batch Operations**: batch_discover, batch_conform, batch_statistics

pub mod businessos_commands;
pub mod execution;
pub mod formatting;
pub mod progress;

pub use businessos_commands::{BosCommand, BosCommandHandler};
pub use execution::CommandExecutor;
pub use formatting::{OutputFormat, ResultFormatter};
pub use progress::ProgressReporter;

/// Result type for command operations
pub type CommandResult<T> = std::result::Result<T, CommandError>;

/// BOS command errors
#[derive(Debug, Clone)]
pub enum CommandError {
    ExecutionFailed(String),
    ValidationFailed(String),
    GatewayError(String),
    SerializationError(String),
    FileIoError(String),
    Timeout(String),
    NotFound(String),
    InvalidArgument(String),
    InternalError(String),
}

impl std::fmt::Display for CommandError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            CommandError::ExecutionFailed(msg) => write!(f, "Execution failed: {}", msg),
            CommandError::ValidationFailed(msg) => write!(f, "Validation failed: {}", msg),
            CommandError::GatewayError(msg) => write!(f, "Gateway error: {}", msg),
            CommandError::SerializationError(msg) => write!(f, "Serialization error: {}", msg),
            CommandError::FileIoError(msg) => write!(f, "File I/O error: {}", msg),
            CommandError::Timeout(msg) => write!(f, "Timeout: {}", msg),
            CommandError::NotFound(msg) => write!(f, "Not found: {}", msg),
            CommandError::InvalidArgument(msg) => write!(f, "Invalid argument: {}", msg),
            CommandError::InternalError(msg) => write!(f, "Internal error: {}", msg),
        }
    }
}

impl std::error::Error for CommandError {}
