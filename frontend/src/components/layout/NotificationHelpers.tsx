/* eslint-disable react-refresh/only-export-components */
import { CheckCircleOutlined, CloseCircleOutlined, ExclamationCircleOutlined, InfoCircleOutlined } from "@ant-design/icons";
import { Tag, theme, Typography } from "antd";
import type { AliasToken } from "antd/es/theme/internal";
import type { BulkOperationMetadata } from "@/types/notificationMetadata";
import "./notifications.css";

const { Text } = Typography;

/**
 * Parsed notification message.
 * If the raw message is JSON (e.g. from a background job result),
 * we extract a human-readable summary and optional numeric stats.
 */
export interface ParsedMessage {
  summary: string;
  stats?: { total?: number; processed?: number; failed?: number };
  platforms?: Record<string, { succeeded: number; failed: number }>;
  failedItems?: Array<{ sku: string; platform: string; error: string; request_id?: string }>;
}

export function getNotificationTypeConfig(type: string, token: AliasToken) {
  const configs = {
    success: { color: token.colorSuccess, icon: <CheckCircleOutlined />, label: "Success" },
    error: { color: token.colorError, icon: <CloseCircleOutlined />, label: "Error" },
    warning: { color: token.colorWarning, icon: <ExclamationCircleOutlined />, label: "Warning" },
    info: { color: token.colorInfo, icon: <InfoCircleOutlined />, label: "Info" },
  };

  return configs[type as keyof typeof configs] ?? {
    color: token.colorTextSecondary,
    icon: <InfoCircleOutlined />,
    label: "Notification",
  };
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

    // Check if this is BulkOperationMetadata format
    if ("operation_type" in parsed) {
      const meta = parsed as BulkOperationMetadata;
      const summary = `${meta.operation_type.replace("_", " ")} completed`;
      const stats = {
        total: meta.total,
        processed: meta.succeeded,
        failed: meta.failed,
      };
      const platforms = meta.platforms || {};
      const failedItems = meta.failed_items || [];
      return { summary, stats, platforms, failedItems };
    }

    // Legacy format (backward compatibility)
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

/** Renders per-platform success/failure breakdown */
export function PlatformBreakdown({
  platforms,
}: {
  platforms: Record<string, { succeeded: number; failed: number }>;
}) {
  const entries = Object.entries(platforms);
  if (entries.length === 0) return null;

  return (
    <div className="notification-helper-platforms">
      {entries.map(([platform, stats]) => (
        <div key={platform} className="notification-helper-platform">
          <Text type="secondary" className="notification-helper-platform__name">
            {platform}:
          </Text>
          {stats.succeeded > 0 && (
            <Tag color="success" className="notification-helper-tag">
              {stats.succeeded} ok
            </Tag>
          )}
          {stats.failed > 0 && (
            <Tag color="error" className="notification-helper-tag">
              {stats.failed} failed
            </Tag>
          )}
        </div>
      ))}
    </div>
  );
}

/** Renders numeric stats as small chips if available */
export function NotificationStats({
  stats,
  failed,
  platforms,
}: {
  stats: NonNullable<ParsedMessage["stats"]>;
  failed: boolean;
  platforms?: Record<string, { succeeded: number; failed: number }>;
}) {
  const { token } = theme.useToken();
  const chipStyle = (isError?: boolean): React.CSSProperties => ({
    display: "inline-flex",
    alignItems: "center",
    gap: 4,
    padding: "1px 6px",
    borderRadius: 6,
    fontSize: 11,
    fontWeight: 600,
    background: isError ? token.colorErrorBg : token.colorSuccessBg,
    color: isError ? token.colorError : token.colorSuccess,
    border: `1px solid ${isError ? token.colorErrorBorder : token.colorSuccessBorder}`,
  });

  return (
    <>
      <div className="notification-helper-stats">
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
      {platforms && <PlatformBreakdown platforms={platforms} />}
    </>
  );
}
