package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// stubOntologyQuerier is a minimal stub that satisfies ontologyBosService.
// Only ExecuteSelect is exercised by the query handler; the rest are no-ops.
type stubOntologyQuerier struct {
	selectResult map[string]interface{}
	selectErr    error
}

func (s *stubOntologyQuerier) ListQueries(_ context.Context) ([]string, error) {
	return nil, nil
}

func (s *stubOntologyQuerier) GetConstructQuery(_ context.Context, _ string) (string, error) {
	return "", nil
}

func (s *stubOntologyQuerier) ExecuteConstruct(_ context.Context, _ string) (string, error) {
	return "", nil
}

func (s *stubOntologyQuerier) ExecuteAll(_ context.Context, _ string) (string, error) {
	return "", nil
}

func (s *stubOntologyQuerier) GenerateQueries(_ context.Context, _ string) (int, error) {
	return 0, nil
}

func (s *stubOntologyQuerier) ExecuteSelect(_ context.Context, _ string) (map[string]interface{}, error) {
	return s.selectResult, s.selectErr
}

// setupQueryRouter returns a Gin router with POST /ontology/query wired to the handler.
func setupQueryRouter(svc ontologyBosService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := newOntologyHandlerFromInterface(svc)
	r.POST("/ontology/query", h.QuerySPARQL)
	return r
}

// TestOntologyQueryHandler_MissingQuery verifies that POST with an empty body returns 400.
func TestOntologyQueryHandler_MissingQuery(t *testing.T) {
	router := setupQueryRouter(&stubOntologyQuerier{})

	req, _ := http.NewRequest(http.MethodPost, "/ontology/query", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 got %d; body: %s", w.Code, w.Body.String())
	}

	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if _, ok := body["error"]; !ok {
		t.Errorf("expected 'error' key in response, got: %v", body)
	}
}

// TestOntologyQueryHandler_InvalidJSON verifies that malformed JSON returns 400.
func TestOntologyQueryHandler_InvalidJSON(t *testing.T) {
	router := setupQueryRouter(&stubOntologyQuerier{})

	req, _ := http.NewRequest(http.MethodPost, "/ontology/query",
		strings.NewReader(`{not valid json`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 got %d; body: %s", w.Code, w.Body.String())
	}

	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if _, ok := body["error"]; !ok {
		t.Errorf("expected 'error' key in response, got: %v", body)
	}
}

// TestOntologyQueryHandler_ValidQuery verifies that a valid SPARQL query returns 200 with the stub result.
func TestOntologyQueryHandler_ValidQuery(t *testing.T) {
	expectedResult := map[string]interface{}{
		"results": map[string]interface{}{
			"bindings": []interface{}{
				map[string]interface{}{
					"subject": map[string]interface{}{
						"type":  "uri",
						"value": "http://example.com/s",
					},
				},
			},
		},
	}

	stub := &stubOntologyQuerier{
		selectResult: expectedResult,
		selectErr:    nil,
	}
	router := setupQueryRouter(stub)

	body, _ := json.Marshal(map[string]string{
		"query": "SELECT ?subject WHERE { ?subject ?p ?o } LIMIT 1",
	})
	req, _ := http.NewRequest(http.MethodPost, "/ontology/query", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d; body: %s", w.Code, w.Body.String())
	}

	var got map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if _, ok := got["results"]; !ok {
		t.Errorf("expected 'results' key in response, got keys: %v", got)
	}
}

// TestOntologyQueryHandler_ServiceError verifies that an ExecuteSelect error returns 500.
func TestOntologyQueryHandler_ServiceError(t *testing.T) {
	stub := &stubOntologyQuerier{
		selectResult: nil,
		selectErr:    errors.New("bos binary not found"),
	}
	router := setupQueryRouter(stub)

	body, _ := json.Marshal(map[string]string{
		"query": "SELECT * WHERE { ?s ?p ?o }",
	})
	req, _ := http.NewRequest(http.MethodPost, "/ontology/query", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 got %d; body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if _, ok := resp["error"]; !ok {
		t.Errorf("expected 'error' key in 500 response, got: %v", resp)
	}
}
