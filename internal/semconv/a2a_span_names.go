package semconv

const (
	// a2a_auction_run is the span name for "a2a.auction.run".
	//
	// Running an A2A capability auction — agents bid for task allocation based on capability and cost.
	// Kind: internal
	// Stability: development
	A2aAuctionRunSpan = "a2a.auction.run"
	// a2a_bid_evaluate is the span name for "a2a.bid.evaluate".
	//
	// Bid evaluation — scoring and ranking agent bids to select the best provider for a task.
	// Kind: internal
	// Stability: development
	A2aBidEvaluateSpan = "a2a.bid.evaluate"
	// a2a_call is the span name for "a2a.call".
	//
	// An agent-to-agent call — one ChatmanGPT service invoking another via the A2A protocol.
	// Kind: client
	// Stability: development
	A2aCallSpan = "a2a.call"
	// a2a_capability_match is the span name for "a2a.capability.match".
	//
	// Matching a capability request to available agents — selecting best provider.
	// Kind: internal
	// Stability: development
	A2aCapabilityMatchSpan = "a2a.capability.match"
	// a2a_capability_negotiate is the span name for "a2a.capability.negotiate".
	//
	// Capability negotiation between two A2A agents — determining what capabilities can be fulfilled.
	// Kind: client
	// Stability: development
	A2aCapabilityNegotiateSpan = "a2a.capability.negotiate"
	// a2a_capability_register is the span name for "a2a.capability.register".
	//
	// Registration of an agent capability in the A2A capability registry.
	// Kind: server
	// Stability: development
	A2aCapabilityRegisterSpan = "a2a.capability.register"
	// a2a_contract_amend is the span name for "a2a.contract.amend".
	//
	// Contract amendment — negotiating a modification to an existing A2A service contract.
	// Kind: client
	// Stability: development
	A2aContractAmendSpan = "a2a.contract.amend"
	// a2a_contract_dispute is the span name for "a2a.contract.dispute".
	//
	// Initiating or updating an A2A contract dispute between agents.
	// Kind: internal
	// Stability: development
	A2aContractDisputeSpan = "a2a.contract.dispute"
	// a2a_contract_execute is the span name for "a2a.contract.execute".
	//
	// Execution of an A2A service contract — running contract obligations and tracking progress toward completion.
	// Kind: internal
	// Stability: development
	A2aContractExecuteSpan = "a2a.contract.execute"
	// a2a_contract_negotiate is the span name for "a2a.contract.negotiate".
	//
	// Negotiation of an A2A service contract — establishing terms, SLA, and obligations between two agents.
	// Kind: client
	// Stability: development
	A2aContractNegotiateSpan = "a2a.contract.negotiate"
	// a2a_create_deal is the span name for "a2a.create_deal".
	//
	// Creation of an A2A deal between two agents.
	// Kind: server
	// Stability: development
	A2aCreateDealSpan = "a2a.create_deal"
	// a2a_deal_status_transition is the span name for "a2a.deal.status_transition".
	//
	// Status transition of an A2A deal through its lifecycle (pending → active → completed).
	// Kind: internal
	// Stability: development
	A2aDealStatusTransitionSpan = "a2a.deal.status_transition"
	// a2a_dispute_resolve is the span name for "a2a.dispute.resolve".
	//
	// Resolution of an A2A dispute between agents — arbitration and settlement process.
	// Kind: internal
	// Stability: development
	A2aDisputeResolveSpan = "a2a.dispute.resolve"
	// a2a_escrow_create is the span name for "a2a.escrow.create".
	//
	// A2A escrow creation — establishing a payment escrow for a deal between two agents.
	// Kind: server
	// Stability: development
	A2aEscrowCreateSpan = "a2a.escrow.create"
	// a2a_escrow_release is the span name for "a2a.escrow.release".
	//
	// A2A escrow release — settling a payment escrow upon deal completion or dispute resolution.
	// Kind: server
	// Stability: development
	A2aEscrowReleaseSpan = "a2a.escrow.release"
	// a2a_knowledge_transfer is the span name for "a2a.knowledge.transfer".
	//
	// Transfer of knowledge or capability data between agents via A2A.
	// Kind: producer
	// Stability: development
	A2aKnowledgeTransferSpan = "a2a.knowledge.transfer"
	// a2a_message_batch is the span name for "a2a.message.batch".
	//
	// Batched delivery of multiple A2A messages — aggregates messages for efficient transport.
	// Kind: producer
	// Stability: development
	A2aMessageBatchSpan = "a2a.message.batch"
	// a2a_message_route is the span name for "a2a.message.route".
	//
	// Routing of an A2A message to the appropriate target agent based on priority and routing rules.
	// Kind: producer
	// Stability: development
	A2aMessageRouteSpan = "a2a.message.route"
	// a2a_negotiate is the span name for "a2a.negotiate".
	//
	// Multi-round deal negotiation between two agents.
	// Kind: client
	// Stability: development
	A2aNegotiateSpan = "a2a.negotiate"
	// a2a_negotiation_state_transition is the span name for "a2a.negotiation.state_transition".
	//
	// State machine transition in an A2A multi-round negotiation protocol.
	// Kind: internal
	// Stability: development
	A2aNegotiationStateTransitionSpan = "a2a.negotiation.state_transition"
	// a2a_penalty_apply is the span name for "a2a.penalty.apply".
	//
	// Applying a penalty or reward to an agent based on contract performance — updates trust score and balance.
	// Kind: server
	// Stability: development
	A2aPenaltyApplySpan = "a2a.penalty.apply"
	// a2a_protocol_negotiate is the span name for "a2a.protocol.negotiate".
	//
	// A2A protocol version negotiation between two agents — determining compatible protocol version.
	// Kind: client
	// Stability: development
	A2aProtocolNegotiateSpan = "a2a.protocol.negotiate"
	// a2a_reputation_decay is the span name for "a2a.reputation.decay".
	//
	// A2A reputation decay event — applying time-based or violation-triggered reputation score reduction.
	// Kind: internal
	// Stability: development
	A2aReputationDecaySpan = "a2a.reputation.decay"
	// a2a_reputation_update is the span name for "a2a.reputation.update".
	//
	// Updating an agent's reputation score based on the outcome of a completed interaction.
	// Kind: internal
	// Stability: development
	A2aReputationUpdateSpan = "a2a.reputation.update"
	// a2a_sla_check is the span name for "a2a.sla.check".
	//
	// SLA validation for an A2A operation — measures actual latency against deadline.
	// Kind: internal
	// Stability: development
	A2aSlaCheckSpan = "a2a.sla.check"
	// a2a_slo_evaluate is the span name for "a2a.slo.evaluate".
	//
	// SLO evaluation — assessing whether A2A operation met service level objectives.
	// Kind: internal
	// Stability: development
	A2aSloEvaluateSpan = "a2a.slo.evaluate"
	// a2a_task_delegate is the span name for "a2a.task.delegate".
	//
	// Delegation of a task from one agent to another via A2A.
	// Kind: producer
	// Stability: development
	A2aTaskDelegateSpan = "a2a.task.delegate"
	// a2a_trust_evaluate is the span name for "a2a.trust.evaluate".
	//
	// Evaluation of an agent's trust score based on reputation history and interaction outcomes.
	// Kind: internal
	// Stability: development
	A2aTrustEvaluateSpan = "a2a.trust.evaluate"
	// a2a_trust_federate is the span name for "a2a.trust.federate".
	//
	// Federated trust evaluation — agent joins or queries a trust ring for cross-federation capability authorization.
	// Kind: client
	// Stability: development
	A2aTrustFederateSpan = "a2a.trust.federate"
)