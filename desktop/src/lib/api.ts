/**
 * Client for the Optimal Engine HTTP API.
 *
 * Endpoints implemented:
 *   GET  /api/status             — engine readiness
 *   GET  /api/metrics            — counters + histograms
 *   GET  /api/search?q=…         — hybrid search hits
 *   POST /api/rag                — wiki-first retrieval envelope
 *   GET  /api/wiki               — list wiki pages
 *   GET  /api/wiki/:slug         — render a single page
 *
 * In dev the Vite proxy forwards `/api` → http://127.0.0.1:4200.
 * In a Tauri build we use `VITE_ENGINE_URL` (defaults to the same host).
 */

const BASE = import.meta.env.VITE_ENGINE_URL ?? "";

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    headers: { "content-type": "application/json" },
    ...init,
  });

  if (!res.ok) {
    throw new Error(`HTTP ${res.status}: ${await res.text()}`);
  }

  return (await res.json()) as T;
}

export interface StatusPayload {
  status: "up" | "degraded" | "down";
  "ok?": boolean;
  checks: Record<string, string>;
  degraded: string[];
}

export interface RagEnvelope {
  body: string;
  format: "plain" | "markdown" | "claude" | "openai";
  sources: string[];
  warnings: string[];
}

export interface RagResult {
  source: "wiki" | "chunks" | "empty";
  envelope: RagEnvelope;
  // Elixir emits keys verbatim including `?` — these match the wire shape.
  trace: {
    "wiki_hit?": boolean;
    n_candidates: number;
    n_delivered: number;
    "truncated?": boolean;
    elapsed_ms: number;
    intent?: unknown;
  };
}

export interface WikiPageSummary {
  slug: string;
  audience: string;
  version: number;
  last_curated: string | null;
  curated_by: string | null;
  size_bytes: number;
}

export async function status(): Promise<StatusPayload> {
  return request<StatusPayload>("/api/status");
}

export async function search(
  q: string,
  limit = 10,
): Promise<{ query: string; results: unknown[] }> {
  const params = new URLSearchParams({ q, limit: String(limit) });
  return request(`/api/search?${params}`);
}

export async function ask(
  query: string,
  opts: { audience?: string; format?: string; bandwidth?: string } = {},
): Promise<RagResult> {
  return request<RagResult>("/api/rag", {
    method: "POST",
    body: JSON.stringify({
      query,
      audience: opts.audience ?? "default",
      format: opts.format ?? "markdown",
      bandwidth: opts.bandwidth ?? "medium",
    }),
  });
}

export async function listWiki(
  tenant = "default",
): Promise<{ pages: WikiPageSummary[] }> {
  return request(`/api/wiki?tenant=${encodeURIComponent(tenant)}`);
}

export async function getWiki(
  slug: string,
  opts: { audience?: string; format?: string } = {},
): Promise<{ slug: string; body: string; warnings: string[] }> {
  const params = new URLSearchParams({
    audience: opts.audience ?? "default",
    format: opts.format ?? "markdown",
  });

  return request(`/api/wiki/${encodeURIComponent(slug)}?${params}`);
}
