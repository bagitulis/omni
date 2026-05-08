import { useState, useMemo } from "react";
import { Button, Empty, Segmented, Typography, theme, Tooltip } from "antd";
import { useNavigate } from "react-router-dom";
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
import {
  formatRelativeTime,
  parseNotificationMessage,
  NotificationStats,
} from "./NotificationHelpers";

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

/**
 * Notification dropdown panel.
 * Uses real-time database-backed NotificationContext.
 */
export function NotificationDropdown({ onClose }: NotificationDropdownProps) {
  const { token } = theme.useToken();
  const navigate = useNavigate();
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
    } else {
      // Navigate to notifications page for full detail view
      onClose?.();
      navigate("/notifications");
    }
  };

  return (
    <>
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
                    background: item.read ? "transparent" : token.colorPrimaryBg + "44",
                    borderBottom: `1px solid ${token.colorBorderSecondary}`,
                    transition: "background 0.15s ease",
                    opacity: item.read ? 0.75 : 1,
                    position: "relative",
                  }}
                  onClick={() => handleItemClick(item)}
                >
                  <div style={{ fontSize: 20, color: cfg.color, flexShrink: 0, marginTop: 2 }}>
                    {cfg.icon}
                  </div>

                  <div style={{ flex: 1, minWidth: 0 }}>
                    <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", gap: 4 }}>
                      <Text strong={!item.read} style={{ fontSize: 14, display: "block", lineHeight: 1.4, color: token.colorText }}>
                        {item.title}
                      </Text>
                      <Tooltip title="Delete">
                        <Button
                          type="text"
                          size="small"
                          icon={<DeleteOutlined />}
                          onClick={(e) => { e.stopPropagation(); deleteNotification(item.id); }}
                          style={{ opacity: 0.45, flexShrink: 0, marginLeft: 4 }}
                        />
                      </Tooltip>
                    </div>

                    {parsed.summary && (
                      <Text type="secondary" style={{ fontSize: 12, display: "block", lineHeight: 1.5, marginTop: 3, wordBreak: "break-word" }}>
                        {parsed.summary}
                      </Text>
                    )}

                    {parsed.stats && (
                      <NotificationStats stats={parsed.stats} failed={item.type === "error"} />
                    )}

                    <Text type="secondary" style={{ fontSize: 11, marginTop: 5, display: "block" }}>
                      {formatRelativeTime(item.created_at)}
                    </Text>
                  </div>

                  {!item.read && (
                    <div style={{ width: 8, height: 8, borderRadius: "50%", background: token.colorPrimary, flexShrink: 0, marginTop: 6 }} />
                  )}
                </div>
              );
            })
          )}
        </div>

        {/* Footer */}
        <div style={{ padding: "10px 16px", borderTop: `1px solid ${token.colorBorderSecondary}`, display: "flex", justifyContent: "space-between", alignItems: "center", background: token.colorFillAlter }}>
          <Text type="secondary" style={{ fontSize: 12 }}>
            Auto-delete: 30 days
          </Text>
          <Button type="link" size="small" style={{ fontSize: 12, padding: 0 }} onClick={() => { onClose?.(); navigate("/notifications"); }}>
            View All
          </Button>
        </div>
      </div>
    </>
  );
}
