package versioning

import (
	"encoding/json"
	"testing"
)

// TestComputeDelta tests the delta computation between two JSON models
func TestComputeDelta(t *testing.T) {
	service := &ModelHistoryService{}

	// Previous model with 2 nodes
	previous := json.RawMessage(`{
		"nodes": [
			{"id": "task_1", "type": "task", "label": "Check"},
			{"id": "task_2", "type": "task", "label": "Process"}
		],
		"edges": [
			{"id": "edge_1", "source": "task_1", "target": "task_2"}
		]
	}`)

	// Current model with 3 nodes (added task_3) and 2 edges
	current := json.RawMessage(`{
		"nodes": [
			{"id": "task_1", "type": "task", "label": "Check"},
			{"id": "task_2", "type": "task", "label": "Process"},
			{"id": "task_3", "type": "task", "label": "Approve"}
		],
		"edges": [
			{"id": "edge_1", "source": "task_1", "target": "task_2"},
			{"id": "edge_2", "source": "task_2", "target": "task_3"}
		]
	}`)

	delta := service.computeDelta(current, previous)

	// Parse the delta patches
	var patches []map[string]interface{}
	if err := json.Unmarshal(delta, &patches); err != nil {
		t.Fatalf("Failed to parse delta: %v", err)
	}

	// Should have 2 additions: task_3 and edge_2
	if len(patches) != 2 {
		t.Errorf("Expected 2 patches, got %d", len(patches))
	}

	// Verify patches have correct structure
	for _, patch := range patches {
		op, ok := patch["op"].(string)
		if !ok || op != "add" {
			t.Errorf("Expected 'add' operation, got %v", patch["op"])
		}

		path, ok := patch["path"].(string)
		if !ok {
			t.Errorf("Expected string path, got %v", patch["path"])
		}

		value, ok := patch["value"].(map[string]interface{})
		if !ok {
			t.Errorf("Expected object value, got %v", patch["value"])
		}

		if path == "/nodes/task_3" {
			if value["id"] != "task_3" {
				t.Errorf("Expected task_3, got %v", value["id"])
			}
		} else if path == "/edges/edge_2" {
			if value["id"] != "edge_2" {
				t.Errorf("Expected edge_2, got %v", value["id"])
			}
		}
	}
}

// TestComputeDeltaRemovals tests delta computation with removed nodes/edges
func TestComputeDeltaRemovals(t *testing.T) {
	service := &ModelHistoryService{}

	// Previous model with 2 nodes
	previous := json.RawMessage(`{
		"nodes": [
			{"id": "task_1", "type": "task"},
			{"id": "task_2", "type": "task"}
		],
		"edges": [
			{"id": "edge_1", "source": "task_1", "target": "task_2"}
		]
	}`)

	// Current model with only 1 node (removed task_2)
	current := json.RawMessage(`{
		"nodes": [
			{"id": "task_1", "type": "task"}
		],
		"edges": []
	}`)

	delta := service.computeDelta(current, previous)

	var patches []map[string]interface{}
	if err := json.Unmarshal(delta, &patches); err != nil {
		t.Fatalf("Failed to parse delta: %v", err)
	}

	// Should have 2 removals: task_2 and edge_1
	if len(patches) != 2 {
		t.Errorf("Expected 2 patches, got %d", len(patches))
	}

	// Verify both are remove operations
	removeCount := 0
	for _, patch := range patches {
		if op, ok := patch["op"].(string); ok && op == "remove" {
			removeCount++
		}
	}

	if removeCount != 2 {
		t.Errorf("Expected 2 remove operations, got %d", removeCount)
	}
}

