import { describe, it, expect, vi, beforeEach } from "vitest";
import {
  cloneProduct,
  batchClone,
  getCloneStatus,
  getProductData,
  getAvailableTargets,
  getClonePreview,
} from "./clone";

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

describe("clone API", () => {
  beforeEach(() => vi.clearAllMocks());

  describe("cloneProduct", () => {
    it("returns clone result on success", async () => {
      const data = { job_id: "job1", status: "pending" };
      mockPost.mockResolvedValue({ success: true, data });
      const req = {
        source_platform: "shopee",
        target_platform: "lazada",
        source_item_id: "123",
        sku: "SKU1",
      };
      const result = await cloneProduct(req);
      expect(mockPost).toHaveBeenCalledWith("/products/clone", req);
      expect(result).toEqual(data);
    });

    it("throws on failure", async () => {
      mockPost.mockResolvedValue({ success: false, error: "Clone failed" });
      await expect(
        cloneProduct({
          source_platform: "shopee",
          target_platform: "lazada",
          source_item_id: "123",
        }),
      ).rejects.toThrow("Clone failed");
    });
  });

  describe("batchClone", () => {
    it("returns batch result on success", async () => {
      const data = { total: 2, success: 2, failed: 0, results: [] };
      mockPost.mockResolvedValue({ success: true, data });
      const req = { items: [] };
      const result = await batchClone(req);
      expect(mockPost).toHaveBeenCalledWith("/products/clone/batch", req);
      expect(result).toEqual(data);
    });

    it("throws on failure", async () => {
      mockPost.mockResolvedValue({ success: false, error: "Batch failed" });
      await expect(batchClone({ items: [] })).rejects.toThrow("Batch failed");
    });
  });

  describe("getCloneStatus", () => {
    it("returns status on success", async () => {
      const data = { job_id: "job1", status: "completed" };
      mockGet.mockResolvedValue({ success: true, data });
      const result = await getCloneStatus("job1");
      expect(mockGet).toHaveBeenCalledWith("/products/clone/status/job1");
      expect(result).toEqual(data);
    });

    it("throws on failure", async () => {
      mockGet.mockResolvedValue({ success: false, error: "Not found" });
      await expect(getCloneStatus("job1")).rejects.toThrow("Not found");
    });
  });

  describe("getProductData", () => {
    it("returns product data on success", async () => {
      const product = { sku: "SKU1", name: "Test Product" };
      mockGet.mockResolvedValue({ success: true, data: { product } });
      const result = await getProductData("shopee", "SKU1");
      expect(mockGet).toHaveBeenCalledWith("/clone/product-data", {
        params: { platform: "shopee", sku: "SKU1" },
      });
      expect(result).toEqual(product);
    });

    it("throws on failure", async () => {
      mockGet.mockResolvedValue({ success: false, error: "Product not found" });
      await expect(getProductData("shopee", "SKU1")).rejects.toThrow(
        "Product not found",
      );
    });
  });

  describe("getAvailableTargets", () => {
    it("returns targets with all fields from response", async () => {
      const data = {
        sku: "SKU1",
        status: { shopee: true, lazada: false, tiktok: true },
        sources: ["shopee"],
        targets: ["lazada"],
      };
      mockGet.mockResolvedValue({ success: true, data });
      const result = await getAvailableTargets("SKU1");
      expect(result).toEqual(data);
    });

    it("defaults missing fields", async () => {
      mockGet.mockResolvedValue({ success: true, data: {} });
      const result = await getAvailableTargets("SKU2");
      expect(result.sku).toBe("SKU2");
      expect(result.status).toEqual({
        shopee: false,
        lazada: false,
        tiktok: false,
      });
      expect(result.sources).toEqual([]);
      expect(result.targets).toEqual([]);
    });

    it("throws on failure", async () => {
      mockGet.mockResolvedValue({ success: false, error: "Failed" });
      await expect(getAvailableTargets("SKU1")).rejects.toThrow("Failed");
    });
  });

  describe("getClonePreview", () => {
    it("returns preview on success", async () => {
      const data = { conflicts: [], can_clone: true };
      mockGet.mockResolvedValue({ success: true, data });
      const params = {
        source_platform: "shopee",
        target_platform: "lazada",
        source_item_id: "123",
        sku: "SKU1",
      };
      const result = await getClonePreview(params);
      expect(mockGet).toHaveBeenCalledWith("/clone/preview", { params });
      expect(result).toEqual(data);
    });

    it("throws on failure", async () => {
      mockGet.mockResolvedValue({ success: false, error: "Preview failed" });
      await expect(
        getClonePreview({
          source_platform: "shopee",
          target_platform: "lazada",
          source_item_id: "123",
        }),
      ).rejects.toThrow("Preview failed");
    });
  });
});
