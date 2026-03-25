package semconv

const (
	// chatmangpt_session_track is the span name for "chatmangpt.session.track".
	//
	// Tracks a ChatmanGPT session — lifecycle from start to end with token and turn accounting.
	// Kind: internal
	// Stability: development
	ChatmangptSessionTrackSpan = "chatmangpt.session.track"
)