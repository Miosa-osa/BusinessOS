package tests

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rhl/businessos-backend/internal/integrations/osa"
	"github.com/rhl/businessos-backend/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockOSAClient is a mock implementation of the OSA client
type MockOSAClient struct {
	mock.Mock
}

func (m *MockOSAClient) GenerateApp(ctx context.Context, req *osa.AppGenerationRequest) (*osa.AppGenerationResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*osa.AppGenerationResponse), args.Error(1)
}

func (m *MockOSAClient) GetAppStatus(ctx context.Context, appID string, userID uuid.UUID) (*osa.AppStatusResponse, error) {
	args := m.Called(ctx, appID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*osa.AppStatusResponse), args.Error(1)
}

func (m *MockOSAClient) Orchestrate(ctx context.Context, req *osa.OrchestrateRequest) (*osa.OrchestrateResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*osa.OrchestrateResponse), args.Error(1)
}

func (m *MockOSAClient) GetWorkspaces(ctx context.Context, userID uuid.UUID) (*osa.WorkspacesResponse, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*osa.WorkspacesResponse), args.Error(1)
}

func (m *MockOSAClient) EditApp(ctx context.Context, req *osa.AppEditRequest) (*osa.AppEditResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*osa.AppEditResponse), args.Error(1)
}

func (m *MockOSAClient) HealthCheck(ctx context.Context) (*osa.HealthResponse, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*osa.HealthResponse), args.Error(1)
}

