import { describe, it, expect, vi, beforeEach } from "vitest";
import {
  getSavedLinks,
  getDetailedSettings,
  validateLink,
  saveLinks,
  updateSettings,
} from "./googleSheets";

const { mockGet, mockPost } = vi.hoisted(() => ({
  mockGet: vi.fn(),
  mockPost: vi.fn(),
}));

vi.mock("./client", () => ({
  default: {
    get: mockGet,
    post: mockPost,
  },
}));

describe("getSavedLinks", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("maps inventory/wallet/shipping/order fields to _url variants", async () => {
    mockGet.mockResolvedValue({
      success: true,
      data: {
        inventory: "https://sheets.google.com/inv",
        wallet: "https://sheets.google.com/wallet",
        shipping: "https://sheets.google.com/shipping",
        order: "https://sheets.google.com/order",
      },
    });

    const result = await getSavedLinks();

    expect(result).toEqual({
      inventory_url: "https://sheets.google.com/inv",
      wallet_url: "https://sheets.google.com/wallet",
      shipping_url: "https://sheets.google.com/shipping",
      order_url: "https://sheets.google.com/order",
    });
  });

  it("defaults missing fields to empty string", async () => {
    mockGet.mockResolvedValue({
      success: true,
      data: { inventory: "https://sheets.google.com/inv" },
    });

    const result = await getSavedLinks();

    expect(result).toEqual({
      inventory_url: "https://sheets.google.com/inv",
      wallet_url: "",
      shipping_url: "",
      order_url: "",
    });
  });

  it("defaults all fields to empty string when data is empty", async () => {
    mockGet.mockResolvedValue({ success: true, data: {} });

    const result = await getSavedLinks();

    expect(result).toEqual({
      inventory_url: "",
      wallet_url: "",
      shipping_url: "",
      order_url: "",
    });
  });

  it("throws when response is not successful", async () => {
    mockGet.mockResolvedValue({
      success: false,
      error: "Unauthorized",
    });

    await expect(getSavedLinks()).rejects.toThrow("Unauthorized");
  });

  it("throws generic message when error field is missing", async () => {
    mockGet.mockResolvedValue({ success: false });

    await expect(getSavedLinks()).rejects.toThrow(
      "Failed to fetch saved links",
    );
  });
});

describe("getDetailedSettings", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("returns data on success", async () => {
    const mockSettings = {
      inventory_url: "https://example.com",
      sync_enabled: true,
    };
    mockGet.mockResolvedValue({ success: true, data: mockSettings });

    const result = await getDetailedSettings();

    expect(result).toEqual(mockSettings);
    expect(mockGet).toHaveBeenCalledWith("/google/settings/detailed");
  });

  it("throws on failure", async () => {
    mockGet.mockResolvedValue({ success: false, error: "Server error" });

    await expect(getDetailedSettings()).rejects.toThrow("Server error");
  });
});

describe("validateLink", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("returns validation result on success", async () => {
    const mockResult = { valid: true, sheet_name: "Inventory" };
    mockPost.mockResolvedValue({ success: true, data: mockResult });

    const result = await validateLink({
      spreadsheet_url: "https://sheets.google.com/test",
      type: "inventory",
    });

    expect(result).toEqual(mockResult);
    expect(mockPost).toHaveBeenCalledWith(
      "/google/settings/validate-link",
      expect.objectContaining({
        spreadsheet_url: "https://sheets.google.com/test",
      }),
    );
  });

  it("throws on validation failure", async () => {
    mockPost.mockResolvedValue({ success: false, error: "Invalid link" });

    await expect(
      validateLink({ spreadsheet_url: "bad-url", type: "inventory" }),
    ).rejects.toThrow("Invalid link");
  });
});

describe("saveLinks", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("resolves without error on success", async () => {
    mockPost.mockResolvedValue({ success: true });

    await expect(
      saveLinks({
        inventory_url: "https://example.com",
        wallet_url: null,
        shipping_url: null,
        order_url: null,
      }),
    ).resolves.toBeUndefined();
    expect(mockPost).toHaveBeenCalledWith(
      "/google/settings/save-links",
      expect.any(Object),
    );
  });

  it("throws on failure", async () => {
    mockPost.mockResolvedValue({ success: false, error: "Save failed" });

    await expect(
      saveLinks({
        inventory_url: "https://example.com",
        wallet_url: null,
        shipping_url: null,
        order_url: null,
      }),
    ).rejects.toThrow("Save failed");
  });
});

describe("updateSettings", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("resolves without error on success", async () => {
    mockPost.mockResolvedValue({ success: true });

    await expect(
      updateSettings({ inventory_spreadsheet_id: "abc123" }),
    ).resolves.toBeUndefined();
    expect(mockPost).toHaveBeenCalledWith(
      "/google/settings/update-detailed",
      expect.any(Object),
    );
  });

  it("throws on failure", async () => {
    mockPost.mockResolvedValue({ success: false, error: "Update failed" });

    await expect(
      updateSettings({ inventory_spreadsheet_id: "" }),
    ).rejects.toThrow("Update failed");
  });
});
