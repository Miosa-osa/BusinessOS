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
	// board_structural_escalation is the span name for "board.structural_escalation".
	//
	// Board escalation emitted for a structural (Conway) violation
	// Kind: internal
	// Stability: development
	BoardStructuralEscalationSpan = "board.structural_escalation"
)