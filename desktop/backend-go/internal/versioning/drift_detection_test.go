package versioning

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/rhl/businessos-backend/internal/observability"
)

// TestDriftDetector_ClassifyDrift_None verifies no drift classification
func TestDriftDetector_ClassifyDrift_None(t *testing.T) {
	threshold := 0.05
	logger := setupTestLogger()
	telemetry := observability.New()
	detector := NewDriftDetector(threshold, logger, telemetry)

	// Test: fitness delta below threshold, minimal structural changes
	fitnessDelta := 0.03 // below 5% threshold
	structuralChanges := 2 // below 3

	severity := detector.classifyDrift(fitnessDelta, structuralChanges)

	if severity != SeverityNone {
		t.Errorf("Expected SeverityNone, got %s", severity)
	}
}

// TestDriftDetector_ClassifyDrift_Minor verifies minor drift classification
func TestDriftDetector_ClassifyDrift_Minor(t *testing.T) {
	threshold := 0.05
	logger := setupTestLogger()
	telemetry := observability.New()
	detector := NewDriftDetector(threshold, logger, telemetry)

	tests := []struct {
		name              string
		fitnessDelta      float64
		structuralChanges int
		expectedSeverity  DriftSeverity
	}{
		{
			name:              "fitness at threshold",
			fitnessDelta:      0.05,
			structuralChanges: 0,
			expectedSeverity:  SeverityMinor,
		},
		{
			name:              "structural changes at threshold",
			fitnessDelta:      0.0,
			structuralChanges: 3,
			expectedSeverity:  SeverityMinor,
		},
		{
			name:              "fitness below moderate threshold",
			fitnessDelta:      0.08,
			structuralChanges: 2,
			expectedSeverity:  SeverityMinor,
		},
		{
			name:              "structural changes below moderate threshold",
			fitnessDelta:      0.03,
			structuralChanges: 4,
			expectedSeverity:  SeverityMinor,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			severity := detector.classifyDrift(tt.fitnessDelta, tt.structuralChanges)
			if severity != tt.expectedSeverity {
				t.Errorf("Expected %s, got %s", tt.expectedSeverity, severity)
			}
		})
	}
}

// TestDriftDetector_ClassifyDrift_Moderate verifies moderate drift classification
func TestDriftDetector_ClassifyDrift_Moderate(t *testing.T) {
	threshold := 0.05
	logger := setupTestLogger()
	telemetry := observability.New()
	detector := NewDriftDetector(threshold, logger, telemetry)

	tests := []struct {
		name              string
		fitnessDelta      float64
		structuralChanges int
		expectedSeverity  DriftSeverity
	}{
		{
			name:              "fitness at moderate threshold",
			fitnessDelta:      0.10,
			structuralChanges: 0,
			expectedSeverity:  SeverityModerate,
		},
		{
			name:              "structural changes at moderate threshold",
			fitnessDelta:      0.0,
			structuralChanges: 6,
			expectedSeverity:  SeverityModerate,
		},
		{
			name:              "fitness below severe threshold",
			fitnessDelta:      0.12,
			structuralChanges: 3,
			expectedSeverity:  SeverityModerate,
		},
		{
			name:              "structural changes below severe threshold",
			fitnessDelta:      0.05,
			structuralChanges: 8,
			expectedSeverity:  SeverityModerate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			severity := detector.classifyDrift(tt.fitnessDelta, tt.structuralChanges)
			if severity != tt.expectedSeverity {
				t.Errorf("Expected %s, got %s", tt.expectedSeverity, severity)
			}
		})
	}
}

// TestDriftDetector_ClassifyDrift_Severe verifies severe drift classification
func TestDriftDetector_ClassifyDrift_Severe(t *testing.T) {
	threshold := 0.05
	logger := setupTestLogger()
	telemetry := observability.New()
	detector := NewDriftDetector(threshold, logger, telemetry)

	tests := []struct {
		name              string
		fitnessDelta      float64
		structuralChanges int
		expectedSeverity  DriftSeverity
	}{
		{
			name:              "fitness above severe threshold",
			fitnessDelta:      0.16,
			structuralChanges: 0,
			expectedSeverity:  SeveritySevere,
		},
		{
			name:              "structural changes above severe threshold",
			fitnessDelta:      0.0,
			structuralChanges: 11,
			expectedSeverity:  SeveritySevere,
		},
		{
			name:              "both metrics severe",
			fitnessDelta:      0.20,
			structuralChanges: 15,
			expectedSeverity:  SeveritySevere,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			severity := detector.classifyDrift(tt.fitnessDelta, tt.structuralChanges)
			if severity != tt.expectedSeverity {
				t.Errorf("Expected %s, got %s", tt.expectedSeverity, severity)
			}
		})
	}
}

