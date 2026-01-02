import { request, getApiBaseUrl } from '../base';
import type { Conversation, Message, SearchResult } from './types';

export async function getConversations(): Promise<Conversation[]> {
  return request<Conversation[]>('/chat/conversations');
}

export async function getConversation(id: string): Promise<Conversation> {
  const response = await request<{ conversation: Conversation; messages: Message[] }>(`/chat/conversations/${id}`);
  return {
    ...response.conversation,
    messages: response.messages || [],
    message_count: response.messages?.length || 0
  } as Conversation;
}

export async function createConversation(title?: string, contextId?: string): Promise<Conversation> {
  return request<Conversation>('/chat/conversations', {
    method: 'POST',
    body: { title, context_id: contextId }
  });
}

export async function deleteConversation(id: string) {
  return request(`/chat/conversations/${id}`, { method: 'DELETE' });
}

export async function updateConversation(id: string, data: { title?: string; context_id?: string | null }) {
  return request<Conversation>(`/chat/conversations/${id}`, { method: 'PUT', body: data });
}

export async function getConversationsByContext(contextId: string) {
  return request<Conversation[]>(`/chat/conversations?context_id=${encodeURIComponent(contextId)}`);
}

// sendMessage returns a streaming response and conversation id header
export async function sendMessage(
  message: string,
  conversationId?: string,
  contextId?: string,
  model?: string,
  options?: {
    structured_output?: boolean;
    output_style?: string;
  }
) {
  const response = await fetch(`${getApiBaseUrl()}/chat/message`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    credentials: 'include',
    body: JSON.stringify({
      message,
      conversation_id: conversationId,
      context_id: contextId,
      model,
      structured_output: options?.structured_output,
      output_style: options?.output_style
    })
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({ detail: 'Chat failed' }));
    throw new Error(error.detail || 'Chat failed');
  }

  return {
    stream: response.body,
    conversationId: response.headers.get('X-Conversation-Id')
  };
}

export async function searchConversations(query: string) {
  return request<SearchResult[]>(`/chat/search?q=${encodeURIComponent(query)}`);
}
