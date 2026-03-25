//! Worker node for distributed process discovery

use super::types::ProcessModel;
use anyhow::Result;
use serde::{Deserialize, Serialize};
use std::collections::{HashMap, HashSet};

/// Worker node managing a partition of event log
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DistributedWorker {
    pub node_id: String,
    pub partition_id: String,
    pub case_ids: HashSet<String>,
    pub activity_counts: HashMap<String, u64>,
    pub local_model: ProcessModel,
    pub case_variants: HashMap<String, Vec<String>>,
    pub processed_events: u64,
}

impl DistributedWorker {
    /// Create a new worker for a partition
    pub fn new(node_id: String, partition_id: String) -> Self {
        Self {
            node_id,
            partition_id: partition_id.clone(),
            case_ids: HashSet::new(),
            activity_counts: HashMap::new(),
            local_model: ProcessModel::new(partition_id),
            case_variants: HashMap::new(),
            processed_events: 0,
        }
    }

    /// Add event to partition discovery
    pub fn process_event(
        &mut self,
        case_id: String,
        activity: String,
        _timestamp: u64,
    ) -> Result<()> {
        // Register case
        self.case_ids.insert(case_id.clone());

        // Count activity
        *self.activity_counts.entry(activity.clone()).or_insert(0) += 1;

        // Get or create case variant sequence
        self.case_variants
            .entry(case_id.clone())
            .or_insert_with(Vec::new)
            .push(activity.clone());

        // Add place for activity
        self.local_model.add_place(format!("p_{}", activity));
        self.local_model.add_transition(format!("t_{}", activity));

        self.processed_events += 1;
        Ok(())
    }

    /// Discover local process model from partition
    pub fn discover_local_model(&mut self) -> Result<ProcessModel> {
        // Build variant-based Petri net
        for (_case_id, activities) in &self.case_variants {
            for i in 0..activities.len() {
                let curr_activity = &activities[i];
                self.local_model
                    .add_transition(format!("t_{}", curr_activity));

                // Add arcs from previous activity
                if i > 0 {
                    let prev_activity = &activities[i - 1];
                    let from = format!("p_{}", prev_activity);
                    let to = format!("t_{}", curr_activity);
                    self.local_model.add_arc(from, to);
                }
            }
        }

        self.local_model.compute_hash();
        Ok(self.local_model.clone())
    }

    /// Get number of cases in this partition
    pub fn case_count(&self) -> usize {
        self.case_ids.len()
    }

    /// Get number of unique activities
    pub fn activity_count(&self) -> usize {
        self.activity_counts.len()
    }

    /// Get top N activities by frequency
    pub fn top_activities(&self, n: usize) -> Vec<(String, u64)> {
        let mut activities: Vec<_> = self
            .activity_counts
            .iter()
            .map(|(k, v)| (k.clone(), *v))
            .collect();
        activities.sort_by(|a, b| b.1.cmp(&a.1));
        activities.into_iter().take(n).collect()
    }

    /// Get variant fingerprint (activities as string)
    pub fn get_variant_fingerprint(&self, case_id: &str) -> Option<String> {
        self.case_variants
            .get(case_id)
            .map(|activities| activities.join("->"))
    }

    /// Check if partition is empty
    pub fn is_empty(&self) -> bool {
        self.case_ids.is_empty()
    }

    /// Get events processed
    pub fn events_processed(&self) -> u64 {
        self.processed_events
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_worker_creation() {
        let worker = DistributedWorker::new("node1".to_string(), "partition_a".to_string());

        assert_eq!(worker.node_id, "node1");
        assert_eq!(worker.partition_id, "partition_a");
        assert!(worker.case_ids.is_empty());
        assert!(worker.is_empty());
    }

    #[test]
    fn test_process_event() {
        let mut worker = DistributedWorker::new("node1".to_string(), "partition_a".to_string());

        worker
            .process_event("case1".to_string(), "invoice".to_string(), 1000)
            .unwrap();
        worker
            .process_event("case1".to_string(), "approval".to_string(), 2000)
            .unwrap();

        assert_eq!(worker.case_count(), 1);
        assert_eq!(worker.activity_count(), 2);
        assert_eq!(worker.events_processed(), 2);
    }

    #[test]
    fn test_multiple_cases() {
        let mut worker = DistributedWorker::new("node1".to_string(), "partition_a".to_string());

        worker
            .process_event("case1".to_string(), "start".to_string(), 1000)
            .unwrap();
        worker
            .process_event("case2".to_string(), "start".to_string(), 1100)
            .unwrap();
        worker
            .process_event("case1".to_string(), "end".to_string(), 2000)
            .unwrap();

        assert_eq!(worker.case_count(), 2);
        assert_eq!(worker.activity_count(), 2);
    }

    #[test]
    fn test_activity_frequency() {
        let mut worker = DistributedWorker::new("node1".to_string(), "partition_a".to_string());

        worker
            .process_event("case1".to_string(), "approve".to_string(), 1000)
            .unwrap();
        worker
            .process_event("case1".to_string(), "approve".to_string(), 2000)
            .unwrap();
        worker
            .process_event("case1".to_string(), "pay".to_string(), 3000)
            .unwrap();

        let counts = worker.activity_counts;
        assert_eq!(counts.get("approve"), Some(&2));
        assert_eq!(counts.get("pay"), Some(&1));
    }

    #[test]
    fn test_top_activities() {
        let mut worker = DistributedWorker::new("node1".to_string(), "partition_a".to_string());

        for i in 0..5 {
            worker
                .process_event("case1".to_string(), "activity_a".to_string(), 1000 + i)
                .unwrap();
        }
        for i in 0..3 {
            worker
                .process_event("case1".to_string(), "activity_b".to_string(), 2000 + i)
                .unwrap();
        }

        let top = worker.top_activities(2);
        assert_eq!(top.len(), 2);
        assert_eq!(top[0].0, "activity_a");
        assert_eq!(top[0].1, 5);
        assert_eq!(top[1].0, "activity_b");
        assert_eq!(top[1].1, 3);
    }

    #[test]
    fn test_discover_local_model() {
        let mut worker = DistributedWorker::new("node1".to_string(), "partition_a".to_string());

        worker
            .process_event("case1".to_string(), "start".to_string(), 1000)
            .unwrap();
        worker
            .process_event("case1".to_string(), "process".to_string(), 2000)
            .unwrap();
        worker
            .process_event("case1".to_string(), "end".to_string(), 3000)
            .unwrap();

        let model = worker.discover_local_model().unwrap();
        assert!(!model.places.is_empty());
        assert!(!model.transitions.is_empty());
        assert!(!model.arcs.is_empty());
        assert!(!model.hash.is_empty());
    }

    #[test]
    fn test_variant_fingerprint() {
        let mut worker = DistributedWorker::new("node1".to_string(), "partition_a".to_string());

        worker
            .process_event("case1".to_string(), "a".to_string(), 1000)
            .unwrap();
        worker
            .process_event("case1".to_string(), "b".to_string(), 2000)
            .unwrap();
        worker
            .process_event("case1".to_string(), "c".to_string(), 3000)
            .unwrap();

        let fp = worker.get_variant_fingerprint("case1");
        assert_eq!(fp, Some("a->b->c".to_string()));
    }

    #[test]
    fn test_nonexistent_variant() {
        let worker = DistributedWorker::new("node1".to_string(), "partition_a".to_string());
        assert!(worker.get_variant_fingerprint("nonexistent").is_none());
    }
}
