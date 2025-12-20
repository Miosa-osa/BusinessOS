import { writable, derived } from 'svelte/store';
import { api, type Conversation, type Message } from '$lib/api/conversations';

interface ChatState {
	conversations: Conversation[];
	currentConversation: Conversation | null;
	messages: Message[];
	loading: boolean;
	streaming: boolean;
	streamingContent: string;
}

function createChatStore() {
	const { subscribe, set, update } = writable<ChatState>({
		conversations: [],
		currentConversation: null,
		messages: [],
		loading: false,
		streaming: false,
		streamingContent: ''
	});

	return {
		subscribe,

		async loadConversations() {
			update((s) => ({ ...s, loading: true }));
			try {
				const conversations = await api.getConversations();
				update((s) => ({ ...s, conversations, loading: false }));
			} catch (error) {
				console.error('Failed to load conversations:', error);
				update((s) => ({ ...s, loading: false }));
			}
		},

		async loadConversation(id: string) {
			update((s) => ({ ...s, loading: true }));
			try {
				const conversation = await api.getConversation(id);
				update((s) => ({
					...s,
					currentConversation: conversation,
					messages: conversation.messages || [],
					loading: false
				}));
			} catch (error) {
				console.error('Failed to load conversation:', error);
				update((s) => ({ ...s, loading: false }));
			}
		},

		async sendMessage(content: string, contextId?: string, model?: string) {
			let currentState: ChatState;
			const unsubscribe = subscribe((s) => (currentState = s));
			unsubscribe();

			// Add user message optimistically
			const userMessage: Message = {
				id: crypto.randomUUID(),
				role: 'user',
				content,
				created_at: new Date().toISOString()
			};

			update((s) => ({
				...s,
				messages: [...s.messages, userMessage],
				streaming: true,
				streamingContent: ''
			}));

			try {
				const { stream, conversationId } = await api.sendMessage(
					content,
					currentState!.currentConversation?.id,
					contextId,
					model
				);

				if (!stream) throw new Error('No stream returned');

				// Update conversation ID if new
				if (conversationId && !currentState!.currentConversation) {
					update((s) => ({
						...s,
						currentConversation: {
							id: conversationId,
							title: content.slice(0, 50),
							context_id: contextId || null,
							created_at: new Date().toISOString(),
							updated_at: new Date().toISOString(),
							messages: []
						}
					}));
				}

				// Read the stream
				const reader = stream.getReader();
				const decoder = new TextDecoder();
				let fullContent = '';

				while (true) {
					const { done, value } = await reader.read();
					if (done) break;

					const chunk = decoder.decode(value, { stream: true });
					fullContent += chunk;
					update((s) => ({ ...s, streamingContent: fullContent }));
				}

				// Add assistant message
				const assistantMessage: Message = {
					id: crypto.randomUUID(),
					role: 'assistant',
					content: fullContent,
					created_at: new Date().toISOString()
				};

				update((s) => ({
					...s,
					messages: [...s.messages, assistantMessage],
					streaming: false,
					streamingContent: ''
				}));

				// Refresh conversations list
				this.loadConversations();
			} catch (error) {
				console.error('Failed to send message:', error);
				update((s) => ({ ...s, streaming: false, streamingContent: '' }));
			}
		},

		async newConversation() {
			update((s) => ({
				...s,
				currentConversation: null,
				messages: []
			}));
		},

		async deleteConversation(id: string) {
			try {
				await api.deleteConversation(id);
				update((s) => ({
					...s,
					conversations: s.conversations.filter((c) => c.id !== id),
					currentConversation: s.currentConversation?.id === id ? null : s.currentConversation,
					messages: s.currentConversation?.id === id ? [] : s.messages
				}));
			} catch (error) {
				console.error('Failed to delete conversation:', error);
			}
		},

		async search(query: string) {
			try {
				return await api.searchConversations(query);
			} catch (error) {
				console.error('Search failed:', error);
				return [];
			}
		}
	};
}

export const chat = createChatStore();
