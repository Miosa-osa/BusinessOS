import {
	documentsStore,
	activeDocumentStore,
	treeStore
} from '../stores/documents';
import type {
	Document,
	DocumentMeta,
	DocumentType,
	Block,
	BlockType,
	RichText,
	PropertyValue
} from '../entities/types';
import * as contextsApi from '$lib/api/contexts';
import type { Context, ContextListItem, Block as ContextBlock } from '$lib/api/contexts/types';

/**
 * Document Service
 * Bridges the knowledge-base module with the existing contexts API.
 * Maps Context ↔ Document for compatibility.
 */

// ============================================================================
// Type Mapping Helpers
// ============================================================================

/**
 * Map ContextType to DocumentType
 */
function mapContextTypeToDocType(type: string): DocumentType {
	switch (type) {
		case 'document':
			return 'document';
		case 'custom':
			return 'document';
		case 'person':
		case 'business':
		case 'project':
			return 'database'; // Context profiles are like databases
		default:
			return 'document';
	}
}

/**
 * Map ContextListItem to DocumentMeta
 */
function mapContextToDocMeta(ctx: ContextListItem): DocumentMeta {
	return {
		id: ctx.id,
		parent_id: ctx.parent_id,
		type: mapContextTypeToDocType(ctx.type),
		title: ctx.name,
		icon: ctx.icon ? { type: 'emoji', value: ctx.icon } : null,
		is_favorite: ctx.properties?.is_favorite === true,
		is_archived: ctx.is_archived,
		updated_at: ctx.updated_at,
		children_count: 0 // Will be calculated from tree structure
	};
}

/**
 * Map full Context to Document
 */
function mapContextToDocument(ctx: Context): Document {
	return {
		id: ctx.id,
		workspace_id: '', // Not provided by API
		parent_id: ctx.parent_id,
		type: mapContextTypeToDocType(ctx.type),
		title: ctx.name,
		icon: ctx.icon ? { type: 'emoji', value: ctx.icon } : null,
		cover: ctx.cover_image || null,
		content: ctx.blocks ? mapContextBlocksToDocBlocks(ctx.blocks) : [],
		properties: (ctx.structured_data || {}) as Record<string, PropertyValue>,
		is_template: ctx.is_template ?? false,
		is_favorite: ctx.properties?.is_favorite === true,
		is_archived: ctx.is_archived,
		created_at: ctx.created_at,
		updated_at: ctx.updated_at,
		created_by: '', // Not provided by API
		last_edited_by: '' // Not provided by API
	};
}

/**
 * Map Context blocks to Document blocks
 */
function mapContextBlocksToDocBlocks(blocks: ContextBlock[]): Block[] {
	return blocks.map((block) => ({
		id: block.id,
		type: block.type as BlockType,
		content: block.content ? parseBlockContent(block.content) : [],
		properties: block.properties || {},
		children: block.children ? mapContextBlocksToDocBlocks(block.children) : [],
		created_at: new Date().toISOString(),
		updated_at: new Date().toISOString()
	}));
}

/**
 * Parse block content string to RichText array
 */
function parseBlockContent(content: string | null): RichText[] {
	if (!content) return [];
	// Simple text content
	return [
		{
			type: 'text',
			text: { content, link: null },
			annotations: {
				bold: false,
				italic: false,
				strikethrough: false,
				underline: false,
				code: false,
				color: 'default'
			},
			plain_text: content,
			href: null
		}
	];
}

/**
 * Map Document blocks to Context blocks for saving
 */
function mapDocBlocksToContextBlocks(blocks: Block[]): ContextBlock[] {
	return blocks.map((block) => ({
		id: block.id,
		type: block.type,
		content: extractPlainText(block.content),
		properties: block.properties as Record<string, unknown>,
		children: block.children ? mapDocBlocksToContextBlocks(block.children) : undefined
	})) as ContextBlock[];
}

/**
 * Extract plain text from RichText array
 */
function extractPlainText(content: RichText[] | string | null): string | null {
	if (!content) return null;
	if (typeof content === 'string') return content;
	return content.map((rt) => rt.plain_text || '').join('');
}

// ============================================================================
// Document CRUD Operations
// ============================================================================

export async function fetchDocuments(parentId?: string | null): Promise<DocumentMeta[]> {
	documentsStore.setLoading(true);
	try {
		const contexts = await contextsApi.getContexts({
			parentId: parentId || undefined
		});

		const documents = contexts.map(mapContextToDocMeta);
		documentsStore.setDocumentMetas(documents);
		return documents;
	} catch (error) {
		const message = error instanceof Error ? error.message : 'Failed to fetch documents';
		documentsStore.setError(message);
		throw error;
	} finally {
		documentsStore.setLoading(false);
	}
}

