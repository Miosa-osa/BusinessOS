import { writable, get } from "svelte/store";

// Check if running in Electron
const isElectron = typeof window !== "undefined" && "electron" in window;

// Check if running in development mode (localhost)
const isDev =
  typeof window !== "undefined" &&
  (window.location.hostname === "localhost" ||
    window.location.hostname === "127.0.0.1");

// Backend URLs
const LOCAL_BACKEND_URL = "http://localhost:8001";
const CLOUD_RUN_URL = "https://businessos-api-460433387676.us-central1.run.app";

// App mode store - 'cloud' or 'local'
export const appMode = writable<"cloud" | "local" | null>(null);
export const cloudServerUrl = writable<string>("");

// Initialize mode from localStorage
if (typeof window !== "undefined") {
  let savedMode = localStorage.getItem("businessos_mode") as
    | "cloud"
    | "local"
    | null;
  let savedUrl = localStorage.getItem("businessos_cloud_url") || "";

  // Auto-configure backend URL based on environment
  if (!savedUrl) {
    // In dev mode, use local backend; in production, use Cloud Run
    savedUrl = isDev ? LOCAL_BACKEND_URL : CLOUD_RUN_URL;
    localStorage.setItem("businessos_cloud_url", savedUrl);
    // Auto-set to cloud mode
    if (!savedMode) {
      savedMode = "cloud";
      localStorage.setItem("businessos_mode", "cloud");
    }
  }

  appMode.set(savedMode);
  cloudServerUrl.set(savedUrl);
}

// Save mode to localStorage
export function setAppMode(mode: "cloud" | "local", serverUrl?: string) {
  appMode.set(mode);
  localStorage.setItem("businessos_mode", mode);
  if (mode === "cloud" && serverUrl) {
    cloudServerUrl.set(serverUrl);
    localStorage.setItem("businessos_cloud_url", serverUrl);
  }
  // Reload to apply new settings
  window.location.reload();
}

// Helper: resolve base URL for auth API calls.
// In web dev mode (not Electron), return "" so fetch uses relative URLs
// routed through the Vite proxy. This keeps cookies on the same origin
// (localhost:5173) so SvelteKit server-side code can read them.
function getAuthBase(serverUrl?: string): string {
  if (serverUrl) return serverUrl;
  if (!isElectron && isDev) return ""; // relative → Vite proxy
  return get(cloudServerUrl);
}

// Helper function to get CSRF token from cookie
function getCSRFToken(): string | null {
  if (typeof document === "undefined") return null;

  const cookies = document.cookie.split(";");
  let foundToken: string | null = null;
  for (const cookie of cookies) {
    const trimmed = cookie.trim();
    // Split only on first '=' to preserve '=' in the token value
    const equalsIndex = trimmed.indexOf("=");
    if (equalsIndex === -1) continue;

    const name = trimmed.substring(0, equalsIndex);
    const value = trimmed.substring(equalsIndex + 1);

    if (name === "csrf_token") {
      foundToken = value;
      // Don't break - log all csrf_token cookies if there are multiple
    }
  }

  return foundToken;
}

// Helper function to add CSRF token to headers
function addCSRFHeaders(headers: HeadersInit = {}): HeadersInit {
  const csrfToken = getCSRFToken();

  if (csrfToken) {
    return {
      ...headers,
      "X-CSRF-Token": csrfToken,
    };
  }
  return headers;
}

// Google OAuth - initiate OAuth flow
// Returns true on success, false on failure
export function initiateGoogleOAuth(serverUrl?: string): boolean {
  const baseUrl = serverUrl || get(cloudServerUrl);
  if (!baseUrl) {
    console.error("No cloud server URL configured");
    return false;
  }

  // In Electron, open system browser for OAuth
  // The callback will redirect back with a token
  const redirectUrl = encodeURIComponent(
    window.location.origin + "/auth/callback",
  );
  const authUrl = `${baseUrl}/api/v1/auth/google?redirect=${redirectUrl}`;

  if (isElectron && (window as any).electron?.openExternal) {
    // Use Electron's shell to open in system browser
    (window as any).electron.openExternal(authUrl);
  } else {
    // Standard web redirect
    window.location.href = authUrl;
  }
  return true;
}

// Email/Password Sign Up
export async function signUpWithEmail(
  email: string,
  password: string,
  name: string,
  serverUrl?: string,
) {
  const baseUrl = getAuthBase(serverUrl);

  try {
    const response = await fetch(`${baseUrl}/api/v1/auth/sign-up/email`, {
      method: "POST",
      headers: addCSRFHeaders({ "Content-Type": "application/json" }),
      credentials: "include",
      body: JSON.stringify({ email, password, name }),
    });

    const data = await response.json();

    if (!response.ok) {
      return { error: { message: data.error || "Sign up failed" } };
    }

    return { data };
  } catch (err) {
    return { error: { message: (err as Error).message || "Network error" } };
  }
}

