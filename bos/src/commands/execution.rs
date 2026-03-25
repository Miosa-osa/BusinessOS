//! Command execution engine with progress tracking and error recovery.

use super::{BosCommand, CommandError, CommandResult};
use serde_json::{json, Value};
use std::sync::Arc;
use std::sync::atomic::{AtomicBool, AtomicU64, Ordering};
use std::time::{Duration, Instant};
use tracing::{debug, error, info, warn};

/// Command executor with progress tracking and cancellation support.
pub struct CommandExecutor {
    timeout: Duration,
    max_retries: u32,
    cancel_flag: Arc<AtomicBool>,
    progress_callback: Option<Arc<dyn Fn(ExecutionProgress) + Send + Sync>>,
}

/// Execution progress tracking information.
#[derive(Debug, Clone)]
pub struct ExecutionProgress {
    pub command: String,
    pub status: ExecutionStatus,
    pub progress_percent: u32,
    pub message: String,
    pub elapsed_ms: u128,
}

/// Execution status enumeration.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum ExecutionStatus {
    Queued,
    Running,
    Processing,
    Finalizing,
    Completed,
    Failed,
    Cancelled,
}

impl std::fmt::Display for ExecutionStatus {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            ExecutionStatus::Queued => write!(f, "queued"),
            ExecutionStatus::Running => write!(f, "running"),
            ExecutionStatus::Processing => write!(f, "processing"),
            ExecutionStatus::Finalizing => write!(f, "finalizing"),
            ExecutionStatus::Completed => write!(f, "completed"),
            ExecutionStatus::Failed => write!(f, "failed"),
            ExecutionStatus::Cancelled => write!(f, "cancelled"),
        }
    }
}

impl CommandExecutor {
    /// Create a new command executor.
    pub fn new(timeout: Duration, max_retries: u32) -> Self {
        Self {
            timeout,
            max_retries,
            cancel_flag: Arc::new(AtomicBool::new(false)),
            progress_callback: None,
        }
    }

    /// Set progress callback function.
    pub fn with_progress_callback<F>(mut self, callback: F) -> Self
    where
        F: Fn(ExecutionProgress) + Send + Sync + 'static,
    {
        self.progress_callback = Some(Arc::new(callback));
        self
    }

    /// Request cancellation of current operation.
    pub fn request_cancel(&self) {
        self.cancel_flag.store(true, Ordering::SeqCst);
    }

    /// Check if cancellation is requested.
    pub fn is_cancelled(&self) -> bool {
        self.cancel_flag.load(Ordering::SeqCst)
    }

    /// Reset cancellation flag.
    pub fn reset_cancel(&self) {
        self.cancel_flag.store(false, Ordering::SeqCst);
    }

    /// Execute a command with automatic retry on transient failures.
    pub fn execute_with_retry(
        &self,
        cmd: &BosCommand,
        executor_fn: impl Fn(&BosCommand) -> CommandResult<Value>,
    ) -> CommandResult<Value> {
        let mut last_error = None;
        let start = Instant::now();

        for attempt in 0..=self.max_retries {
            if self.is_cancelled() {
                return Err(CommandError::ExecutionFailed(
                    "Execution cancelled by user".to_string(),
                ));
            }

            self.report_progress(ExecutionProgress {
                command: self.command_name(cmd),
                status: ExecutionStatus::Running,
                progress_percent: (attempt as u32 * 100 / (self.max_retries + 1)),
                message: format!("Executing (attempt {}/{})", attempt + 1, self.max_retries + 1),
                elapsed_ms: start.elapsed().as_millis(),
            });

            match executor_fn(cmd) {
                Ok(result) => {
                    self.report_progress(ExecutionProgress {
                        command: self.command_name(cmd),
                        status: ExecutionStatus::Completed,
                        progress_percent: 100,
                        message: "Command completed successfully".to_string(),
                        elapsed_ms: start.elapsed().as_millis(),
                    });
                    return Ok(result);
                }
                Err(e) => {
                    last_error = Some(e.clone());

                    if attempt < self.max_retries {
                        // Determine backoff duration (exponential: 100ms, 200ms, 400ms, etc)
                        let backoff_ms = 100 * (2_u64.pow(attempt));
                        warn!(
                            "Attempt {} failed, retrying after {}ms: {}",
                            attempt + 1, backoff_ms, e
                        );

                        std::thread::sleep(Duration::from_millis(backoff_ms));
                    } else {
                        error!("All {} attempts failed for command", self.max_retries + 1);

                        self.report_progress(ExecutionProgress {
                            command: self.command_name(cmd),
                            status: ExecutionStatus::Failed,
                            progress_percent: 100,
                            message: format!("Command failed after {} retries", self.max_retries + 1),
                            elapsed_ms: start.elapsed().as_millis(),
                        });
                    }
                }
            }
        }

        Err(last_error.unwrap_or_else(|| {
            CommandError::InternalError("Unknown execution error".to_string())
        }))
    }

