package yawlv6_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rhl/businessos-backend/internal/integrations/yawlv6"
)

func TestHealth_WhenEngineUnreachable(t *testing.T) {
	// Point to a port nobody is listening on so Health() must return an error.
	t.Setenv("YAWLV6_URL", "http://localhost:19999")
	c := yawlv6.NewClient()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := c.Health(ctx)
	assert.Error(t, err, "Health should return error when engine is unreachable")
}

func TestBuildSequenceSpec_ContainsAllTasks(t *testing.T) {
	xml := yawlv6.BuildSequenceSpec([]string{"step_a", "step_b", "step_c"})

	assert.Contains(t, xml, `id="step_a"`)
	assert.Contains(t, xml, `id="step_b"`)
	assert.Contains(t, xml, `id="step_c"`)
	assert.Contains(t, xml, "specificationSet")
	assert.Contains(t, xml, `join code="xor"`)
}

func TestBuildSequenceSpec_EmptyTasks_ReturnsEmpty(t *testing.T) {
	xml := yawlv6.BuildSequenceSpec([]string{})
	assert.Empty(t, xml)
}

func TestBuildSequenceSpec_ChainOrder(t *testing.T) {
	xml := yawlv6.BuildSequenceSpec([]string{"a", "b"})

	// "a" must flow into "b", "b" must flow into OutputCondition.
	aIdx := strings.Index(xml, `id="a"`)
	bIdx := strings.Index(xml, `id="b"`)
	require.True(t, aIdx >= 0 && bIdx >= 0)

	assert.Contains(t, xml, `nextElementRef id="b"`)
	assert.Contains(t, xml, `nextElementRef id="OutputCondition"`)
}

func TestBuildParallelSplitSpec_HasAndSplit(t *testing.T) {
	xml := yawlv6.BuildParallelSplitSpec("start", []string{"branch_a", "branch_b"})

	assert.Contains(t, xml, `split code="and"`)
	assert.Contains(t, xml, `id="branch_a"`)
	assert.Contains(t, xml, `id="branch_b"`)
	require.True(t, strings.Contains(xml, "specificationSet"))
	assert.Contains(t, xml, "OSA_ParallelSplit")
}

func TestBuildParallelSplitSpec_TriggerFlowsToBothBranches(t *testing.T) {
	xml := yawlv6.BuildParallelSplitSpec("trigger", []string{"left", "right"})

	assert.Contains(t, xml, `nextElementRef id="left"`)
	assert.Contains(t, xml, `nextElementRef id="right"`)
}

func TestConformanceResult_JSONRoundTrip(t *testing.T) {
	raw := []byte(`{"fitness":0.95,"violations":["v1"],"is_sound":true}`)

	var result yawlv6.ConformanceResult
	err := json.Unmarshal(raw, &result)
	require.NoError(t, err)

	assert.Equal(t, 0.95, result.Fitness)
	assert.Equal(t, []string{"v1"}, result.Violations)
	assert.True(t, result.IsSound)
}

// TestHealth_ReachableServer_ReturnsNoError tests that Health() succeeds
// when the YAWL engine responds with 200 at /health.jsp.
func TestHealth_ReachableServer_ReturnsNoError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/health.jsp", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"running","version":"6.0","engine_running":true}`))
	}))
	defer srv.Close()

	t.Setenv("YAWLV6_URL", srv.URL)
	c := yawlv6.NewClient()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := c.Health(ctx)
	assert.NoError(t, err, "Health should return nil error when engine responds 200")
}

// TestCheckConformance_ValidSpec_ReturnsResult tests that CheckConformance
// successfully parses the YAWL engine response and returns the result.
func TestCheckConformance_ValidSpec_ReturnsResult(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/process-mining/conformance", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"fitness":0.95,"violations":[],"is_sound":true}`))
	}))
	defer srv.Close()

	t.Setenv("YAWLV6_URL", srv.URL)
	c := yawlv6.NewClient()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	result, err := c.CheckConformance(ctx, "<specificationSet/>", []byte(`[]`))
	require.NoError(t, err)
	assert.Equal(t, 0.95, result.Fitness)
	assert.True(t, result.IsSound)
	assert.Empty(t, result.Violations)
}

// TestCheckConformance_ServerError_ReturnsError tests that CheckConformance
// returns an error when the server returns a non-JSON response body.
// (CheckConformance does not check status code; only json.Unmarshal failure triggers error.)
func TestCheckConformance_ServerError_ReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
	}))
	defer srv.Close()

	t.Setenv("YAWLV6_URL", srv.URL)
	c := yawlv6.NewClient()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := c.CheckConformance(ctx, "<specificationSet/>", []byte(`[]`))
	assert.Error(t, err, "CheckConformance should return error when server returns non-JSON body")
}

