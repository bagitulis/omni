import { useState } from "react";
import { Badge, Button, Popover } from "antd";
import { BellOutlined } from "@ant-design/icons";
import { NotificationDropdown } from "./NotificationDropdown";
import { useNotifications } from "@/contexts/NotificationContext";

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
  const { unreadCount } = useNotifications();
  const [isOpen, setIsOpen] = useState(false);

  return (
    <Popover
      content={<NotificationDropdown onClose={() => setIsOpen(false)} />}
      trigger="click"
      open={isOpen}
      onOpenChange={(open) => setIsOpen(open)}
      placement="bottomRight"
      arrow={false}
      overlayInnerStyle={{ padding: 0 }}
      overlayStyle={{ width: 380 }}
    >
      <Badge count={unreadCount} size="small" offset={[-4, 4]}>
        <Button
          type="text"
          icon={<BellOutlined />}
          aria-label={`Notifications${unreadCount > 0 ? ` (${unreadCount} unread)` : ""}`}
          style={{ fontSize: 16 }}
        />
      </Badge>
    </Popover>
  );
}
