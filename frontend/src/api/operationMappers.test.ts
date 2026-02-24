import { describe, it, expect } from "vitest";
import {
  mapShopeeOperation,
  mapLazadaOperation,
  mapTikTokOperation,
  mapSheetsOperation,
  mapOrderExport,
  isDirectOrderExport,
  isDirectSheetsOperation,
  getPlatformOperationMapping,
} from "./operationMappers";

describe("mapShopeeOperation", () => {
  it("maps known operations to correct endpoints", () => {
    expect(mapShopeeOperation("update_stock")).toEqual({
      endpoint: "/inventory/update-stock",
      method: "POST",
    });
    expect(mapShopeeOperation("update_price")).toEqual({
      endpoint: "/inventory/update-price",
      method: "POST",
    });
    expect(mapShopeeOperation("export_orders")).toEqual({
      endpoint: "/shopee/orders/export",
      method: "POST",
    });
    expect(mapShopeeOperation("get_token")).toEqual({
      endpoint: "/platform-auth/initiate/shopee",
      method: "POST",
    });
    expect(mapShopeeOperation("refresh_token")).toEqual({
      endpoint: "/tokens/refresh/shopee",
      method: "POST",
    });
    expect(mapShopeeOperation("update_code")).toEqual({
      endpoint: "/shopee/operations/update-code",
      method: "POST",
    });
  });

  it("falls back to /execute for unknown operations", () => {
    expect(mapShopeeOperation("unknown_op")).toEqual({
      endpoint: "/execute",
      method: "POST",
    });
  });
});

describe("mapLazadaOperation", () => {
  it("maps known operations to correct endpoints", () => {
    expect(mapLazadaOperation("update_stock").endpoint).toBe(
      "/inventory/update-stock",
    );
    expect(mapLazadaOperation("update_price").endpoint).toBe(
      "/inventory/update-price",
    );
    expect(mapLazadaOperation("get_token").endpoint).toBe(
      "/platform-auth/initiate/lazada",
    );
    expect(mapLazadaOperation("refresh_token").endpoint).toBe(
      "/tokens/refresh/lazada",
    );
  });

  it("falls back to /execute for unknown operations", () => {
    expect(mapLazadaOperation("unknown_op").endpoint).toBe("/execute");
  });
});

describe("mapTikTokOperation", () => {
  it("maps known operations to correct endpoints", () => {
    expect(mapTikTokOperation("update_stock").endpoint).toBe(
      "/inventory/update-stock",
    );
    expect(mapTikTokOperation("update_price").endpoint).toBe(
      "/inventory/update-price",
    );
    expect(mapTikTokOperation("export_orders").endpoint).toBe(
      "/tiktok/orders/export",
    );
    expect(mapTikTokOperation("get_token").endpoint).toBe(
      "/platform-auth/initiate/tiktok",
    );
  });

  it("falls back to /execute for unknown operations", () => {
    expect(mapTikTokOperation("unknown_op").endpoint).toBe("/execute");
  });
});

describe("mapSheetsOperation", () => {
  it("maps wallet_to_sheets", () => {
    expect(mapSheetsOperation("wallet_to_sheets")).toBe(
      "/shopee/wallet/export-to-sheets",
    );
  });

  it("maps shipping_fee_to_sheets", () => {
    expect(mapSheetsOperation("shipping_fee_to_sheets")).toBe(
      "/shopee/shipping/export-to-sheets",
    );
  });

  it("falls back to /execute-sheets for unknown operations", () => {
    expect(mapSheetsOperation("unknown_op")).toBe("/execute-sheets");
  });
});

describe("mapOrderExport", () => {
  it("maps shopee unpaid/cancelled to correct endpoints", () => {
    expect(mapOrderExport("shopee", "unpaid")).toBe(
      "/orders/export/shopee/unpaid",
    );
    expect(mapOrderExport("shopee", "cancelled")).toBe(
      "/orders/export/shopee/cancelled",
    );
  });

  it("maps lazada unpaid/cancelled to correct endpoints", () => {
    expect(mapOrderExport("lazada", "unpaid")).toBe(
      "/orders/export/lazada/unpaid",
    );
    expect(mapOrderExport("lazada", "cancelled")).toBe(
      "/orders/export/lazada/cancelled",
    );
  });

  it("maps tiktok unpaid/cancelled to correct endpoints", () => {
    expect(mapOrderExport("tiktok", "unpaid")).toBe(
      "/orders/export/tiktok/unpaid",
    );
    expect(mapOrderExport("tiktok", "cancelled")).toBe(
      "/orders/export/tiktok/cancelled",
    );
  });

  it("normalizes orderType to lowercase", () => {
    expect(mapOrderExport("shopee", "UNPAID")).toBe(
      "/orders/export/shopee/unpaid",
    );
  });

  it("falls back to /orders/export/by-platform for unknown", () => {
    expect(mapOrderExport("unknown", "type")).toBe(
      "/orders/export/by-platform",
    );
  });
});

describe("isDirectOrderExport", () => {
  it("returns true for known direct export keys", () => {
    expect(isDirectOrderExport("shopee", "unpaid")).toBe(true);
    expect(isDirectOrderExport("shopee", "cancelled")).toBe(true);
    expect(isDirectOrderExport("lazada", "unpaid")).toBe(true);
    expect(isDirectOrderExport("lazada", "cancelled")).toBe(true);
    expect(isDirectOrderExport("tiktok", "unpaid")).toBe(true);
    expect(isDirectOrderExport("tiktok", "cancelled")).toBe(true);
  });

  it("returns false for unknown combinations", () => {
    expect(isDirectOrderExport("shopee", "pending")).toBe(false);
    expect(isDirectOrderExport("unknown", "unpaid")).toBe(false);
  });

  it("normalizes orderType to lowercase", () => {
    expect(isDirectOrderExport("shopee", "UNPAID")).toBe(true);
  });
});

describe("isDirectSheetsOperation", () => {
  it("returns true for known direct operations", () => {
    expect(isDirectSheetsOperation("wallet_to_sheets")).toBe(true);
    expect(isDirectSheetsOperation("shipping_fee_to_sheets")).toBe(true);
  });

  it("returns false for unknown operations", () => {
    expect(isDirectSheetsOperation("unknown_op")).toBe(false);
  });
});

describe("getPlatformOperationMapping", () => {
  it("returns EndpointMapping for shopee_operation", () => {
    const result = getPlatformOperationMapping("shopee_operation", {
      operation: "update_stock",
    });
    expect(result).toEqual({
      endpoint: "/inventory/update-stock",
      method: "POST",
    });
  });

  it("returns EndpointMapping for lazada_operation", () => {
    const result = getPlatformOperationMapping("lazada_operation", {
      operation: "refresh_token",
    });
    expect(result).not.toBeNull();
    expect(result?.endpoint).toBe("/tokens/refresh/lazada");
  });

  it("returns EndpointMapping for tiktok_operation", () => {
    const result = getPlatformOperationMapping("tiktok_operation", {
      operation: "export_orders",
    });
    expect(result?.endpoint).toBe("/tiktok/orders/export");
  });

  it("returns null for unknown operation type", () => {
    expect(getPlatformOperationMapping("unknown_platform", {})).toBeNull();
  });
});
