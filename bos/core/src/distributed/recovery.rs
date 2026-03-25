//! Fault recovery for distributed process mining

use super::types::NodeHealth;
use anyhow::Result;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::time::{SystemTime, UNIX_EPOCH};

/// Manages fault detection and recovery
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct FaultRecovery {
    node_id: String,
    health_map: HashMap<String, NodeHealth>,
    crash_count: u32,
    recovery_log: Vec<RecoveryEvent>,
}

/// Recovery event for audit trail
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RecoveryEvent {
    pub timestamp: u64,
    pub event_type: RecoveryEventType,
    pub node_id: String,
    pub details: String,
}

/// Types of recovery events
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum RecoveryEventType {
    NodeCrash,
    NodeRecovered,
    LeaderElection,
    PartitionDetected,
    ReplicationStarted,
}

impl FaultRecovery {
    /// Create a new fault recovery manager
    pub fn new(node_id: String) -> Self {
        Self {
            node_id,
            health_map: HashMap::new(),
            crash_count: 0,
            recovery_log: Vec::new(),
        }
    }

    /// Register a node for health monitoring
    pub fn register_node(&mut self, node_id: String) {
        self.health_map
            .insert(node_id.clone(), NodeHealth::new(node_id));
    }

    /// Check health of a node
    pub fn health_check(&mut self, node_id: &str, is_responsive: bool) -> Result<bool> {
        let should_log_crash = if let Some(health) = self.health_map.get_mut(node_id) {
            if is_responsive {
                health.mark_healthy();
                false
            } else {
                health.mark_failed_check();
                if !health.is_alive {
                    self.crash_count += 1;
                    true
                } else {
                    false
                }
            }
        } else {
            return Ok(true); // Unknown node treated as alive
        };

        if should_log_crash {
            self.log_event(
                RecoveryEventType::NodeCrash,
                node_id.to_string(),
                "Node declared dead".to_string(),
            );
        }

        Ok(self.health_map.get(node_id).map(|h| h.is_alive).unwrap_or(true))
    }

    /// Get node health
    pub fn get_node_health(&self, node_id: &str) -> Option<NodeHealth> {
        self.health_map.get(node_id).cloned()
    }

    /// Check if node is healthy
    pub fn is_node_healthy(&self, node_id: &str) -> bool {
        self.health_map
            .get(node_id)
            .map(|h| h.is_alive)
            .unwrap_or(false)
    }

    /// Start node recovery
    pub fn start_recovery(&mut self, node_id: String) -> Result<()> {
        if let Some(health) = self.health_map.get_mut(&node_id) {
            health.mark_healthy();
            self.log_event(
                RecoveryEventType::NodeRecovered,
                node_id.clone(),
                "Recovery initiated".to_string(),
            );
        }
        Ok(())
    }

    /// Log recovery event
    fn log_event(&mut self, event_type: RecoveryEventType, node_id: String, details: String) {
        let timestamp = SystemTime::now()
            .duration_since(UNIX_EPOCH)
            .unwrap_or_default()
            .as_secs();

        self.recovery_log.push(RecoveryEvent {
            timestamp,
            event_type,
            node_id,
            details,
        });
    }

    /// Get all health statuses
    pub fn get_all_health(&self) -> Vec<NodeHealth> {
        self.health_map.values().cloned().collect()
    }

    /// Count alive nodes
    pub fn count_alive_nodes(&self) -> usize {
        self.health_map.values().filter(|h| h.is_alive).count()
    }

    /// Trigger leader re-election
    pub fn trigger_reelection(&mut self, node_id: String) -> Result<()> {
        self.log_event(
            RecoveryEventType::LeaderElection,
            node_id,
            "Leader re-election triggered".to_string(),
        );
        Ok(())
    }

    /// Detect network partition
    pub fn detect_partition(&mut self, partition_id: String, isolated_nodes: Vec<String>) -> Result<()> {
        self.log_event(
            RecoveryEventType::PartitionDetected,
            partition_id,
            format!("Isolated nodes: {:?}", isolated_nodes),
        );
        Ok(())
    }

    /// Get crash count
    pub fn crash_count(&self) -> u32 {
        self.crash_count
    }

    /// Get recovery log
    pub fn get_recovery_log(&self) -> Vec<RecoveryEvent> {
        self.recovery_log.clone()
    }

    /// Can system tolerate another failure (Byzantine tolerance)
    pub fn can_tolerate_failure(&self, total_nodes: usize) -> bool {
        let alive_count = self.count_alive_nodes();
        let quorum_needed = (total_nodes / 2) + 1;
        alive_count > quorum_needed
    }

    /// Get failed node count
    pub fn failed_node_count(&self) -> usize {
        self.health_map.values().filter(|h| !h.is_alive).count()
    }

