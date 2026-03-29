package sync

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockLogger implements Logger interface for testing
type MockLogger struct{}

func (m *MockLogger) Info(msg string, args ...any)  {}
func (m *MockLogger) Warn(msg string, args ...any)  {}
func (m *MockLogger) Error(msg string, args ...any) {}

func TestDetectWorkspaceConflict_NoConflict_RemoteClearlyNewer(t *testing.T) {
	cd := &ConflictDetector{logger: &MockLogger{}}

	workspaceID := uuid.New()
	now := time.Now()

	local := &Workspace{
		ID:        workspaceID,
		UserID:    uuid.New(),
		Name:      "Test Workspace",
		Mode:      "2d",
		UpdatedAt: now,
	}

	remote := &Workspace{
		ID:        workspaceID,
		UserID:    local.UserID,
		Name:      "Test Workspace Updated",
		Mode:      "2d",
		UpdatedAt: now.Add(10 * time.Second), // 10 seconds newer
	}

	conflict, err := cd.DetectWorkspaceConflict(nil, local, remote)

	require.NoError(t, err)
	assert.Nil(t, conflict, "No conflict should be detected when remote is >5 seconds newer")
}

func TestDetectWorkspaceConflict_NoConflict_LocalClearlyNewer(t *testing.T) {
	cd := &ConflictDetector{logger: &MockLogger{}}

	workspaceID := uuid.New()
	now := time.Now()

	local := &Workspace{
		ID:        workspaceID,
		UserID:    uuid.New(),
		Name:      "Test Workspace",
		Mode:      "2d",
		UpdatedAt: now,
	}

	remote := &Workspace{
		ID:        workspaceID,
		UserID:    local.UserID,
		Name:      "Test Workspace Old",
		Mode:      "2d",
		UpdatedAt: now.Add(-10 * time.Second), // 10 seconds older
	}

	conflict, err := cd.DetectWorkspaceConflict(nil, local, remote)

	require.NoError(t, err)
	assert.Nil(t, conflict, "No conflict should be detected when local is >5 seconds newer")
}

func TestDetectWorkspaceConflict_NoConflict_IdenticalData(t *testing.T) {
	cd := &ConflictDetector{logger: &MockLogger{}}

	workspaceID := uuid.New()
	now := time.Now()

	local := &Workspace{
		ID:        workspaceID,
		UserID:    uuid.New(),
		Name:      "Test Workspace",
		Mode:      "2d",
		Layout:    json.RawMessage(`{"x": 10, "y": 20}`),
		UpdatedAt: now,
	}

	remote := &Workspace{
		ID:        workspaceID,
		UserID:    local.UserID,
		Name:      "Test Workspace",
		Mode:      "2d",
		Layout:    json.RawMessage(`{"x": 10, "y": 20}`),
		UpdatedAt: now.Add(2 * time.Second), // Within threshold
	}

	conflict, err := cd.DetectWorkspaceConflict(nil, local, remote)

	require.NoError(t, err)
	assert.Nil(t, conflict, "No conflict should be detected when data is identical")
}

func TestDetectWorkspaceConflict_ConflictDetected_NameChanged(t *testing.T) {
	cd := &ConflictDetector{logger: &MockLogger{}}

	workspaceID := uuid.New()
	now := time.Now()

	local := &Workspace{
		ID:        workspaceID,
		UserID:    uuid.New(),
		Name:      "Local Name",
		Mode:      "2d",
		UpdatedAt: now,
	}

	remote := &Workspace{
		ID:        workspaceID,
		UserID:    local.UserID,
		Name:      "Remote Name",
		Mode:      "2d",
		UpdatedAt: now.Add(2 * time.Second), // Within 5 second threshold
	}

	conflict, err := cd.DetectWorkspaceConflict(nil, local, remote)

	require.NoError(t, err)
	require.NotNil(t, conflict, "Conflict should be detected")
	assert.Equal(t, "workspace", conflict.EntityType)
	assert.Equal(t, workspaceID, conflict.EntityID)
	assert.Contains(t, conflict.ConflictFields, "name")
}

