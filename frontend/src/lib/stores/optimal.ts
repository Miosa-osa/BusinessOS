import { writable } from "svelte/store";
import { browser } from "$app/environment";
import { getApiBaseUrl, getCSRFToken } from "$lib/api/base";

// ─── Types ───────────────────────────────────────────────────────────────────

export interface OptimalNode {
  slug: string;
  name: string;
  type: string;
  context_md: string;
  signal_md: string;
  signal_count: number;
  has_signals: boolean;
}

export interface FileEntry {
  name: string;
  path: string;
  is_dir: boolean;
  size: number;
  children?: FileEntry[];
}

export interface SelectedFile {
  content: string;
  path: string;
  slug: string;
}

export interface OptimalSignalFile {
  slug: string;
  date: string;
  name: string;
  path: string;
  content: string;
}

export interface OptimalSearchResult {
  path: string;
  abstract: string;
  score: number;
}

export interface RhythmDay {
  date: string;
  content: string;
}

// ─── Store State ─────────────────────────────────────────────────────────────

interface OptimalState {
  nodes: OptimalNode[];
  selectedNode: OptimalNode | null;
  selectedNodeSignals: OptimalSignalFile[];
  todayRhythm: RhythmDay | null;
  searchResults: OptimalSearchResult[];
  fileTree: FileEntry[];
  selectedFile: SelectedFile | null;
  loading: boolean;
  error: string | null;
}

// ─── Store ───────────────────────────────────────────────────────────────────

