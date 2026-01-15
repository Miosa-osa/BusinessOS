// OSA-5 API Types
// Matches backend types from internal/handlers/osa_api.go and internal/integrations/osa/client.go

export interface OSAHealthResponse {
	enabled: boolean;
	status: 'healthy' | 'unhealthy' | 'degraded';
	version: string;
	message?: string;
}

export interface GenerateAppRequest {
	workspace_id?: string;
	name: string;
	description: string;
	type: 'web' | 'mobile' | 'api' | 'fullstack' | 'cli' | 'desktop';
	parameters?: Record<string, unknown>;
}

export interface GenerateAppResponse {
	app_id: string;
	status: 'queued' | 'processing' | 'completed' | 'failed';
	message?: string;
	workspace_id?: string;
}

export interface AppGenerationStatus {
	app_id: string;
	status: 'queued' | 'processing' | 'completed' | 'failed';
	progress: number; // 0-100
	phase: string; // e.g., "Analyzing requirements", "Generating code", "Deploying"
	message?: string;
	created_at: string;
	updated_at: string;
	completed_at?: string;
	error?: string;
	result?: AppGenerationResult;
}

export interface AppGenerationResult {
	repository_url?: string;
	deployment_url?: string;
	files_generated: number;
	preview_available: boolean;
	artifacts: GeneratedArtifact[];
}

export interface GeneratedArtifact {
	type: 'code' | 'schema' | 'config' | 'documentation' | 'deployment';
	title: string;
	path: string;
	content?: string;
	language?: string;
}

export interface OSAWorkspace {
	id: string;
	name: string;
	description?: string;
	created_at: string;
	app_count: number;
}

export interface OSAWorkspacesResponse {
	workspaces: OSAWorkspace[];
	total: number;
}

// Event types for SSE streaming during generation
export type OSAGenerationEvent =
	| OSAProgressEvent
	| OSAPhaseEvent
	| OSAArtifactEvent
	| OSACompleteEvent
	| OSAErrorEvent;

export interface OSAProgressEvent {
	type: 'progress';
	app_id: string;
	progress: number;
	message: string;
}

export interface OSAPhaseEvent {
	type: 'phase';
	app_id: string;
	phase: string;
	description: string;
}

export interface OSAArtifactEvent {
	type: 'artifact';
	app_id: string;
	artifact: GeneratedArtifact;
}

export interface OSACompleteEvent {
	type: 'complete';
	app_id: string;
	result: AppGenerationResult;
}

export interface OSAErrorEvent {
	type: 'error';
	app_id: string;
	error: string;
}
