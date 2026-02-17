import { render, screen, fireEvent } from "@testing-library/react";
import { describe, it, expect, vi } from "vitest";
import { PlatformIndicator } from "./PlatformIndicator";
import type { PlatformIndicatorData } from "@/types/shared";
import "@testing-library/jest-dom";

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

  it("renders correctly for linked platform (SUCCESS state)", () => {
    render(<PlatformIndicator data={mockData} />);
    const indicator = screen.getByTestId("platform-indicator-shopee");

    expect(indicator).toBeInTheDocument();
  });

  it("renders NOT_LINKED state correctly (opacity 0.2 + grayscale)", () => {
    const notLinkedData: PlatformIndicatorData = {
      ...mockData,
      linked: false,
      sync_state: "idle",
    };
    render(<PlatformIndicator data={notLinkedData} />);
    const indicator = screen.getByTestId("platform-indicator-shopee");

    expect(indicator).toHaveStyle({
      opacity: "0.2",
      filter: "grayscale(100%)",
    });
  });

  it("renders ERROR state correctly (red border + animation)", () => {
    const errorData: PlatformIndicatorData = {
      ...mockData,
      sync_state: "error",
      error_message: "Sync failed",
    };
    render(<PlatformIndicator data={errorData} />);
    const indicator = screen.getByTestId("platform-indicator-shopee");

    // Check animation style presence
    expect(indicator).toHaveStyle({
      animation: "shake 0.4s ease-in-out",
    });
  });

  it("renders SYNCING state correctly (blue border)", () => {
    const syncingData: PlatformIndicatorData = {
      ...mockData,
      sync_state: "syncing",
    };
    render(<PlatformIndicator data={syncingData} />);
    const indicator = screen.getByTestId("platform-indicator-shopee");

    expect(indicator).toBeInTheDocument();
  });

  it("renders HAS_UPDATE state correctly (amber border)", () => {
    const updateData: PlatformIndicatorData = {
      ...mockData,
      has_update: true,
      sync_state: "success",
    };
    render(<PlatformIndicator data={updateData} />);
    const indicator = screen.getByTestId("platform-indicator-shopee");
    expect(indicator).toBeInTheDocument();
  });

  it("handles onClick event", () => {
    const handleClick = vi.fn();
    render(<PlatformIndicator data={mockData} onClick={handleClick} />);

    const indicator = screen.getByTestId("platform-indicator-shopee");
    fireEvent.click(indicator);

    expect(handleClick).toHaveBeenCalledTimes(1);
    expect(handleClick).toHaveBeenCalledWith("shopee");
  });

  it("renders small size correctly", () => {
    render(<PlatformIndicator data={mockData} size="small" />);
    const indicator = screen.getByTestId("platform-indicator-shopee");

    expect(indicator).toHaveStyle({
      width: "24px",
      height: "24px",
    });
  });
});
