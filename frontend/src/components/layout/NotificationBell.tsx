import { BellOutlined } from "@ant-design/icons";
import { Badge, Button, Popover } from "antd";
import { useState } from "react";
import { useNotifications } from "@/contexts/NotificationContext";
import { NotificationDropdown } from "./NotificationDropdown";
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
  const { unreadCount, counts } = useNotifications();
  const [isOpen, setIsOpen] = useState(false);
  // Counts are authoritative — the loaded list may be capped at 50, so the
  // bell would previously under-report errors/warnings.
  const criticalCount = counts.by_severity?.["50"] ?? 0;
  const highCount = counts.by_severity?.["40"] ?? 0;
  const hasErrors = criticalCount > 0 || highCount > 0;
  const labelParts = [`Notifications`, `${unreadCount} unread`];
  if (criticalCount > 0) labelParts.push(`${criticalCount} critical`);
  if (highCount > 0) labelParts.push(`${highCount} high`);

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
        <Badge
          count={unreadCount}
          dot={hasErrors}
          status={hasErrors ? "error" : "default"}
          size="small"
        >
          <Button
            type="text"
            icon={<BellOutlined />}
            aria-label={labelParts.join(", ")}
            className="notification-bell-button"
          />
        </Badge>
      </span>
    </Popover>
  );
}
