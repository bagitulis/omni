export function formatDuration(ms: number): string {
  if (!Number.isFinite(ms) || ms <= 0) {
    return "-";
  }

  if (ms < 1000) {
    return `${Math.round(ms)}ms`;
  }

  return `${(ms / 1000).toFixed(2)}s`;
}

export function formatTimestamp(date: string): string {
  if (!date) {
    return "-";
  }

  const parsed = new Date(date);
  if (Number.isNaN(parsed.getTime())) {
    return "-";
  }

  return parsed.toLocaleString();
}

export function formatStatus(status: string): {
  color: string;
  label: string;
} {
  switch (status) {
    case "completed":
      return { color: "success", label: "COMPLETED" };
    case "failed":
      return { color: "error", label: "FAILED" };
    case "cancelled":
      return { color: "warning", label: "CANCELLED" };
    default:
      return { color: "default", label: (status || "unknown").toUpperCase() };
  }
}
