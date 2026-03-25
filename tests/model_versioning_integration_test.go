package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

/**
 * Integration Test Suite: Process Model Versioning
 *
 * Tests complete workflows across HTTP API, Go backend, and PostgreSQL
 * Validates that discovered models can be versioned, compared, and rolled back safely
 */

// TestModelVersioningE2E validates the complete versioning workflow
func TestModelVersioningE2E(t *testing.T) {
	baseURL := "http://localhost:8001/api"
	client := &http.Client{Timeout: 10 * time.Second}

	// Step 1: Create initial model
	t.Log("Step 1: Creating initial process model")

	modelPayload := map[string]interface{}{
		"name":        "Loan Processing",
		"description": "E2E test for model versioning",
	}

	modelJSON, _ := json.Marshal(modelPayload)
	resp, err := client.Post(
		baseURL+"/process-models",
		"application/json",
		bytes.NewReader(modelJSON),
	)
	if err != nil {
		t.Fatalf("Create model failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Expected 201, got %d", resp.StatusCode)
	}

	var modelResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&modelResp)
	modelID := modelResp["id"].(string)
	t.Logf("Created model: %s", modelID)

	// Step 2: Discover and create first version
	t.Log("Step 2: Creating first model version")

	version1Payload := map[string]interface{}{
		"model": map[string]interface{}{
			"nodes": []map[string]interface{}{
				{"id": "start", "type": "event", "label": "Loan Request"},
				{"id": "check", "type": "task", "label": "Check Credit"},
				{"id": "decision", "type": "xor_gateway", "label": "Approved?"},
				{"id": "approve", "type": "task", "label": "Send Approval"},
				{"id": "reject", "type": "task", "label": "Send Rejection"},
				{"id": "end", "type": "event", "label": "Process Complete"},
			},
			"edges": []map[string]interface{}{
				{"id": "e1", "source": "start", "target": "check"},
				{"id": "e2", "source": "check", "target": "decision"},
				{"id": "e3", "source": "decision", "target": "approve", "label": "yes"},
				{"id": "e4", "source": "decision", "target": "reject", "label": "no"},
				{"id": "e5", "source": "approve", "target": "end"},
				{"id": "e6", "source": "reject", "target": "end"},
			},
		},
		"metrics": map[string]interface{}{
			"nodes_count":      6,
			"edges_count":      6,
			"fitness":          0.87,
			"average_duration": 45.3,
			"covered_traces":   250,
			"variants":         2,
		},
		"change_type":     "patch",
		"description":     "Initial model from inductive mining",
		"created_by":      "test-discovery-engine",
		"discovery_source": "inductive",
		"tags":            []string{"initial", "automated"},
	}

	v1JSON, _ := json.Marshal(version1Payload)
	resp, err = client.Post(
		fmt.Sprintf("%s/process-models/%s/versions", baseURL, modelID),
		"application/json",
		bytes.NewReader(v1JSON),
	)
	if err != nil {
		t.Fatalf("Create version 1 failed: %v", err)
	}
	defer resp.Body.Close()

	var v1Resp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&v1Resp)
	v1Version := v1Resp["version"].(string)
	t.Logf("Created version: %s", v1Version)

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Expected 201, got %d: %v", resp.StatusCode, v1Resp)
	}

	// Step 3: Retrieve version and verify metadata
	t.Log("Step 3: Retrieving version metadata")

	resp, err = client.Get(
		fmt.Sprintf("%s/process-models/%s/versions/%s", baseURL, modelID, v1Version),
	)
	if err != nil {
		t.Fatalf("Get version failed: %v", err)
	}
	defer resp.Body.Close()

	var v1Full map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&v1Full)

	if v1Full["fitness"].(float64) != 0.87 {
		t.Errorf("Fitness mismatch: expected 0.87, got %v", v1Full["fitness"])
	}
	if !v1Full["is_released"].(bool) == false {
		t.Error("Version should not be released initially")
	}

	// Step 4: Create improved version
	t.Log("Step 4: Creating improved version with additional nodes")

	time.Sleep(100 * time.Millisecond) // Ensure different timestamp

	version2Payload := map[string]interface{}{
		"model": map[string]interface{}{
			"nodes": []map[string]interface{}{
				{"id": "start", "type": "event", "label": "Loan Request"},
				{"id": "validate", "type": "task", "label": "Validate Application"},
				{"id": "check", "type": "task", "label": "Check Credit"},
				{"id": "review", "type": "task", "label": "Manual Review"},
				{"id": "decision", "type": "xor_gateway", "label": "Approved?"},
				{"id": "approve", "type": "task", "label": "Send Approval"},
				{"id": "reject", "type": "task", "label": "Send Rejection"},
				{"id": "end", "type": "event", "label": "Process Complete"},
			},
			"edges": []map[string]interface{}{
				{"id": "e0", "source": "start", "target": "validate"},
				{"id": "e1", "source": "validate", "target": "check"},
				{"id": "e2", "source": "check", "target": "review"},
				{"id": "e3", "source": "review", "target": "decision"},
				{"id": "e4", "source": "decision", "target": "approve", "label": "yes"},
				{"id": "e5", "source": "decision", "target": "reject", "label": "no"},
				{"id": "e6", "source": "approve", "target": "end"},
				{"id": "e7", "source": "reject", "target": "end"},
			},
		},
		"metrics": map[string]interface{}{
			"nodes_count":      8,
			"edges_count":      8,
			"fitness":          0.93,
			"average_duration": 52.1,
			"covered_traces":   285,
			"variants":         2,
		},
		"change_type": "minor",
		"description": "Added validation and manual review steps to improve accuracy",
		"created_by":  "test-discovery-engine",
		"tags":        []string{"improved", "validation"},
	}

	v2JSON, _ := json.Marshal(version2Payload)
	resp, err = client.Post(
		fmt.Sprintf("%s/process-models/%s/versions", baseURL, modelID),
		"application/json",
		bytes.NewReader(v2JSON),
	)
	if err != nil {
		t.Fatalf("Create version 2 failed: %v", err)
	}
	defer resp.Body.Close()

	var v2Resp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&v2Resp)
	v2Version := v2Resp["version"].(string)
	t.Logf("Created version: %s", v2Version)

	// Step 5: Compare versions
	t.Log("Step 5: Comparing versions")

	resp, err = client.Get(
		fmt.Sprintf("%s/process-models/%s/versions/compare?from=%s&to=%s",
			baseURL, modelID, v1Version, v2Version),
	)
	if err != nil {
		t.Fatalf("Compare versions failed: %v", err)
	}
	defer resp.Body.Close()

	var diffResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&diffResp)

	// Verify structural changes detected
	structDiff := diffResp["structural_diff"].(map[string]interface{})
	nodesAdded := structDiff["nodes_added"].([]interface{})
	if len(nodesAdded) != 2 {
		t.Errorf("Expected 2 nodes added, got %d", len(nodesAdded))
	}

	// Verify metrics diff
	metricsDiff := diffResp["metrics_diff"].(map[string]interface{})
	fitnessDiff := metricsDiff["fitness"].(map[string]interface{})
	if fitnessDiff["delta"].(float64) < 0.05 {
		t.Errorf("Expected fitness improvement >= 0.05, got %v", fitnessDiff["delta"])
	}

	t.Logf("Differences detected: %v breaking changes", len(diffResp["breaking_changes"].([]interface{})))

	// Step 6: Release version
	t.Log("Step 6: Releasing version with sufficient fitness")

	releasePayload := map[string]interface{}{
		"release_notes": "Production-ready model with improved accuracy and validation workflow",
	}

	releaseJSON, _ := json.Marshal(releasePayload)
	resp, err = client.Post(
		fmt.Sprintf("%s/process-models/%s/versions/%s/release", baseURL, modelID, v2Version),
		"application/json",
		bytes.NewReader(releaseJSON),
	)
	if err != nil {
		t.Fatalf("Release version failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected 200, got %d: %s", resp.StatusCode, string(body))
	}

	// Step 7: Verify version is released
	t.Log("Step 7: Verifying version is released")

	resp, err = client.Get(
		fmt.Sprintf("%s/process-models/%s/versions/%s", baseURL, modelID, v2Version),
	)
	if err != nil {
		t.Fatalf("Get version failed: %v", err)
	}
	defer resp.Body.Close()

	var v2Full map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&v2Full)

	if !v2Full["is_released"].(bool) {
		t.Error("Version should be marked as released")
	}

	// Step 8: List version history
	t.Log("Step 8: Listing version history")

	resp, err = client.Get(
		fmt.Sprintf("%s/process-models/%s/versions?limit=10", baseURL, modelID),
	)
	if err != nil {
		t.Fatalf("List versions failed: %v", err)
	}
	defer resp.Body.Close()

	var historyResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&historyResp)

	versions := historyResp["versions"].([]interface{})
	if len(versions) != 2 {
		t.Errorf("Expected 2 versions, got %d", len(versions))
	}

	t.Logf("Version history: %d versions found", len(versions))

	// Step 9: Test rollback impact analysis
	t.Log("Step 9: Analyzing rollback impact")

	resp, err = client.Get(
		fmt.Sprintf("%s/process-models/%s/versions/%s/rollback-impact", baseURL, modelID, v1Version),
	)
	if err != nil {
		t.Fatalf("Analyze rollback impact failed: %v", err)
	}
	defer resp.Body.Close()

	var impactResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&impactResp)

	t.Logf("Rollback impact: %d instances to pause, breaking changes: %v",
		int(impactResp["instances_to_pause"].(float64)),
		impactResp["breaking_changes"])

	t.Log("✓ Complete versioning workflow validated successfully")
}

