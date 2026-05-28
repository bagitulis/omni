import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { Notification } from "@/api/notifications";
import { NotificationProvider, useNotifications } from "./NotificationContext";

const { mockGetSSETicket, mockGetValidToken, mockList, mockUnreadCount, mockMarkAsRead, mockMarkAllAsRead, mockDeleteNotif, mockInfo, mockAuthState } = vi.hoisted(() => {
  const getValidToken = vi.fn();
  return {
    mockGetSSETicket: vi.fn(),
    mockGetValidToken: getValidToken,
    mockList: vi.fn(),
    mockUnreadCount: vi.fn(),
    mockMarkAsRead: vi.fn(),
    mockMarkAllAsRead: vi.fn(),
    mockDeleteNotif: vi.fn(),
    mockInfo: vi.fn(),
    mockAuthState: {
      tenantId: "tenant_test",
      getValidToken,
    },
  };
});

vi.mock("@/api/auth", () => ({
  getSSETicket: mockGetSSETicket,
}));

vi.mock("@/api/notifications", () => ({
  notificationApi: {
    list: mockList,
    getUnreadCount: mockUnreadCount,
    markAsRead: mockMarkAsRead,
    markAllAsRead: mockMarkAllAsRead,
    delete: mockDeleteNotif,
  },
}));

vi.mock("@/stores/authStore", () => ({
  useAuthStore: () => mockAuthState,
}));

vi.mock("@/lib/constants", () => ({
  API_BASE_URL: "/api",
}));

vi.mock("antd", () => ({
  App: {
    useApp: () => ({
      notification: {
        success: vi.fn(),
        error: vi.fn(),
        warning: vi.fn(),
        info: mockInfo,
      },
    }),
  },
}));

describe("NotificationProvider SSE setup", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockGetValidToken.mockResolvedValue("raw-jwt-token");
    mockList.mockResolvedValue({ success: true, data: { items: [] } });
    mockUnreadCount.mockResolvedValue({ success: true, data: { unread_count: 0 } });
  });

  it("does not open an SSE stream with a raw token when ticket creation fails", async () => {
    const eventSourceSpy = vi.fn();
    vi.stubGlobal("EventSource", eventSourceSpy);
    mockGetSSETicket.mockResolvedValue(null);

    render(
      <NotificationProvider>
        <div>child</div>
      </NotificationProvider>,
    );

    await waitFor(() => expect(mockGetSSETicket).toHaveBeenCalled());
    expect(eventSourceSpy).not.toHaveBeenCalled();
  });

  it("opens an SSE stream with a one-time ticket when ticket creation succeeds", async () => {
    const eventSourceSpy = vi.fn().mockImplementation(() => ({
      close: vi.fn(),
      addEventListener: vi.fn(),
    }));
    vi.stubGlobal("EventSource", eventSourceSpy);
    mockGetSSETicket.mockResolvedValue("sse-ticket");

    render(
      <NotificationProvider>
        <div>child</div>
      </NotificationProvider>,
    );

    await waitFor(() => expect(eventSourceSpy).toHaveBeenCalledWith("/api/notifications/stream?ticket=sse-ticket"));
  });
});

// ---------------------------------------------------------------------------
// Test consumer for inspecting context state
// ---------------------------------------------------------------------------
function NotificationStateConsumer() {
  const { notifications, unreadCount, loading, markAllAsRead, deleteNotification, fetchNotifications } = useNotifications();
  return (
    <div>
      <span data-testid="loading">{loading ? "loading" : "loaded"}</span>
      <span data-testid="notification-count">{notifications.length}</span>
      <span data-testid="unread-count">{unreadCount}</span>
      <span data-testid="notification-ids">{notifications.map(n => n.id).join(",")}</span>
      <span data-testid="computed-unread">
        {notifications.filter(n => !n.read).length}
      </span>
      <button type="button" onClick={() => void markAllAsRead()}>mark all</button>
      <button type="button" onClick={() => void deleteNotification(1)}>delete first</button>
      <button type="button" onClick={() => void fetchNotifications()}>fetch</button>
    </div>
  );
}

