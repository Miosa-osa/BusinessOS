// Backend URLs
const LOCAL_BACKEND_URL = "http://localhost:8001";
const CLOUD_RUN_URL = "https://businessos-api-460433387676.us-central1.run.app";

// API Version (centralized configuration)
const API_VERSION = "v1";

// Shared fetch logic copied from the original ApiClient.request implementation
function getApiBase(): string {
  if (typeof window === "undefined") {
    const result = import.meta.env.VITE_API_URL || `/api/${API_VERSION}`;
    return result;
  }

  const isElectron = "electron" in window;
  const isDev =
    window.location.hostname === "localhost" ||
    window.location.hostname === "127.0.0.1";

  if (isElectron) {
    const mode = localStorage.getItem("businessos_mode");
    let cloudUrl = localStorage.getItem("businessos_cloud_url");

    // Auto-configure URL if not set
    if (!cloudUrl) {
      cloudUrl = isDev ? LOCAL_BACKEND_URL : CLOUD_RUN_URL;
      localStorage.setItem("businessos_cloud_url", cloudUrl);
    }

    if (mode === "cloud" && cloudUrl) {
      const result = `${cloudUrl}/api/${API_VERSION}`;
      return result;
    } else if (mode === "local") {
      return `http://localhost:18080/api/${API_VERSION}`;
    }
    const result = `${cloudUrl}/api/${API_VERSION}`;
    return result;
  }

  // Web app: use env var, or auto-detect based on environment
  if (import.meta.env.VITE_API_URL) {
    return import.meta.env.VITE_API_URL;
  }
  // In development, use relative URLs through Vite proxy to ensure CSRF cookies work
  // (same origin for cookie set and API calls)
  const result = isDev
    ? `/api/${API_VERSION}`
    : `${CLOUD_RUN_URL}/api/${API_VERSION}`;
  return result;
}

export const getApiBaseUrl = () => getApiBase();
export const API_BASE = getApiBase();

// Get CSRF token from cookie
export function getCSRFToken(): string | null {
  if (typeof document === "undefined") return null;

  const cookies = document.cookie.split(";");
  for (const cookie of cookies) {
    const trimmed = cookie.trim();
    const eqIndex = trimmed.indexOf("=");
    if (eqIndex === -1) continue;

    const name = trimmed.substring(0, eqIndex);
    const value = trimmed.substring(eqIndex + 1); // Get everything after first =

    if (name === "csrf_token") {
      return value;
    }
  }
  return null;
}

// Initialize CSRF token by calling the backend endpoint
// This should be called before any state-changing requests (POST, PUT, DELETE)
export async function initCSRF(): Promise<void> {
  if (typeof window === "undefined") {
    return;
  }

  try {
    // Use relative URL to go through Vite proxy in development
    // This ensures the CSRF cookie is set in the correct domain context (localhost:5173)
    // and can be read by subsequent fetch calls
    const isDev =
      window.location.hostname === "localhost" ||
      window.location.hostname === "127.0.0.1";
    const csrfUrl = isDev ? "/api/auth/csrf" : `${getApiBaseUrl()}/auth/csrf`;

    const response = await fetch(csrfUrl, {
      method: "GET",
      credentials: "include",
    });

    if (response.ok) {
      await response.json();
      if (import.meta.env.DEV) {
        console.log(
          "[CSRF] Token initialized:",
          document.cookie.includes("csrf_token")
            ? "cookie set"
            : "cookie missing",
        );
      }
    } else {
      console.error("[CSRF] Init failed, status:", response.status);
    }
  } catch (error) {
    console.warn("[CSRF] Failed to initialize CSRF token:", error);
  }
}

// Add CSRF token to headers for state-changing requests
function addCSRFToken(
  method: string,
  headers: Record<string, string>,
): Record<string, string> {
  const stateChangingMethods = ["POST", "PUT", "PATCH", "DELETE"];
  if (stateChangingMethods.includes(method.toUpperCase())) {
    const csrfToken = getCSRFToken();
    if (csrfToken) {
      headers["X-CSRF-Token"] = csrfToken;
    }
  }
  return headers;
}

/**
 * Get the backend server base URL (without /api suffix)
 * Use this for image URLs and other non-API resources
 */
export function getBackendUrl(): string {
  if (typeof window === "undefined") {
    return "";
  }

  const isElectron = "electron" in window;
  const isDev =
    window.location.hostname === "localhost" ||
    window.location.hostname === "127.0.0.1";

  if (isElectron) {
    const mode = localStorage.getItem("businessos_mode");
    let cloudUrl = localStorage.getItem("businessos_cloud_url");

    if (!cloudUrl) {
      cloudUrl = isDev ? LOCAL_BACKEND_URL : CLOUD_RUN_URL;
    }

    if (mode === "cloud" && cloudUrl) {
      return cloudUrl;
    } else if (mode === "local") {
      return "http://localhost:18080";
    }
    return cloudUrl;
  }

  // Web app: use env var base, or auto-detect based on environment
  if (import.meta.env.VITE_BACKEND_URL) {
    return import.meta.env.VITE_BACKEND_URL;
  }
  return isDev ? LOCAL_BACKEND_URL : CLOUD_RUN_URL;
}

