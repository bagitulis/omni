import { useState, useMemo } from "react";
import { Button, Empty, Segmented, Typography, theme } from "antd";
import {
  CheckCircleOutlined,
  CloseCircleOutlined,
  ExclamationCircleOutlined,
  InfoCircleOutlined,
  CheckOutlined,
  DeleteOutlined,
} from "@ant-design/icons";
import {
  useNotificationStore,
  type NotificationItem,
  type NotificationType,
} from "@/stores/notificationStore";

const { Text } = Typography;

type TabKey = "all" | "unread";

/** Color & icon map for notification types */
const TYPE_CONFIG: Record<
  NotificationType,
  { color: string; icon: React.ReactNode }
> = {
  success: { color: "#52c41a", icon: <CheckCircleOutlined /> },
  error: { color: "#ff4d4f", icon: <CloseCircleOutlined /> },
  warning: { color: "#faad14", icon: <ExclamationCircleOutlined /> },
  info: { color: "#1677ff", icon: <InfoCircleOutlined /> },
};

function formatRelativeTime(timestamp: number): string {
  const diff = Date.now() - timestamp;
  const mins = Math.floor(diff / 60_000);
  if (mins < 1) return "just now";
  if (mins < 60) return `${mins}m ago`;
  const hours = Math.floor(mins / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.floor(hours / 24);
  return `${days}d ago`;
}

/**
 * Notification dropdown panel shown inside the bell popover.
 * Supports All / Unread tabs, mark-all-read, and per-item read.
 */
export function NotificationDropdown() {
  const { token } = theme.useToken();
  const [tab, setTab] = useState<TabKey>("all");

  const notifications = useNotificationStore((s) => s.notifications);
  const markAsRead = useNotificationStore((s) => s.markAsRead);
  const markAllAsRead = useNotificationStore((s) => s.markAllAsRead);
  const clearAll = useNotificationStore((s) => s.clearAll);
  const closeDropdown = useNotificationStore((s) => s.closeDropdown);

  const visible = useMemo(() => {
    const list =
      tab === "unread" ? notifications.filter((n) => !n.read) : notifications;
    return list.slice(0, 30);
  }, [tab, notifications]);

  const unreadExists = notifications.some((n) => !n.read);

  const handleItemClick = (item: NotificationItem) => {
    if (!item.read) markAsRead(item.id);
    if (item.actionUrl) {
      closeDropdown();
      window.location.href = item.actionUrl;
    }
  };

  return (
    <div
      style={{
        maxHeight: 440,
        display: "flex",
        flexDirection: "column",
      }}
    >
      {/* Header */}
      <div
        style={{
          padding: "12px 16px 8px",
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
          borderBottom: `1px solid ${token.colorBorderSecondary}`,
        }}
      >
        <Text strong style={{ fontSize: 15 }}>
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
      <div style={{ padding: "8px 16px 4px" }}>
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
      <div style={{ overflowY: "auto", flex: 1, padding: "4px 0" }}>
        {visible.length === 0 ? (
          <Empty
            image={Empty.PRESENTED_IMAGE_SIMPLE}
            description={
              tab === "unread"
                ? "No unread notifications"
                : "No notifications yet"
            }
            style={{ margin: "24px 0" }}
          />
        ) : (
          visible.map((item) => {
            const cfg = TYPE_CONFIG[item.type];
            return (
              <div
                key={item.id}
                onClick={() => handleItemClick(item)}
                role="button"
                tabIndex={0}
                onKeyDown={(e) => {
                  if (e.key === "Enter") handleItemClick(item);
                }}
                style={{
                  padding: "10px 16px",
                  cursor: "pointer",
                  display: "flex",
                  gap: 10,
                  alignItems: "flex-start",
                  background: item.read
                    ? "transparent"
                    : token.colorPrimaryBg,
                  borderBottom: `1px solid ${token.colorBorderSecondary}`,
                  transition: "background 0.15s",
                }}
              >
                {/* Unread dot */}
                <div style={{ width: 8, paddingTop: 6, flexShrink: 0 }}>
                  {!item.read && (
                    <div
                      style={{
                        width: 8,
                        height: 8,
                        borderRadius: "50%",
                        background: token.colorPrimary,
                      }}
                    />
                  )}
                </div>

                {/* Type icon */}
                <div
                  style={{
                    fontSize: 18,
                    color: cfg.color,
                    flexShrink: 0,
                    lineHeight: 1,
                    paddingTop: 2,
                  }}
                >
                  {cfg.icon}
                </div>

                {/* Content */}
                <div style={{ flex: 1, minWidth: 0 }}>
                  <Text
                    strong
                    style={{
                      fontSize: 13,
                      display: "block",
                      lineHeight: 1.3,
                    }}
                  >
                    {item.title}
                  </Text>
                  {item.message && (
                    <Text
                      type="secondary"
                      style={{
                        fontSize: 12,
                        display: "block",
                        lineHeight: 1.4,
                        marginTop: 2,
                      }}
                      ellipsis={{ tooltip: item.message }}
                    >
                      {item.message}
                    </Text>
                  )}
                  <Text
                    type="secondary"
                    style={{ fontSize: 11, marginTop: 4, display: "block" }}
                  >
                    {formatRelativeTime(item.timestamp)}
                  </Text>
                </div>
              </div>
            );
          })
        )}
      </div>

      {/* Footer */}
      {notifications.length > 0 && (
        <div
          style={{
            padding: "8px 16px",
            borderTop: `1px solid ${token.colorBorderSecondary}`,
            display: "flex",
            justifyContent: "center",
          }}
        >
          <Button
            type="text"
            size="small"
            icon={<DeleteOutlined />}
            danger
            onClick={clearAll}
            style={{ fontSize: 12 }}
          >
            Clear all
          </Button>
        </div>
      )}
    </div>
  );
}
