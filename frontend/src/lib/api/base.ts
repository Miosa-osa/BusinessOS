// Backend URLs
const LOCAL_BACKEND_URL = 'http://localhost:8080';
const CLOUD_RUN_URL = 'https://businessos-api-460433387676.us-central1.run.app';

// Shared fetch logic copied from the original ApiClient.request implementation
function getApiBase(): string {
  if (typeof window === 'undefined') {
    return import.meta.env.VITE_API_URL || '/api';
  }

  const isElectron = 'electron' in window;
  const isDev = window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1';

  if (isElectron) {
    const mode = localStorage.getItem('businessos_mode');
    let cloudUrl = localStorage.getItem('businessos_cloud_url');

    // Auto-configure URL if not set
    if (!cloudUrl) {
      cloudUrl = isDev ? LOCAL_BACKEND_URL : CLOUD_RUN_URL;
      localStorage.setItem('businessos_cloud_url', cloudUrl);
    }

    if (mode === 'cloud' && cloudUrl) {
      return `${cloudUrl}/api`;
    } else if (mode === 'local') {
      return 'http://localhost:18080/api';
    }
    return `${cloudUrl}/api`;
  }

  // Web app: use env var, or auto-detect based on environment
  if (import.meta.env.VITE_API_URL) {
    return import.meta.env.VITE_API_URL;
  }
  return isDev ? `${LOCAL_BACKEND_URL}/api` : `${CLOUD_RUN_URL}/api`;
}

export const getApiBaseUrl = () => getApiBase();
export const API_BASE = getApiBase();

export interface RequestOptions {
  method?: string;
  body?: unknown;
  headers?: Record<string, string>;
}

export async function request<T>(endpoint: string, options: RequestOptions = {}): Promise<T> {
  const { method = 'GET', body, headers = {} } = options;

  if (body && !headers['Content-Type']) {
    headers['Content-Type'] = 'application/json';
  }

  const baseUrl = getApiBaseUrl();
  const response = await fetch(`${baseUrl}${endpoint}`, {
    method,
    headers,
    credentials: 'include',
    body: body ? JSON.stringify(body) : undefined,
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({ detail: 'Request failed' }));
    const errorMessage = error.detail || error.message || 'Request failed';
    console.error(`[API] ${method} ${endpoint} failed with status ${response.status}: ${errorMessage}`);
    throw new Error(`${errorMessage} (HTTP ${response.status})`);
  }

  return response.json();
}

// For raw response access (like original apiClient)
export const raw = {
  async get(endpoint: string): Promise<Response> {
    return fetch(`${getApiBaseUrl()}${endpoint}`, { method: 'GET', credentials: 'include' });
  },
  async post(endpoint: string, body?: unknown): Promise<Response> {
    return fetch(`${getApiBaseUrl()}${endpoint}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: body ? JSON.stringify(body) : undefined,
    });
  },
  async postFormData(endpoint: string, formData: FormData): Promise<Response> {
    return fetch(`${getApiBaseUrl()}${endpoint}`, { method: 'POST', credentials: 'include', body: formData });
  },
  async put(endpoint: string, body?: unknown): Promise<Response> {
    return fetch(`${getApiBaseUrl()}${endpoint}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: body ? JSON.stringify(body) : undefined,
    });
  },
  async delete(endpoint: string): Promise<Response> {
    return fetch(`${getApiBaseUrl()}${endpoint}`, { method: 'DELETE', credentials: 'include' });
  },
};
