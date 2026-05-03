import { beforeEach, describe, expect, it, vi } from "vitest";
import {
  checkPlatformStatus,
  getAvailableColumns,
  getInventoryConfig,
  getSelectedColumns,
  saveSelectedColumns,
  updateStock,
  updateInventoryRecord,
  updatePriceBatch,
  updateStockBatch,
} from "./inventory";

const { mockGet, mockPost, mockPut, mockClientGet } = vi.hoisted(() => ({
  mockGet: vi.fn(),
  mockPost: vi.fn(),
  mockPut: vi.fn(),
  mockClientGet: vi.fn(),
}));

vi.mock("./client", () => ({
  default: {
    get: mockGet,
    post: mockPost,
    put: mockPut,
    client: {
      get: mockClientGet,
    },
  },
}));

describe("inventory api contract adapters", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("normalizes legacy update-price-batch array response", async () => {
    mockPost.mockResolvedValue({
      success: true,
      data: [
        { sku: "SKU-1", success: true, platforms: {}, errors: [], skipped: [] },
        {
          sku: "SKU-2",
          success: false,
          platforms: {},
          errors: [],
          skipped: [],
        },
      ],
    });

    const result = await updatePriceBatch([
      { sku: "SKU-1", price: 1000 },
      { sku: "SKU-2", price: 2000 },
    ]);

    expect(result).toMatchObject({
      total: 2,
      successful: 1,
      failed: 1,
      skipped: 0,
    });
  });

  it("normalizes structured update-price-batch response", async () => {
    mockPost.mockResolvedValue({
      success: true,
      data: {
        total: 3,
        success: 2,
        failed: 1,
        results: [
          {
            sku: "SKU-1",
            success: true,
            platforms: {},
            errors: [],
            skipped: [],
          },
          {
            sku: "SKU-2",
            success: true,
            platforms: {},
            errors: [],
            skipped: [],
          },
          {
            sku: "SKU-3",
            success: false,
            platforms: {},
            errors: [],
            skipped: [],
          },
        ],
      },
    });

    const result = await updatePriceBatch([{ sku: "SKU-1", price: 1000 }]);

    expect(result).toMatchObject({
      total: 3,
      successful: 2,
      failed: 1,
      skipped: 0,
    });
  });

  it("supports selected_columns and data response shapes", async () => {
    mockClientGet.mockResolvedValueOnce({
      data: {
        success: true,
        selected_columns: ["SKU", "Stock"],
      },
    });

    await expect(getSelectedColumns()).resolves.toEqual(["SKU", "Stock"]);

    mockClientGet.mockResolvedValueOnce({
      data: {
        success: true,
        data: ["SKU", "Price"],
      },
    });

    await expect(getSelectedColumns()).resolves.toEqual(["SKU", "Price"]);
  });

  it("supports available columns from legacy and modular response shapes", async () => {
    mockClientGet.mockResolvedValueOnce({
      data: {
        success: true,
        columns: [{ name: "SKU" }, { name: "Stock" }],
      },
    });

    await expect(getAvailableColumns()).resolves.toEqual(["SKU", "Stock"]);

    mockClientGet.mockResolvedValueOnce({
      data: {
        success: true,
        data: [
          { key: "sku", label: "SKU" },
          { key: "price", label: "Price" },
        ],
      },
    });

    await expect(getAvailableColumns()).resolves.toEqual(["SKU", "Price"]);
  });

  it("normalizes inventory config when selected_columns is an array", async () => {
    mockGet.mockResolvedValueOnce({
      success: true,
      data: {
        spreadsheet_id: "sheet-1",
        sheet_name: "Inventory",
        selected_columns: ["SKU", "Stock"],
        key_column: "SKU",
      },
    });

    await expect(getInventoryConfig()).resolves.toMatchObject({
      spreadsheet_id: "sheet-1",
      sheet_name: "Inventory",
      selected_columns: '["SKU","Stock"]',
      key_column: "SKU",
    });
  });

  it("maps key_column_name fallback when key_column is missing", async () => {
    mockGet.mockResolvedValueOnce({
      success: true,
      data: {
        selected_columns: "SKU,Stock",
        key_column_name: "SKU",
      },
    });

    await expect(getInventoryConfig()).resolves.toMatchObject({
      selected_columns: "SKU,Stock",
      key_column: "SKU",
    });
  });

  it("sends backward and forward compatible selected-columns payload", async () => {
    mockPost.mockResolvedValue({ success: true });

    await saveSelectedColumns(["SKU", "Stock"]);

    expect(mockPost).toHaveBeenCalledWith("/inventory/columns/selected", {
      selected_columns: ["SKU", "Stock"],
      columns: ["SKU", "Stock"],
    });
  });

  it("sends legacy and modular stock-batch fields", async () => {
    mockPost.mockResolvedValue({ success: true });

    await updateStockBatch(["SKU-1", "SKU-2"], ["shopee"]);

    expect(mockPost).toHaveBeenCalledWith("/inventory/update-stock-batch", {
      skus: ["SKU-1", "SKU-2"],
      platforms: ["shopee"],
      items: [
        { sku: "SKU-1", platforms: ["shopee"] },
        { sku: "SKU-2", platforms: ["shopee"] },
      ],
    });
  });

  it("sends per-item stock payload for stock sync modal flow", async () => {
    mockPost.mockResolvedValue({ success: true });

    await updateStockBatch([
      { sku: "SKU-1", stock: 7, platforms: ["shopee", "tiktok"] },
      { sku: "SKU-2", stock: 5, platforms: ["lazada"] },
    ]);

    expect(mockPost).toHaveBeenCalledWith("/inventory/update-stock-batch", {
      skus: ["SKU-1", "SKU-2"],
      platforms: undefined,
      items: [
        { sku: "SKU-1", stock: 7, platforms: ["shopee", "tiktok"] },
        { sku: "SKU-2", stock: 5, platforms: ["lazada"] },
      ],
    });
  });

  it("includes stock in single update payload when provided", async () => {
    mockPost.mockResolvedValue({ success: true });

    await updateStock("SKU-1", ["shopee"], 12);

    expect(mockPost).toHaveBeenCalledWith("/inventory/update-stock", {
      sku: "SKU-1",
      platforms: ["shopee"],
      stock: 12,
    });
  });

  it("throws raw backend error when stock batch endpoint returns success false", async () => {
    mockPost.mockResolvedValue({
      success: false,
      error: "shopee API error [code=E1001]: Invalid access token",
    });

    await expect(updateStockBatch(["SKU-1"], ["shopee"])).rejects.toThrow(
      "shopee API error [code=E1001]: Invalid access token",
    );
  });

  it("throws when single stock sync result reports platform failure", async () => {
    mockPost.mockResolvedValue({
      success: true,
      data: {
        success: false,
        platforms: {
          shopee: {
            success: false,
            error: "shopee API error [code=E1001]: Invalid access token",
          },
        },
      },
    });

    await expect(updateStock("SKU-1", ["shopee"], 12)).rejects.toThrow(
      "shopee API error [code=E1001]: Invalid access token",
    );
  });

  it("returns structured result when batch stock sync has failed sku results", async () => {
    mockPost.mockResolvedValue({
      success: true,
      data: [
        {
          sku: "SKU-1",
          success: true,
          platforms: {
            shopee: { success: true },
          },
        },
        {
          sku: "SKU-2",
          success: false,
          platforms: {
            shopee: {
              success: false,
              error: "shopee API error [code=E2002]: Model not found",
            },
          },
        },
      ],
    });

    const result = await updateStockBatch([
      { sku: "SKU-1", stock: 7, platforms: ["shopee"] },
      { sku: "SKU-2", stock: 5, platforms: ["shopee"] },
    ]);
    expect(result.total).toBe(2);
    expect(result.succeeded).toBe(1);
    expect(result.failed).toBe(1);
    expect(result.results[1].platforms?.shopee?.error).toBe(
      "shopee API error [code=E2002]: Model not found",
    );
  });

  it("returns structured result when batch stock sync fails from structured data payload", async () => {
    mockPost.mockResolvedValue({
      success: true,
      data: {
        data: [
          {
            sku: "SKU-3",
            success: false,
            platforms: {
              lazada: {
                success: false,
                error: "lazada API error [code=1000]: Invalid seller sku",
              },
            },
          },
        ],
      },
    });

    const result = await updateStockBatch([
      { sku: "SKU-3", stock: 2, platforms: ["lazada"] },
    ]);
    expect(result.total).toBe(1);
    expect(result.succeeded).toBe(0);
    expect(result.failed).toBe(1);
    expect(result.results[0].platforms?.lazada?.error).toBe(
      "lazada API error [code=1000]: Invalid seller sku",
    );
  });

  it("checks platform status using selected skus", async () => {
    mockPost.mockResolvedValue({
      success: true,
      data: {
        results: [{ sku: "SKU-1", shopee: true, lazada: false, tiktok: false }],
      },
    });

    const result = await checkPlatformStatus(["SKU-1"]);

    expect(result).toEqual([
      { sku: "SKU-1", shopee: true, lazada: false, tiktok: false },
    ]);
    expect(mockPost).toHaveBeenCalledWith("/inventory/batch-check-sku", {
      skus: ["SKU-1"],
    });
  });

  it("updates inventory record through key-value route", async () => {
    mockPut.mockResolvedValue({
      success: true,
      data: {
        id: "1",
        key_value: "SKU-1",
        key_column_name: "SKU",
        data: { SKU: "SKU-1", Stock: 10 },
        created_at: "2026-01-01T00:00:00Z",
        updated_at: "2026-01-01T00:00:00Z",
      },
    });

    await updateInventoryRecord("SKU-1", { SKU: "SKU-1", Stock: 10 });

    expect(mockPut).toHaveBeenCalledWith("/inventory/SKU-1", {
      SKU: "SKU-1",
      Stock: 10,
    });
  });
});
