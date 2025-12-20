// Integrations API Types

// Google
export interface GoogleAuthResponse {
  auth_url: string;
}

export interface GoogleConnectionStatus {
  connected: boolean;
  email?: string;
  connected_at?: string;
}

// Slack
export interface SlackAuthResponse {
  auth_url: string;
}

export interface SlackConnectionStatus {
  connected: boolean;
  workspace_name?: string;
  workspace_id?: string;
  user_name?: string;
  user_id?: string;
  connected_at?: string;
}

export interface SlackChannel {
  id: string;
  name: string;
  is_private: boolean;
  is_member: boolean;
  num_members?: number;
}

export interface SlackNotification {
  id: string;
  channel_id: string;
  channel_name: string;
  text: string;
  user_id: string;
  user_name: string;
  timestamp: string;
  thread_ts?: string;
}

export interface SlackChannelsResponse {
  channels: SlackChannel[];
}

export interface SlackNotificationsResponse {
  notifications: SlackNotification[];
  has_more: boolean;
}

// Notion
export interface NotionAuthResponse {
  auth_url: string;
}

export interface NotionConnectionStatus {
  connected: boolean;
  workspace_name?: string;
  workspace_id?: string;
  workspace_icon?: string;
  connected_at?: string;
}

export interface NotionDatabase {
  id: string;
  title: string;
  description?: string;
  icon?: string;
  url: string;
  created_time: string;
  last_edited_time: string;
}

export interface NotionPage {
  id: string;
  title: string;
  icon?: string;
  url: string;
  parent_type: 'database' | 'page' | 'workspace';
  parent_id?: string;
  created_time: string;
  last_edited_time: string;
}

export interface NotionDatabasesResponse {
  databases: NotionDatabase[];
}

export interface NotionPagesResponse {
  pages: NotionPage[];
  has_more: boolean;
  next_cursor?: string;
}

export interface NotionSyncResponse {
  success: boolean;
  synced_count: number;
  message?: string;
}
