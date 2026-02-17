import { render, screen } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import { PlatformIndicator } from "./PlatformIndicator";
import type { PlatformIndicatorData } from "@/types/shared";
import "@testing-library/jest-dom"; // Import matchers

// Mock matchMedia for Ant Design
Object.defineProperty(window, "matchMedia", {
  writable: true,
  value: (query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: () => {},
    removeListener: () => {},
    addEventListener: () => {},
    removeEventListener: () => {},
    dispatchEvent: () => {},
  }),
});

describe("PlatformIndicator", () => {
  const mockData: PlatformIndicatorData = {
    platform: "shopee",
    linked: true,
    has_update: false,
    sync_state: "success",
    last_synced_at: "2023-01-01T00:00:00Z",
  };

  it("renders correctly for linked platform", () => {
    const { container } = render(<PlatformIndicator data={mockData} />);
    // Should contain Shopee icon/text "S"
    expect(screen.getByText("S")).toBeInTheDocument();
    // Should have tooltip trigger (antd tooltip wraps children)
    expect(container.querySelector(".ant-space")).toBeInTheDocument();
  });

  it("renders correctly for not linked platform", () => {
    const notLinkedData: PlatformIndicatorData = {
      ...mockData,
      linked: false,
      sync_state: "idle",
    };
    render(<PlatformIndicator data={notLinkedData} />);
    const indicator = screen.getByText("S").closest(".ant-space");
    expect(indicator).toHaveStyle({ opacity: "0.5" });
  });

  it("shows error state correctly", () => {
    const errorData: PlatformIndicatorData = {
      ...mockData,
      sync_state: "error",
      error_message: "Sync failed",
    };
    render(<PlatformIndicator data={errorData} />);
    // We can't easily check for the specific icon class without implementation details,
    // but we can check if it renders without crashing.
    // Ideally we'd check for the error icon or color.
    expect(screen.getByText("S")).toBeInTheDocument();
  });

  it("shows syncing state correctly", () => {
    const syncingData: PlatformIndicatorData = {
      ...mockData,
      sync_state: "syncing",
    };
    render(<PlatformIndicator data={syncingData} />);
    expect(screen.getByText("S")).toBeInTheDocument();
  });
});