// Email/Password Sign In
export async function signInWithEmail(
  email: string,
  password: string,
  serverUrl?: string,
) {
  const baseUrl = getAuthBase(serverUrl);
  const url = `${baseUrl}/api/v1/auth/sign-in/email`;
  const csrfToken = getCSRFToken();
  const headers = addCSRFHeaders({ "Content-Type": "application/json" });
  const bodyPayload = { email, password };

  if (import.meta.env.DEV) {
    console.log("[AUTH DEBUG] signInWithEmail called");
    console.log("[AUTH DEBUG] baseUrl:", JSON.stringify(baseUrl));
    console.log("[AUTH DEBUG] full URL:", url);
    console.log("[AUTH DEBUG] CSRF token present:", !!csrfToken);
    console.log("[AUTH DEBUG] email:", email);
    console.log("[AUTH DEBUG] password length:", password.length);
    console.log("[AUTH DEBUG] headers:", JSON.stringify(headers));
    console.log("[AUTH DEBUG] cookies:", document.cookie);
  }

  try {
    const response = await fetch(url, {
      method: "POST",
      headers,
      credentials: "include",
      body: JSON.stringify(bodyPayload),
    });

    const data = await response.json();

    if (import.meta.env.DEV) {
      console.log("[AUTH DEBUG] response status:", response.status);
      console.log("[AUTH DEBUG] response data:", JSON.stringify(data));
    }

    if (!response.ok) {
      return { error: { message: data.error || "Sign in failed" } };
    }

    return { data };
  } catch (err) {
    if (import.meta.env.DEV) {
      console.error("[AUTH DEBUG] fetch error:", err);
    }
    return { error: { message: (err as Error).message || "Network error" } };
  }
}

// Get current session from server
export async function getSession(serverUrl?: string) {
  const baseUrl = getAuthBase(serverUrl);

  try {
    const response = await fetch(`${baseUrl}/api/v1/auth/session`, {
      method: "GET",
      credentials: "include",
    });

    if (!response.ok) {
      return { data: null, error: "Not authenticated" };
    }

    const data = await response.json();
    return { data, error: null };
  } catch (err) {
    return { data: null, error: (err as Error).message || "Network error" };
  }
}

// Check if user needs onboarding
// Returns { needsOnboarding: true/false, hasSession: true/false }
export async function checkOnboardingStatus(serverUrl?: string): Promise<{
  needsOnboarding: boolean;
  hasSession: boolean;
  error?: string;
}> {
  const baseUrl = getAuthBase(serverUrl);

  try {
    const response = await fetch(`${baseUrl}/api/v1/onboarding/status`, {
      method: "GET",
      credentials: "include",
    });

    if (!response.ok) {
      // If 401, user isn't authenticated
      if (response.status === 401) {
        return {
          needsOnboarding: false,
          hasSession: false,
          error: "Not authenticated",
        };
      }
      return {
        needsOnboarding: false,
        hasSession: false,
        error: "Failed to check onboarding status",
      };
    }

    const data = await response.json();
    return {
      needsOnboarding: data.needs_onboarding ?? false,
      hasSession: data.has_session ?? false,
    };
  } catch (err) {
    return {
      needsOnboarding: false,
      hasSession: false,
      error: (err as Error).message || "Network error",
    };
  }
}

// Sign out
// Returns { success: true } or { success: false, error: string }
export async function signOutFromServer(
  serverUrl?: string,
): Promise<{ success: boolean; error?: string }> {
  const baseUrl = getAuthBase(serverUrl);

  try {
    const response = await fetch(`${baseUrl}/api/auth/logout`, {
      method: "POST",
      credentials: "include",
    });

    if (!response.ok) {
      const errorText = await response.text().catch(() => "Unknown error");
      console.error(
        `Sign out failed with status ${response.status}: ${errorText}`,
      );
      // Still redirect to clear client-side state even if server fails
      window.location.href = "/";
      return { success: false, error: `Server returned ${response.status}` };
    }
  } catch (err) {
    const errorMessage = err instanceof Error ? err.message : "Network error";
    console.error("Sign out error:", errorMessage);
    // Still redirect to clear client-side state
    window.location.href = "/";
    return { success: false, error: errorMessage };
  }

  // Clear local session state
  window.location.href = "/";
  return { success: true };
}

// For Local mode: Create a local-only session
// IMPORTANT: This is a synthetic session for Electron local mode only.
// It does NOT represent actual server authentication.
// The `isLocalMode` flag distinguishes this from real authenticated sessions.
const localSession = writable({
  isPending: false,
  isLocalMode: true, // Flag to indicate this is a synthetic local session
  data: {
    user: {
      id: "local-user",
      email: "local@businessos.app",
      name: "Local User",
      image: undefined as string | undefined,
    },
    session: {
      id: "local-session",
    },
  },
  error: null,
});

