//! Distributed process mining with Raft consensus
//!
//! Provides multi-node process discovery with Byzantine-fault-tolerant consensus.
//! Coordinator node leads quorum-based decisions, worker nodes discover locally,
//! and models merge at the coordinator while preserving soundness.

pub mod coordinator;
pub mod worker;
pub mod consensus;
pub mod merger;
pub mod recovery;
pub mod types;

pub use coordinator::{RaftCoordinator, PeerTransport, HttpPeerTransport};
pub use worker::DistributedWorker;
pub use consensus::ConsensusProtocol;
pub use merger::ModelMerger;
pub use recovery::FaultRecovery;
pub use types::*;

// Re-export recovery types
pub use recovery::{RecoveryEvent, RecoveryEventType};
pub use merger::MergeStats;

use anyhow::Result;
use std::sync::Arc;
use tokio::sync::RwLock;
use std::collections::HashMap;

/// Distributed process mining system coordinator
pub struct DistributedPM {
    coordinator: Arc<RwLock<RaftCoordinator>>,
    workers: Arc<RwLock<HashMap<String, DistributedWorker>>>,
    consensus: Arc<ConsensusProtocol>,
    merger: Arc<ModelMerger>,
    recovery: Arc<FaultRecovery>,
}

impl DistributedPM {
    /// Create a new distributed PM system with given node ID and peer list
    pub fn new(
        node_id: String,
        peers: Vec<String>,
        _quorum_size: usize,
    ) -> Self {
        let consensus = Arc::new(ConsensusProtocol::new(node_id.clone(), peers.len()));
        let recovery = Arc::new(FaultRecovery::new(node_id.clone()));

        Self {
            coordinator: Arc::new(RwLock::new(RaftCoordinator::new(node_id.clone(), peers))),
            workers: Arc::new(RwLock::new(HashMap::new())),
            consensus,
            merger: Arc::new(ModelMerger::new()),
            recovery,
        }
    }

    /// Start Raft leader election
    pub async fn start_election(&self) -> Result<()> {
        let mut coordinator = self.coordinator.write().await;
        coordinator.start_election(self.consensus.clone()).await
    }

    /// Register a worker node for a log partition
    pub async fn register_worker(&self, partition_id: String, worker: DistributedWorker) -> Result<()> {
        let mut workers = self.workers.write().await;
        workers.insert(partition_id, worker);
        Ok(())
    }

    /// Submit vote for consensus
    pub async fn submit_vote(&self, vote: Vote) -> Result<()> {
        self.consensus.submit_vote(vote).await
    }

    /// Get current leader
    pub async fn get_leader(&self) -> Option<String> {
        self.coordinator.read().await.get_leader()
    }

    /// Check if this node is leader
    pub async fn is_leader(&self) -> bool {
        self.coordinator.read().await.is_leader()
    }

    /// Send heartbeat to all followers
    pub async fn send_heartbeat(&self) -> Result<()> {
        let coordinator = self.coordinator.read().await;
        coordinator.send_heartbeat().await
    }
}

impl Default for DistributedPM {
    fn default() -> Self {
        Self::new("node1".to_string(), vec![], 1)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn test_distributed_pm_creation() {
        let pm = DistributedPM::new(
            "node1".to_string(),
            vec!["node2".to_string(), "node3".to_string()],
            3,
        );
        assert!(!pm.is_leader().await);
    }

    #[tokio::test]
    async fn test_register_worker() {
        let pm = DistributedPM::new("node1".to_string(), vec![], 1);
        let worker = DistributedWorker::new("node1".to_string(), "partition_a".to_string());
        pm.register_worker("partition_a".to_string(), worker)
            .await
            .unwrap();

        let workers = pm.workers.read().await;
        assert!(workers.contains_key("partition_a"));
    }
}
