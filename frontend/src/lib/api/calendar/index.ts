export * from './types';
export * from './calendar';

import * as calendarApi from './calendar';

export const api = {
  getCalendarEvents: calendarApi.getCalendarEvents,
  getCalendarEvent: calendarApi.getCalendarEvent,
  createCalendarEvent: calendarApi.createCalendarEvent,
  updateCalendarEvent: calendarApi.updateCalendarEvent,
  deleteCalendarEvent: calendarApi.deleteCalendarEvent,
  syncCalendar: calendarApi.syncCalendar,
  getTodayEvents: calendarApi.getTodayEvents,
  getUpcomingEvents: calendarApi.getUpcomingEvents,
  getGoogleConnectionStatus: calendarApi.getGoogleConnectionStatus,
};

export default api;
