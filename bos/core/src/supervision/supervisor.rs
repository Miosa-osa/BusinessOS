//! Root supervisor for managing worker processes.
//!
//! Implements Joe Armstrong-style supervision tree with:
//! - Crash detection and recovery
//! - Configurable restart strategies
//! - Exponential backoff
//! - Memory bounds enforcement
//! - State recovery on restart

use std::collections::HashMap;
use std::sync::Arc;
use std::time::Duration;
use tokio::sync::RwLock;
use tokio::task::JoinHandle;
use uuid::Uuid;

use super::error::{SupervisionError, SupervisionResult};
use super::strategy::{RestartPolicy, RestartStrategy, RestartTracker};
use super::worker::{WorkerConfig, WorkerHandle, WorkerMessage, WorkerState};

/// Configuration for the supervisor.
#[derive(Debug, Clone)]
pub struct SupervisorConfig {
    /// Name of the supervisor
    pub name: String,

    /// Restart policy for children
    pub restart_policy: RestartPolicy,

    /// Maximum number of markings in reachability graph
    pub max_markings: usize,
}

impl SupervisorConfig {
    /// Create a new supervisor configuration.
    pub fn new(name: impl Into<String>) -> Self {
        Self {
            name: name.into(),
            restart_policy: RestartPolicy::default(),
            max_markings: 100_000,
        }
    }

    /// Root supervisor for the system.
    pub fn root() -> Self {
        Self::new("SupervisorRoot")
    }
}

/// Handle to interact with a supervisor.
pub struct SupervisorHandle {
    /// Supervisor unique ID
    id: Uuid,
    /// Supervisor configuration
    config: Arc<SupervisorConfig>,
    /// Active workers
    workers: Arc<RwLock<HashMap<String, WorkerHandle>>>,
    /// Restart trackers per worker
    restart_trackers: Arc<RwLock<HashMap<String, RestartTracker>>>,
    /// Supervision tasks
    tasks: Arc<RwLock<Vec<JoinHandle<SupervisionResult<()>>>>>,
}

impl SupervisorHandle {
    /// Create a new supervisor.
    pub fn new(config: SupervisorConfig) -> Self {
        Self {
            id: Uuid::new_v4(),
            config: Arc::new(config),
            workers: Arc::new(RwLock::new(HashMap::new())),
            restart_trackers: Arc::new(RwLock::new(HashMap::new())),
            tasks: Arc::new(RwLock::new(Vec::new())),
        }
    }

    /// Get the supervisor ID.
    pub fn id(&self) -> Uuid {
        self.id
    }

    /// Get the supervisor name.
    pub fn name(&self) -> &str {
        &self.config.name
    }

    /// Add a child worker to the supervisor.
    pub async fn add_worker(&self, config: WorkerConfig) -> SupervisionResult<WorkerHandle> {
        let worker = WorkerHandle::new(config);
        self.workers
            .write()
            .await
            .insert(worker.name().to_string(), worker.clone());

        let mut trackers = self.restart_trackers.write().await;
        trackers.insert(
            worker.name().to_string(),
            RestartTracker::new(&self.config.restart_policy),
        );

        Ok(worker)
    }

    /// Get a worker by name.
    pub async fn get_worker(&self, name: &str) -> SupervisionResult<WorkerHandle> {
        self.workers
            .read()
            .await
            .get(name)
            .cloned()
            .ok_or_else(|| SupervisionError::WorkerNotFound {
                name: name.to_string(),
            })
    }

    /// Monitor a worker for crashes and restart on failure.
    pub async fn supervise_worker<F>(
        &self,
        worker: WorkerHandle,
        work_fn: F,
    ) -> SupervisionResult<()>
    where
        F: Fn() -> std::pin::Pin<Box<dyn std::future::Future<Output = SupervisionResult<()>> + Send>>
            + Send
            + Sync
            + 'static,
    {
        let worker_name = worker.name().to_string();
        let config = self.config.clone();
        let trackers = self.restart_trackers.clone();
        let workers = self.workers.clone();
        let work_fn = Arc::new(work_fn);

        let supervision_task = tokio::spawn(async move {
            loop {
                // Run the worker
                worker.set_state(WorkerState::Running).await;
                let result = work_fn().await;

                // Worker exited (crashed or shutdown)
                match result {
                    Ok(()) => {
                        // Clean shutdown
                        worker.set_state(WorkerState::Stopped).await;
                        break;
                    }
                    Err(e) => {
                        worker.set_state(WorkerState::Crashed).await;

                        // Check restart limit
                        let mut tracker = trackers.write().await;
                        let tracker_entry = tracker
                            .entry(worker_name.clone())
                            .or_insert_with(|| RestartTracker::new(&config.restart_policy));

                        match tracker_entry.record_restart() {
                            Ok(attempt) => {
                                let backoff =
                                    config.restart_policy.backoff_delay(attempt - 1);
                                tracing::warn!(
                                    worker = %worker_name,
                                    attempt = attempt,
                                    error = %e,
                                    backoff_ms = backoff.as_millis(),
                                    "Worker crashed, restarting with backoff"
                                );

                                worker.set_state(WorkerState::Restarting).await;
                                drop(tracker); // Release lock before sleeping
                                tokio::time::sleep(backoff).await;
                            }
                            Err(restart_count) => {
                                tracing::error!(
                                    worker = %worker_name,
                                    max_restarts = config.restart_policy.max_restarts,
                                    window_secs = config.restart_policy.time_window_secs,
                                    "Worker exceeded restart limit"
                                );
                                drop(tracker);

                                // Remove from active workers
                                workers.write().await.remove(&worker_name);

                                return Err(SupervisionError::RestartLimitExceeded {
                                    max_restarts: config.restart_policy.max_restarts,
                                    window_secs: config.restart_policy.time_window_secs,
                                });
                            }
                        }
                    }
                }
            }

            Ok(())
        });

        let mut tasks = self.tasks.write().await;
        tasks.push(supervision_task);

        Ok(())
    }