func TestDetectWorkspaceConflict_ConflictDetected_MultipleFields(t *testing.T) {
	cd := &ConflictDetector{logger: &MockLogger{}}

	workspaceID := uuid.New()
	now := time.Now()

	local := &Workspace{
		ID:            workspaceID,
		UserID:        uuid.New(),
		Name:          "Local Name",
		Mode:          "2d",
		Layout:        json.RawMessage(`{"x": 10}`),
		ActiveModules: []uuid.UUID{uuid.New()},
		Settings:      json.RawMessage(`{"theme": "dark"}`),
		UpdatedAt:     now,
	}

	remote := &Workspace{
		ID:            workspaceID,
		UserID:        local.UserID,
		Name:          "Remote Name",
		Mode:          "3d",
		Layout:        json.RawMessage(`{"y": 20}`),
		ActiveModules: []uuid.UUID{uuid.New()},
		Settings:      json.RawMessage(`{"theme": "light"}`),
		UpdatedAt:     now.Add(1 * time.Second),
	}

	conflict, err := cd.DetectWorkspaceConflict(nil, local, remote)

	require.NoError(t, err)
	require.NotNil(t, conflict)
	assert.Contains(t, conflict.ConflictFields, "name")
	assert.Contains(t, conflict.ConflictFields, "mode")
	assert.Contains(t, conflict.ConflictFields, "layout")
	assert.Contains(t, conflict.ConflictFields, "active_modules")
	assert.Contains(t, conflict.ConflictFields, "settings")
}

func TestResolveConflict_TimestampBased_RemoteWins(t *testing.T) {
	cd := &ConflictDetector{logger: &MockLogger{}}

	workspaceID := uuid.New()
	now := time.Now()

	local := &Workspace{
		ID:        workspaceID,
		Name:      "Local",
		UpdatedAt: now,
	}

	remote := &Workspace{
		ID:        workspaceID,
		Name:      "Remote",
		UpdatedAt: now.Add(10 * time.Second),
	}

	localJSON, _ := json.Marshal(local)
	remoteJSON, _ := json.Marshal(remote)

	conflict := &Conflict{
		EntityType:      "workspace",
		EntityID:        workspaceID,
		LocalData:       localJSON,
		RemoteData:      remoteJSON,
		LocalUpdatedAt:  now,
		RemoteUpdatedAt: now.Add(10 * time.Second),
		ConflictFields:  []string{"name"},
	}

	resolution, err := cd.ResolveConflict(nil, conflict)

	require.NoError(t, err)
	require.NotNil(t, resolution)
	assert.Equal(t, ResolutionTimestampBased, resolution.Strategy)
	assert.Nil(t, resolution.ResolvedBy, "Should be automatic resolution")

	var resolved Workspace
	err = json.Unmarshal(resolution.ResolvedData, &resolved)
	require.NoError(t, err)
	assert.Equal(t, "Remote", resolved.Name, "Remote should win")
}

func TestResolveConflict_TimestampBased_LocalWins(t *testing.T) {
	cd := &ConflictDetector{logger: &MockLogger{}}

	workspaceID := uuid.New()
	now := time.Now()

	local := &Workspace{
		ID:        workspaceID,
		Name:      "Local",
		UpdatedAt: now,
	}

	remote := &Workspace{
		ID:        workspaceID,
		Name:      "Remote",
		UpdatedAt: now.Add(-10 * time.Second),
	}

	localJSON, _ := json.Marshal(local)
	remoteJSON, _ := json.Marshal(remote)

	conflict := &Conflict{
		EntityType:      "workspace",
		EntityID:        workspaceID,
		LocalData:       localJSON,
		RemoteData:      remoteJSON,
		LocalUpdatedAt:  now,
		RemoteUpdatedAt: now.Add(-10 * time.Second),
		ConflictFields:  []string{"name"},
	}

	resolution, err := cd.ResolveConflict(nil, conflict)

	require.NoError(t, err)
	require.NotNil(t, resolution)
	assert.Equal(t, ResolutionTimestampBased, resolution.Strategy)

	var resolved Workspace
	err = json.Unmarshal(resolution.ResolvedData, &resolved)
	require.NoError(t, err)
	assert.Equal(t, "Local", resolved.Name, "Local should win")
}

