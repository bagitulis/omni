import { useRef, useEffect, useCallback } from "react";
import { Badge, Button, Popover } from "antd";
import { BellOutlined } from "@ant-design/icons";
import {
  useNotificationStore,
  selectUnreadCount,
} from "@/stores/notificationStore";
import { NotificationDropdown } from "./NotificationDropdown";

/**
 * Bell icon with unread badge.  Clicking opens a popover with
 * the NotificationDropdown panel.
 */
export function NotificationBell() {
  const unreadCount = useNotificationStore(selectUnreadCount);
  const isOpen = useNotificationStore((s) => s.isDropdownOpen);
  const toggleDropdown = useNotificationStore((s) => s.toggleDropdown);
  const closeDropdown = useNotificationStore((s) => s.closeDropdown);

  // Close on outside click
  const containerRef = useRef<HTMLDivElement>(null);
  const handleClickOutside = useCallback(
    (e: MouseEvent) => {
      if (
        isOpen &&
        containerRef.current &&
        !containerRef.current.contains(e.target as Node)
      ) {
        closeDropdown();
      }
    },
    [isOpen, closeDropdown],
  );

  useEffect(() => {
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, [handleClickOutside]);

  return (
    <div ref={containerRef}>
      <Popover
        content={<NotificationDropdown />}
        trigger="click"
        open={isOpen}
        onOpenChange={(open) => {
          if (open) toggleDropdown();
          else closeDropdown();
        }}
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