describe("NotificationProvider state consistency", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockGetValidToken.mockResolvedValue("raw-jwt-token");
    mockGetSSETicket.mockResolvedValue("sse-ticket");
  });

  it("derives unread count consistently from normalized server-backed data", async () => {
    const notifs: Notification[] = [
      { id: 1, type: "info", category: "sync", title: "A", message: "", read: false, created_at: new Date().toISOString() },
      { id: 2, type: "success", category: "order", title: "B", message: "", read: true, created_at: new Date().toISOString() },
      { id: 3, type: "warning", category: "system", title: "C", message: "", read: false, created_at: new Date().toISOString() },
    ];

    mockList.mockResolvedValue({ success: true, data: { items: notifs, count: 3 } });
    mockUnreadCount.mockResolvedValue({ success: true, data: { unread_count: 2 } });

    const mockEs = { close: vi.fn(), addEventListener: vi.fn() };
    vi.stubGlobal("EventSource", vi.fn(() => mockEs));

    render(
      <NotificationProvider>
        <NotificationStateConsumer />
      </NotificationProvider>,
    );

    await waitFor(() => {
      expect(screen.getByTestId("notification-count")).toHaveTextContent("3");
    });

    // Context reports 2 unread (matches API)
    expect(screen.getByTestId("unread-count")).toHaveTextContent("2");

    // Computed from local notifications array is also 2
    expect(screen.getByTestId("computed-unread")).toHaveTextContent("2");
  });

  it("does not duplicate notifications when SSE event arrives with existing id", async () => {
    const notif: Notification = {
      id: 42,
      type: "info",
      category: "sync",
      title: "Existing",
      message: "",
      read: false,
      created_at: new Date().toISOString(),
    };

    mockList.mockResolvedValue({ success: true, data: { items: [notif], count: 1 } });
    mockUnreadCount.mockResolvedValue({ success: true, data: { unread_count: 1 } });

    let triggerEvent: ((data: unknown) => void) | null = null;
    const mockEs = {
      close: vi.fn(),
      addEventListener: vi.fn((_type: string, handler: (event: MessageEvent) => void) => {
        triggerEvent = (data: unknown) => {
          handler({ data: JSON.stringify(data) } as MessageEvent);
        };
      }),
    };
    vi.stubGlobal("EventSource", vi.fn(() => mockEs));

    render(
      <NotificationProvider>
        <NotificationStateConsumer />
      </NotificationProvider>,
    );

    await waitFor(() => {
      expect(screen.getByTestId("notification-count")).toHaveTextContent("1");
    });

    // Dispatch duplicate SSE event with same id 42
    triggerEvent?.(notif);

    // Still 1 notification — idempotent upsert
    await waitFor(() => {
      expect(screen.getByTestId("notification-count")).toHaveTextContent("1");
    });

    // Verify it's the same single id 42, not duplicated
    expect(screen.getByTestId("notification-ids")).toHaveTextContent("42");
  });

});

// ---------------------------------------------------------------------------
// Action consumer for testing context mutation methods
// ---------------------------------------------------------------------------
function NotificationActionConsumer() {
  const { notifications, unreadCount, markAllAsRead, deleteNotification, fetchNotifications } = useNotifications();
  return (
    <div>
      <span data-testid="notification-count">{notifications.length}</span>
      <span data-testid="unread-count-global">{unreadCount}</span>
      <span data-testid="notification-ids">{notifications.map(n => n.id).join(",")}</span>
      <button type="button" data-testid="mark-all-read-btn" onClick={markAllAsRead}>Mark All Read</button>
      <button type="button" data-testid="delete-first-btn" onClick={() => notifications[0] && deleteNotification(notifications[0].id)}>Delete First</button>
      <button type="button" data-testid="fetch-btn" onClick={fetchNotifications}>Fetch</button>
    </div>
  );
}

