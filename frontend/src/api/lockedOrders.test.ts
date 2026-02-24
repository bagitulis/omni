import { describe, it, expect, vi, beforeEach } from "vitest";
import { syncLockedToday, getLockedOrders } from "./lockedOrders";

const { mockPost, mockGet } = vi.hoisted(() => ({
  mockPost: vi.fn(),
  mockGet: vi.fn(),
}));

vi.mock("./client", () => ({
  default: {
    get: mockGet,
    post: mockPost,
    put: vi.fn(),
    patch: vi.fn(),
    delete: vi.fn(),
    client: { get: vi.fn(), post: vi.fn() },
  },
}));

beforeEach(() => vi.clearAllMocks());

describe("syncLockedToday", () => {
  it("returns items from response.data.items", async () => {
    const items = [
      { sku: "SKU-1", product_name: "Product 1", qty: 3 },
      { sku: "SKU-2", product_name: "Product 2", qty: 1 },
    ];
    mockPost.mockResolvedValue({ success: true, data: { items } });

    const result = await syncLockedToday();

    expect(result).toEqual(items);
  });

  it("falls back to raw.items when response.data.items is undefined", async () => {
    const items = [{ sku: "SKU-X", product_name: "Fallback Product", qty: 2 }];
    // Simulate backend returning items at top level, not inside data
    mockPost.mockResolvedValue({ success: true, data: undefined, items });

    const result = await syncLockedToday();

    expect(result).toEqual(items);
  });

  it("returns empty array when no items at all", async () => {
    mockPost.mockResolvedValue({ success: true, data: {} });

    const result = await syncLockedToday();

    expect(result).toEqual([]);
  });

  it("passes default days=7 to POST body", async () => {
    mockPost.mockResolvedValue({ success: true, data: { items: [] } });

    await syncLockedToday();

    expect(mockPost).toHaveBeenCalledWith("/orders/locked-today", { days: 7 });
  });

  it("passes custom days value to POST body", async () => {
    mockPost.mockResolvedValue({ success: true, data: { items: [] } });

    await syncLockedToday(14);

    expect(mockPost).toHaveBeenCalledWith("/orders/locked-today", { days: 14 });
  });

  it("throws when response.success is false", async () => {
    mockPost.mockResolvedValue({ success: false, error: "Sync failed" });

    await expect(syncLockedToday()).rejects.toThrow("Sync failed");
  });

  it("throws default message when error field is missing", async () => {
    mockPost.mockResolvedValue({ success: false });

    await expect(syncLockedToday()).rejects.toThrow(
      "Failed to sync locked orders",
    );
  });
});

describe("getLockedOrders", () => {
  it("returns items from response.data.items", async () => {
    const items = [{ sku: "SKU-3", product_name: "Get Product", qty: 5 }];
    mockGet.mockResolvedValue({ success: true, data: { items } });

    const result = await getLockedOrders();

    expect(result).toEqual(items);
  });

  it("falls back to raw.items when response.data.items is undefined", async () => {
    const items = [{ sku: "SKU-Y", product_name: "Raw Item", qty: 1 }];
    mockGet.mockResolvedValue({ success: true, data: undefined, items });

    const result = await getLockedOrders();

    expect(result).toEqual(items);
  });

  it("returns empty array when no items", async () => {
    mockGet.mockResolvedValue({ success: true, data: {} });

    const result = await getLockedOrders();

    expect(result).toEqual([]);
  });

  it("calls GET /orders/locked-today", async () => {
    mockGet.mockResolvedValue({ success: true, data: { items: [] } });

    await getLockedOrders();

    expect(mockGet).toHaveBeenCalledWith("/orders/locked-today");
  });

  it("throws when response.success is false", async () => {
    mockGet.mockResolvedValue({ success: false, error: "Fetch error" });

    await expect(getLockedOrders()).rejects.toThrow("Fetch error");
  });

  it("throws default message when error field is missing", async () => {
    mockGet.mockResolvedValue({ success: false });

    await expect(getLockedOrders()).rejects.toThrow(
      "Failed to fetch locked orders",
    );
  });
});
