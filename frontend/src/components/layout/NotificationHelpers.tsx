/* eslint-disable react-refresh/only-export-components */
import { CheckCircleOutlined, CloseCircleOutlined, ExclamationCircleOutlined, InfoCircleOutlined } from "@ant-design/icons";
import { Descriptions, Tag, Typography } from "antd";
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

export type NotificationType = "success" | "error" | "warning" | "info";

const notificationTypes: readonly NotificationType[] = ["success", "error", "warning", "info"];

export function normalizeNotificationType(type?: string | null): NotificationType {
  if (!type) return "info";
  return notificationTypes.includes(type as NotificationType)
    ? (type as NotificationType)
    : "info";
}

export function getNotificationTypeConfig(type?: string | null) {
  const configs = {
    success: { type: "success", icon: <CheckCircleOutlined />, label: "Success" },
    error: { type: "error", icon: <CloseCircleOutlined />, label: "Error" },
    warning: { type: "warning", icon: <ExclamationCircleOutlined />, label: "Warning" },
    info: { type: "info", icon: <InfoCircleOutlined />, label: "Info" },
  };

  return configs[normalizeNotificationType(type)];
}

export function parseNotificationMetadata(metadata?: string | null): Record<string, string> {
  if (!metadata?.trim()) return {};

  try {
    const parsed: unknown = JSON.parse(metadata);
    if (typeof parsed !== "object" || parsed === null || Array.isArray(parsed)) return {};

    return Object.entries(parsed).reduce<Record<string, string>>((acc, [key, value]) => {
      if (value === null || value === undefined || value === "") return acc;
      acc[key] = typeof value === "object" ? JSON.stringify(value) : String(value);
      return acc;
    }, {});
  } catch {
    return {};
  }
}

export function NotificationMetadata({ metadata }: { metadata?: string | null }) {
  const details = parseNotificationMetadata(metadata);
  const entries = Object.entries(details);
  if (entries.length === 0) return null;

  return (
    <Descriptions size="small" column={1} className="notification-metadata">
      {entries.map(([key, value]) => (
        <Descriptions.Item key={key} label={formatMetadataKey(key)}>
          <Text className="notification-metadata__value" ellipsis={{ tooltip: value }}>
            {value}
          </Text>
        </Descriptions.Item>
      ))}
    </Descriptions>
  );
}

function formatMetadataKey(key: string): string {
  return key.replace(/_/g, " ").replace(/\b\w/g, (char) => char.toUpperCase());
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
  return (
    <>
      <div className="notification-helper-stats">
        {stats.total !== undefined && (
          <Tag color="success" className="notification-helper-stat-tag">Total: {stats.total}</Tag>
        )}
        {stats.processed !== undefined && (
          <Tag color="success" className="notification-helper-stat-tag">Done: {stats.processed}</Tag>
        )}
        {stats.failed !== undefined && stats.failed > 0 && (
          <Tag color="error" className="notification-helper-stat-tag">Failed: {stats.failed}</Tag>
        )}
        {failed && stats.failed === 0 && (
          <Tag color="success" className="notification-helper-stat-tag">0 Failed</Tag>
        )}
      </div>
      {platforms && <PlatformBreakdown platforms={platforms} />}
    </>
  );
}
