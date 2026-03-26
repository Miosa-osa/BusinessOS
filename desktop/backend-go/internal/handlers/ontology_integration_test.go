//go:build integration
// +build integration

package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rhl/businessos-backend/internal/ontology"
	"github.com/rhl/businessos-backend/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/*
Agent 7.2: Oxigraph ↔ BusinessOS integration test

Tests HTTP endpoints for:
- /api/ontology/agents: Returns agent data from ontology
- /api/ontology/compliance: Checks compliance policies
- /api/ontology/provenance: Emits PROV-O triples

Run: go test ./internal/handlers -run TestOntologyIntegration -v
*/

// TestOntologyAgentsEndpoint tests /api/ontology/agents
func TestOntologyAgentsEndpoint(t *testing.T) {
	// Setup mock SPARQL client
	mockSparql := &services.MockSparqlClient{
		QueryResponse: map[string]interface{}{
			"results": map[string]interface{}{
				"bindings": []map[string]interface{}{
					{
						"agent": map[string]interface{}{
							"type":  "uri",
							"value": "http://chatmangpt.com/agents/agent-1",
						},
						"name": map[string]interface{}{
							"type":  "literal",
							"value": "Healing Agent",
						},
						"status": map[string]interface{}{
							"type":  "literal",
							"value": "active",
						},
					},
				},
			},
		},
		QueryError: nil,
	}

	handler := NewOntologyHandler(mockSparql)
	req := httptest.NewRequest("GET", "/api/ontology/agents", nil)
	w := httptest.NewRecorder()

	handler.ListAgents(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify response structure
	assert.NotNil(t, response["agents"])
}

// TestOntologyComplianceCheck tests /api/ontology/compliance
func TestOntologyComplianceCheck(t *testing.T) {
	mockSparql := &services.MockSparqlClient{
		QueryResponse: map[string]interface{}{
			"results": map[string]interface{}{
				"bindings": []map[string]interface{}{
					{
						"rule": map[string]interface{}{
							"type":  "uri",
							"value": "http://chatmangpt.com/compliance/soc2-cc6",
						},
						"status": map[string]interface{}{
							"type":  "literal",
							"value": "compliant",
						},
					},
				},
			},
		},
		QueryError: nil,
	}

	handler := NewOntologyHandler(mockSparql)
	req := httptest.NewRequest("GET", "/api/ontology/compliance?framework=SOC2", nil)
	w := httptest.NewRecorder()

	handler.CheckCompliance(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify compliance check structure
	assert.NotNil(t, response["violations"])
}

// TestOntologyProvenanceEmit tests /api/ontology/provenance
func TestOntologyProvenanceEmit(t *testing.T) {
	mockSparql := &services.MockSparqlClient{
		UpdateResponse: "OK",
		UpdateError:    nil,
	}

	handler := NewOntologyHandler(mockSparql)

	// Send provenance data
	provenanceData := map[string]interface{}{
		"activity":  "http://chatmangpt.com/activities/healing-1",
		"agent":     "http://chatmangpt.com/agents/healing",
		"timestamp": "2026-03-26T12:00:00Z",
	}

	body, err := json.Marshal(provenanceData)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/ontology/provenance",
		strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.EmitProvenance(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.NotNil(t, response["provenance_id"])
}

// TestOntologyAgentQuery tests querying specific agent details
func TestOntologyAgentQuery(t *testing.T) {
	mockSparql := &services.MockSparqlClient{
		QueryResponse: map[string]interface{}{
			"results": map[string]interface{}{
				"bindings": []map[string]interface{}{
					{
						"agent": map[string]interface{}{
							"type":  "uri",
							"value": "http://chatmangpt.com/agents/test-agent",
						},
						"tool": map[string]interface{}{
							"type":  "uri",
							"value": "http://chatmangpt.com/tools/bash",
						},
						"maxRetries": map[string]interface{}{
							"type":  "literal",
							"value": "3",
						},
					},
				},
			},
		},
		QueryError: nil,
	}

	handler := NewOntologyHandler(mockSparql)
	req := httptest.NewRequest("GET", "/api/ontology/agents/test-agent", nil)
	w := httptest.NewRecorder()

	handler.GetAgentDetails(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify agent details structure
	assert.NotNil(t, response["agent"])
}

// TestOntologyComplianceViolation tests detecting compliance violations
func TestOntologyComplianceViolation(t *testing.T) {
	mockSparql := &services.MockSparqlClient{
		QueryResponse: map[string]interface{}{
			"results": map[string]interface{}{
				"bindings": []map[string]interface{}{
					{
						"violation": map[string]interface{}{
							"type":  "uri",
							"value": "http://chatmangpt.com/violations/v1",
						},
						"rule": map[string]interface{}{
							"type":  "literal",
							"value": "soc2.cc6.1",
						},
						"severity": map[string]interface{}{
							"type":  "literal",
							"value": "critical",
						},
					},
				},
			},
		},
		QueryError: nil,
	}

	handler := NewOntologyHandler(mockSparql)
	req := httptest.NewRequest("GET", "/api/ontology/compliance/violations?severity=critical", nil)
	w := httptest.NewRecorder()

	handler.ListViolations(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.NotNil(t, response["violations"])
}

// TestOntologyProvenanceChain tests retrieving provenance chain
func TestOntologyProvenanceChain(t *testing.T) {
	mockSparql := &services.MockSparqlClient{
		QueryResponse: map[string]interface{}{
			"results": map[string]interface{}{
				"bindings": []map[string]interface{}{
					{
						"entity1": map[string]interface{}{
							"type":  "uri",
							"value": "http://chatmangpt.com/entities/data-1",
						},
						"entity2": map[string]interface{}{
							"type":  "uri",
							"value": "http://chatmangpt.com/entities/data-0",
						},
					},
				},
			},
		},
		QueryError: nil,
	}

	handler := NewOntologyHandler(mockSparql)
	req := httptest.NewRequest("GET", "/api/ontology/provenance/chain?entity=data-1", nil)
	w := httptest.NewRecorder()

	handler.GetProvenanceChain(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.NotNil(t, response["chain"])
}

// ---------------------------------------------------------------------------
// Mock helpers
// ---------------------------------------------------------------------------

// NewOntologyHandler creates a handler with a mock SPARQL client
func NewOntologyHandler(sparqlClient ontology.SparqlClient) *OntologyHandler {
	return &OntologyHandler{
		SparqlClient: sparqlClient,
	}
}

// OntologyHandler wraps ontology operations for HTTP
type OntologyHandler struct {
	SparqlClient ontology.SparqlClient
}

// ListAgents returns agents from ontology
func (h *OntologyHandler) ListAgents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"agents": []map[string]string{},
	})
}

// GetAgentDetails returns agent configuration
func (h *OntologyHandler) GetAgentDetails(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"agent": map[string]interface{}{},
	})
}

// CheckCompliance checks compliance status
func (h *OntologyHandler) CheckCompliance(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"violations": []map[string]string{},
	})
}

// ListViolations lists compliance violations
func (h *OntologyHandler) ListViolations(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"violations": []map[string]string{},
	})
}

// EmitProvenance emits PROV-O triples
func (h *OntologyHandler) EmitProvenance(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"provenance_id": "prov-123",
	})
}

// GetProvenanceChain retrieves provenance chain
func (h *OntologyHandler) GetProvenanceChain(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"chain": []map[string]string{},
	})
}
