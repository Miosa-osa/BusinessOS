package semconv

const (
	// canopy_adapter_call is the span name for "canopy.adapter_call".
	//
	// Canopy adapter invocation — calling an external service via a Canopy adapter.
	// Kind: client
	// Stability: development
	CanopyAdapterCallSpan = "canopy.adapter_call"
	// canopy_broadcast is the span name for "canopy.broadcast".
	//
	// Broadcast of a signal or command to all connected agents.
	// Kind: producer
	// Stability: development
	CanopyBroadcastSpan = "canopy.broadcast"
	// canopy_command is the span name for "canopy.command".
	//
	// Command dispatch through the Canopy workspace protocol.
	// Kind: producer
	// Stability: development
	CanopyCommandSpan = "canopy.command"
	// canopy_heartbeat is the span name for "canopy.heartbeat".
	//
	// Canopy heartbeat dispatch — periodic health signal sent to connected services.
	// Kind: internal
	// Stability: development
	CanopyHeartbeatSpan = "canopy.heartbeat"
	// canopy_heartbeat_probe is the span name for "canopy.heartbeat.probe".
	//
	// Individual heartbeat probe — one RTT measurement to a single OSA node.
	// Kind: internal
	// Stability: development
	CanopyHeartbeatProbeSpan = "canopy.heartbeat.probe"
	// canopy_session_create is the span name for "canopy.session.create".
	//
	// Canopy workspace session creation — initializing a new collaboration session.
	// Kind: server
	// Stability: development
	CanopySessionCreateSpan = "canopy.session.create"
	// canopy_snapshot_create is the span name for "canopy.snapshot.create".
	//
	// Creating a point-in-time snapshot of the canopy workspace state.
	// Kind: internal
	// Stability: development
	CanopySnapshotCreateSpan = "canopy.snapshot.create"
	// canopy_workspace_reconcile is the span name for "canopy.workspace.reconcile".
	//
	// Reconciling workspace state between peers — resolving conflicts and applying updates.
	// Kind: internal
	// Stability: development
	CanopyWorkspaceReconcileSpan = "canopy.workspace.reconcile"
	// canopy_workspace_sync is the span name for "canopy.workspace.sync".
	//
	// Synchronization of workspace state across connected agents.
	// Kind: internal
	// Stability: development
	CanopyWorkspaceSyncSpan = "canopy.workspace.sync"
)