export interface Message {
  id: string;
  role: 'user' | 'assistant' | 'system';
  content: string;
  created_at: string;
  message_metadata?: Record<string, unknown>;
}

export interface Conversation {
  id: string;
  title: string;
  context_id: string | null;
  created_at: string;
  updated_at: string;
  messages: Message[];
  message_count?: number;
}

export interface SearchResult {
  message_id: string;
  conversation_id: string;
  content: string;
  role: string;
  created_at: string;
}
