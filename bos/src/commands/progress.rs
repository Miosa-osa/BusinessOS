//! Progress reporting for long-running operations.

use std::sync::atomic::{AtomicU32, Ordering};
use std::sync::Arc;
use std::time::{Duration, Instant};

/// Progress reporter for tracking command execution progress.
pub struct ProgressReporter {
    start_time: Instant,
    total_items: u32,
    completed_items: Arc<AtomicU32>,
    current_message: Arc<std::sync::Mutex<String>>,
}

impl ProgressReporter {
    /// Create a new progress reporter.
    pub fn new(total_items: u32) -> Self {
        Self {
            start_time: Instant::now(),
            total_items,
            completed_items: Arc::new(AtomicU32::new(0)),
            current_message: Arc::new(std::sync::Mutex::new(String::new())),
        }
    }

    /// Update progress with completed item count.
    pub fn update(&self, completed: u32) {
        self.completed_items.store(completed, Ordering::SeqCst);
    }

    /// Increment progress by one item.
    pub fn increment(&self) {
        let current = self.completed_items.load(Ordering::SeqCst);
        self.completed_items.store(current + 1, Ordering::SeqCst);
    }

    /// Set current status message.
    pub fn set_message(&self, message: impl Into<String>) {
        if let Ok(mut msg) = self.current_message.lock() {
            *msg = message.into();
        }
    }

    /// Get current progress percentage.
    pub fn percentage(&self) -> u32 {
        if self.total_items == 0 {
            return 0;
        }
        let completed = self.completed_items.load(Ordering::SeqCst);
        ((completed as f64 / self.total_items as f64) * 100.0) as u32
    }

    /// Get elapsed time.
    pub fn elapsed(&self) -> Duration {
        self.start_time.elapsed()
    }

    /// Get estimated time remaining.
    pub fn eta(&self) -> Option<Duration> {
        let completed = self.completed_items.load(Ordering::SeqCst);
        if completed == 0 {
            return None;
        }

        let elapsed = self.elapsed();
        let rate = completed as f64 / elapsed.as_secs_f64();
        let remaining = self.total_items as f64 - completed as f64;
        let eta_secs = remaining / rate;

        Some(Duration::from_secs_f64(eta_secs))
    }

    /// Format progress bar.
    pub fn format_bar(&self, width: usize) -> String {
        let percent = self.percentage();
        let filled = (width as f64 * percent as f64 / 100.0) as usize;
        let empty = width.saturating_sub(filled);

        format!(
            "[{}{}] {}%",
            "=".repeat(filled),
            " ".repeat(empty),
            percent
        )
    }

    /// Format full progress report.
    pub fn format_report(&self) -> String {
        let completed = self.completed_items.load(Ordering::SeqCst);
        let percent = self.percentage();
        let elapsed = self.elapsed();
        let message = self
            .current_message
            .lock()
            .map(|m| m.clone())
            .unwrap_or_default();

        let eta_str = self
            .eta()
            .map(|e| format!("ETA: {}s", e.as_secs()))
            .unwrap_or_else(|| "ETA: calculating".to_string());

        format!(
            "{} [{}/{}] {}% | {:.1}s elapsed | {}",
            self.format_bar(30),
            completed,
            self.total_items,
            percent,
            elapsed.as_secs_f64(),
            eta_str
        )
    }

    /// Get progress as JSON.
    pub fn to_json(&self) -> serde_json::Value {
        use serde_json::json;

        let completed = self.completed_items.load(Ordering::SeqCst);
        let elapsed = self.elapsed();
        let message = self
            .current_message
            .lock()
            .map(|m| m.clone())
            .unwrap_or_default();

        json!({
            "completed": completed,
            "total": self.total_items,
            "percentage": self.percentage(),
            "elapsed_ms": elapsed.as_millis(),
            "eta_ms": self.eta().map(|e| e.as_millis()),
            "message": message,
        })
    }
}

/// Multi-stage progress tracker.
pub struct StageProgressTracker {
    stages: Vec<Stage>,
    current_stage: usize,
}

struct Stage {
    name: String,
    total_items: u32,
    completed_items: u32,
    reporter: ProgressReporter,
}

impl StageProgressTracker {
    /// Create a new stage tracker.
    pub fn new() -> Self {
        Self {
            stages: vec![],
            current_stage: 0,
        }
    }

    /// Add a new stage.
    pub fn add_stage(&mut self, name: impl Into<String>, total_items: u32) {
        self.stages.push(Stage {
            name: name.into(),
            total_items,
            completed_items: 0,
            reporter: ProgressReporter::new(total_items),
        });
    }

    /// Move to next stage.
    pub fn next_stage(&mut self) -> bool {
        if self.current_stage + 1 < self.stages.len() {
            self.current_stage += 1;
            true
        } else {
            false
        }
    }

    /// Update current stage progress.
    pub fn update_current(&mut self, completed: u32) {
        if let Some(stage) = self.stages.get_mut(self.current_stage) {
            stage.completed_items = completed;
            stage.reporter.update(completed);
        }
    }

    /// Get overall progress percentage.
    pub fn overall_percentage(&self) -> u32 {
        if self.stages.is_empty() {
            return 0;
        }

        let total_items: u32 = self.stages.iter().map(|s| s.total_items).sum();
        if total_items == 0 {
            return 0;
        }

        let completed: u32 = self.stages.iter().map(|s| s.completed_items).sum();
        ((completed as f64 / total_items as f64) * 100.0) as u32
    }

    /// Format all stages report.
    pub fn format_report(&self) -> String {
        let mut output = String::new();
        output.push_str("=== Progress Report ===\n");
        output.push_str(&format!(
            "Overall: {}%\n\n",
            self.overall_percentage()
        ));

        for (idx, stage) in self.stages.iter().enumerate() {
            let marker = if idx == self.current_stage { ">" } else { " " };
            output.push_str(&format!(
                "{} [{}] {}/{} ({}%)\n",
                marker,
                stage.name,
                stage.completed_items,
                stage.total_items,
                stage.reporter.percentage()
            ));
        }

        output
    }
}

impl Default for StageProgressTracker {
    fn default() -> Self {
        Self::new()
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::thread;

    #[test]
    fn test_progress_percentage() {
        let reporter = ProgressReporter::new(100);
        reporter.update(50);
        assert_eq!(reporter.percentage(), 50);
    }

    #[test]
    fn test_progress_bar() {
        let reporter = ProgressReporter::new(100);
        reporter.update(50);
        let bar = reporter.format_bar(20);
        assert!(bar.contains("50%"));
    }

    #[test]
    fn test_eta_calculation() {
        let reporter = ProgressReporter::new(100);
        reporter.update(10);
        thread::sleep(Duration::from_millis(100));
        // ETA should be available
        assert!(reporter.eta().is_some());
    }

    #[test]
    fn test_stage_tracker() {
        let mut tracker = StageProgressTracker::new();
        tracker.add_stage("Phase 1", 10);
        tracker.add_stage("Phase 2", 20);
        tracker.update_current(5);
        assert_eq!(tracker.overall_percentage(), 11); // 5/30
    }
}
