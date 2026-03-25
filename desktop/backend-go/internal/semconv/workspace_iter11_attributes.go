// Code generated from semconv/model/workspace/registry.yaml. DO NOT EDIT.
// Wave 9 iteration 11

package semconv

import "go.opentelemetry.io/otel/attribute"

// Workspace tool category and context window attributes (iter11).

const (
	// WorkspaceToolCategoryKey is the OTel attribute key for workspace.tool.category.
	// Functional category of the tool active in the workspace.
	WorkspaceToolCategoryKey = attribute.Key("workspace.tool.category")
	// WorkspaceContextWindowSizeKey is the OTel attribute key for workspace.context.window_size.
	// Maximum context window size in tokens for the current workspace model.
	WorkspaceContextWindowSizeKey = attribute.Key("workspace.context.window_size")
)

// WorkspaceToolCategory returns an attribute KeyValue for workspace.tool.category.
func WorkspaceToolCategory(val string) attribute.KeyValue {
	return WorkspaceToolCategoryKey.String(val)
}

// WorkspaceContextWindowSize returns an attribute KeyValue for workspace.context.window_size.
func WorkspaceContextWindowSize(val int) attribute.KeyValue {
	return WorkspaceContextWindowSizeKey.Int(val)
}
