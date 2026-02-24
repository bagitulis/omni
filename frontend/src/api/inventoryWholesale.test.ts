import { describe, it, expect, vi, beforeEach } from "vitest";
import {
  getInventoryWholesaleTiers,
  updateInventoryWholesaleTiers,
  getInventoryWholesaleSettings,
  updateInventoryWholesaleSettings,
  batchUpdateInventoryWholesale,
  batchDeleteInventoryWholesale,
  getInventoryWholesaleInfo,
  getInventoryMpqSettings,
  updateInventoryMpqSettings,
  batchUpdateInventoryMpq,
} from "./inventoryWholesale";

const { mockGet, mockPost, mockPut } = vi.hoisted(() => ({
  mockGet: vi.fn(),
  mockPost: vi.fn(),
  mockPut: vi.fn(),
}));

vi.mock("./client", () => ({
  default: {
    get: mockGet,
    post: mockPost,
    put: mockPut,
  },
}));

describe("getInventoryWholesaleTiers", () => {
  beforeEach(() => vi.clearAllMocks());

  it("returns tiers array on success", async () => {
    const tiers = [{ min_qty: 1, price: 10000 }];
    mockGet.mockResolvedValue({ success: true, data: tiers });

    const result = await getInventoryWholesaleTiers("SKU-001");

    expect(result).toEqual(tiers);
    expect(mockGet).toHaveBeenCalledWith("/inventory/wholesale/SKU-001");
  });

  it("returns empty array when data is null", async () => {
    mockGet.mockResolvedValue({ success: true, data: null });

    const result = await getInventoryWholesaleTiers("SKU-001");

    expect(result).toEqual([]);
  });

  it("throws on failure", async () => {
    mockGet.mockResolvedValue({ success: false, error: "Not found" });

    await expect(getInventoryWholesaleTiers("SKU-001")).rejects.toThrow(
      "Not found",
    );
  });
});

describe("updateInventoryWholesaleTiers", () => {
  beforeEach(() => vi.clearAllMocks());

  it("resolves on success", async () => {
    mockPut.mockResolvedValue({ success: true });

    await expect(
      updateInventoryWholesaleTiers("SKU-001", [{ sku: "SKU-001", min_qty: 1, price: 10000 }]),
    ).resolves.toBeUndefined();
    expect(mockPut).toHaveBeenCalledWith("/inventory/wholesale/SKU-001", {
      tiers: [{ sku: "SKU-001", min_qty: 1, price: 10000 }],
    });
  });

  it("throws on failure", async () => {
    mockPut.mockResolvedValue({ success: false, error: "Update failed" });

    await expect(updateInventoryWholesaleTiers("SKU-001", [])).rejects.toThrow(
      "Update failed",
    );
  });
});

describe("getInventoryWholesaleSettings", () => {
  beforeEach(() => vi.clearAllMocks());

  it("returns settings on success", async () => {
    const settings = { enabled: true, discount_type: "percent" };
    mockGet.mockResolvedValue({ success: true, data: settings });

    const result = await getInventoryWholesaleSettings();

    expect(result).toEqual(settings);
  });

  it("returns null when data is null", async () => {
    mockGet.mockResolvedValue({ success: true, data: null });

    const result = await getInventoryWholesaleSettings();

    expect(result).toBeNull();
  });

  it("throws on failure", async () => {
    mockGet.mockResolvedValue({
      success: false,
      error: "Failed to fetch wholesale settings",
    });

    await expect(getInventoryWholesaleSettings()).rejects.toThrow(
      "Failed to fetch wholesale settings",
    );
  });
});

describe("updateInventoryWholesaleSettings", () => {
  beforeEach(() => vi.clearAllMocks());

  it("returns updated settings on success", async () => {
    const updated = { enabled: true, discount_type: "percent" };
    mockPut.mockResolvedValue({ success: true, data: updated });

    const result = await updateInventoryWholesaleSettings({ enabled: true });

    expect(result).toEqual(updated);
  });

  it("throws when data is null even on success", async () => {
    mockPut.mockResolvedValue({ success: true, data: null });

    await expect(
      updateInventoryWholesaleSettings({ enabled: true }),
    ).rejects.toThrow("no data in response");
  });

  it("throws on failure", async () => {
    mockPut.mockResolvedValue({ success: false, error: "Settings error" });

    await expect(
      updateInventoryWholesaleSettings({ enabled: false }),
    ).rejects.toThrow("Settings error");
  });
});

