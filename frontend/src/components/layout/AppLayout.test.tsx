import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import AppLayout from "@/components/layout/AppLayout";
import { BrowserRouter } from "react-router-dom";

// Mock child components
vi.mock("@/components/layout/Sidebar", () => ({
  default: ({
    collapsed,
    onCollapse,
  }: {
    collapsed: boolean;
    onCollapse: (v: boolean) => void;
  }) => (
    <div data-testid="sidebar">
      Sidebar Collapsed: {collapsed.toString()}
      <button onClick={() => onCollapse(!collapsed)}>Toggle Sidebar</button>
    </div>
  ),
}));

vi.mock("@/components/layout/Header", () => ({
  default: ({
    collapsed,
    onCollapse,
  }: {
    collapsed: boolean;
    onCollapse: () => void;
  }) => (
    <div data-testid="header">
      Header Collapsed: {collapsed.toString()}
      <button onClick={onCollapse}>Toggle Header</button>
    </div>
  ),
}));

vi.mock("@/components/layout/MobileNav", () => ({
  default: () => <div data-testid="mobile-nav">MobileNav</div>,
}));

// Mock react-router-dom Outlet
vi.mock("react-router-dom", async () => {
  const actual =
    await vi.importActual<typeof import("react-router-dom")>(
      "react-router-dom",
    );
  return {
    ...actual,
    Outlet: () => <div data-testid="outlet">Outlet Content</div>,
  };
});

describe("AppLayout", () => {
  const renderComponent = () =>
    render(
      <BrowserRouter>
        <AppLayout />
      </BrowserRouter>,
    );

  it("renders all layout components", () => {
    renderComponent();
    expect(screen.getByTestId("sidebar")).toBeInTheDocument();
    expect(screen.getByTestId("header")).toBeInTheDocument();
    expect(screen.getByTestId("mobile-nav")).toBeInTheDocument();
    expect(screen.getByTestId("outlet")).toBeInTheDocument();
  });

  it("manages collapsed state correctly", () => {
    renderComponent();

    // Initial state
    expect(screen.getByText("Sidebar Collapsed: false")).toBeInTheDocument();
    expect(screen.getByText("Header Collapsed: false")).toBeInTheDocument();

    // Toggle from sidebar
    fireEvent.click(screen.getByText("Toggle Sidebar"));
    expect(screen.getByText("Sidebar Collapsed: true")).toBeInTheDocument();
    expect(screen.getByText("Header Collapsed: true")).toBeInTheDocument();

    // Toggle from header
    fireEvent.click(screen.getByText("Toggle Header"));
    expect(screen.getByText("Sidebar Collapsed: false")).toBeInTheDocument();
    expect(screen.getByText("Header Collapsed: false")).toBeInTheDocument();
  });

  it("renders skip to content link for accessibility", () => {
    renderComponent();
    const skipLink = screen.getByText("Skip to main content");
    expect(skipLink).toBeInTheDocument();
    expect(skipLink).toHaveAttribute("href", "#main-content");
  });
});
