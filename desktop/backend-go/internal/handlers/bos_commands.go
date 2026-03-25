// Package handlers provides HTTP handlers for BusinessOS.
package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BOSCommandInvoker invokes BOS CLI commands and handles results.
type BOSCommandInvoker struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
	bos    *BOSGatewayHandler
	mu     sync.RWMutex
}

// CommandInvokeRequest represents a BOS command invocation.
type CommandInvokeRequest struct {
	Command   string                 `json:"command" binding:"required"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
	Format    string                 `json:"format,omitempty"` // json, table, csv, plaintext
	Timeout   int                    `json:"timeout_seconds,omitempty"`
}

// CommandInvokeResponse represents the result of a command invocation.
type CommandInvokeResponse struct {
	Status      string          `json:"status"`
	Command     string          `json:"command"`
	Output      json.RawMessage `json:"output"`
	ErrorMsg    string          `json:"error,omitempty"`
	DurationMs  int64           `json:"duration_ms"`
	Timestamp   string          `json:"timestamp"`
	ExitCode    int             `json:"exit_code"`
}

// NewBOSCommandInvoker creates a new BOS command invoker.
func NewBOSCommandInvoker(pool *pgxpool.Pool, logger *slog.Logger, bos *BOSGatewayHandler) *BOSCommandInvoker {
	if logger == nil {
		logger = slog.Default()
	}
	return &BOSCommandInvoker{
		pool:   pool,
		logger: logger,
		bos:    bos,
	}
}

// RegisterBOSCommandRoutes registers command invocation routes.
func RegisterBOSCommandRoutes(api *gin.RouterGroup, invoker *BOSCommandInvoker) {
	commands := api.Group("/bos/commands")
	{
		commands.POST("/invoke", invoker.InvokeCommand)
		commands.POST("/discover", invoker.InvokeDiscover)
		commands.POST("/conform", invoker.InvokeConform)
		commands.POST("/statistics", invoker.InvokeStatistics)
		commands.POST("/quality-check", invoker.InvokeQualityCheck)
		commands.POST("/fingerprint", invoker.InvokeFingerprint)
		commands.POST("/variability", invoker.InvokeVariability)
		commands.POST("/org-evolution", invoker.InvokeOrgEvolution)
		commands.POST("/variant-analysis", invoker.InvokeVariantAnalysis)
		commands.POST("/export-model", invoker.InvokeExportModel)
		commands.POST("/batch-discover", invoker.InvokeBatchDiscover)
		commands.GET("/help", invoker.GetCommandHelp)
	}
}

// InvokeCommand invokes an arbitrary BOS command.
func (inv *BOSCommandInvoker) InvokeCommand(c *gin.Context) {
	var req CommandInvokeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	inv.logger.Info("Invoking BOS command",
		slog.String("command", req.Command),
		slog.String("format", req.Format),
	)

	// Set default format and timeout
	if req.Format == "" {
		req.Format = "json"
	}
	if req.Timeout == 0 {
		req.Timeout = 30
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(req.Timeout)*time.Second)
	defer cancel()

	result, err := inv.executeCommand(ctx, req)
	if err != nil {
		inv.logger.Error("Command execution failed",
			slog.String("command", req.Command),
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// InvokeDiscover invokes a discovery command on an event log.
func (inv *BOSCommandInvoker) InvokeDiscover(c *gin.Context) {
	type DiscoverRequest struct {
		LogPath   string `json:"log_path" binding:"required"`
		Algorithm string `json:"algorithm,omitempty"` // inductive, alpha, heuristic, ilp
		Format    string `json:"format,omitempty"`
	}

	var req DiscoverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate file exists
	if _, err := os.Stat(req.LogPath); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Log file not found: %s", req.LogPath)})
		return
	}

	inv.logger.Info("Discover command",
		slog.String("log_path", req.LogPath),
		slog.String("algorithm", req.Algorithm),
	)

	cmd := CommandInvokeRequest{
		Command: "discover",
		Arguments: map[string]interface{}{
			"log_path":  req.LogPath,
			"algorithm": req.Algorithm,
		},
		Format: req.Format,
	}

	result, err := inv.executeCommand(context.Background(), cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// InvokeConform invokes a conformance check.
func (inv *BOSCommandInvoker) InvokeConform(c *gin.Context) {
	type ConformRequest struct {
		LogPath string `json:"log_path" binding:"required"`
		ModelID string `json:"model_id" binding:"required"`
		Format  string `json:"format,omitempty"`
	}

	var req ConformRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if _, err := os.Stat(req.LogPath); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Log file not found: %s", req.LogPath)})
		return
	}

	inv.logger.Info("Conformance check",
		slog.String("log_path", req.LogPath),
		slog.String("model_id", req.ModelID),
	)

	cmd := CommandInvokeRequest{
		Command: "conform",
		Arguments: map[string]interface{}{
			"log_path": req.LogPath,
			"model_id": req.ModelID,
		},
		Format: req.Format,
	}

	result, err := inv.executeCommand(context.Background(), cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// InvokeStatistics invokes statistics extraction.
func (inv *BOSCommandInvoker) InvokeStatistics(c *gin.Context) {
	type StatsRequest struct {
		LogPath string `json:"log_path" binding:"required"`
		Format  string `json:"format,omitempty"`
	}

	var req StatsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if _, err := os.Stat(req.LogPath); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Log file not found: %s", req.LogPath)})
		return
	}

	inv.logger.Info("Statistics extraction", slog.String("log_path", req.LogPath))

	cmd := CommandInvokeRequest{
		Command: "statistics",
		Arguments: map[string]interface{}{
			"log_path": req.LogPath,
		},
		Format: req.Format,
	}

	result, err := inv.executeCommand(context.Background(), cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// InvokeQualityCheck invokes a quality check.
func (inv *BOSCommandInvoker) InvokeQualityCheck(c *gin.Context) {
	type QCRequest struct {
		DataPath string   `json:"data_path" binding:"required"`
		Metrics  []string `json:"metrics,omitempty"`
		Report   bool     `json:"report,omitempty"`
		Format   string   `json:"format,omitempty"`
	}

	var req QCRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	inv.logger.Info("Quality check", slog.String("data_path", req.DataPath))

	cmd := CommandInvokeRequest{
		Command: "quality_check",
		Arguments: map[string]interface{}{
			"data_path": req.DataPath,
			"metrics":   req.Metrics,
			"report":    req.Report,
		},
		Format: req.Format,
	}

	result, err := inv.executeCommand(context.Background(), cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// InvokeFingerprint invokes fingerprint calculation.
func (inv *BOSCommandInvoker) InvokeFingerprint(c *gin.Context) {
	type FPRequest struct {
		LogPath        string `json:"log_path" binding:"required"`
		BaselineModel  string `json:"baseline_model,omitempty"`
		Format         string `json:"format,omitempty"`
	}

	var req FPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	inv.logger.Info("Fingerprint calculation", slog.String("log_path", req.LogPath))

	cmd := CommandInvokeRequest{
		Command: "fingerprint",
		Arguments: map[string]interface{}{
			"log_path":        req.LogPath,
			"baseline_model":  req.BaselineModel,
		},
		Format: req.Format,
	}

	result, err := inv.executeCommand(context.Background(), cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// InvokeVariability invokes variability analysis.
func (inv *BOSCommandInvoker) InvokeVariability(c *gin.Context) {
	type VarRequest struct {
		LogPath            string  `json:"log_path" binding:"required"`
		BaselineVariant    string  `json:"baseline_variant,omitempty"`
		VarianceThreshold  float64 `json:"variance_threshold,omitempty"`
		Format             string  `json:"format,omitempty"`
	}

	var req VarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	inv.logger.Info("Variability analysis", slog.String("log_path", req.LogPath))

	cmd := CommandInvokeRequest{
		Command: "variability",
		Arguments: map[string]interface{}{
			"log_path":           req.LogPath,
			"baseline_variant":   req.BaselineVariant,
			"variance_threshold": req.VarianceThreshold,
		},
		Format: req.Format,
	}

	result, err := inv.executeCommand(context.Background(), cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// InvokeOrgEvolution invokes organizational evolution analysis.
func (inv *BOSCommandInvoker) InvokeOrgEvolution(c *gin.Context) {
	type OERequest struct {
		LogPath     string `json:"log_path" binding:"required"`
		StartDate   string `json:"start_date,omitempty"`
		EndDate     string `json:"end_date,omitempty"`
		Granularity string `json:"granularity,omitempty"` // daily, weekly, monthly
		Format      string `json:"format,omitempty"`
	}

	var req OERequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	inv.logger.Info("Org evolution analysis", slog.String("log_path", req.LogPath))

	cmd := CommandInvokeRequest{
		Command: "org_evolution",
		Arguments: map[string]interface{}{
			"log_path":     req.LogPath,
			"start_date":   req.StartDate,
			"end_date":     req.EndDate,
			"granularity":  req.Granularity,
		},
		Format: req.Format,
	}

	result, err := inv.executeCommand(context.Background(), cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// InvokeVariantAnalysis invokes variant analysis.
func (inv *BOSCommandInvoker) InvokeVariantAnalysis(c *gin.Context) {
	type VARequest struct {
		LogPath             string  `json:"log_path" binding:"required"`
		TopN                int     `json:"top_n,omitempty"`
		SimilarityThreshold float64 `json:"similarity_threshold,omitempty"`
		Format              string  `json:"format,omitempty"`
	}

	var req VARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	inv.logger.Info("Variant analysis", slog.String("log_path", req.LogPath))

	cmd := CommandInvokeRequest{
		Command: "variant_analysis",
		Arguments: map[string]interface{}{
			"log_path":              req.LogPath,
			"top_n":                 req.TopN,
			"similarity_threshold":  req.SimilarityThreshold,
		},
		Format: req.Format,
	}

	result, err := inv.executeCommand(context.Background(), cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// InvokeExportModel invokes model export.
func (inv *BOSCommandInvoker) InvokeExportModel(c *gin.Context) {
	type ExportRequest struct {
		SourceID     string `json:"source_id" binding:"required"`
		OutputPath   string `json:"output_path" binding:"required"`
		Format       string `json:"format,omitempty"`
		WithMetadata bool   `json:"with_metadata,omitempty"`
	}

	var req ExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ensure output directory exists
	outDir := filepath.Dir(req.OutputPath)
	if err := os.MkdirAll(outDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Cannot create output dir: %v", err)})
		return
	}

	inv.logger.Info("Export model",
		slog.String("source_id", req.SourceID),
		slog.String("output_path", req.OutputPath),
	)

	cmd := CommandInvokeRequest{
		Command: "export_model",
		Arguments: map[string]interface{}{
			"source_id":      req.SourceID,
			"output_path":    req.OutputPath,
			"format":         req.Format,
			"with_metadata":  req.WithMetadata,
		},
		Format: "json",
	}

	result, err := inv.executeCommand(context.Background(), cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// InvokeBatchDiscover invokes batch discovery.
func (inv *BOSCommandInvoker) InvokeBatchDiscover(c *gin.Context) {
	type BatchRequest struct {
		LogDirectory string `json:"log_directory" binding:"required"`
		Pattern      string `json:"pattern,omitempty"`
		Algorithm    string `json:"algorithm,omitempty"`
		Workers      int    `json:"workers,omitempty"`
	}

	var req BatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if _, err := os.Stat(req.LogDirectory); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Directory not found: %s", req.LogDirectory)})
		return
	}

	inv.logger.Info("Batch discover",
		slog.String("directory", req.LogDirectory),
		slog.String("pattern", req.Pattern),
	)

	cmd := CommandInvokeRequest{
		Command: "batch_discover",
		Arguments: map[string]interface{}{
			"log_directory": req.LogDirectory,
			"pattern":       req.Pattern,
			"algorithm":     req.Algorithm,
			"workers":       req.Workers,
		},
		Format: "json",
	}

	result, err := inv.executeCommand(context.Background(), cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetCommandHelp returns help for BOS commands.
func (inv *BOSCommandInvoker) GetCommandHelp(c *gin.Context) {
	help := gin.H{
		"commands": []gin.H{
			{"name": "discover", "description": "Discover process model from event log"},
			{"name": "conform", "description": "Check process conformance"},
			{"name": "statistics", "description": "Extract log statistics"},
			{"name": "quality_check", "description": "Check data quality"},
			{"name": "fingerprint", "description": "Calculate trace fingerprint"},
			{"name": "variability", "description": "Analyze process variability"},
			{"name": "org_evolution", "description": "Analyze organizational evolution"},
			{"name": "variant_analysis", "description": "Analyze process variants"},
			{"name": "export_model", "description": "Export process model"},
			{"name": "batch_discover", "description": "Batch discover multiple logs"},
		},
	}
	c.JSON(http.StatusOK, help)
}

// executeCommand executes a BOS command via subprocess.
func (inv *BOSCommandInvoker) executeCommand(ctx context.Context, req CommandInvokeRequest) (*CommandInvokeResponse, error) {
	start := time.Now()

	// Build command arguments
	args := []string{req.Command, "--format", req.Format}

	// Add arguments to command
	for key, value := range req.Arguments {
		if value != nil && value != "" && value != false {
			if v, ok := value.(string); ok && v != "" {
				args = append(args, fmt.Sprintf("--%s", key), v)
			} else if v, ok := value.(bool); ok && v {
				args = append(args, fmt.Sprintf("--%s", key))
			}
		}
	}

	// Execute BOS CLI command
	// Note: In production, ensure 'bos' binary is in PATH
	cmd := exec.CommandContext(ctx, "bos", args...)
	output, err := cmd.CombinedOutput()

	duration := time.Since(start)
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
	}

	resp := &CommandInvokeResponse{
		Command:    req.Command,
		DurationMs: duration.Milliseconds(),
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		ExitCode:   exitCode,
	}

	if err != nil {
		resp.Status = "failed"
		resp.ErrorMsg = string(output)
		inv.logger.Error("Command execution failed",
			slog.String("command", req.Command),
			slog.Int("exit_code", exitCode),
		)
	} else {
		resp.Status = "success"
		resp.Output = json.RawMessage(output)
		inv.logger.Info("Command executed successfully",
			slog.String("command", req.Command),
			slog.Int64("duration_ms", resp.DurationMs),
		)
	}

	return resp, nil
}
