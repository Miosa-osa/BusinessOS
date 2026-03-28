package handlers

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTrackHealthcareData_Success(t *testing.T) {
	router := gin.Default()
	logger := slog.Default()
	router.POST("/healthcare/track", TrackHealthcareData(logger))

	req := HealthcareRESTAPIRequest{
		ResourceID:   "p123",
		ResourceType: "Patient",
		PatientID:    "patient_001",
		Data: map[string]interface{}{
			"name":  "John Doe",
			"email": "john@example.com",
		},
	}

	body, _ := json.Marshal(req)
	request := httptest.NewRequest("POST", "/healthcare/track", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, request)

	assert.Equal(t, http.StatusCreated, w.Code)
	var result map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &result)
	assert.Equal(t, "p123", result["resource_id"])
	assert.Equal(t, "Patient", result["resource_type"])
	assert.Equal(t, true, result["hipaa_check_passed"])
}

func TestTrackHealthcareData_MissingResourceID(t *testing.T) {
	router := gin.Default()
	logger := slog.Default()
	router.POST("/healthcare/track", TrackHealthcareData(logger))

	req := HealthcareRESTAPIRequest{
		ResourceID:   "",
		ResourceType: "Patient",
		PatientID:    "patient_001",
		Data: map[string]interface{}{
			"name": "John Doe",
		},
	}

	body, _ := json.Marshal(req)
	request := httptest.NewRequest("POST", "/healthcare/track", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, request)

	// Missing required field triggers validation error
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetAuditTrailData_Success(t *testing.T) {
	router := gin.Default()
	logger := slog.Default()
	router.GET("/healthcare/audit/:id", GetAuditTrailData(logger))

	patientID := "patient_001"
	request := httptest.NewRequest("GET", "/healthcare/audit/"+patientID, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, request)

	assert.Equal(t, http.StatusOK, w.Code)
	var result map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &result)
	assert.Equal(t, patientID, result["patient_id"])
	assert.Equal(t, "last_90_days", result["period"])
}

func TestGetAuditTrailData_WithCustomDays(t *testing.T) {
	router := gin.Default()
	logger := slog.Default()
	router.GET("/healthcare/audit/:id", GetAuditTrailData(logger))

	patientID := "patient_001"
	request := httptest.NewRequest("GET", "/healthcare/audit/"+patientID+"?days=30", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, request)

	assert.Equal(t, http.StatusOK, w.Code)
	var result map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &result)
	assert.Equal(t, patientID, result["patient_id"])
	assert.Equal(t, "last_30_days", result["period"])
}

func TestVerifyConsentData_NoValidConsent(t *testing.T) {
	router := gin.Default()
	logger := slog.Default()
	router.POST("/healthcare/consent/verify", VerifyConsentData(logger))

	req := VerifyConsentAPIRequest{
		PatientID: "patient_001",
	}

	body, _ := json.Marshal(req)
	request := httptest.NewRequest("POST", "/healthcare/consent/verify", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, request)

	assert.Equal(t, http.StatusForbidden, w.Code)
	var result map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &result)
	assert.Equal(t, "no_valid_consent", result["error"])
}

func TestVerifyConsentData_MissingPatientID(t *testing.T) {
	router := gin.Default()
	logger := slog.Default()
	router.POST("/healthcare/consent/verify", VerifyConsentData(logger))

	req := VerifyConsentAPIRequest{
		PatientID: "",
	}

	body, _ := json.Marshal(req)
	request := httptest.NewRequest("POST", "/healthcare/consent/verify", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, request)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var result map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &result)
	// Validation error from struct binding - check nested structure
	assert.Contains(t, result, "error")
	errorMap := result["error"].(map[string]interface{})
	assert.Contains(t, errorMap, "code")
	assert.Equal(t, "BAD_REQUEST", errorMap["code"])
}

func TestDeleteHealthcareData_Success(t *testing.T) {
	router := gin.Default()
	logger := slog.Default()
	router.DELETE("/healthcare/:id", DeleteHealthcareData(logger))

	resourceID := "p123"
	request := httptest.NewRequest("DELETE", "/healthcare/"+resourceID+"?type=Patient", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, request)

	assert.Equal(t, http.StatusOK, w.Code)
	var result map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &result)
	assert.Equal(t, resourceID, result["resource_id"])
	assert.Equal(t, true, result["fully_deleted"])
	assert.Equal(t, 0, int(result["triple_count"].(float64)))
}