describe("NotificationProvider mutations", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockGetValidToken.mockResolvedValue("raw-jwt-token");
    mockGetSSETicket.mockResolvedValue("sse-ticket");
  });

  it("markAllAsRead updates all notifications to read", async () => {
    const notifs: Notification[] = [
      { id: 1, type: "info", category: "sync", title: "A", message: "", read: false, created_at: new Date().toISOString() },
      { id: 2, type: "warning", category: "system", title: "B", message: "", read: false, created_at: new Date().toISOString() },
    ];

    const mockEs = { close: vi.fn(), addEventListener: vi.fn() };
    vi.stubGlobal("EventSource", vi.fn(() => mockEs));

    mockList.mockResolvedValue({ success: true, data: { items: notifs, count: 2 } });
    mockUnreadCount.mockResolvedValue({ success: true, data: { unread_count: 2 } });
    mockMarkAllAsRead.mockResolvedValue({ success: true });

    render(
      <NotificationProvider>
        <NotificationActionConsumer />
      </NotificationProvider>,
    );

    await waitFor(() => {
      expect(screen.getByTestId("notification-count")).toHaveTextContent("2");
    });

    fireEvent.click(screen.getByTestId("mark-all-read-btn"));

    await waitFor(() => {
      expect(screen.getByTestId("unread-count-global")).toHaveTextContent("0");
    });

    expect(mockMarkAllAsRead).toHaveBeenCalledTimes(1);
  });

  it("deleteNotification removes notification from list", async () => {
    const notifs: Notification[] = [
      { id: 10, type: "info", category: "sync", title: "Delete Me", message: "", read: false, created_at: new Date().toISOString() },
    ];

    const mockEs = { close: vi.fn(), addEventListener: vi.fn() };
    vi.stubGlobal("EventSource", vi.fn(() => mockEs));

    mockList.mockResolvedValue({ success: true, data: { items: notifs, count: 1 } });
    mockUnreadCount.mockResolvedValue({ success: true, data: { unread_count: 1 } });
    mockDeleteNotif.mockResolvedValue({ success: true });

    render(
      <NotificationProvider>
        <NotificationActionConsumer />
      </NotificationProvider>,
    );

    await waitFor(() => {
      expect(screen.getByTestId("notification-count")).toHaveTextContent("1");
      expect(screen.getByTestId("unread-count-global")).toHaveTextContent("1");
    });

    fireEvent.click(screen.getByTestId("delete-first-btn"));

    await waitFor(() => {
      expect(screen.getByTestId("notification-count")).toHaveTextContent("0");
    });

    expect(mockDeleteNotif).toHaveBeenCalledWith(10);
  });

  it("fetchNotifications merges by id without duplicates", async () => {
    // Initial state: notification with id 1 exists
    const existingNotif: Notification = {
      id: 1, type: "info", category: "sync", title: "Existing", message: "", read: false, created_at: new Date().toISOString(),
    };

    const mockEs = { close: vi.fn(), addEventListener: vi.fn() };
    vi.stubGlobal("EventSource", vi.fn(() => mockEs));

    mockList.mockResolvedValue({ success: true, data: { items: [existingNotif], count: 1 } });
    mockUnreadCount.mockResolvedValue({ success: true, data: { unread_count: 1 } });

    render(
      <NotificationProvider>
        <NotificationActionConsumer />
      </NotificationProvider>,
    );

    await waitFor(() => {
      expect(screen.getByTestId("notification-count")).toHaveTextContent("1");
    });

    // Re-fetch with same notification (updated title) + new one
    const updatedExisting: Notification = {
      ...existingNotif,
      title: "Updated Title",
    };
    const newNotif: Notification = {
      id: 2, type: "success", category: "order", title: "New", message: "", read: false, created_at: new Date().toISOString(),
    };

    // Second fetch returns updated existing + new
    mockList.mockResolvedValue({ success: true, data: { items: [updatedExisting, newNotif], count: 2 } });

    fireEvent.click(screen.getByTestId("fetch-btn"));

    await waitFor(() => {
      // Should have 2 notifications (updated existing + new), no duplicates
      expect(screen.getByTestId("notification-count")).toHaveTextContent("2");
    });

    // Both ids 1 and 2 present
    const ids = screen.getByTestId("notification-ids").textContent;
    expect(ids).toContain("1");
    expect(ids).toContain("2");

    // Only one occurrence of id 1 (no duplicate)
    expect(ids?.split(",").filter(id => id === "1").length).toBe(1);
  });

  it("deleteNotification decrements unread count correctly for read notification", async () => {
    const notifs: Notification[] = [
      { id: 20, type: "info", category: "sync", title: "Read Notif", message: "", read: true, created_at: new Date().toISOString() },
    ];

    const mockEs = { close: vi.fn(), addEventListener: vi.fn() };
    vi.stubGlobal("EventSource", vi.fn(() => mockEs));

    mockList.mockResolvedValue({ success: true, data: { items: notifs, count: 1 } });
    mockUnreadCount.mockResolvedValue({ success: true, data: { unread_count: 0 } });
    mockDeleteNotif.mockResolvedValue({ success: true });

    render(
      <NotificationProvider>
        <NotificationActionConsumer />
      </NotificationProvider>,
    );

    await waitFor(() => {
      expect(screen.getByTestId("notification-count")).toHaveTextContent("1");
    });

    // Unread count was 0 and deleting a read notification keeps it 0
    expect(screen.getByTestId("unread-count-global")).toHaveTextContent("0");

    fireEvent.click(screen.getByTestId("delete-first-btn"));

    await waitFor(() => {
      expect(screen.getByTestId("notification-count")).toHaveTextContent("0");
    });

    expect(screen.getByTestId("unread-count-global")).toHaveTextContent("0");
  });
});
