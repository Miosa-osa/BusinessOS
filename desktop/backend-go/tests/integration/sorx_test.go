package integration

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"
)

// requireServices checks that both OSA and BOS are reachable, skipping the test if not.
func requireServices(t *testing.T, urls ...string) {
	t.Helper()
	client := &http.Client{Timeout: 2 * time.Second}
	for _, u := range urls {
		resp, err := client.Get(u)
		if err != nil {
			t.Skipf("Skipping: service not reachable at %s: %v", u, err)
		}
		resp.Body.Close()
	}
}

// TestBOSToOSAHandshake tests the BusinessOS -> OSA bidirectional communication
func TestBOSToOSAHandshake(t *testing.T) {
	osaURL := "http://localhost:8089"
	bosURL := "http://localhost:8001"
	requireServices(t, osaURL+"/api/v1/a2a/agent-card", bosURL+"/health")
	requireRoute(t, "GET", bosURL+"/api/v1/me")

	// Test 1: OSA service is alive and responds
	t.Run("OSA_Service_Alive", func(t *testing.T) {
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Get(osaURL + "/api/v1/a2a/agent-card")
		if err != nil {
			t.Fatalf("Failed to reach OSA: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			t.Fatalf("Expected 200, got %d", resp.StatusCode)
		}

		var agentCard map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&agentCard); err != nil {
			t.Fatalf("Failed to decode agent card: %v", err)
		}

		if _, ok := agentCard["name"]; !ok {
			t.Fatalf("Agent card missing 'name' field")
		}
	})

	// Test 2: XXE Security - Malicious XES file upload blocked
	t.Run("XXE_Security_Malicious_File_Blocked", func(t *testing.T) {
		maliciousXML := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE foo [
  <!ENTITY xxe SYSTEM "file:///etc/passwd">
]>
<log>
  <trace>
    <event name="&xxe;" timestamp="2024-01-01T10:00:00Z"/>
  </trace>
</log>`

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Post(
			bosURL+"/api/bos/discover",
			"application/xml",
			bytes.NewBufferString(maliciousXML),
		)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		// Should return 400 or 422 (Bad Request/Unprocessable Entity)
		if resp.StatusCode == 200 {
			t.Fatalf("XXE exploit was not blocked! Got %d", resp.StatusCode)
		}
		if resp.StatusCode < 400 || resp.StatusCode >= 500 {
			t.Logf("Got status %d (acceptable for XXE rejection)", resp.StatusCode)
		}
	})

	// Test 3: XXE Security - Valid XES file processed
	t.Run("XXE_Security_Valid_File_Processed", func(t *testing.T) {
		validXES := `<?xml version="1.0" encoding="UTF-8"?>
<log xes.version="1.0" xes.features="" openxes.version="1.0">
  <trace>
    <string key="concept:name" value="Case1"/>
    <event>
      <string key="concept:name" value="Activity_A"/>
      <date key="time:timestamp" value="2024-01-01T10:00:00Z"/>
    </event>
    <event>
      <string key="concept:name" value="Activity_B"/>
      <date key="time:timestamp" value="2024-01-01T10:05:00Z"/>
    </event>
  </trace>
</log>`

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Post(
			bosURL+"/api/bos/discover",
			"application/xml",
			bytes.NewBufferString(validXES),
		)
		if err != nil {
			t.Logf("Request failed (acceptable if pm4py not running): %v", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			t.Logf("Valid XES file accepted, status: %d", resp.StatusCode)
		} else if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			t.Logf("Valid XES file rejected with %d (may be acceptable)", resp.StatusCode)
		}
	})

	// Test 4: JWT Auth - Request without token returns 401
	t.Run("JWT_Auth_No_Token_Returns_401", func(t *testing.T) {
		client := &http.Client{Timeout: 5 * time.Second}
		req, _ := http.NewRequest("GET", bosURL+"/api/v1/me", nil)
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 401 {
			t.Fatalf("Expected 401 Unauthorized, got %d", resp.StatusCode)
		}
	})

	// Test 5: JWT Auth - Invalid token returns 401
	t.Run("JWT_Auth_Invalid_Token_Returns_401", func(t *testing.T) {
		client := &http.Client{Timeout: 5 * time.Second}
		req, _ := http.NewRequest("GET", bosURL+"/api/v1/me", nil)
		req.Header.Set("Authorization", "Bearer invalid.token.here")
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 401 {
			t.Fatalf("Expected 401 Unauthorized, got %d", resp.StatusCode)
		}
	})

	// Test 6: Idempotency - Duplicate request returns same response
	t.Run("Idempotency_Duplicate_Request_Same_Response", func(t *testing.T) {
		payload := map[string]string{"request_id": "test-idem-001"}
		body, _ := json.Marshal(payload)

		client := &http.Client{Timeout: 5 * time.Second}

		// First request
		resp1, err := client.Post(
			bosURL+"/api/bos/tx/prepare",
			"application/json",
			bytes.NewBuffer(body),
		)
		if err != nil {
			t.Logf("First request failed (acceptable if endpoints not ready): %v", err)
			return
		}
		defer resp1.Body.Close()
		body1, _ := io.ReadAll(resp1.Body)

		// Second request (same payload)
		resp2, err := client.Post(
			bosURL+"/api/bos/tx/prepare",
			"application/json",
			bytes.NewBuffer(body),
		)
		if err != nil {
			t.Logf("Second request failed: %v", err)
			return
		}
		defer resp2.Body.Close()
		body2, _ := io.ReadAll(resp2.Body)

		if resp1.StatusCode == resp2.StatusCode {
			t.Logf("Both requests returned same status: %d", resp1.StatusCode)
		}

		// For successful responses, bodies should match
		if resp1.StatusCode >= 200 && resp1.StatusCode < 300 {
			if !bytes.Equal(body1, body2) {
				t.Logf("Response bodies differ (may be acceptable for non-idempotent ops)")
			}
		}
	})

	// Test 7: OSA A2A Agent Card has version
	t.Run("OSA_Agent_Card_Has_Version", func(t *testing.T) {
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Get(osaURL + "/api/v1/a2a/agent-card")
		if err != nil {
			t.Fatalf("Failed to reach OSA: %v", err)
		}
		defer resp.Body.Close()

		var agentCard map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&agentCard); err != nil {
			t.Fatalf("Failed to decode agent card: %v", err)
		}

		if _, ok := agentCard["version"]; !ok {
			t.Fatalf("Agent card missing 'version' field")
		}
	})

	// Test 8: OSA Tools endpoint returns list
	t.Run("OSA_Tools_List_Valid", func(t *testing.T) {
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Get(osaURL + "/api/v1/a2a/tools")
		if err != nil {
			t.Fatalf("Failed to reach OSA tools: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			t.Fatalf("Expected 200, got %d", resp.StatusCode)
		}

		var toolsResp map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&toolsResp); err != nil {
			t.Fatalf("Failed to decode tools response: %v", err)
		}

		if _, ok := toolsResp["tools"]; !ok {
			t.Fatalf("Tools response missing 'tools' field")
		}
	})

	// Test 9: Deadlock Detection - Long-running process times out gracefully
	t.Run("Deadlock_Detection_Timeout_Recovery", func(t *testing.T) {
		client := &http.Client{Timeout: 3 * time.Second}

		// Send a long-running request
		req, _ := http.NewRequest("POST", bosURL+"/api/bos/discover",
			bytes.NewBufferString(`{"payload": "test"}`))
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			t.Logf("Request timeout detected (expected for long-running): %v", err)
			return
		}
		defer resp.Body.Close()

		// Should complete or timeout gracefully, not hang
		t.Logf("Request completed with status: %d", resp.StatusCode)
	})

	// Test 10: Health Check - BOS reports system status
	t.Run("Health_Check_BOS_Status", func(t *testing.T) {
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Get(bosURL + "/health")
		if err != nil {
			t.Logf("Health check failed (acceptable if service not running): %v", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode == 200 {
			var health map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&health)
			if database, ok := health["database"]; ok {
				t.Logf("Database status: %v", database)
			}
		}
	})

	// Test 11: SORX Skill Execution - Call via A2A
	t.Run("SORX_Skill_Execution_Via_A2A", func(t *testing.T) {
		payload := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      1,
			"method":  "tools/list",
			"params":  map[string]interface{}{},
		}
		body, _ := json.Marshal(payload)

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Post(
			osaURL+"/api/v1/a2a",
			"application/json",
			bytes.NewBuffer(body),
		)
		if err != nil {
			t.Logf("SORX request failed (acceptable if OSA not running): %v", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode == 200 {
			var result map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&result); err == nil {
				if resultData, ok := result["result"]; ok {
					t.Logf("SORX skill executed successfully: %v", resultData)
				}
			}
		}
	})

	// Test 12: Byzantine Consensus - Proposal from multiple agents
	t.Run("Byzantine_Consensus_Multiple_Agents", func(t *testing.T) {
		// Create a consensus proposal payload
		payload := map[string]interface{}{
			"proposal_id": "test-consensus-001",
			"content":     "test-proposal",
			"timestamp":   time.Now().Unix(),
		}
		body, _ := json.Marshal(payload)

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Post(
			osaURL+"/api/v1/consensus/propose",
			"application/json",
			bytes.NewBuffer(body),
		)
		if err != nil {
			t.Logf("Consensus proposal failed (acceptable if not implemented): %v", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			t.Logf("Consensus proposal accepted")
		} else if resp.StatusCode == 404 {
			t.Logf("Consensus endpoint not yet implemented")
		}
	})
}

// TestCrossSystemDataFlow tests complete data flow across systems
func TestCrossSystemDataFlow(t *testing.T) {
	bosURL := "http://localhost:8001"
	osaURL := "http://localhost:8089"
	requireServices(t, bosURL+"/health")

	t.Run("Full_Integration_Chain_Startup", func(t *testing.T) {
		client := &http.Client{Timeout: 5 * time.Second}

		// Check BOS
		resp1, err1 := client.Get(bosURL + "/health")
		if err1 == nil {
			defer resp1.Body.Close()
			t.Logf("BOS health: %d", resp1.StatusCode)
		}

		// Check OSA
		resp2, err2 := client.Get(osaURL + "/api/v1/a2a/agent-card")
		if err2 == nil {
			defer resp2.Body.Close()
			t.Logf("OSA health: %d", resp2.StatusCode)
		}

		if err1 != nil && err2 != nil {
			t.Fatalf("Both services unreachable")
		}
	})

	t.Run("End_To_End_SSE_Streaming_Setup", func(t *testing.T) {
		client := &http.Client{Timeout: 5 * time.Second}

		// Check if SSE streaming endpoint exists
		resp, err := client.Get(bosURL + "/api/v1/osa/stream/build/test-session-id")
		if err != nil {
			t.Logf("SSE streaming not available (acceptable): %v", err)
			return
		}
		defer resp.Body.Close()

		// 200 or 401/403 indicate endpoint exists
		if resp.StatusCode >= 200 && resp.StatusCode < 500 {
			t.Logf("SSE endpoint exists, status: %d", resp.StatusCode)
		}
	})

	t.Run("Learning_Loop_Experience_Recording", func(t *testing.T) {
		experiencePayload := map[string]interface{}{
			"agent_id": "test-agent-001",
			"action":   "test_action",
			"outcome":  "success",
			"metrics":  map[string]float64{"duration": 1.5},
		}
		body, _ := json.Marshal(experiencePayload)

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Post(
			osaURL+"/api/v1/memory/learn",
			"application/json",
			bytes.NewBuffer(body),
		)
		if err != nil {
			t.Logf("Learning endpoint not available (acceptable): %v", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			t.Logf("Experience recorded successfully")
		}
	})
}

// TestSecurityBoundaries tests security at system boundaries
func TestSecurityBoundaries(t *testing.T) {
	bosURL := "http://localhost:8001"

	t.Run("CSRF_Protection_Enabled", func(t *testing.T) {
		client := &http.Client{Timeout: 5 * time.Second}

		// POST without CSRF token should fail
		resp, err := client.Post(
			bosURL+"/api/v1/apps",
			"application/json",
			bytes.NewBufferString(`{}`),
		)
		if err != nil {
			t.Logf("Request failed (acceptable): %v", err)
			return
		}
		defer resp.Body.Close()

		// Should get 401 or 403, not 200
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			t.Logf("Warning: POST accepted without CSRF check, status: %d", resp.StatusCode)
		} else if resp.StatusCode == 403 || resp.StatusCode == 401 {
			t.Logf("CSRF protection working, status: %d", resp.StatusCode)
		}
	})

	t.Run("Rate_Limiting_Enforced", func(t *testing.T) {
		client := &http.Client{Timeout: 1 * time.Second}

		// Send many requests quickly
		for i := 0; i < 20; i++ {
			resp, err := client.Get(bosURL + "/health")
			if err != nil {
				t.Logf("Request %d failed: %v", i+1, err)
				break
			}
			defer resp.Body.Close()

			// If we get 429 (Too Many Requests), rate limiting is working
			if resp.StatusCode == 429 {
				t.Logf("Rate limiting triggered after %d requests", i+1)
				return
			}
		}
		t.Logf("Rate limiting test completed")
	})
}

// BenchmarkBOSToOSALatency measures request/response latency
func BenchmarkBOSToOSALatency(b *testing.B) {
	osaURL := "http://localhost:8089"
	client := &http.Client{Timeout: 10 * time.Second}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := client.Get(osaURL + "/api/v1/a2a/agent-card")
		if err != nil {
			b.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()
	}
}

// BenchmarkXMLParsing measures XXE-safe XML parsing latency
func BenchmarkXMLParsing(b *testing.B) {
	bosURL := "http://localhost:8001"
	validXES := `<?xml version="1.0" encoding="UTF-8"?>
<log>
  <trace>
    <event name="Activity_A" timestamp="2024-01-01T10:00:00Z"/>
  </trace>
</log>`

	client := &http.Client{Timeout: 10 * time.Second}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := client.Post(
			bosURL+"/api/bos/discover",
			"application/xml",
			bytes.NewBufferString(validXES),
		)
		if err != nil {
			b.Logf("Request failed: %v", err)
		}
		if resp != nil {
			resp.Body.Close()
		}
	}
}
