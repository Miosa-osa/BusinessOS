package semconv

const (
	// board_briefing_render is the span name for "board.briefing_render".
	//
	// Board chair briefing rendered from L3 intelligence data
	// Kind: internal
	// Stability: development
	BoardBriefingRenderSpan = "board.briefing_render"
	// board_conway_check is the span name for "board.conway_check".
	//
	// Conway's Law violation check for a department process
	// Kind: internal
	// Stability: development
	BoardConwayCheckSpan = "board.conway_check"
	// board_conway_check_summary is the span name for "board.conway_check_summary".
	//
	// Periodic Conway + Little's Law monitoring check summary
	// Kind: internal
	// Stability: development
	BoardConwayCheckSummarySpan = "board.conway_check_summary"
	// board_kpi_compute is the span name for "board.kpi_compute".
	//
	// Board KPIs computed from process mining event log
	// Kind: internal
	// Stability: development
	BoardKpiComputeSpan = "board.kpi_compute"
	// board_l0_sync is the span name for "board.l0_sync".
	//
	// Periodic L0 sync — exports BusinessOS cases and handoffs to Oxigraph as RDF facts via bos CLI.
	// Kind: internal
	// Stability: development
	BoardL0SyncSpan = "board.l0_sync"
	// board_structural_escalation is the span name for "board.structural_escalation".
	//
	// Board escalation emitted for a structural (Conway) violation
	// Kind: internal
	// Stability: development
	BoardStructuralEscalationSpan = "board.structural_escalation"
)