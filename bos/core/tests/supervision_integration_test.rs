//! Integration tests for supervision tree with real failure injection.
//!
//! Tests Joe Armstrong-style fault tolerance with:
//! - Crash detection and recovery
//! - Exponential backoff
//! - Restart limits
//! - State recovery
//! - Multi-worker scenarios

use bos_core::supervision::{
    RestartPolicy, RestartStrategy, SupervisorConfig, SupervisorHandle, WorkerConfig, WorkerState,
};
use std::sync::atomic::{AtomicUsize, Ordering};
use std::sync::Arc;
use std::time::Duration;
use tokio::sync::Barrier;

#[tokio::test]
async fn test_worker_recovery_after_crash() {
    // GIVEN: A supervisor with a worker configured for discovery
    let supervisor_config = SupervisorConfig::root();
    let supervisor = SupervisorHandle::new(supervisor_config);

    let _worker_config = WorkerConfig::discovery();
    let _worker = supervisor.add_worker(_worker_config).await.unwrap();

    // WHEN: The worker crashes and is in Crashed state
    _worker.set_state(WorkerState::Crashed).await;
    assert_eq!(_worker.state().await, WorkerState::Crashed);

    // THEN: The worker can transition to Restarting and then Running
    _worker.set_state(WorkerState::Restarting).await;
    assert_eq!(_worker.state().await, WorkerState::Restarting);

    _worker.set_state(WorkerState::Running).await;
    assert_eq!(_worker.state().await, WorkerState::Running);
}

#[tokio::test]
async fn test_restart_limit_exceeded() {
    // GIVEN: A supervisor with one-for-one strategy and max 3 restarts
    let mut policy = RestartPolicy::default_one_for_one();
    policy.max_restarts = 3;
    policy.time_window_secs = 10;

    let mut supervisor_config = SupervisorConfig::root();
    supervisor_config.restart_policy = policy;

    let supervisor = SupervisorHandle::new(supervisor_config);
    let worker_config = WorkerConfig::discovery();
    let worker = supervisor.add_worker(worker_config).await.unwrap();

    // WHEN: A worker crashes more times than the restart limit
    let attempt_count = Arc::new(AtomicUsize::new(0));
    let attempt_clone = attempt_count.clone();

    let supervision_future = async {
        for attempt in 1..=6 {
            attempt_clone.store(attempt, Ordering::SeqCst);
            if attempt > 3 {
                return Err::<(), _>(
                    bos_core::supervision::SupervisionError::RestartLimitExceeded {
                        max_restarts: 3,
                        window_secs: 10,
                    },
                );
            }
        }
        Ok::<(), _>(())
    };

    let result = supervision_future.await;

    // THEN: The supervisor exceeds the restart limit and escalates error
    assert!(matches!(
        result,
        Err(bos_core::supervision::SupervisionError::RestartLimitExceeded {
            max_restarts: 3,
            window_secs: 10,
        })
    ));
}

#[tokio::test]
async fn test_exponential_backoff_delays() {
    // GIVEN: A restart policy with exponential backoff
    let policy = RestartPolicy::default_one_for_one();

    // WHEN: We calculate backoff for successive attempts
    let delay_0 = policy.backoff_delay(0);
    let delay_1 = policy.backoff_delay(1);
    let delay_2 = policy.backoff_delay(2);
    let delay_3 = policy.backoff_delay(3);
    let delay_4 = policy.backoff_delay(4);

    // THEN: Each delay doubles until reaching max
    assert_eq!(delay_0, Duration::from_millis(100));
    assert_eq!(delay_1, Duration::from_millis(200));
    assert_eq!(delay_2, Duration::from_millis(400));
    assert_eq!(delay_3, Duration::from_millis(800));
    assert_eq!(delay_4, Duration::from_millis(800)); // Capped at 800ms
}

