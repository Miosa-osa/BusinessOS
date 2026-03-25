//! Restart strategies for supervision tree (Joe Armstrong style).
//!
//! Implements three core strategies:
//! - OneForOne: Restart only the failed child
//! - OneForAll: Restart failed child and all siblings
//! - RestForOne: Restart failed child and those started after it

use std::time::{Duration, Instant};
use serde::{Deserialize, Serialize};

/// Restart strategy for a supervisor.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum RestartStrategy {
    /// Restart only the failed child process
    OneForOne,
    /// Restart failed child and terminate all sibling processes
    OneForAll,
    /// Restart failed child and all children started after it
    RestForOne,
}

/// Policy controlling restart behavior and limits.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RestartPolicy {
    /// Strategy to use when a child crashes
    pub strategy: RestartStrategy,

    /// Maximum number of restarts allowed in the time window
    pub max_restarts: usize,

    /// Time window (seconds) for restart counting
    pub time_window_secs: u64,

    /// Initial backoff delay (milliseconds)
    pub initial_backoff_ms: u64,

    /// Maximum backoff delay (milliseconds)
    pub max_backoff_ms: u64,

    /// Enable exponential backoff (doubles on each restart)
    pub exponential_backoff: bool,
}

impl RestartPolicy {
    /// Default policy: OneForOne with 5 restarts in 60 seconds, exponential backoff.
    pub fn default_one_for_one() -> Self {
        Self {
            strategy: RestartStrategy::OneForOne,
            max_restarts: 5,
            time_window_secs: 60,
            initial_backoff_ms: 100,
            max_backoff_ms: 800,
            exponential_backoff: true,
        }
    }

    /// Strict policy: OneForAll with 3 restarts in 60 seconds.
    pub fn strict_one_for_all() -> Self {
        Self {
            strategy: RestartStrategy::OneForAll,
            max_restarts: 3,
            time_window_secs: 60,
            initial_backoff_ms: 100,
            max_backoff_ms: 800,
            exponential_backoff: true,
        }
    }

    /// Calculate backoff delay for restart attempt number.
    pub fn backoff_delay(&self, attempt: usize) -> Duration {
        if !self.exponential_backoff {
            return Duration::from_millis(self.initial_backoff_ms);
        }

        let backoff = self.initial_backoff_ms * 2_u64.saturating_pow(attempt as u32);
        let capped = backoff.min(self.max_backoff_ms);
        Duration::from_millis(capped)
    }
}

impl Default for RestartPolicy {
    fn default() -> Self {
        Self::default_one_for_one()
    }
}

/// Tracks restart history for a process.
#[derive(Debug, Clone)]
pub struct RestartTracker {
    /// Time window for counting restarts
    window: Duration,
    /// Maximum restarts allowed in window
    max_restarts: usize,
    /// Timestamps of recent restarts
    restart_times: Vec<Instant>,
}

impl RestartTracker {
    /// Create a new restart tracker with the given policy.
    pub fn new(policy: &RestartPolicy) -> Self {
        Self {
            window: Duration::from_secs(policy.time_window_secs),
            max_restarts: policy.max_restarts,
            restart_times: Vec::with_capacity(policy.max_restarts),
        }
    }

    /// Record a restart and check if limit is exceeded.
    /// Returns Ok(attempt_number) or Err if limit exceeded.
    pub fn record_restart(&mut self) -> Result<usize, usize> {
        let now = Instant::now();
        let cutoff = now - self.window;

        // Remove restart timestamps outside the window
        self.restart_times.retain(|&t| t > cutoff);

        if self.restart_times.len() >= self.max_restarts {
            return Err(self.restart_times.len());
        }

        self.restart_times.push(now);
        Ok(self.restart_times.len())
    }

    /// Get the current number of restarts in the window.
    pub fn current_restarts(&self) -> usize {
        let now = Instant::now();
        let cutoff = now - self.window;
        self.restart_times.iter().filter(|&&t| t > cutoff).count()
    }

    /// Reset the restart tracker.
    pub fn reset(&mut self) {
        self.restart_times.clear();
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_restart_policy_defaults() {
        let policy = RestartPolicy::default();
        assert_eq!(policy.strategy, RestartStrategy::OneForOne);
        assert_eq!(policy.max_restarts, 5);
        assert_eq!(policy.time_window_secs, 60);
    }

    #[test]
    fn test_exponential_backoff() {
        let policy = RestartPolicy::default();

        // Attempt 0: 100ms
        assert_eq!(policy.backoff_delay(0), Duration::from_millis(100));

        // Attempt 1: 200ms
        assert_eq!(policy.backoff_delay(1), Duration::from_millis(200));

        // Attempt 2: 400ms
        assert_eq!(policy.backoff_delay(2), Duration::from_millis(400));

        // Attempt 3: 800ms
        assert_eq!(policy.backoff_delay(3), Duration::from_millis(800));

        // Attempt 4+: capped at 800ms
        assert_eq!(policy.backoff_delay(4), Duration::from_millis(800));
        assert_eq!(policy.backoff_delay(10), Duration::from_millis(800));
    }

    #[test]
    fn test_restart_tracker_records_within_limit() {
        let policy = RestartPolicy::default();
        let mut tracker = RestartTracker::new(&policy);

        // Should succeed for first 5 attempts
        for i in 1..=5 {
            let attempt = tracker.record_restart().unwrap();
            assert_eq!(attempt, i);
        }

        assert_eq!(tracker.current_restarts(), 5);
    }

    #[test]
    fn test_restart_tracker_exceeds_limit() {
        let policy = RestartPolicy::default();
        let mut tracker = RestartTracker::new(&policy);

        // Fill the limit
        for _ in 0..5 {
            let _ = tracker.record_restart();
        }

        // Next attempt should fail
        let result = tracker.record_restart();
        assert!(result.is_err());
        assert_eq!(result.unwrap_err(), 5);
    }

    #[test]
    fn test_restart_tracker_resets_over_time() {
        let mut policy = RestartPolicy::default();
        policy.time_window_secs = 1; // 1 second window for testing
        let mut tracker = RestartTracker::new(&policy);

        // Record a restart
        tracker.record_restart().unwrap();
        assert_eq!(tracker.current_restarts(), 1);

        // Wait for window to expire
        std::thread::sleep(Duration::from_millis(1100));

        // Verify old restart is forgotten
        assert_eq!(tracker.current_restarts(), 0);

        // New restart should succeed
        tracker.record_restart().unwrap();
        assert_eq!(tracker.current_restarts(), 1);
    }
}
