package handlers

import (
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
)

// registerOptimalRoutes wires up all OptimalOS data endpoints at
// /api/optimal/*. The runtime is chosen at startup:
//
//	OPTIMAL_ENGINE_URL set     → Elixir engine via reverse proxy (canonical)
//	OPTIMAL_ENGINE_URL unset   → in-process Go port (fallback)
//
// The proxy mode is the production path — it delegates every request to
// the live Elixir OptimalEngine running on http://localhost:4200 (default)
// so the engine's classify / route / ingest / RAG / wiki logic runs in its
// native runtime. The Go port stays available for environments that don't
// want a second process.
//
// These routes are intentionally unauthenticated for the Tauri desktop
// scenario; in cloud deployments tighten via the parent api group's auth
// middleware.
func (h *Handlers) registerOptimalRoutes(api *gin.RouterGroup) {
	if engineURL := os.Getenv("OPTIMAL_ENGINE_URL"); engineURL != "" {
		proxy, err := NewOptimalProxy(engineURL)
		if err != nil {
			slog.Error("OptimalOS proxy: invalid OPTIMAL_ENGINE_URL — falling back to in-process Go port",
				"engine_url", engineURL, "error", err)
		} else {
			RegisterOptimalProxyRoutes(api, proxy)
			slog.Info("OptimalOS routes proxied to live Elixir engine",
				"engine_url", engineURL)
			return
		}
	}

	if h.optimalHandler == nil {
		slog.Warn("OptimalOS routes skipped: optimalHandler not initialized and OPTIMAL_ENGINE_URL unset")
		return
	}
	RegisterOptimalRoutes(api, h.optimalHandler)
	slog.Info("OptimalOS routes served by in-process Go port (set OPTIMAL_ENGINE_URL to switch to live Elixir engine)")
}

// registerOptimalEngineRoutes mounts /api/optimal/{ask,recall/*,memory/*}
// and /api/brief/{morning,evening}. Distinct from registerOptimalRoutes
// (which is the catch-all reverse proxy) — these are typed BO endpoints
// that call the engine's HTTP API via the shared optimalengine.Client
// and inject BusinessOS user context (auth, default workspace, etc.).
func (h *Handlers) registerOptimalEngineRoutes(api *gin.RouterGroup, auth gin.HandlerFunc) {
	handler := h.buildOptimalEngineHandler()
	RegisterOptimalEngineRoutes(api, handler, auth)
	if handler != nil && handler.engine != nil && handler.engine.Enabled() {
		slog.Info("Engine-backed routes registered",
			"routes", "/api/engine/{ask,recall/*,memory}, /api/brief/{morning,evening}")
	} else {
		slog.Info("Engine-backed routes registered (will return 503 — OPTIMAL_ENGINE_URL not set)")
	}
}
