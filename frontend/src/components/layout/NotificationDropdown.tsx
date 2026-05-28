import { useState, useMemo } from "react";
import { Button, Empty, Segmented, Skeleton, Typography, theme, Tooltip } from "antd";
import { useNavigate } from "react-router-dom";
import {
  CheckOutlined,
  DeleteOutlined,
} from "@ant-design/icons";
import { useNotifications } from "@/contexts/NotificationContext";
import type { Notification } from "@/api/notifications";
import {
  formatRelativeTime,
  parseNotificationMessage,
  NotificationStats,
  getNotificationTypeConfig,
} from "./NotificationHelpers";
import "./notifications.css";

const { Text } = Typography;

type TabKey = "all" | "unread";

interface NotificationDropdownProps {
  onClose?: () => void;
}

/**
 * Notification dropdown panel.
 * Uses real-time database-backed NotificationContext.
 */
export function NotificationDropdown({ onClose }: NotificationDropdownProps) {
  const { token } = theme.useToken();
  const navigate = useNavigate();
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
  const unreadCount = notifications.filter((n) => !n.read).length;

  const handleItemClick = (item: Notification) => {
    if (!item.read) markAsRead(item.id);
    if (item.action_url) {
      onClose?.();
      window.location.href = item.action_url;
    } else {
      // Navigate to notifications page with this notification expanded
      onClose?.();
      navigate(`/notifications?expand=${item.id}`);
    }
  };

  return (
      <div className="notification-dropdown" data-testid="notification-dropdown">
        {/* Header */}
        <div className="notification-dropdown__header">
          <Text strong className="notification-dropdown__title">
            Notifications
          </Text>
          {unreadExists && (
            <Button
              type="link"
              size="small"
              icon={<CheckOutlined />}
              onClick={markAllAsRead}
              className="notification-dropdown__action"
            >
              Mark all read
            </Button>
          )}
        </div>

        {/* Tabs */}
        <div className="notification-dropdown__tabs">
          <Segmented
            block
            size="small"
            value={tab}
            onChange={(v) => setTab(v as TabKey)}
            options={[
              { label: `All (${notifications.length})`, value: "all" },
              { label: `Unread (${unreadCount})`, value: "unread" },
            ]}
          />
        </div>

        {/* List */}
        <div className="notification-dropdown__list">
          {loading && visible.length === 0 ? (
            <div className="notification-dropdown__state">
              <Skeleton active avatar paragraph={{ rows: 3 }} />
            </div>
          ) : visible.length === 0 ? (
            <Empty
              image={Empty.PRESENTED_IMAGE_SIMPLE}
              description={
                tab === "unread"
                  ? "No unread notifications"
                  : "No notifications yet"
              }
              className="notification-dropdown__state"
            />
          ) : (
            visible.map((item) => {
              const cfg = getNotificationTypeConfig(item.type, token);
              const parsed = parseNotificationMessage(item.message);

              return (
                <button
                  key={item.id}
                  type="button"
                  className={`notification-item ${item.read ? "notification-item--read" : "notification-item--unread"}`}
                  onClick={() => handleItemClick(item)}
                >
                  <div className="notification-item__icon" style={{ color: cfg.color }}>
                    {cfg.icon}
                  </div>

                  <div className="notification-item__body">
                    <div className="notification-item__header">
                      <Text strong={!item.read} ellipsis className="notification-item__title">
                        {item.title}
                      </Text>
                      <Tooltip title="Delete">
                        <Button
                          type="text"
                          size="small"
                          icon={<DeleteOutlined />}
                          onClick={(e) => { e.stopPropagation(); deleteNotification(item.id); }}
                          className="notification-item__delete"
                        />
                      </Tooltip>
                    </div>

                    {parsed.summary && (
                      <Text type="secondary" className="notification-item__message">
                        {parsed.summary}
                      </Text>
                    )}

                    {parsed.stats && (
                      <NotificationStats stats={parsed.stats} failed={item.type === "error"} platforms={parsed.platforms} />
                    )}

                    <Text type="secondary" className="notification-item__time">
                      {formatRelativeTime(item.created_at)}
                    </Text>
                  </div>

                  {!item.read && (
                    <div className="notification-item__unread-dot" />
                  )}
                </button>
              );
            })
          )}
        </div>

        {/* Footer */}
        <div className="notification-dropdown__footer">
          <Text type="secondary" className="notification-item__message">
            Auto-delete: 30 days
          </Text>
          <Button type="link" size="small" className="notification-dropdown__action" onClick={() => { onClose?.(); navigate("/notifications"); }}>
            View All
          </Button>
        </div>
      </div>
  );
}