export async function fetchDocument(id: string): Promise<Document> {
	activeDocumentStore.setLoading(true);
	try {
		const context = await contextsApi.getContext(id);
		const document = mapContextToDocument(context);
		documentsStore.setDocument(document);
		return document;
	} catch (error) {
		const message = error instanceof Error ? error.message : 'Failed to fetch document';
		activeDocumentStore.setError(message);
		throw error;
	} finally {
		activeDocumentStore.setLoading(false);
	}
}

export async function createDocument(params: {
	title: string;
	type?: DocumentType;
	parent_id?: string | null;
	icon?: string | null;
	content?: Block[];
}): Promise<Document> {
	try {
		const context = await contextsApi.createContext({
			name: params.title || 'Untitled',
			type: 'document',
			parent_id: params.parent_id || undefined,
			icon: typeof params.icon === 'string' ? params.icon : undefined,
			blocks: params.content ? mapDocBlocksToContextBlocks(params.content) : [
				{ id: crypto.randomUUID(), type: 'paragraph', content: '' }
			]
		});

		const document = mapContextToDocument(context);
		documentsStore.setDocument(document);

		// Expand parent if exists
		if (params.parent_id) {
			treeStore.setExpanded(params.parent_id, true);
		}

		return document;
	} catch (error) {
		const message = error instanceof Error ? error.message : 'Failed to create document';
		documentsStore.setError(message);
		throw error;
	}
}

export async function updateDocument(id: string, updates: Partial<Document>): Promise<Document> {
	activeDocumentStore.setSaving(true);
	try {
		// Map document updates to context updates
		const contextUpdates: Parameters<typeof contextsApi.updateContext>[1] = {};

		if (updates.title !== undefined) contextUpdates.name = updates.title;
		if (updates.icon !== undefined) {
			contextUpdates.icon = typeof updates.icon === 'object' && updates.icon?.value
				? updates.icon.value
				: (updates.icon as string | undefined);
		}
		if (updates.cover !== undefined) contextUpdates.cover_image = updates.cover || undefined;
		if (updates.parent_id !== undefined) contextUpdates.parent_id = updates.parent_id;
		if (updates.is_archived !== undefined) contextUpdates.is_archived = updates.is_archived;
		if (updates.is_template !== undefined) contextUpdates.is_template = updates.is_template;
		if (updates.content !== undefined) {
			contextUpdates.blocks = mapDocBlocksToContextBlocks(updates.content);
		}
		if (updates.is_favorite !== undefined) {
			contextUpdates.properties = { is_favorite: updates.is_favorite };
		}

		const context = await contextsApi.updateContext(id, contextUpdates);
		const document = mapContextToDocument(context);
		documentsStore.updateDocument(id, document);
		activeDocumentStore.setLastSaved(new Date().toISOString());
		return document;
	} catch (error) {
		const message = error instanceof Error ? error.message : 'Failed to update document';
		activeDocumentStore.setError(message);
		throw error;
	} finally {
		activeDocumentStore.setSaving(false);
	}
}

export async function deleteDocument(id: string, permanent = false): Promise<void> {
	try {
		if (permanent) {
			await contextsApi.deleteContext(id);
			documentsStore.removeDocument(id);
		} else {
			await contextsApi.archiveContext(id);
			documentsStore.updateDocument(id, { is_archived: true });
		}
	} catch (error) {
		const message = error instanceof Error ? error.message : 'Failed to delete document';
		documentsStore.setError(message);
		throw error;
	}
}

export async function restoreDocument(id: string): Promise<void> {
	try {
		await contextsApi.unarchiveContext(id);
		documentsStore.updateDocument(id, { is_archived: false });
	} catch (error) {
		const message = error instanceof Error ? error.message : 'Failed to restore document';
		documentsStore.setError(message);
		throw error;
	}
}

export async function moveDocument(id: string, newParentId: string | null): Promise<void> {
	try {
		await contextsApi.updateContext(id, { parent_id: newParentId });
		documentsStore.updateDocument(id, { parent_id: newParentId });

		// Expand new parent
		if (newParentId) {
			treeStore.setExpanded(newParentId, true);
		}
	} catch (error) {
		const message = error instanceof Error ? error.message : 'Failed to move document';
		documentsStore.setError(message);
		throw error;
	}
}

export async function duplicateDocument(id: string): Promise<Document> {
	try {
		const context = await contextsApi.duplicateContext(id);
		const document = mapContextToDocument(context);
		documentsStore.setDocument(document);
		return document;
	} catch (error) {
		const message = error instanceof Error ? error.message : 'Failed to duplicate document';
		documentsStore.setError(message);
		throw error;
	}
}

