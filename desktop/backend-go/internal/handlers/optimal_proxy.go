package handlers

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// OptimalProxy forwards every /api/optimal/* request to a live Elixir
// OptimalEngine HTTP server (default port 4200). When wired, it bypasses
// the in-process Go port entirely — the Elixir engine is the runtime
// authority for graph, search, ingest, wiki, and architectures.
//
// Auth still happens at the BusinessOS layer: this handler is mounted
// behind whatever middleware the caller decides. The proxy preserves
// cookies and request headers verbatim so the Elixir engine sees the same
// authenticated context.
type OptimalProxy struct {
	target *url.URL
	rp     *httputil.ReverseProxy
}

// NewOptimalProxy returns a proxy targeted at engineURL (e.g.
// "http://localhost:4200"). The url is parsed at construction so a bad
// value fails fast at startup, not on the first request.
func NewOptimalProxy(engineURL string) (*OptimalProxy, error) {
	if strings.TrimSpace(engineURL) == "" {
		return nil, errors.New("optimal proxy: engine URL is empty")
	}
	u, err := url.Parse(engineURL)
	if err != nil {
		return nil, err
	}
	if u.Scheme == "" || u.Host == "" {
		return nil, errors.New("optimal proxy: engine URL must include scheme + host")
	}

	rp := httputil.NewSingleHostReverseProxy(u)

	// Surface upstream failures with structured slog instead of the default
	// "http: proxy error: …" line. Operators triaging a Pages graph stall
	// can grep for "optimal proxy:" and see the engine URL + cause.
	rp.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		slog.ErrorContext(r.Context(), "optimal proxy: upstream call failed",
			"engine_url", u.String(),
			"path", r.URL.Path,
			"method", r.Method,
			"error", err,
		)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		_, _ = io.WriteString(w,
			`{"error":"optimal engine unreachable","upstream":"`+u.String()+`"}`)
	}

	// Tighten the upstream timeout so a hung Elixir process doesn't tie up
	// our request goroutines. The default Transport has no overall timeout.
	rp.Transport = &http.Transport{
		ResponseHeaderTimeout: 30 * time.Second,
	}

	return &OptimalProxy{target: u, rp: rp}, nil
}

// RegisterOptimalProxyRoutes mounts a single catch-all route at
// /api/optimal/*path that forwards to the Elixir engine. Mounted INSTEAD OF
// the in-process Go handler when OPTIMAL_ENGINE_URL is configured.
//
// We deliberately don't preserve any of the Go handler's per-route auth
// or rate limits — those decisions belong to the engine itself. BusinessOS
// only enforces session auth via the api group's middleware chain.
func RegisterOptimalProxyRoutes(api *gin.RouterGroup, p *OptimalProxy) {
	api.Any("/optimal/*path", p.Forward)
}

// Forward is the single proxy entrypoint. We bind it explicitly (rather
// than calling rp.ServeHTTP from a closure) so middleware can decorate it
// like any other handler.
func (p *OptimalProxy) Forward(c *gin.Context) {
	// Inject a per-request deadline so a slow upstream can't dangle.
	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()
	p.rp.ServeHTTP(c.Writer, c.Request.WithContext(ctx))
}