export interface RequestOptions {
  method?: string;
  body?: unknown;
  headers?: Record<string, string>;
  timeout?: number; // Timeout in milliseconds
  skipCache?: boolean; // Bypass the GET response cache
  skipAuthRedirect?: boolean; // Don't auto-redirect to /login on 401
}

// ============================================================================
// Pagination Types
// ============================================================================

export interface PaginationParams {
  page?: number;
  page_size?: number;
}

export interface PaginationMeta {
  page: number;
  page_size: number;
  total_items: number;
  total_pages: number;
  has_more: boolean;
}

export interface PaginatedResponse<T> {
  data: T[];
  pagination: PaginationMeta;
}

/**
 * Build a query string from pagination params, appending to an existing endpoint.
 * Skips undefined values.
 */
export function buildPaginatedEndpoint(
  endpoint: string,
  params: PaginationParams,
): string {
  const qs = new URLSearchParams();
  if (params.page !== undefined) qs.set("page", String(params.page));
  if (params.page_size !== undefined)
    qs.set("page_size", String(params.page_size));
  const queryString = qs.toString();
  return queryString ? `${endpoint}?${queryString}` : endpoint;
}

/**
 * Make a paginated GET request.
 *
 * Handles both paginated (new backend format) and non-paginated (legacy) responses
 * for backward compatibility. When the response is a plain array or an object
 * without a `pagination` field, it wraps it into a PaginatedResponse so callers
 * always receive the same shape.
 *
 * @param endpoint - API path (e.g. "/osa/module-instances")
 * @param params   - Pagination params (page, page_size)
 * @param dataKey  - Key in non-paginated responses that holds the array (e.g. "apps")
 */
export async function requestPaginated<T>(
  endpoint: string,
  params: PaginationParams = {},
  dataKey?: string,
): Promise<PaginatedResponse<T>> {
  const paginatedEndpoint = buildPaginatedEndpoint(endpoint, params);

  // Use skipCache:true for paginated requests so page changes always hit the
  // network rather than returning stale cached results.
  const raw = await request<unknown>(paginatedEndpoint, { skipCache: true });

  // New paginated format: { data: T[], pagination: { ... } }
  if (
    raw !== null &&
    typeof raw === "object" &&
    !Array.isArray(raw) &&
    "pagination" in raw &&
    "data" in raw
  ) {
    return raw as PaginatedResponse<T>;
  }

  // Legacy array format: T[]
  if (Array.isArray(raw)) {
    return {
      data: raw as T[],
      pagination: {
        page: 1,
        page_size: raw.length,
        total_items: raw.length,
        total_pages: 1,
        has_more: false,
      },
    };
  }

  // Legacy object format: { [dataKey]: T[], ... }
  if (
    raw !== null &&
    typeof raw === "object" &&
    dataKey &&
    dataKey in (raw as Record<string, unknown>)
  ) {
    const items = ((raw as Record<string, unknown>)[dataKey] as T[]) ?? [];
    return {
      data: items,
      pagination: {
        page: 1,
        page_size: items.length,
        total_items: items.length,
        total_pages: 1,
        has_more: false,
      },
    };
  }

  // Fallback: treat entire response as a single-page result
  return {
    data: [],
    pagination: {
      page: 1,
      page_size: 0,
      total_items: 0,
      total_pages: 0,
      has_more: false,
    },
  };
}

// ============================================================================
// GET Request Cache — prevents re-fetching the same data during tab switches
// ============================================================================
const GET_CACHE_TTL_MS = 30_000; // 30 seconds
const getCache = new Map<string, { data: unknown; expiry: number }>();
const inflightRequests = new Map<string, Promise<unknown>>();

function getCacheKey(endpoint: string): string {
  return endpoint;
}

/**
 * Invalidate all cached GET responses (call after mutations)
 */
export function invalidateApiCache(): void {
  getCache.clear();
}

/**
 * Invalidate cached GET responses matching a prefix
 */
export function invalidateApiCacheByPrefix(prefix: string): void {
  for (const key of getCache.keys()) {
    if (key.startsWith(prefix)) {
      getCache.delete(key);
    }
  }
}

