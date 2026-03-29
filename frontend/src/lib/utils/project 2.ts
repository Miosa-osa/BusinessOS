// Shared project utility functions

export function getStatusColor(status: string): string {
  switch (status) {
    case "active":
      return "bg-emerald-100 text-emerald-700 border-emerald-200";
    case "paused":
      return "bg-amber-100 text-amber-700 border-amber-200";
    case "completed":
      return "bg-blue-100 text-blue-700 border-blue-200";
    case "archived":
      return "bg-gray-100 text-gray-600 border-gray-200";
    default:
      return "bg-gray-100 text-gray-600 border-gray-200";
  }
}

export function getStatusIcon(status: string): string {
  switch (status) {
    case "active":
      return "M13 10V3L4 14h7v7l9-11h-7z";
    case "paused":
      return "M10 9v6m4-6v6m7-3a9 9 0 11-18 0 9 9 0 0118 0z";
    case "completed":
      return "M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z";
    default:
      return "M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4";
  }
}

export function getPriorityColor(priority: string): string {
  switch (priority) {
    case "critical":
      return "text-red-600 bg-red-50";
    case "high":
      return "text-orange-600 bg-orange-50";
    case "medium":
      return "text-yellow-600 bg-yellow-50";
    case "low":
      return "text-green-600 bg-green-50";
    default:
      return "text-gray-600 bg-gray-50";
  }
}

export function getTypeLabel(type: string): string {
  switch (type) {
    case "internal":
      return "Internal";
    case "client_work":
      return "Client Work";
    case "learning":
      return "Learning";
    default:
      return type;
  }
}

export function getTypeIcon(type: string): string {
  switch (type) {
    case "internal":
      return "🏢";
    case "client_work":
      return "👥";
    case "learning":
      return "📚";
    default:
      return "📁";
  }
}

export function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString(undefined, {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
}

export function formatTime(dateStr: string): string {
  return new Date(dateStr).toLocaleTimeString(undefined, {
    hour: "numeric",
    minute: "2-digit",
  });
}
