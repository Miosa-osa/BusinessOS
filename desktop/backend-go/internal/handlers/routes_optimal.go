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
	// Always register the Go handler for frontend-facing routes (/nodes,
	// /dashboard, /rhythm, /team, /projects, /revenue). These read from
	// the local filesystem (OPTIMAL_NODES_ROOT) and are what the SvelteKit
	// frontend expects.
	if h.optimalHandler != nil {
		RegisterOptimalRoutes(api, h.optimalHandler)
		slog.Info("OptimalOS Go handler registered (nodes, dashboard, rhythm)")
	} else {
		slog.Warn("OptimalOS Go handler skipped: optimalHandler not initialized")
	}

	// If the Elixir engine is running, mount engine-native routes that the
	// Go handler doesn't cover (graph analysis, wiki, RAG, grep, health,
	// reindex). These go under /optimal-engine/* to avoid conflicts with
	// the Go handler's /optimal/* routes.
	if engineURL := os.Getenv("OPTIMAL_ENGINE_URL"); engineURL != "" {
		proxy, err := NewOptimalProxy(engineURL)
		if err != nil {
			slog.Error("OptimalOS proxy: invalid OPTIMAL_ENGINE_URL",
				"engine_url", engineURL, "error", err)
		} else {
			api.Any("/optimal-engine/*path", proxy.Forward)
			slog.Info("OptimalOS Elixir engine proxied at /optimal-engine/*",
				"engine_url", engineURL)
		}
	}
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
