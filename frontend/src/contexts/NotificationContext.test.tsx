import { render, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { NotificationProvider } from "./NotificationContext";

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
