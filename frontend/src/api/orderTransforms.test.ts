import { describe, it, expect } from "vitest";
import { transformOrder, computeUniquePlatformCounts } from "./orderTransforms";
import type { Order } from "@/types/order";

describe("orderTransforms", () => {
  describe("transformOrder", () => {
    it("maps order_no and status from backend fields", () => {
      const result = transformOrder({
        order_no: "ORD-001",
        status: "paid",
        platform: "Shopee",
      });

      expect(result.order_no).toBe("ORD-001");
      expect(result.order_sn).toBe("ORD-001");
      expect(result.status).toBe("paid");
      expect(result.order_status).toBe("paid");
      expect(result.platform).toBe("shopee");
    });

    it("falls back to order_sn when order_no is missing", () => {
      const result = transformOrder({ order_sn: "SN-001", status: "pending" });

      expect(result.order_no).toBe("SN-001");
      expect(result.order_sn).toBe("SN-001");
    });

    it("falls back to order_status when status is missing", () => {
      const result = transformOrder({
        order_no: "ORD-001",
        order_status: "shipped",
      });

      expect(result.status).toBe("shipped");
      expect(result.order_status).toBe("shipped");
    });

    it("lowercases the platform", () => {
      const result = transformOrder({
        order_no: "ORD-001",
        platform: "TikTok",
      });
      expect(result.platform).toBe("tiktok");
    });

    it("uses courier fallback for shipping_carrier", () => {
      const result = transformOrder({
        order_no: "ORD-001",
        courier: "JNE",
      });
      expect(result.shipping_carrier).toBe("JNE");
    });

    it("uses tracking_no fallback for tracking_number", () => {
      const result = transformOrder({
        order_no: "ORD-001",
        tracking_no: "TRK-001",
      });
      expect(result.tracking_number).toBe("TRK-001");
    });

    it("uses seller_sku fallback for sku", () => {
      const result = transformOrder({
        order_no: "ORD-001",
        seller_sku: "MY-SKU",
      });
      expect(result.sku).toBe("MY-SKU");
    });

    it("uses quantity fallback for qty", () => {
      const result = transformOrder({ order_no: "ORD-001", quantity: 5 });
      expect(result.qty).toBe(5);
    });

    it("defaults qty to 1 when not provided", () => {
      const result = transformOrder({ order_no: "ORD-001" });
      expect(result.qty).toBe(1);
    });

    it("uses synced_at fallback for created_at", () => {
      const result = transformOrder({
        order_no: "ORD-001",
        synced_at: "2024-01-01T00:00:00Z",
      });
      expect(result.created_at).toBe("2024-01-01T00:00:00Z");
    });

    it("defaults currency to IDR when not provided", () => {
      const result = transformOrder({ order_no: "ORD-001" });
      expect(result.currency).toBe("IDR");
    });

    it("sets id from order_no when id not provided", () => {
      const result = transformOrder({ order_no: "ORD-001" });
      expect(result.id).toBe("ORD-001");
    });

    it("uses provided id when present", () => {
      const result = transformOrder({ id: "custom-id", order_no: "ORD-001" });
      expect(result.id).toBe("custom-id");
    });

    it("defaults total_amount to 0", () => {
      const result = transformOrder({ order_no: "ORD-001" });
      expect(result.total_amount).toBe(0);
    });

    it("preserves existing fields from spread", () => {
      const result = transformOrder({
        order_no: "ORD-001",
        buyer_username: "john_doe",
        product_name: "Test Product",
      });
      expect(result.buyer_username).toBe("john_doe");
      expect(result.product_name).toBe("Test Product");
    });
  });

  describe("computeUniquePlatformCounts", () => {
    it("counts unique orders per platform", () => {
      const orders: Order[] = [
        { order_sn: "O1", platform: "shopee", order_no: "O1" } as Order,
        { order_sn: "O2", platform: "shopee", order_no: "O2" } as Order,
        { order_sn: "O3", platform: "lazada", order_no: "O3" } as Order,
      ];

      const result = computeUniquePlatformCounts(orders);

      expect(result).toEqual({ shopee: 2, lazada: 1 });
    });

    it("lowercases platform names for grouping", () => {
      const orders: Order[] = [
        { order_sn: "O1", platform: "Shopee", order_no: "O1" } as Order,
        { order_sn: "O2", platform: "SHOPEE", order_no: "O2" } as Order,
      ];

      const result = computeUniquePlatformCounts(orders);
      expect(result).toEqual({ shopee: 2 });
    });

    it("deduplicates same order_no within same platform", () => {
      const orders: Order[] = [
        { order_sn: "O1", platform: "shopee", order_no: "O1" } as Order,
        { order_sn: "O1", platform: "shopee", order_no: "O1" } as Order,
      ];

      const result = computeUniquePlatformCounts(orders);
      expect(result.shopee).toBe(1);
    });

    it("uses order_sn when order_no is missing", () => {
      const orders: Order[] = [
        { order_sn: "SN-001", platform: "tiktok", order_no: "" } as Order,
      ];

      const result = computeUniquePlatformCounts(orders);
      expect(result.tiktok).toBe(1);
    });

    it("skips orders without platform", () => {
      const orders: Order[] = [
        { order_sn: "O1", platform: "", order_no: "O1" } as Order,
      ];

      const result = computeUniquePlatformCounts(orders);
      expect(Object.keys(result)).toHaveLength(0);
    });

    it("skips orders without order id", () => {
      const orders: Order[] = [
        { order_sn: "", platform: "shopee", order_no: "" } as Order,
      ];

      const result = computeUniquePlatformCounts(orders);
      expect(Object.keys(result)).toHaveLength(0);
    });

    it("returns empty object for empty input", () => {
      const result = computeUniquePlatformCounts([]);
      expect(result).toEqual({});
    });
  });
});
