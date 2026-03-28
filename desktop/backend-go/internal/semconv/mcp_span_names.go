package semconv

const (
	// mcp_call is the span name for "mcp.call".
	//
	// An MCP tool invocation — request from an agent to execute a named tool via MCP protocol.
	// Kind: client
	// Stability: development
	McpCallSpan = "mcp.call"
	// mcp_connection_establish is the span name for "mcp.connection.establish".
	//
	// MCP client-server connection establishment — transport negotiation and capability exchange.
	// Kind: client
	// Stability: development
	McpConnectionEstablishSpan = "mcp.connection.establish"
	// mcp_connection_pool_acquire is the span name for "mcp.connection.pool.acquire".
	//
	// Acquiring a connection from the MCP connection pool for use in a client-server interaction.
	// Kind: internal
	// Stability: development
	McpConnectionPoolAcquireSpan = "mcp.connection.pool.acquire"
	// mcp_registry_discover is the span name for "mcp.registry.discover".
	//
	// MCP tool discovery — listing available tools from a connected server.
	// Kind: client
	// Stability: development
	McpRegistryDiscoverSpan = "mcp.registry.discover"
	// mcp_resource_read is the span name for "mcp.resource.read".
	//
	// Reading an MCP resource — fetching content from a resource URI exposed by an MCP server.
	// Kind: client
	// Stability: development
	McpResourceReadSpan = "mcp.resource.read"
	// mcp_server_health_check is the span name for "mcp.server.health_check".
	//
	// Health check of an MCP server — verifying tool availability and server responsiveness.
	// Kind: internal
	// Stability: development
	McpServerHealthCheckSpan = "mcp.server.health_check"
	// mcp_server_metrics_collect is the span name for "mcp.server.metrics.collect".
	//
	// Collecting aggregated metrics from an MCP server instance.
	// Kind: internal
	// Stability: development
	McpServerMetricsCollectSpan = "mcp.server.metrics.collect"
	// mcp_session_create is the span name for "mcp.session.create".
	//
	// New MCP session allocation by the StreamableHttpService server.
	// Kind: server
	// Stability: development
	McpSessionCreateSpan = "mcp.session.create"
	// mcp_tool_analytics_record is the span name for "mcp.tool.analytics.record".
	//
	// MCP tool analytics recording — capturing tool usage statistics for performance monitoring and capacity planning.
	// Kind: internal
	// Stability: development
	McpToolAnalyticsRecordSpan = "mcp.tool.analytics.record"
	// mcp_tool_cache_lookup is the span name for "mcp.tool.cache.lookup".
	//
	// MCP tool cache lookup — checking response cache before executing tool.
	// Kind: internal
	// Stability: development
	McpToolCacheLookupSpan = "mcp.tool.cache.lookup"
	// mcp_tool_compose is the span name for "mcp.tool.compose".
	//
	// Composition of multiple MCP tools into a chain — sequential, parallel, or fallback execution.
	// Kind: internal
	// Stability: development
	McpToolComposeSpan = "mcp.tool.compose"
	// mcp_tool_deprecate is the span name for "mcp.tool.deprecate".
	//
	// MCP tool deprecation lifecycle event — marking a tool as deprecated and scheduling its removal.
	// Kind: internal
	// Stability: development
	McpToolDeprecateSpan = "mcp.tool.deprecate"
	// mcp_tool_retry is the span name for "mcp.tool.retry".
	//
	// A retry attempt for a previously failed MCP tool execution.
	// Kind: client
	// Stability: development
	McpToolRetrySpan = "mcp.tool.retry"
	// mcp_tool_timeout is the span name for "mcp.tool.timeout".
	//
	// MCP tool execution timed out — tool did not respond within the configured budget.
	// Kind: client
	// Stability: development
	McpToolTimeoutSpan = "mcp.tool.timeout"
	// mcp_tool_validate is the span name for "mcp.tool.validate".
	//
	// Validating MCP tool input/output schema before execution.
	// Kind: internal
	// Stability: development
	McpToolValidateSpan = "mcp.tool.validate"
	// mcp_tool_version_check is the span name for "mcp.tool.version_check".
	//
	// Version compatibility check for an MCP tool — validates client version against server tool version.
	// Kind: internal
	// Stability: development
	McpToolVersionCheckSpan = "mcp.tool.version_check"
	// mcp_tool_execute is the span name for "mcp.tool_execute".
	//
	// Server-side execution of an MCP tool — the handler running the tool logic.
	// Kind: server
	// Stability: development
	McpToolExecuteSpan = "mcp.tool_execute"
	// mcp_transport_connect is the span name for "mcp.transport.connect".
	//
	// Establishment of an MCP transport connection — initial handshake and protocol negotiation.
	// Kind: client
	// Stability: development
	McpTransportConnectSpan = "mcp.transport.connect"
)