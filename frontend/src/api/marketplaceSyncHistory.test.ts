import { beforeEach, describe, expect, it, vi } from "vitest";
import {
  getMarketplaceSyncHistory,
  createMarketplaceSyncHistoryEntry,
} from "./marketplaceSyncHistory";

const { mockGet, mockPost } = vi.hoisted(() => ({
  mockGet: vi.fn(),
  mockPost: vi.fn(),
}));

vi.mock("./client", () => ({
  default: {
    get: mockGet,
    post: mockPost,
  },
}));

describe("marketplaceSyncHistory api", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  // --- getMarketplaceSyncHistory ---

  it("calls GET /marketplace-sync-history with no query string when no filter provided", async () => {
    mockGet.mockResolvedValue({
      success: true,
      data: { entries: [], total: 0, page: 1, page_size: 20 },
    });

    await getMarketplaceSyncHistory();

    expect(mockGet).toHaveBeenCalledWith("/marketplace-sync-history");
  });

  it("builds query string with all provided filter params", async () => {
    mockGet.mockResolvedValue({
      success: true,
      data: { entries: [], total: 0, page: 2, page_size: 10 },
    });

    await getMarketplaceSyncHistory({
      page: 2,
      page_size: 10,
      platform: "shopee",
      operation: "stock_update",
      status: "success",
      sku_search: "SKU-001",
      date_from: "2025-01-01",
      date_to: "2025-01-31",
    });

    const calledUrl = mockGet.mock.calls[0]?.[0] as string;
    expect(calledUrl).toContain("/marketplace-sync-history?");
    expect(calledUrl).toContain("page=2");
    expect(calledUrl).toContain("page_size=10");
    expect(calledUrl).toContain("platform=shopee");
    expect(calledUrl).toContain("operation=stock_update");
    expect(calledUrl).toContain("status=success");
    expect(calledUrl).toContain("sku_search=SKU-001");
    expect(calledUrl).toContain("date_from=2025-01-01");
    expect(calledUrl).toContain("date_to=2025-01-31");
  });

  it("omits undefined/absent filter params from query string", async () => {
    mockGet.mockResolvedValue({
      success: true,
      data: { entries: [], total: 0, page: 1, page_size: 20 },
    });

    await getMarketplaceSyncHistory({ platform: "tiktok" });

    const calledUrl = mockGet.mock.calls[0]?.[0] as string;
    expect(calledUrl).toContain("platform=tiktok");
    expect(calledUrl).not.toContain("page=");
    expect(calledUrl).not.toContain("status=");
    expect(calledUrl).not.toContain("sku_search=");
  });

  it("returns the data payload on successful response", async () => {
    const expectedData = {
      entries: [
        {
          id: "entry-1",
          tenant_id: "t1",
          sku: "SKU-001",
          platform: "shopee" as const,
          operation: "stock_update" as const,
          status: "success" as const,
          created_at: "2025-01-01T00:00:00Z",
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
    };

    mockGet.mockResolvedValue({ success: true, data: expectedData });

    const result = await getMarketplaceSyncHistory();

    expect(result).toEqual(expectedData);
  });

  it("throws raw backend error message when success=false", async () => {
    mockGet.mockResolvedValue({
      success: false,
      error: "marketplace API error: tenant not found",
    });

    await expect(getMarketplaceSyncHistory()).rejects.toThrow(
      "marketplace API error: tenant not found",
    );
  });

  it("throws fallback error when success=false and no error message", async () => {
    mockGet.mockResolvedValue({ success: false });

    await expect(getMarketplaceSyncHistory()).rejects.toThrow(
      "Failed to fetch marketplace sync history",
    );
  });

  // --- createMarketplaceSyncHistoryEntry ---

  it("calls POST /marketplace-sync-history with entry data", async () => {
    const entry = {
      sku: "SKU-001",
      platform: "shopee" as const,
      operation: "stock_update" as const,
      status: "success" as const,
      request_data: '{"stock":10}',
      response_data: '{"updated":true}',
    };

    mockPost.mockResolvedValue({
      success: true,
      data: {
        id: "new-entry-1",
        tenant_id: "t1",
        created_at: "2025-01-01T00:00:00Z",
        ...entry,
      },
    });

    const result = await createMarketplaceSyncHistoryEntry(entry);

    expect(mockPost).toHaveBeenCalledWith("/marketplace-sync-history", entry);
    expect(result.id).toBe("new-entry-1");
    expect(result.sku).toBe("SKU-001");
  });

  it("throws error on failed createMarketplaceSyncHistoryEntry", async () => {
    mockPost.mockResolvedValue({
      success: false,
      error: "tiktok API error [code=ERR_01]: invalid SKU",
    });

    await expect(
      createMarketplaceSyncHistoryEntry({
        sku: "BAD-SKU",
        platform: "tiktok",
        operation: "price_update",
        status: "failed",
      }),
    ).rejects.toThrow("tiktok API error [code=ERR_01]: invalid SKU");
  });
});