// TestComputeStructuralDiff tests structural difference computation
func TestComputeStructuralDiff(t *testing.T) {
	from := map[string]interface{}{
		"nodes": []interface{}{
			map[string]interface{}{"id": "task_1", "type": "task", "label": "Check"},
			map[string]interface{}{"id": "task_2", "type": "task", "label": "Process"},
		},
		"edges": []interface{}{
			map[string]interface{}{"id": "edge_1", "source": "task_1", "target": "task_2"},
		},
	}

	to := map[string]interface{}{
		"nodes": []interface{}{
			map[string]interface{}{"id": "task_1", "type": "task", "label": "Check"},
			map[string]interface{}{"id": "task_3", "type": "gateway", "label": "Decision"},
		},
		"edges": []interface{}{
			map[string]interface{}{"id": "edge_1", "source": "task_1", "target": "task_3"},
			map[string]interface{}{"id": "edge_2", "source": "task_3", "target": "task_1"},
		},
	}

	diff := computeStructuralDiff(from, to)

	// Should have:
	// - 1 node added (task_3)
	// - 1 node removed (task_2)
	// - 1 edge added (edge_2)
	if len(diff.NodesAdded) != 1 {
		t.Errorf("Expected 1 node added, got %d", len(diff.NodesAdded))
	}

	if len(diff.NodesRemoved) != 1 {
		t.Errorf("Expected 1 node removed, got %d", len(diff.NodesRemoved))
	}

	if len(diff.EdgesAdded) != 1 {
		t.Errorf("Expected 1 edge added, got %d", len(diff.EdgesAdded))
	}

	// Verify added node details
	if diff.NodesAdded[0].ID != "task_3" {
		t.Errorf("Expected added node ID task_3, got %s", diff.NodesAdded[0].ID)
	}

	if diff.NodesAdded[0].Type != "gateway" {
		t.Errorf("Expected added node type gateway, got %s", diff.NodesAdded[0].Type)
	}

	// Verify removed node details
	if diff.NodesRemoved[0].ID != "task_2" {
		t.Errorf("Expected removed node ID task_2, got %s", diff.NodesRemoved[0].ID)
	}
}

// TestComputeStructuralDiffEmptyModels tests with empty models
func TestComputeStructuralDiffEmptyModels(t *testing.T) {
	from := map[string]interface{}{}
	to := map[string]interface{}{}

	diff := computeStructuralDiff(from, to)

	if len(diff.NodesAdded) != 0 {
		t.Errorf("Expected 0 nodes added, got %d", len(diff.NodesAdded))
	}

	if len(diff.NodesRemoved) != 0 {
		t.Errorf("Expected 0 nodes removed, got %d", len(diff.NodesRemoved))
	}

	if len(diff.EdgesAdded) != 0 {
		t.Errorf("Expected 0 edges added, got %d", len(diff.EdgesAdded))
	}

	if len(diff.EdgesRemoved) != 0 {
		t.Errorf("Expected 0 edges removed, got %d", len(diff.EdgesRemoved))
	}
}

// TestGetStringField tests the helper function
func TestGetStringField(t *testing.T) {
	tests := []struct {
		name          string
		m             map[string]interface{}
		key           string
		defaultValue  string
		expectedValue string
	}{
		{
			name: "existing string field",
			m: map[string]interface{}{
				"type": "task",
			},
			key:           "type",
			defaultValue:  "unknown",
			expectedValue: "task",
		},
		{
			name: "missing field returns default",
			m: map[string]interface{}{
				"label": "Test",
			},
			key:           "type",
			defaultValue:  "unknown",
			expectedValue: "unknown",
		},
		{
			name: "non-string field returns default",
			m: map[string]interface{}{
				"count": 42,
			},
			key:           "count",
			defaultValue:  "0",
			expectedValue: "0",
		},
		{
			name:          "empty map",
			m:             map[string]interface{}{},
			key:           "type",
			defaultValue:  "unknown",
			expectedValue: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getStringField(tt.m, tt.key, tt.defaultValue)
			if result != tt.expectedValue {
				t.Errorf("Expected %q, got %q", tt.expectedValue, result)
			}
		})
	}
}
