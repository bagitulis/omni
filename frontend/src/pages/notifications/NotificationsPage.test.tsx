import type { Notification } from "@/api/notifications";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import NotificationsPage from "./NotificationsPage";
import { MemoryRouter } from "react-router-dom";
import { beforeEach, describe, expect, it, vi } from "vitest";

const mockUseNotifications = vi.fn();

vi.mock("@/contexts/NotificationContext", () => ({
  useNotifications: () => mockUseNotifications(),
}));

function renderPage() {
  return render(
    <MemoryRouter>
      <NotificationsPage />
    </MemoryRouter>,
  );
}

function contextValue() {
  return {
    notifications: [] as Notification[],
    unreadCount: 0,
    loading: false,
    connected: true,
    markAsRead: vi.fn(),
    markAllAsRead: vi.fn(),
    deleteNotification: vi.fn(),
    fetchNotifications: vi.fn(),
    addNotification: vi.fn(),
  };
}

describe("NotificationsPage", () => {
  beforeEach(() => {
    mockUseNotifications.mockReset();
  });

  it("renders loading and empty states", () => {
    mockUseNotifications.mockReturnValue({ ...contextValue(), loading: true });
    const { rerender } = renderPage();

    expect(document.querySelector(".ant-skeleton")).toBeInTheDocument();

    mockUseNotifications.mockReturnValue(contextValue());
    rerender(
      <MemoryRouter>
        <NotificationsPage />
      </MemoryRouter>,
    );

    expect(screen.getByText("No notifications")).toBeInTheDocument();
  });

  it("renders an error state with retry when refresh fails", async () => {
    const fetchNotifications = vi.fn().mockRejectedValue(new Error("network"));
    mockUseNotifications.mockReturnValue({ ...contextValue(), fetchNotifications });

    renderPage();
    fireEvent.click(screen.getByRole("button", { name: /refresh/i }));

    await waitFor(() => {
      expect(screen.getByText("Unable to refresh notifications. Try again in a moment.")).toBeInTheDocument();
    });
  });

  it("renders long titles with the page ellipsis class", () => {
    const longTitle = "Very long notification title that needs ellipsis handling on the notifications page";
    mockUseNotifications.mockReturnValue({
      ...contextValue(),
      notifications: [
        {
          id: 7,
          type: "success",
          category: "sync",
          title: longTitle,
          message: "Message",
          read: false,
          created_at: new Date().toISOString(),
        },
      ],
    });

    renderPage();

    expect(screen.getByText(longTitle).closest(".notification-page-item__title")).toBeInTheDocument();
  });
});
