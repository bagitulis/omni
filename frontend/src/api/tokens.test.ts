import { describe, it, expect, vi, beforeEach } from "vitest";

const { mockGet, mockPost } = vi.hoisted(() => ({
  mockGet: vi.fn(),
  mockPost: vi.fn(),
}));

vi.mock("@/api/client", () => ({
  default: { get: mockGet, post: mockPost },
}));

import {
  getAllTokenStatus,
  getTokenStatus,
  refreshToken,
  refreshAllTokens,
} from "./tokens";

describe("getAllTokenStatus", () => {
  beforeEach(() => vi.clearAllMocks());

  it("returns all token status on success", async () => {
    const data = {
      shopee: { platform: "shopee", isValid: true, needsRefresh: false },
    };
    mockGet.mockResolvedValue({ success: true, data });
    const result = await getAllTokenStatus();
    expect(result).toEqual(data);
    expect(mockGet).toHaveBeenCalledWith("/tokens/status");
  });

  it("throws when success is false", async () => {
    mockGet.mockResolvedValue({ success: false, error: "Unauthorized" });
    await expect(getAllTokenStatus()).rejects.toThrow("Unauthorized");
  });

  it("throws when data is null even if success is true", async () => {
    mockGet.mockResolvedValue({ success: true, data: null });
    await expect(getAllTokenStatus()).rejects.toThrow(
      "Failed to fetch token status",
    );
  });

  it("throws with default message when no error provided", async () => {
    mockGet.mockResolvedValue({ success: false, data: null });
    await expect(getAllTokenStatus()).rejects.toThrow(
      "Failed to fetch token status",
    );
  });
});

describe("getTokenStatus", () => {
  beforeEach(() => vi.clearAllMocks());

  it("returns token status for a platform", async () => {
    const data = { platform: "shopee", isValid: true, needsRefresh: false };
    mockGet.mockResolvedValue({ success: true, data });
    const result = await getTokenStatus("shopee");
    expect(result).toEqual(data);
    expect(mockGet).toHaveBeenCalledWith("/tokens/status/shopee");
  });

  it("returns data even if success is false (only checks data presence)", async () => {
    const data = { platform: "lazada", isValid: false, needsRefresh: true };
    mockGet.mockResolvedValue({ success: false, data });
    const result = await getTokenStatus("lazada");
    expect(result).toEqual(data);
  });

  it("throws when data is null", async () => {
    mockGet.mockResolvedValue({ success: true, data: null });
    await expect(getTokenStatus("shopee")).rejects.toThrow(
      "Failed to fetch platform token status",
    );
  });

  it("throws with API error message when data is null", async () => {
    mockGet.mockResolvedValue({
      success: false,
      data: null,
      error: "Token not found",
    });
    await expect(getTokenStatus("tiktok")).rejects.toThrow("Token not found");
  });
});

describe("refreshToken", () => {
  beforeEach(() => vi.clearAllMocks());

  it("refreshes token for a platform", async () => {
    const data = { platform: "shopee", isValid: true };
    mockPost.mockResolvedValue({ success: true, data });
    const result = await refreshToken("shopee");
    expect(result).toEqual(data);
    expect(mockPost).toHaveBeenCalledWith("/tokens/refresh/shopee");
  });

  it("throws when success is false", async () => {
    mockPost.mockResolvedValue({ success: false, error: "Refresh failed" });
    await expect(refreshToken("shopee")).rejects.toThrow("Refresh failed");
  });

  it("throws when data is null", async () => {
    mockPost.mockResolvedValue({ success: true, data: null });
    await expect(refreshToken("shopee")).rejects.toThrow(
      "Failed to refresh token",
    );
  });
});

describe("refreshAllTokens", () => {
  beforeEach(() => vi.clearAllMocks());

  it("refreshes all tokens with default force=false", async () => {
    const data = { shopee: { success: true, isValid: true } };
    mockPost.mockResolvedValue({ success: true, data });
    const result = await refreshAllTokens();
    expect(result).toEqual(data);
    expect(mockPost).toHaveBeenCalledWith("/tokens/refresh-all", null, {
      params: { force: false },
    });
  });

  it("refreshes all tokens with force=true", async () => {
    const data = { shopee: { success: true } };
    mockPost.mockResolvedValue({ success: true, data });
    await refreshAllTokens(true);
    expect(mockPost).toHaveBeenCalledWith("/tokens/refresh-all", null, {
      params: { force: true },
    });
  });

  it("throws when success is false", async () => {
    mockPost.mockResolvedValue({ success: false, error: "Refresh all failed" });
    await expect(refreshAllTokens()).rejects.toThrow("Refresh all failed");
  });

  it("throws when data is null", async () => {
    mockPost.mockResolvedValue({ success: true, data: null });
    await expect(refreshAllTokens()).rejects.toThrow(
      "Failed to refresh all tokens",
    );
  });
});