export async function toggleFavorite(id: string): Promise<void> {
	const doc = documentsStore.getDocument(id);
	if (!doc) return;

	await updateDocument(id, { is_favorite: !doc.is_favorite });
}

// ============================================================================
// Block Operations
// ============================================================================

export function createEmptyParagraph(): Block {
	return {
		id: crypto.randomUUID(),
		type: 'paragraph',
		content: [],
		properties: {},
		children: [],
		created_at: new Date().toISOString(),
		updated_at: new Date().toISOString()
	};
}

export function createBlock(type: BlockType, content?: string): Block {
	const richText: RichText[] = content
		? [
				{
					type: 'text',
					text: { content, link: null },
					annotations: {
						bold: false,
						italic: false,
						strikethrough: false,
						underline: false,
						code: false,
						color: 'default'
					},
					plain_text: content,
					href: null
				}
			]
		: [];

	// Initialize type-specific properties
	let properties: Record<string, unknown> = {};

	switch (type) {
		case 'to_do':
			properties = { checked: false };
			break;
		case 'toggle':
			properties = { expanded: false };
			break;
		case 'callout':
			properties = { icon: 'Lightbulb' };
			break;
		case 'code':
			properties = { language: 'plaintext' };
			break;
		case 'table':
			properties = {
				tableData: {
					rows: [
						['', '', ''],
						['', '', '']
					],
					headerRow: true
				}
			};
			break;
		case 'image':
			properties = { url: null, caption: '' };
			break;
		case 'bookmark':
			properties = { url: null, title: '', description: '' };
			break;
		default:
			properties = {};
	}

	return {
		id: crypto.randomUUID(),
		type,
		content: richText,
		properties,
		children: [],
		created_at: new Date().toISOString(),
		updated_at: new Date().toISOString()
	};
}

// ============================================================================
// Search Operations
// ============================================================================

export async function searchDocuments(query: string): Promise<DocumentMeta[]> {
	if (!query.trim()) return [];

	try {
		const contexts = await contextsApi.getContexts({ search: query });
		return contexts.map(mapContextToDocMeta);
	} catch (error) {
		console.error('Search error:', error);
		return [];
	}
}

// ============================================================================
// Navigation Helpers
// ============================================================================

export function openDocument(id: string) {
	activeDocumentStore.setActiveDocument(id);
}

export function closeDocument() {
	activeDocumentStore.setActiveDocument(null);
}

export async function openAndFetchDocument(id: string): Promise<Document> {
	activeDocumentStore.setActiveDocument(id);
	return await fetchDocument(id);
}

// ============================================================================
// Profile Operations
// ============================================================================

export type ProfileType = 'person' | 'business' | 'project';

export async function fetchProfiles(type?: ProfileType): Promise<DocumentMeta[]> {
	documentsStore.setLoading(true);
	try {
		const contexts = await contextsApi.getContexts({
			type: type || undefined
		});

		// Filter by profile types if no specific type given
		const profileTypes: ProfileType[] = ['person', 'business', 'project'];
		const filtered = type
			? contexts
			: contexts.filter((ctx) => profileTypes.includes(ctx.type as ProfileType));

		const documents = filtered.map(mapContextToDocMeta);
		return documents;
	} catch (error) {
		const message = error instanceof Error ? error.message : 'Failed to fetch profiles';
		documentsStore.setError(message);
		throw error;
	} finally {
		documentsStore.setLoading(false);
	}
}

export async function createProfile(params: {
	name: string;
	type: ProfileType;
	description?: string;
	icon?: string | null;
	properties?: Record<string, unknown>;
}): Promise<Document> {
	try {
		const context = await contextsApi.createContext({
			name: params.name || 'Untitled',
			type: params.type,
			icon: typeof params.icon === 'string' ? params.icon : undefined,
			properties: params.properties,
			blocks: params.description
				? [{ id: crypto.randomUUID(), type: 'paragraph', content: params.description }]
				: [{ id: crypto.randomUUID(), type: 'paragraph', content: '' }]
		});

		const document = mapContextToDocument(context);
		documentsStore.setDocument(document);
		return document;
	} catch (error) {
		const message = error instanceof Error ? error.message : 'Failed to create profile';
		documentsStore.setError(message);
		throw error;
	}
}

// Default icons for profile types (SVG icon names from PageIconPicker)
export const defaultProfileIcons: Record<ProfileType, string> = {
	person: 'User',
	business: 'Building',
	project: 'FolderKanban'
};
