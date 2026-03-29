package handlers

// OCPMHandler proxies Object-Centric Process Mining (OCPM) requests to
// pm4py-rust (port 8090) and OSA (port 8089).
//
// Routes:
//   POST /api/ocpm/throughput   → pm4py-rust /api/ocpm/performance/throughput
//   POST /api/ocpm/bottleneck   → pm4py-rust /api/ocpm/performance/bottleneck
//   POST /api/ocpm/query        → pm4py-rust /api/ocpm/llm/query
//   GET  /api/ocpm/export       → OSA        /api/ocel/export
//
// WvdA: every outbound HTTP call uses a shared context timeout (deadlock freedom).
// Armstrong: upstream errors surface as 502 Bad Gateway — no silent swallow.

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

const ocpmTimeout = 60 * time.Second

// pm4py-rust OCPM endpoint paths.
const (
	pm4pyOCPMThroughput  = "/api/ocpm/performance/throughput"
	pm4pyOCPMBottleneck  = "/api/ocpm/performance/bottleneck"
	pm4pyOCPMLLMQuery    = "/api/ocpm/llm/query"
)

// OSA OCEL export endpoint path.
const osaOCELExport = "/api/ocel/export"

// OCPMHandler proxies OCPM requests between BusinessOS, pm4py-rust, and OSA.
type OCPMHandler struct {
	pm4pyURL string
	osaURL   string
	logger   *slog.Logger
	client   *http.Client
}

// NewOCPMHandler creates an OCPMHandler.
// pm4pyURL defaults to PM4PY_RUST_URL env var (fallback: http://localhost:8090).
// osaURL defaults to OSA_URL env var (fallback: http://localhost:8089).
func NewOCPMHandler(pm4pyURL, osaURL string) *OCPMHandler {
	if pm4pyURL == "" {
		pm4pyURL = os.Getenv("PM4PY_RUST_URL")
		if pm4pyURL == "" {
			pm4pyURL = "http://localhost:8090"
		}
	}
	if osaURL == "" {
		osaURL = os.Getenv("OSA_URL")
		if osaURL == "" {
			osaURL = "http://localhost:8089"
		}
	}
	return &OCPMHandler{
		pm4pyURL: pm4pyURL,
		osaURL:   osaURL,
		logger:   slog.Default(),
		client:   &http.Client{Timeout: ocpmTimeout},
	}
}

// RegisterRoutes wires all /ocpm routes onto the given router group.
func (h *OCPMHandler) RegisterRoutes(r *gin.RouterGroup) {
	g := r.Group("/ocpm")
	g.POST("/throughput", h.Throughput)
	g.POST("/bottleneck", h.Bottleneck)
	g.POST("/query", h.Query)
	g.GET("/export", h.ExportOCEL)
}

// Throughput proxies POST /api/ocpm/throughput → pm4py-rust /api/ocpm/performance/throughput.
// Accepts OCEL 2.0 JSON; returns throughput statistics per object type.
func (h *OCPMHandler) Throughput(c *gin.Context) {
	h.proxyPostToPm4py(c, pm4pyOCPMThroughput, "ocpm.throughput")
}

// Bottleneck proxies POST /api/ocpm/bottleneck → pm4py-rust /api/ocpm/performance/bottleneck.
// Accepts OCEL JSON + optional top_n; returns ranked bottleneck list.
func (h *OCPMHandler) Bottleneck(c *gin.Context) {
	h.proxyPostToPm4py(c, pm4pyOCPMBottleneck, "ocpm.bottleneck")
}

// Query proxies POST /api/ocpm/query → pm4py-rust /api/ocpm/llm/query.
// Accepts {question, ocel, api_key}; returns {answer, grounded}.
func (h *OCPMHandler) Query(c *gin.Context) {
	h.proxyPostToPm4py(c, pm4pyOCPMLLMQuery, "ocpm.query")
}

// ExportOCEL proxies GET /api/ocpm/export → OSA /api/ocel/export.
// Returns OCEL 2.0 JSON produced by OSA.
func (h *OCPMHandler) ExportOCEL(c *gin.Context) {
	target := h.osaURL + osaOCELExport
	reqCtx := c.Request.Context()

	httpReq, err := http.NewRequestWithContext(reqCtx, http.MethodGet, target, nil)
	if err != nil {
		h.logger.Error("ocpm.export: failed to build OSA request", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to build request"})
		return
	}
	// Forward Accept header so OSA can content-negotiate if needed.
	if accept := c.GetHeader("Accept"); accept != "" {
		httpReq.Header.Set("Accept", accept)
	}

	resp, err := h.client.Do(httpReq)
	if err != nil {
		h.logger.Warn("ocpm.export: OSA unreachable", "osa_url", target, "error", err)
		c.JSON(http.StatusBadGateway, gin.H{
			"error":   "OSA unreachable",
			"osa_url": target,
		})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		h.logger.Error("ocpm.export: failed to read OSA response", "error", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to read OSA response"})
		return
	}

	if resp.StatusCode != http.StatusOK {
		h.logger.Warn("ocpm.export: OSA returned non-200",
			"status", resp.StatusCode,
			"osa_url", target,
		)
		c.JSON(http.StatusBadGateway, gin.H{
			"error":      fmt.Sprintf("OSA export returned %d", resp.StatusCode),
			"osa_status": resp.StatusCode,
		})
		return
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/json"
	}
	c.Data(resp.StatusCode, contentType, body)
}

// proxyPostToPm4py is a shared helper that reads the raw request body, forwards
// it to pm4py-rust at targetPath, and writes the response back to the caller.
// operationName is used only for structured log messages.
//
// WvdA: the caller's context carries the request deadline; no secondary timeout
// is imposed here so that the outer 60-second client timeout applies cleanly.
func (h *OCPMHandler) proxyPostToPm4py(c *gin.Context, targetPath, operationName string) {
	target := h.pm4pyURL + targetPath
	reqCtx := c.Request.Context()

	// Read the full request body so we can forward it.
	reqBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.logger.Error(operationName+": failed to read request body", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
		return
	}

	httpReq, err := http.NewRequestWithContext(reqCtx, http.MethodPost, target, bytes.NewReader(reqBody))
	if err != nil {
		h.logger.Error(operationName+": failed to build pm4py request", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to build request"})
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(httpReq)
	if err != nil {
		h.logger.Warn(operationName+": pm4py-rust unreachable", "pm4py_url", target, "error", err)
		c.JSON(http.StatusBadGateway, gin.H{
			"error":      "pm4py-rust unreachable",
			"pm4py_url":  target,
		})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		h.logger.Error(operationName+": failed to read pm4py response", "error", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to read pm4py response"})
		return
	}

	if resp.StatusCode != http.StatusOK {
		h.logger.Warn(operationName+": pm4py returned non-200",
			"status", resp.StatusCode,
			"pm4py_url", target,
		)
		c.JSON(http.StatusBadGateway, gin.H{
			"error":        fmt.Sprintf("pm4py returned %d", resp.StatusCode),
			"pm4py_status": resp.StatusCode,
		})
		return
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/json"
	}
	c.Data(resp.StatusCode, contentType, body)
}
