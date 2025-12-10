import { createAuthClient } from 'better-auth/svelte';

export const authClient = createAuthClient({
	baseURL: typeof window !== 'undefined' ? window.location.origin : 'http://localhost:5174',
	fetchOptions: {
		credentials: 'include' // Ensure cookies are sent with requests
	}
});

export const { signIn, signUp, signOut, useSession } = authClient;
