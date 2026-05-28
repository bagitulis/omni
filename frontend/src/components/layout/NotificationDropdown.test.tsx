import type { Notification } from "@/api/notifications";
import { render, screen } from "@testing-library/react";
import { NotificationDropdown } from "./NotificationDropdown";
import { MemoryRouter } from "react-router-dom";
import { beforeEach, describe, expect, it, vi } from "vitest";

const mockUseNotifications = vi.fn();

vi.mock("@/contexts/NotificationContext", () => ({
  useNotifications: () => mockUseNotifications(),
}));

function renderDropdown() {
  return render(
    <MemoryRouter>
      <NotificationDropdown />
    </MemoryRouter>,
  );
}

function baseContext(overrides: Partial<ReturnType<typeof contextValue>> = {}) {
  return {
    ...contextValue(),
    ...overrides,
  };
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

describe("NotificationDropdown", () => {
  beforeEach(() => {
    mockUseNotifications.mockReset();
  });

  it("renders a skeleton while loading", () => {
    mockUseNotifications.mockReturnValue(baseContext({ loading: true }));

    renderDropdown();

    expect(document.querySelector(".ant-skeleton")).toBeInTheDocument();
  });

  it("renders the no notifications empty state", () => {
    mockUseNotifications.mockReturnValue(baseContext());

    renderDropdown();

    expect(screen.getByText("No notifications")).toBeInTheDocument();
  });

  it("renders long notification titles through the ellipsis title class", () => {
    const longTitle = "Very long notification title that needs to stay inside the dropdown without overflowing";
    mockUseNotifications.mockReturnValue(baseContext({
      notifications: [
        {
          id: 1,
          type: "info",
          category: "system",
          title: longTitle,
          message: "Body",
          read: false,
          created_at: new Date().toISOString(),
        },
      ],
      unreadCount: 1,
    }));

    renderDropdown();

    expect(screen.getByText(longTitle).closest(".notification-item__title")).toBeInTheDocument();
  });
});
