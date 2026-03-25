// Code generated from semconv/model/workspace/registry.yaml. DO NOT EDIT.
// Wave 9 iteration 8

package semconv

import "go.opentelemetry.io/otel/attribute"

// Workspace Attributes

const (
	// WorkspaceSessionIDKey is the OTel attribute key for workspace.session.id.
	// Unique identifier for the workspace session.
	WorkspaceSessionIDKey = attribute.Key("workspace.session.id")
	// WorkspaceContextSizeKey is the OTel attribute key for workspace.context.size.
	// Number of tokens or items in the current workspace context.
	WorkspaceContextSizeKey = attribute.Key("workspace.context.size")
	// WorkspaceToolNameKey is the OTel attribute key for workspace.tool.name.
	// Name of the tool currently active in the workspace.
	WorkspaceToolNameKey = attribute.Key("workspace.tool.name")
	// WorkspaceToolCountKey is the OTel attribute key for workspace.tool.count.
	// Number of tools available in the workspace.
	WorkspaceToolCountKey = attribute.Key("workspace.tool.count")
	// WorkspaceAgentRoleKey is the OTel attribute key for workspace.agent.role.
	// Role of the agent operating in this workspace session.
	WorkspaceAgentRoleKey = attribute.Key("workspace.agent.role")
	// WorkspacePhaseKey is the OTel attribute key for workspace.phase.
	// Current lifecycle phase of the workspace session.
	WorkspacePhaseKey = attribute.Key("workspace.phase")
)

// WorkspaceSessionID returns an attribute KeyValue for workspace.session.id.
func WorkspaceSessionID(val string) attribute.KeyValue { return WorkspaceSessionIDKey.String(val) }

// WorkspaceContextSize returns an attribute KeyValue for workspace.context.size.
func WorkspaceContextSize(val int) attribute.KeyValue { return WorkspaceContextSizeKey.Int(val) }

// WorkspaceToolName returns an attribute KeyValue for workspace.tool.name.
func WorkspaceToolName(val string) attribute.KeyValue { return WorkspaceToolNameKey.String(val) }

// WorkspaceToolCount returns an attribute KeyValue for workspace.tool.count.
func WorkspaceToolCount(val int) attribute.KeyValue { return WorkspaceToolCountKey.Int(val) }

// WorkspaceAgentRole returns an attribute KeyValue for workspace.agent.role.
func WorkspaceAgentRole(val string) attribute.KeyValue { return WorkspaceAgentRoleKey.String(val) }

// WorkspacePhase returns an attribute KeyValue for workspace.phase.
func WorkspacePhase(val string) attribute.KeyValue { return WorkspacePhaseKey.String(val) }

// WorkspaceAgentRole* constants are the known enum values for workspace.agent.role.
const (
	WorkspaceAgentRolePlanner     = "planner"
	WorkspaceAgentRoleExecutor    = "executor"
	WorkspaceAgentRoleReviewer    = "reviewer"
	WorkspaceAgentRoleCoordinator = "coordinator"
	WorkspaceAgentRoleResearcher  = "researcher"
)

// WorkspacePhase* constants are the known enum values for workspace.phase.
const (
	WorkspacePhaseStartup  = "startup"
	WorkspacePhaseActive   = "active"
	WorkspacePhaseIdle     = "idle"
	WorkspacePhaseShutdown = "shutdown"
)
