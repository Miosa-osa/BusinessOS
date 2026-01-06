// Gmail API functions
// Uses new integration infrastructure: /integrations/google/gmail/*
import { apiClient } from '../client';
import type { Email, ComposeEmail, GmailAccessStatus, GmailStats, SyncResult, GetEmailsParams } from './types';

// ============================================
// Gmail API - Uses new integration infrastructure
// All routes now under /integrations/google/gmail/*
// ============================================

const GMAIL_BASE = '/integrations/google/gmail';

/**
 * Check if the user has Gmail access
 * Uses the Google integration status endpoint
 */
export async function checkGmailAccess(): Promise<GmailAccessStatus> {
  const res = await apiClient.get('/integrations/google/status');
  if (!res.ok) {
    throw new Error('Failed to check Gmail access');
  }
  const data = await res.json();
  // Map the integration status to Gmail access status format
  return {
    has_access: data.connected && data.scopes?.includes('gmail'),
    requires_upgrade: data.connected && !data.scopes?.includes('gmail'),
    email: data.email
  };
}

/**
 * Request Gmail access (initiates OAuth flow with Gmail scopes)
 * Uses the Google auth endpoint which includes Gmail scopes
 */
export async function requestGmailAccess(): Promise<{ auth_url: string }> {
  const res = await apiClient.get('/integrations/google/auth');
  if (!res.ok) {
    throw new Error('Failed to request Gmail access');
  }
  return res.json();
}

/**
 * Get emails from the user's Gmail
 */
export async function getEmails(params?: GetEmailsParams): Promise<Email[]> {
  const searchParams = new URLSearchParams();
  if (params?.folder) searchParams.set('folder', params.folder);
  if (params?.limit) searchParams.set('limit', params.limit.toString());
  if (params?.offset) searchParams.set('offset', params.offset.toString());

  const res = await apiClient.get(`${GMAIL_BASE}/emails?${searchParams.toString()}`);
  if (!res.ok) {
    const data = await res.json();
    if (data.requires_upgrade) {
      throw new Error('REQUIRES_UPGRADE');
    }
    throw new Error('Failed to get emails');
  }
  return res.json();
}

/**
 * Get a single email by ID
 */
export async function getEmail(id: string): Promise<Email> {
  const res = await apiClient.get(`${GMAIL_BASE}/emails/${id}`);
  if (!res.ok) {
    throw new Error('Failed to get email');
  }
  return res.json();
}

/**
 * Send an email
 */
export async function sendEmail(email: ComposeEmail): Promise<{ message: string }> {
  const res = await apiClient.post(`${GMAIL_BASE}/emails/send`, email);
  if (!res.ok) {
    const data = await res.json();
    throw new Error(data.error || 'Failed to send email');
  }
  return res.json();
}

/**
 * Mark an email as read
 */
export async function markAsRead(id: string): Promise<{ message: string }> {
  const res = await apiClient.post(`${GMAIL_BASE}/emails/${id}/read`);
  if (!res.ok) {
    throw new Error('Failed to mark email as read');
  }
  return res.json();
}

/**
 * Archive an email
 */
export async function archiveEmail(id: string): Promise<{ message: string }> {
  const res = await apiClient.post(`${GMAIL_BASE}/emails/${id}/archive`);
  if (!res.ok) {
    throw new Error('Failed to archive email');
  }
  return res.json();
}

/**
 * Delete an email (moves to trash)
 */
export async function deleteEmail(id: string): Promise<{ message: string }> {
  const res = await apiClient.delete(`${GMAIL_BASE}/emails/${id}`);
  if (!res.ok) {
    throw new Error('Failed to delete email');
  }
  return res.json();
}

/**
 * Sync emails from Gmail
 */
export async function syncEmails(maxResults?: number): Promise<{ message: string; result: SyncResult }> {
  const searchParams = new URLSearchParams();
  if (maxResults) searchParams.set('max', maxResults.toString());

  const res = await apiClient.post(`${GMAIL_BASE}/sync?${searchParams.toString()}`);
  if (!res.ok) {
    const data = await res.json();
    if (data.requires_upgrade) {
      throw new Error('REQUIRES_UPGRADE');
    }
    throw new Error(data.error || 'Failed to sync emails');
  }
  return res.json();
}

/**
 * Get Gmail statistics
 * NOTE: Stats endpoint not in new infrastructure, returns empty stats for now
 */
export async function getGmailStats(): Promise<GmailStats> {
  // Stats can be computed from emails or added to new infrastructure later
  return {
    total_emails: 0,
    unread_count: 0,
    last_sync: null
  };
}