// TestDriftDetector_DetectDrift_NoDrift verifies no drift detection scenario
func TestDriftDetector_DetectDrift_NoDrift(t *testing.T) {
	ctx := context.Background()
	logger := setupTestLogger()
	telemetry := observability.New()
	detector := NewDriftDetector(0.05, logger, telemetry)

	newModel := &ProcessModelVersion{
		ID:          uuid.New(),
		Version:     "1.0.1",
		Fitness:     0.92,
		NodesAdded:  1,
		NodesRemoved: 0,
		EdgesAdded:  1,
		EdgesRemoved: 0,
	}

	deployedModel := &ProcessModelVersion{
		ID:          uuid.New(),
		Version:     "1.0.0",
		Fitness:     0.94, // 2% drop, below 5% threshold
		NodesAdded:  0,
		NodesRemoved: 0,
		EdgesAdded:  0,
		EdgesRemoved: 0,
	}

	event := detector.DetectDrift(ctx, newModel, deployedModel)

	if event.Severity != SeverityNone {
		t.Errorf("Expected SeverityNone, got %s", event.Severity)
	}

	if event.FitnessDelta < 0.019 || event.FitnessDelta > 0.021 {
		t.Errorf("Expected fitness delta ~0.02, got %.4f", event.FitnessDelta)
	}

	if event.StructuralChanges != 2 {
		t.Errorf("Expected 2 structural changes, got %d", event.StructuralChanges)
	}

	// Verify no OTEL span was emitted (only spans with severity != None)
	spans := telemetry.GetSpans()
	if len(spans) != 0 {
		t.Errorf("Expected no spans for no-drift event, got %d", len(spans))
	}
}

// TestDriftDetector_DetectDrift_MinorDrift verifies minor drift detection with OTEL span
func TestDriftDetector_DetectDrift_MinorDrift(t *testing.T) {
	ctx := context.Background()
	logger := setupTestLogger()
	telemetry := observability.New()
	detector := NewDriftDetector(0.05, logger, telemetry)

	newModel := &ProcessModelVersion{
		ID:          uuid.New(),
		Version:     "1.0.1",
		Fitness:     0.89, // 5% drop from deployed
		NodesAdded:  2,
		NodesRemoved: 0,
		EdgesAdded:  2,
		EdgesRemoved: 0,
	}

	deployedModel := &ProcessModelVersion{
		ID:          uuid.New(),
		Version:     "1.0.0",
		Fitness:     0.94,
		NodesAdded:  0,
		NodesRemoved: 0,
		EdgesAdded:  0,
		EdgesRemoved: 0,
	}

	event := detector.DetectDrift(ctx, newModel, deployedModel)

	if event.Severity != SeverityMinor {
		t.Errorf("Expected SeverityMinor, got %s", event.Severity)
	}

	if event.FitnessDelta < 0.049 || event.FitnessDelta > 0.051 {
		t.Errorf("Expected fitness delta ~0.05, got %.4f", event.FitnessDelta)
	}

	if event.StructuralChanges != 4 {
		t.Errorf("Expected 4 structural changes, got %d", event.StructuralChanges)
	}

	// Verify OTEL span was emitted
	spans := telemetry.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("Expected 1 span, got %d", len(spans))
	}

	span := spans[0]
	if span.SpanName != "versioning.drift_detected" {
		t.Errorf("Expected span name 'versioning.drift_detected', got %s", span.SpanName)
	}

	if span.Status != "ok" {
		t.Errorf("Expected span status 'ok', got %s", span.Status)
	}

	// Verify span attributes
	if severity, ok := span.Attributes["drift.severity"].(string); !ok || severity != "minor" {
		t.Errorf("Expected drift.severity 'minor', got %v", span.Attributes["drift.severity"])
	}

	if fitnessDelta, ok := span.Attributes["drift.fitness_delta"].(float64); !ok || fitnessDelta < 0.049 || fitnessDelta > 0.051 {
		t.Errorf("Expected drift.fitness_delta ~0.05, got %v", span.Attributes["drift.fitness_delta"])
	}

	if structuralChanges, ok := span.Attributes["drift.structural_changes"].(int); !ok || structuralChanges != 4 {
		t.Errorf("Expected drift.structural_changes 4, got %v", span.Attributes["drift.structural_changes"])
	}
}

