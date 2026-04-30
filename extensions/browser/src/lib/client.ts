import type {
  ClipPayload,
  Memory,
  ProfileSummary,
  Result,
  SearchResult,
  Workspace,
} from "./types";
import { getSettings } from "./storage";

// ─── Internal helpers ─────────────────────────────────────────────────────────

async function buildHeaders(): Promise<Record<string, string>> {
  const settings = await getSettings();
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    Accept: "application/json",
  };
  if (settings.apiKey) {
    headers["X-API-Key"] = settings.apiKey;
  }
  return headers;
}

async function baseUrl(): Promise<string> {
  const settings = await getSettings();
  return settings.engineUrl.replace(/\/$/, "");
}

async function get<T>(path: string): Promise<Result<T>> {
  try {
    const [base, headers] = await Promise.all([baseUrl(), buildHeaders()]);
    const res = await fetch(`${base}${path}`, { method: "GET", headers });
    if (!res.ok) {
      const text = await res.text().catch(() => res.statusText);
      return { ok: false, error: `HTTP ${res.status}: ${text}` };
    }
    const data = (await res.json()) as T;
    return { ok: true, value: data };
  } catch (err) {
    return {
      ok: false,
      error: err instanceof Error ? err.message : "Network error",
    };
  }
}

async function post<T>(path: string, body: unknown): Promise<Result<T>> {
  try {
    const [base, headers] = await Promise.all([baseUrl(), buildHeaders()]);
    const res = await fetch(`${base}${path}`, {
      method: "POST",
      headers,
      body: JSON.stringify(body),
    });
    if (!res.ok) {
      const text = await res.text().catch(() => res.statusText);
      return { ok: false, error: `HTTP ${res.status}: ${text}` };
    }
    const data = (await res.json()) as T;
    return { ok: true, value: data };
  } catch (err) {
    return {
      ok: false,
      error: err instanceof Error ? err.message : "Network error",
    };
  }
}

// ─── Public API ───────────────────────────────────────────────────────────────

/**
 * List workspaces for a given tenant (optional).
 */
export async function listWorkspaces(
  tenant?: string,
): Promise<Result<Workspace[]>> {
  const qs = tenant ? `?tenant=${encodeURIComponent(tenant)}` : "";
  return get<Workspace[]>(`/api/workspaces${qs}`);
}

/**
 * Search memories in a workspace.
 */
export async function searchMemories(
  query: string,
  workspace: string,
  limit = 20,
): Promise<Result<SearchResult[]>> {
  const qs = new URLSearchParams({
    q: query,
    workspace,
    limit: String(limit),
  });
  return get<SearchResult[]>(`/api/search?${qs}`);
}

/**
 * Clip content to the engine as a new memory.
 */
export async function clipMemory(
  payload: ClipPayload,
): Promise<Result<Memory>> {
  return post<Memory>("/api/memory", payload);
}

/**
 * Fetch recent memories for a workspace.
 */
export async function listMemories(
  workspace: string,
  limit = 10,
): Promise<Result<Memory[]>> {
  const qs = new URLSearchParams({ workspace, limit: String(limit) });
  return get<Memory[]>(`/api/memory?${qs}`);
}

/**
 * Get workspace profile summary (active wiki page, memory count, etc.).
 */
export async function getProfile(
  workspace: string,
): Promise<Result<ProfileSummary>> {
  const qs = new URLSearchParams({ workspace, bandwidth: "l1" });
  return get<ProfileSummary>(`/api/profile?${qs}`);
}