func TestDeleteHealthcareData_MissingType(t *testing.T) {
	router := gin.Default()
	logger := slog.Default()
	router.DELETE("/healthcare/:id", DeleteHealthcareData(logger))

	resourceID := "p123"
	request := httptest.NewRequest("DELETE", "/healthcare/"+resourceID, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, request)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var result map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &result)
	assert.Equal(t, "missing_parameters", result["error"])
}

func TestVerifyHIPAACompliance_Success(t *testing.T) {
	router := gin.Default()
	logger := slog.Default()
	router.GET("/healthcare/hipaa/verify", VerifyHIPAACompliance(logger))

	request := httptest.NewRequest("GET", "/healthcare/hipaa/verify", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, request)

	assert.Equal(t, http.StatusOK, w.Code)
	var result map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &result)
	assert.Equal(t, true, result["compliant"])
	assert.Equal(t, true, result["access_control_pass"])
	assert.Equal(t, true, result["audit_log_pass"])
	assert.Equal(t, true, result["encryption_pass"])
	assert.Equal(t, true, result["integrity_pass"])
	assert.Equal(t, 1.0, result["compliance_score"])
}

func TestTrackHealthcareData_InvalidJSON(t *testing.T) {
	router := gin.Default()
	logger := slog.Default()
	router.POST("/healthcare/track", TrackHealthcareData(logger))

	request := httptest.NewRequest("POST", "/healthcare/track", bytes.NewBuffer([]byte("{invalid json")))
	request.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, request)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMultipleTrackCalls(t *testing.T) {
	router := gin.Default()
	logger := slog.Default()
	router.POST("/healthcare/track", TrackHealthcareData(logger))

	for i := 1; i <= 3; i++ {
		req := HealthcareRESTAPIRequest{
			ResourceID:   "p12" + string(rune('0'+byte(i))),
			ResourceType: "Patient",
			PatientID:    "patient_001",
			Data: map[string]interface{}{
				"name": "Test Patient",
			},
		}

		body, _ := json.Marshal(req)
		request := httptest.NewRequest("POST", "/healthcare/track", bytes.NewBuffer(body))
		request.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, request)

		require.Equal(t, http.StatusCreated, w.Code)
		var result map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &result)
		assert.Equal(t, req.ResourceID, result["resource_id"])
	}
}

func TestVerifyHIPAACompliance_ComplianceScore(t *testing.T) {
	router := gin.Default()
	logger := slog.Default()
	router.GET("/healthcare/hipaa/verify", VerifyHIPAACompliance(logger))

	request := httptest.NewRequest("GET", "/healthcare/hipaa/verify", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, request)

	assert.Equal(t, http.StatusOK, w.Code)
	var result map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &result)

	score := result["compliance_score"].(float64)
	assert.GreaterOrEqual(t, score, 0.0)
	assert.LessOrEqual(t, score, 1.0)

	assert.Contains(t, result, "access_control_pass")
	assert.Contains(t, result, "audit_log_pass")
	assert.Contains(t, result, "encryption_pass")
	assert.Contains(t, result, "integrity_pass")
}

func TestDeleteHealthcareData_TripleCountZero(t *testing.T) {
	router := gin.Default()
	logger := slog.Default()
	router.DELETE("/healthcare/:id", DeleteHealthcareData(logger))

	resourceID := "p123"
	request := httptest.NewRequest("DELETE", "/healthcare/"+resourceID+"?type=Patient", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, request)

	assert.Equal(t, http.StatusOK, w.Code)
	var result map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &result)

	tripleCount := int(result["triple_count"].(float64))
	assert.Equal(t, 0, tripleCount)
	assert.Equal(t, true, result["rdf_clean_confirmed"])
}

func TestGetAuditTrailData_EmptyAuditTrail(t *testing.T) {
	router := gin.Default()
	logger := slog.Default()
	router.GET("/healthcare/audit/:id", GetAuditTrailData(logger))

	patientID := "patient_001"
	request := httptest.NewRequest("GET", "/healthcare/audit/"+patientID, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, request)

	assert.Equal(t, http.StatusOK, w.Code)
	var result map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &result)

	entries := result["entries"]
	assert.NotNil(t, entries)
	totalEntries := int(result["total_entries"].(float64))
	assert.Equal(t, 0, totalEntries)
}
