import "@testing-library/jest-dom";
import { render, screen, fireEvent } from "@testing-library/react";
import type { ProductPlatform } from "../types";
import { ProductSyncTab } from "./ProductSyncTab";

Object.defineProperty(window, "matchMedia", {
  writable: true,
  value: vi.fn().mockImplementation((query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
});

const mockPlatforms: ProductPlatform[] = [
  { platform: "shopee", status: "synced", last_sync: "2024-01-01" },
  { platform: "lazada", status: "pending", last_sync: "2024-01-02" },
  { platform: "tokopedia", status: "failed", last_sync: "2024-01-03" },
];

describe("ProductSyncTab", () => {
  const onSync = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders table with platform rows", () => {
    render(
      <ProductSyncTab
        platforms={mockPlatforms}
        onSync={onSync}
        loading={false}
      />,
    );
    expect(screen.getByText("shopee")).toBeInTheDocument();
    expect(screen.getByText("lazada")).toBeInTheDocument();
    expect(screen.getByText("tokopedia")).toBeInTheDocument();
  });

  it("renders column headers", () => {
    render(
      <ProductSyncTab
        platforms={mockPlatforms}
        onSync={onSync}
        loading={false}
      />,
    );
    expect(screen.getByText("Platform")).toBeInTheDocument();
    expect(screen.getByText("Status")).toBeInTheDocument();
    expect(screen.getByText("Last Sync")).toBeInTheDocument();
    expect(screen.getByText("Action")).toBeInTheDocument();
  });

  it("renders uppercase status text", () => {
    render(
      <ProductSyncTab
        platforms={mockPlatforms}
        onSync={onSync}
        loading={false}
      />,
    );
    expect(screen.getByText("SYNCED")).toBeInTheDocument();
    expect(screen.getByText("PENDING")).toBeInTheDocument();
    expect(screen.getByText("FAILED")).toBeInTheDocument();
  });

  it("renders Sync Now button for each platform", () => {
    render(
      <ProductSyncTab
        platforms={mockPlatforms}
        onSync={onSync}
        loading={false}
      />,
    );
    const buttons = screen.getAllByRole("button", { name: /sync now/i });
    expect(buttons).toHaveLength(3);
  });

  it("calls onSync with correct platform when Sync Now clicked", () => {
    render(
      <ProductSyncTab
        platforms={mockPlatforms}
        onSync={onSync}
        loading={false}
      />,
    );
    const buttons = screen.getAllByRole("button", { name: /sync now/i });
    fireEvent.click(buttons[0]);
    expect(onSync).toHaveBeenCalledWith("shopee");
  });

  it("renders empty table when no platforms", () => {
    render(<ProductSyncTab platforms={[]} onSync={onSync} loading={false} />);
    // Table still renders but with no rows
    expect(screen.getByText("Platform")).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: /sync now/i }),
    ).not.toBeInTheDocument();
  });

  it("renders last_sync dates in table", () => {
    render(
      <ProductSyncTab
        platforms={mockPlatforms}
        onSync={onSync}
        loading={false}
      />,
    );
    expect(screen.getByText("2024-01-01")).toBeInTheDocument();
    expect(screen.getByText("2024-01-02")).toBeInTheDocument();
  });
});
