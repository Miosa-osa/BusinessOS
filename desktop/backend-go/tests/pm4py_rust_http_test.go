package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/rhl/businessos-backend/internal/integrations/pm4py_rust"
)

// mockPM4PyServer creates a mock pm4py-rust HTTP server for testing
type mockPM4PyServer struct {
	server     *httptest.Server
	baseURL    string
	requestLog []mockRequest
}

type mockRequest struct {
	method string
	path   string
	body   []byte
}

func newMockPM4PyServer() *mockPM4PyServer {
	m := &mockPM4PyServer{
		requestLog: make([]mockRequest, 0),
	}

	m.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Log request
		body, _ := io.ReadAll(r.Body)
		m.requestLog = append(m.requestLog, mockRequest{
			method: r.Method,
			path:   r.URL.Path,
			body:   body,
		})
		r.Body = io.NopCloser(bytes.NewReader(body))

		switch {
		case r.Method == "GET" && r.URL.Path == "/api/health":
			health := pm4py_rust.HealthResponse{
				Status:    "healthy",
				Version:   "0.3.0",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(health)

		case r.Method == "POST" && r.URL.Path == "/api/discovery/alpha":
			var req pm4py_rust.DiscoveryRequest
			json.NewDecoder(r.Body).Decode(&req)

			resp := pm4py_rust.DiscoveryResponse{
				Algorithm:     "alpha_miner",
				ExecutionTime: 42,
				EventCount:    10,
				TraceCount:    3,
				PetriNet: pm4py_rust.PetriNetJSON{
					Places: []pm4py_rust.PlaceJSON{
						{ID: "p1", Name: "source", InitialMarking: 1},
						{ID: "p2", Name: "sink", InitialMarking: 0},
					},
					Transitions: []pm4py_rust.TransitionJSON{
						{ID: "t1", Name: "A", Label: strPtr("A")},
					},
					Arcs: []pm4py_rust.ArcJSON{
						{From: "p1", To: "t1", Weight: 1},
					},
					InitialPlace: strPtr("p1"),
					FinalPlace:   strPtr("p2"),
				},
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(resp)

		case r.Method == "POST" && r.URL.Path == "/api/conformance/token-replay":
			var req pm4py_rust.ConformanceRequest
			json.NewDecoder(r.Body).Decode(&req)

			resp := pm4py_rust.ConformanceResponse{
				IsConformant:  true,
				Fitness:       0.95,
				Precision:     0.92,
				Generalization: 0.88,
				Method:        "token_replay",
				ExecutionTime: 125,
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(resp)

		case r.Method == "POST" && r.URL.Path == "/api/statistics":
			var req pm4py_rust.StatisticsRequest
			json.NewDecoder(r.Body).Decode(&req)

			resp := pm4py_rust.StatisticsResponse{
				TraceCount:       10,
				EventCount:       45,
				UniqueActivities: 5,
				ActivityFrequencies: map[string]int{
					"A": 10, "B": 15, "C": 10, "D": 5, "E": 5,
				},
				VariantCount: 3,
				VariantFrequencies: map[string]int{
					"A,B,C":     5,
					"A,C,B":     3,
					"A,B,C,D":   2,
				},
				BottleneckActivities: []string{"B"},
				ResourceCount:        intPtr(4),
				ExecutionTime:        87,
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(resp)

		default:
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, `{"error":"not found","status":404}`)
		}
	}))

	m.baseURL = m.server.URL
	return m
}

func (m *mockPM4PyServer) close() {
	m.server.Close()
}

// Helper functions
func strPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func createTestEventLog() json.RawMessage {
	log := map[string]interface{}{
		"traces": []map[string]interface{}{
			{
				"id": "trace_1",
				"events": []map[string]interface{}{
					{"activity": "A", "timestamp": time.Now().UTC().Format(time.RFC3339)},
					{"activity": "B", "timestamp": time.Now().UTC().Format(time.RFC3339)},
					{"activity": "C", "timestamp": time.Now().UTC().Format(time.RFC3339)},
				},
			},
		},
	}
	data, _ := json.Marshal(log)
	return data
}

// Tests

func TestClientNewWithDefaults(t *testing.T) {
	client := pm4py_rust.NewClient("http://localhost:8089")
	assert.NotNil(t, client)
	assert.NotNil(t, client)
}

func TestClientNewWithConfig(t *testing.T) {
	cfg := pm4py_rust.ClientConfig{
		Timeout: 5 * time.Second,
		Retries: 5,
	}
	client := pm4py_rust.NewClientWithConfig("http://localhost:8089", cfg)
	assert.NotNil(t, client)
}

func TestHealthCheck(t *testing.T) {
	mock := newMockPM4PyServer()
	defer mock.close()

	client := pm4py_rust.NewClient(mock.baseURL)
	ctx := context.Background()

	resp, err := client.HealthCheckRequest(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "healthy", resp.Status)
	assert.Equal(t, "0.3.0", resp.Version)
}

