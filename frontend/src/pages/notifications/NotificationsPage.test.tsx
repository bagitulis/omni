import type { Notification, NotificationCounts } from "@/api/notifications";
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

function makeCounts(): NotificationCounts {
  return { total: 0, unread: 0, by_severity: {} };
}

function contextValue() {
  return {
    notifications: [] as Notification[],
    counts: makeCounts(),
    unreadCount: 0,
    loading: false,
    markAsRead: vi.fn(),
    markAllAsRead: vi.fn(),
    bulkMarkRead: vi.fn(),
    bulkDelete: vi.fn(),
    snooze: vi.fn(),
    deleteNotification: vi.fn(),
    fetchNotifications: vi.fn(),
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
      expect(
        screen.getByText("Unable to refresh notifications. Try again in a moment."),
      ).toBeInTheDocument();
    });
  });

  it("renders long titles inside the master list item", () => {
    const longTitle =
      "Very long notification title that needs ellipsis handling on the notifications page";
    mockUseNotifications.mockReturnValue({
      ...contextValue(),
      counts: { total: 1, unread: 1, by_severity: { "20": 1 } },
      notifications: [
        {
          id: 7,
          type: "success",
          category: "sync",
          severity: 20,
          title: longTitle,
          message: "Message",
          dedup_count: 1,
          read: false,
          created_at: new Date().toISOString(),
        },
      ],
    });

    renderPage();

    // The new master-detail layout renders the title inside .notification-list-item__title.
    expect(
      screen.getByText(longTitle).closest(".notification-list-item__title"),
    ).toBeInTheDocument();
  });
});
