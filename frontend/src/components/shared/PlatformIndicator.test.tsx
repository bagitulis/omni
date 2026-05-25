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
  const baseData: PlatformIndicatorData = {
    platform: "shopee",
    linked: true,
    has_update: false,
    sync_state: "idle",
    last_synced_at: "2023-01-01T00:00:00Z",
  };

  const renderIndicator = (overrides: Partial<PlatformIndicatorData> = {}) => {
    const data: PlatformIndicatorData = {
      ...baseData,
      ...overrides,
    };

    render(<PlatformIndicator data={data} />);
    return screen.getByTestId(`platform-indicator-${data.platform}`);
  };

  const getTooltipTrigger = (indicator: HTMLElement): HTMLElement => {
    return indicator.parentElement ?? indicator;
  };

  it("renders platform icon image for each platform", () => {
    const { rerender } = render(<PlatformIndicator data={baseData} />);

    const shopeeImg = screen.getByTestId("platform-indicator-shopee").querySelector("img");
    expect(shopeeImg).toBeInTheDocument();
    expect(shopeeImg).toHaveAttribute("alt", "Shopee");

    rerender(<PlatformIndicator data={{ ...baseData, platform: "tiktok" }} />);
    const tiktokImg = screen.getByTestId("platform-indicator-tiktok").querySelector("img");
    expect(tiktokImg).toBeInTheDocument();
    expect(tiktokImg).toHaveAttribute("alt", "TikTok");

    rerender(<PlatformIndicator data={{ ...baseData, platform: "lazada" }} />);
    const lazadaImg = screen.getByTestId("platform-indicator-lazada").querySelector("img");
    expect(lazadaImg).toBeInTheDocument();
    expect(lazadaImg).toHaveAttribute("alt", "Lazada");
  });

  it("renders NOT_LINKED state, keeps tooltip visible, and enforces non-interactive behavior", async () => {
    const handleClick = vi.fn();
    const data: PlatformIndicatorData = {
      ...baseData,
      linked: false,
    };

    render(<PlatformIndicator data={data} onClick={handleClick} />);
    const indicator = screen.getByTestId("platform-indicator-shopee");

    expect(indicator).toHaveStyle({
      opacity: "0.4",
      filter: "grayscale(80%)",
      cursor: "default",
    });
    expect(getComputedStyle(indicator).borderColor).toBe("rgb(217, 217, 217)");
    expect(indicator).toHaveAttribute("tabindex", "-1");

    fireEvent.mouseEnter(getTooltipTrigger(indicator));
    expect(await screen.findByText((content) => content.includes("Shopee: Not linked"))).toBeInTheDocument();

    fireEvent.click(indicator);
    expect(handleClick).not.toHaveBeenCalled();
  });

  it("renders LINKED idle state with exact green border and check badge", () => {
    const indicator = renderIndicator();

    expect(indicator).toHaveAttribute("data-state", "linked");
    expect(indicator).toHaveStyle({ opacity: "1" });
    expect(indicator).toHaveStyle({ borderRadius: "3px" });
    expect(getComputedStyle(indicator).borderColor).toBe("rgb(82, 196, 26)");
    expect(indicator).toHaveTextContent("✓");
  });

  it("shows LINKED tooltip content on hover", async () => {
    const indicator = renderIndicator({ last_synced_at: undefined });

    fireEvent.mouseEnter(getTooltipTrigger(indicator));
    expect(await screen.findByText("Shopee: Linked")).toBeInTheDocument();
  });

  it("renders HAS_UPDATE state with exact amber border and warning badge", () => {
    const indicator = renderIndicator({ has_update: true });

    expect(indicator).toHaveAttribute("data-state", "has_update");
    expect(getComputedStyle(indicator).borderColor).toBe("rgb(250, 173, 20)");
    expect(indicator).toHaveTextContent("!");
  });

  it("renders SYNCING state with exact blue border and spinner overlay", () => {
    const indicator = renderIndicator({
      sync_state: "syncing",
      has_update: true,
    });

    expect(indicator).toHaveAttribute("data-state", "syncing");
    expect(getComputedStyle(indicator).borderColor).toBe("rgb(22, 119, 255)");
    expect(indicator.querySelector(".platform-spinner")).toBeInTheDocument();
  });

  it("renders SUCCESS state with exact green border and popIn animation", () => {
    const indicator = renderIndicator({ sync_state: "success" });

    expect(indicator).toHaveAttribute("data-state", "success");
    expect(getComputedStyle(indicator).borderColor).toBe("rgb(82, 196, 26)");
    expect(indicator).toHaveStyle({ animation: "popIn 0.3s ease" });
  });

  it("renders ERROR state with exact red border and shake animation", () => {
    const indicator = renderIndicator({
      sync_state: "error",
      error_message: "Sync failed from API",
    });

    expect(indicator).toHaveAttribute("data-state", "error");
    expect(getComputedStyle(indicator).borderColor).toBe("rgb(255, 77, 79)");
    expect(indicator).toHaveStyle({ animation: "shake 0.5s ease" });
  });

  it("uses sync state precedence over linked state", () => {
    const indicator = renderIndicator({ linked: false, sync_state: "syncing" });

    expect(indicator).toHaveAttribute("data-state", "syncing");
    expect(getComputedStyle(indicator).borderColor).toBe("rgb(22, 119, 255)");
  });

  it("triggers onClick only when linked", () => {
    const handleClick = vi.fn();
    const { rerender } = render(
      <PlatformIndicator data={baseData} onClick={handleClick} />,
    );

    fireEvent.click(screen.getByTestId("platform-indicator-shopee"));
    expect(handleClick).toHaveBeenCalledWith("shopee");
    expect(handleClick).toHaveBeenCalledTimes(1);

    rerender(
      <PlatformIndicator
        data={{ ...baseData, linked: false }}
        onClick={handleClick}
      />,
    );
    fireEvent.click(screen.getByTestId("platform-indicator-shopee"));
    expect(handleClick).toHaveBeenCalledTimes(1);
  });

  it("renders small size correctly", () => {
    render(<PlatformIndicator data={baseData} size="small" />);
    const indicator = screen.getByTestId("platform-indicator-shopee");

    expect(indicator).toHaveStyle({
      width: "24px",
      height: "24px",
    });
  });
});