function createOptimalStore() {
  const initialState: OptimalState = {
    nodes: [],
    selectedNode: null,
    selectedNodeSignals: [],
    todayRhythm: null,
    searchResults: [],
    fileTree: [],
    selectedFile: null,
    loading: false,
    error: null,
  };

  const { subscribe, update } = writable<OptimalState>(initialState);

  function buildHeaders(includeContentType = false): Record<string, string> {
    const headers: Record<string, string> = {};
    const csrfToken = getCSRFToken();
    if (csrfToken) headers["X-CSRF-Token"] = csrfToken;
    if (includeContentType) headers["Content-Type"] = "application/json";
    return headers;
  }

  return {
    subscribe,

    /** Fetch all OptimalOS nodes — GET /api/optimal/nodes */
    async loadNodes(): Promise<void> {
      if (!browser) return;

      update((s) => ({ ...s, loading: true, error: null }));

      try {
        const res = await fetch(`${getApiBaseUrl()}/optimal/nodes`, {
          method: "GET",
          headers: buildHeaders(),
          credentials: "include",
          signal: AbortSignal.timeout(8000),
        });

        if (!res.ok) throw new Error(`HTTP ${res.status}`);

        const data: { nodes?: OptimalNode[] } = await res.json();
        const nodes = Array.isArray(data.nodes) ? data.nodes : [];

        update((s) => ({ ...s, nodes, loading: false }));
      } catch (err) {
        const message =
          err instanceof Error ? err.message : "Failed to load nodes";
        update((s) => ({ ...s, loading: false, error: message }));
      }
    },

    /**
     * Select a node by slug — GET /api/optimal/nodes/:slug
     * Also loads signals for the selected node.
     */
    async selectNode(slug: string): Promise<void> {
      if (!browser) return;

      update((s) => ({
        ...s,
        loading: true,
        error: null,
        selectedNode: null,
        selectedNodeSignals: [],
      }));

      try {
        const res = await fetch(
          `${getApiBaseUrl()}/optimal/nodes/${encodeURIComponent(slug)}`,
          {
            method: "GET",
            headers: buildHeaders(),
            credentials: "include",
            signal: AbortSignal.timeout(8000),
          },
        );

        if (!res.ok) throw new Error(`HTTP ${res.status}`);

        const selectedNode: OptimalNode = await res.json();

        update((s) => ({ ...s, selectedNode, loading: false }));
      } catch (err) {
        const message =
          err instanceof Error ? err.message : "Failed to load node";
        update((s) => ({ ...s, loading: false, error: message }));
        return;
      }

      // Load signals in parallel after node is resolved
      await this.loadNodeSignals(slug);
    },

    /** Load signal files for a node — GET /api/optimal/nodes/:slug/signals */
    async loadNodeSignals(slug: string): Promise<void> {
      if (!browser) return;

      try {
        const res = await fetch(
          `${getApiBaseUrl()}/optimal/nodes/${encodeURIComponent(slug)}/signals`,
          {
            method: "GET",
            headers: buildHeaders(),
            credentials: "include",
            signal: AbortSignal.timeout(8000),
          },
        );

        if (!res.ok) throw new Error(`HTTP ${res.status}`);

        const data: { signals?: OptimalSignalFile[] } = await res.json();
        const signals = Array.isArray(data.signals) ? data.signals : [];

        update((s) => ({ ...s, selectedNodeSignals: signals }));
      } catch (err) {
        const message =
          err instanceof Error ? err.message : "Failed to load signals";
        update((s) => ({ ...s, error: message }));
      }
    },

    /** Load today's rhythm file — GET /api/optimal/rhythm/today */
    async loadTodayRhythm(): Promise<void> {
      if (!browser) return;

      try {
        const res = await fetch(`${getApiBaseUrl()}/optimal/rhythm/today`, {
          method: "GET",
          headers: buildHeaders(),
          credentials: "include",
          signal: AbortSignal.timeout(5000),
        });

        if (!res.ok) throw new Error(`HTTP ${res.status}`);

        const todayRhythm: RhythmDay = await res.json();

        update((s) => ({ ...s, todayRhythm }));
      } catch {
        // Today's rhythm is optional — silently ignore missing/unavailable
      }
    },

    /** Search the OptimalOS context engine — POST /api/optimal/search */
    async search(query: string, limit = 10): Promise<void> {
      if (!browser || !query.trim()) return;

      update((s) => ({ ...s, loading: true, error: null, searchResults: [] }));

      try {
        const res = await fetch(`${getApiBaseUrl()}/optimal/search`, {
          method: "POST",
          headers: buildHeaders(true),
          credentials: "include",
          signal: AbortSignal.timeout(10000),
          body: JSON.stringify({ query: query.trim(), limit }),
        });

        if (!res.ok) throw new Error(`HTTP ${res.status}`);

        const data: { results?: OptimalSearchResult[] } = await res.json();
        const searchResults = Array.isArray(data.results) ? data.results : [];

        update((s) => ({ ...s, searchResults, loading: false }));
      } catch (err) {
        const message = err instanceof Error ? err.message : "Search failed";
        update((s) => ({ ...s, loading: false, error: message }));
      }
    },

    /** Ingest a signal into the OptimalOS engine — POST /api/optimal/ingest */
    async ingest(text: string, genre?: string): Promise<void> {
      if (!browser || !text.trim()) return;

      update((s) => ({ ...s, loading: true, error: null }));

      try {
        const body: Record<string, string> = { text: text.trim() };
        if (genre) body.genre = genre;

        const res = await fetch(`${getApiBaseUrl()}/optimal/ingest`, {
          method: "POST",
          headers: buildHeaders(true),
          credentials: "include",
          signal: AbortSignal.timeout(15000),
          body: JSON.stringify(body),
        });

        if (!res.ok) throw new Error(`HTTP ${res.status}`);

        update((s) => ({ ...s, loading: false }));
      } catch (err) {
        const message = err instanceof Error ? err.message : "Ingest failed";
        update((s) => ({ ...s, loading: false, error: message }));
      }
    },

    /** Load file tree for a node — GET /api/optimal/nodes/:slug/files */
    async loadFileTree(slug: string): Promise<void> {
      if (!browser) return;

      update((s) => ({
        ...s,
        loading: true,
        error: null,
        fileTree: [],
        selectedFile: null,
      }));

      try {
        const res = await fetch(
          `${getApiBaseUrl()}/optimal/nodes/${encodeURIComponent(slug)}/files`,
          {
            method: "GET",
            headers: buildHeaders(),
            credentials: "include",
            signal: AbortSignal.timeout(8000),
          },
        );

        if (!res.ok) throw new Error(`HTTP ${res.status}`);

        const data: { files?: FileEntry[] } = await res.json();
        const fileTree = Array.isArray(data.files) ? data.files : [];

        update((s) => ({ ...s, fileTree, loading: false }));
      } catch (err) {
        const message =
          err instanceof Error ? err.message : "Failed to load file tree";
        update((s) => ({ ...s, loading: false, error: message }));
      }
    },

    /** Load a single file — GET /api/optimal/nodes/:slug/file/*filepath */
    async loadFile(slug: string, filePath: string): Promise<void> {
      if (!browser) return;

      update((s) => ({ ...s, loading: true, error: null }));

      try {
        // Strip the node slug prefix from the path if present
        // (file tree returns "04-ai-masters/community/file.md" but API expects "community/file.md")
        let cleanPath = filePath;
        if (cleanPath.startsWith(slug + "/")) {
          cleanPath = cleanPath.slice(slug.length + 1);
        }
        const encoded = cleanPath
          .split("/")
          .map((seg) => encodeURIComponent(seg))
          .join("/");

        const res = await fetch(
          `${getApiBaseUrl()}/optimal/nodes/${encodeURIComponent(slug)}/file/${encoded}`,
          {
            method: "GET",
            headers: buildHeaders(),
            credentials: "include",
            signal: AbortSignal.timeout(8000),
          },
        );

        if (!res.ok) throw new Error(`HTTP ${res.status}`);

        const selectedFile: SelectedFile = await res.json();

        update((s) => ({ ...s, selectedFile, loading: false }));
      } catch (err) {
        const message =
          err instanceof Error ? err.message : "Failed to load file";
        update((s) => ({ ...s, loading: false, error: message }));
      }
    },

    /** Clear the currently selected file */
    clearSelectedFile(): void {
      update((s) => ({ ...s, selectedFile: null }));
    },

    /** Clear search results */
    clearSearch(): void {
      update((s) => ({ ...s, searchResults: [] }));
    },

    /** Clear selected node and its signals */
    clearSelectedNode(): void {
      update((s) => ({
        ...s,
        selectedNode: null,
        selectedNodeSignals: [],
      }));
    },

    /** Dismiss the current error */
    clearError(): void {
      update((s) => ({ ...s, error: null }));
    },
  };
}

export const optimalStore = createOptimalStore();
