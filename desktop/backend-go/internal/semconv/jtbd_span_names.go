package semconv

const (
	// jtbd_loop is the span name for "jtbd.loop".
	//
	// A complete iteration of a 10-scenario JTBD loop execution across ChatmanGPT integration chain.
	// Kind: internal
	// Stability: development
	JtbdLoopSpan = "jtbd.loop"
	// jtbd_scenario is the span name for "jtbd.scenario".
	//
	// A single step in a JTBD (Jobs-to-be-Done) scenario execution across ChatmanGPT systems.
	// Kind: internal
	// Stability: development
	JtbdScenarioSpan = "jtbd.scenario"
	// jtbd_scenario_contract_closure is the span name for "jtbd.scenario.contract_closure".
	//
	// A JTBD scenario step for closing and signing contracts with blockchain validation.
	// Kind: internal
	// Stability: development
	JtbdScenarioContractClosureSpan = "jtbd.scenario.contract_closure"
	// jtbd_scenario_deal_progression is the span name for "jtbd.scenario.deal_progression".
	//
	// A JTBD scenario step for progressing deals through CRM pipeline stages.
	// Kind: internal
	// Stability: development
	JtbdScenarioDealProgressionSpan = "jtbd.scenario.deal_progression"
	// jtbd_scenario_icp_qualification is the span name for "jtbd.scenario.icp_qualification".
	//
	// A JTBD scenario step for ICP (Ideal Customer Profile) qualification in RevOps workflows.
	// Kind: internal
	// Stability: development
	JtbdScenarioIcpQualificationSpan = "jtbd.scenario.icp_qualification"
	// jtbd_scenario_outreach_sequence_execution is the span name for "jtbd.scenario.outreach_sequence_execution".
	//
	// A JTBD scenario step for executing multi-step outreach sequences in RevOps.
	// Kind: internal
	// Stability: development
	JtbdScenarioOutreachSequenceExecutionSpan = "jtbd.scenario.outreach_sequence_execution"
	// jtbd_scenario_process_intelligence_query is the span name for "jtbd.scenario.process_intelligence_query".
	//
	// A JTBD scenario step for executing natural language queries against process intelligence engine.
	// Kind: internal
	// Stability: development
	JtbdScenarioProcessIntelligenceQuerySpan = "jtbd.scenario.process_intelligence_query"
	// jtbd_scenario_retrofit_complexity_scoring is the span name for "jtbd.scenario.retrofit_complexity_scoring".
	//
	// A JTBD scenario step for assessing Java 26 retrofit complexity.
	// Kind: internal
	// Stability: development
	JtbdScenarioRetrofitComplexityScoringSpan = "jtbd.scenario.retrofit_complexity_scoring"
)