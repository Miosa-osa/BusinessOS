/**
 * OSA Build Onboarding Types
 * Types for the OSA onboarding flow integration
 * Matches backend types from services.OSAOnboardingService
 */

/**
 * User analysis insights from backend
 */
export interface UserAnalysisResult {
	insights: string[];
	interests: string[];
	tools_used: string[];
	profile_summary: string;
	raw_data: Record<string, any>;
}

/**
 * Status of app generation
 */
export type AppGenerationStatus = 'generating' | 'ready' | 'failed';

/**
 * A single starter app suggestion
 */
export interface StarterApp {
	id: string;
	title: string;
	description: string;
	icon_emoji: string;
	icon_url: string;
	reasoning: string;
	category: string;
	status: AppGenerationStatus;
	workflow_id: string;
}

/**
 * Request payload for user analysis
 */
export interface AnalyzeUserRequest {
	email: string;
	gmail_connected: boolean;
	calendar_connected?: boolean;
}

/**
 * Response from analyze endpoint
 * Backend returns: { analysis: UserAnalysisResult }
 */
export interface AnalyzeUserResponse {
	analysis: UserAnalysisResult;
}

/**
 * Request payload for generating starter apps
 */
export interface GenerateAppsRequest {
	workspace_id: string;
	analysis: UserAnalysisResult;
}

/**
 * Response from generate-apps endpoint
 * Backend returns: { starter_apps: StarterApp[], ready_to_launch: boolean }
 */
export interface GenerateAppsResponse {
	starter_apps: StarterApp[];
	ready_to_launch: boolean;
}

/**
 * Response from apps-status endpoint
 * Backend returns: { analysis: UserAnalysisResult, starter_apps: StarterApp[], ready_to_launch: boolean }
 */
export interface AppsStatusResponse {
	analysis: UserAnalysisResult;
	starter_apps: StarterApp[];
	ready_to_launch: boolean;
}

/**
 * Response from profile endpoint
 * Backend returns: { analysis: UserAnalysisResult, starter_apps: StarterApp[] }
 */
export interface GetProfileResponse {
	analysis: UserAnalysisResult;
	starter_apps: StarterApp[];
}
