import { describe, it, expect, vi, beforeEach } from "vitest";
import { syncProductsToDb } from "./lazadaDb";

const { mockPost } = vi.hoisted(() => ({
  mockPost: vi.fn(),
}));

vi.mock("./client", () => ({
  default: {
    get: vi.fn(),
    post: mockPost,
    put: vi.fn(),
    patch: vi.fn(),
    delete: vi.fn(),
    client: { get: vi.fn(), post: vi.fn() },
  },
}));

vi.mock("@/lib/constants", () => ({
  API_TIMEOUT: { EXTRA_LONG: 120000 },
}));

beforeEach(() => vi.clearAllMocks());

describe("syncProductsToDb", () => {
  it("returns processed count from detail_saved when it is a number", async () => {
    mockPost.mockResolvedValue({
      success: true,
      data: { detail_saved: 42, processed: 10, total: 50 },
    });

    const result = await syncProductsToDb();

    expect(result.processed).toBe(42);
    expect(result.total).toBe(50);
  });

  it("falls back to syncResult.processed when detail_saved is not a number", async () => {
    mockPost.mockResolvedValue({
      success: true,
      data: { processed: 15, total: 20 },
    });

    const result = await syncProductsToDb();

    expect(result.processed).toBe(15);
  });

  it("calls the correct endpoint with EXTRA_LONG timeout", async () => {
    mockPost.mockResolvedValue({
      success: true,
      data: {},
    });

    await syncProductsToDb();

    expect(mockPost).toHaveBeenCalledWith("/lazada/sync/products", undefined, {
      timeout: 120000,
    });
  });

  it("throws when response.success is false", async () => {
    mockPost.mockResolvedValue({
      success: false,
      error: "Lazada API unavailable",
    });

    await expect(syncProductsToDb()).rejects.toThrow("Lazada API unavailable");
  });

  it("throws default message when error field is empty", async () => {
    mockPost.mockResolvedValue({ success: false });

    await expect(syncProductsToDb()).rejects.toThrow(
      "Failed to sync Lazada products",
    );
  });

  it("spreads all other fields from syncResult into result", async () => {
    mockPost.mockResolvedValue({
      success: true,
      data: { message: "done", status: "ok", queued: 5, detail_saved: 3 },
    });

    const result = await syncProductsToDb();

    expect(result.message).toBe("done");
    expect(result.status).toBe("ok");
    expect(result.queued).toBe(5);
    expect(result.processed).toBe(3);
  });
});
