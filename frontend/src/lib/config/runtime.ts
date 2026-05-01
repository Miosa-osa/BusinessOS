export const API_VERSION = "v1";

export const LOCAL_BACKEND_URL =
  import.meta.env.VITE_LOCAL_BACKEND_URL || "http://localhost:8001";
export const LOCAL_OSA_URL = "http://localhost:18080";
export const PRODUCTION_BACKEND_URL =
  import.meta.env.VITE_BACKEND_URL || "https://api.businessos.dev";

function isBrowser(): boolean {
  return typeof window !== "undefined";
}

function isLocalhost(): boolean {
  if (!isBrowser()) return false;

  return (
    window.location.hostname === "localhost" ||
    window.location.hostname === "127.0.0.1"
  );
}

function isElectron(): boolean {
  return isBrowser() && "electron" in window;
}

function getStoredMode(): string | null {
  if (!isBrowser()) return null;
  return localStorage.getItem("businessos_mode");
}

function getStoredCloudUrl(): string | null {
  if (!isBrowser()) return null;
  return localStorage.getItem("businessos_cloud_url");
}

export function getDefaultCloudBackendUrl(): string {
  return isLocalhost() ? LOCAL_BACKEND_URL : PRODUCTION_BACKEND_URL;
}

export function getBackendUrl(): string {
  if (!isBrowser()) return "";

  if (isElectron()) {
    const mode = getStoredMode();
    const cloudUrl = getStoredCloudUrl() || getDefaultCloudBackendUrl();

    if (mode === "local") return LOCAL_OSA_URL;
    return cloudUrl;
  }

  if (import.meta.env.VITE_BACKEND_URL) {
    return import.meta.env.VITE_BACKEND_URL;
  }

  return getDefaultCloudBackendUrl();
}

export function getApiBaseUrl(version = API_VERSION): string {
  if (!isBrowser()) {
    return import.meta.env.VITE_API_URL || `/api/${version}`;
  }

  if (isElectron()) {
    const mode = getStoredMode();
    if (mode === "local") return `${LOCAL_OSA_URL}/api/${version}`;
    return `${getDefaultCloudBackendUrl()}/api/${version}`;
  }

  if (import.meta.env.VITE_API_URL) {
    return import.meta.env.VITE_API_URL;
  }

  return isLocalhost()
    ? `/api/${version}`
    : `${PRODUCTION_BACKEND_URL}/api/${version}`;
}

export function getLegacyApiBaseUrl(): string {
  if (!isBrowser()) {
    return import.meta.env.VITE_API_URL || "/api";
  }

  if (isElectron()) {
    const mode = getStoredMode();
    const cloudUrl = getStoredCloudUrl();

    if (mode === "cloud" && cloudUrl) return `${cloudUrl}/api`;
    if (mode === "local") return `${LOCAL_OSA_URL}/api`;
    return `${LOCAL_BACKEND_URL}/api`;
  }

  return import.meta.env.VITE_API_URL || "/api";
}
