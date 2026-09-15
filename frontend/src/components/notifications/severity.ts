import type { NotificationSeverity } from "@/api/notifications";

/**
 * Severity metadata used by the UI: color rail, label, and tag color.
 * The color values map to Ant Design theme tokens so light/dark themes
 * stay consistent.
 */
export const SEVERITY_META: Record<
  NotificationSeverity,
  { label: string; token: string; tag: "default" | "processing" | "success" | "warning" | "error"; rank: number }
> = {
  10: { label: "Info", token: "var(--color-primary)", tag: "processing", rank: 1 },
  20: { label: "Low", token: "var(--color-success)", tag: "success", rank: 2 },
  30: { label: "Medium", token: "var(--color-warning)", tag: "warning", rank: 3 },
  40: { label: "High", token: "var(--color-error)", tag: "error", rank: 4 },
  50: { label: "Critical", token: "#b8003c", tag: "error", rank: 5 },
};

export function severityColor(sev?: NotificationSeverity): string {
  return SEVERITY_META[sev ?? 10]?.token ?? "var(--color-primary)";
}

export function severityLabel(sev?: NotificationSeverity): string {
  return SEVERITY_META[sev ?? 10]?.label ?? "Info";
}
