import { useState, useMemo } from "react";
import { Button, Empty, Segmented, Typography, theme, Tooltip } from "antd";
import {
  CheckCircleOutlined,
  CloseCircleOutlined,
  ExclamationCircleOutlined,
  InfoCircleOutlined,
  CheckOutlined,
  DeleteOutlined,
} from "@ant-design/icons";
import { useNotifications } from "@/contexts/NotificationContext";
import { Notification } from "@/api/notifications";

const { Text } = Typography;

type TabKey = "all" | "unread";

interface NotificationDropdownProps {
  onClose?: () => void;
}

/** Color & icon map for notification types */
function useTypeConfig() {
  const { token } = theme.useToken();
  return {
    success: { color: token.colorSuccess, icon: <CheckCircleOutlined /> },
    error: { color: token.colorError, icon: <CloseCircleOutlined /> },
    warning: { color: token.colorWarning, icon: <ExclamationCircleOutlined /> },
    info: { color: token.colorPrimary, icon: <InfoCircleOutlined /> },
  };
}

function formatRelativeTime(dateStr: string): string {
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
 * Parsed notification message.
 * If the raw message is JSON (e.g. from a background job result),
 * we extract a human-readable summary and optional numeric stats.
 */
interface ParsedMessage {
  summary: string;
  stats?: { total?: number; processed?: number; failed?: number };
}

/**
 * Parses a notification message that may be a raw JSON string
 * from a background job handler (e.g. escrow sync result).
 *
 * Expected JSON shape (best effort, not enforced):
 * { message: string, total_orders?: number, processed_orders?: number, failed_orders?: number }
 */
function parseNotificationMessage(raw: string): ParsedMessage {
  if (!raw) return { summary: "" };

  try {
    const parsed = JSON.parse(raw);

    // Must be a plain object to proceed
    if (typeof parsed !== "object" || parsed === null || Array.isArray(parsed)) {
      return { summary: raw };
    }

    // Use the human-readable "message" field if present
    const summary =
      typeof parsed.message === "string" && parsed.message.trim()
        ? parsed.message.trim()
        : raw;

    // Extract optional numeric stats
    const stats: ParsedMessage["stats"] = {};
    if (typeof parsed.total_orders === "number") stats.total = parsed.total_orders;
    if (typeof parsed.processed_orders === "number") stats.processed = parsed.processed_orders;
    if (typeof parsed.failed_orders === "number") stats.failed = parsed.failed_orders;

    const hasStats = Object.keys(stats).length > 0;
    return { summary, stats: hasStats ? stats : undefined };
  } catch {
    // Not valid JSON — display as plain text
    return { summary: raw };
  }
}

/** Renders numeric stats as small chips if available */
function NotificationStats({
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

/**
 * Notification dropdown panel.
 * Uses real-time database-backed NotificationContext.
 */
export function NotificationDropdown({ onClose }: NotificationDropdownProps) {
  const { token } = theme.useToken();
  const typeConfig = useTypeConfig();
  const [tab, setTab] = useState<TabKey>("all");

  const {
    notifications,
    markAsRead,
    markAllAsRead,
    deleteNotification,
    loading,
  } = useNotifications();

  const visible = useMemo(() => {
    return tab === "unread"
      ? notifications.filter((n) => !n.read)
      : notifications;
  }, [tab, notifications]);

  const unreadExists = notifications.some((n) => !n.read);

  const handleItemClick = (item: Notification) => {
    if (!item.read) markAsRead(item.id);
    if (item.action_url) {
      onClose?.();
      window.location.href = item.action_url;
    }
    // No close if no action_url — user wants to read the notification in place
  };

  return (
    <>
      {/* Inline hover style — no external CSS dependency */}
      <style>{`
        .notif-item:hover {
          background: ${token.colorFillTertiary} !important;
        }
      `}</style>

      <div
        style={{
          maxHeight: 500,
          display: "flex",
          flexDirection: "column",
          background: token.colorBgElevated,
          borderRadius: token.borderRadiusLG,
          boxShadow: token.boxShadowSecondary,
        }}
      >
        {/* Header */}
        <div
          style={{
            padding: "12px 16px",
            display: "flex",
            alignItems: "center",
            justifyContent: "space-between",
            borderBottom: `1px solid ${token.colorBorderSecondary}`,
          }}
        >
          <Text strong style={{ fontSize: 16 }}>
            Notifications
          </Text>
          {unreadExists && (
            <Button
              type="link"
              size="small"
              icon={<CheckOutlined />}
              onClick={markAllAsRead}
              style={{ fontSize: 12, padding: 0 }}
            >
              Mark all read
            </Button>
          )}
        </div>

        {/* Tabs */}
        <div style={{ padding: "8px 16px" }}>
          <Segmented
            block
            size="small"
            value={tab}
            onChange={(v) => setTab(v as TabKey)}
            options={[
              { label: "All", value: "all" },
              { label: "Unread", value: "unread" },
            ]}
          />
        </div>

        {/* List */}
        <div style={{ overflowY: "auto", flex: 1, minHeight: 100, maxHeight: 400 }}>
          {visible.length === 0 ? (
            <Empty
              image={Empty.PRESENTED_IMAGE_SIMPLE}
              description={
                loading
                  ? "Loading..."
                  : tab === "unread"
                  ? "No unread notifications"
                  : "No notifications yet"
              }
              style={{ margin: "32px 0" }}
            />
          ) : (
            visible.map((item) => {
              const cfg =
                typeConfig[item.type as keyof typeof typeConfig] ||
                typeConfig.info;
              const parsed = parseNotificationMessage(item.message);

              return (
                <div
                  key={item.id}
                  className="notif-item"
                  style={{
                    padding: "12px 16px",
                    cursor: "pointer",
                    display: "flex",
                    gap: 12,
                    alignItems: "flex-start",
                    background: item.read
                      ? "transparent"
                      : token.colorPrimaryBg + "44",
                    borderBottom: `1px solid ${token.colorBorderSecondary}`,
                    transition: "background 0.15s ease",
                    opacity: item.read ? 0.75 : 1,
                    position: "relative",
                  }}
                  onClick={() => handleItemClick(item)}
                >
                  {/* Type icon */}
                  <div
                    style={{
                      fontSize: 20,
                      color: cfg.color,
                      flexShrink: 0,
                      marginTop: 2,
                    }}
                  >
                    {cfg.icon}
                  </div>

                  {/* Content */}
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <div
                      style={{
                        display: "flex",
                        justifyContent: "space-between",
                        alignItems: "flex-start",
                        gap: 4,
                      }}
                    >
                      <Text
                        strong={!item.read}
                        style={{
                          fontSize: 14,
                          display: "block",
                          lineHeight: 1.4,
                          color: token.colorText,
                        }}
                      >
                        {item.title}
                      </Text>

                      {/* Delete button — stopPropagation prevents triggering handleItemClick */}
                      <Tooltip title="Delete">
                        <Button
                          type="text"
                          size="small"
                          icon={<DeleteOutlined />}
                          onClick={(e) => {
                            e.stopPropagation();
                            deleteNotification(item.id);
                          }}
                          style={{
                            opacity: 0.45,
                            flexShrink: 0,
                            marginLeft: 4,
                          }}
                        />
                      </Tooltip>
                    </div>

                    {/* Message — rendered as human-readable, not raw JSON */}
                    {parsed.summary && (
                      <Text
                        type="secondary"
                        style={{
                          fontSize: 12,
                          display: "block",
                          lineHeight: 1.5,
                          marginTop: 3,
                          wordBreak: "break-word",
                        }}
                      >
                        {parsed.summary}
                      </Text>
                    )}

                    {/* Stats chips for job results */}
                    {parsed.stats && (
                      <NotificationStats
                        stats={parsed.stats}
                        failed={item.type === "error"}
                      />
                    )}

                    <Text
                      type="secondary"
                      style={{ fontSize: 11, marginTop: 5, display: "block" }}
                    >
                      {formatRelativeTime(item.created_at)}
                    </Text>
                  </div>

                  {/* Unread dot */}
                  {!item.read && (
                    <div
                      style={{
                        width: 8,
                        height: 8,
                        borderRadius: "50%",
                        background: token.colorPrimary,
                        flexShrink: 0,
                        marginTop: 6,
                      }}
                    />
                  )}
                </div>
              );
            })
          )}
        </div>

        {/* Footer */}
        <div
          style={{
            padding: "10px 16px",
            borderTop: `1px solid ${token.colorBorderSecondary}`,
            textAlign: "center",
            background: token.colorFillAlter,
          }}
        >
          <Text type="secondary" style={{ fontSize: 12 }}>
            Showing recent 50 notifications
          </Text>
        </div>
      </div>
    </>
  );
}