// TestVersionReleaseQualityGate validates fitness requirements
func TestVersionReleaseQualityGate(t *testing.T) {
	baseURL := "http://localhost:8001/api"
	client := &http.Client{Timeout: 10 * time.Second}

	// Create model
	modelPayload := map[string]interface{}{
		"name": fmt.Sprintf("QualityGateTest-%s", uuid.New().String()[:8]),
	}
	modelJSON, _ := json.Marshal(modelPayload)
	resp, _ := client.Post(baseURL+"/process-models", "application/json", bytes.NewReader(modelJSON))
	var modelResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&modelResp)
	resp.Body.Close()

	modelID := modelResp["id"].(string)

	// Create version with LOW fitness (should fail release)
	lowFitnessPayload := map[string]interface{}{
		"model": map[string]interface{}{"nodes": []map[string]interface{}{}},
		"metrics": map[string]interface{}{
			"nodes_count":      0,
			"edges_count":      0,
			"fitness":          0.80, // Below 0.85 threshold
			"average_duration": 0,
			"covered_traces":   0,
			"variants":         0,
		},
		"change_type": "patch",
		"description": "Low fitness test",
		"created_by":  "test",
	}

	lowJSON, _ := json.Marshal(lowFitnessPayload)
	resp, _ = client.Post(
		fmt.Sprintf("%s/process-models/%s/versions", baseURL, modelID),
		"application/json",
		bytes.NewReader(lowJSON),
	)
	var lowResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&lowResp)
	resp.Body.Close()

	lowVersion := lowResp["version"].(string)

	// Try to release low fitness version (should fail)
	releasePayload := map[string]interface{}{
		"release_notes": "This should fail",
	}
	releaseJSON, _ := json.Marshal(releasePayload)
	resp, _ = client.Post(
		fmt.Sprintf("%s/process-models/%s/versions/%s/release", baseURL, modelID, lowVersion),
		"application/json",
		bytes.NewReader(releaseJSON),
	)

	if resp.StatusCode == http.StatusOK {
		t.Error("Release should fail for fitness < 0.85")
	}
	resp.Body.Close()

	t.Log("✓ Quality gate enforcement verified")
}

