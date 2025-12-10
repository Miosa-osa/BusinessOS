import { writable } from 'svelte/store';
import { api, type Context, type CreateContextData } from '$lib/api/client';

interface ContextsState {
	contexts: Context[];
	currentContext: Context | null;
	loading: boolean;
}

function createContextsStore() {
	const { subscribe, update } = writable<ContextsState>({
		contexts: [],
		currentContext: null,
		loading: false
	});

	return {
		subscribe,

		async loadContexts(type?: string) {
			update((s) => ({ ...s, loading: true }));
			try {
				const contexts = await api.getContexts(type);
				update((s) => ({ ...s, contexts, loading: false }));
			} catch (error) {
				console.error('Failed to load contexts:', error);
				update((s) => ({ ...s, loading: false }));
			}
		},

		async loadContext(id: string) {
			update((s) => ({ ...s, loading: true }));
			try {
				const context = await api.getContext(id);
				update((s) => ({ ...s, currentContext: context, loading: false }));
			} catch (error) {
				console.error('Failed to load context:', error);
				update((s) => ({ ...s, loading: false }));
			}
		},

		async createContext(data: CreateContextData) {
			try {
				const context = await api.createContext(data);
				update((s) => ({ ...s, contexts: [context, ...s.contexts] }));
				return context;
			} catch (error) {
				console.error('Failed to create context:', error);
				throw error;
			}
		},

		async updateContext(id: string, data: Partial<CreateContextData>) {
			try {
				const context = await api.updateContext(id, data);
				update((s) => ({
					...s,
					contexts: s.contexts.map((c) => (c.id === id ? context : c)),
					currentContext: s.currentContext?.id === id ? context : s.currentContext
				}));
				return context;
			} catch (error) {
				console.error('Failed to update context:', error);
				throw error;
			}
		},

		async deleteContext(id: string) {
			try {
				await api.deleteContext(id);
				update((s) => ({
					...s,
					contexts: s.contexts.filter((c) => c.id !== id),
					currentContext: s.currentContext?.id === id ? null : s.currentContext
				}));
			} catch (error) {
				console.error('Failed to delete context:', error);
				throw error;
			}
		},

		clearCurrent() {
			update((s) => ({ ...s, currentContext: null }));
		}
	};
}

export const contexts = createContextsStore();
