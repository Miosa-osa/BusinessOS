import { writable, derived } from 'svelte/store';
import {
	getCustomAgents,
	getCustomAgent,
	createCustomAgent,
	updateCustomAgent,
	deleteCustomAgent,
	getAgentsByCategory,
	getAgentPresets,
	getAgentPreset,
	createFromPreset,
	testAgent,
	testSandbox
} from '$lib/api/ai/ai';
import type { CustomAgent, AgentPreset } from '$lib/api/ai/types';

interface AgentFilters {
	category: string | null;
	search: string;
	status: 'active' | 'inactive' | null;
}

interface AgentsState {
	agents: CustomAgent[];
	currentAgent: CustomAgent | null;
	presets: AgentPreset[];
	loading: boolean;
	error: string | null;
	filters: AgentFilters;
}

function createAgentsStore() {
	const { subscribe, update } = writable<AgentsState>({
		agents: [],
		currentAgent: null,
		presets: [],
		loading: false,
		error: null,
		filters: {
			category: null,
			search: '',
			status: null
		}
	});

	// Request versioning to prevent race conditions
	let loadRequestId = 0;

	return {
		subscribe,

		// ============ Agent Methods ============

		async loadAgents(filters?: Partial<AgentFilters>) {
			// Increment request ID to track latest request
			const thisRequestId = ++loadRequestId;
			update((s) => ({ ...s, loading: true, error: null }));
			try {
				// Apply filters if provided
				if (filters) {
					update((s) => ({
						...s,
						filters: { ...s.filters, ...filters }
					}));
				}

				let state: AgentsState;
				update((s) => {
					state = s;
					return s;
				});

				// Get agents (include inactive if filter is null or explicitly set to inactive)
				const includeInactive = state!.filters.status === null || state!.filters.status === 'inactive';
				const response = await getCustomAgents(includeInactive);
				let agents = response.agents;

				// Apply client-side filters
				if (state!.filters.category) {
					agents = agents.filter((a) => a.category === state!.filters.category);
				}

				if (state!.filters.search) {
					const searchLower = state!.filters.search.toLowerCase();
					agents = agents.filter(
						(a) =>
							a.name.toLowerCase().includes(searchLower) ||
							a.display_name.toLowerCase().includes(searchLower) ||
							a.description?.toLowerCase().includes(searchLower)
					);
				}

				if (state!.filters.status === 'active') {
					agents = agents.filter((a) => a.is_active);
				} else if (state!.filters.status === 'inactive') {
					agents = agents.filter((a) => !a.is_active);
				}

				// Only update if this is still the latest request
				if (thisRequestId === loadRequestId) {
					update((s) => ({ ...s, agents, loading: false }));
				}
			} catch (error) {
				console.error('Failed to load agents:', error);

				// Only update error if this is still the latest request
				if (thisRequestId === loadRequestId) {
					update((s) => ({
						...s,
						loading: false,
						error: error instanceof Error ? error.message : 'Failed to load agents'
					}));
				}
			}
		},

		async loadAgent(id: string) {
			update((s) => ({ ...s, loading: true, error: null }));
			try {
				const agent = await getCustomAgent(id);
				update((s) => ({ ...s, currentAgent: agent, loading: false }));
				return agent;
			} catch (error) {
				console.error('Failed to load agent:', error);
				update((s) => ({
					...s,
					loading: false,
					error: error instanceof Error ? error.message : 'Failed to load agent'
				}));
				return null;
			}
		},

		async createAgent(data: Partial<CustomAgent>) {
			try {
				const agent = await createCustomAgent(data);
				update((s) => ({ ...s, agents: [agent, ...s.agents] }));
				return agent;
			} catch (error) {
				console.error('Failed to create agent:', error);
				throw error;
			}
		},

		async updateAgent(id: string, data: Partial<CustomAgent>) {
			try {
				const agent = await updateCustomAgent(id, data);
				update((s) => ({
					...s,
					agents: s.agents.map((a) => (a.id === id ? agent : a)),
					currentAgent: s.currentAgent?.id === id ? agent : s.currentAgent
				}));
				return agent;
			} catch (error) {
				console.error('Failed to update agent:', error);
				throw error;
			}
		},

		async deleteAgent(id: string) {
			try {
				await deleteCustomAgent(id);
				update((s) => ({
					...s,
					agents: s.agents.filter((a) => a.id !== id),
					currentAgent: s.currentAgent?.id === id ? null : s.currentAgent
				}));
			} catch (error) {
				console.error('Failed to delete agent:', error);
				throw error;
			}
		},

		// ============ Current Agent Methods ============

		setCurrentAgent(agent: CustomAgent | null) {
			update((s) => ({ ...s, currentAgent: agent }));
		},

		clearCurrent() {
			update((s) => ({ ...s, currentAgent: null }));
		},

		// ============ Filter Methods ============

		setFilters(filters: Partial<AgentFilters>) {
			update((s) => ({
				...s,
				filters: { ...s.filters, ...filters }
			}));
		},

		clearFilters() {
			update((s) => ({
				...s,
				filters: {
					category: null,
					search: '',
					status: null
				}
			}));
		},

		// ============ Error Methods ============

		clearError() {
			update((s) => ({ ...s, error: null }));
		},

		// ============ Test Utilities ============

		reset() {
			update((s) => ({
				agents: [],
				currentAgent: null,
				presets: [],
				loading: false,
				error: null,
				filters: {
					category: null,
					search: '',
					status: null
				}
			}));
		},

		// ============ Preset Methods ============

		async loadPresets() {
			update((s) => ({ ...s, loading: true, error: null }));
			try {
				const response = await getAgentPresets();
				update((s) => ({ ...s, presets: response.presets, loading: false }));
			} catch (error) {
				console.error('Failed to load agent presets:', error);
				update((s) => ({
					...s,
					loading: false,
					error: error instanceof Error ? error.message : 'Failed to load agent presets'
				}));
			}
		},

		async loadPreset(id: string) {
			try {
				return await getAgentPreset(id);
			} catch (error) {
				console.error('Failed to load agent preset:', error);
				throw error;
			}
		},

		async createFromPreset(presetId: string, name?: string) {
			try {
				const agent = await createFromPreset(presetId, name);
				update((s) => ({ ...s, agents: [agent, ...s.agents] }));
				return agent;
			} catch (error) {
				console.error('Failed to create agent from preset:', error);
				throw error;
			}
		},

		// ============ Testing Methods ============

		async testAgent(id: string, message: string): Promise<ReadableStream<Uint8Array> | null> {
			try {
				return await testAgent(id, message);
			} catch (error) {
				console.error('Failed to test agent:', error);
				throw error;
			}
		},

		async testSandbox(config: {
			system_prompt: string;
			message: string;
			model?: string;
			temperature?: number;
		}): Promise<ReadableStream<Uint8Array> | null> {
			try {
				// Convert message to test_message for API
				const apiConfig = {
					system_prompt: config.system_prompt,
					test_message: config.message,
					model: config.model,
					temperature: config.temperature
				};
				return await testSandbox(apiConfig);
			} catch (error) {
				console.error('Failed to test in sandbox:', error);
				throw error;
			}
		}
	};
}

export const agents = createAgentsStore();

// ============ Derived Stores ============

export const selectedAgent = derived(agents, ($agents) => $agents.currentAgent);

export const agentsByCategory = derived(agents, ($agents) => {
	const byCategory: Record<string, CustomAgent[]> = {};

	for (const agent of $agents.agents) {
		const category = agent.category || 'uncategorized';
		if (!byCategory[category]) {
			byCategory[category] = [];
		}
		byCategory[category].push(agent);
	}

	return byCategory;
});

export const activeAgents = derived(agents, ($agents) =>
	$agents.agents.filter((a) => a.is_active)
);

// ============ UI Constants ============

export const categoryColors: Record<string, string> = {
	general: 'bg-blue-50 text-blue-700',
	specialist: 'bg-purple-50 text-purple-700',
	system: 'bg-gray-100 text-gray-600',
	custom: 'bg-emerald-50 text-emerald-700',
	uncategorized: 'bg-orange-50 text-orange-700'
};

export const categoryLabels: Record<string, string> = {
	general: 'General',
	specialist: 'Specialist',
	system: 'System',
	custom: 'Custom',
	uncategorized: 'Uncategorized'
};