func (m *MockOSAClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

// TestBuildWelcomeAppPrompt tests the prompt building logic
func TestBuildWelcomeAppPrompt(t *testing.T) {
	// Note: This is a demonstration test. In a real implementation,
	// we would need to expose buildWelcomeAppPrompt or create a test helper.
	// For now, this shows what the test structure would look like.

	tests := []struct {
		name          string
		workspaceName string
		data          services.ExtractedOnboardingData
		integrations  []string
		wantContains  []string
	}{
		{
			name:          "Agency with client management",
			workspaceName: "Acme Agency",
			data: services.ExtractedOnboardingData{
				BusinessType: "agency",
				TeamSize:     "2-5",
				Role:         "founder",
				Challenge:    "client management",
			},
			integrations: []string{"hubspot", "slack"},
			wantContains: []string{
				"Acme Agency",
				"agency",
				"client management",
				"hubspot, slack",
				"client management features",
				"project tracking",
			},
		},
		{
			name:          "Startup with product focus",
			workspaceName: "TechCo Startup",
			data: services.ExtractedOnboardingData{
				BusinessType: "startup",
				TeamSize:     "6-10",
				Role:         "cto",
				Challenge:    "scaling team collaboration",
			},
			integrations: []string{"linear", "notion"},
			wantContains: []string{
				"TechCo Startup",
				"startup",
				"scaling team collaboration",
				"product roadmap",
				"team collaboration",
			},
		},
		{
			name:          "Freelancer with time tracking",
			workspaceName: "Solo Dev Shop",
			data: services.ExtractedOnboardingData{
				BusinessType: "freelance",
				TeamSize:     "solo",
				Role:         "developer",
				Challenge:    "tracking billable hours",
			},
			integrations: []string{"google", "fathom"},
			wantContains: []string{
				"Solo Dev Shop",
				"freelance",
				"tracking billable hours",
				"time tracking",
				"invoice",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test is a placeholder showing expected behavior
			// In production, we would either:
			// 1. Export buildWelcomeAppPrompt (not recommended)
			// 2. Create a test helper method
			// 3. Test via integration test

			// For now, we document expected behavior
			t.Skip("Placeholder test - buildWelcomeAppPrompt is private")

			// Example of what we would assert:
			// prompt := service.buildWelcomeAppPrompt(tt.workspaceName, tt.data, tt.integrations)
			// for _, want := range tt.wantContains {
			//     assert.Contains(t, prompt, want)
			// }
		})
	}
}

// TestGenerateInitialWorkspaceAppSuccess tests successful OSA app generation
func TestGenerateInitialWorkspaceAppSuccess(t *testing.T) {
	// Setup
	mockClient := new(MockOSAClient)
	userID := uuid.New()
	workspaceID := uuid.New()

	expectedReq := &osa.AppGenerationRequest{
		UserID:      userID,
		WorkspaceID: workspaceID,
		Name:        "Welcome Workspace",
		Description: mock.AnythingOfType("string"),
		Type:        "full-stack",
		Parameters: map[string]interface{}{
			"workspace_name": "Test Workspace",
			"business_type":  "startup",
			"team_size":      "2-5",
			"role":           "founder",
			"challenge":      "scaling the team",
			"integrations":   []string{"slack", "linear"},
			"prompt":         mock.AnythingOfType("string"),
		},
	}

	expectedResp := &osa.AppGenerationResponse{
		AppID:       "app_123",
		Status:      "pending",
		WorkspaceID: workspaceID.String(),
		Message:     "App generation started",
		CreatedAt:   time.Now(),
	}

	// Mock expectation
	mockClient.On("GenerateApp", mock.Anything, mock.MatchedBy(func(req *osa.AppGenerationRequest) bool {
		return req.UserID == userID &&
			req.WorkspaceID == workspaceID &&
			req.Name == "Welcome Workspace" &&
			req.Type == "full-stack"
	})).Return(expectedResp, nil)

	// Note: In a real test, we would need to create an OnboardingService with the mock client
	// and call generateInitialWorkspaceApp. Since it's private, we test via CompleteOnboarding.

	// This test demonstrates the expected behavior
	t.Skip("Integration test - requires full service setup")
}

// TestGenerateInitialWorkspaceAppFailure tests OSA failure handling
func TestGenerateInitialWorkspaceAppFailure(t *testing.T) {
	// Setup
	mockClient := new(MockOSAClient)
	userID := uuid.New()
	workspaceID := uuid.New()

	// Mock expectation - OSA fails
	mockClient.On("GenerateApp", mock.Anything, mock.Anything).Return(nil, assert.AnError)

	// Expected behavior: Error logged but onboarding still succeeds
	// The goroutine should handle the error gracefully

	// This test demonstrates the expected behavior
	t.Skip("Integration test - requires full service setup with logging verification")
}

// TestOnboardingOSAIntegrationDisabled tests behavior when OSA client is nil
func TestOnboardingOSAIntegrationDisabled(t *testing.T) {
	// When osaClient is nil, CompleteOnboarding should:
	// 1. Still succeed
	// 2. Not attempt to call OSA
	// 3. Not log any OSA-related errors

	// This test demonstrates the expected behavior
	t.Skip("Integration test - requires full service setup")
}

// TestPromptCustomizationByBusinessType tests prompt variations
func TestPromptCustomizationByBusinessType(t *testing.T) {
	businessTypes := []struct {
		businessType    string
		expectedFeature string
	}{
		{"agency", "client management"},
		{"consulting", "client management"},
		{"startup", "product roadmap"},
		{"freelance", "time tracking"},
		{"other", "task management"},
	}

	for _, tt := range businessTypes {
		t.Run(tt.businessType, func(t *testing.T) {
			// Test that prompts are customized based on business type
			// This would verify that different business types get appropriate features
			t.Skip("Placeholder - requires exposing prompt builder")

			// Example assertion:
			// prompt := buildWelcomeAppPrompt(...)
			// assert.Contains(t, prompt, tt.expectedFeature)
		})
	}
}

// TestOSARequestParameters tests that all onboarding data is passed correctly
func TestOSARequestParameters(t *testing.T) {
	// Verify that OSA request includes:
	// - workspace_name
	// - business_type
	// - team_size
	// - role
	// - challenge
	// - integrations
	// - prompt

	t.Skip("Integration test - requires full service setup")
}

// TestAsyncExecutionDoesNotBlockOnboarding tests that OSA runs in background
func TestAsyncExecutionDoesNotBlockOnboarding(t *testing.T) {
	// Verify that:
	// 1. CompleteOnboarding returns immediately
	// 2. OSA generation happens in background goroutine
	// 3. Slow OSA response doesn't delay onboarding completion

	t.Skip("Integration test - requires timing verification")
}

// TestOSATimeoutHandling tests 30-second timeout in OSA call
func TestOSATimeoutHandling(t *testing.T) {
	// Verify that:
	// 1. Context has 30-second timeout
	// 2. Timeout triggers cancellation
	// 3. Error is logged appropriately

	t.Skip("Integration test - requires timeout simulation")
}