export async function request<T>(
  endpoint: string,
  options: RequestOptions = {},
): Promise<T> {
  const {
    method = "GET",
    body,
    headers = {},
    timeout = 10_000,
    skipCache = false,
    skipAuthRedirect = false,
  } = options;

  // --- GET cache: return cached data if fresh ---
  const isGet = method.toUpperCase() === "GET";
  if (isGet && !skipCache) {
    const cacheKey = getCacheKey(endpoint);
    const cached = getCache.get(cacheKey);
    if (cached && Date.now() < cached.expiry) {
      return cached.data as T;
    }

    // Deduplicate in-flight GET requests to the same endpoint
    const inflight = inflightRequests.get(cacheKey);
    if (inflight) {
      return inflight as Promise<T>;
    }
  }

  if (body && !headers["Content-Type"]) {
    headers["Content-Type"] = "application/json";
  }

  // Add CSRF token for state-changing requests
  const finalHeaders = addCSRFToken(method, headers);

  const baseUrl = getApiBaseUrl();

  // Create abort controller for timeout
  const controller = new AbortController();
  let timeoutId: NodeJS.Timeout | number | undefined;

  if (timeout && timeout > 0) {
    timeoutId = setTimeout(() => controller.abort(), timeout);
  }

  const doFetch = async (): Promise<T> => {
    try {
      const response = await fetch(`${baseUrl}${endpoint}`, {
        method,
        headers: finalHeaders,
        credentials: "include",
        body: body ? JSON.stringify(body) : undefined,
        signal: controller.signal,
      });

      if (!response.ok) {
        // Handle 401 for non-auth endpoints: session expired, redirect to login
        if (
          response.status === 401 &&
          !endpoint.includes("/auth/") &&
          !skipAuthRedirect
        ) {
          const { clearSession } = await import("$lib/auth-client");
          const { goto } = await import("$app/navigation");
          clearSession();
          goto("/login");
          throw new Error("Session expired");
        }

        const error = await response
          .json()
          .catch(() => ({ detail: "Request failed" }));
        const errorMessage = error.detail || error.message || "Request failed";
        console.error(
          `[API] ${method} ${endpoint} failed with status ${response.status}: ${errorMessage}`,
        );
        throw new Error(`${errorMessage} (HTTP ${response.status})`);
      }

      const data: T = await response.json();

      // Cache successful GET responses
      if (isGet && !skipCache) {
        getCache.set(getCacheKey(endpoint), {
          data,
          expiry: Date.now() + GET_CACHE_TTL_MS,
        });
      }

      // Invalidate cache on mutations so subsequent GETs get fresh data
      if (!isGet) {
        invalidateApiCache();
      }

      return data;
    } catch (error) {
      if (error instanceof Error && error.name === "AbortError") {
        throw new Error(`Request timeout after ${timeout}ms`);
      }
      throw error;
    } finally {
      if (timeoutId !== undefined) {
        clearTimeout(timeoutId as number);
      }
      if (isGet && !skipCache) {
        inflightRequests.delete(getCacheKey(endpoint));
      }
    }
  };

  // Deduplicate in-flight GET requests
  if (isGet && !skipCache) {
    const cacheKey = getCacheKey(endpoint);
    const promise = doFetch();
    inflightRequests.set(cacheKey, promise);
    return promise;
  }

  return doFetch();
}

// For raw response access (like original apiClient)
export const raw = {
  async get(endpoint: string): Promise<Response> {
    return fetch(`${getApiBaseUrl()}${endpoint}`, {
      method: "GET",
      credentials: "include",
    });
  },
  async post(endpoint: string, body?: unknown): Promise<Response> {
    const headers = addCSRFToken("POST", {
      "Content-Type": "application/json",
    });
    return fetch(`${getApiBaseUrl()}${endpoint}`, {
      method: "POST",
      headers,
      credentials: "include",
      body: body ? JSON.stringify(body) : undefined,
    });
  },
  async postFormData(endpoint: string, formData: FormData): Promise<Response> {
    const headers = addCSRFToken("POST", {});
    return fetch(`${getApiBaseUrl()}${endpoint}`, {
      method: "POST",
      headers,
      credentials: "include",
      body: formData,
    });
  },
  async put(endpoint: string, body?: unknown): Promise<Response> {
    const headers = addCSRFToken("PUT", { "Content-Type": "application/json" });
    return fetch(`${getApiBaseUrl()}${endpoint}`, {
      method: "PUT",
      headers,
      credentials: "include",
      body: body ? JSON.stringify(body) : undefined,
    });
  },
  async delete(endpoint: string): Promise<Response> {
    const headers = addCSRFToken("DELETE", {});
    return fetch(`${getApiBaseUrl()}${endpoint}`, {
      method: "DELETE",
      headers,
      credentials: "include",
    });
  },
};