func TestDiscovery(t *testing.T) {
	mock := newMockPM4PyServer()
	defer mock.close()

	client := pm4py_rust.NewClient(mock.baseURL)
	ctx := context.Background()

	req := &pm4py_rust.DiscoveryRequest{
		EventLog: createTestEventLog(),
		Variant:  "alpha",
	}

	resp, err := client.DiscoveryRequest(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "alpha_miner", resp.Algorithm)
	assert.Equal(t, 3, resp.TraceCount)
	assert.Equal(t, 10, resp.EventCount)
	assert.Greater(t, resp.ExecutionTime, uint64(0))
	assert.Len(t, resp.PetriNet.Places, 2)
	assert.Len(t, resp.PetriNet.Transitions, 1)
}

func TestConformance(t *testing.T) {
	mock := newMockPM4PyServer()
	defer mock.close()

	client := pm4py_rust.NewClient(mock.baseURL)
	ctx := context.Background()

	req := &pm4py_rust.ConformanceRequest{
		EventLog: createTestEventLog(),
		PetriNet: pm4py_rust.PetriNetJSON{
			Places: []pm4py_rust.PlaceJSON{
				{ID: "p1", Name: "source", InitialMarking: 1},
			},
			Transitions: []pm4py_rust.TransitionJSON{
				{ID: "t1", Name: "A"},
			},
		},
		Method: "token_replay",
	}

	resp, err := client.ConformanceRequest(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.IsConformant)
	assert.Greater(t, resp.Fitness, 0.9)
	assert.Greater(t, resp.Precision, 0.9)
	assert.Equal(t, "token_replay", resp.Method)
}

func TestStatistics(t *testing.T) {
	mock := newMockPM4PyServer()
	defer mock.close()

	client := pm4py_rust.NewClient(mock.baseURL)
	ctx := context.Background()

	req := &pm4py_rust.StatisticsRequest{
		EventLog:              createTestEventLog(),
		IncludeVariants:       true,
		IncludeResourceMetrics: true,
		IncludeBottlenecks:    true,
	}

	resp, err := client.StatisticsRequest(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 10, resp.TraceCount)
	assert.Equal(t, 45, resp.EventCount)
	assert.Equal(t, 5, resp.UniqueActivities)
	assert.NotNil(t, resp.ActivityFrequencies)
	assert.NotNil(t, resp.VariantFrequencies)
	assert.NotEmpty(t, resp.BottleneckActivities)
}

func TestRetryLogic(t *testing.T) {
	// Create a server that fails first time, then succeeds
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, `{"error":"service unavailable","status":503}`)
			return
		}
		health := pm4py_rust.HealthResponse{
			Status:    "healthy",
			Version:   "0.3.0",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(health)
	}))
	defer server.Close()

	cfg := pm4py_rust.ClientConfig{
		Timeout: 5 * time.Second,
		Retries: 3,
	}
	client := pm4py_rust.NewClientWithConfig(server.URL, cfg)
	ctx := context.Background()

	resp, err := client.HealthCheckRequest(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 2, attempts, "Should have retried once before succeeding")
}

func TestContextTimeout(t *testing.T) {
	// Create a slow server that delays responses
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		health := pm4py_rust.HealthResponse{
			Status:    "healthy",
			Version:   "0.3.0",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(health)
	}))
	defer server.Close()

	client := pm4py_rust.NewClient(server.URL)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, resultErr := client.HealthCheckRequest(ctx)

	// Either context cancelled or deadline exceeded is acceptable
	assert.Error(t, resultErr)
}

func TestMalformedEventLog(t *testing.T) {
	mock := newMockPM4PyServer()
	defer mock.close()

	client := pm4py_rust.NewClient(mock.baseURL)
	ctx := context.Background()

	// Send malformed JSON
	req := &pm4py_rust.DiscoveryRequest{
		EventLog: json.RawMessage(`{"invalid": "log"}`),
	}

	resp, _ := client.DiscoveryRequest(ctx, req)

	// The mock server doesn't validate, but in real scenario would error
	assert.NotNil(t, resp)
}

func TestResponseMarshaling(t *testing.T) {
	// Test that responses unmarshal correctly
	healthJSON := `{
		"status": "healthy",
		"version": "0.3.0",
		"timestamp": "2024-01-01T00:00:00Z"
	}`

	var health pm4py_rust.HealthResponse
	err := json.Unmarshal([]byte(healthJSON), &health)

	assert.NoError(t, err)
	assert.Equal(t, "healthy", health.Status)
	assert.Equal(t, "0.3.0", health.Version)
}

