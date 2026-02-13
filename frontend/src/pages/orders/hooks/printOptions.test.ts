import { describe, expect, it } from "vitest";
import type { Order, OrderListResponse } from "@/types/order";
import {
  buildBulkPrintOptions,
  shouldPromptTikTokPackingSlip,
} from "./printOptions";

function createOrder(overrides: Partial<Order>): Order {
  return {
    id: "1",
    order_sn: "ORDER-1",
    order_no: "ORDER-1",
    order_status: "UNPAID",
    status: "UNPAID",
    platform: "shopee",
    category: "default",
    buyer_username: "buyer",
    total_amount: 10000,
    currency: "IDR",
    payment_method: "cod",
    shipping_carrier: "JNE",
    ship_by_date: 0,
    sku: "SKU-1",
    product_name: "Product 1",
    variation_name: "Default",
    qty: 1,
    price: 10000,
    product_image: "",
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
    ...overrides,
  };
}

function createData(orders: Order[]): OrderListResponse {
  return {
    orders,
    total: orders.length,
    page: 1,
    page_size: 10,
  };
}

describe("shouldPromptTikTokPackingSlip", () => {
  it("returns true for explicit TikTok platform", () => {
    expect(shouldPromptTikTokPackingSlip("tiktok", undefined, [])).toBe(true);
  });

  it("returns false for non-all non-TikTok platform", () => {
    expect(shouldPromptTikTokPackingSlip("shopee", undefined, [])).toBe(false);
  });

  it("returns true for all platform when selected order is TikTok", () => {
    const data = createData([
      createOrder({ order_sn: "SHP-1", order_no: "SHP-1", platform: "shopee" }),
      createOrder({ order_sn: "TT-1", order_no: "TT-1", platform: "tiktok" }),
    ]);

    expect(shouldPromptTikTokPackingSlip("all", data, ["TT-1"])).toBe(true);
  });

  it("returns false for all platform when selected orders have no TikTok", () => {
    const data = createData([
      createOrder({ order_sn: "SHP-1", order_no: "SHP-1", platform: "shopee" }),
      createOrder({ order_sn: "LZD-1", order_no: "LZD-1", platform: "lazada" }),
    ]);

    expect(shouldPromptTikTokPackingSlip("all", data, ["SHP-1"])).toBe(false);
  });
});

describe("buildBulkPrintOptions", () => {
  it("normalizes and includes platform and include_products", () => {
    expect(buildBulkPrintOptions(" TikTok ", true)).toEqual({
      platform: "tiktok",
      include_products: true,
    });
  });

  it("omits platform when selected platform is all", () => {
    expect(buildBulkPrintOptions("all", false)).toEqual({
      include_products: false,
    });
  });

  it("omits include_products when not provided", () => {
    expect(buildBulkPrintOptions("shopee")).toEqual({
      platform: "shopee",
    });
  });
});