func TestResolveConflict_FieldLevelMerge_NonCriticalFields(t *testing.T) {
	cd := &ConflictDetector{logger: &MockLogger{}}

	workspaceID := uuid.New()
	userID := uuid.New()
	now := time.Now()

	local := &Workspace{
		ID:       workspaceID,
		UserID:   userID,
		Name:     "Workspace",
		Mode:     "2d",
		Layout:   json.RawMessage(`{"x": 10}`),
		Settings: json.RawMessage(`{"theme": "dark"}`),
		UpdatedAt: now,
	}

	remote := &Workspace{
		ID:       workspaceID,
		UserID:   userID,
		Name:     "Workspace",
		Mode:     "2d",
		Layout:   json.RawMessage(`{"y": 20}`),
		Settings: json.RawMessage(`{"lang": "en"}`),
		UpdatedAt: now.Add(2 * time.Second),
	}

	localJSON, _ := json.Marshal(local)
	remoteJSON, _ := json.Marshal(remote)

	conflict := &Conflict{
		EntityType:      "workspace",
		EntityID:        workspaceID,
		LocalData:       localJSON,
		RemoteData:      remoteJSON,
		LocalUpdatedAt:  now,
		RemoteUpdatedAt: now.Add(2 * time.Second),
		ConflictFields:  []string{"layout", "settings"}, // Only non-critical fields
	}

	resolution, err := cd.ResolveConflict(nil, conflict)

	require.NoError(t, err)
	require.NotNil(t, resolution)
	assert.Equal(t, ResolutionFieldLevelMerge, resolution.Strategy)
	assert.Nil(t, resolution.ResolvedBy, "Should be automatic resolution")

	var resolved Workspace
	err = json.Unmarshal(resolution.ResolvedData, &resolved)
	require.NoError(t, err)

	// Check that JSON fields were merged
	var resolvedLayout map[string]interface{}
	json.Unmarshal(resolved.Layout, &resolvedLayout)
	assert.Contains(t, resolvedLayout, "x")
	assert.Contains(t, resolvedLayout, "y")
}

func TestResolveConflict_ManualReview_CriticalFields(t *testing.T) {
	cd := &ConflictDetector{logger: &MockLogger{}}

	workspaceID := uuid.New()
	now := time.Now()

	local := &Workspace{
		ID:        workspaceID,
		Name:      "Local Name",
		Mode:      "2d",
		UpdatedAt: now,
	}

	remote := &Workspace{
		ID:        workspaceID,
		Name:      "Remote Name",
		Mode:      "3d",
		UpdatedAt: now.Add(2 * time.Second),
	}

	localJSON, _ := json.Marshal(local)
	remoteJSON, _ := json.Marshal(remote)

	conflict := &Conflict{
		EntityType:      "workspace",
		EntityID:        workspaceID,
		LocalData:       localJSON,
		RemoteData:      remoteJSON,
		LocalUpdatedAt:  now,
		RemoteUpdatedAt: now.Add(2 * time.Second),
		ConflictFields:  []string{"name", "mode"}, // Critical fields
	}

	resolution, err := cd.ResolveConflict(nil, conflict)

	require.NoError(t, err)
	require.NotNil(t, resolution)
	assert.Equal(t, ResolutionManualReview, resolution.Strategy)
	assert.Nil(t, resolution.ResolvedData, "Should have no automatic resolution")
	assert.Contains(t, resolution.Reasoning, "Critical fields")
}

func TestCanAutoMerge(t *testing.T) {
	cd := &ConflictDetector{logger: &MockLogger{}}

	tests := []struct {
		name           string
		conflictFields []string
		canMerge       bool
	}{
		{
			name:           "only non-critical fields",
			conflictFields: []string{"layout", "settings"},
			canMerge:       true,
		},
		{
			name:           "contains name (critical)",
			conflictFields: []string{"name", "layout"},
			canMerge:       false,
		},
		{
			name:           "contains mode (critical)",
			conflictFields: []string{"mode", "settings"},
			canMerge:       false,
		},
		{
			name:           "both critical fields",
			conflictFields: []string{"name", "mode"},
			canMerge:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cd.canAutoMerge(tt.conflictFields)
			assert.Equal(t, tt.canMerge, result)
		})
	}
}

