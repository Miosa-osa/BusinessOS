//! Types for distributed process mining

use serde::{Deserialize, Serialize};
use std::collections::HashSet;

/// Raft log term (election epoch)
#[derive(Debug, Clone, Copy, PartialEq, Eq, PartialOrd, Ord, Serialize, Deserialize)]
pub struct Term(pub u64);

impl Term {
    pub fn new() -> Self {
        Term(0)
    }

    pub fn increment(&mut self) {
        self.0 += 1;
    }
}

impl Default for Term {
    fn default() -> Self {
        Term::new()
    }
}

/// Node state in Raft consensus
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum NodeState {
    Follower,
    Candidate,
    Leader,
}

/// Vote in Raft election
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Vote {
    pub term: Term,
    pub candidate_id: String,
    pub last_log_index: u64,
    pub last_log_term: Term,
    pub voter_id: String,
    pub granted: bool,
}

/// Heartbeat message from leader
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Heartbeat {
    pub term: Term,
    pub leader_id: String,
    pub prev_log_index: u64,
    pub prev_log_term: Term,
    pub leader_commit: u64,
    pub entries: Vec<LogEntry>,
}

/// Log entry in Raft log
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct LogEntry {
    pub term: Term,
    pub index: u64,
    pub command: LogCommand,
}

/// Commands that can be logged
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum LogCommand {
    DiscoverPartition {
        partition_id: String,
        case_ids: Vec<String>,
    },
    MergeModels {
        partition_id: String,
        model_hash: String,
    },
    ElectLeader {
        leader_id: String,
        term: u64,
    },
}

/// Quorum acknowledgment for Byzantine tolerance
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct QuorumAck {
    pub entry_index: u64,
    pub acknowledged_by: HashSet<String>,
    pub required_count: usize,
}

impl QuorumAck {
    pub fn new(entry_index: u64, required_count: usize) -> Self {
        Self {
            entry_index,
            acknowledged_by: HashSet::new(),
            required_count,
        }
    }

    pub fn add_ack(&mut self, node_id: String) {
        self.acknowledged_by.insert(node_id);
    }

    pub fn is_quorum_reached(&self) -> bool {
        self.acknowledged_by.len() >= self.required_count
    }
}

/// Process model for merging
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ProcessModel {
    pub partition_id: String,
    pub places: Vec<String>,
    pub transitions: Vec<String>,
    pub arcs: Vec<(String, String)>,
    pub hash: String,
}

impl ProcessModel {
    pub fn new(partition_id: String) -> Self {
        Self {
            partition_id,
            places: Vec::new(),
            transitions: Vec::new(),
            arcs: Vec::new(),
            hash: String::new(),
        }
    }

    pub fn add_place(&mut self, place: String) {
        if !self.places.contains(&place) {
            self.places.push(place);
        }
    }

    pub fn add_transition(&mut self, transition: String) {
        if !self.transitions.contains(&transition) {
            self.transitions.push(transition);
        }
    }

    pub fn add_arc(&mut self, from: String, to: String) {
        if !self.arcs.contains(&(from.clone(), to.clone())) {
            self.arcs.push((from, to));
        }
    }

    pub fn compute_hash(&mut self) {
        use sha2::{Sha256, Digest};
        let mut hasher = Sha256::new();
        hasher.update(format!("{:?}{:?}{:?}", self.places, self.transitions, self.arcs));
        self.hash = format!("{:x}", hasher.finalize());
    }
}

/// Node health status
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct NodeHealth {
    pub node_id: String,
    pub is_alive: bool,
    pub last_heartbeat: u64,
    pub failed_checks: u32,
}

impl NodeHealth {
    pub fn new(node_id: String) -> Self {
        Self {
            node_id,
            is_alive: true,
            last_heartbeat: 0,
            failed_checks: 0,
        }
    }

    pub fn mark_healthy(&mut self) {
        self.is_alive = true;
        self.failed_checks = 0;
    }

    pub fn mark_failed_check(&mut self) {
        self.failed_checks += 1;
        if self.failed_checks > 3 {
            self.is_alive = false;
        }
    }
}

/// Distribution strategy for event log partitions
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum PartitionStrategy {
    ByCase,
    ByTime,
    ByActivity,
    ByRoundRobin,
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_term_increment() {
        let mut term = Term::new();
        assert_eq!(term.0, 0);
        term.increment();
        assert_eq!(term.0, 1);
    }

    #[test]
    fn test_quorum_ack() {
        let mut ack = QuorumAck::new(1, 3);
        assert!(!ack.is_quorum_reached());

        ack.add_ack("node1".to_string());
        ack.add_ack("node2".to_string());
        assert!(!ack.is_quorum_reached());

        ack.add_ack("node3".to_string());
        assert!(ack.is_quorum_reached());
    }

    #[test]
    fn test_process_model_creation() {
        let mut model = ProcessModel::new("partition_a".to_string());
        model.add_place("p1".to_string());
        model.add_transition("t1".to_string());
        model.add_arc("p1".to_string(), "t1".to_string());

        assert_eq!(model.places.len(), 1);
        assert_eq!(model.transitions.len(), 1);
        assert_eq!(model.arcs.len(), 1);
    }

    #[test]
    fn test_node_health_failure_count() {
        let mut health = NodeHealth::new("node1".to_string());
        assert!(health.is_alive);

        health.mark_failed_check();
        health.mark_failed_check();
        health.mark_failed_check();
        assert!(health.is_alive);

        health.mark_failed_check();
        assert!(!health.is_alive);
    }

    #[test]
    fn test_duplicate_places() {
        let mut model = ProcessModel::new("partition_a".to_string());
        model.add_place("p1".to_string());
        model.add_place("p1".to_string());
        assert_eq!(model.places.len(), 1);
    }
}
