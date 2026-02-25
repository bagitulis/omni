import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import Header from "@/components/layout/Header";
import { BrowserRouter, useNavigate } from "react-router-dom";
import { useAuthStore } from "@/stores/authStore";
import { useTheme } from "@/contexts/ThemeContext.hooks";
import apiClient from "@/api/client";

// Mock external modules and stores
vi.mock("@/stores/authStore", () => ({
  useAuthStore: vi.fn(),
}));

vi.mock("@/contexts/ThemeContext.hooks", () => ({
  useTheme: vi.fn(),
}));

vi.mock("@/api/client", () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
  },
}));

vi.mock("react-router-dom", async () => {
  const actual =
    await vi.importActual<typeof import("react-router-dom")>(
      "react-router-dom",
    );
  return {
    ...actual,
    useNavigate: vi.fn(),
  };
});

// Mock TokenStatusDropdown since it's used inside Header
vi.mock("@/components/layout/TokenStatusDropdown", () => ({
  TokenStatusDropdown: () => <div data-testid="token-status-dropdown" />,
}));

describe("Header", () => {
  const mockLogout = vi.fn();
  const mockNavigate = vi.fn();
  const mockToggleTheme = vi.fn();
  const mockUser = { username: "Test User", role: "admin" };

  beforeEach(() => {
    vi.resetAllMocks();
    (useAuthStore as any).mockReturnValue({
      user: mockUser,
      logout: mockLogout,
      tenantId: "test-tenant",
      setAuth: vi.fn(),
    });
    (useTheme as any).mockReturnValue({
      isDark: false,
      toggle: mockToggleTheme,
    });
    (useNavigate as any).mockReturnValue(mockNavigate);
  });

  const renderHeader = (props = {}) => {
    const defaultProps = {
      collapsed: false,
      onCollapse: vi.fn(),
    };
    return render(
      <BrowserRouter>
        <Header {...defaultProps} {...props} />
      </BrowserRouter>,
    );
  };

  it("renders header elements correctly", () => {
    renderHeader();
    expect(screen.getByText("OMNI")).toBeInTheDocument();
    expect(screen.getByText("Test User")).toBeInTheDocument();
    expect(screen.getByTestId("token-status-dropdown")).toBeInTheDocument();
  });

  it("calls onCollapse when toggle button is clicked", () => {
    const onCollapse = vi.fn();
    renderHeader({ onCollapse });
    const toggleBtn = screen.getByLabelText("Collapse sidebar");
    fireEvent.click(toggleBtn);
    expect(onCollapse).toHaveBeenCalled();
  });

  it("toggles theme when theme button is clicked", () => {
    renderHeader();
    const themeBtn = screen.getByLabelText("Switch to dark mode");
    fireEvent.click(themeBtn);
    expect(mockToggleTheme).toHaveBeenCalled();
  });

  it("renders developer tenant switcher only for developer role", async () => {
    // 1. Test as Admin (should NOT see switcher)
    renderHeader();
    expect(screen.queryByRole("combobox")).not.toBeInTheDocument();

    // 2. Test as Developer (should see switcher)
    (useAuthStore as any).mockReturnValue({
      user: { ...mockUser, role: "developer" },
      logout: mockLogout,
      tenantId: "test-tenant",
    });

    (apiClient.get as any).mockResolvedValue({
      success: true,
      tenants: [{ id: "tenant1", shop_name: "Shop 1" }],
    });

    renderHeader();

    await waitFor(() => {
      // Ant Design Select renders an input with role="combobox" (or similar depending on version/aria)
      // We can check for the class "tenant-switcher" which is applied in the component
      const switcher = document.querySelector(".tenant-switcher");
      expect(switcher).toBeInTheDocument();
    });
  });
});
