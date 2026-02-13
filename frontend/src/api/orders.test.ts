import { beforeEach, describe, expect, it, vi } from "vitest";
import { getOrders, syncOrdersByCategory } from "./orders";

const { mockClientGet, mockClientPost } = vi.hoisted(() => ({
  mockClientGet: vi.fn(),
  mockClientPost: vi.fn(),
}));

vi.mock("./client", () => ({
  default: {
    client: {
      get: mockClientGet,
      post: mockClientPost,
    },
  },
}));

describe("syncOrdersByCategory", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("syncs processed orders for all platforms with 30-day window", async () => {
    mockClientPost.mockResolvedValue({
      data: { success: true },
    });

    await syncOrdersByCategory("processed");

    expect(mockClientPost).toHaveBeenCalledWith(
      "/orders/sync/processed",
      { days: 30 },
      { timeout: 120_000 },
    );
  });

  it("throws backend error when processed sync fails", async () => {
    mockClientPost.mockResolvedValue({
      data: { success: false, error: "sync failed from backend" },
    });

    await expect(syncOrdersByCategory("processed")).rejects.toThrow(
      "sync failed from backend",
    );
  });

  it("syncs non-processed tab with 7-day window", async () => {
    mockClientPost.mockResolvedValue({
      data: { success: true },
    });

    await syncOrdersByCategory("unprocess");

    expect(mockClientPost).toHaveBeenCalledWith(
      "/orders/sync/unprocess",
      { days: 7 },
      { timeout: 120_000 },
    );
  });

  it("syncs platform-specific tab key with normalized category and platform filter", async () => {
    mockClientPost.mockResolvedValue({
      data: { success: true },
    });

    await syncOrdersByCategory("READY_TO_SHIP", "Shopee");

    expect(mockClientPost).toHaveBeenCalledWith(
      "/orders/sync/unprocess?platforms=shopee",
      { days: 7 },
      { timeout: 120_000 },
    );
  });

  it("does nothing for non-syncable tab", async () => {
    await syncOrdersByCategory("locked", "all");

    expect(mockClientPost).not.toHaveBeenCalled();
  });
});

describe("getOrders", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("maps platform-specific tab to correct endpoint", async () => {
    mockClientGet.mockResolvedValue({
      data: {
        success: true,
        data: [],
        count: 0,
      },
    });

    await getOrders({ status: "READY_TO_SHIP", platform: "shopee" });

    expect(mockClientGet).toHaveBeenCalledWith("/orders/unprocess", {
      params: {
        page: undefined,
        pageSize: undefined,
        platform: "shopee",
        search: undefined,
        start_date: undefined,
        end_date: undefined,
      },
    });
  });

  it("normalizes today payload fields to table-compatible order fields", async () => {
    mockClientPost.mockResolvedValue({
      data: {
        success: true,
        data: [
          {
            id: 1,
            order_sn: "ORDER-TODAY-1",
            platform: "TIKTOK",
            tracking_no: "TK-123",
            courier: "JNE",
            seller_sku: "SKU-TK",
            product_name: "Product Name",
            variation_name: "Variant",
            quantity: 2,
            synced_at: "2026-01-01T00:00:00Z",
          },
        ],
        count: 1,
      },
    });

    const result = await getOrders({ status: "today" });

    expect(result.orders[0]).toMatchObject({
      order_no: "ORDER-TODAY-1",
      order_sn: "ORDER-TODAY-1",
      tracking_number: "TK-123",
      shipping_carrier: "JNE",
      sku: "SKU-TK",
      qty: 2,
      platform: "tiktok",
    });
  });
});
