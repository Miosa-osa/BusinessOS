package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// BOSCommandType represents a BOS CLI command category
type BOSCommandType string

const (
	DiscoverCommand    BOSCommandType = "discover"
	ConformanceCommand BOSCommandType = "conformance"
	StatisticsCommand  BOSCommandType = "statistics"
	AnalyticsCommand   BOSCommandType = "analytics"
	ExportCommand      BOSCommandType = "export"
	WorkspaceCommand   BOSCommandType = "ws"
	BatchCommand       BOSCommandType = "batch"
)

// BOSInvoker provides typed access to BOS CLI commands
type BOSInvoker struct {
	bosPath string
	logger  *slog.Logger
}

// NewBOSInvoker creates a new BOS command invoker
func NewBOSInvoker(bosPath string, logger *slog.Logger) *BOSInvoker {
	return &BOSInvoker{
		bosPath: bosPath,
		logger:  logger,
	}
}

// CommandRequest represents a BOS CLI command request
type CommandRequest struct {
	Verb      string                 `json:"verb"`
	Args      map[string]interface{} `json:"args"`
	Timeout   time.Duration          `json:"timeout,omitempty"`
	JSONInput map[string]interface{} `json:"json_input,omitempty"`
}

// CommandResponse represents a BOS CLI command response
type CommandResponse struct {
	Status        string          `json:"status"`
	Command       string          `json:"command"`
	Timestamp     string          `json:"timestamp"`
	DurationMS    int64           `json:"duration_ms"`
	Output        json.RawMessage `json:"output"`
	Errors        []string        `json:"errors"`
	ExitCode      int             `json:"exit_code"`
	StderrMessage string          `json:"stderr_message,omitempty"`
}

// DiscoverModelRequest represents a discover model command
type DiscoverModelRequest struct {
	LogPath      string `json:"log_path"`
	Algorithm    string `json:"algorithm,omitempty"` // alpha, inductive, heuristic, dfg
	OutputFormat string `json:"output_format,omitempty"`
	Timeout      int    `json:"timeout,omitempty"` // seconds
}

// ConformanceCheckRequest represents a conformance check command
type ConformanceCheckRequest struct {
	LogPath   string `json:"log_path"`
	ModelPath string `json:"model_path,omitempty"`
	Timeout   int    `json:"timeout,omitempty"`
}

// StatisticsAnalysisRequest represents a statistics analysis command
type StatisticsAnalysisRequest struct {
	LogPath         string `json:"log_path"`
	IncludeVariants bool   `json:"include_variants,omitempty"`
	Timeout         int    `json:"timeout,omitempty"`
}

// ExportRequest represents an export command
type ExportRequest struct {
	Source       string `json:"source"`
	Format       string `json:"format,omitempty"` // pnml, json, svg, png
	Output       string `json:"output,omitempty"`
	AnalysisType string `json:"analysis_type,omitempty"` // for report export
	Timeout      int    `json:"timeout,omitempty"`
}

// BatchOperationRequest represents a batch operation
type BatchOperationRequest struct {
	InputDir  string `json:"input_dir"`
	ModelDir  string `json:"model_dir,omitempty"`
	Algorithm string `json:"algorithm,omitempty"`
	Workers   int    `json:"workers,omitempty"`
	Timeout   int    `json:"timeout,omitempty"`
}

// ============================================================================
// DISCOVERY COMMANDS
// ============================================================================

// DiscoverModel invokes: bos discover model <log> --algorithm <algo> --output-format <fmt>
func (b *BOSInvoker) DiscoverModel(ctx context.Context, req DiscoverModelRequest) (*CommandResponse, error) {
	args := []string{"discover", "model", req.LogPath}

	if req.Algorithm != "" {
		args = append(args, "--algorithm", req.Algorithm)
	}
	if req.OutputFormat != "" {
		args = append(args, "--output-format", req.OutputFormat)
	}

	timeout := time.Duration(req.Timeout) * time.Second
	if req.Timeout == 0 {
		timeout = 30 * time.Second
	}

	return b.executeCommand(ctx, "discover model", args, timeout)
}

// AnalyzeVariants invokes: bos discover variants <log> --top-n <n>
func (b *BOSInvoker) AnalyzeVariants(ctx context.Context, logPath string, topN int) (*CommandResponse, error) {
	args := []string{"discover", "variants", logPath}

	if topN > 0 {
		args = append(args, "--top-n", fmt.Sprintf("%d", topN))
	}

	return b.executeCommand(ctx, "discover variants", args, 30*time.Second)
}

// ============================================================================
// CONFORMANCE COMMANDS
// ============================================================================

// CheckConformance invokes: bos conformance check <log> --model <model>
func (b *BOSInvoker) CheckConformance(ctx context.Context, req ConformanceCheckRequest) (*CommandResponse, error) {
	args := []string{"conformance", "check", req.LogPath}

	if req.ModelPath != "" {
		args = append(args, "--model", req.ModelPath)
	}

	timeout := time.Duration(req.Timeout) * time.Second
	if req.Timeout == 0 {
		timeout = 30 * time.Second
	}

	return b.executeCommand(ctx, "conformance check", args, timeout)
}

