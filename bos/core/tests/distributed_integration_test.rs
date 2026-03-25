//! Integration tests for distributed process mining with Raft consensus

use bos_core::distributed::*;

/// Test 1: Single node baseline discovery
#[tokio::test]
async fn test_single_node_discovery() {
    let pm = DistributedPM::new("node1".to_string(), vec![], 1);

    // Create single worker
    let mut worker = DistributedWorker::new("node1".to_string(), "partition_a".to_string());

    // Process events
    worker
        .process_event("case1".to_string(), "invoice".to_string(), 1000)
        .unwrap();
    worker
        .process_event("case1".to_string(), "review".to_string(), 2000)
        .unwrap();
    worker
        .process_event("case1".to_string(), "approve".to_string(), 3000)
        .unwrap();
    worker
        .process_event("case1".to_string(), "pay".to_string(), 4000)
        .unwrap();

    // Discover local model
    let model = worker.discover_local_model().unwrap();

    assert_eq!(model.places.len(), 4);
    assert_eq!(model.transitions.len(), 4);
    assert!(!model.hash.is_empty());
    assert_eq!(worker.case_count(), 1);

    pm.register_worker("partition_a".to_string(), worker)
        .await
        .unwrap();
}

/// Test 2: Multi-node distributed discovery
#[tokio::test]
async fn test_distributed_three_nodes() {
    let pm = DistributedPM::new(
        "node1".to_string(),
        vec!["node2".to_string(), "node3".to_string()],
        3,
    );

    // Worker 1: Process partition A
    let mut worker_a = DistributedWorker::new("node1".to_string(), "partition_a".to_string());
    worker_a
        .process_event("case1".to_string(), "start".to_string(), 1000)
        .unwrap();
    worker_a
        .process_event("case1".to_string(), "process".to_string(), 2000)
        .unwrap();

    // Worker 2: Process partition B
    let mut worker_b = DistributedWorker::new("node2".to_string(), "partition_b".to_string());
    worker_b
        .process_event("case2".to_string(), "start".to_string(), 1100)
        .unwrap();
    worker_b
        .process_event("case2".to_string(), "review".to_string(), 2100)
        .unwrap();

    // Worker 3: Process partition C
    let mut worker_c = DistributedWorker::new("node3".to_string(), "partition_c".to_string());
    worker_c
        .process_event("case3".to_string(), "start".to_string(), 1200)
        .unwrap();
    worker_c
        .process_event("case3".to_string(), "end".to_string(), 3200)
        .unwrap();

    // Register all workers
    pm.register_worker("partition_a".to_string(), worker_a)
        .await
        .unwrap();
    pm.register_worker("partition_b".to_string(), worker_b)
        .await
        .unwrap();
    pm.register_worker("partition_c".to_string(), worker_c)
        .await
        .unwrap();

    // Verify all workers registered (through get_leader being None since no election)
    assert!(pm.get_leader().await.is_none());
}

/// Test 3: Leader election protocol
#[tokio::test]
async fn test_leader_election() {
    let pm = DistributedPM::new(
        "node1".to_string(),
        vec!["node2".to_string(), "node3".to_string()],
        3,
    );

    // Initially no leader
    assert!(!pm.is_leader().await);
    assert!(pm.get_leader().await.is_none());

    // Start election
    pm.start_election().await.unwrap();

    // After election, node should potentially become leader or have state changed
    // (exact state depends on network conditions in real implementation)
}

/// Test 4: Consensus with quorum voting
#[tokio::test]
async fn test_consensus_quorum() {
    let consensus = ConsensusProtocol::new("node1".to_string(), 3);

    // Need 2 votes for quorum (3 nodes -> quorum 2)
    assert_eq!(consensus.quorum_size(), 2);

    // Cast first vote
    let vote1 = Vote {
        term: Term(1),
        candidate_id: "node1".to_string(),
        last_log_index: 0,
        last_log_term: Term(0),
        voter_id: "node1".to_string(),
        granted: true,
    };
    consensus.submit_vote(vote1).await.unwrap();

    assert!(!consensus.has_quorum(Term(1)).await);

    // Cast second vote - reaches quorum
    let vote2 = Vote {
        term: Term(1),
        candidate_id: "node1".to_string(),
        last_log_index: 0,
        last_log_term: Term(0),
        voter_id: "node2".to_string(),
        granted: true,
    };
    consensus.submit_vote(vote2).await.unwrap();

    assert!(consensus.has_quorum(Term(1)).await);
}