// Log when local mode session is used (helps debug auth issues)
if (typeof window !== "undefined" && isElectron) {
  const mode = localStorage.getItem("businessos_mode");
  if (mode === "local") {
    console.info("[Auth] Using local mode - no server authentication required");
  }
}

// For when mode is not yet selected - return a "pending" state
const pendingSession = writable({
  isPending: true,
  data: null,
  error: null,
});

// Get the base URL for auth
function getBaseURL(): string {
  if (typeof window === "undefined") return "http://localhost:5174";

  const mode = get(appMode);
  const serverUrl = get(cloudServerUrl);

  // Cloud mode with server URL
  if (mode === "cloud" && serverUrl) {
    return serverUrl;
  }

  // Local mode in Electron - use local backend
  if (isElectron) {
    return LOCAL_BACKEND_URL;
  }

  // Web app - use current origin
  return window.location.origin;
}

// Cloud session store - fetched from server
const cloudSession = writable<{
  isPending: boolean;
  data: {
    user: { id: string; email: string; name: string; image?: string; createdAt?: string };
    session: { id: string };
  } | null;
  error: string | null;
}>({
  isPending: true,
  data: null,
  error: null,
});

// Fetch cloud session on init (only in cloud mode)
async function initCloudSession() {
  const mode = get(appMode);
  if (mode !== "cloud") return;

  cloudSession.set({ isPending: true, data: null, error: null });

  try {
    const result = await getSession();
    if (result.data?.user) {
      cloudSession.set({ isPending: false, data: result.data, error: null });
    } else {
      cloudSession.set({
        isPending: false,
        data: null,
        error: result.error || null,
      });
    }
  } catch (err) {
    cloudSession.set({
      isPending: false,
      data: null,
      error: (err as Error).message,
    });
  }
}

// Clear session data (call this when receiving 401 from API)
export function clearSession() {
  cloudSession.set({ isPending: false, data: null, error: "Session expired" });
}

// Re-check session from server
export async function refreshSession() {
  await initCloudSession();
}

// Initialize cloud session when mode changes to cloud
if (typeof window !== "undefined") {
  appMode.subscribe((mode) => {
    if (mode === "cloud") {
      initCloudSession();
    }
  });
}

// Local mode auth functions (for compatibility)
const localSignIn = {
  email: async ({ email, password }: { email: string; password: string }) => {
    return signInWithEmail(email, password);
  },
  social: async () => ({ data: get(localSession).data, error: null }),
};
const localSignUp = {
  email: async ({
    email,
    password,
    name,
  }: {
    email: string;
    password: string;
    name: string;
  }) => {
    return signUpWithEmail(email, password, name);
  },
};
const localSignOut = async () => {
  if (typeof window !== "undefined") window.location.href = "/";
  return {};
};

// Cloud mode auth functions
const cloudSignIn = {
  email: async ({ email, password }: { email: string; password: string }) => {
    const result = await signInWithEmail(email, password);
    if (result.data) {
      await initCloudSession();
    }
    return result;
  },
  social: async () => {
    initiateGoogleOAuth();
    return { data: null, error: null };
  },
};
const cloudSignUp = {
  email: async ({
    email,
    password,
    name,
  }: {
    email: string;
    password: string;
    name: string;
  }) => {
    const result = await signUpWithEmail(email, password, name);
    if (result.data) {
      await initCloudSession();
    }
    return result;
  },
};
const cloudSignOut = async () => {
  await signOutFromServer();
  cloudSession.set({ isPending: false, data: null, error: null });
  return {};
};

// Export auth functions — each method checks mode at call time, not at import time.
// This avoids the frozen-IIFE problem where mode was captured once on module load.
export const signIn = {
  email: async (params: { email: string; password: string }) => {
    const mode = typeof window !== "undefined" ? get(appMode) : null;
    if (isElectron && mode === "local") return localSignIn.email(params);
    return cloudSignIn.email(params);
  },
  social: async () => {
    const mode = typeof window !== "undefined" ? get(appMode) : null;
    if (isElectron && mode === "local") return localSignIn.social();
    return cloudSignIn.social();
  },
};

export const signUp = {
  email: async (params: { email: string; password: string; name: string }) => {
    const mode = typeof window !== "undefined" ? get(appMode) : null;
    if (isElectron && mode === "local") return localSignUp.email(params);
    return cloudSignUp.email(params);
  },
};

export const signOut = async () => {
  const mode = typeof window !== "undefined" ? get(appMode) : null;
  if (isElectron && mode === "local") return localSignOut();
  return cloudSignOut();
};

export const useSession = () => {
  const mode = typeof window !== "undefined" ? get(appMode) : null;
  // In Electron with no mode selected, return pending session
  if (isElectron && mode === null) return pendingSession;
  // In local mode, return local session
  if (isElectron && mode === "local") return localSession;
  // In cloud mode or web, use cloud session
  return cloudSession;
};