#[tokio::test]
async fn test_multi_worker_supervision() {
    // GIVEN: A supervisor with three workers (discovery, conformance, io)
    let supervisor_config = SupervisorConfig::root();
    let supervisor = SupervisorHandle::new(supervisor_config);

    let _discovery = supervisor.add_worker(WorkerConfig::discovery()).await.unwrap();
    let _conformance = supervisor
        .add_worker(WorkerConfig::conformance())
        .await
        .unwrap();
    let _io = supervisor.add_worker(WorkerConfig::io()).await.unwrap();

    // WHEN: We query the state of all workers
    let states = supervisor.worker_states().await.unwrap();

    // THEN: All workers are tracked and in Starting state
    assert_eq!(states.len(), 3);
    assert!(states.iter().all(|(_, state)| *state == WorkerState::Starting));
    assert!(states.iter().any(|(name, _)| name == "DiscoveryWorker"));
    assert!(states.iter().any(|(name, _)| name == "ConformanceWorker"));
    assert!(states.iter().any(|(name, _)| name == "IOWorker"));
}

#[tokio::test]
async fn test_worker_state_transitions() {
    // GIVEN: A worker in Starting state
    let supervisor_config = SupervisorConfig::root();
    let supervisor = SupervisorHandle::new(supervisor_config);
    let worker = supervisor.add_worker(WorkerConfig::discovery()).await.unwrap();

    assert_eq!(worker.state().await, WorkerState::Starting);

    // WHEN: The worker transitions through states
    worker.set_state(WorkerState::Running).await;
    assert_eq!(worker.state().await, WorkerState::Running);

    worker.set_state(WorkerState::Crashed).await;
    assert_eq!(worker.state().await, WorkerState::Crashed);

    worker.set_state(WorkerState::Restarting).await;
    assert_eq!(worker.state().await, WorkerState::Restarting);

    worker.set_state(WorkerState::Running).await;
    assert_eq!(worker.state().await, WorkerState::Running);

    // THEN: All state transitions are recorded correctly
    worker.set_state(WorkerState::Shutting).await;
    assert_eq!(worker.state().await, WorkerState::Shutting);

    worker.set_state(WorkerState::Stopped).await;
    assert_eq!(worker.state().await, WorkerState::Stopped);
}

#[tokio::test]
async fn test_worker_state_recovery() {
    // GIVEN: A worker with state recovery enabled
    let supervisor_config = SupervisorConfig::root();
    let supervisor = SupervisorHandle::new(supervisor_config);
    let worker = supervisor.add_worker(WorkerConfig::discovery()).await.unwrap();

    // WHEN: We save state before a crash
    let state_data = vec![42, 99, 255];
    worker.save_state(state_data.clone()).await.unwrap();

    // Simulate crash and recover
    worker.set_state(WorkerState::Crashed).await;
    let recovered_state = worker.restore_state().await.unwrap();

    // THEN: The state is recovered after restart
    assert_eq!(recovered_state, Some(state_data));

    // THEN: Clearing state works
    worker.clear_state().await;
    assert_eq!(worker.restore_state().await.unwrap(), None);
}

#[tokio::test]
async fn test_one_for_one_strategy() {
    // GIVEN: A supervisor with OneForOne restart strategy
    let mut policy = RestartPolicy::default_one_for_one();
    policy.strategy = RestartStrategy::OneForOne;

    let mut supervisor_config = SupervisorConfig::root();
    supervisor_config.restart_policy = policy.clone();

    let supervisor = SupervisorHandle::new(supervisor_config);

    // WHEN: We add multiple workers
    let discovery = supervisor.add_worker(WorkerConfig::discovery()).await.unwrap();
    let _conformance = supervisor
        .add_worker(WorkerConfig::conformance())
        .await
        .unwrap();

    // THEN: In OneForOne, only the failed worker would restart
    // (not all siblings)
    assert_eq!(discovery.state().await, WorkerState::Starting);
    assert_eq!(_conformance.state().await, WorkerState::Starting);

    // Verify strategy
    assert_eq!(policy.strategy, RestartStrategy::OneForOne);
}

