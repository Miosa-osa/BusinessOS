import { request } from '../base';
import type { CalendarEvent, CreateCalendarEventData, UpdateCalendarEventData } from './types';

// ============================================
// Calendar API - Uses new integration infrastructure
// All routes now under /integrations/google/calendar/*
// ============================================

const CALENDAR_BASE = '/integrations/google/calendar';

export async function getCalendarEvents(filters?: { start?: string; end?: string; meetingType?: string; contextId?: string; projectId?: string; clientId?: string }) {
  const params = new URLSearchParams();
  if (filters?.start) params.set('start', filters.start);
  if (filters?.end) params.set('end', filters.end);
  if (filters?.meetingType) params.set('meeting_type', filters.meetingType);
  if (filters?.contextId) params.set('context_id', filters.contextId);
  if (filters?.projectId) params.set('project_id', filters.projectId);
  if (filters?.clientId) params.set('client_id', filters.clientId);
  const query = params.toString();
  return request<CalendarEvent[]>(`${CALENDAR_BASE}/events${query ? `?${query}` : ''}`);
}

export async function getCalendarEvent(id: string) {
  return request<CalendarEvent>(`${CALENDAR_BASE}/events/${id}`);
}

export async function createCalendarEvent(data: CreateCalendarEventData) {
  return request<CalendarEvent>(`${CALENDAR_BASE}/events`, { method: 'POST', body: data });
}

export async function updateCalendarEvent(id: string, data: UpdateCalendarEventData) {
  return request<CalendarEvent>(`${CALENDAR_BASE}/events/${id}`, { method: 'PUT', body: data });
}

export async function deleteCalendarEvent(id: string) {
  return request(`${CALENDAR_BASE}/events/${id}`, { method: 'DELETE' });
}

export async function syncCalendar(): Promise<{ message: string; synced_count: number }> {
  return request<{ message: string; synced_count: number }>(`${CALENDAR_BASE}/sync`, { method: 'POST' });
}

export async function getTodayEvents() {
  // Today events uses query param on main events endpoint
  const today = new Date().toISOString().split('T')[0];
  const params = new URLSearchParams({ start: today, end: today });
  return request<CalendarEvent[]>(`${CALENDAR_BASE}/events?${params}`);
}

export async function getUpcomingEvents(limit?: number) {
  const params = new URLSearchParams();
  if (limit) params.set('limit', String(limit));
  const query = params.toString();
  return request<CalendarEvent[]>(`${CALENDAR_BASE}/events${query ? `?${query}` : ''}`);
}

// NOTE: getGoogleConnectionStatus is in integrations module
// Use: import { getGoogleConnectionStatus } from '$lib/api/integrations'