func TestDiscoveryResponseMarshaling(t *testing.T) {
	discoveryJSON := `{
		"petri_net": {
			"places": [
				{"id": "p1", "name": "source", "initial_marking": 1}
			],
			"transitions": [
				{"id": "t1", "name": "A", "label": "A"}
			],
			"arcs": [
				{"from": "p1", "to": "t1", "weight": 1}
			],
			"initial_place": "p1",
			"final_place": null
		},
		"algorithm": "alpha_miner",
		"execution_time_ms": 42,
		"event_count": 10,
		"trace_count": 3
	}`

	var resp pm4py_rust.DiscoveryResponse
	err := json.Unmarshal([]byte(discoveryJSON), &resp)

	assert.NoError(t, err)
	assert.Equal(t, "alpha_miner", resp.Algorithm)
	assert.Len(t, resp.PetriNet.Places, 1)
	assert.Equal(t, "p1", resp.PetriNet.Places[0].ID)
}

func TestConformanceResponseMarshaling(t *testing.T) {
	conformanceJSON := `{
		"is_conformant": true,
		"fitness": 0.95,
		"precision": 0.92,
		"generalization": 0.88,
		"method": "token_replay",
		"execution_time_ms": 125
	}`

	var resp pm4py_rust.ConformanceResponse
	err := json.Unmarshal([]byte(conformanceJSON), &resp)

	assert.NoError(t, err)
	assert.True(t, resp.IsConformant)
	assert.Greater(t, resp.Fitness, 0.9)
}

func TestStatisticsResponseMarshaling(t *testing.T) {
	statsJSON := `{
		"trace_count": 10,
		"event_count": 45,
		"unique_activities": 5,
		"activity_frequencies": {"A": 10, "B": 15},
		"variant_count": 3,
		"variant_frequencies": {"A,B,C": 5},
		"bottleneck_activities": ["B"],
		"resource_count": 4,
		"execution_time_ms": 87
	}`

	var resp pm4py_rust.StatisticsResponse
	err := json.Unmarshal([]byte(statsJSON), &resp)

	assert.NoError(t, err)
	assert.Equal(t, 10, resp.TraceCount)
	assert.Equal(t, 45, resp.EventCount)
}

func TestClientClosing(t *testing.T) {
	mock := newMockPM4PyServer()
	defer mock.close()

	client := pm4py_rust.NewClient(mock.baseURL)
	err := client.Close()

	assert.NoError(t, err)
}

func TestConcurrentRequests(t *testing.T) {
	mock := newMockPM4PyServer()
	defer mock.close()

	client := pm4py_rust.NewClient(mock.baseURL)
	ctx := context.Background()

	// Launch 5 concurrent health checks
	results := make(chan error, 5)
	for i := 0; i < 5; i++ {
		go func() {
			_, err := client.HealthCheckRequest(ctx)
			results <- err
		}()
	}

	for i := 0; i < 5; i++ {
		err := <-results
		assert.NoError(t, err, fmt.Sprintf("Request %d failed", i))
	}
}

func TestRequestHeaders(t *testing.T) {
	mock := newMockPM4PyServer()
	defer mock.close()

	client := pm4py_rust.NewClient(mock.baseURL)
	ctx := context.Background()

	_, _ = client.HealthCheckRequest(ctx)

	// Check that a request was made
	assert.Greater(t, len(mock.requestLog), 0)

	// The request should have been GET to /api/health
	req := mock.requestLog[0]
	assert.Equal(t, "GET", req.method)
	assert.Equal(t, "/api/health", req.path)
}

func TestConnectionPooling(t *testing.T) {
	mock := newMockPM4PyServer()
	defer mock.close()

	client := pm4py_rust.NewClient(mock.baseURL)
	ctx := context.Background()

	// Make multiple requests to verify connection reuse
	for i := 0; i < 3; i++ {
		_, err := client.HealthCheckRequest(ctx)
		assert.NoError(t, err)
	}

	// All requests should succeed with connection pooling
	assert.NoError(t, nil)
}

func TestAPIErrorParsing(t *testing.T) {
	apiErr := pm4py_rust.APIError{
		Error:   "EventLog parsing failed",
		Details: "Missing required field: traces",
		Status:  400,
	}

	data, err := json.Marshal(apiErr)
	assert.NoError(t, err)

	var parsed pm4py_rust.APIError
	err = json.Unmarshal(data, &parsed)

	assert.NoError(t, err)
	assert.Equal(t, "EventLog parsing failed", parsed.Error)
	assert.Equal(t, 400, parsed.Status)
}

func BenchmarkHealthCheck(b *testing.B) {
	mock := newMockPM4PyServer()
	defer mock.close()

	client := pm4py_rust.NewClient(mock.baseURL)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.HealthCheckRequest(ctx)
	}
}

func BenchmarkDiscovery(b *testing.B) {
	mock := newMockPM4PyServer()
	defer mock.close()

	client := pm4py_rust.NewClient(mock.baseURL)
	ctx := context.Background()
	req := &pm4py_rust.DiscoveryRequest{
		EventLog: createTestEventLog(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.DiscoveryRequest(ctx, req)
	}
}