// TestLoadSpec_ValidFile_ReturnsContent tests that LoadSpec successfully reads
// a YAWL spec XML file from the local directory structure and returns its content.
func TestLoadSpec_ValidFile_ReturnsContent(t *testing.T) {
	// Create the directory structure LoadSpec expects.
	specsDir := t.TempDir()
	catDir := specsDir + "/wcp-patterns/basic"
	err := os.MkdirAll(catDir, 0o755)
	require.NoError(t, err)

	content := `<?xml version="1.0"?><specificationSet><specification uri="WCP01"/></specificationSet>`
	err = os.WriteFile(catDir+"/WCP01_sequence.xml", []byte(content), 0o644)
	require.NoError(t, err)

	t.Setenv("YAWLV6_SPECS_PATH", specsDir)
	c := yawlv6.NewClient()

	result, err := c.LoadSpec("WCP-1")
	require.NoError(t, err)
	assert.Contains(t, result, "<specificationSet>")
}

// TestLoadSpec_MissingFile_ReturnsError tests that LoadSpec returns an error
// when the spec file does not exist in the expected directory structure.
func TestLoadSpec_MissingFile_ReturnsError(t *testing.T) {
	// Point to an empty specs directory — no wcp-patterns subdirs exist.
	specsDir := t.TempDir()
	t.Setenv("YAWLV6_SPECS_PATH", specsDir)
	c := yawlv6.NewClient()

	_, err := c.LoadSpec("WCP-99")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// TestListPatterns_ReturnsAllSpecs verifies that ListPatterns discovers all XML files
// in wcp-patterns/* subdirectories, parsing ID and name from filenames.
func TestListPatterns_ReturnsAllSpecs(t *testing.T) {
	specsDir := t.TempDir()
	catDir := specsDir + "/wcp-patterns/basic"
	require.NoError(t, os.MkdirAll(catDir, 0o755))
	require.NoError(t, os.WriteFile(catDir+"/WCP01_Sequence.xml", []byte("<spec/>"), 0o644))
	require.NoError(t, os.WriteFile(catDir+"/WCP02_ParallelSplit.xml", []byte("<spec/>"), 0o644))

	t.Setenv("YAWLV6_SPECS_PATH", specsDir)
	c := yawlv6.NewClient()

	entries, err := c.ListPatterns()
	require.NoError(t, err)
	require.Len(t, entries, 2)
	assert.Equal(t, "WCP-1", entries[0].ID)
	assert.Equal(t, "Sequence", entries[0].Name)
	assert.Equal(t, "basic", entries[0].Category)
	assert.Equal(t, "WCP-2", entries[1].ID)
}

// TestListPatterns_EmptyDir_ReturnsEmpty verifies that ListPatterns returns an empty
// slice when no wcp-patterns directories exist.
func TestListPatterns_EmptyDir_ReturnsEmpty(t *testing.T) {
	specsDir := t.TempDir()
	t.Setenv("YAWLV6_SPECS_PATH", specsDir)
	c := yawlv6.NewClient()

	entries, err := c.ListPatterns()
	require.NoError(t, err)
	assert.Empty(t, entries)
}

// TestLoadRealData_ValidFile_ReturnsXML verifies that LoadRealData reads a real-data spec.
func TestLoadRealData_ValidFile_ReturnsXML(t *testing.T) {
	specsDir := t.TempDir()
	realDir := specsDir + "/real-data"
	require.NoError(t, os.MkdirAll(realDir, 0o755))
	content := `<?xml version="1.0"?><specificationSet><specification uri="RepairProcess"/></specificationSet>`
	require.NoError(t, os.WriteFile(realDir+"/RepairProcess.yawl.xml", []byte(content), 0o644))

	t.Setenv("YAWLV6_SPECS_PATH", specsDir)
	c := yawlv6.NewClient()

	xml, err := c.LoadRealData("repair-process")
	require.NoError(t, err)
	assert.Contains(t, xml, "RepairProcess")
}

// TestLoadRealData_UnknownName_ReturnsError verifies that LoadRealData rejects unknown names.
func TestLoadRealData_UnknownName_ReturnsError(t *testing.T) {
	specsDir := t.TempDir()
	t.Setenv("YAWLV6_SPECS_PATH", specsDir)
	c := yawlv6.NewClient()

	_, err := c.LoadRealData("unknown-dataset")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown real-data dataset")
}

// TestListRealData_ValidDir_ReturnsEntries verifies that ListRealData discovers
// all known real-data files present on disk.
func TestListRealData_ValidDir_ReturnsEntries(t *testing.T) {
	specsDir := t.TempDir()
	realDir := specsDir + "/real-data"
	require.NoError(t, os.MkdirAll(realDir, 0o755))
	require.NoError(t, os.WriteFile(realDir+"/OrderManagement.yawl.xml", []byte("<spec/>"), 0o644))
	require.NoError(t, os.WriteFile(realDir+"/RepairProcess.yawl.xml", []byte("<spec/>"), 0o644))

	t.Setenv("YAWLV6_SPECS_PATH", specsDir)
	c := yawlv6.NewClient()

	entries, err := c.ListRealData()
	require.NoError(t, err)
	require.Len(t, entries, 2)
	names := make([]string, len(entries))
	for i, e := range entries {
		names[i] = e.Name
	}
	assert.ElementsMatch(t, []string{"order-management", "repair-process"}, names)
}
