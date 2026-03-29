import type { CalendarEvent, MeetingType } from "$lib/api";

export type ViewMode = "day" | "week" | "month" | "agenda";

export interface SyncStats {
  totalEvents: number;
  googleEvents: number;
  localEvents: number;
  dateRange: { from: string | null; to: string | null } | null;
  lastSync: string | null;
}

export interface EventFormData {
  title: string;
  description: string;
  start_date: string;
  start_time: string;
  end_date: string;
  end_time: string;
  all_day: boolean;
  location: string;
  meeting_type: MeetingType | "";
  meeting_link: string;
}

export const weekDays = ["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"];
export const hours = Array.from({ length: 24 }, (_, i) => i);

/**
 * Meeting type color definitions using raw hex values.
 * These are applied as inline CSS custom properties (--cal-ev-color)
 * to avoid Tailwind color class dependencies.
 */
export const meetingTypeColorValues: Record<string, string> = {
  team: "#3b82f6",
  sales: "#10b981",
  client: "#8b5cf6",
  onboarding: "#eab308",
  kickoff: "#f97316",
  implementation: "#06b6d4",
  standup: "#6366f1",
  planning: "#ec4899",
  review: "#14b8a6",
  one_on_one: "#f43f5e",
  retrospective: "#f59e0b",
  internal: "#64748b",
  external: "#059669",
  other: "#6b7280",
  default: "#3b82f6",
};

export function getEventColor(event: CalendarEvent): string {
  const type = event.meeting_type || "default";
  return meetingTypeColorValues[type] || meetingTypeColorValues.default;
}

/**
 * @deprecated Use getEventColor() + inline styles instead.
 * Kept for backward compat during migration.
 */
export const meetingTypeColors: Record<
  string,
  { bg: string; border: string; text: string }
> = {
  team: { bg: "bg-blue-100", border: "border-blue-300", text: "text-blue-800" },
  sales: { bg: "bg-green-100", border: "border-green-300", text: "text-green-800" },
  client: { bg: "bg-purple-100", border: "border-purple-300", text: "text-purple-800" },
  onboarding: { bg: "bg-yellow-100", border: "border-yellow-300", text: "text-yellow-800" },
  kickoff: { bg: "bg-orange-100", border: "border-orange-300", text: "text-orange-800" },
  implementation: { bg: "bg-cyan-100", border: "border-cyan-300", text: "text-cyan-800" },
  standup: { bg: "bg-indigo-100", border: "border-indigo-300", text: "text-indigo-800" },
  planning: { bg: "bg-pink-100", border: "border-pink-300", text: "text-pink-800" },
  review: { bg: "bg-teal-100", border: "border-teal-300", text: "text-teal-800" },
  one_on_one: { bg: "bg-rose-100", border: "border-rose-300", text: "text-rose-800" },
  retrospective: { bg: "bg-amber-100", border: "border-amber-300", text: "text-amber-800" },
  internal: { bg: "bg-slate-100", border: "border-slate-300", text: "text-slate-800" },
  external: { bg: "bg-emerald-100", border: "border-emerald-300", text: "text-emerald-800" },
  other: { bg: "bg-gray-100", border: "border-gray-300", text: "text-gray-800" },
  default: { bg: "bg-blue-50", border: "border-blue-200", text: "text-blue-700" },
};

export function getEventColors(event: CalendarEvent): {
  bg: string;
  border: string;
  text: string;
} {
  const type = event.meeting_type || "default";
  return meetingTypeColors[type] || meetingTypeColors.default;
}

export function isToday(date: Date): boolean {
  const today = new Date();
  return (
    date.getFullYear() === today.getFullYear() &&
    date.getMonth() === today.getMonth() &&
    date.getDate() === today.getDate()
  );
}

export function formatHour(hour: number): string {
  if (hour === 0) return "12 AM";
  if (hour === 12) return "12 PM";
  return hour > 12 ? `${hour - 12} PM` : `${hour} AM`;
}

/** Strip dangerous HTML, allow basic formatting tags only. */
export function sanitizeHtml(html: string): string {
  if (!html) return "";
  const temp = document.createElement("div");
  temp.innerHTML = html;
  temp
    .querySelectorAll("script, style, iframe, object, embed")
    .forEach((el) => el.remove());
  temp.querySelectorAll("*").forEach((el) => {
    Array.from(el.attributes).forEach((attr) => {
      if (
        attr.name.startsWith("on") ||
        (attr.name === "href" && attr.value.startsWith("javascript:"))
      ) {
        el.removeAttribute(attr.name);
      }
    });
  });
  return temp.innerHTML;
}

export function getEventsForDate(
  events: CalendarEvent[],
  date: Date,
): CalendarEvent[] {
  return events.filter((event) => {
    const eventDate = new Date(event.start_time);
    return (
      eventDate.getFullYear() === date.getFullYear() &&
      eventDate.getMonth() === date.getMonth() &&
      eventDate.getDate() === date.getDate()
    );
  });
}

export function getEventsForHour(
  events: CalendarEvent[],
  date: Date,
  hour: number,
): CalendarEvent[] {
  return events.filter((event) => {
    const eventStart = new Date(event.start_time);
    return (
      eventStart.getFullYear() === date.getFullYear() &&
      eventStart.getMonth() === date.getMonth() &&
      eventStart.getDate() === date.getDate() &&
      eventStart.getHours() === hour
    );
  });
}

export function buildDateRange(
  viewMode: ViewMode,
  currentDate: Date,
): { start: Date; end: Date } {
  const start = new Date(currentDate);
  const end = new Date(currentDate);

  if (viewMode === "day") {
    start.setHours(0, 0, 0, 0);
    end.setHours(23, 59, 59, 999);
  } else if (viewMode === "week") {
    start.setDate(start.getDate() - start.getDay());
    start.setHours(0, 0, 0, 0);
    end.setDate(start.getDate() + 6);
    end.setHours(23, 59, 59, 999);
  } else if (viewMode === "month") {
    start.setDate(1);
    start.setHours(0, 0, 0, 0);
    end.setMonth(end.getMonth() + 1, 0);
    end.setHours(23, 59, 59, 999);
  } else {
    // agenda — 30 days from today
    start.setHours(0, 0, 0, 0);
    end.setDate(end.getDate() + 30);
    end.setHours(23, 59, 59, 999);
  }

  return { start, end };
}

export function buildMonthData(
  currentDate: Date,
  dateRange: { start: Date; end: Date },
): Date[][] {
  const firstDayOfMonth = dateRange.start.getDay();
  const daysInMonth = new Date(dateRange.end).getDate();

  const weeks: Date[][] = [];
  let currentWeek: Date[] = [];

  for (let i = 0; i < firstDayOfMonth; i++) {
    const prevDate = new Date(dateRange.start);
    prevDate.setDate(prevDate.getDate() - (firstDayOfMonth - i));
    currentWeek.push(prevDate);
  }

  for (let day = 1; day <= daysInMonth; day++) {
    const date = new Date(
      currentDate.getFullYear(),
      currentDate.getMonth(),
      day,
    );
    currentWeek.push(date);
    if (currentWeek.length === 7) {
      weeks.push(currentWeek);
      currentWeek = [];
    }
  }

  if (currentWeek.length > 0) {
    const nextMonth = new Date(dateRange.end);
    nextMonth.setDate(nextMonth.getDate() + 1);
    while (currentWeek.length < 7) {
      currentWeek.push(new Date(nextMonth));
      nextMonth.setDate(nextMonth.getDate() + 1);
    }
    weeks.push(currentWeek);
  }

  return weeks;
}