func TestMergeJSON(t *testing.T) {
	cd := &ConflictDetector{logger: &MockLogger{}}

	local := json.RawMessage(`{"a": 1, "b": 2}`)
	remote := json.RawMessage(`{"b": 3, "c": 4}`)

	// Test useLocal = true
	merged := cd.mergeJSON(local, remote, true)
	var mergedMap map[string]interface{}
	json.Unmarshal(merged, &mergedMap)

	assert.Equal(t, float64(1), mergedMap["a"], "Should keep local 'a'")
	assert.Equal(t, float64(2), mergedMap["b"], "Should keep local 'b'")
	assert.Equal(t, float64(4), mergedMap["c"], "Should add remote 'c'")

	// Test useLocal = false
	merged = cd.mergeJSON(local, remote, false)
	json.Unmarshal(merged, &mergedMap)

	assert.Equal(t, float64(1), mergedMap["a"], "Should keep local 'a'")
	assert.Equal(t, float64(3), mergedMap["b"], "Should use remote 'b'")
	assert.Equal(t, float64(4), mergedMap["c"], "Should add remote 'c'")
}

func TestMergeUUIDSlices(t *testing.T) {
	id1 := uuid.New()
	id2 := uuid.New()
	id3 := uuid.New()

	a := []uuid.UUID{id1, id2}
	b := []uuid.UUID{id2, id3}

	merged := mergeUUIDSlices(a, b)

	assert.Len(t, merged, 3, "Should have 3 unique UUIDs")
	assert.Contains(t, merged, id1)
	assert.Contains(t, merged, id2)
	assert.Contains(t, merged, id3)
}

func TestJSONEqual(t *testing.T) {
	tests := []struct {
		name  string
		a     json.RawMessage
		b     json.RawMessage
		equal bool
	}{
		{
			name:  "identical objects",
			a:     json.RawMessage(`{"x": 10, "y": 20}`),
			b:     json.RawMessage(`{"x": 10, "y": 20}`),
			equal: true,
		},
		{
			name:  "same content different order",
			a:     json.RawMessage(`{"x": 10, "y": 20}`),
			b:     json.RawMessage(`{"y": 20, "x": 10}`),
			equal: true,
		},
		{
			name:  "different values",
			a:     json.RawMessage(`{"x": 10}`),
			b:     json.RawMessage(`{"x": 20}`),
			equal: false,
		},
		{
			name:  "empty objects",
			a:     json.RawMessage(`{}`),
			b:     json.RawMessage(`{}`),
			equal: true,
		},
		{
			name:  "both nil/empty",
			a:     json.RawMessage{},
			b:     json.RawMessage{},
			equal: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := jsonEqual(tt.a, tt.b)
			assert.Equal(t, tt.equal, result)
		})
	}
}

func TestUUIDSliceEqual(t *testing.T) {
	id1 := uuid.New()
	id2 := uuid.New()
	id3 := uuid.New()

	tests := []struct {
		name  string
		a     []uuid.UUID
		b     []uuid.UUID
		equal bool
	}{
		{
			name:  "identical slices",
			a:     []uuid.UUID{id1, id2},
			b:     []uuid.UUID{id1, id2},
			equal: true,
		},
		{
			name:  "same IDs different order",
			a:     []uuid.UUID{id1, id2},
			b:     []uuid.UUID{id2, id1},
			equal: true,
		},
		{
			name:  "different IDs",
			a:     []uuid.UUID{id1, id2},
			b:     []uuid.UUID{id1, id3},
			equal: false,
		},
		{
			name:  "different lengths",
			a:     []uuid.UUID{id1},
			b:     []uuid.UUID{id1, id2},
			equal: false,
		},
		{
			name:  "both empty",
			a:     []uuid.UUID{},
			b:     []uuid.UUID{},
			equal: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := uuidSliceEqual(tt.a, tt.b)
			assert.Equal(t, tt.equal, result)
		})
	}
}
