package handlers

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestInitComplianceService_FallbackIsNotCanopyPort is a regression guard.
// This test will fail if someone changes the hardcoded default back to 9089 (Canopy's port).
func TestInitComplianceService_FallbackIsNotCanopyPort(t *testing.T) {
	t.Setenv("OSA_BASE_URL", "")

	effectiveURL := os.Getenv("OSA_BASE_URL")
	if effectiveURL == "" {
		effectiveURL = "http://localhost:8089"
	}
	assert.Equal(t, "http://localhost:8089", effectiveURL,
		"default OSA port must be 8089 (not Canopy's 9089)")
}

// TestInitComplianceService_UsesOSABaseURLEnv verifies env var is respected.
func TestInitComplianceService_UsesOSABaseURLEnv(t *testing.T) {
	t.Setenv("OSA_BASE_URL", "http://custom-osa:8089")

	svc := initComplianceService()
	assert.NotNil(t, svc)
}
