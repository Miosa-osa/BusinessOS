import { sveltekit } from "@sveltejs/kit/vite";
import tailwindcss from "@tailwindcss/vite";
import { defineConfig } from "vite";

export default defineConfig(({ mode }) => ({
  plugins: [sveltekit(), tailwindcss()],
  esbuild: {
    drop: mode === "production" ? ["console", "debugger"] : [],
  },
  ssr: {
    // Force these CommonJS modules to be bundled for SSR
    noExternal: ["ms"],
  },
  optimizeDeps: {
    include: ["ms"],
  },
  server: {
    proxy: {
      // Versioned API (v1) - catch-all for all /api/v1/* routes
      // This ensures CSRF cookies work correctly (same origin)
      "/api/v1": {
        target: "http://localhost:8001",
        changeOrigin: true,
      },
      // Auth routes - proxied to Go backend (NOT handled by SvelteKit)
      "/api/auth": {
        target: "http://localhost:8001",
        changeOrigin: true,
      },
      // Terminal WebSocket - requires ws: true for WebSocket upgrade
      // changeOrigin: false preserves original Origin header (localhost:5174)
      // which is already in ALLOWED_ORIGINS - critical for CORS!
      "/api/terminal": {
        target: "http://localhost:8001",
        changeOrigin: false,
        ws: true,
        configure: (proxy, _options) => {
          proxy.on("error", (err, _req, _res) => {
            console.log("[Vite Proxy] Error:", err);
          });
          proxy.on("proxyReq", (proxyReq, req, _res) => {
            console.log("[Vite Proxy] Request:", req.method, req.url);
          });
          proxy.on("proxyReqWs", (proxyReq, req, socket, options, head) => {
            console.log("[Vite Proxy] WebSocket upgrade:", req.url);
          });
        },
      },
      "/api/chat": {
        target: "http://localhost:8001",
        changeOrigin: true,
      },
      "/api/projects": {
        target: "http://localhost:8001",
        changeOrigin: true,
      },
      "/api/contexts": {
        target: "http://localhost:8001",
        changeOrigin: true,
      },
      "/api/team": {
        target: "http://localhost:8001",
        changeOrigin: true,
      },
      "/api/dashboard": {
        target: "http://localhost:8001",
        changeOrigin: true,
      },
      "/api/mcp": {
        target: "http://localhost:8001",
        changeOrigin: true,
      },
      "/api/daily": {
        target: "http://localhost:8001",
        changeOrigin: true,
      },
      "/api/settings": {
        target: "http://localhost:8001",
        changeOrigin: true,
      },
      "/api/artifacts": {
        target: "http://localhost:8001",
        changeOrigin: true,
      },
      "/api/nodes": {
        target: "http://localhost:8001",
        changeOrigin: true,
      },
      "/api/clients": {
        target: "http://localhost:8001",
        changeOrigin: true,
      },
      "/api/deals": {
        target: "http://localhost:8001",
        changeOrigin: true,
      },
      "/api/transcribe": {
        target: "http://localhost:8001",
        changeOrigin: true,
      },
      "/api/voice-notes": {
        target: "http://localhost:8001",
        changeOrigin: true,
      },
      "/api/ai": {
        target: "http://localhost:8001",
        changeOrigin: true,
      },
      "/api/calendar": {
        target: "http://localhost:8001",
        changeOrigin: true,
      },
      "/api/integrations": {
        target: "http://localhost:8001",
        changeOrigin: true,
      },
      "/api/profile": {
        target: "http://localhost:8001",
        changeOrigin: true,
      },
      "/api/filesystem": {
        target: "http://localhost:8001",
        changeOrigin: true,
      },
      "/api/usage": {
        target: "http://localhost:8001",
        changeOrigin: true,
      },
      // OSA Integration APIs (SSE streams need longer timeout + no buffering)
      "/api/osa": {
        target: "http://localhost:8001",
        changeOrigin: true,
        timeout: 600000, // 10 min timeout for SSE streams
        configure: (proxy) => {
          proxy.on("proxyRes", (proxyRes) => {
            // Disable buffering for SSE streams
            if (
              proxyRes.headers["content-type"]?.includes("text/event-stream")
            ) {
              proxyRes.headers["x-accel-buffering"] = "no";
              proxyRes.headers["cache-control"] = "no-cache";
            }
          });
        },
      },
      // Apps API
      "/api/apps": {
        target: "http://localhost:8001",
        changeOrigin: true,
      },
      // Pedro Tasks APIs
      "/api/documents": {
        target: "http://localhost:8001",
        changeOrigin: true,
      },
      "/api/memories": {
        target: "http://localhost:8001",
        changeOrigin: true,
      },
      "/api/learning": {
        target: "http://localhost:8001",
        changeOrigin: true,
      },
      "/api/app-profiles": {
        target: "http://localhost:8001",
        changeOrigin: true,
      },
      "/api/intelligence": {
        target: "http://localhost:8001",
        changeOrigin: true,
      },
      "/api/onboarding": {
        target: "http://localhost:8001",
        changeOrigin: true,
      },
      "/api/users": {
        target: "http://localhost:8001",
        changeOrigin: true,
      },
      "/health": {
        target: "http://localhost:8001",
        changeOrigin: true,
      },
    },
  },
}));
