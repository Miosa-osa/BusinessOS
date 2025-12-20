import { request } from '../base';
import type {
  GoogleAuthResponse,
  GoogleConnectionStatus,
  SlackAuthResponse,
  SlackConnectionStatus,
  SlackChannelsResponse,
  SlackNotificationsResponse,
  NotionAuthResponse,
  NotionConnectionStatus,
  NotionDatabasesResponse,
  NotionPagesResponse,
  NotionSyncResponse
} from './types';

// ============================================
// Google OAuth Integration
// ============================================
export async function initiateGoogleAuth() {
  return request<GoogleAuthResponse>('/integrations/google/auth');
}

export async function getGoogleConnectionStatus() {
  return request<GoogleConnectionStatus>('/integrations/google/status');
}

export async function disconnectGoogle() {
  return request('/integrations/google', { method: 'DELETE' });
}

// ============================================
// Slack Integration
// ============================================

/**
 * Initiate Slack OAuth flow
 * Returns URL to redirect user to for Slack authorization
 */
export async function initiateSlackAuth() {
  return request<SlackAuthResponse>('/integrations/slack/auth');
}

/**
 * Get current Slack connection status
 */
export async function getSlackConnectionStatus() {
  return request<SlackConnectionStatus>('/integrations/slack/status');
}

/**
 * Disconnect Slack integration
 */
export async function disconnectSlack() {
  return request('/integrations/slack', { method: 'DELETE' });
}

/**
 * Get list of Slack channels the user has access to
 */
export async function getSlackChannels() {
  return request<SlackChannelsResponse>('/integrations/slack/channels');
}

/**
 * Get Slack notifications/messages
 * @param limit - Number of notifications to fetch (default 50)
 * @param cursor - Pagination cursor for fetching more results
 */
export async function getSlackNotifications(limit = 50, cursor?: string) {
  const params = new URLSearchParams({ limit: String(limit) });
  if (cursor) params.append('cursor', cursor);
  return request<SlackNotificationsResponse>(`/integrations/slack/notifications?${params}`);
}

// ============================================
// Notion Integration
// ============================================

/**
 * Initiate Notion OAuth flow
 * Returns URL to redirect user to for Notion authorization
 */
export async function initiateNotionAuth() {
  return request<NotionAuthResponse>('/integrations/notion/auth');
}

/**
 * Get current Notion connection status
 */
export async function getNotionConnectionStatus() {
  return request<NotionConnectionStatus>('/integrations/notion/status');
}

/**
 * Disconnect Notion integration
 */
export async function disconnectNotion() {
  return request('/integrations/notion', { method: 'DELETE' });
}

/**
 * Get list of Notion databases the integration has access to
 */
export async function getNotionDatabases() {
  return request<NotionDatabasesResponse>('/integrations/notion/databases');
}

/**
 * Get Notion pages
 * @param databaseId - Optional database ID to filter pages
 * @param cursor - Pagination cursor for fetching more results
 */
export async function getNotionPages(databaseId?: string, cursor?: string) {
  const params = new URLSearchParams();
  if (databaseId) params.append('database_id', databaseId);
  if (cursor) params.append('cursor', cursor);
  const query = params.toString();
  return request<NotionPagesResponse>(`/integrations/notion/pages${query ? `?${query}` : ''}`);
}

/**
 * Sync a Notion database to BusinessOS
 * @param databaseId - The Notion database ID to sync
 */
export async function syncNotionDatabase(databaseId: string) {
  return request<NotionSyncResponse>('/integrations/notion/sync', {
    method: 'POST',
    body: { database_id: databaseId }
  });
}
