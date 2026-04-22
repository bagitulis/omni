import { theme } from "antd";

/**
 * Parsed notification message.
 * If the raw message is JSON (e.g. from a background job result),
 * we extract a human-readable summary and optional numeric stats.
 */
export interface ParsedMessage {
  summary: string;
  stats?: { total?: number; processed?: number; failed?: number };
}

export function formatRelativeTime(dateStr: string): string {
  const date = new Date(dateStr);
  const diff = Date.now() - date.getTime();
  const mins = Math.floor(diff / 60_000);
  if (mins < 1) return "just now";
  if (mins < 60) return `${mins}m ago`;
  const hours = Math.floor(mins / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.floor(hours / 24);
  return `${days}d ago`;
}

/**
 * Parses a notification message that may be a raw JSON string
 * from a background job handler (e.g. escrow sync result).
 */
export function parseNotificationMessage(raw: string): ParsedMessage {
  if (!raw) return { summary: "" };

  try {
    const parsed = JSON.parse(raw);

    if (typeof parsed !== "object" || parsed === null || Array.isArray(parsed)) {
      return { summary: raw };
    }

    const summary =
      typeof parsed.message === "string" && parsed.message.trim()
        ? parsed.message.trim()
        : raw;

    const stats: ParsedMessage["stats"] = {};
    if (typeof parsed.total_orders === "number") stats.total = parsed.total_orders;
    if (typeof parsed.processed_orders === "number") stats.processed = parsed.processed_orders;
    if (typeof parsed.failed_orders === "number") stats.failed = parsed.failed_orders;

    const hasStats = Object.keys(stats).length > 0;
    return { summary, stats: hasStats ? stats : undefined };
  } catch {
    return { summary: raw };
  }
}

/** Renders numeric stats as small chips if available */
export function NotificationStats({
  stats,
  failed,
}: {
  stats: NonNullable<ParsedMessage["stats"]>;
  failed: boolean;
}) {
  const { token } = theme.useToken();
  const chipStyle = (isError?: boolean): React.CSSProperties => ({
    display: "inline-flex",
    alignItems: "center",
    gap: 3,
    padding: "1px 6px",
    borderRadius: 10,
    fontSize: 11,
    fontWeight: 600,
    background: isError ? token.colorErrorBg : token.colorSuccessBg,
    color: isError ? token.colorError : token.colorSuccess,
    border: `1px solid ${isError ? token.colorErrorBorder : token.colorSuccessBorder}`,
  });

  return (
    <div style={{ display: "flex", gap: 4, marginTop: 4, flexWrap: "wrap" }}>
      {stats.total !== undefined && (
        <span style={chipStyle()}>Total: {stats.total}</span>
      )}
      {stats.processed !== undefined && (
        <span style={chipStyle()}>Done: {stats.processed}</span>
      )}
      {stats.failed !== undefined && stats.failed > 0 && (
        <span style={chipStyle(true)}>Failed: {stats.failed}</span>
      )}
      {failed && stats.failed === 0 && (
        <span style={chipStyle()}>0 Failed</span>
      )}
    </div>
  );
}