    /// Wait for all supervision tasks to complete.
    pub async fn wait_all(&self) -> SupervisionResult<()> {
        let mut tasks = self.tasks.write().await;
        while let Some(task) = tasks.pop() {
            task.await.map_err(|e| SupervisionError::SupervisorCrashed {
                reason: e.to_string(),
            })??;
        }
        Ok(())
    }

    /// Gracefully shut down all workers.
    pub async fn shutdown(&self) -> SupervisionResult<()> {
        let workers = self.workers.read().await;
        for worker in workers.values() {
            worker.shutdown().await?;
        }
        Ok(())
    }

    /// Get current state of all workers.
    pub async fn worker_states(&self) -> SupervisionResult<Vec<(String, WorkerState)>> {
        let workers = self.workers.read().await;
        let mut states = Vec::new();
        for (name, worker) in workers.iter() {
            let state = worker.state().await;
            states.push((name.clone(), state));
        }
        Ok(states)
    }

    /// Get restart statistics for a worker.
    pub async fn restart_stats(&self, name: &str) -> SupervisionResult<usize> {
        let trackers = self.restart_trackers.read().await;
        let tracker = trackers.get(name).ok_or_else(|| {
            SupervisionError::WorkerNotFound {
                name: name.to_string(),
            }
        })?;
        Ok(tracker.current_restarts())
    }
}

impl Clone for SupervisorHandle {
    fn clone(&self) -> Self {
        Self {
            id: self.id,
            config: self.config.clone(),
            workers: self.workers.clone(),
            restart_trackers: self.restart_trackers.clone(),
            tasks: self.tasks.clone(),
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_supervisor_creation() {
        let config = SupervisorConfig::root();
        let supervisor = SupervisorHandle::new(config);

        assert_eq!(supervisor.name(), "SupervisorRoot");
    }

    #[tokio::test]
    async fn test_add_worker() {
        let config = SupervisorConfig::root();
        let supervisor = SupervisorHandle::new(config);

        let worker_config = WorkerConfig::discovery();
        let worker = supervisor.add_worker(worker_config).await.unwrap();

        assert_eq!(worker.name(), "DiscoveryWorker");
        assert_eq!(worker.state().await, WorkerState::Starting);
    }

    #[tokio::test]
    async fn test_get_worker() {
        let config = SupervisorConfig::root();
        let supervisor = SupervisorHandle::new(config);

        let worker_config = WorkerConfig::conformance();
        supervisor.add_worker(worker_config).await.unwrap();

        let worker = supervisor.get_worker("ConformanceWorker").await;
        assert!(worker.is_ok());
        assert_eq!(worker.unwrap().name(), "ConformanceWorker");
    }

    #[tokio::test]
    async fn test_get_nonexistent_worker() {
        let config = SupervisorConfig::root();
        let supervisor = SupervisorHandle::new(config);

        let result = supervisor.get_worker("NonExistent").await;
        assert!(result.is_err());
    }

    #[tokio::test]
    async fn test_worker_states() {
        let config = SupervisorConfig::root();
        let supervisor = SupervisorHandle::new(config);

        supervisor.add_worker(WorkerConfig::discovery()).await.unwrap();
        supervisor.add_worker(WorkerConfig::conformance()).await.unwrap();

        let states = supervisor.worker_states().await.unwrap();
        assert_eq!(states.len(), 2);

        let state_map: HashMap<_, _> = states.into_iter().collect();
        assert_eq!(
            state_map.get("DiscoveryWorker"),
            Some(&WorkerState::Starting)
        );
        assert_eq!(
            state_map.get("ConformanceWorker"),
            Some(&WorkerState::Starting)
        );
    }
}