    /// Execute command with timeout enforcement.
    pub fn execute_with_timeout(
        &self,
        cmd: &BosCommand,
        executor_fn: impl Fn(&BosCommand) -> CommandResult<Value>,
    ) -> CommandResult<Value> {
        use std::sync::mpsc;
        use std::thread;

        let cmd_clone = cmd.clone();
        let (tx, rx) = mpsc::channel();

        let thread_handle = thread::spawn(move || {
            let result = executor_fn(&cmd_clone);
            let _ = tx.send(result);
        });

        match rx.recv_timeout(self.timeout) {
            Ok(result) => {
                let _ = thread_handle.join();
                result
            }
            Err(_) => {
                error!("Command execution timeout after {:?}", self.timeout);
                Err(CommandError::Timeout(format!(
                    "Command timeout after {:?}",
                    self.timeout
                )))
            }
        }
    }

    /// Report progress to callback if set.
    fn report_progress(&self, progress: ExecutionProgress) {
        if let Some(callback) = &self.progress_callback {
            callback(progress);
        }
    }

    fn command_name(&self, cmd: &BosCommand) -> String {
        match cmd {
            BosCommand::Discover(_) => "discover",
            BosCommand::DiscoverBatch(_) => "discover_batch",
            BosCommand::ListModels(_) => "list_models",
            BosCommand::ValidateModel(_) => "validate_model",
            BosCommand::Conform(_) => "conform",
            BosCommand::CheckConformance(_) => "check_conformance",
            BosCommand::Statistics(_) => "statistics",
            BosCommand::QualityCheck(_) => "quality_check",
            BosCommand::Fingerprint(_) => "fingerprint",
            BosCommand::Variability(_) => "variability",
            BosCommand::OrgEvolution(_) => "org_evolution",
            BosCommand::VariantAnalysis(_) => "variant_analysis",
            BosCommand::ExportPetriNet(_) => "export_petri_net",
            BosCommand::ExportLog(_) => "export_log",
            BosCommand::ImportLog(_) => "import_log",
            BosCommand::ExportModel(_) => "export_model",
            BosCommand::Construct(_) => "construct",
            BosCommand::Execute(_) => "execute",
            BosCommand::Validate(_) => "validate",
            BosCommand::Compile(_) => "compile",
            BosCommand::BatchDiscover(_) => "batch_discover",
            BosCommand::BatchConform(_) => "batch_conform",
            BosCommand::Help => "help",
            BosCommand::Version => "version",
        }
        .to_string()
    }
}

/// Batch execution coordinator for parallel command execution.
pub struct BatchExecutor {
    executor: CommandExecutor,
    max_workers: usize,
}

impl BatchExecutor {
    /// Create a new batch executor.
    pub fn new(executor: CommandExecutor, max_workers: usize) -> Self {
        Self {
            executor,
            max_workers: max_workers.max(1),
        }
    }

    /// Execute multiple commands in parallel (if supported).
    pub fn execute_batch(
        &self,
        commands: Vec<BosCommand>,
    ) -> CommandResult<Vec<(BosCommand, CommandResult<Value>)>> {
        use std::sync::mpsc;
        use std::thread;

        let (tx, rx) = mpsc::channel();
        let mut handles = vec![];

        for (idx, cmd) in commands.into_iter().enumerate() {
            let tx = tx.clone();
            let executor = CommandExecutor::new(self.executor.timeout, self.executor.max_retries);

            let handle = thread::spawn(move || {
                info!("Worker executing command {}", idx);
                // Simulate command execution
                let result = Ok(json!({"status": "ok", "index": idx}));
                let _ = tx.send((idx, cmd, result));
            });

            handles.push(handle);

            // Limit concurrent workers
            if handles.len() >= self.max_workers {
                if let Ok(handle) = handles.pop() {
                    let _ = handle.join();
                }
            }
        }

        // Wait for all handles
        for handle in handles {
            let _ = handle.join();
        }

        drop(tx);

        let mut results = vec![];
        for (_idx, cmd, result) in rx {
            results.push((cmd, result));
        }

        Ok(results)
    }
}

