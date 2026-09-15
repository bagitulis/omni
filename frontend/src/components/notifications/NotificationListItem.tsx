import { Badge, Checkbox, Tag, Typography } from "antd";
import type { Notification } from "@/api/notifications";
import { formatRelativeTime, getNotificationTypeConfig } from "@/components/layout/NotificationHelpers";
import { severityColor, severityLabel } from "./severity";

const { Text } = Typography;

interface NotificationListItemProps {
  item: Notification;
  active: boolean;
  selected: boolean;
  onSelectToggle: (id: number) => void;
  onOpen: (item: Notification) => void;
}

/**
 * Single row in the notifications list. Owns the left severity rail,
 * checkbox, icon, title, dedup badge, and timestamp.
 */
export function NotificationListItem({ item, active, selected, onSelectToggle, onOpen }: NotificationListItemProps) {
  const typeConfig = getNotificationTypeConfig(item.type);
  return (
    <div
      className={`notification-list-item${active ? " notification-list-item--active" : ""}${!item.read ? " notification-list-item--unread" : ""}`}
      style={{ borderLeftColor: severityColor(item.severity) }}
    >
      <div className="notification-list-item__select" onClick={(e) => e.stopPropagation()}>
        <Checkbox checked={selected} onChange={() => onSelectToggle(item.id)} aria-label={`Select ${item.title}`} />
      </div>
      <button
        type="button"
        className="notification-list-item__content"
        onClick={() => onOpen(item)}
        aria-label={`Open notification: ${item.title}`}
      >
        <span className="notification-list-item__icon" data-notification-type={typeConfig.type}>
          {typeConfig.icon}
        </span>
        <div className="notification-list-item__main">
          <div className="notification-list-item__title-row">
            <Text strong={!item.read} ellipsis className="notification-list-item__title">
              {item.title}
            </Text>
            {item.dedup_count > 1 && (
              <Tag color="default" className="notification-list-item__dedup">×{item.dedup_count}</Tag>
            )}
          </div>
          <div className="notification-list-item__meta">
            <Text type="secondary" className="notification-list-item__severity">
              {severityLabel(item.severity)}
            </Text>
            <span aria-hidden="true"> · </span>
            <Text type="secondary" className="notification-list-item__category">
              {item.category}
            </Text>
            <span aria-hidden="true"> · </span>
            <Text type="secondary" className="notification-list-item__time">
              {formatRelativeTime(item.created_at)}
            </Text>
          </div>
        </div>
        {!item.read && <Badge status="processing" className="notification-list-item__unread-dot" />}
      </button>
    </div>
  );
}
