// Calendar API Types

export type MeetingType = 'team' | 'sales' | 'onboarding' | 'kickoff' | 'implementation' | 'standup' | 'retrospective' | 'planning' | 'review' | 'one_on_one' | 'client' | 'internal' | 'external' | 'other';
export type EventSource = 'google' | 'businessos';

export interface GoogleConnectionStatus {
  connected: boolean;
  email?: string;
  connected_at?: string;
}

export interface CalendarAttendee {
  email: string;
  name?: string;
  response_status?: string;
}

export interface ExternalLink {
  name: string;
  url: string;
  type?: string;
}

export interface ActionItem {
  id: string;
  text: string;
  completed: boolean;
  assignee_id?: string;
  due_date?: string;
}

export interface CalendarEvent {
  id: string;
  user_id: string;
  google_event_id: string | null;
  calendar_id: string | null;
  title: string | null;
  description: string | null;
  start_time: string;
  end_time: string;
  all_day: boolean;
  location: string | null;
  attendees: CalendarAttendee[];
  status: string | null;
  visibility: string | null;
  html_link: string | null;
  source: EventSource;
  meeting_type: MeetingType;
  context_id: string | null;
  project_id: string | null;
  client_id: string | null;
  recording_url: string | null;
  meeting_link: string | null;
  external_links: ExternalLink[];
  meeting_notes: string | null;
  meeting_summary: string | null;
  action_items: ActionItem[];
  synced_at: string | null;
  created_at: string;
  updated_at: string;
}

export interface CreateCalendarEventData {
  title: string;
  description?: string;
  start_time: string;
  end_time: string;
  all_day?: boolean;
  location?: string;
  attendees?: CalendarAttendee[];
  meeting_type?: MeetingType;
  context_id?: string;
  project_id?: string;
  client_id?: string;
  recording_url?: string;
  meeting_link?: string;
  external_links?: ExternalLink[];
  meeting_notes?: string;
  action_items?: ActionItem[];
}

export interface UpdateCalendarEventData {
  title?: string;
  description?: string;
  start_time?: string;
  end_time?: string;
  all_day?: boolean;
  location?: string;
  attendees?: CalendarAttendee[];
  meeting_type?: MeetingType;
  context_id?: string | null;
  project_id?: string | null;
  client_id?: string | null;
  recording_url?: string;
  meeting_link?: string;
  external_links?: ExternalLink[];
  meeting_notes?: string;
  action_items?: ActionItem[];
}
