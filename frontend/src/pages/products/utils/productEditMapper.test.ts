import { describe, expect, it } from "vitest";
import type { MasterProduct } from "@/types/product";
import {
  buildPlatformRows,
  mapMasterProductToProductData,
} from "./productEditMapper";

const baseProduct: MasterProduct = {
  id: 99,
  tenant_id: "tenant-1",
  title: "Test Product",
  description: "Test Description",
  images: ["/uploads/initial.webp"],
  status: "draft",
  created_at: "2026-02-18T00:00:00Z",
  updated_at: "2026-02-18T00:00:00Z",
  skus: [],
};

describe("productEditMapper", () => {
  it("maps product SKUs and preserves IDs for edit payload", () => {
    const product: MasterProduct = {
      ...baseProduct,
      skus: [
        {
          id: 7,
          tenant_id: "tenant-1",
          master_product_id: 99,
          seller_sku: "SKU-7",
          variant_name: "Red",
          variant_data: { color: "red" },
          price: 15000,
          stock: 5,
          created_at: "2026-02-18T00:00:00Z",
          updated_at: "2026-02-18T00:00:00Z",
          platform_links: [],
        },
      ],
    };

    const result = mapMasterProductToProductData(product);

    expect(result.skus).toHaveLength(1);
    expect(result.skus?.[0]).toMatchObject({
      key: "7",
      id: 7,
      seller_sku: "SKU-7",
      variant_name: "Red",
      price: 15000,
      stock: 5,
    });
  });

  it("marks platform as synced when SKU presence IDs exist", () => {
    const product: MasterProduct = {
      ...baseProduct,
      skus: [
        {
          id: 8,
          tenant_id: "tenant-1",
          master_product_id: 99,
          seller_sku: "SKU-8",
          variant_name: "Blue",
          variant_data: {},
          price: 17000,
          stock: 4,
          created_at: "2026-02-18T00:00:00Z",
          updated_at: "2026-02-18T00:00:00Z",
          platform_links: [
            {
              id: 1,
              master_product_id: 99,
              master_sku_id: 8,
              platform: "shopee",
              platform_product_id: "",
              platform_sku_id: "30011",
              platform_item_id: 10022,
              sync_status: "failed",
              last_synced_at: "2026-02-18T10:00:00Z",
            },
          ],
        },
      ],
    };

    const rows = buildPlatformRows(product);
    const shopee = rows.find((row) => row.platform === "shopee");
    const lazada = rows.find((row) => row.platform === "lazada");

    expect(shopee?.status).toBe("synced");
    expect(shopee?.last_sync).toBe("2026-02-18T10:00:00Z");
    expect(lazada?.status).toBe("not_synced");
  });
});
