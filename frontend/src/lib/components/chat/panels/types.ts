// Shared types for chat panel components
// Re-exported here because TypeScript cannot export types from .svelte files

export interface ThinkingStep {
	type: 'explore' | 'analyze' | 'synthesize' | 'conclude' | 'verify' | 'fallback' | 'understand' | 'plan' | 'reason' | 'evaluate';
	content: string;
	duration?: number;
}

export interface ThinkingTrace {
	id?: string;
	content?: string;
	thinking_content?: string;
	steps?: ThinkingStep[];
	metadata?: {
		tokenCount?: number;
		duration?: number;
		model?: string;
	};
	thinking_tokens?: number;
	model_used?: string;
	duration_ms?: number;
	timestamp?: number;
}

export interface DelegatedTask {
	id: string;
	title: string;
	status: 'pending' | 'working' | 'done' | 'failed';
	conversationId: string;
	conversationTitle: string;
	agent: string;
	createdAt: string;
	progress?: number;
}

export interface ActiveResource {
	id: string;
	type: 'document' | 'artifact' | 'project' | 'context' | 'search_result';
	name?: string;
	source?: string;
	title?: string;
	contextId?: string;
	tokenCount?: number;
}
