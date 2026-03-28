//! Raft coordinator with leader election and heartbeat mechanism

use super::types::{Term, NodeState, Vote, Heartbeat, LogEntry, LogCommand};
use super::consensus::ConsensusProtocol;
use anyhow::Result;
use std::collections::HashMap;
use std::sync::Arc;
use tokio::time::{Duration, Instant};

/// Transport abstraction for sending heartbeats to peers.
#[async_trait::async_trait]
pub trait PeerTransport: Send + Sync {
    async fn send(&self, peer_addr: &str, heartbeat: &Heartbeat) -> Result<()>;
}

/// HTTP-based peer transport: POSTs heartbeat JSON to `http://{peer_addr}/raft/heartbeat`.
pub struct HttpPeerTransport {
    client: reqwest::Client,
}

impl HttpPeerTransport {
    pub fn new() -> Self {
        Self {
            client: reqwest::Client::new(),
        }
    }
}

impl Default for HttpPeerTransport {
    fn default() -> Self {
        Self::new()
    }
}

#[async_trait::async_trait]
impl PeerTransport for HttpPeerTransport {
    async fn send(&self, peer_addr: &str, heartbeat: &Heartbeat) -> Result<()> {
        let url = format!("http://{}/raft/heartbeat", peer_addr);
        self.client
            .post(&url)
            .json(heartbeat)
            .send()
            .await
            .map_err(|e| anyhow::anyhow!("heartbeat send failed to {}: {}", peer_addr, e))?;
        Ok(())
    }
}

/// Raft coordinator managing leader election and log replication
pub struct RaftCoordinator {
    pub node_id: String,
    pub state: NodeState,
    pub current_term: Term,
    pub voted_for: Option<String>,
    pub log: Vec<LogEntry>,
    pub commit_index: u64,
    pub last_applied: u64,
    pub peers: Vec<String>,
    pub next_index: HashMap<String, u64>,
    pub match_index: HashMap<String, u64>,
    pub last_heartbeat: Instant,
    pub transport: Arc<dyn PeerTransport>,
}

impl RaftCoordinator {
    /// Create a new coordinator with given node ID and peer list.
    /// Uses `HttpPeerTransport` by default.
    pub fn new(node_id: String, peers: Vec<String>) -> Self {
        Self::with_transport(node_id, peers, Arc::new(HttpPeerTransport::new()))
    }

    /// Create a new coordinator with a custom transport (useful for testing).
    pub fn with_transport(node_id: String, peers: Vec<String>, transport: Arc<dyn PeerTransport>) -> Self {
        let mut next_index = HashMap::new();
        let mut match_index = HashMap::new();
        for peer in &peers {
            next_index.insert(peer.clone(), 0);
            match_index.insert(peer.clone(), 0);
        }

        Self {
            node_id,
            state: NodeState::Follower,
            current_term: Term::new(),
            voted_for: None,
            log: Vec::new(),
            commit_index: 0,
            last_applied: 0,
            peers,
            next_index,
            match_index,
            last_heartbeat: Instant::now(),
            transport,
        }
    }

    /// Get current leader if known
    pub fn get_leader(&self) -> Option<String> {
        if self.state == NodeState::Leader {
            Some(self.node_id.clone())
        } else {
            None
        }
    }

    /// Check if this node is the leader
    pub fn is_leader(&self) -> bool {
        self.state == NodeState::Leader
    }

    /// Start leader election (transition to Candidate)
    pub async fn start_election(&mut self, consensus: std::sync::Arc<ConsensusProtocol>) -> Result<()> {
        self.current_term.increment();
        self.state = NodeState::Candidate;
        self.voted_for = Some(self.node_id.clone());

        // Submit vote for self
        let vote = Vote {
            term: self.current_term,
            candidate_id: self.node_id.clone(),
            last_log_index: self.log.len() as u64,
            last_log_term: self.log.last().map(|e| e.term).unwrap_or(Term::new()),
            voter_id: self.node_id.clone(),
            granted: true,
        };

        consensus.submit_vote(vote).await?;
        Ok(())
    }

    /// Become leader (only after winning election)
    pub fn become_leader(&mut self) {
        self.state = NodeState::Leader;
        let next_index = self.log.len() as u64;
        for peer in &self.peers {
            self.next_index.insert(peer.clone(), next_index);
            self.match_index.insert(peer.clone(), 0);
        }
    }

    /// Handle incoming vote from another node
    pub fn handle_vote(&mut self, vote: Vote) -> bool {
        // Reject if term is outdated
        if vote.term < self.current_term {
            return false;
        }

        // Update term if newer
        if vote.term > self.current_term {
            self.current_term = vote.term;
            self.state = NodeState::Follower;
            self.voted_for = None;
        }

        // Check if we've already voted in this term
        if let Some(ref voted_for) = self.voted_for {
            if voted_for != &vote.candidate_id {
                return false;
            }
        }

        // Check candidate log is at least as up-to-date as ours
        let last_log_term = self.log.last().map(|e| e.term).unwrap_or(Term::new());
        let last_log_index = self.log.len() as u64;

        if vote.last_log_term < last_log_term
            || (vote.last_log_term == last_log_term && vote.last_log_index < last_log_index)
        {
            return false;
        }

        self.voted_for = Some(vote.candidate_id.clone());
        true
    }

