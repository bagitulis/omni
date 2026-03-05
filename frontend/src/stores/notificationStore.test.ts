import { describe, it, expect, beforeEach } from "vitest";
import { useNotificationStore, selectUnreadCount } from "./notificationStore";

describe("notificationStore", () => {
  beforeEach(() => {
    // Reset store between tests
    useNotificationStore.setState({ notifications: [], isDropdownOpen: false });
  });

  it("starts with empty notifications", () => {
    const state = useNotificationStore.getState();
    expect(state.notifications).toEqual([]);
    expect(state.isDropdownOpen).toBe(false);
  });

  it("adds a notification with auto-generated id, timestamp, and read=false", () => {
    const { addNotification } = useNotificationStore.getState();
    addNotification({
      type: "success",
      category: "sync",
      title: "Sync Complete",
      message: "3 products synced",
    });

    const { notifications } = useNotificationStore.getState();
    expect(notifications).toHaveLength(1);
    expect(notifications[0].title).toBe("Sync Complete");
    expect(notifications[0].read).toBe(false);
    expect(notifications[0].id).toBeTruthy();
    expect(notifications[0].timestamp).toBeGreaterThan(0);
  });

  it("prepends new notifications (newest first)", () => {
    const { addNotification } = useNotificationStore.getState();
    addNotification({
      type: "success",
      category: "sync",
      title: "First",
      message: "",
    });
    addNotification({
      type: "error",
      category: "order",
      title: "Second",
      message: "",
    });

    const { notifications } = useNotificationStore.getState();
    expect(notifications[0].title).toBe("Second");
    expect(notifications[1].title).toBe("First");
  });

  it("marks a notification as read", () => {
    const { addNotification } = useNotificationStore.getState();
    addNotification({
      type: "info",
      category: "system",
      title: "Test",
      message: "",
    });

    const id = useNotificationStore.getState().notifications[0].id;
    useNotificationStore.getState().markAsRead(id);

    expect(useNotificationStore.getState().notifications[0].read).toBe(true);
  });

  it("marks all as read", () => {
    const { addNotification } = useNotificationStore.getState();
    addNotification({
      type: "info",
      category: "system",
      title: "A",
      message: "",
    });
    addNotification({
      type: "error",
      category: "order",
      title: "B",
      message: "",
    });

    useNotificationStore.getState().markAllAsRead();

    const { notifications } = useNotificationStore.getState();
    expect(notifications.every((n) => n.read)).toBe(true);
  });

  it("removes a notification by id", () => {
    const { addNotification } = useNotificationStore.getState();
    addNotification({
      type: "success",
      category: "export",
      title: "Export",
      message: "",
    });

    const id = useNotificationStore.getState().notifications[0].id;
    useNotificationStore.getState().removeNotification(id);

    expect(useNotificationStore.getState().notifications).toHaveLength(0);
  });

  it("clears all notifications", () => {
    const { addNotification } = useNotificationStore.getState();
    addNotification({
      type: "success",
      category: "sync",
      title: "A",
      message: "",
    });
    addNotification({
      type: "error",
      category: "sync",
      title: "B",
      message: "",
    });

    useNotificationStore.getState().clearAll();
    expect(useNotificationStore.getState().notifications).toHaveLength(0);
  });

  it("toggles dropdown open/closed", () => {
    const { toggleDropdown } = useNotificationStore.getState();
    expect(useNotificationStore.getState().isDropdownOpen).toBe(false);
    toggleDropdown();
    expect(useNotificationStore.getState().isDropdownOpen).toBe(true);
    toggleDropdown();
    expect(useNotificationStore.getState().isDropdownOpen).toBe(false);
  });

  it("caps at 100 notifications", () => {
    const { addNotification } = useNotificationStore.getState();
    for (let i = 0; i < 110; i++) {
      addNotification({
        type: "info",
        category: "system",
        title: `#${i}`,
        message: "",
      });
    }
    expect(useNotificationStore.getState().notifications.length).toBeLessThanOrEqual(100);
  });
});

describe("selectUnreadCount", () => {
  beforeEach(() => {
    useNotificationStore.setState({ notifications: [], isDropdownOpen: false });
  });

  it("returns 0 when no notifications", () => {
    expect(selectUnreadCount(useNotificationStore.getState())).toBe(0);
  });

  it("counts only unread notifications", () => {
    const { addNotification, markAsRead } = useNotificationStore.getState();
    addNotification({
      type: "success",
      category: "sync",
      title: "A",
      message: "",
    });
    addNotification({
      type: "error",
      category: "order",
      title: "B",
      message: "",
    });

    const id = useNotificationStore.getState().notifications[0].id;
    markAsRead(id);

    expect(selectUnreadCount(useNotificationStore.getState())).toBe(1);
  });
});
