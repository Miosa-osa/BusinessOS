package handlers

// PM4PyDashboardHandler aggregates process mining KPIs from pm4py-rust in a single call.
//
// POST /api/pm4py/dashboard-kpi
//
// Calls pm4py-rust /api/statistics and /api/conformance/token-replay concurrently,
// then merges results into a single KPI response for the 4 new KPI dashboard widgets.
//
// WvdA: both concurrent calls have a shared 30s context timeout (deadlock freedom).
// Armstrong: stats error surfaces as 502 — no silent swallow.

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/integrations/pm4py_rust"
)

const pm4pyDashboardTimeout = 30 * time.Second

// ProcessMiningKPIRequest accepts raw event log JSON and an optional petri net.
type ProcessMiningKPIRequest struct {
	EventLog json.RawMessage `json:"event_log" binding:"required"`
	PetriNet json.RawMessage `json:"petri_net,omitempty"`
}

// ProcessMiningKPIResponse aggregates all process mining KPIs for dashboard widgets.
type ProcessMiningKPIResponse struct {
	// Conformance KPIs (widget 1 + 2)
	ConformanceFitness   float64 `json:"conformance_fitness"`
	ConformancePrecision float64 `json:"conformance_precision"`
	IsConformant         bool    `json:"is_conformant"`

	// Variant KPIs (widget 3)
	VariantCount int            `json:"variant_count"`
	TopVariants  []VariantEntry `json:"top_variants"`

	// Activity / bottleneck KPIs (widget 4)
	BottleneckActivities []BottleneckEntry `json:"bottleneck_activities"`
	ActivityFrequencies  map[string]int    `json:"activity_frequencies"`

	// Event log metadata
	EventCount int `json:"event_count"`
	TraceCount int `json:"trace_count"`

	// Metadata
	FetchedAt string `json:"fetched_at"`
}

// VariantEntry describes one process variant and its relative frequency.
type VariantEntry struct {
	Label      string  `json:"label"`
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
}

// BottleneckEntry describes one activity identified as a bottleneck.
type BottleneckEntry struct {
	Activity  string `json:"activity"`
	Frequency int    `json:"frequency"`
}

// PM4PyDashboardHandler handles POST /api/pm4py/dashboard-kpi.
type PM4PyDashboardHandler struct {
	client  *pm4py_rust.Client
	logger  *slog.Logger
	timeout time.Duration
}

// NewPM4PyDashboardHandler constructs a handler using an existing pm4py-rust client.
func NewPM4PyDashboardHandler(client *pm4py_rust.Client) *PM4PyDashboardHandler {
	return &PM4PyDashboardHandler{
		client:  client,
		logger:  slog.Default(),
		timeout: pm4pyDashboardTimeout,
	}
}

// NewPM4PyDashboardHandlerWithTimeout constructs a handler with a caller-supplied timeout.
// Use this in tests or when the deployment environment requires a non-default deadline.
func NewPM4PyDashboardHandlerWithTimeout(client *pm4py_rust.Client, timeout time.Duration) *PM4PyDashboardHandler {
	return &PM4PyDashboardHandler{
		client:  client,
		logger:  slog.Default(),
		timeout: timeout,
	}
}

// GetDashboardKPI handles POST /api/pm4py/dashboard-kpi.
// It fans out to pm4py-rust statistics + conformance concurrently, then merges results.
func (h *PM4PyDashboardHandler) GetDashboardKPI(c *gin.Context) {
	var req ProcessMiningKPIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "event_log is required"})
		return
	}

	// WvdA: shared deadline covers both goroutines.
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.timeout)
	defer cancel()

	var (
		statsResult *pm4py_rust.StatisticsResponse
		statsErr    error
		confResult  *pm4py_rust.ConformanceResponse
		confErr     error
		wg          sync.WaitGroup
	)

	wg.Add(2)

	go func() {
		defer wg.Done()
		statsResult, statsErr = h.callStatistics(ctx, req.EventLog)
	}()

	go func() {
		defer wg.Done()
		confResult, confErr = h.callConformance(ctx, req.EventLog, req.PetriNet)
	}()

	wg.Wait()

	// Armstrong: surface stats failure visibly rather than returning partial/corrupt data.
	if statsErr != nil {
		h.logger.ErrorContext(ctx, "pm4py dashboard: statistics call failed", "error", statsErr)
		c.JSON(http.StatusBadGateway, gin.H{"error": "pm4py statistics unavailable"})
		return
	}

	// Conformance errors are non-fatal (petri net may be absent or pm4py unavailable).
	if confErr != nil {
		h.logger.WarnContext(ctx, "pm4py dashboard: conformance call failed (non-fatal)", "error", confErr)
	}

	resp := h.buildResponse(statsResult, confResult)
	c.JSON(http.StatusOK, resp)
}

