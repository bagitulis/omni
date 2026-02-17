import { renderHook } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import { usePlatformStatus, derivePlatformStatus } from "./usePlatformStatus";
import type { UnifiedProductRow, Platform } from "@/types/shared";

describe("usePlatformStatus", () => {
  const createMockProduct = (
    overrides: Partial<UnifiedProductRow> = {},
  ): UnifiedProductRow => ({
    id: 1,
    title: "Test Product",
    description: "Test description",
    images: [],
    status: "active",
    skus: [],
    primary_sku: "TEST-SKU",
    primary_price: 100,
    primary_stock: 10,
    platform_summary: {
      shopee: "not_linked",
      tiktok: "not_linked",
      lazada: "not_linked",
    },
    created_at: "2024-01-01T00:00:00Z",
    updated_at: "2024-01-01T00:00:00Z",
    ...overrides,
  });

  describe("derivePlatformStatus", () => {
    it("returns idle state for null product", () => {
      const result = derivePlatformStatus(null, "shopee");

      expect(result).toEqual({
        platform: "shopee",
        linked: false,
        has_update: false,
        sync_state: "idle",
      });
    });

    it("returns idle state for undefined product", () => {
      const result = derivePlatformStatus(undefined, "tiktok");

      expect(result).toEqual({
        platform: "tiktok",
        linked: false,
        has_update: false,
        sync_state: "idle",
      });
    });

    it("returns not_linked state correctly", () => {
      const product = createMockProduct({
        platform_summary: {
          shopee: "not_linked",
          tiktok: "not_linked",
          lazada: "not_linked",
        },
      });

      const result = derivePlatformStatus(product, "shopee");

      expect(result).toEqual({
        platform: "shopee",
        linked: false,
        has_update: false,
        sync_state: "idle",
      });
    });

    it("returns linked state correctly", () => {
      const product = createMockProduct({
        platform_summary: {
          shopee: "linked",
          tiktok: "not_linked",
          lazada: "not_linked",
        },
        skus: [
          {
            id: 1,
            seller_sku: "SKU-001",
            variant_name: "Default",
            price: 100,
            stock: 10,
            platform_links: [
              {
                platform: "shopee",
                platform_product_id: "PROD-123",
                platform_sku_id: "SKU-456",
                sync_status: "synced",
                last_synced_at: "2024-01-01T12:00:00Z",
              },
            ],
          },
        ],
      });

      const result = derivePlatformStatus(product, "shopee");

      expect(result).toEqual({
        platform: "shopee",
        linked: true,
        has_update: false,
        sync_state: "success",
        platform_product_id: "PROD-123",
        platform_sku_id: "SKU-456",
        last_synced_at: "2024-01-01T12:00:00Z",
      });
    });

    it("returns pending state correctly", () => {
      const product = createMockProduct({
        platform_summary: {
          shopee: "pending",
          tiktok: "not_linked",
          lazada: "not_linked",
        },
        skus: [
          {
            id: 1,
            seller_sku: "SKU-001",
            variant_name: "Default",
            price: 100,
            stock: 10,
            platform_links: [
              {
                platform: "shopee",
                sync_status: "pending",
              },
            ],
          },
        ],
      });

      const result = derivePlatformStatus(product, "shopee");

      expect(result).toEqual({
        platform: "shopee",
        linked: false,
        has_update: false,
        sync_state: "syncing",
      });
    });

    it("returns error state with error message", () => {
      const product = createMockProduct({
        platform_summary: {
          shopee: "error",
          tiktok: "not_linked",
          lazada: "not_linked",
        },
        skus: [
          {
            id: 1,
            seller_sku: "SKU-001",
            variant_name: "Default",
            price: 100,
            stock: 10,
            platform_links: [
              {
                platform: "shopee",
                sync_status: "error",
                error_message: "API rate limit exceeded",
              },
            ],
          },
        ],
      });

      const result = derivePlatformStatus(product, "shopee");

      expect(result).toEqual({
        platform: "shopee",
        linked: false,
        has_update: false,
        sync_state: "error",
        error_message: "API rate limit exceeded",
      });
    });

    it("detects has_update for outdated sync status", () => {
      const product = createMockProduct({
        platform_summary: {
          shopee: "linked",
          tiktok: "not_linked",
          lazada: "not_linked",
        },
        skus: [
          {
            id: 1,
            seller_sku: "SKU-001",
            variant_name: "Default",
            price: 100,
            stock: 10,
            platform_links: [
              {
                platform: "shopee",
                platform_product_id: "PROD-123",
                sync_status: "outdated",
              },
            ],
          },
        ],
      });

      const result = derivePlatformStatus(product, "shopee");

      expect(result.has_update).toBe(true);
    });

    it("detects has_update for pending with existing link", () => {
      const product = createMockProduct({
        platform_summary: {
          shopee: "pending",
          tiktok: "not_linked",
          lazada: "not_linked",
        },
        skus: [
          {
            id: 1,
            seller_sku: "SKU-001",
            variant_name: "Default",
            price: 100,
            stock: 10,
            platform_links: [
              {
                platform: "shopee",
                platform_product_id: "PROD-123",
                sync_status: "pending",
              },
            ],
          },
        ],
      });

      const result = derivePlatformStatus(product, "shopee");

      expect(result.has_update).toBe(true);
    });

    it("aggregates metadata from multiple SKUs", () => {
      const product = createMockProduct({
        platform_summary: {
          shopee: "linked",
          tiktok: "not_linked",
          lazada: "not_linked",
        },
        skus: [
          {
            id: 1,
            seller_sku: "SKU-001",
            variant_name: "Variant 1",
            price: 100,
            stock: 10,
            platform_links: [
              {
                platform: "shopee",
                platform_product_id: "PROD-123",
                sync_status: "synced",
              },
            ],
          },
          {
            id: 2,
            seller_sku: "SKU-002",
            variant_name: "Variant 2",
            price: 120,
            stock: 5,
            platform_links: [
              {
                platform: "shopee",
                platform_sku_id: "SKU-456",
                sync_status: "synced",
                last_synced_at: "2024-01-01T12:00:00Z",
              },
            ],
          },
        ],
      });

      const result = derivePlatformStatus(product, "shopee");

      expect(result.platform_product_id).toBe("PROD-123");
      expect(result.platform_sku_id).toBe("SKU-456");
      expect(result.last_synced_at).toBe("2024-01-01T12:00:00Z");
    });

    it("prioritizes error message from first SKU with error", () => {
      const product = createMockProduct({
        platform_summary: {
          shopee: "error",
          tiktok: "not_linked",
          lazada: "not_linked",
        },
        skus: [
          {
            id: 1,
            seller_sku: "SKU-001",
            variant_name: "Variant 1",
            price: 100,
            stock: 10,
            platform_links: [
              {
                platform: "shopee",
                sync_status: "synced",
              },
            ],
          },
          {
            id: 2,
            seller_sku: "SKU-002",
            variant_name: "Variant 2",
            price: 120,
            stock: 5,
            platform_links: [
              {
                platform: "shopee",
                sync_status: "error",
                error_message: "First error message",
              },
            ],
          },
          {
            id: 3,
            seller_sku: "SKU-003",
            variant_name: "Variant 3",
            price: 130,
            stock: 3,
            platform_links: [
              {
                platform: "shopee",
                sync_status: "error",
                error_message: "Second error message",
              },
            ],
          },
        ],
      });

      const result = derivePlatformStatus(product, "shopee");

      expect(result.error_message).toBe("First error message");
    });
  });

  describe("usePlatformStatus hook", () => {
    it("returns status for all platforms", () => {
      const product = createMockProduct({
        platform_summary: {
          shopee: "linked",
          tiktok: "pending",
          lazada: "not_linked",
        },
        skus: [
          {
            id: 1,
            seller_sku: "SKU-001",
            variant_name: "Default",
            price: 100,
            stock: 10,
            platform_links: [
              {
                platform: "shopee",
                platform_product_id: "SHOP-123",
                sync_status: "synced",
              },
              {
                platform: "tiktok",
                sync_status: "pending",
              },
            ],
          },
        ],
      });

      const { result } = renderHook(() => usePlatformStatus(product));

      expect(result.current.shopee.linked).toBe(true);
      expect(result.current.shopee.sync_state).toBe("success");
      expect(result.current.tiktok.linked).toBe(false);
      expect(result.current.tiktok.sync_state).toBe("syncing");
      expect(result.current.lazada.linked).toBe(false);
      expect(result.current.lazada.sync_state).toBe("idle");
    });

    it("memoizes result for same product reference", () => {
      const product = createMockProduct();

      const { result, rerender } = renderHook(() => usePlatformStatus(product));
      const firstResult = result.current;

      rerender();
      const secondResult = result.current;

      expect(firstResult).toBe(secondResult);
    });

    it("updates result when product changes", () => {
      const product1 = createMockProduct({
        platform_summary: {
          shopee: "not_linked",
          tiktok: "not_linked",
          lazada: "not_linked",
        },
      });
      const product2 = createMockProduct({
        platform_summary: {
          shopee: "linked",
          tiktok: "not_linked",
          lazada: "not_linked",
        },
      });

      const { result, rerender } = renderHook(({ p }) => usePlatformStatus(p), {
        initialProps: { p: product1 },
      });

      expect(result.current.shopee.linked).toBe(false);

      rerender({ p: product2 });

      expect(result.current.shopee.linked).toBe(true);
    });

    it("handles null product gracefully", () => {
      const { result } = renderHook(() => usePlatformStatus(null));

      const platforms: Platform[] = ["shopee", "tiktok", "lazada"];
      platforms.forEach((platform) => {
        expect(result.current[platform]).toEqual({
          platform,
          linked: false,
          has_update: false,
          sync_state: "idle",
        });
      });
    });

    it("handles undefined product gracefully", () => {
      const { result } = renderHook(() => usePlatformStatus(undefined));

      const platforms: Platform[] = ["shopee", "tiktok", "lazada"];
      platforms.forEach((platform) => {
        expect(result.current[platform]).toEqual({
          platform,
          linked: false,
          has_update: false,
          sync_state: "idle",
        });
      });
    });
  });
});
