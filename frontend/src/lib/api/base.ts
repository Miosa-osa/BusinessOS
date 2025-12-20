// Shared fetch logic copied from the original ApiClient.request implementation
function getApiBase(): string {
  if (typeof window === 'undefined') {
    return import.meta.env.VITE_API_URL || '/api';
  }

  const isElectron = 'electron' in window;

  if (isElectron) {
    const mode = localStorage.getItem('businessos_mode');
    const cloudUrl = localStorage.getItem('businessos_cloud_url');

    if (mode === 'cloud' && cloudUrl) {
      return `${cloudUrl}/api`;
    } else if (mode === 'local') {
      return 'http://localhost:18080/api';
    }
    return 'http://localhost:8000/api';
  }

  return import.meta.env.VITE_API_URL || (import.meta.env.DEV ? 'http://localhost:8000/api' : '/api');
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
    throw new Error(error.detail || 'Request failed');
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
