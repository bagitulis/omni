import { describe, it, expect, vi, beforeEach } from "vitest";
import {
  registerSheet,
  getSheets,
  getSheet,
  lockSheet,
  unlockSheet,
  deleteSheet,
  exportInventoryToSheet,
  importInventoryFromSheet,
  checkSyncStatus,
} from "./sheetRegistry";

const { mockGet, mockPost, mockDelete } = vi.hoisted(() => ({
  mockGet: vi.fn(),
  mockPost: vi.fn(),
  mockDelete: vi.fn(),
}));

vi.mock("./client", () => ({
  default: {
    get: mockGet,
    post: mockPost,
    put: vi.fn(),
    patch: vi.fn(),
    delete: mockDelete,
    client: { get: vi.fn(), post: vi.fn() },
  },
}));

beforeEach(() => vi.clearAllMocks());

const mockSheet = {
  id: "sheet-123",
  spreadsheet_id: "1abc",
  spreadsheet_name: "Test Sheet",
  spreadsheet_url: "https://docs.google.com/spreadsheets/d/1abc",
  sheets: [],
  registered_by: "test-user",
  registered_at: "2026-01-01T00:00:00Z",
  is_locked: false,
  purpose: "inventory",
  sync_settings: { auto_sync: false, sync_interval: 60 },
  is_active: true,
  created_at: "2026-01-01T00:00:00Z",
  updated_at: "2026-01-01T00:00:00Z",
};

describe("registerSheet", () => {
  it("returns spreadsheet data on success", async () => {
    mockPost.mockResolvedValue({ success: true, data: mockSheet });

    const result = await registerSheet(
      "https://docs.google.com/spreadsheets/d/1abc",
      "inventory",
    );

    expect(result).toEqual(mockSheet);
  });

  it("calls POST /google/registry/register", async () => {
    mockPost.mockResolvedValue({ success: true, data: mockSheet });

    await registerSheet(
      "https://docs.google.com/spreadsheets/d/1abc",
      "inventory",
    );

    expect(mockPost).toHaveBeenCalledWith("/google/registry/register", {
      url: "https://docs.google.com/spreadsheets/d/1abc",
      purpose: "inventory",
    });
  });

  it("throws when response.success is false", async () => {
    mockPost.mockResolvedValue({ success: false, error: "Invalid URL" });

    await expect(registerSheet("bad-url", "inventory")).rejects.toThrow(
      "Invalid URL",
    );
  });

  it("throws when data is null", async () => {
    mockPost.mockResolvedValue({ success: true, data: null });

    await expect(
      registerSheet("https://docs.google.com/spreadsheets/d/1abc", "inventory"),
    ).rejects.toThrow("Registration failed");
  });
});

describe("getSheets", () => {
  it("returns array of sheets", async () => {
    mockGet.mockResolvedValue({ success: true, data: [mockSheet] });

    const result = await getSheets();

    expect(result).toEqual([mockSheet]);
  });

  it("returns empty array when data is undefined", async () => {
    mockGet.mockResolvedValue({ success: true, data: undefined });

    const result = await getSheets();

    expect(result).toEqual([]);
  });

  it("throws when response.success is false", async () => {
    mockGet.mockResolvedValue({ success: false, error: "Not authorized" });

    await expect(getSheets()).rejects.toThrow("Not authorized");
  });
});

describe("getSheet", () => {
  it("returns a single sheet by id", async () => {
    mockGet.mockResolvedValue({ success: true, data: mockSheet });

    const result = await getSheet("sheet-123");

    expect(result).toEqual(mockSheet);
  });

  it("returns null when data is undefined", async () => {
    mockGet.mockResolvedValue({ success: true, data: undefined });

    const result = await getSheet("sheet-123");

    expect(result).toBeNull();
  });

  it("throws when response.success is false", async () => {
    mockGet.mockResolvedValue({ success: false, error: "Sheet not found" });

    await expect(getSheet("sheet-999")).rejects.toThrow("Sheet not found");
  });
});

describe("lockSheet / unlockSheet / deleteSheet", () => {
  it("lockSheet returns true when response.success is true", async () => {
    mockPost.mockResolvedValue({ success: true });

    const result = await lockSheet("sheet-123");

    expect(result).toBe(true);
    expect(mockPost).toHaveBeenCalledWith("/google/registry/sheet-123/lock");
  });

  it("lockSheet returns false when response.success is false", async () => {
    mockPost.mockResolvedValue({ success: false });

    const result = await lockSheet("sheet-123");

    expect(result).toBe(false);
  });

  it("unlockSheet returns true when response.success is true", async () => {
    mockPost.mockResolvedValue({ success: true });

    const result = await unlockSheet("sheet-123");

    expect(result).toBe(true);
    expect(mockPost).toHaveBeenCalledWith("/google/registry/sheet-123/unlock");
  });

  it("deleteSheet returns true when response.success is true", async () => {
    mockDelete.mockResolvedValue({ success: true });

    const result = await deleteSheet("sheet-123");

    expect(result).toBe(true);
    expect(mockDelete).toHaveBeenCalledWith("/google/registry/sheet-123");
  });

  it("deleteSheet returns false when response.success is false", async () => {
    mockDelete.mockResolvedValue({ success: false });

    const result = await deleteSheet("sheet-123");

    expect(result).toBe(false);
  });
});

describe("exportInventoryToSheet", () => {
  it("returns stats on success", async () => {
    const stats = { exported: 50 };
    mockPost.mockResolvedValue({ success: true, data: { stats } });

    const result = await exportInventoryToSheet("sheet-123", [
      { sku: "S1", price: 100 },
    ]);

    expect(result).toEqual({ stats });
  });

  it("throws when response.success is false", async () => {
    mockPost.mockResolvedValue({ success: false, error: "Export failed" });

    await expect(exportInventoryToSheet("sheet-123", [])).rejects.toThrow(
      "Export failed",
    );
  });
});

describe("importInventoryFromSheet", () => {
  it("returns imported data on success", async () => {
    const data = [{ sku: "S1" }, { sku: "S2" }];
    mockPost.mockResolvedValue({ success: true, data });

    const result = await importInventoryFromSheet("sheet-123", "Sheet1");

    expect(result).toEqual(data);
  });

  it("throws when response.success is false", async () => {
    mockPost.mockResolvedValue({ success: false, error: "Import failed" });

    await expect(importInventoryFromSheet("sheet-123")).rejects.toThrow(
      "Import failed",
    );
  });
});

describe("checkSyncStatus", () => {
  it("returns sync status on success", async () => {
    const status = { last_sync: "2026-01-01", status: "ok", changes: 5 };
    mockPost.mockResolvedValue({ success: true, data: status });

    const result = await checkSyncStatus("sheet-123");

    expect(result).toEqual(status);
  });

  it("throws when response.success is false", async () => {
    mockPost.mockResolvedValue({
      success: false,
      error: "Status check failed",
    });

    await expect(checkSyncStatus("sheet-123")).rejects.toThrow(
      "Status check failed",
    );
  });
});
