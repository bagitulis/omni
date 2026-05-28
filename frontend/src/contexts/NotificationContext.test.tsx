import { render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { NotificationProvider, useNotifications } from "./NotificationContext";
import type { Notification } from "@/api/notifications";

const { mockGetSSETicket, mockGetValidToken, mockList, mockUnreadCount, mockInfo, mockAuthState } = vi.hoisted(() => {
  const getValidToken = vi.fn();
  return {
    mockGetSSETicket: vi.fn(),
    mockGetValidToken: getValidToken,
    mockList: vi.fn(),
    mockUnreadCount: vi.fn(),
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
  const { notifications, unreadCount, loading } = useNotifications();
  return (
    <div>
      <span data-testid="loading">{loading ? "loading" : "loaded"}</span>
      <span data-testid="notification-count">{notifications.length}</span>
      <span data-testid="unread-count">{unreadCount}</span>
      <span data-testid="notification-ids">{notifications.map(n => n.id).join(",")}</span>
      <span data-testid="computed-unread">
        {notifications.filter(n => !n.read).length}
      </span>
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
