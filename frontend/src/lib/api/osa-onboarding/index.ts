/**
 * OSA Onboarding API Client
 * Handles communication with the OSA Build onboarding backend
 */

import { request } from '../base';
import type {
	UserAnalysisResult,
	AnalyzeUserResponse,
	GenerateAppsResponse,
	AppsStatusResponse,
	GetProfileResponse
} from './types';

/**
 * Analyze user data and get personalized insights
 * POST /api/osa-onboarding/analyze
 *
 * @param email - User's email address
 * @param gmailConnected - Whether Gmail is connected for analysis
 * @param calendarConnected - Whether Calendar is connected for analysis
 * @returns Analysis result with insights, interests, and tools
 */
export async function analyzeUser(
	email: string,
	gmailConnected: boolean,
	calendarConnected = false
): Promise<AnalyzeUserResponse> {
	return request<AnalyzeUserResponse>('/osa-onboarding/analyze', {
		method: 'POST',
		body: {
			email,
			gmail_connected: gmailConnected,
			calendar_connected: calendarConnected
		}
	});
}

/**
 * Generate 4 personalized starter apps based on user analysis
 * POST /api/osa-onboarding/generate-apps
 *
 * @param workspaceId - Workspace identifier
 * @param analysis - User analysis result
 * @returns Array of starter app suggestions
 */
export async function generateStarterApps(
	workspaceId: string,
	analysis: UserAnalysisResult
): Promise<GenerateAppsResponse> {
	return request<GenerateAppsResponse>('/osa-onboarding/generate-apps', {
		method: 'POST',
		body: {
			workspace_id: workspaceId,
			analysis
		}
	});
}

/**
 * Check the status of app generation
 * GET /api/osa-onboarding/apps-status?workspace_id=xxx
 *
 * @param workspaceId - Workspace identifier
 * @returns Current status of all apps and analysis
 */
export async function checkAppsStatus(workspaceId: string): Promise<AppsStatusResponse> {
	return request<AppsStatusResponse>(
		`/osa-onboarding/apps-status?workspace_id=${encodeURIComponent(workspaceId)}`
	);
}

/**
 * Get saved onboarding profile for a workspace
 * GET /api/osa-onboarding/profile?workspace_id=xxx
 *
 * @param workspaceId - Workspace identifier
 * @returns Complete onboarding profile with analysis and apps
 */
export async function getProfile(workspaceId: string): Promise<GetProfileResponse> {
	return request<GetProfileResponse>(
		`/osa-onboarding/profile?workspace_id=${encodeURIComponent(workspaceId)}`
	);
}

// Re-export types for convenience
export type * from './types';