// DetectDeviations invokes: bos conformance deviations <log> --baseline <baseline>
func (b *BOSInvoker) DetectDeviations(ctx context.Context, logPath string, baseline string) (*CommandResponse, error) {
	args := []string{"conformance", "deviations", logPath}

	if baseline != "" {
		args = append(args, "--baseline", baseline)
	}

	return b.executeCommand(ctx, "conformance deviations", args, 30*time.Second)
}

// ============================================================================
// STATISTICS COMMANDS
// ============================================================================

// AnalyzeStatistics invokes: bos statistics analyze <log> --include-variants <bool>
func (b *BOSInvoker) AnalyzeStatistics(ctx context.Context, req StatisticsAnalysisRequest) (*CommandResponse, error) {
	args := []string{"statistics", "analyze", req.LogPath}

	if req.IncludeVariants {
		args = append(args, "--include-variants", "true")
	}

	timeout := time.Duration(req.Timeout) * time.Second
	if req.Timeout == 0 {
		timeout = 30 * time.Second
	}

	return b.executeCommand(ctx, "statistics analyze", args, timeout)
}

// AssessQuality invokes: bos statistics quality --workspace <path>
func (b *BOSInvoker) AssessQuality(ctx context.Context, workspacePath string) (*CommandResponse, error) {
	args := []string{"statistics", "quality"}

	if workspacePath != "" {
		args = append(args, "--workspace", workspacePath)
	}

	return b.executeCommand(ctx, "statistics quality", args, 30*time.Second)
}

// ============================================================================
// ANALYTICS COMMANDS
// ============================================================================

// GenerateFingerprint invokes: bos analytics fingerprint <log> --algorithm <algo>
func (b *BOSInvoker) GenerateFingerprint(ctx context.Context, logPath string, algorithm string) (*CommandResponse, error) {
	args := []string{"analytics", "fingerprint", logPath}

	if algorithm != "" {
		args = append(args, "--algorithm", algorithm)
	}

	return b.executeCommand(ctx, "analytics fingerprint", args, 30*time.Second)
}

// AnalyzeEvolution invokes: bos analytics evolution <log> --period <period>
func (b *BOSInvoker) AnalyzeEvolution(ctx context.Context, logPath string, period string) (*CommandResponse, error) {
	args := []string{"analytics", "evolution", logPath}

	if period != "" {
		args = append(args, "--period", period)
	}

	return b.executeCommand(ctx, "analytics evolution", args, 30*time.Second)
}

// ============================================================================
// EXPORT COMMANDS
// ============================================================================

// ExportModel invokes: bos export model <source> --format <fmt> --output <path>
func (b *BOSInvoker) ExportModel(ctx context.Context, req ExportRequest) (*CommandResponse, error) {
	args := []string{"export", "model", req.Source}

	if req.Format != "" {
		args = append(args, "--format", req.Format)
	}
	if req.Output != "" {
		args = append(args, "--output", req.Output)
	}

	timeout := time.Duration(req.Timeout) * time.Second
	if req.Timeout == 0 {
		timeout = 30 * time.Second
	}

	return b.executeCommand(ctx, "export model", args, timeout)
}

// ExportReport invokes: bos export report <analysis-type> --format <fmt> --output <path>
func (b *BOSInvoker) ExportReport(ctx context.Context, req ExportRequest) (*CommandResponse, error) {
	args := []string{"export", "report", req.AnalysisType}

	if req.Format != "" {
		args = append(args, "--format", req.Format)
	}
	if req.Output != "" {
		args = append(args, "--output", req.Output)
	}

	timeout := time.Duration(req.Timeout) * time.Second
	if req.Timeout == 0 {
		timeout = 30 * time.Second
	}

	return b.executeCommand(ctx, "export report", args, timeout)
}

// ============================================================================
// WORKSPACE COMMANDS
// ============================================================================

// WorkspaceStats invokes: bos ws stats --path <path>
func (b *BOSInvoker) WorkspaceStats(ctx context.Context, workspacePath string) (*CommandResponse, error) {
	args := []string{"ws", "stats"}

	if workspacePath != "" {
		args = append(args, "--path", workspacePath)
	}

	return b.executeCommand(ctx, "ws stats", args, 30*time.Second)
}

// RefreshWorkspace invokes: bos ws refresh --path <path> --deep <bool>
func (b *BOSInvoker) RefreshWorkspace(ctx context.Context, workspacePath string, deep bool) (*CommandResponse, error) {
	args := []string{"ws", "refresh"}

	if workspacePath != "" {
		args = append(args, "--path", workspacePath)
	}
	if deep {
		args = append(args, "--deep", "true")
	}

	timeout := 60 * time.Second
	if deep {
		timeout = 120 * time.Second
	}

	return b.executeCommand(ctx, "ws refresh", args, timeout)
}

// ============================================================================
// BATCH COMMANDS
// ============================================================================