// TestDriftDetector_DetectDrift_ModerateDrift verifies moderate drift detection
func TestDriftDetector_DetectDrift_ModerateDrift(t *testing.T) {
	ctx := context.Background()
	logger := setupTestLogger()
	telemetry := observability.New()
	detector := NewDriftDetector(0.05, logger, telemetry)

	newModel := &ProcessModelVersion{
		ID:          uuid.New(),
		Version:     "1.1.0",
		Fitness:     0.84, // 10% drop
		NodesAdded:  5,
		NodesRemoved: 1,
		EdgesAdded:  3,
		EdgesRemoved: 1,
	}

	deployedModel := &ProcessModelVersion{
		ID:          uuid.New(),
		Version:     "1.0.0",
		Fitness:     0.94,
		NodesAdded:  0,
		NodesRemoved: 0,
		EdgesAdded:  0,
		EdgesRemoved: 0,
	}

	event := detector.DetectDrift(ctx, newModel, deployedModel)

	if event.Severity != SeverityModerate {
		t.Errorf("Expected SeverityModerate, got %s", event.Severity)
	}

	if event.FitnessDelta < 0.099 || event.FitnessDelta > 0.101 {
		t.Errorf("Expected fitness delta ~0.10, got %.4f", event.FitnessDelta)
	}

	if event.StructuralChanges != 10 {
		t.Errorf("Expected 10 structural changes, got %d", event.StructuralChanges)
	}
}

// TestDriftDetector_DetectDrift_SevereDrift verifies severe drift detection
func TestDriftDetector_DetectDrift_SevereDrift(t *testing.T) {
	ctx := context.Background()
	logger := setupTestLogger()
	telemetry := observability.New()
	detector := NewDriftDetector(0.05, logger, telemetry)

	newModel := &ProcessModelVersion{
		ID:          uuid.New(),
		Version:     "2.0.0",
		Fitness:     0.74, // 20% drop, severe
		NodesAdded:  8,
		NodesRemoved: 4,
		EdgesAdded:  6,
		EdgesRemoved: 3,
	}

	deployedModel := &ProcessModelVersion{
		ID:          uuid.New(),
		Version:     "1.0.0",
		Fitness:     0.94,
		NodesAdded:  0,
		NodesRemoved: 0,
		EdgesAdded:  0,
		EdgesRemoved: 0,
	}

	event := detector.DetectDrift(ctx, newModel, deployedModel)

	if event.Severity != SeveritySevere {
		t.Errorf("Expected SeveritySevere, got %s", event.Severity)
	}

	if event.FitnessDelta < 0.199 || event.FitnessDelta > 0.201 {
		t.Errorf("Expected fitness delta ~0.20, got %.4f", event.FitnessDelta)
	}

	if event.StructuralChanges != 21 {
		t.Errorf("Expected 21 structural changes, got %d", event.StructuralChanges)
	}

	// Verify OTEL span contains severe attributes
	spans := telemetry.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("Expected 1 span, got %d", len(spans))
	}

	span := spans[0]
	if severity, ok := span.Attributes["drift.severity"].(string); !ok || severity != "severe" {
		t.Errorf("Expected drift.severity 'severe', got %v", span.Attributes["drift.severity"])
	}
}

// TestDriftDetector_ShouldRollback verifies rollback decision logic
func TestDriftDetector_ShouldRollback(t *testing.T) {
	logger := setupTestLogger()
	telemetry := observability.New()
	detector := NewDriftDetector(0.05, logger, telemetry)

	tests := []struct {
		name      string
		severity  DriftSeverity
		expectRollback bool
	}{
		{
			name:      "severe drift should rollback",
			severity:  SeveritySevere,
			expectRollback: true,
		},
		{
			name:      "moderate drift should rollback",
			severity:  SeverityModerate,
			expectRollback: true,
		},
		{
			name:      "minor drift should not rollback",
			severity:  SeverityMinor,
			expectRollback: false,
		},
		{
			name:      "no drift should not rollback",
			severity:  SeverityNone,
			expectRollback: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := &DriftEvent{
				Severity: tt.severity,
			}
			shouldRollback := detector.ShouldRollback(event)
			if shouldRollback != tt.expectRollback {
				t.Errorf("Expected rollback=%v, got %v", tt.expectRollback, shouldRollback)
			}
		})
	}
}

