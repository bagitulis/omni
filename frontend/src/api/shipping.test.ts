import { describe, it, expect, vi, beforeEach } from "vitest";

const {
  mockGet,
  mockPost,
  mockClientGet,
  mockOpen,
  mockCreateObjectURL,
  mockRevokeObjectURL,
  mockClick,
  mockAppendChild,
  mockRemoveChild,
} = vi.hoisted(() => ({
  mockGet: vi.fn(),
  mockPost: vi.fn(),
  mockClientGet: vi.fn(),
  mockOpen: vi.fn(),
  mockCreateObjectURL: vi.fn().mockReturnValue("blob:fake-url"),
  mockRevokeObjectURL: vi.fn(),
  mockClick: vi.fn(),
  mockAppendChild: vi.fn(),
  mockRemoveChild: vi.fn(),
}));

vi.mock("./client", () => ({
  default: {
    get: mockGet,
    post: mockPost,
    client: { get: mockClientGet },
  },
}));

// DOM mocks
Object.defineProperty(window, "open", { value: mockOpen, writable: true });
Object.defineProperty(window, "URL", {
  value: {
    createObjectURL: mockCreateObjectURL,
    revokeObjectURL: mockRevokeObjectURL,
  },
  writable: true,
});

const fakeLinkElement = {
  href: "",
  download: "",
  click: mockClick,
};

vi.spyOn(document, "createElement").mockImplementation((tag: string) => {
  if (tag === "a") return fakeLinkElement as unknown as HTMLElement;
  return document.createElement(tag);
});
vi.spyOn(document.body, "appendChild").mockImplementation(mockAppendChild);
vi.spyOn(document.body, "removeChild").mockImplementation(mockRemoveChild);

import {
  downloadBase64PDF,
  downloadShippingLabel,
  downloadShippingLabelsBatch,
} from "./shipping";

describe("downloadBase64PDF", () => {
  beforeEach(() => vi.clearAllMocks());

  it("creates a blob and triggers download", () => {
    const fakeBase64 = btoa("fake pdf content");
    downloadBase64PDF(fakeBase64, "test.pdf");

    expect(mockCreateObjectURL).toHaveBeenCalledOnce();
    expect(mockClick).toHaveBeenCalledOnce();
    expect(mockRevokeObjectURL).toHaveBeenCalledWith("blob:fake-url");
    expect(fakeLinkElement.download).toBe("test.pdf");
  });
});

describe("downloadShippingLabel", () => {
  beforeEach(() => vi.clearAllMocks());

  it("throws for unsupported platform", async () => {
    await expect(
      downloadShippingLabel("unknown-platform", "order-001"),
    ).rejects.toThrow("Unsupported platform: unknown-platform");
  });

  it("throws when response is not successful", async () => {
    mockGet.mockResolvedValue({ success: false, error: "Not found" });
    await expect(downloadShippingLabel("shopee", "order-001")).rejects.toThrow(
      "Not found",
    );
  });

  it("opens doc_url in new tab for shopee", async () => {
    mockGet.mockResolvedValue({
      success: true,
      data: { doc_url: "https://example.com/label.pdf" },
    });
    await downloadShippingLabel("shopee", "order-001");
    expect(mockOpen).toHaveBeenCalledWith(
      "https://example.com/label.pdf",
      "_blank",
      "noopener,noreferrer",
    );
  });

  it("uses correct endpoint for tiktok", async () => {
    mockGet.mockResolvedValue({
      success: true,
      data: { doc_url: "https://tiktok.com/label.pdf" },
    });
    await downloadShippingLabel("tiktok", "order-002");
    expect(mockGet).toHaveBeenCalledWith(
      "/tiktok/shipping/download/order/order-002",
      expect.any(Object),
    );
  });

  it("uses correct endpoint for lazada", async () => {
    mockGet.mockResolvedValue({
      success: true,
      data: { doc_url: "https://lazada.com/label.pdf" },
    });
    await downloadShippingLabel("lazada", "order-003");
    expect(mockGet).toHaveBeenCalledWith(
      "/lazada/shipping/label/order-003",
      expect.any(Object),
    );
  });

  it("downloads base64 file_data as PDF", async () => {
    const fakeBase64 = btoa("pdf bytes");
    mockGet.mockResolvedValue({
      success: true,
      data: { file_data: fakeBase64 },
    });
    await downloadShippingLabel("shopee", "order-004");
    expect(mockCreateObjectURL).toHaveBeenCalledOnce();
    expect(fakeLinkElement.download).toBe("shipping-label-order-004.pdf");
  });

  it("downloads from server when file_path is provided", async () => {
    mockGet.mockResolvedValue({
      success: true,
      data: { file_path: "/some/path/label.pdf" },
    });
    mockClientGet.mockResolvedValue({ data: new Blob(["pdf"]) });
    await downloadShippingLabel("shopee", "order-005");
    expect(mockClientGet).toHaveBeenCalledWith(
      "/uploads/labels/label.pdf",
      expect.objectContaining({ responseType: "blob" }),
    );
  });

  it("throws when no label data is available", async () => {
    mockGet.mockResolvedValue({
      success: true,
      data: {},
    });
    await expect(downloadShippingLabel("shopee", "order-006")).rejects.toThrow(
      "No shipping label available",
    );
  });
});

describe("downloadShippingLabelsBatch", () => {
  beforeEach(() => vi.clearAllMocks());

  it("throws for non-tiktok platform", async () => {
    await expect(
      downloadShippingLabelsBatch("shopee", ["order-001"]),
    ).rejects.toThrow("Batch download only supported for TikTok");
  });

  it("throws for lazada platform", async () => {
    await expect(
      downloadShippingLabelsBatch("lazada", ["order-001"]),
    ).rejects.toThrow("Batch download only supported for TikTok");
  });

  it("throws when response is not successful", async () => {
    mockPost.mockResolvedValue({ success: false, error: "Batch failed" });
    await expect(
      downloadShippingLabelsBatch("tiktok", ["order-001"]),
    ).rejects.toThrow("Batch failed");
  });

  it("returns batch result without file_data field", async () => {
    mockPost.mockResolvedValue({
      success: true,
      data: {
        total: 1,
        success: 1,
        failed: 0,
        results: [
          {
            order_id: "order-001",
            status: "SUCCESS",
            file_data: btoa("pdf"),
          },
        ],
      },
    });
    const result = await downloadShippingLabelsBatch("tiktok", ["order-001"]);
    expect(result.total).toBe(1);
    expect(result.success).toBe(1);
    expect(result.failed).toBe(0);
    expect(result.results[0]).not.toHaveProperty("file_data");
    expect(result.results[0].order_id).toBe("order-001");
  });

  it("downloads base64 for SUCCESS results", async () => {
    mockPost.mockResolvedValue({
      success: true,
      data: {
        total: 1,
        success: 1,
        failed: 0,
        results: [
          {
            order_id: "order-007",
            status: "SUCCESS",
            file_data: btoa("pdf bytes"),
          },
        ],
      },
    });
    await downloadShippingLabelsBatch("tiktok", ["order-007"]);
    expect(mockCreateObjectURL).toHaveBeenCalledOnce();
  });

  it("skips download for non-SUCCESS results", async () => {
    mockPost.mockResolvedValue({
      success: true,
      data: {
        total: 1,
        success: 0,
        failed: 1,
        results: [
          {
            order_id: "order-008",
            status: "FAILED",
            error: "Some error",
          },
        ],
      },
    });
    await downloadShippingLabelsBatch("tiktok", ["order-008"]);
    expect(mockCreateObjectURL).not.toHaveBeenCalled();
  });
});
