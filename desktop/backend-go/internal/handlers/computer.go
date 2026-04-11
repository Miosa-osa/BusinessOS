package handlers

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/config"
	"github.com/rhl/businessos-backend/internal/utils"
)

// ComputerHandler serves cloud computer status and lifecycle endpoints.
// All responses are mock/realistic data — no MIOSA integration yet.
type ComputerHandler struct {
	cfg *config.Config
}

// NewComputerHandler constructs a ComputerHandler.
func NewComputerHandler(cfg *config.Config) *ComputerHandler {
	return &ComputerHandler{cfg: cfg}
}

// ── request types ─────────────────────────────────────────────────────────────

type createComputerRequest struct {
	Plan string `json:"plan" binding:"required"`
}

type upgradeComputerRequest struct {
	Plan string `json:"plan" binding:"required"`
}

type runtimeActionRequest struct {
	// name comes from the URL param; body is intentionally empty for now
}

// ── mock data helpers ─────────────────────────────────────────────────────────

var defaultRuntimes = []Runtime{
	{Name: "claude-code", Status: "active", MemoryMB: 2100, Uptime: 3600},
	{Name: "ollama", Status: "idle", MemoryMB: 1200, Uptime: 7200},
}

func mockComputer(plan string) Computer {
	specs := planSpecs(plan)
	now := time.Now().UTC()
	return Computer{
		ID:          "cmp_01j9xk2mhz8r4v7n",
		Status:      "active",
		Plan:        plan,
		RAM:         specs.RAMGB,
		CPUs:        specs.CPUs,
		StorageGB:   specs.StorageGB,
		StorageUsed: 3.2,
		Domain:      "acme.app.businessos.com",
		Runtimes:    defaultRuntimes,
		CreatedAt:   now.Add(-72 * time.Hour),
		UpdatedAt:   now,
	}
}

// planSpecs returns the resource allocation for a plan, reusing PlanInfo.
func planSpecs(plan string) PlanInfo {
	for _, p := range availablePlans {
		if p.ID == plan {
			return p
		}
	}
	// fallback for unknown plans — treat as pro
	return availablePlans[0]
}

// ── handlers ──────────────────────────────────────────────────────────────────

// GetComputer handles GET /api/computer.
// In cloud mode returns a realistic active computer.
// In local mode returns a null computer with an explanatory message.
func (h *ComputerHandler) GetComputer(c *gin.Context) {
	if !h.cfg.IsCloudDeployment() {
		c.JSON(http.StatusOK, gin.H{
			"computer": nil,
			"message":  "No cloud computer. Running locally.",
		})
		return
	}

	comp := mockComputer("pro")
	c.JSON(http.StatusOK, gin.H{"computer": comp})
}

// CreateComputer handles POST /api/computer.
// Accepts {"plan": "pro"} and returns a mock provisioning response.
func (h *ComputerHandler) CreateComputer(c *gin.Context) {
	var req createComputerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, slog.Default(), err)
		return
	}

	comp := mockComputer(req.Plan)
	comp.Status = "provisioning"
	comp.UpdatedAt = time.Now().UTC()

	slog.InfoContext(c.Request.Context(), "computer: create requested", "plan", req.Plan)

	c.JSON(http.StatusAccepted, gin.H{
		"computer": comp,
		"status":   "provisioning",
		"message":  "Your computer is being created. This takes about 60 seconds.",
	})
}

// UpgradeComputer handles PUT /api/computer/upgrade.
// Accepts {"plan": "growth"} and returns an updated computer.
func (h *ComputerHandler) UpgradeComputer(c *gin.Context) {
	var req upgradeComputerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, slog.Default(), err)
		return
	}

	comp := mockComputer(req.Plan)
	comp.UpdatedAt = time.Now().UTC()

	slog.InfoContext(c.Request.Context(), "computer: upgrade requested", "plan", req.Plan)

	c.JSON(http.StatusOK, gin.H{
		"computer": comp,
		"message":  "Plan upgraded to " + req.Plan + ".",
	})
}

// DeleteComputer handles DELETE /api/computer.
// Returns a terminated status response.
func (h *ComputerHandler) DeleteComputer(c *gin.Context) {
	slog.InfoContext(c.Request.Context(), "computer: terminate requested")
	c.JSON(http.StatusOK, gin.H{"status": "terminated"})
}

// GetMetrics handles GET /api/computer/metrics.
// Returns realistic resource usage metrics.
func (h *ComputerHandler) GetMetrics(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"cpu_percent":      23.5,
		"ram_used_mb":      2048,
		"ram_total_mb":     4096,
		"storage_used_gb":  3.2,
		"storage_total_gb": 10,
		"network_in_mb":    145.2,
		"network_out_mb":   89.7,
		"uptime_seconds":   86400,
	})
}

// GetRuntimes handles GET /api/computer/runtimes.
// Returns the list of processes running inside the computer.
func (h *ComputerHandler) GetRuntimes(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"runtimes": defaultRuntimes,
		"count":    len(defaultRuntimes),
	})
}

// StartRuntime handles POST /api/computer/runtimes/:name/start.
func (h *ComputerHandler) StartRuntime(c *gin.Context) {
	name := c.Param("name")
	slog.InfoContext(c.Request.Context(), "computer: runtime start requested", "name", name)
	c.JSON(http.StatusAccepted, gin.H{
		"status": "starting",
		"name":   name,
	})
}

// StopRuntime handles POST /api/computer/runtimes/:name/stop.
func (h *ComputerHandler) StopRuntime(c *gin.Context) {
	name := c.Param("name")
	slog.InfoContext(c.Request.Context(), "computer: runtime stop requested", "name", name)
	c.JSON(http.StatusOK, gin.H{
		"status": "stopped",
		"name":   name,
	})
}
