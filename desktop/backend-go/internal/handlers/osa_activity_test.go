package handlers

import (
	"github.com/rhl/businessos-backend/internal/integrations/osa"
	"github.com/rhl/businessos-backend/internal/streaming"
	"testing"
)

func TestOSAToolActivityPreservesFailure(t *testing.T) {
	for _, tc := range []struct {
		phase   string
		success bool
		status  string
		kind    streaming.EventType
	}{
		{"start", true, "calling", streaming.EventTypeToolCall},
		{"end", true, "success", streaming.EventTypeToolResult},
		{"end", false, "error", streaming.EventTypeToolResult},
	} {
		thinking := false
		mapped := mapSingleEvent(osa.Event{Type: "tool_call", Data: map[string]interface{}{"name": "computer_use", "phase": tc.phase, "success": tc.success}}, &thinking)
		if mapped == nil || mapped.Type != tc.kind {
			t.Fatalf("unexpected mapping: %+v", mapped)
		}
		call := mapped.Data.(streaming.ToolCallEvent)
		if call.ToolName != "computer_use" || call.Status != tc.status {
			t.Fatalf("unexpected tool activity: %+v", call)
		}
	}
}

func TestOSAExposesPermissionWaitWithoutExposingDiagnosticContent(t *testing.T) {
	thinking := false
	mapped := mapSingleEvent(osa.Event{Type: "system_event", Data: map[string]interface{}{"event": "permission_required", "content": "private diagnostic text"}}, &thinking)
	if mapped == nil || mapped.Type != streaming.EventTypeThinkingChunk {
		t.Fatal("permission wait was hidden")
	}
	status := mapped.Data.(streaming.ThinkingStep)
	if status.Step != "status" || status.Content != "Waiting for a tool permission decision" {
		t.Fatalf("unexpected status: %+v", status)
	}
	if mapSingleEvent(osa.Event{Type: "system_event", Data: map[string]interface{}{"event": "diagnostic", "content": "private diagnostic text"}}, &thinking) != nil {
		t.Fatal("diagnostic text was exposed")
	}
}
