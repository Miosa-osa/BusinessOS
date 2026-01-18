import { request } from './base';

// Response types matching backend
export interface CheckUsernameResponse {
	available: boolean;
	reason?: string;
}

export interface SetUsernameResponse {
	success: boolean;
	username: string;
}

export interface UserProfile {
	id: string;
	username?: string;
	email: string;
	full_name?: string;
	has_username: boolean;
	username_claimed_at?: string;
}

/**
 * Check if a username is available
 * GET /api/users/check-username/:username
 */
export async function checkUsernameAvailability(username: string): Promise<CheckUsernameResponse> {
	console.log('[API] Checking username availability:', username);
	try {
		const result = await request<CheckUsernameResponse>(`/users/check-username/${username}`, {
			method: 'GET',
		});
		console.log('[API] Username check result:', result);
		return result;
	} catch (error) {
		console.error('[API] Username check failed:', error);
		throw error;
	}
}

/**
 * Set or update the username for the current authenticated user
 * PATCH /api/users/me/username
 */
export async function setUsername(username: string): Promise<SetUsernameResponse> {
	return request<SetUsernameResponse>('/users/me/username', {
		method: 'PATCH',
		body: { username },
	});
}

/**
 * Get current user profile
 * GET /api/users/me
 */
export async function getCurrentUser(): Promise<UserProfile> {
	return request<UserProfile>('/users/me', {
		method: 'GET',
	});
}

/**
 * Mark onboarding as complete
 * POST /api/users/me/complete-onboarding
 */
export async function completeOnboarding(): Promise<{ success: boolean; message: string }> {
	return request<{ success: boolean; message: string }>('/users/me/complete-onboarding', {
		method: 'POST',
	});
}