// TestDriftDetector_GetRecommendation verifies human-readable recommendations
func TestDriftDetector_GetRecommendation(t *testing.T) {
	logger := setupTestLogger()
	telemetry := observability.New()
	detector := NewDriftDetector(0.05, logger, telemetry)

	tests := []struct {
		name         string
		severity     DriftSeverity
		fitnessDelta float64
		changes      int
		expectContains []string
	}{
		{
			name:     "severe recommendation",
			severity: SeveritySevere,
			fitnessDelta: 0.20,
			changes:  15,
			expectContains: []string{"Severe drift", "Immediate rollback", "0.2000", "15"},
		},
		{
			name:     "moderate recommendation",
			severity: SeverityModerate,
			fitnessDelta: 0.12,
			changes:  8,
			expectContains: []string{"Moderate drift", "Rollback recommended", "0.1200", "8"},
		},
		{
			name:     "minor recommendation",
			severity: SeverityMinor,
			fitnessDelta: 0.06,
			changes:  4,
			expectContains: []string{"Minor drift", "Monitor closely", "0.0600", "4"},
		},
		{
			name:     "none recommendation",
			severity: SeverityNone,
			fitnessDelta: 0.02,
			changes:  1,
			expectContains: []string{"No significant drift", "acceptable threshold"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := &DriftEvent{
				Severity:        tt.severity,
				FitnessDelta:    tt.fitnessDelta,
				StructuralChanges: tt.changes,
			}
			recommendation := detector.GetRecommendation(event)

			for _, expected := range tt.expectContains {
				if !contains(recommendation, expected) {
					t.Errorf("Recommendation missing expected text '%s'\nGot: %s", expected, recommendation)
				}
			}
		})
	}
}

// TestDriftDetector_CountStructuralChanges verifies structural change counting
func TestDriftDetector_CountStructuralChanges(t *testing.T) {
	logger := setupTestLogger()
	telemetry := observability.New()
	detector := NewDriftDetector(0.05, logger, telemetry)

	newModel := &ProcessModelVersion{
		NodesAdded:   5,
		NodesRemoved: 2,
		EdgesAdded:   8,
		EdgesRemoved: 3,
	}

	deployedModel := &ProcessModelVersion{
		NodesAdded:   0,
		NodesRemoved: 0,
		EdgesAdded:   0,
		EdgesRemoved: 0,
	}

	changes := detector.countStructuralChanges(newModel, deployedModel)
	expected := 5 + 2 + 8 + 3 // 18 total changes

	if changes != expected {
		t.Errorf("Expected %d structural changes, got %d", expected, changes)
	}
}

// TestDriftDetector_NewDetector verifies constructor with default threshold
func TestDriftDetector_NewDetector(t *testing.T) {
	logger := setupTestLogger()
	telemetry := observability.New()

	// Test with valid threshold
	detector := NewDriftDetector(0.10, logger, telemetry)
	if detector.threshold != 0.10 {
		t.Errorf("Expected threshold 0.10, got %.4f", detector.threshold)
	}

	// Test with invalid threshold (should default to 0.05)
	detector = NewDriftDetector(0.0, logger, telemetry)
	if detector.threshold != 0.05 {
		t.Errorf("Expected default threshold 0.05, got %.4f", detector.threshold)
	}

	// Test with negative threshold (should default to 0.05)
	detector = NewDriftDetector(-0.05, logger, telemetry)
	if detector.threshold != 0.05 {
		t.Errorf("Expected default threshold 0.05, got %.4f", detector.threshold)
	}
}

// TestDriftDetector_SpanAttributes verifies OTEL span contains all required attributes
func TestDriftDetector_SpanAttributes(t *testing.T) {
	ctx := context.Background()
	logger := setupTestLogger()
	telemetry := observability.New()
	detector := NewDriftDetector(0.05, logger, telemetry)

	newModel := &ProcessModelVersion{
		ID:          uuid.New(),
		Version:     "1.0.1",
		Fitness:     0.88,
		NodesAdded:  3,
		NodesRemoved: 1,
		EdgesAdded:  2,
		EdgesRemoved: 1,
	}

	deployedModel := &ProcessModelVersion{
		ID:          uuid.New(),
		Version:     "1.0.0",
		Fitness:     0.94,
		NodesAdded:  0,
		NodesRemoved: 0,
		EdgesAdded:  0,
		EdgesRemoved: 0,
	}

	event := detector.DetectDrift(ctx, newModel, deployedModel)

	if event.Severity == SeverityNone {
		t.Skip("Skipping span attribute test - no drift detected")
	}

	spans := telemetry.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("Expected 1 span, got %d", len(spans))
	}

	span := spans[0]
	requiredAttrs := []string{
		"drift.severity",
		"drift.fitness_delta",
		"drift.structural_changes",
		"new_model.version",
		"new_model.fitness",
		"deployed_model.version",
		"deployed_model.fitness",
		"drift.threshold",
	}

	for _, attr := range requiredAttrs {
		if _, exists := span.Attributes[attr]; !exists {
			t.Errorf("Missing required span attribute: %s", attr)
		}
	}
}

// Helper functions

func setupTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
