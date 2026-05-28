import {
  CheckOutlined,
  DeleteOutlined,
} from "@ant-design/icons";
import { Button, Empty, Segmented, Skeleton, Tooltip, Typography } from "antd";
import { useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import type { Notification } from "@/api/notifications";
import { useNotifications } from "@/contexts/NotificationContext";
import {
  formatRelativeTime,
  getNotificationTypeConfig,
  NotificationMetadata,
  NotificationStats,
  parseNotificationMessage,
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
      <section className="notification-dropdown" data-testid="notification-dropdown" aria-label="Notifications">
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
          {loading ? (
            <div className="notification-dropdown__state">
              <Skeleton active avatar paragraph={{ rows: 2 }} />
            </div>
          ) : visible.length === 0 ? (
            <Empty
              image={Empty.PRESENTED_IMAGE_SIMPLE}
              description="No notifications"
              className="notification-dropdown__state"
            />
          ) : (
            visible.map((item) => {
              const cfg = getNotificationTypeConfig(item.type);
              const parsed = parseNotificationMessage(item.message);

              return (
                <div
                  key={item.id}
                  className={`notification-item ${item.read ? "notification-item--read" : "notification-item--unread"}`}
                >
                  <button type="button" className="notification-item__content" onClick={() => handleItemClick(item)}>
                    <div className="notification-item__icon" data-notification-type={cfg.type}>
                      {cfg.icon}
                    </div>

                    <div className="notification-item__body">
                      <div className="notification-item__header">
                        <Text strong={!item.read} ellipsis={{ tooltip: item.title }} className="notification-item__title">
                          {item.title}
                        </Text>
                      </div>

                      {parsed.summary && (
                        <Text type="secondary" className="notification-item__message">
                          {parsed.summary}
                        </Text>
                      )}

                      {parsed.stats && (
                        <NotificationStats stats={parsed.stats} failed={item.type === "error"} platforms={parsed.platforms} />
                      )}

                      <NotificationMetadata metadata={item.metadata} />

                      <Text type="secondary" className="notification-item__time">
                        {formatRelativeTime(item.created_at)}
                      </Text>
                    </div>
                  </button>

                  <Tooltip title="Delete">
                    <Button
                      type="text"
                      size="small"
                      icon={<DeleteOutlined />}
                      onClick={() => deleteNotification(item.id)}
                      className="notification-item__delete"
                    />
                  </Tooltip>

                  {!item.read && (
                    <div className="notification-item__unread-dot" />
                  )}
                </div>
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
      </section>
  );
}
