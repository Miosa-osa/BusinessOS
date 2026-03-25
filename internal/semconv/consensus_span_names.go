package semconv

const (
	// consensus_block_commit is the span name for "consensus.block.commit".
	//
	// Committing a decided value as a block in the HotStuff BFT log.
	// Kind: internal
	// Stability: development
	ConsensusBlockCommitSpan = "consensus.block.commit"
	// consensus_byzantine_recover is the span name for "consensus.byzantine.recover".
	//
	// Byzantine fault recovery — adjusts quorum and restores consensus after detecting byzantine behavior.
	// Kind: internal
	// Stability: development
	ConsensusByzantineRecoverSpan = "consensus.byzantine.recover"
	// consensus_epoch_advance is the span name for "consensus.epoch.advance".
	//
	// Epoch advancement — consensus protocol advances to a new epoch after configuration change or key rotation.
	// Kind: internal
	// Stability: development
	ConsensusEpochAdvanceSpan = "consensus.epoch.advance"
	// consensus_epoch_finalize is the span name for "consensus.epoch.finalize".
	//
	// Epoch finalization — collecting signatures and committing the final state of a consensus epoch.
	// Kind: internal
	// Stability: development
	ConsensusEpochFinalizeSpan = "consensus.epoch.finalize"
	// consensus_epoch_key_rotate is the span name for "consensus.epoch.key_rotate".
	//
	// Epoch key rotation — rotating cryptographic keys for a consensus epoch after a configuration change or compromise.
	// Kind: internal
	// Stability: development
	ConsensusEpochKeyRotateSpan = "consensus.epoch.key_rotate"
	// consensus_epoch_quorum_snapshot is the span name for "consensus.epoch.quorum_snapshot".
	//
	// Epoch quorum snapshot — capturing the quorum membership set at an epoch boundary.
	// Kind: internal
	// Stability: development
	ConsensusEpochQuorumSnapshotSpan = "consensus.epoch.quorum_snapshot"
	// consensus_epoch_transition is the span name for "consensus.epoch.transition".
	//
	// Epoch transition in the consensus protocol — moving from one epoch to the next.
	// Kind: internal
	// Stability: development
	ConsensusEpochTransitionSpan = "consensus.epoch.transition"
	// consensus_fork_detect is the span name for "consensus.fork.detect".
	//
	// Fork detection in the consensus chain — identifies diverged branches and applies resolution strategy.
	// Kind: internal
	// Stability: development
	ConsensusForkDetectSpan = "consensus.fork.detect"
	// consensus_leader_rotate is the span name for "consensus.leader.rotate".
	//
	// Leader rotation event — current leader yields and new leader is selected via scoring.
	// Kind: internal
	// Stability: development
	ConsensusLeaderRotateSpan = "consensus.leader.rotate"
	// consensus_leader_election is the span name for "consensus.leader_election".
	//
	// Leader election event in HotStuff BFT — new leader selected after view change.
	// Kind: internal
	// Stability: development
	ConsensusLeaderElectionSpan = "consensus.leader_election"
	// consensus_liveness_check is the span name for "consensus.liveness.check".
	//
	// Verifying liveness of the consensus protocol — confirming progress is being made.
	// Kind: internal
	// Stability: development
	ConsensusLivenessCheckSpan = "consensus.liveness.check"
	// consensus_network_recovery is the span name for "consensus.network.recovery".
	//
	// Network recovery — restoring consensus network connectivity after partition or node failure.
	// Kind: internal
	// Stability: development
	ConsensusNetworkRecoverySpan = "consensus.network.recovery"
	// consensus_network_topology is the span name for "consensus.network.topology".
	//
	// Network topology snapshot — capturing current consensus cluster topology for analysis and fault diagnosis.
	// Kind: internal
	// Stability: development
	ConsensusNetworkTopologySpan = "consensus.network.topology"
	// consensus_partition_recover is the span name for "consensus.partition.recover".
	//
	// Network partition recovery — restoring consensus after a partition splits the replica set.
	// Kind: internal
	// Stability: development
	ConsensusPartitionRecoverSpan = "consensus.partition.recover"
	// consensus_quorum_grow is the span name for "consensus.quorum.grow".
	//
	// Quorum growth operation — adding new replicas to expand the consensus quorum size.
	// Kind: internal
	// Stability: development
	ConsensusQuorumGrowSpan = "consensus.quorum.grow"
	// consensus_quorum_shrink is the span name for "consensus.quorum.shrink".
	//
	// Quorum shrink operation — removing nodes from the consensus quorum safely.
	// Kind: internal
	// Stability: development
	ConsensusQuorumShrinkSpan = "consensus.quorum.shrink"
	// consensus_replica_sync is the span name for "consensus.replica.sync".
	//
	// Synchronization of a replica to catch up with the consensus leader.
	// Kind: internal
	// Stability: development
	ConsensusReplicaSyncSpan = "consensus.replica.sync"
	// consensus_round is the span name for "consensus.round".
	//
	// A single round in the OSA HotStuff BFT consensus protocol.
	// Kind: internal
	// Stability: development
	ConsensusRoundSpan = "consensus.round"
	// consensus_safety_check is the span name for "consensus.safety.check".
	//
	// Checking consensus safety — validating that quorum meets safety threshold before committing.
	// Kind: internal
	// Stability: development
	ConsensusSafetyCheckSpan = "consensus.safety.check"
	// consensus_safety_monitor is the span name for "consensus.safety.monitor".
	//
	// Ongoing safety monitoring — continuously verifies BFT safety invariants across replica set.
	// Kind: internal
	// Stability: development
	ConsensusSafetyMonitorSpan = "consensus.safety.monitor"
	// consensus_safety_violation is the span name for "consensus.safety.violation".
	//
	// Safety violation detected in the consensus protocol — double voting, equivocation, or quorum breach.
	// Kind: internal
	// Stability: development
	ConsensusSafetyViolationSpan = "consensus.safety.violation"
	// consensus_threshold_adapt is the span name for "consensus.threshold.adapt".
	//
	// Consensus threshold adaptation — dynamically adjusting the quorum threshold based on observed fault rates and network conditions.
	// Kind: internal
	// Stability: development
	ConsensusThresholdAdaptSpan = "consensus.threshold.adapt"
	// consensus_threshold_vote is the span name for "consensus.threshold.vote".
	//
	// Consensus threshold voting — executing a threshold-based vote among replicas.
	// Kind: internal
	// Stability: development
	ConsensusThresholdVoteSpan = "consensus.threshold.vote"
	// consensus_timeout_event is the span name for "consensus.timeout_event".
	//
	// View timeout event — current view timed out, triggering view change protocol.
	// Kind: internal
	// Stability: development
	ConsensusTimeoutEventSpan = "consensus.timeout_event"
	// consensus_view_change is the span name for "consensus.view_change".
	//
	// View change event — leader timeout triggered, transitioning to new leader.
	// Kind: internal
	// Stability: development
	ConsensusViewChangeSpan = "consensus.view_change"
	// consensus_view_change_optimize is the span name for "consensus.view_change.optimize".
	//
	// Optimized view change with exponential backoff — reduces thrashing during network instability.
	// Kind: internal
	// Stability: development
	ConsensusViewChangeOptimizeSpan = "consensus.view_change.optimize"
	// consensus_vote is the span name for "consensus.vote".
	//
	// Casting or receiving a single vote in a HotStuff BFT round.
	// Kind: internal
	// Stability: development
	ConsensusVoteSpan = "consensus.vote"
)