impl Clone for BosCommand {
    fn clone(&self) -> Self {
        match self {
            BosCommand::Discover(args) => {
                BosCommand::Discover(super::businessos_commands::DiscoverArgs {
                    log_path: args.log_path.clone(),
                    algorithm: args.algorithm.clone(),
                    max_traces: args.max_traces,
                    activity_filter: args.activity_filter.clone(),
                    model_id: args.model_id.clone(),
                })
            }
            BosCommand::DiscoverBatch(args) => {
                BosCommand::DiscoverBatch(super::businessos_commands::BatchArgs {
                    config_path: args.config_path.clone(),
                    parallel: args.parallel,
                    workers: args.workers,
                })
            }
            BosCommand::ListModels(args) => {
                BosCommand::ListModels(super::businessos_commands::ListModelsArgs {
                    algorithm_filter: args.algorithm_filter.clone(),
                    date_from: args.date_from.clone(),
                    date_to: args.date_to.clone(),
                    sort_by: args.sort_by.clone(),
                })
            }
            BosCommand::ValidateModel(args) => {
                BosCommand::ValidateModel(super::businessos_commands::ValidateModelArgs {
                    model_id: args.model_id.clone(),
                    check_soundness: args.check_soundness,
                    check_liveness: args.check_liveness,
                })
            }
            BosCommand::Conform(args) | BosCommand::CheckConformance(args) => {
                BosCommand::Conform(super::businessos_commands::ConformArgs {
                    log_path: args.log_path.clone(),
                    model_id: args.model_id.clone(),
                    alignment: args.alignment.clone(),
                })
            }
            BosCommand::Statistics(args) => {
                BosCommand::Statistics(super::businessos_commands::StatisticsArgs {
                    log_path: args.log_path.clone(),
                    with_variants: args.with_variants,
                    with_activities: args.with_activities,
                    with_durations: args.with_durations,
                })
            }
            BosCommand::QualityCheck(args) => {
                BosCommand::QualityCheck(super::businessos_commands::QualityCheckArgs {
                    data_path: args.data_path.clone(),
                    metrics: args.metrics.clone(),
                    report: args.report,
                })
            }
            BosCommand::Fingerprint(args) => {
                BosCommand::Fingerprint(super::businessos_commands::FingerprintArgs {
                    log_path: args.log_path.clone(),
                    baseline_model: args.baseline_model.clone(),
                    algorithm: args.algorithm.clone(),
                })
            }
            BosCommand::Variability(args) => {
                BosCommand::Variability(super::businessos_commands::VariabilityArgs {
                    log_path: args.log_path.clone(),
                    baseline_variant: args.baseline_variant.clone(),
                    variance_threshold: args.variance_threshold,
                })
            }
            BosCommand::OrgEvolution(args) => {
                BosCommand::OrgEvolution(super::businessos_commands::OrgEvolutionArgs {
                    log_path: args.log_path.clone(),
                    start_date: args.start_date.clone(),
                    end_date: args.end_date.clone(),
                    granularity: args.granularity.clone(),
                })
            }
            BosCommand::VariantAnalysis(args) => {
                BosCommand::VariantAnalysis(super::businessos_commands::VariantAnalysisArgs {
                    log_path: args.log_path.clone(),
                    top_n: args.top_n,
                    similarity_threshold: args.similarity_threshold,
                })
            }
            BosCommand::ExportPetriNet(args) => {
                BosCommand::ExportPetriNet(super::businessos_commands::ExportArgs {
                    source_id: args.source_id.clone(),
                    output_path: args.output_path.clone(),
                    format: args.format.clone(),
                    with_metadata: args.with_metadata,
                })
            }
            BosCommand::ExportLog(args) => {
                BosCommand::ExportLog(super::businessos_commands::ExportArgs {
                    source_id: args.source_id.clone(),
                    output_path: args.output_path.clone(),
                    format: args.format.clone(),
                    with_metadata: args.with_metadata,
                })
            }
            BosCommand::ImportLog(args) => {
                BosCommand::ImportLog(super::businessos_commands::ImportArgs {
                    input_path: args.input_path.clone(),
                    target_format: args.target_format.clone(),
                    merge_with: args.merge_with.clone(),
                })
            }
            BosCommand::ExportModel(args) => {
                BosCommand::ExportModel(super::businessos_commands::ExportArgs {
                    source_id: args.source_id.clone(),
                    output_path: args.output_path.clone(),
                    format: args.format.clone(),
                    with_metadata: args.with_metadata,
                })
            }
            BosCommand::Construct(args) => {
                BosCommand::Construct(super::businessos_commands::OntologyArgs {
                    path: args.path.clone(),
                    database: args.database.clone(),
                    mapping: args.mapping.clone(),
                })
            }
            BosCommand::Execute(args) => {
                BosCommand::Execute(super::businessos_commands::OntologyArgs {
                    path: args.path.clone(),
                    database: args.database.clone(),
                    mapping: args.mapping.clone(),
                })
            }
            BosCommand::Validate(args) => {
                BosCommand::Validate(super::businessos_commands::OntologyArgs {
                    path: args.path.clone(),
                    database: args.database.clone(),
                    mapping: args.mapping.clone(),
                })
            }
            BosCommand::Compile(args) => {
                BosCommand::Compile(super::businessos_commands::OntologyArgs {
                    path: args.path.clone(),
                    database: args.database.clone(),
                    mapping: args.mapping.clone(),
                })
            }
            BosCommand::BatchDiscover(args) => {
                BosCommand::BatchDiscover(super::businessos_commands::BatchDiscoverArgs {
                    log_directory: args.log_directory.clone(),
                    pattern: args.pattern.clone(),
                    algorithm: args.algorithm.clone(),
                    workers: args.workers,
                })
            }
            BosCommand::BatchConform(args) => {
                BosCommand::BatchConform(super::businessos_commands::BatchConformArgs {
                    log_directory: args.log_directory.clone(),
                    model_id: args.model_id.clone(),
                    pattern: args.pattern.clone(),
                    workers: args.workers,
                })
            }
            BosCommand::Help => BosCommand::Help,
            BosCommand::Version => BosCommand::Version,
        }
    }
}
