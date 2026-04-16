import { useRef, useState, useCallback, useEffect } from "react";
import { Badge, Button, Popover } from "antd";
import { BellOutlined } from "@ant-design/icons";
import { NotificationDropdown } from "./NotificationDropdown";
import { useNotifications } from "@/contexts/NotificationContext";

/**
 * Bell icon with unread badge.  Clicking opens a popover with
 * the NotificationDropdown panel.
 */
export function NotificationBell() {
  const { unreadCount } = useNotifications();
  const [isOpen, setIsOpen] = useState(false);

  // Close on outside click
  const containerRef = useRef<HTMLDivElement>(null);
  const handleClickOutside = useCallback(
    (e: MouseEvent) => {
      if (
        isOpen &&
        containerRef.current &&
        !containerRef.current.contains(e.target as Node)
      ) {
        setIsOpen(false);
      }
    },
    [isOpen],
  );

  useEffect(() => {
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, [handleClickOutside]);

  return (
    <div ref={containerRef}>
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
    </div>
  );
}
