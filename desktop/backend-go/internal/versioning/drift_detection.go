package versioning

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/rhl/businessos-backend/internal/observability"
)

// DriftSeverity indicates the impact level of detected drift
type DriftSeverity string

const (
	// SeverityNone indicates no significant drift
	SeverityNone DriftSeverity = "none"
	// SeverityMinor indicates minor drift requiring monitoring
	SeverityMinor DriftSeverity = "minor"
	// SeverityModerate indicates moderate drift requiring attention
	SeverityModerate DriftSeverity = "moderate"
	// SeveritySevere indicates severe drift requiring immediate action
	SeveritySevere DriftSeverity = "severe"
)

// DriftEvent represents a detected drift event between model versions
type DriftEvent struct {
	Severity           DriftSeverity `json:"severity"`
	FitnessDelta       float64       `json:"fitness_delta"`
	StructuralChanges  int           `json:"structural_changes"`
	NewModelVersion    string        `json:"new_model_version"`
	DeployedModelVersion string      `json:"deployed_model_version"`
}

// DriftDetector analyzes process model versions for drift detection
type DriftDetector struct {
	threshold float64           // 0.05 = 5% fitness drop threshold
	logger    *slog.Logger      // structured logger
	telemetry *observability.Telemetry // OTEL span emission
}

// NewDriftDetector creates a new drift detector with the specified threshold
func NewDriftDetector(threshold float64, logger *slog.Logger, telemetry *observability.Telemetry) *DriftDetector {
	if threshold <= 0 {
		threshold = 0.05 // default to 5%
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &DriftDetector{
		threshold: threshold,
		logger:    logger,
		telemetry: telemetry,
	}
}

// DetectDrift analyzes drift between a new model and the deployed model
func (d *DriftDetector) DetectDrift(
	ctx context.Context,
	newModel *ProcessModelVersion,
	deployedModel *ProcessModelVersion,
) *DriftEvent {
	// Calculate fitness delta (deployed - new, so positive delta = degradation)
	fitnessDelta := deployedModel.Fitness - newModel.Fitness

	// Count structural changes
	structuralChanges := d.countStructuralChanges(newModel, deployedModel)

	// Classify drift severity
	severity := d.classifyDrift(fitnessDelta, structuralChanges)

	// Emit OTEL span if drift detected
	if severity != SeverityNone {
		d.emitDriftSpan(ctx, newModel, deployedModel, severity, fitnessDelta, structuralChanges)
	}

	event := &DriftEvent{
		Severity:             severity,
		FitnessDelta:         fitnessDelta,
		StructuralChanges:    structuralChanges,
		NewModelVersion:      newModel.Version,
		DeployedModelVersion: deployedModel.Version,
	}

	d.logger.Info("drift detection completed",
		"severity", severity,
		"fitness_delta", fitnessDelta,
		"structural_changes", structuralChanges,
		"new_version", newModel.Version,
		"deployed_version", deployedModel.Version,
	)

	return event
}

// classifyDrift determines severity based on fitness delta and structural changes
func (d *DriftDetector) classifyDrift(fitnessDelta float64, structuralChanges int) DriftSeverity {
	// Check for severe drift first
	if fitnessDelta > 0.15 || structuralChanges > 10 {
		return SeveritySevere
	}

	// Check for moderate drift (fitness 0.10-0.15 OR structural 6-10)
	if fitnessDelta >= 0.10 || structuralChanges >= 6 {
		return SeverityModerate
	}

	// Check for minor drift (fitness threshold-0.10 OR structural 3-5)
	if fitnessDelta >= d.threshold || structuralChanges >= 3 {
		return SeverityMinor
	}

	// No significant drift
	return SeverityNone
}

// countStructuralChanges calculates total structural modifications between versions
func (d *DriftDetector) countStructuralChanges(newModel, deployedModel *ProcessModelVersion) int {
	// Use the change summary fields from ProcessModelVersion
	nodesChanged := newModel.NodesAdded + newModel.NodesRemoved
	edgesChanged := newModel.EdgesAdded + newModel.EdgesRemoved

	return nodesChanged + edgesChanged
}

// emitDriftSpan creates an OTEL span for drift detection events
func (d *DriftDetector) emitDriftSpan(
	ctx context.Context,
	newModel *ProcessModelVersion,
	deployedModel *ProcessModelVersion,
	severity DriftSeverity,
	fitnessDelta float64,
	structuralChanges int,
) {
	attributes := map[string]any{
		"drift.severity":          string(severity),
		"drift.fitness_delta":     fitnessDelta,
		"drift.structural_changes": structuralChanges,
		"new_model.version":       newModel.Version,
		"new_model.fitness":       newModel.Fitness,
		"deployed_model.version":  deployedModel.Version,
		"deployed_model.fitness":  deployedModel.Fitness,
		"drift.threshold":         d.threshold,
	}

	span, _ := d.telemetry.StartSpan(ctx, "versioning.drift_detected", attributes)
	d.telemetry.EndSpan(span, "ok", "")
}

// ShouldRollback determines if the detected drift warrants a rollback
func (d *DriftDetector) ShouldRollback(event *DriftEvent) bool {
	// Rollback recommended for severe or moderate drift
	return event.Severity == SeveritySevere || event.Severity == SeverityModerate
}

// GetRecommendation returns human-readable guidance based on drift event
func (d *DriftDetector) GetRecommendation(event *DriftEvent) string {
	switch event.Severity {
	case SeveritySevere:
		return fmt.Sprintf("Severe drift detected (fitness drop: %.4f, %d structural changes). "+
			"Immediate rollback recommended. Review process discovery configuration and data quality.",
			event.FitnessDelta, event.StructuralChanges)
	case SeverityModerate:
		return fmt.Sprintf("Moderate drift detected (fitness drop: %.4f, %d structural changes). "+
			"Rollback recommended. Investigate process changes and validate new model.",
			event.FitnessDelta, event.StructuralChanges)
	case SeverityMinor:
		return fmt.Sprintf("Minor drift detected (fitness drop: %.4f, %d structural changes). "+
			"Monitor closely. Consider holding deployment for further validation.",
			event.FitnessDelta, event.StructuralChanges)
	case SeverityNone:
		return "No significant drift detected. New model is within acceptable threshold."
	default:
		return "Unable to provide recommendation: unknown drift severity."
	}
}
