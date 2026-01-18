import { writable, get } from 'svelte/store';

// Check if running in Electron
const isElectron = typeof window !== 'undefined' && 'electron' in window;

// Check if running in development mode (localhost)
const isDev = typeof window !== 'undefined' &&
	(window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1');

// Backend URLs
const LOCAL_BACKEND_URL = 'http://localhost:8001';
const CLOUD_RUN_URL = 'https://businessos-api-460433387676.us-central1.run.app';

// App mode store - 'cloud' or 'local'
export const appMode = writable<'cloud' | 'local' | null>(null);
export const cloudServerUrl = writable<string>('');

// Initialize mode from localStorage
if (typeof window !== 'undefined') {
	let savedMode = localStorage.getItem('businessos_mode') as 'cloud' | 'local' | null;
	let savedUrl = localStorage.getItem('businessos_cloud_url') || '';

	// Auto-configure backend URL based on environment
	if (!savedUrl) {
		// In dev mode, use local backend; in production, use Cloud Run
		savedUrl = isDev ? LOCAL_BACKEND_URL : CLOUD_RUN_URL;
		localStorage.setItem('businessos_cloud_url', savedUrl);
		// Auto-set to cloud mode
		if (!savedMode) {
			savedMode = 'cloud';
			localStorage.setItem('businessos_mode', 'cloud');
		}
	}

	appMode.set(savedMode);
	cloudServerUrl.set(savedUrl);
}

// Save mode to localStorage
export function setAppMode(mode: 'cloud' | 'local', serverUrl?: string) {
	appMode.set(mode);
	localStorage.setItem('businessos_mode', mode);
	if (mode === 'cloud' && serverUrl) {
		cloudServerUrl.set(serverUrl);
		localStorage.setItem('businessos_cloud_url', serverUrl);
	}
	// Reload to apply new settings
	window.location.reload();
}

// Google OAuth - initiate OAuth flow
// Returns true on success, false on failure
export function initiateGoogleOAuth(serverUrl?: string): boolean {
	const baseUrl = serverUrl || get(cloudServerUrl);
	if (!baseUrl) {
		console.error('No cloud server URL configured');
		return false;
	}

	// Build the OAuth URL with redirect back to our auth callback
	const redirectUrl = encodeURIComponent(window.location.origin + '/auth/callback');
	const authUrl = `${baseUrl}/api/auth/google?redirect=${redirectUrl}`;

	// IMPORTANT: In Electron, we navigate WITHIN the window (not external browser)
	// This ensures cookies are set in Electron's session, not the system browser
	// The OAuth flow will redirect back to our app after authentication
	window.location.href = authUrl;
	return true;
}

// Email/Password Sign Up
export async function signUpWithEmail(email: string, password: string, name: string, serverUrl?: string) {
	const baseUrl = serverUrl || get(cloudServerUrl);
	if (!baseUrl) {
		return { error: { message: 'No cloud server URL configured' } };
	}

	try {
		const response = await fetch(`${baseUrl}/api/auth/sign-up/email`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			credentials: 'include',
			body: JSON.stringify({ email, password, name })
		});

		const data = await response.json();

		if (!response.ok) {
			return { error: { message: data.error || 'Sign up failed' } };
		}

		return { data };
	} catch (err) {
		return { error: { message: (err as Error).message || 'Network error' } };
	}
}

// Email/Password Sign In
export async function signInWithEmail(email: string, password: string, serverUrl?: string) {
	const baseUrl = serverUrl || get(cloudServerUrl);
	if (!baseUrl) {
		return { error: { message: 'No cloud server URL configured' } };
	}

	try {
		const response = await fetch(`${baseUrl}/api/auth/sign-in/email`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			credentials: 'include',
			body: JSON.stringify({ email, password })
		});

		const data = await response.json();

		if (!response.ok) {
			return { error: { message: data.error || 'Sign in failed' } };
		}

		return { data };
	} catch (err) {
		return { error: { message: (err as Error).message || 'Network error' } };
	}
}

// Get current session from server
export async function getSession(serverUrl?: string) {
	const baseUrl = serverUrl || get(cloudServerUrl);
	if (!baseUrl) {
		return { data: null, error: 'No cloud server URL configured' };
	}

	try {
		const response = await fetch(`${baseUrl}/api/auth/session`, {
			method: 'GET',
			credentials: 'include'
		});

		if (!response.ok) {
			return { data: null, error: 'Not authenticated' };
		}

		const data = await response.json();
		return { data, error: null };
	} catch (err) {
		return { data: null, error: (err as Error).message || 'Network error' };
	}
}

