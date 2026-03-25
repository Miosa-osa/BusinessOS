//! Raft consensus protocol implementation

use super::types::{Term, Vote, QuorumAck};
use anyhow::Result;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::sync::atomic::{AtomicUsize, Ordering};
use std::sync::Arc;
use tokio::sync::RwLock;

/// Consensus protocol manager (non-serializable due to runtime state)
#[derive(Debug)]
pub struct ConsensusProtocol {
    node_id: String,
    total_nodes: usize,
    quorum_size: usize,
    votes: Arc<RwLock<HashMap<u64, Vec<Vote>>>>,
    quorum_acks: Arc<RwLock<HashMap<u64, QuorumAck>>>,
    election_count: Arc<AtomicUsize>,
}

impl ConsensusProtocol {
    /// Create a new consensus protocol instance
    pub fn new(node_id: String, total_nodes: usize) -> Self {
        let quorum_size = (total_nodes / 2) + 1;

        Self {
            node_id,
            total_nodes,
            quorum_size,
            votes: Arc::new(RwLock::new(HashMap::new())),
            quorum_acks: Arc::new(RwLock::new(HashMap::new())),
            election_count: Arc::new(AtomicUsize::new(0)),
        }
    }

    /// Get quorum size for this cluster
    pub fn quorum_size(&self) -> usize {
        self.quorum_size
    }

    /// Get total nodes in cluster
    pub fn total_nodes(&self) -> usize {
        self.total_nodes
    }

    /// Submit vote for consensus
    pub async fn submit_vote(&self, vote: Vote) -> Result<()> {
        let term_num = vote.term.0;

        let mut votes = self.votes.write().await;
        votes
            .entry(term_num)
            .or_insert_with(Vec::new)
            .push(vote);

        Ok(())
    }

    /// Check if a term has reached quorum
    pub async fn has_quorum(&self, term: Term) -> bool {
        let votes = self.votes.read().await;
        if let Some(term_votes) = votes.get(&term.0) {
            let granted = term_votes.iter().filter(|v| v.granted).count();
            granted >= self.quorum_size
        } else {
            false
        }
    }

    /// Get votes for a term
    pub async fn get_votes(&self, term: Term) -> Vec<Vote> {
        let votes = self.votes.read().await;
        votes
            .get(&term.0)
            .map(|v| v.clone())
            .unwrap_or_default()
    }

    /// Count granted votes for a term
    pub async fn count_granted_votes(&self, term: Term) -> usize {
        let votes = self.votes.read().await;
        votes
            .get(&term.0)
            .map(|term_votes| term_votes.iter().filter(|v| v.granted).count())
            .unwrap_or(0)
    }

    /// Register quorum ack for log entry
    pub async fn register_quorum_ack(&self, entry_index: u64) -> QuorumAck {
        let mut acks = self.quorum_acks.write().await;
        let ack = QuorumAck::new(entry_index, self.quorum_size);
        acks.insert(entry_index, ack.clone());
        ack
    }

    /// Acknowledge a log entry from a node
    pub async fn ack_log_entry(&self, entry_index: u64, node_id: String) -> Result<bool> {
        let mut acks = self.quorum_acks.write().await;

        if let Some(ack) = acks.get_mut(&entry_index) {
            ack.add_ack(node_id);
            Ok(ack.is_quorum_reached())
        } else {
            Ok(false)
        }
    }

    /// Check if quorum has been reached for entry
    pub async fn is_entry_committed(&self, entry_index: u64) -> bool {
        let acks = self.quorum_acks.read().await;
        acks.get(&entry_index)
            .map(|ack| ack.is_quorum_reached())
            .unwrap_or(false)
    }

    /// Get quorum ack state
    pub async fn get_quorum_ack(&self, entry_index: u64) -> Option<QuorumAck> {
        let acks = self.quorum_acks.read().await;
        acks.get(&entry_index).cloned()
    }

    /// Clear votes for a term (cleanup)
    pub async fn clear_votes(&self, term: Term) {
        let mut votes = self.votes.write().await;
        votes.remove(&term.0);
    }

    /// Get election count
    pub fn election_count(&self) -> usize {
        self.election_count.load(Ordering::SeqCst)
    }

    /// Increment election count
    pub fn increment_elections(&self) {
        self.election_count.fetch_add(1, Ordering::SeqCst);
    }

