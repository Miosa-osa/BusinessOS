import { request } from '../base';
import type { CalendarEvent, CreateCalendarEventData, UpdateCalendarEventData, GoogleConnectionStatus } from './types';

export async function getCalendarEvents(filters?: { start?: string; end?: string; meetingType?: string; contextId?: string; projectId?: string; clientId?: string }) {
  const params = new URLSearchParams();
  if (filters?.start) params.set('start', filters.start);
  if (filters?.end) params.set('end', filters.end);
  if (filters?.meetingType) params.set('meeting_type', filters.meetingType);
  if (filters?.contextId) params.set('context_id', filters.contextId);
  if (filters?.projectId) params.set('project_id', filters.projectId);
  if (filters?.clientId) params.set('client_id', filters.clientId);
  const query = params.toString();
  return request<CalendarEvent[]>(`/calendar/events${query ? `?${query}` : ''}`);
}

export async function getCalendarEvent(id: string) {
  return request<CalendarEvent>(`/calendar/events/${id}`);
}

export async function createCalendarEvent(data: CreateCalendarEventData) {
  return request<CalendarEvent>('/calendar/events', { method: 'POST', body: data });
}

export async function updateCalendarEvent(id: string, data: UpdateCalendarEventData) {
  return request<CalendarEvent>(`/calendar/events/${id}`, { method: 'PUT', body: data });
}

export async function deleteCalendarEvent(id: string) {
  return request(`/calendar/events/${id}`, { method: 'DELETE' });
}

export async function syncCalendar() {
  return request(`/calendar/sync`, { method: 'POST' });
}

export async function getTodayEvents() {
  return request<CalendarEvent[]>('/calendar/today');
}

export async function getUpcomingEvents(limit?: number) {
  const params = limit ? `?limit=${limit}` : '';
  return request<CalendarEvent[]>(`/calendar/upcoming${params}`);
}

export async function getGoogleConnectionStatus() {
  return request<GoogleConnectionStatus>('/integrations/google/status');
}
