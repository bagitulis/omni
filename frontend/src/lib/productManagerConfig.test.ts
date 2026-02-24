import { describe, it, expect } from "vitest";
import {
  getPlatformConfig,
  platformConfigs,
  type PlatformConfig,
} from "./productManagerConfig";

describe("platformConfigs", () => {
  it("has exactly three platforms: lazada, shopee, tiktok", () => {
    expect(Object.keys(platformConfigs)).toEqual([
      "lazada",
      "shopee",
      "tiktok",
    ]);
  });

  it("each platform config has required fields", () => {
    const requiredFields: (keyof PlatformConfig)[] = [
      "platform",
      "columnFields",
      "columnLabels",
      "availableStatuses",
      "filterableFields",
      "searchFields",
      "apiEndpoint",
    ];
    for (const config of Object.values(platformConfigs)) {
      for (const field of requiredFields) {
        expect(config).toHaveProperty(field);
      }
    }
  });

  it("all platforms use /product as apiEndpoint", () => {
    for (const config of Object.values(platformConfigs)) {
      expect(config.apiEndpoint).toBe("/product");
    }
  });
});

describe("lazada config", () => {
  const config = platformConfigs["lazada"];

  it("has platform = lazada", () => {
    expect(config.platform).toBe("lazada");
  });

  it("includes expected columnFields", () => {
    expect(config.columnFields).toContain("item_id");
    expect(config.columnFields).toContain("sku_id");
    expect(config.columnFields).toContain("sku_name");
    expect(config.columnFields).toContain("price");
    expect(config.columnFields).toContain("quantity");
    expect(config.columnFields).toContain("status");
  });

  it("columnLabels map field to human-readable string", () => {
    expect(config.columnLabels["item_id"]).toBe("Item ID");
    expect(config.columnLabels["item_name"]).toBe("Product Name");
    expect(config.columnLabels["quantity"]).toBe("Stock");
  });

  it("availableStatuses are ACTIVE, INACTIVE, DELISTED, UNKNOWN", () => {
    expect(config.availableStatuses).toEqual([
      "ACTIVE",
      "INACTIVE",
      "DELISTED",
      "UNKNOWN",
    ]);
  });

  it("searchFields are a subset of columnFields", () => {
    for (const field of config.searchFields) {
      expect(config.columnFields).toContain(field);
    }
  });
});

describe("shopee config", () => {
  const config = platformConfigs["shopee"];

  it("has platform = shopee", () => {
    expect(config.platform).toBe("shopee");
  });

  it("uses model_id and sku (not sku_id) as identifiers", () => {
    expect(config.columnFields).toContain("model_id");
    expect(config.columnFields).toContain("sku");
    expect(config.columnFields).not.toContain("sku_id");
  });

  it("uses stock (not quantity) for inventory", () => {
    expect(config.columnFields).toContain("stock");
    expect(config.columnFields).not.toContain("quantity");
  });

  it("availableStatuses include NORMAL, BANNED, DELETED, UNLIST", () => {
    expect(config.availableStatuses).toEqual([
      "NORMAL",
      "BANNED",
      "DELETED",
      "UNLIST",
    ]);
  });

  it("price label is Current Price", () => {
    expect(config.columnLabels["price"]).toBe("Current Price");
  });
});

describe("tiktok config", () => {
  const config = platformConfigs["tiktok"];

  it("has platform = tiktok", () => {
    expect(config.platform).toBe("tiktok");
  });

  it("uses product_id as primary identifier", () => {
    expect(config.columnFields).toContain("product_id");
  });

  it("includes seller_sku field", () => {
    expect(config.columnFields).toContain("seller_sku");
  });

  it("availableStatuses include DRAFT and ACTIVATE", () => {
    expect(config.availableStatuses).toContain("DRAFT");
    expect(config.availableStatuses).toContain("ACTIVATE");
  });

  it("has six available statuses", () => {
    expect(config.availableStatuses).toHaveLength(6);
  });
});

describe("getPlatformConfig", () => {
  it("returns lazada config for 'lazada'", () => {
    const config = getPlatformConfig("lazada");
    expect(config.platform).toBe("lazada");
  });

  it("returns shopee config for 'shopee'", () => {
    const config = getPlatformConfig("shopee");
    expect(config.platform).toBe("shopee");
  });

  it("returns tiktok config for 'tiktok'", () => {
    const config = getPlatformConfig("tiktok");
    expect(config.platform).toBe("tiktok");
  });

  it("falls back to lazada config for unknown platform", () => {
    const config = getPlatformConfig("unknown");
    expect(config.platform).toBe("lazada");
  });

  it("falls back to lazada config for empty string", () => {
    const config = getPlatformConfig("");
    expect(config.platform).toBe("lazada");
  });

  it("returns a config that has all required fields", () => {
    const config = getPlatformConfig("shopee");
    expect(config.columnFields.length).toBeGreaterThan(0);
    expect(config.searchFields.length).toBeGreaterThan(0);
    expect(Object.keys(config.columnLabels).length).toBeGreaterThan(0);
  });
});
