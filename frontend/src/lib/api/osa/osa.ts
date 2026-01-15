// OSA-5 API Client
// Follows patterns from /lib/api/ai/ai.ts and /lib/api/base.ts

import { request, raw } from '../base';
import type {
	OSAHealthResponse,
	GenerateAppRequest,
	GenerateAppResponse,
	AppGenerationStatus,
	OSAWorkspacesResponse,
	OSAGenerationEvent
} from './types';

/**
 * Check OSA integration health and availability
 */
export async function checkOSAHealth(): Promise<OSAHealthResponse> {
	return request<OSAHealthResponse>('/osa/health');
}

/**
 * Generate a new application using OSA-5 orchestrator
 * @param req - Application generation request
 * @returns Generation response with app_id for tracking
 */
export async function generateApp(req: GenerateAppRequest): Promise<GenerateAppResponse> {
	return request<GenerateAppResponse>('/osa/generate', {
		method: 'POST',
		body: req
	});
}

/**
 * Get the current status of an app generation
 * @param appId - App ID returned from generateApp()
 */
export async function getAppStatus(appId: string): Promise<AppGenerationStatus> {
	return request<AppGenerationStatus>(`/osa/status/${appId}`);
}

/**
 * Get list of available OSA workspaces
 */
export async function getWorkspaces(): Promise<OSAWorkspacesResponse> {
	return request<OSAWorkspacesResponse>('/osa/workspaces');
}

/**
 * Stream app generation progress using Server-Sent Events
 * Similar to pullModel() pattern in ai.ts
 * @param appId - App ID to stream progress for
 * @returns EventSource for listening to generation events
 */
export function streamAppGeneration(appId: string): EventSource | null {
	try {
		const eventSource = new EventSource(`/api/osa/generate/${appId}/stream`);
		return eventSource;
	} catch (error) {
		console.error('Failed to create EventSource for OSA generation:', error);
		return null;
	}
}

/**
 * Parse OSA generation event from SSE stream
 * @param event - MessageEvent from EventSource
 * @returns Parsed OSA generation event
 */
export function parseGenerationEvent(event: MessageEvent): OSAGenerationEvent | null {
	try {
		const data = JSON.parse(event.data);
		return data as OSAGenerationEvent;
	} catch (error) {
		console.error('Failed to parse OSA generation event:', error);
		return null;
	}
}

/**
 * Cancel an in-progress app generation
 * @param appId - App ID to cancel
 */
export async function cancelGeneration(appId: string): Promise<void> {
	return request<void>(`/osa/cancel/${appId}`, {
		method: 'POST'
	});
}