    /// Get last N recovery events
    pub fn get_recent_events(&self, count: usize) -> Vec<RecoveryEvent> {
        self.recovery_log
            .iter()
            .rev()
            .take(count)
            .cloned()
            .collect()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_fault_recovery_creation() {
        let recovery = FaultRecovery::new("node1".to_string());
        assert_eq!(recovery.crash_count, 0);
        assert!(recovery.health_map.is_empty());
    }

    #[test]
    fn test_register_node() {
        let mut recovery = FaultRecovery::new("node1".to_string());
        recovery.register_node("node2".to_string());

        assert!(recovery.health_map.contains_key("node2"));
    }

    #[test]
    fn test_health_check_responsive() {
        let mut recovery = FaultRecovery::new("node1".to_string());
        recovery.register_node("node2".to_string());

        let alive = recovery.health_check("node2", true).unwrap();
        assert!(alive);
    }

    #[test]
    fn test_health_check_unresponsive() {
        let mut recovery = FaultRecovery::new("node1".to_string());
        recovery.register_node("node2".to_string());

        // First check: mark as failed
        recovery.health_check("node2", false).unwrap();
        recovery.health_check("node2", false).unwrap();
        recovery.health_check("node2", false).unwrap();
        assert!(recovery.is_node_healthy("node2"));

        // Fourth check triggers death
        recovery.health_check("node2", false).unwrap();
        assert!(!recovery.is_node_healthy("node2"));
    }

    #[test]
    fn test_crash_count() {
        let mut recovery = FaultRecovery::new("node1".to_string());
        recovery.register_node("node2".to_string());

        for _ in 0..4 {
            recovery.health_check("node2", false).unwrap();
        }

        assert_eq!(recovery.crash_count(), 1);
    }

    #[test]
    fn test_recovery_log() {
        let mut recovery = FaultRecovery::new("node1".to_string());
        recovery.register_node("node2".to_string());

        recovery.health_check("node2", false).unwrap();
        recovery.health_check("node2", false).unwrap();
        recovery.health_check("node2", false).unwrap();
        recovery.health_check("node2", false).unwrap();

        let log = recovery.get_recovery_log();
        assert!(!log.is_empty());
    }

    #[test]
    fn test_start_recovery() {
        let mut recovery = FaultRecovery::new("node1".to_string());
        recovery.register_node("node2".to_string());

        // Mark node as dead
        for _ in 0..4 {
            recovery.health_check("node2", false).unwrap();
        }
        assert!(!recovery.is_node_healthy("node2"));

        // Recover it
        recovery.start_recovery("node2".to_string()).unwrap();
        assert!(recovery.is_node_healthy("node2"));
    }

    #[test]
    fn test_count_alive_nodes() {
        let mut recovery = FaultRecovery::new("node1".to_string());
        recovery.register_node("node2".to_string());
        recovery.register_node("node3".to_string());

        assert_eq!(recovery.count_alive_nodes(), 2);

        for _ in 0..4 {
            recovery.health_check("node2", false).unwrap();
        }

        assert_eq!(recovery.count_alive_nodes(), 1);
    }

    #[test]
    fn test_can_tolerate_failure_3_nodes() {
        let mut recovery = FaultRecovery::new("node1".to_string());
        recovery.register_node("node2".to_string());
        recovery.register_node("node3".to_string());

        // 3 nodes total: node1 (implicit) + node2 + node3 = 3, but we only registered 2 explicitly
        // Active: all 2 registered + implicitly node1 = 3 total
        // Quorum needed: (3 / 2) + 1 = 2
        // Can tolerate: alive_count > quorum_needed → need more than 2 alive
        // Currently 2 alive, so cannot tolerate another failure
        assert_eq!(recovery.count_alive_nodes(), 2);
        assert!(!recovery.can_tolerate_failure(3));
    }

    #[test]
    fn test_can_tolerate_failure_5_nodes() {
        let mut recovery = FaultRecovery::new("node1".to_string());
        for i in 2..=5 {
            recovery.register_node(format!("node{}", i));
        }

        // 5 nodes: quorum 3, can tolerate 1 failure (4 alive > 3)
        assert_eq!(recovery.count_alive_nodes(), 4);
        assert!(recovery.can_tolerate_failure(5));
    }

    #[test]
    fn test_trigger_reelection() {
        let mut recovery = FaultRecovery::new("node1".to_string());
        recovery.trigger_reelection("node1".to_string()).unwrap();

        let log = recovery.get_recovery_log();
        assert_eq!(log.len(), 1);
    }

    #[test]
    fn test_detect_partition() {
        let mut recovery = FaultRecovery::new("node1".to_string());
        recovery
            .detect_partition("partition_a".to_string(), vec!["node2".to_string()])
            .unwrap();

        let log = recovery.get_recovery_log();
        assert_eq!(log.len(), 1);
    }

    #[test]
    fn test_recent_events() {
        let mut recovery = FaultRecovery::new("node1".to_string());

        recovery.trigger_reelection("node1".to_string()).unwrap();
        recovery.trigger_reelection("node1".to_string()).unwrap();
        recovery.trigger_reelection("node1".to_string()).unwrap();

        let recent = recovery.get_recent_events(2);
        assert_eq!(recent.len(), 2);
    }

    #[test]
    fn test_failed_node_count() {
        let mut recovery = FaultRecovery::new("node1".to_string());
        recovery.register_node("node2".to_string());
        recovery.register_node("node3".to_string());

        for _ in 0..4 {
            recovery.health_check("node2", false).unwrap();
        }

        assert_eq!(recovery.failed_node_count(), 1);
    }

    #[test]
    fn test_get_all_health() {
        let mut recovery = FaultRecovery::new("node1".to_string());
        recovery.register_node("node2".to_string());
        recovery.register_node("node3".to_string());

        let health_statuses = recovery.get_all_health();
        assert_eq!(health_statuses.len(), 2);
    }
}