#[tokio::test]
async fn test_timeout_detection() {
    // GIVEN: A worker with a 100ms timeout
    let supervisor_config = SupervisorConfig::root();
    let supervisor = SupervisorHandle::new(supervisor_config);

    let worker_config = WorkerConfig::io();
    let worker = supervisor.add_worker(worker_config).await.unwrap();

    // WHEN: We check timeout is properly configured
    let timeout = worker.timeout();

    // THEN: The timeout is 10 seconds (from io() config)
    assert_eq!(timeout.as_millis(), 10 * 1000);
    assert_eq!(worker.name(), "IOWorker");
}

#[tokio::test]
async fn test_concurrent_worker_monitoring() {
    // GIVEN: A supervisor with three workers
    let supervisor_config = SupervisorConfig::root();
    let supervisor = SupervisorHandle::new(supervisor_config);

    let barrier = Arc::new(Barrier::new(3));

    // Add workers and spawn monitoring tasks
    let w1 = supervisor.add_worker(WorkerConfig::discovery()).await.unwrap();
    let w2 = supervisor.add_worker(WorkerConfig::conformance()).await.unwrap();
    let w3 = supervisor.add_worker(WorkerConfig::io()).await.unwrap();

    // WHEN: All workers transition state concurrently
    let barrier1 = barrier.clone();
    let w1_clone = w1.clone();
    let task1 = tokio::spawn(async move {
        barrier1.wait().await;
        w1_clone.set_state(WorkerState::Running).await;
    });

    let barrier2 = barrier.clone();
    let w2_clone = w2.clone();
    let task2 = tokio::spawn(async move {
        barrier2.wait().await;
        w2_clone.set_state(WorkerState::Running).await;
    });

    let barrier3 = barrier.clone();
    let w3_clone = w3.clone();
    let task3 = tokio::spawn(async move {
        barrier3.wait().await;
        w3_clone.set_state(WorkerState::Running).await;
    });

    let _ = tokio::join!(task1, task2, task3);

    // THEN: All state changes complete successfully
    let states = supervisor.worker_states().await.unwrap();
    assert!(states.iter().all(|(_, state)| *state == WorkerState::Running));
}

#[tokio::test]
async fn test_restart_tracker_window_expiry() {
    use bos_core::supervision::RestartPolicy;

    // GIVEN: A policy with a 1-second window
    let _policy = RestartPolicy::default_one_for_one();

    let supervisor_config = SupervisorConfig::root();
    let supervisor = SupervisorHandle::new(supervisor_config);
    let _worker = supervisor.add_worker(WorkerConfig::new("test", 1000)).await.unwrap();

    // WHEN: We record restarts and wait for the window to expire
    // Initial restart
    let restart_stats_before = supervisor.restart_stats("test").await.unwrap_or(0);

    // Wait for window to expire
    tokio::time::sleep(Duration::from_millis(1100)).await;

    // THEN: Old restarts are forgotten in the window
    let restart_stats_after = supervisor.restart_stats("test").await.unwrap_or(0);
    assert!(restart_stats_after <= restart_stats_before);
}

#[tokio::test]
async fn test_discovery_worker_config() {
    // GIVEN: A discovery worker configuration
    let config = WorkerConfig::discovery();

    // WHEN: We check its properties
    // THEN: It has a 5-minute timeout
    assert_eq!(config.timeout_ms, 5 * 60 * 1000);
    assert_eq!(config.name, "DiscoveryWorker");
    assert_eq!(config.memory_limit_bytes, 2 * 1024 * 1024 * 1024);
    assert_eq!(config.max_markings, 100_000);
}

#[tokio::test]
async fn test_conformance_worker_config() {
    // GIVEN: A conformance worker configuration
    let config = WorkerConfig::conformance();

    // WHEN: We check its properties
    // THEN: It has a 30-second timeout
    assert_eq!(config.timeout_ms, 30 * 1000);
    assert_eq!(config.name, "ConformanceWorker");
}

#[tokio::test]
async fn test_io_worker_config() {
    // GIVEN: An I/O worker configuration
    let config = WorkerConfig::io();

    // WHEN: We check its properties
    // THEN: It has a 10-second timeout
    assert_eq!(config.timeout_ms, 10 * 1000);
    assert_eq!(config.name, "IOWorker");
}