describe("batchUpdateInventoryWholesale", () => {
  beforeEach(() => vi.clearAllMocks());

  it("returns batch result on success", async () => {
    const batchResult = { updated: 3, failed: 0 };
    mockPost.mockResolvedValue({ success: true, data: batchResult });

    const result = await batchUpdateInventoryWholesale([
      { sku: "SKU-001", tiers: [] },
    ]);

    expect(result).toEqual(batchResult);
    expect(mockPost).toHaveBeenCalledWith(
      "/inventory/wholesale/batch-update",
      expect.objectContaining({ items: expect.any(Array) }),
    );
  });

  it("throws when data is null even on success", async () => {
    mockPost.mockResolvedValue({ success: true, data: null });

    await expect(
      batchUpdateInventoryWholesale([{ sku: "SKU-001", tiers: [] }]),
    ).rejects.toThrow("no data in response");
  });

  it("throws on failure", async () => {
    mockPost.mockResolvedValue({ success: false, error: "Batch error" });

    await expect(batchUpdateInventoryWholesale([])).rejects.toThrow(
      "Batch error",
    );
  });
});

describe("batchDeleteInventoryWholesale", () => {
  beforeEach(() => vi.clearAllMocks());

  it("resolves on success", async () => {
    mockPost.mockResolvedValue({ success: true });

    await expect(
      batchDeleteInventoryWholesale(["SKU-001", "SKU-002"]),
    ).resolves.toBeUndefined();
    expect(mockPost).toHaveBeenCalledWith(
      "/wholesale/shopee/batch-delete-skus",
      { skus: ["SKU-001", "SKU-002"] },
    );
  });

  it("throws on failure", async () => {
    mockPost.mockResolvedValue({ success: false, error: "Delete failed" });

    await expect(batchDeleteInventoryWholesale(["SKU-001"])).rejects.toThrow(
      "Delete failed",
    );
  });
});

describe("getInventoryWholesaleInfo", () => {
  beforeEach(() => vi.clearAllMocks());

  it("returns info on success", async () => {
    const info = { sku: "SKU-001", total_tiers: 3 };
    mockGet.mockResolvedValue({ success: true, data: info });

    const result = await getInventoryWholesaleInfo("SKU-001");

    expect(result).toEqual(info);
    expect(mockGet).toHaveBeenCalledWith("/inventory/wholesale/SKU-001/info");
  });

  it("throws when data is null even on success", async () => {
    mockGet.mockResolvedValue({ success: true, data: null });

    await expect(getInventoryWholesaleInfo("SKU-001")).rejects.toThrow(
      "no data in response",
    );
  });

  it("throws on failure", async () => {
    mockGet.mockResolvedValue({ success: false, error: "Info not found" });

    await expect(getInventoryWholesaleInfo("SKU-001")).rejects.toThrow(
      "Info not found",
    );
  });
});

describe("getInventoryMpqSettings", () => {
  beforeEach(() => vi.clearAllMocks());

  it("returns MPQ settings array on success", async () => {
    const settings = [{ sku: "SKU-001", min_purchase_qty: 2 }];
    mockGet.mockResolvedValue({ success: true, data: settings });

    const result = await getInventoryMpqSettings();

    expect(result).toEqual(settings);
  });

  it("returns empty array when data is null", async () => {
    mockGet.mockResolvedValue({ success: true, data: null });

    const result = await getInventoryMpqSettings();

    expect(result).toEqual([]);
  });

  it("throws on failure", async () => {
    mockGet.mockResolvedValue({ success: false, error: "MPQ fetch failed" });

    await expect(getInventoryMpqSettings()).rejects.toThrow("MPQ fetch failed");
  });
});

describe("updateInventoryMpqSettings", () => {
  beforeEach(() => vi.clearAllMocks());

  it("resolves on success", async () => {
    mockPut.mockResolvedValue({ success: true });

    await expect(
      updateInventoryMpqSettings([{ sku: "SKU-001", min_purchase_qty: 5, enabled: true }]),
    ).resolves.toBeUndefined();
    expect(mockPut).toHaveBeenCalledWith("/inventory/mpq/settings", {
      settings: [{ sku: "SKU-001", min_purchase_qty: 5, enabled: true }],
    });
  });

  it("throws on failure", async () => {
    mockPut.mockResolvedValue({ success: false, error: "MPQ update failed" });

    await expect(updateInventoryMpqSettings([])).rejects.toThrow(
      "MPQ update failed",
    );
  });
});

describe("batchUpdateInventoryMpq", () => {
  beforeEach(() => vi.clearAllMocks());

  it("returns batch result on success", async () => {
    const result = { updated: 2, failed: 0 };
    mockPost.mockResolvedValue({ success: true, data: result });

    const batchResult = await batchUpdateInventoryMpq([
      { sku: "SKU-001", min_purchase_qty: 3 },
    ]);

    expect(batchResult).toEqual(result);
  });

  it("throws when data is null even on success", async () => {
    mockPost.mockResolvedValue({ success: true, data: null });

    await expect(
      batchUpdateInventoryMpq([{ sku: "SKU-001", min_purchase_qty: 3 }]),
    ).rejects.toThrow("no data in response");
  });

  it("throws on failure", async () => {
    mockPost.mockResolvedValue({ success: false, error: "MPQ batch error" });

    await expect(batchUpdateInventoryMpq([])).rejects.toThrow(
      "MPQ batch error",
    );
  });
});
