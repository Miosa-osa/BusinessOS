import { describe, it, expect, vi, beforeEach } from 'vitest';
import { get } from 'svelte/store';
import { chat } from './chat';
import { mockConversation, mockMessage } from '$lib/test-utils/mocks';

// Mock the API module
vi.mock('$lib/api/conversations', () => ({
	api: {
		getConversations: vi.fn(),
		getConversation: vi.fn(),
		sendMessage: vi.fn(),
		deleteConversation: vi.fn(),
		searchConversations: vi.fn()
	}
}));

import { api } from '$lib/api/conversations';
const mockedApi = vi.mocked(api);

describe('Chat Store', () => {
	beforeEach(() => {
		vi.clearAllMocks();
		// Reset store to initial state
		chat.newConversation();
	});

	describe('loadConversations', () => {
		it('loads conversations successfully', async () => {
			const mockConversations = [mockConversation(), mockConversation()];
			mockedApi.getConversations.mockResolvedValueOnce(mockConversations);

			await chat.loadConversations();

			const state = get(chat);
			expect(state.conversations).toEqual(mockConversations);
			expect(state.loading).toBe(false);
		});

		it('sets loading state during fetch', async () => {
			mockedApi.getConversations.mockImplementation(
				() => new Promise((resolve) => setTimeout(resolve, 100))
			);

			const loadPromise = chat.loadConversations();
			const stateDuringLoad = get(chat);
			expect(stateDuringLoad.loading).toBe(true);

			await loadPromise;
		});

		it('handles errors gracefully', async () => {
			const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
			mockedApi.getConversations.mockRejectedValueOnce(new Error('API error'));

			await chat.loadConversations();

			const state = get(chat);
			expect(state.loading).toBe(false);
			expect(consoleErrorSpy).toHaveBeenCalled();

			consoleErrorSpy.mockRestore();
		});
	});

	describe('loadConversation', () => {
		it('loads a specific conversation with messages', async () => {
			const conversation = mockConversation({
				id: 'conv-123',
				messages: [mockMessage(), mockMessage()]
			});
			mockedApi.getConversation.mockResolvedValueOnce(conversation);

			await chat.loadConversation('conv-123');

			const state = get(chat);
			expect(state.currentConversation).toEqual(conversation);
			expect(state.messages).toEqual(conversation.messages);
			expect(state.loading).toBe(false);
		});

		it('handles missing messages array', async () => {
			const conversation = mockConversation({ id: 'conv-123' });
			delete (conversation as any).messages;
			mockedApi.getConversation.mockResolvedValueOnce(conversation);

			await chat.loadConversation('conv-123');

			const state = get(chat);
			expect(state.messages).toEqual([]);
		});
	});

	describe('newConversation', () => {
		it('resets current conversation and messages', async () => {
			// First set a conversation
			const conversation = mockConversation();
			mockedApi.getConversation.mockResolvedValueOnce(conversation);
			await chat.loadConversation('conv-123');

			// Then create new conversation
			await chat.newConversation();

			const state = get(chat);
			expect(state.currentConversation).toBeNull();
			expect(state.messages).toEqual([]);
		});
	});

	describe('deleteConversation', () => {
		it('deletes a conversation and removes it from list', async () => {
			const conversations = [
				mockConversation({ id: 'conv-1' }),
				mockConversation({ id: 'conv-2' })
			];
			mockedApi.getConversations.mockResolvedValueOnce(conversations);
			await chat.loadConversations();

			mockedApi.deleteConversation.mockResolvedValueOnce(undefined);
			await chat.deleteConversation('conv-1');

			const state = get(chat);
			expect(state.conversations).toHaveLength(1);
			expect(state.conversations[0].id).toBe('conv-2');
		});

		it('clears current conversation if it is deleted', async () => {
			const conversation = mockConversation({ id: 'conv-123' });
			mockedApi.getConversation.mockResolvedValueOnce(conversation);
			await chat.loadConversation('conv-123');

			mockedApi.deleteConversation.mockResolvedValueOnce(undefined);
			await chat.deleteConversation('conv-123');

			const state = get(chat);
			expect(state.currentConversation).toBeNull();
			expect(state.messages).toEqual([]);
		});

		it('preserves current conversation if different one is deleted', async () => {
			const conversation = mockConversation({ id: 'conv-current' });
			mockedApi.getConversation.mockResolvedValueOnce(conversation);
			await chat.loadConversation('conv-current');

			mockedApi.deleteConversation.mockResolvedValueOnce(undefined);
			await chat.deleteConversation('conv-other');

			const state = get(chat);
			expect(state.currentConversation?.id).toBe('conv-current');
		});
	});

	describe('search', () => {
		it('returns search results', async () => {
			const mockResults = [
				{
					message_id: 'msg-1',
					conversation_id: 'conv-1',
					content: 'Found content',
					role: 'user',
					created_at: new Date().toISOString()
				}
			];
			mockedApi.searchConversations.mockResolvedValueOnce(mockResults);

			const results = await chat.search('test query');

			expect(results).toEqual(mockResults);
			expect(mockedApi.searchConversations).toHaveBeenCalledWith('test query');
		});

		it('returns empty array on error', async () => {
			const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
			mockedApi.searchConversations.mockRejectedValueOnce(new Error('Search failed'));

			const results = await chat.search('test query');

			expect(results).toEqual([]);
			expect(consoleErrorSpy).toHaveBeenCalled();

			consoleErrorSpy.mockRestore();
		});
	});

	describe('archive functionality', () => {
		it('toggles archive view', () => {
			const initialState = get(chat);
			expect(initialState.showArchived).toBe(false);

			chat.toggleArchiveView();
			expect(get(chat).showArchived).toBe(true);

			chat.toggleArchiveView();
			expect(get(chat).showArchived).toBe(false);
		});

		it('sets archive view state', () => {
			chat.setShowArchived(true);
			expect(get(chat).showArchived).toBe(true);

			chat.setShowArchived(false);
			expect(get(chat).showArchived).toBe(false);
		});

		it('archives a conversation (client-side)', async () => {
			const conversations = [
				mockConversation({ id: 'conv-1' }),
				mockConversation({ id: 'conv-2' })
			];
			mockedApi.getConversations.mockResolvedValueOnce(conversations);
			await chat.loadConversations();

			await chat.archiveConversation('conv-1');

			const state = get(chat);
			expect(state.conversations).toHaveLength(1);
			expect(state.conversations[0].id).toBe('conv-2');
			expect(state.archivedConversations).toHaveLength(1);
			expect(state.archivedConversations[0].id).toBe('conv-1');
			expect(state.archivedConversations[0].is_archived).toBe(true);
		});

		it('unarchives a conversation (client-side)', async () => {
			// First archive a conversation
			const conversations = [mockConversation({ id: 'conv-1' })];
			mockedApi.getConversations.mockResolvedValueOnce(conversations);
			await chat.loadConversations();
			await chat.archiveConversation('conv-1');

			// Then unarchive it
			await chat.unarchiveConversation('conv-1');

			const state = get(chat);
			expect(state.archivedConversations).toHaveLength(0);
			expect(state.conversations).toHaveLength(1);
			expect(state.conversations[0].id).toBe('conv-1');
			expect(state.conversations[0].is_archived).toBe(false);
		});
	});

	describe('pinConversation', () => {
		it('toggles pin status', async () => {
			const conversations = [mockConversation({ id: 'conv-1', pinned: false })];
			mockedApi.getConversations.mockResolvedValueOnce(conversations);
			await chat.loadConversations();

			await chat.pinConversation('conv-1');
			expect(get(chat).conversations[0].pinned).toBe(true);

			await chat.pinConversation('conv-1');
			expect(get(chat).conversations[0].pinned).toBe(false);
		});
	});

	describe('renameConversation', () => {
		it('renames a conversation in the list', async () => {
			const conversations = [mockConversation({ id: 'conv-1', title: 'Old Title' })];
			mockedApi.getConversations.mockResolvedValueOnce(conversations);
			await chat.loadConversations();

			await chat.renameConversation('conv-1', 'New Title');

			const state = get(chat);
			expect(state.conversations[0].title).toBe('New Title');
		});

		it('renames current conversation if it matches', async () => {
			const conversation = mockConversation({ id: 'conv-1', title: 'Old Title' });
			mockedApi.getConversation.mockResolvedValueOnce(conversation);
			await chat.loadConversation('conv-1');

			await chat.renameConversation('conv-1', 'New Title');

			const state = get(chat);
			expect(state.currentConversation?.title).toBe('New Title');
		});
	});

	describe('sendMessage', () => {
		it('adds user message optimistically', async () => {
			const mockStream = new ReadableStream({
				start(controller) {
					controller.enqueue(new TextEncoder().encode('Hello'));
					controller.close();
				}
			});

			mockedApi.sendMessage.mockResolvedValueOnce({
				stream: mockStream,
				conversationId: 'conv-new'
			});
			mockedApi.getConversations.mockResolvedValueOnce([]);

			// We need to wait for the promise but can check state during
			const sendPromise = chat.sendMessage('Test message');

			// Small delay to let optimistic update happen
			await new Promise((resolve) => setTimeout(resolve, 10));

			const stateDuringStream = get(chat);
			const userMessage = stateDuringStream.messages.find((m) => m.role === 'user');
			expect(userMessage).toBeDefined();
			expect(userMessage?.content).toBe('Test message');

			await sendPromise;
		});

		it('sets streaming state during message send', async () => {
			const mockStream = new ReadableStream({
				start(controller) {
					setTimeout(() => {
						controller.enqueue(new TextEncoder().encode('Response'));
						controller.close();
					}, 50);
				}
			});

			mockedApi.sendMessage.mockResolvedValueOnce({
				stream: mockStream,
				conversationId: 'conv-123'
			});
			mockedApi.getConversations.mockResolvedValueOnce([]);

			const sendPromise = chat.sendMessage('Test');

			// Check state during streaming
			await new Promise((resolve) => setTimeout(resolve, 20));
			const stateDuringStream = get(chat);
			expect(stateDuringStream.streaming).toBe(true);

			await sendPromise;

			// Check state after streaming completes
			const stateAfter = get(chat);
			expect(stateAfter.streaming).toBe(false);
		});
	});
});