// callStatistics requests statistics from pm4py-rust.
// Returns nil+error on failure; caller decides whether to surface as 502.
func (h *PM4PyDashboardHandler) callStatistics(ctx context.Context, eventLog json.RawMessage) (*pm4py_rust.StatisticsResponse, error) {
	if h.client == nil {
		return nil, nil
	}
	return h.client.StatisticsRequest(ctx, &pm4py_rust.StatisticsRequest{
		EventLog:           eventLog,
		IncludeVariants:    true,
		IncludeBottlenecks: true,
	})
}

// callConformance requests token-replay conformance from pm4py-rust.
// Returns nil+nil when no petri net is supplied (graceful degradation).
func (h *PM4PyDashboardHandler) callConformance(ctx context.Context, eventLog, petriNet json.RawMessage) (*pm4py_rust.ConformanceResponse, error) {
	if h.client == nil {
		return nil, nil
	}
	if len(petriNet) == 0 || string(petriNet) == "null" || string(petriNet) == "" {
		return nil, nil // No petri net — skip conformance gracefully.
	}

	// Unmarshal the raw PetriNet JSON into the typed struct the client expects.
	var pn pm4py_rust.PetriNetJSON
	if err := json.Unmarshal(petriNet, &pn); err != nil {
		return nil, err
	}

	return h.client.ConformanceRequest(ctx, &pm4py_rust.ConformanceRequest{
		EventLog: eventLog,
		PetriNet: pn,
	})
}

// buildResponse merges statistics and (optional) conformance into one KPI payload.
func (h *PM4PyDashboardHandler) buildResponse(
	stats *pm4py_rust.StatisticsResponse,
	conf *pm4py_rust.ConformanceResponse,
) ProcessMiningKPIResponse {

	resp := ProcessMiningKPIResponse{
		TopVariants:          []VariantEntry{},
		BottleneckActivities: []BottleneckEntry{},
		ActivityFrequencies:  map[string]int{},
		FetchedAt:            time.Now().UTC().Format(time.RFC3339),
	}

	if stats != nil {
		resp.EventCount = stats.EventCount
		resp.TraceCount = stats.TraceCount
		resp.VariantCount = stats.VariantCount

		if stats.ActivityFrequencies != nil {
			resp.ActivityFrequencies = stats.ActivityFrequencies
		}

		// Build top variants list from variant frequencies map.
		if len(stats.VariantFrequencies) > 0 {
			totalTraces := stats.TraceCount
			if totalTraces == 0 {
				totalTraces = 1 // avoid division by zero
			}
			for label, count := range stats.VariantFrequencies {
				resp.TopVariants = append(resp.TopVariants, VariantEntry{
					Label:      label,
					Count:      count,
					Percentage: float64(count) / float64(totalTraces) * 100.0,
				})
			}
		}

		// Map bottleneck activity names to BottleneckEntry with frequency.
		for _, activity := range stats.BottleneckActivities {
			freq := 0
			if stats.ActivityFrequencies != nil {
				freq = stats.ActivityFrequencies[activity]
			}
			resp.BottleneckActivities = append(resp.BottleneckActivities, BottleneckEntry{
				Activity:  activity,
				Frequency: freq,
			})
		}
	}

	if conf != nil {
		resp.ConformanceFitness = conf.Fitness
		resp.ConformancePrecision = conf.Precision
		resp.IsConformant = conf.IsConformant
	}

	return resp
}