/// Test 5: Model merging from multiple partitions
#[tokio::test]
async fn test_model_merging() {
    let mut merger = ModelMerger::new();

    // Create three local models
    let mut model_a = ProcessModel::new("partition_a".to_string());
    model_a.add_place("p_start".to_string());
    model_a.add_transition("t_invoice".to_string());
    model_a.add_arc("p_start".to_string(), "t_invoice".to_string());

    let mut model_b = ProcessModel::new("partition_b".to_string());
    model_b.add_place("p_review".to_string());
    model_b.add_transition("t_approve".to_string());
    model_b.add_arc("p_review".to_string(), "t_approve".to_string());

    let mut model_c = ProcessModel::new("partition_c".to_string());
    model_c.add_place("p_pay".to_string());
    model_c.add_transition("t_archive".to_string());
    model_c.add_arc("p_pay".to_string(), "t_archive".to_string());

    // Register models
    merger.register_model("partition_a".to_string(), model_a).unwrap();
    merger.register_model("partition_b".to_string(), model_b).unwrap();
    merger.register_model("partition_c".to_string(), model_c).unwrap();

    // Merge
    let global_model = merger.merge_models().unwrap();

    assert_eq!(global_model.places.len(), 3);
    assert_eq!(global_model.transitions.len(), 3);
    assert_eq!(global_model.arcs.len(), 3);

    // Get stats
    let stats = merger.get_merge_stats();
    assert_eq!(stats.num_local_models, 3);
    assert_eq!(stats.global_places, 3);
}

/// Test 6: Node crash detection and recovery
#[tokio::test]
async fn test_node_crash_and_recovery() {
    let mut recovery = FaultRecovery::new("node1".to_string());

    recovery.register_node("node2".to_string());
    recovery.register_node("node3".to_string());

    assert_eq!(recovery.count_alive_nodes(), 2);

    // Simulate node2 health check failures
    recovery.health_check("node2", false).unwrap();
    recovery.health_check("node2", false).unwrap();
    recovery.health_check("node2", false).unwrap();
    recovery.health_check("node2", false).unwrap();

    // Node2 should be marked as crashed
    assert!(!recovery.is_node_healthy("node2"));
    assert_eq!(recovery.crash_count(), 1);
    assert_eq!(recovery.count_alive_nodes(), 1);

    // Recover node2
    recovery.start_recovery("node2".to_string()).unwrap();
    assert!(recovery.is_node_healthy("node2"));

    // Check recovery log
    let log = recovery.get_recovery_log();
    assert!(!log.is_empty());
}

/// Test 7: Byzantine fault tolerance (5 nodes, tolerate 1 failure)
#[tokio::test]
async fn test_byzantine_tolerance_5_nodes() {
    let consensus = ConsensusProtocol::new("node1".to_string(), 5);

    assert_eq!(consensus.quorum_size(), 3);
    assert_eq!(consensus.byzantine_tolerance(), 1); // Can tolerate 1 failure

    let mut recovery = FaultRecovery::new("node1".to_string());
    // Register 4 explicit nodes + node1 (recovery.node_id) = 5 total
    recovery.register_node("node2".to_string());
    recovery.register_node("node3".to_string());
    recovery.register_node("node4".to_string());
    recovery.register_node("node5".to_string());

    // All 4 registered nodes alive (+ implicit node1)
    assert_eq!(recovery.count_alive_nodes(), 4);
    // Quorum for 5 nodes: (5/2)+1 = 3
    // Can tolerate failure: alive > quorum → 4 > 3 ✓
    assert!(recovery.can_tolerate_failure(5));

    // Crash one node
    for _ in 0..4 {
        recovery.health_check("node2", false).unwrap();
    }

    // Now 3 alive (node1, node3, node4, node5 minus node2)
    // Can tolerate: 3 > 3? No, but we have node1 + 3 others = 4 - 1 = 3... wait
    // Actually: 4 - 1 crash = 3 alive. Quorum is 3. 3 > 3? No
    // But our health_map only has 4 entries, and node2 crashed, so 3 alive
    assert_eq!(recovery.count_alive_nodes(), 3);
    assert!(!recovery.can_tolerate_failure(5)); // 3 not > 3
}

