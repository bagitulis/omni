export function truncateId(id: string): string {
  if (!id) return "-";
  return id.length > 12 ? id.substring(0, 12) + "..." : id;
}

export function truncateError(error: string): string {
  if (!error) return "-";
  return error.length > 50 ? error.substring(0, 50) + "..." : error;
}

export function getStatusEmoji(status: string): string {
  switch (status) {
    case "completed":
      return "✅";
    case "failed":
      return "❌";
    case "pending":
      return "⏳";
    default:
      return "❓";
  }
}

export function formatDuration(ms?: number): string {
  if (!ms) return "-";
  if (ms < 1000) return `${ms}ms`;
  const seconds = Math.floor(ms / 1000);
  if (seconds < 60) return `${seconds}s`;
  const minutes = Math.floor(seconds / 60);
  const remainingSeconds = seconds % 60;
  return `${minutes}m ${remainingSeconds}s`;
}

export function formatDateTime(date: Date | string | undefined): string {
  if (!date) return "-";

  const d = typeof date === "string" ? new Date(date) : date;

  return new Intl.DateTimeFormat("en-US", {
    timeZone: "Asia/Jakarta",
    month: "short",
    day: "numeric",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
    hour12: true,
  }).format(d);
}