// TestConcurrentVersionCreation validates thread-safety
func TestConcurrentVersionCreation(t *testing.T) {
	baseURL := "http://localhost:8001/api"

	// Create model
	modelPayload := map[string]interface{}{
		"name": fmt.Sprintf("ConcurrentTest-%s", uuid.New().String()[:8]),
	}
	modelJSON, _ := json.Marshal(modelPayload)
	resp, _ := http.Post(baseURL+"/process-models", "application/json", bytes.NewReader(modelJSON))
	var modelResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&modelResp)
	resp.Body.Close()

	modelID := modelResp["id"].(string)

	// Create 5 versions concurrently
	done := make(chan error, 5)
	for i := 0; i < 5; i++ {
		go func(idx int) {
			vPayload := map[string]interface{}{
				"model": map[string]interface{}{
					"id": fmt.Sprintf("concurrent_%d", idx),
				},
				"metrics": map[string]interface{}{
					"nodes_count":      idx + 1,
					"edges_count":      idx,
					"fitness":          0.90,
					"average_duration": float64(idx * 10),
					"covered_traces":   100 + idx*10,
					"variants":         idx + 1,
				},
				"change_type": "patch",
				"description": fmt.Sprintf("Concurrent version %d", idx),
				"created_by":  "test",
			}

			vJSON, _ := json.Marshal(vPayload)
			resp, err := http.Post(
				fmt.Sprintf("%s/process-models/%s/versions", baseURL, modelID),
				"application/json",
				bytes.NewReader(vJSON),
			)

			if err != nil {
				done <- err
				return
			}

			if resp.StatusCode != http.StatusCreated {
				done <- fmt.Errorf("Expected 201, got %d", resp.StatusCode)
				resp.Body.Close()
				return
			}

			resp.Body.Close()
			done <- nil
		}(i)
	}

	// Wait for all to complete
	for i := 0; i < 5; i++ {
		if err := <-done; err != nil {
			t.Errorf("Concurrent creation failed: %v", err)
		}
	}

	t.Log("✓ Concurrent version creation validated")
}