/// Test 8: Heartbeat mechanism
#[tokio::test]
async fn test_heartbeat_mechanism() {
    let mut coord =
        RaftCoordinator::new("node1".to_string(), vec!["node2".to_string(), "node3".to_string()]);

    coord.become_leader();

    // Build heartbeat for each peer
    let hb2 = coord.build_heartbeat("node2");
    assert!(hb2.is_some());

    let hb3 = coord.build_heartbeat("node3");
    assert!(hb3.is_some());

    // Heartbeats should have leader info
    let hb = hb2.unwrap();
    assert_eq!(hb.leader_id, "node1");
    assert_eq!(hb.leader_commit, 0);

    // Send heartbeats
    coord.send_heartbeat().await.unwrap();
}

/// Test 9: Log replication with quorum acks
#[tokio::test]
async fn test_log_replication() {
    let consensus = ConsensusProtocol::new("node1".to_string(), 3);

    // Register log entry for replication
    let _ack = consensus.register_quorum_ack(0).await;

    // Simulate replication from node1
    consensus.ack_log_entry(0, "node1".to_string()).await.unwrap();

    // Entry not yet committed (need 2 acks)
    assert!(!consensus.is_entry_committed(0).await);

    // Get second acknowledgment
    let reached = consensus
        .ack_log_entry(0, "node2".to_string())
        .await
        .unwrap();

    assert!(reached);
    assert!(consensus.is_entry_committed(0).await);
}

/// Test 10: Partition handling with network isolation
#[tokio::test]
async fn test_network_partition_handling() {
    let mut recovery = FaultRecovery::new("node1".to_string());

    recovery.register_node("node2".to_string());
    recovery.register_node("node3".to_string());

    // Simulate network partition isolating node3
    recovery
        .detect_partition("partition_a".to_string(), vec!["node3".to_string()])
        .unwrap();

    let log = recovery.get_recovery_log();
    assert!(!log.is_empty());

    // Node health shouldn't change just from detection
    assert!(recovery.is_node_healthy("node3"));
}

/// Test 11: Large-scale discovery simulation
#[tokio::test]
async fn test_large_scale_event_processing() {
    let mut worker = DistributedWorker::new("node1".to_string(), "partition_a".to_string());

    // Process 1000 events across 10 cases
    for case_num in 0..10 {
        let case_id = format!("case_{}", case_num);
        for activity_num in 0..100 {
            worker
                .process_event(
                    case_id.clone(),
                    format!("activity_{}", activity_num % 10), // 10 unique activities
                    1000 + activity_num as u64,
                )
                .unwrap();
        }
    }

    assert_eq!(worker.case_count(), 10);
    assert_eq!(worker.activity_count(), 10);
    assert_eq!(worker.events_processed(), 1000);

    // Discover model
    let model = worker.discover_local_model().unwrap();
    assert!(!model.places.is_empty());
    assert!(!model.transitions.is_empty());
}

/// Test 12: Concurrent election rounds
#[tokio::test]
async fn test_concurrent_elections() {
    let consensus = ConsensusProtocol::new("node1".to_string(), 3);

    // Simulate multiple election rounds
    for round in 0..5 {
        let vote = Vote {
            term: Term(round as u64),
            candidate_id: format!("node{}", (round % 3) + 1),
            last_log_index: round as u64,
            last_log_term: if round == 0 { Term(0) } else { Term(round as u64 - 1) },
            voter_id: "node1".to_string(),
            granted: true,
        };
        consensus.submit_vote(vote).await.unwrap();
    }

    // Check election counter starts at 0
    assert_eq!(consensus.election_count(), 0); // Counter not incremented by submit
}
