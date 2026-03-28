package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupBOSSchemaRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	h := NewBOSGatewayHandler(nil, nil)
	schema := router.Group("/api/bos/schema")
	schema.POST("/import", h.SchemaImport)
	schema.GET("/export/:schema_id", h.SchemaExport)
	schema.POST("/validate/:schema_id", h.SchemaValidate)
	schema.POST("/update", h.SchemaUpdate)
	return router
}

func TestSchemaImport_ReturnsOK(t *testing.T) {
	router := setupBOSSchemaRouter()

	body, _ := json.Marshal(map[string]interface{}{
		"schema": map[string]interface{}{
			"schema_name": "test_fibo",
			"tables": []interface{}{
				map[string]interface{}{"name": "fibo_deal"},
				map[string]interface{}{"name": "fibo_party"},
			},
		},
		"format": "json",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/bos/schema/import", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp bosSchemaImportResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Status != "ok" {
		t.Errorf("expected status 'ok', got %q", resp.Status)
	}
	if resp.SchemaID == "" {
		t.Errorf("expected non-empty schema_id")
	}
	if resp.TablesImported != 2 {
		t.Errorf("expected tables_imported=2, got %d", resp.TablesImported)
	}
	if resp.ContentHash == "" {
		t.Errorf("expected non-empty content_hash")
	}
	if resp.DurationMs < 0 {
		t.Errorf("expected duration_ms >= 0, got %d", resp.DurationMs)
	}
}

func TestSchemaImport_MissingBody_Returns400(t *testing.T) {
	router := setupBOSSchemaRouter()

	body, _ := json.Marshal(map[string]interface{}{
		"format": "json",
		// neither "schema" nor "data" present
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/bos/schema/import", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestSchemaExport_ReturnsOK(t *testing.T) {
	router := setupBOSSchemaRouter()

	schemaID := "test-schema-id-123"
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/bos/schema/export/"+schemaID+"?format=json", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp bosSchemaExportResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Status != "ok" {
		t.Errorf("expected status 'ok', got %q", resp.Status)
	}
	if resp.SchemaID != schemaID {
		t.Errorf("expected schema_id=%q, got %q", schemaID, resp.SchemaID)
	}
	if resp.Data == "" {
		t.Errorf("expected non-empty data field")
	}
}

func TestSchemaValidate_ReturnsOK(t *testing.T) {
	router := setupBOSSchemaRouter()

	schemaID := "test-schema-id-456"
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/bos/schema/validate/"+schemaID, nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if valid, ok := resp["valid"].(bool); !ok || !valid {
		t.Errorf("expected valid=true, got %v", resp["valid"])
	}
}

func TestSchemaUpdate_ReturnsOK(t *testing.T) {
	router := setupBOSSchemaRouter()

	body, _ := json.Marshal(map[string]interface{}{
		"schema_id": "test-schema-id-789",
		"schema": map[string]interface{}{
			"schema_name": "test_fibo_updated",
			"tables": []interface{}{
				map[string]interface{}{"name": "fibo_deal"},
			},
		},
		"format": "json",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/bos/schema/update", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp bosSchemaImportResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Status != "ok" {
		t.Errorf("expected status 'ok', got %q", resp.Status)
	}
}