    /// Append log entry (as leader)
    pub fn append_log_entry(&mut self, command: LogCommand) -> u64 {
        let entry = LogEntry {
            term: self.current_term,
            index: self.log.len() as u64,
            command,
        };
        self.log.push(entry);
        self.log.len() as u64 - 1
    }

    /// Build heartbeat message for followers
    pub fn build_heartbeat(&self, peer: &str) -> Option<Heartbeat> {
        if self.state != NodeState::Leader {
            return None;
        }

        let prev_index = self.next_index.get(peer)?.saturating_sub(1);
        let prev_term = if prev_index == 0 {
            Term::new()
        } else {
            self.log.get(prev_index as usize).map(|e| e.term).unwrap_or(Term::new())
        };

        let mut entries = Vec::new();
        if let Some(next_idx) = self.next_index.get(peer) {
            for i in *next_idx..self.log.len() as u64 {
                if let Some(entry) = self.log.get(i as usize) {
                    entries.push(entry.clone());
                }
            }
        }

        Some(Heartbeat {
            term: self.current_term,
            leader_id: self.node_id.clone(),
            prev_log_index: prev_index,
            prev_log_term: prev_term,
            leader_commit: self.commit_index,
            entries,
        })
    }

    /// Send heartbeat to all followers via the configured transport.
    pub async fn send_heartbeat(&self) -> Result<()> {
        if self.state != NodeState::Leader {
            return Ok(());
        }

        for peer in &self.peers {
            if let Some(heartbeat) = self.build_heartbeat(peer) {
                self.transport.send(peer, &heartbeat).await?;
            }
        }
        Ok(())
    }

    /// Check if node election timeout has elapsed
    pub fn election_timeout_elapsed(&self) -> bool {
        self.last_heartbeat.elapsed() > Duration::from_millis(150)
    }

    /// Reset heartbeat timer
    pub fn reset_heartbeat_timer(&mut self) {
        self.last_heartbeat = Instant::now();
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_coordinator_creation() {
        let peers = vec!["node2".to_string(), "node3".to_string()];
        let coord = RaftCoordinator::new("node1".to_string(), peers.clone());

        assert_eq!(coord.node_id, "node1");
        assert_eq!(coord.state, NodeState::Follower);
        assert_eq!(coord.current_term.0, 0);
        assert_eq!(coord.peers, peers);
        assert!(!coord.is_leader());
    }

    #[test]
    fn test_become_candidate() {
        let coord = RaftCoordinator::new("node1".to_string(), vec![]);
        assert_eq!(coord.state, NodeState::Follower);
        assert_eq!(coord.current_term.0, 0);

        // Not testing full election, just state transition
    }

    #[test]
    fn test_become_leader() {
        let mut coord = RaftCoordinator::new("node1".to_string(), vec!["node2".to_string()]);
        coord.become_leader();

        assert!(coord.is_leader());
        assert_eq!(coord.state, NodeState::Leader);
        assert!(coord.next_index.contains_key("node2"));
    }

    #[test]
    fn test_append_log_entry() {
        let mut coord = RaftCoordinator::new("node1".to_string(), vec![]);
        coord.become_leader();

        let cmd = LogCommand::DiscoverPartition {
            partition_id: "partition_a".to_string(),
            case_ids: vec!["case1".to_string()],
        };

        let index = coord.append_log_entry(cmd);
        assert_eq!(index, 0);
        assert_eq!(coord.log.len(), 1);
    }

    #[test]
    fn test_handle_vote_valid() {
        let mut coord = RaftCoordinator::new("node1".to_string(), vec![]);

        let vote = Vote {
            term: Term(1),
            candidate_id: "node2".to_string(),
            last_log_index: 0,
            last_log_term: Term(0),
            voter_id: "node1".to_string(),
            granted: true,
        };

        assert!(coord.handle_vote(vote.clone()));
        assert_eq!(coord.current_term.0, 1);
    }

    #[test]
    fn test_handle_vote_outdated_term() {
        let mut coord = RaftCoordinator::new("node1".to_string(), vec![]);
        coord.current_term = Term(5);

        let vote = Vote {
            term: Term(3),
            candidate_id: "node2".to_string(),
            last_log_index: 0,
            last_log_term: Term(0),
            voter_id: "node1".to_string(),
            granted: true,
        };

        assert!(!coord.handle_vote(vote));
    }

    #[test]
    fn test_heartbeat_election_timeout() {
        let coord = RaftCoordinator::new("node1".to_string(), vec![]);
        // Fresh coordinator should not have timeout elapsed yet
        assert!(!coord.election_timeout_elapsed());
    }

    #[test]
    fn test_build_heartbeat_as_leader() {
        let mut coord = RaftCoordinator::new("node1".to_string(), vec!["node2".to_string()]);
        coord.become_leader();

        let hb = coord.build_heartbeat("node2");
        assert!(hb.is_some());
        let hb = hb.unwrap();
        assert_eq!(hb.leader_id, "node1");
        assert_eq!(hb.term, coord.current_term);
    }

    #[test]
    fn test_build_heartbeat_as_follower() {
        let coord = RaftCoordinator::new("node1".to_string(), vec!["node2".to_string()]);
        // Follower cannot build heartbeat
        assert!(coord.build_heartbeat("node2").is_none());
    }
}
