import { beforeEach, describe, expect, it, vi } from "vitest";
import { batchTiktokMpq, getSettings } from "./wholesale";

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

describe("wholesale api adapters", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("reads settings from standard data payload", async () => {
    mockGet.mockResolvedValue({
      success: true,
      data: {
        admin_fee: 1500,
        min_order_1: 2,
        max_order_1: 3,
        max_order_tier_3: 1000,
      },
    });

    await expect(getSettings()).resolves.toEqual({
      admin_fee: 1500,
      min_order_1: 2,
      max_order_1: 3,
      max_order_tier_3: 1000,
    });
  });

  it("normalizes backend settings envelope with legacy qty fields", async () => {
    mockGet.mockResolvedValue({
      success: true,
      settings: {
        min_qty_1: 5,
        min_qty_2: 10,
        min_qty_3: 20,
      },
    });

    await expect(getSettings()).resolves.toEqual({
      admin_fee: 0,
      min_order_1: 5,
      max_order_1: 9,
      max_order_tier_3: 20,
    });
  });

  it("sends products payload for TikTok MPQ route", async () => {
    mockPost.mockResolvedValue({ success: true, data: { total: 1 } });

    await batchTiktokMpq([{ sku: "SKU-TK-1", price: 50000 }], 2);

    expect(mockPost).toHaveBeenCalledWith("/wholesale/tiktok/batch-mpq", {
      products: [{ product_id: "SKU-TK-1", sku_id: "SKU-TK-1", mpq: 2 }],
      items: [{ sku: "SKU-TK-1", price: 50000 }],
      mpq: 2,
    });
  });

  it("preserves backend raw error message for TikTok MPQ failures", async () => {
    mockPost.mockResolvedValue({
      success: false,
      error: "tiktok API error [code=10003]: invalid product_id",
    });

    await expect(batchTiktokMpq([{ sku: "SKU-TK-2" }], 2)).rejects.toThrow(
      "tiktok API error [code=10003]: invalid product_id",
    );
  });
});
