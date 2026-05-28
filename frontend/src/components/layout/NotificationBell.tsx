import { useState } from "react";
import { Badge, Button, Popover, theme } from "antd";
import { BellOutlined } from "@ant-design/icons";
import { NotificationDropdown } from "./NotificationDropdown";
import { useNotifications } from "@/contexts/NotificationContext";
import "./notifications.css";

/**
 * Bell icon with unread badge. Clicking opens a popover with
 * the NotificationDropdown panel.
 *
 * NOTE: We deliberately avoid a manual click-outside handler here.
 * Ant Design Popover renders its content via a React Portal to <body>,
 * so any containerRef.contains() check would incorrectly treat clicks
 * inside the dropdown as "outside" and close it prematurely.
 * AntD's built-in Popover handles click-outside correctly on its own.
 */
export function NotificationBell() {
  const { token } = theme.useToken();
  const { unreadCount, notifications } = useNotifications();
  const [isOpen, setIsOpen] = useState(false);
  const errorCount = notifications.filter((n) => !n.read && n.type === "error").length;
  const warningCount = notifications.filter((n) => !n.read && n.type === "warning").length;
  const severityColor = errorCount > 0 ? token.colorError : warningCount > 0 ? token.colorWarning : token.colorInfo;
  const labelParts = [`Notifications`, `${unreadCount} unread`];
  if (errorCount > 0) labelParts.push(`${errorCount} error`);
  if (warningCount > 0) labelParts.push(`${warningCount} warning`);

  return (
    <Popover
      content={<NotificationDropdown onClose={() => setIsOpen(false)} />}
      trigger="click"
      open={isOpen}
      onOpenChange={(open) => setIsOpen(open)}
      placement="bottomRight"
      arrow={false}
      overlayClassName="notification-popover"
    >
      <span className="notification-bell-wrapper">
        <Badge count={unreadCount} size="small" offset={[-4, 4]} color={severityColor}>
          <Button
            type="text"
            icon={<BellOutlined />}
            aria-label={labelParts.join(", ")}
            className="notification-bell-button"
          />
        </Badge>
        {(errorCount > 0 || warningCount > 0) && (
          <span
            className="notification-severity-dot"
            style={{ backgroundColor: severityColor }}
            aria-hidden="true"
          />
        )}
      </span>
    </Popover>
  );
}
