import { writable } from 'svelte/store';
import { api, type Context, type ContextListItem, type CreateContextData, type UpdateContextData, type Block, type ShareResponse } from '$lib/api/contexts';

interface ContextsState {
	contexts: ContextListItem[];
	currentContext: Context | null;
	loading: boolean;
}

interface ContextFilters {
	type?: string;
	includeArchived?: boolean;
	templatesOnly?: boolean;
	parentId?: string;
	search?: string;
}

function createContextsStore() {
	const { subscribe, update } = writable<ContextsState>({
		contexts: [],
		currentContext: null,
		loading: false
	});

	return {
		subscribe,

		async loadContexts(filters?: ContextFilters | string) {
			update((s) => ({ ...s, loading: true }));
			try {
				// Support both old string format and new filters object
				const filterObj = typeof filters === 'string' ? { type: filters } : filters;
				const contexts = await api.getContexts(filterObj);
				update((s) => ({ ...s, contexts, loading: false }));
			} catch (error) {
				console.error('Failed to load contexts:', error);
				update((s) => ({ ...s, loading: false }));
			}
		},

		async loadContext(id: string) {
			update((s) => ({ ...s, loading: true }));
			try {
				const ctx = await api.getContext(id);
				update((s) => ({ ...s, currentContext: ctx, loading: false }));
				return ctx;
			} catch (error) {
				console.error('Failed to load context:', error);
				update((s) => ({ ...s, loading: false }));
				throw error;
			}
		},

		async createContext(data: CreateContextData) {
			try {
				const ctx = await api.createContext(data);
				update((s) => ({
					...s,
					contexts: [{
						id: ctx.id,
						name: ctx.name,
						type: ctx.type,
						icon: ctx.icon,
						cover_image: ctx.cover_image,
						parent_id: ctx.parent_id,
						is_template: ctx.is_template,
						is_archived: ctx.is_archived,
						word_count: ctx.word_count,
						property_schema: ctx.property_schema,
						properties: ctx.properties,
						client_id: ctx.client_id,
						updated_at: ctx.updated_at
					}, ...s.contexts]
				}));
				return ctx;
			} catch (error) {
				console.error('Failed to create context:', error);
				throw error;
			}
		},

		async updateContext(id: string, data: UpdateContextData) {
			try {
				const ctx = await api.updateContext(id, data);
				update((s) => ({
					...s,
					contexts: s.contexts.map((c) => (c.id === id ? {
						id: ctx.id,
						name: ctx.name,
						type: ctx.type,
						icon: ctx.icon,
						cover_image: ctx.cover_image,
						parent_id: ctx.parent_id,
						is_template: ctx.is_template,
						is_archived: ctx.is_archived,
						word_count: ctx.word_count,
						property_schema: ctx.property_schema,
						properties: ctx.properties,
						client_id: ctx.client_id,
						updated_at: ctx.updated_at
					} : c)),
					currentContext: s.currentContext?.id === id ? ctx : s.currentContext
				}));
				return ctx;
			} catch (error) {
				console.error('Failed to update context:', error);
				throw error;
			}
		},

		async updateBlocks(id: string, blocks: Block[], wordCount?: number) {
			try {
				const ctx = await api.updateContextBlocks(id, { blocks, word_count: wordCount });
				update((s) => ({
					...s,
					currentContext: s.currentContext?.id === id ? ctx : s.currentContext
				}));
				return ctx;
			} catch (error) {
				console.error('Failed to update blocks:', error);
				throw error;
			}
		},

		async enableSharing(id: string): Promise<ShareResponse> {
			try {
				const response = await api.enableContextSharing(id);
				update((s) => ({
					...s,
					currentContext: s.currentContext?.id === id
						? { ...s.currentContext, is_public: true, share_id: response.share_id }
						: s.currentContext
				}));
				return response;
			} catch (error) {
				console.error('Failed to enable sharing:', error);
				throw error;
			}
		},

		async disableSharing(id: string) {
			try {
				await api.disableContextSharing(id);
				update((s) => ({
					...s,
					currentContext: s.currentContext?.id === id
						? { ...s.currentContext, is_public: false }
						: s.currentContext
				}));
			} catch (error) {
				console.error('Failed to disable sharing:', error);
				throw error;
			}
		},

		async duplicateContext(id: string) {
			try {
				const ctx = await api.duplicateContext(id);
				update((s) => ({
					...s,
					contexts: [{
						id: ctx.id,
						name: ctx.name,
						type: ctx.type,
						icon: ctx.icon,
						cover_image: ctx.cover_image,
						parent_id: ctx.parent_id,
						is_template: ctx.is_template,
						is_archived: ctx.is_archived,
						word_count: ctx.word_count,
						property_schema: ctx.property_schema,
						properties: ctx.properties,
						client_id: ctx.client_id,
						updated_at: ctx.updated_at
					}, ...s.contexts]
				}));
				return ctx;
			} catch (error) {
				console.error('Failed to duplicate context:', error);
				throw error;
			}
		},

		async archiveContext(id: string) {
			try {
				await api.archiveContext(id);
				update((s) => ({
					...s,
					contexts: s.contexts.filter((c) => c.id !== id),
					currentContext: s.currentContext?.id === id
						? { ...s.currentContext, is_archived: true }
						: s.currentContext
				}));
			} catch (error) {
				console.error('Failed to archive context:', error);
				throw error;
			}
		},

		async unarchiveContext(id: string) {
			try {
				await api.unarchiveContext(id);
				update((s) => ({
					...s,
					currentContext: s.currentContext?.id === id
						? { ...s.currentContext, is_archived: false }
						: s.currentContext
				}));
			} catch (error) {
				console.error('Failed to unarchive context:', error);
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