// Sign out
// Returns { success: true } or { success: false, error: string }
export async function signOutFromServer(serverUrl?: string): Promise<{ success: boolean; error?: string }> {
	const baseUrl = serverUrl || get(cloudServerUrl);
	if (!baseUrl) {
		console.error('No cloud server URL configured for sign out');
		// Still redirect to clear client-side state
		window.location.href = '/';
		return { success: false, error: 'No cloud server URL configured' };
	}

	try {
		const response = await fetch(`${baseUrl}/api/auth/logout`, {
			method: 'POST',
			credentials: 'include'
		});

		if (!response.ok) {
			const errorText = await response.text().catch(() => 'Unknown error');
			console.error(`Sign out failed with status ${response.status}: ${errorText}`);
			// Still redirect to clear client-side state even if server fails
			window.location.href = '/';
			return { success: false, error: `Server returned ${response.status}` };
		}
	} catch (err) {
		const errorMessage = err instanceof Error ? err.message : 'Network error';
		console.error('Sign out error:', errorMessage);
		// Still redirect to clear client-side state
		window.location.href = '/';
		return { success: false, error: errorMessage };
	}

	// Clear local session state
	window.location.href = '/';
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
			id: 'local-user',
			email: 'local@businessos.app',
			name: 'Local User',
			image: undefined as string | undefined,
		},
		session: {
			id: 'local-session',
		}
	},
	error: null
});

// Log when local mode session is used (helps debug auth issues)
if (typeof window !== 'undefined' && isElectron) {
	const mode = localStorage.getItem('businessos_mode');
	if (mode === 'local') {
		console.info('[Auth] Using local mode - no server authentication required');
	}
}

// For when mode is not yet selected - return a "pending" state
const pendingSession = writable({
	isPending: true,
	data: null,
	error: null
});

// Get the base URL for auth
function getBaseURL(): string {
	if (typeof window === 'undefined') return 'http://localhost:5174';

	const mode = get(appMode);
	const serverUrl = get(cloudServerUrl);

	// Cloud mode with server URL
	if (mode === 'cloud' && serverUrl) {
		return serverUrl;
	}

	// Local mode in Electron - use local backend
	if (isElectron) {
		return 'http://localhost:8001';
	}

	// Web app - use current origin
	return window.location.origin;
}

// Cloud session store - fetched from server
const cloudSession = writable<{
	isPending: boolean;
	data: { user: { id: string; email: string; name: string; image?: string }; session: { id: string } } | null;
	error: string | null;
}>({
	isPending: true,
	data: null,
	error: null
});

// Fetch cloud session on init (only in cloud mode)
async function initCloudSession() {
	const mode = get(appMode);
	if (mode !== 'cloud') return;

	cloudSession.set({ isPending: true, data: null, error: null });

	try {
		const result = await getSession();
		if (result.data?.user) {
			cloudSession.set({ isPending: false, data: result.data, error: null });
			// Only log in dev mode to reduce spam
			if (import.meta.env.DEV) {
				console.debug('[Auth] Session restored');
			}
		} else {
			cloudSession.set({ isPending: false, data: null, error: result.error || null });
		}
	} catch (err) {
		cloudSession.set({ isPending: false, data: null, error: (err as Error).message });
		console.error('[Auth] Session check failed:', err);
	}
}

// Clear session data (call this when receiving 401 from API)
export function clearSession() {
	cloudSession.set({ isPending: false, data: null, error: 'Session expired' });
}

// Re-check session from server
export async function refreshSession() {
	await initCloudSession();
}

// Initialize cloud session when mode changes to cloud
if (typeof window !== 'undefined') {
	appMode.subscribe((mode) => {
		if (mode === 'cloud') {
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
	email: async ({ email, password, name }: { email: string; password: string; name: string }) => {
		return signUpWithEmail(email, password, name);
	},
};
const localSignOut = async () => {
	if (typeof window !== 'undefined') window.location.href = '/';
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
	email: async ({ email, password, name }: { email: string; password: string; name: string }) => {
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

// Export auth functions - check mode dynamically at call time (not module load time)
// This fixes timing issues in Electron where mode isn't set until after module load
export const signIn = {
	email: async ({ email, password }: { email: string; password: string }) => {
		const mode = typeof window !== 'undefined' ? get(appMode) : null;
		if (isElectron && mode === 'local') {
			return localSignIn.email({ email, password });
		}
		return cloudSignIn.email({ email, password });
	},
	social: async () => {
		const mode = typeof window !== 'undefined' ? get(appMode) : null;
		if (isElectron && mode === 'local') {
			return localSignIn.social();
		}
		return cloudSignIn.social();
	},
};

export const signUp = {
	email: async ({ email, password, name }: { email: string; password: string; name: string }) => {
		const mode = typeof window !== 'undefined' ? get(appMode) : null;
		if (isElectron && mode === 'local') {
			return localSignUp.email({ email, password, name });
		}
		return cloudSignUp.email({ email, password, name });
	},
};

export const signOut = async () => {
	const mode = typeof window !== 'undefined' ? get(appMode) : null;
	// Sign out - mode determines local vs cloud behavior
	if (isElectron && mode === 'local') {
		return localSignOut();
	}
	return cloudSignOut();
};

export function useSession() {
	const mode = typeof window !== 'undefined' ? get(appMode) : null;
	// In Electron with no mode selected, return pending session
	if (isElectron && mode === null) return pendingSession;
	// In local mode, return local session
	if (isElectron && mode === 'local') return localSession;
	// In cloud mode or web, use cloud session
	return cloudSession;
}