    /// Check Byzantine tolerance (can tolerate up to f failures where 2f+1 <= quorum_size)
    pub fn byzantine_tolerance(&self) -> usize {
        (self.quorum_size - 1) / 2
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn test_consensus_creation() {
        let cp = ConsensusProtocol::new("node1".to_string(), 3);

        assert_eq!(cp.total_nodes(), 3);
        assert_eq!(cp.quorum_size(), 2);
        assert_eq!(cp.byzantine_tolerance(), 0); // 3 nodes: 1f+1 = 2
    }

    #[tokio::test]
    async fn test_submit_vote() {
        let cp = ConsensusProtocol::new("node1".to_string(), 3);

        let vote = Vote {
            term: Term(1),
            candidate_id: "node2".to_string(),
            last_log_index: 0,
            last_log_term: Term(0),
            voter_id: "node1".to_string(),
            granted: true,
        };

        cp.submit_vote(vote).await.unwrap();

        let votes = cp.get_votes(Term(1)).await;
        assert_eq!(votes.len(), 1);
    }

    #[tokio::test]
    async fn test_quorum_size_3_nodes() {
        let cp = ConsensusProtocol::new("node1".to_string(), 3);
        assert_eq!(cp.quorum_size(), 2);
    }

    #[tokio::test]
    async fn test_quorum_size_5_nodes() {
        let cp = ConsensusProtocol::new("node1".to_string(), 5);
        assert_eq!(cp.quorum_size(), 3);
    }

    #[tokio::test]
    async fn test_has_quorum_false() {
        let cp = ConsensusProtocol::new("node1".to_string(), 3);

        let vote = Vote {
            term: Term(1),
            candidate_id: "node2".to_string(),
            last_log_index: 0,
            last_log_term: Term(0),
            voter_id: "node1".to_string(),
            granted: true,
        };

        cp.submit_vote(vote).await.unwrap();

        // Only 1 vote, quorum needs 2
        assert!(!cp.has_quorum(Term(1)).await);
    }

    #[tokio::test]
    async fn test_has_quorum_true() {
        let cp = ConsensusProtocol::new("node1".to_string(), 3);

        for i in 1..=2 {
            let vote = Vote {
                term: Term(1),
                candidate_id: "node2".to_string(),
                last_log_index: 0,
                last_log_term: Term(0),
                voter_id: format!("node{}", i),
                granted: true,
            };
            cp.submit_vote(vote).await.unwrap();
        }

        assert!(cp.has_quorum(Term(1)).await);
    }

    #[tokio::test]
    async fn test_count_granted_votes() {
        let cp = ConsensusProtocol::new("node1".to_string(), 5);

        for i in 0..3 {
            let vote = Vote {
                term: Term(1),
                candidate_id: "node1".to_string(),
                last_log_index: 0,
                last_log_term: Term(0),
                voter_id: format!("node{}", i),
                granted: i != 2, // One vote denied
            };
            cp.submit_vote(vote).await.unwrap();
        }

        assert_eq!(cp.count_granted_votes(Term(1)).await, 2);
    }

    #[tokio::test]
    async fn test_log_entry_quorum_ack() {
        let cp = ConsensusProtocol::new("node1".to_string(), 3);

        let ack = cp.register_quorum_ack(1).await;
        assert!(!ack.is_quorum_reached());

        // Add one acknowledgment
        cp.ack_log_entry(1, "node1".to_string()).await.unwrap();
        let committed = cp.is_entry_committed(1).await;
        assert!(!committed); // Need 2 for quorum

        // Add second acknowledgment
        let reached = cp.ack_log_entry(1, "node2".to_string()).await.unwrap();
        assert!(reached);
        assert!(cp.is_entry_committed(1).await);
    }

    #[tokio::test]
    async fn test_byzantine_tolerance_3_nodes() {
        let cp = ConsensusProtocol::new("node1".to_string(), 3);
        // 3 nodes: quorum 2, byzantine tolerance 0 (can't tolerate failure)
        assert_eq!(cp.byzantine_tolerance(), 0);
    }

    #[tokio::test]
    async fn test_byzantine_tolerance_5_nodes() {
        let cp = ConsensusProtocol::new("node1".to_string(), 5);
        // 5 nodes: quorum 3, byzantine tolerance 1
        assert_eq!(cp.byzantine_tolerance(), 1);
    }

    #[tokio::test]
    async fn test_clear_votes() {
        let cp = ConsensusProtocol::new("node1".to_string(), 3);

        let vote = Vote {
            term: Term(1),
            candidate_id: "node2".to_string(),
            last_log_index: 0,
            last_log_term: Term(0),
            voter_id: "node1".to_string(),
            granted: true,
        };

        cp.submit_vote(vote).await.unwrap();
        let votes = cp.get_votes(Term(1)).await;
        assert_eq!(votes.len(), 1);

        cp.clear_votes(Term(1)).await;
        let votes = cp.get_votes(Term(1)).await;
        assert_eq!(votes.len(), 0);
    }

    #[test]
    fn test_election_counter() {
        let cp = ConsensusProtocol::new("node1".to_string(), 3);
        assert_eq!(cp.election_count(), 0);

        cp.increment_elections();
        assert_eq!(cp.election_count(), 1);

        cp.increment_elections();
        assert_eq!(cp.election_count(), 2);
    }
}