// BatchDiscover invokes: bos batch discover <input-dir> --algorithm <algo> --workers <n>
func (b *BOSInvoker) BatchDiscover(ctx context.Context, req BatchOperationRequest) (*CommandResponse, error) {
	args := []string{"batch", "discover", req.InputDir}

	if req.Algorithm != "" {
		args = append(args, "--algorithm", req.Algorithm)
	}
	if req.Workers > 0 {
		args = append(args, "--workers", fmt.Sprintf("%d", req.Workers))
	}

	timeout := time.Duration(req.Timeout) * time.Second
	if req.Timeout == 0 {
		timeout = 120 * time.Second
	}

	return b.executeCommand(ctx, "batch discover", args, timeout)
}

// BatchConform invokes: bos batch conform <log-dir> --model-dir <dir>
func (b *BOSInvoker) BatchConform(ctx context.Context, req BatchOperationRequest) (*CommandResponse, error) {
	args := []string{"batch", "conform", req.InputDir}

	if req.ModelDir != "" {
		args = append(args, "--model-dir", req.ModelDir)
	}

	timeout := time.Duration(req.Timeout) * time.Second
	if req.Timeout == 0 {
		timeout = 120 * time.Second
	}

	return b.executeCommand(ctx, "batch conform", args, timeout)
}

// ============================================================================
// INTERNAL EXECUTION
// ============================================================================

// executeCommand executes a BOS CLI command and returns parsed response
func (b *BOSInvoker) executeCommand(
	ctx context.Context,
	commandName string,
	args []string,
	timeout time.Duration,
) (*CommandResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	start := time.Now()
	cmd := exec.CommandContext(ctx, b.bosPath, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	b.logger.InfoContext(ctx, "executing bos command",
		slog.String("command", commandName),
		slog.String("args", strings.Join(args, " ")),
		slog.Duration("timeout", timeout),
	)

	err := cmd.Run()
	duration := time.Since(start)

	resp := &CommandResponse{
		Command:    commandName,
		Timestamp:  time.Now().Format(time.RFC3339),
		DurationMS: duration.Milliseconds(),
		ExitCode:   0,
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			resp.ExitCode = exitErr.ExitCode()
			resp.Status = "error"
			resp.Errors = []string{err.Error()}
			resp.StderrMessage = stderr.String()
		} else if ctx.Err() == context.DeadlineExceeded {
			resp.Status = "timeout"
			resp.Errors = []string{fmt.Sprintf("command timed out after %s", timeout)}
			resp.ExitCode = -1
		} else {
			resp.Status = "error"
			resp.Errors = []string{err.Error()}
			resp.StderrMessage = stderr.String()
			resp.ExitCode = -1
		}

		b.logger.ErrorContext(ctx, "bos command failed",
			slog.String("command", commandName),
			slog.Int("exit_code", resp.ExitCode),
			slog.Duration("duration", duration),
			slog.String("stderr", resp.StderrMessage),
		)

		return resp, err
	}

	resp.Status = "success"
	resp.Output = stdout.Bytes()

	b.logger.InfoContext(ctx, "bos command succeeded",
		slog.String("command", commandName),
		slog.Duration("duration", duration),
		slog.Int("output_size", len(resp.Output)),
	)

	return resp, nil
}

// ParseJSONResponse parses JSON output from command response
func (b *BOSInvoker) ParseJSONResponse(resp *CommandResponse, target interface{}) error {
	if resp.Status != "success" {
		return fmt.Errorf("command failed: %s", strings.Join(resp.Errors, "; "))
	}

	if err := json.Unmarshal(resp.Output, target); err != nil {
		return fmt.Errorf("failed to parse response: %w (output: %s)", err, string(resp.Output))
	}

	return nil
}

// ============================================================================
// VALIDATION HELPERS
// ============================================================================

// ValidateLogPath validates that a log file exists and is readable
func ValidateLogPath(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("invalid log path: %w", err)
	}

	// In production, check file exists with os.Stat
	// For now, just validate it's not empty
	if absPath == "" {
		return fmt.Errorf("log path cannot be empty")
	}

	return nil
}

// ValidateAlgorithm validates a discovery algorithm choice
func ValidateAlgorithm(algo string) error {
	validAlgos := map[string]bool{
		"alpha":     true,
		"inductive": true,
		"tree":      true,
		"heuristic": true,
		"dfg":       true,
	}

	if !validAlgos[algo] {
		return fmt.Errorf("invalid algorithm: %s (valid: alpha, inductive, tree, heuristic, dfg)", algo)
	}

	return nil
}

// ValidateFormat validates an export format
func ValidateFormat(format string) error {
	validFormats := map[string]bool{
		"pnml": true,
		"json": true,
		"svg":  true,
		"png":  true,
		"pdf":  true,
		"html": true,
		"md":   true,
	}

	if !validFormats[format] {
		return fmt.Errorf("invalid format: %s", format)
	}

	return nil
}

// ValidatePeriod validates a time period for analysis
func ValidatePeriod(period string) error {
	validPeriods := map[string]bool{
		"daily":   true,
		"weekly":  true,
		"monthly": true,
		"yearly":  true,
	}

	if !validPeriods[period] {
		return fmt.Errorf("invalid period: %s (valid: daily, weekly, monthly, yearly)", period)
	}

	return nil
}
