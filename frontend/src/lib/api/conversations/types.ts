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
  /** Preview text from the last message */
  preview?: string;
  /** Whether this conversation is archived */
  is_archived?: boolean;
  /** Type of conversation - regular chat or focus mode session */
  conversation_type?: 'chat' | 'focus';
  /** Associated project ID if linked */
  project_id?: string;
  /** Whether this conversation is pinned */
  pinned?: boolean;
}

export interface SearchResult {
  message_id: string;
  conversation_id: string;
  content: string;
  role: string;
  created_at: string;
}
