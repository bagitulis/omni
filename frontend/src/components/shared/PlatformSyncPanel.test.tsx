import "@testing-library/jest-dom";
import {
  fireEvent,
  render,
  screen,
  waitFor,
  within,
} from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { importFromStaging } from "@/api/products";
import { syncProductsToDb } from "@/api/lazadaDb";
import { syncShopeeProducts } from "@/api/shopeeDb";
import { searchProducts } from "@/api/tiktokDb";
import { PlatformSyncPanel } from "./PlatformSyncPanel";

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

globalThis.ResizeObserver = vi.fn().mockImplementation(() => ({
  observe: vi.fn(),
  unobserve: vi.fn(),
  disconnect: vi.fn(),
}));

vi.mock("@/api/products", () => ({
  importFromStaging: vi.fn(),
}));

vi.mock("@/api/shopeeDb", () => ({
  syncShopeeProducts: vi.fn(),
}));

vi.mock("@/api/tiktokDb", () => ({
  searchProducts: vi.fn(),
}));

vi.mock("@/api/lazadaDb", () => ({
  syncProductsToDb: vi.fn(),
}));

describe("PlatformSyncPanel", () => {
  beforeEach(() => {
    vi.clearAllMocks();

    vi.mocked(syncShopeeProducts).mockResolvedValue({
      message: "Synced 20 products",
      processed: 20,
    });
    vi.mocked(searchProducts).mockResolvedValue([
      { product_id: "TK-1" },
      { product_id: "TK-2" },
    ]);
    vi.mocked(syncProductsToDb).mockResolvedValue({
      message: "Synced 10 products",
      processed: 10,
    });
  });

  it("renders all platform cards", () => {
    render(<PlatformSyncPanel />);

    expect(screen.getByTestId("platform-sync-panel")).toBeInTheDocument();
    expect(screen.getByTestId("platform-sync-card-shopee")).toBeInTheDocument();
    expect(screen.getByTestId("platform-sync-card-tiktok")).toBeInTheDocument();
    expect(screen.getByTestId("platform-sync-card-lazada")).toBeInTheDocument();
  });

  it("imports from staging and refreshes products callback", async () => {
    const onImportCompleted = vi.fn().mockResolvedValue(undefined);
    vi.mocked(importFromStaging).mockResolvedValue({
      products_created: 2,
      products_matched: 1,
      products_skipped: 0,
      skus_created: 3,
      skus_skipped: 0,
      links_created: 3,
      errors: [],
    });

    render(<PlatformSyncPanel onImportCompleted={onImportCompleted} />);

    fireEvent.click(screen.getByTestId("platform-import-master-shopee"));

    await waitFor(() => {
      expect(importFromStaging).toHaveBeenCalledWith("shopee");
    });

    await waitFor(() => {
      expect(onImportCompleted).toHaveBeenCalled();
    });

    const resultBox = screen.getByTestId("platform-result-shopee");
    expect(resultBox).toBeInTheDocument();
    expect(within(resultBox).getByText(/Created P: 2/)).toBeInTheDocument();
    expect(within(resultBox).getByText(/SKUs: 3/)).toBeInTheDocument();
  });

  it("shows raw import error message", async () => {
    vi.mocked(importFromStaging).mockRejectedValue(
      new Error("shopee API error [E1001]: Invalid access token"),
    );

    render(<PlatformSyncPanel />);

    fireEvent.click(screen.getByTestId("platform-import-master-shopee"));

    await waitFor(() => {
      expect(
        screen.getByText("shopee API error [E1001]: Invalid access token"),
      ).toBeInTheDocument();
    });
  });

  it("shows sync message after TikTok sync", async () => {
    render(<PlatformSyncPanel />);

    fireEvent.click(screen.getByTestId("platform-sync-db-tiktok"));

    await waitFor(() => {
      expect(searchProducts).toHaveBeenCalled();
    });

    expect(screen.getByText("Synced 2 products")).toBeInTheDocument();
  });
});
