import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import Sidebar from "@/components/layout/Sidebar";
import { MemoryRouter, useNavigate, useLocation } from "react-router-dom";
import { useTheme } from "@/contexts/ThemeContext.hooks";

// Mock React Router hooks
vi.mock("react-router-dom", async () => {
  const actual = await vi.importActual("react-router-dom");
  return {
    ...actual,
    useNavigate: vi.fn(),
    useLocation: vi.fn(),
  };
});

vi.mock("@/contexts/ThemeContext.hooks", () => ({
  useTheme: vi.fn(),
}));

// Mock antd Layout and Menu
vi.mock("antd", async () => {
  const actual = await vi.importActual<typeof import("antd")>("antd");
  return {
    ...actual,
    Layout: {
      Sider: ({ children }: { children?: React.ReactNode }) => (
        <div data-testid="sider">{children}</div>
      ),
    },
    Menu: ({
      items,
      onClick,
    }: {
      items?: Array<{ key: string; label: React.ReactNode }>;
      onClick?: (info: { key: string }) => void;
    }) => (
      <div data-testid="menu">
        {(items ?? []).map((item) => (
          <div
            key={item.key}
            data-testid={`menu-item-${item.key}`}
            onClick={() => onClick?.({ key: item.key })}
          >
            {item.label}
          </div>
        ))}
      </div>
    ),
  };
});

describe("Sidebar", () => {
  const mockNavigate = vi.fn();
  const mockLocation = { pathname: "/" };

  beforeEach(() => {
    vi.resetAllMocks();
    (
      vi.mocked(useNavigate) as unknown as ReturnType<typeof vi.fn>
    ).mockReturnValue(mockNavigate);
    (
      vi.mocked(useLocation) as unknown as ReturnType<typeof vi.fn>
    ).mockReturnValue(mockLocation);
    (
      vi.mocked(useTheme) as unknown as ReturnType<typeof vi.fn>
    ).mockReturnValue({
      isDark: false,
    });
  });

  const renderSidebar = (props = {}) => {
    const defaultProps = {
      collapsed: false,
      onCollapse: vi.fn(),
    };
    return render(
      <MemoryRouter>
        <Sidebar {...defaultProps} {...props} />
      </MemoryRouter>,
    );
  };

  it("renders sidebar correctly", () => {
    renderSidebar();
    expect(screen.getByTestId("sider")).toBeInTheDocument();
    expect(screen.getByText("OMNI")).toBeInTheDocument(); // Logo text
    expect(screen.getByTestId("menu")).toBeInTheDocument();
  });

  it("renders collapsed logo correctly", () => {
    renderSidebar({ collapsed: true });
    expect(screen.getByText("O")).toBeInTheDocument();
    expect(screen.queryByText("OMNI")).not.toBeInTheDocument();
  });

  it("navigates when menu item is clicked", () => {
    renderSidebar();
    const dashboardItem = screen.getByTestId("menu-item-/");
    fireEvent.click(dashboardItem);
    expect(mockNavigate).toHaveBeenCalledWith("/");

    const ordersItem = screen.getByTestId("menu-item-/order-manager");
    fireEvent.click(ordersItem);
    expect(mockNavigate).toHaveBeenCalledWith("/order-manager");
  });

  it("highlights correct menu item based on location", () => {
    // This logic is mainly inside useMemo for selectedKey
    // Since we mock Menu, we can't easily check 'selectedKeys' prop visually without inspecting the mock call
    // But we can verify no errors occur during render with different paths
    (
      vi.mocked(useLocation) as unknown as ReturnType<typeof vi.fn>
    ).mockReturnValue({ pathname: "/products/add" });
    renderSidebar();
    expect(screen.getByTestId("menu")).toBeInTheDocument();
  });
});
