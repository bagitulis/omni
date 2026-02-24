import { describe, it, expect, vi, beforeEach } from "vitest";
import { exportOrders } from "./exports";

const { mockClientPost } = vi.hoisted(() => ({
  mockClientPost: vi.fn(),
}));

vi.mock("@/api/client", () => ({
  default: {
    client: {
      post: mockClientPost,
    },
  },
}));

vi.mock("@/lib/logger", () => ({
  logger: {
    error: vi.fn(),
    info: vi.fn(),
    debug: vi.fn(),
  },
}));

const mockLink = {
  href: "",
  download: "",
  click: vi.fn(),
  setAttribute: vi.fn(),
  parentNode: { removeChild: vi.fn() },
};

describe("exportOrders", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    global.URL.createObjectURL = vi.fn().mockReturnValue("blob:url");
    global.URL.revokeObjectURL = vi.fn();
    vi.spyOn(document, "createElement").mockReturnValue(
      mockLink as unknown as HTMLElement,
    );
    vi.spyOn(document.body, "appendChild").mockImplementation(
      () => mockLink as unknown as Node,
    );
    mockLink.href = "";
    mockLink.download = "";
    mockLink.click.mockClear();
    mockLink.setAttribute.mockClear();
  });

  it("returns true and triggers click on successful CSV export", async () => {
    mockClientPost.mockResolvedValue({
      data: new Blob(["csv data"]),
      headers: {},
    });

    const result = await exportOrders({
      date_from: "2024-01-01",
      date_to: "2024-01-31",
      format: "csv",
    });

    expect(result).toBe(true);
    expect(mockClientPost).toHaveBeenCalledWith(
      "/orders/export",
      {
        date_from: "2024-01-01",
        date_to: "2024-01-31",
        format: "csv",
      },
      {
        responseType: "blob",
        timeout: 60000,
      },
    );
    expect(mockLink.click).toHaveBeenCalledTimes(1);
    expect(URL.createObjectURL).toHaveBeenCalledTimes(1);
    expect(URL.revokeObjectURL).toHaveBeenCalledWith("blob:url");
  });

  it("returns true and triggers click on successful Excel export", async () => {
    mockClientPost.mockResolvedValue({
      data: new Blob(["excel data"]),
      headers: {},
    });

    const result = await exportOrders({
      date_from: "2024-01-01",
      date_to: "2024-01-31",
      format: "excel",
    });

    expect(result).toBe(true);
    expect(mockLink.click).toHaveBeenCalledTimes(1);
  });

  it("uses filename from content-disposition header when available", async () => {
    mockClientPost.mockResolvedValue({
      data: new Blob(["csv data"]),
      headers: {
        "content-disposition": 'attachment; filename="my_export.csv"',
      },
    });

    await exportOrders({
      date_from: "2024-01-01",
      date_to: "2024-01-31",
      format: "csv",
    });

    expect(mockLink.setAttribute).toHaveBeenCalledWith(
      "download",
      "my_export.csv",
    );
  });

  it("generates default filename when content-disposition is missing", async () => {
    mockClientPost.mockResolvedValue({
      data: new Blob(["data"]),
      headers: {},
    });

    await exportOrders({
      date_from: "2024-01-01",
      date_to: "2024-01-31",
      format: "excel",
    });

    const callArgs = mockLink.setAttribute.mock.calls[0];
    expect(callArgs[0]).toBe("download");
    expect(callArgs[1]).toMatch(/orders_export_.*\.xlsx/);
  });

  it("passes platform and status filters to API", async () => {
    mockClientPost.mockResolvedValue({
      data: new Blob(["csv"]),
      headers: {},
    });

    await exportOrders({
      platform: ["shopee", "tiktok"],
      status: ["PAID"],
      date_from: "2024-01-01",
      date_to: "2024-01-31",
      format: "csv",
    });

    expect(mockClientPost).toHaveBeenCalledWith(
      "/orders/export",
      expect.objectContaining({
        platform: ["shopee", "tiktok"],
        status: ["PAID"],
      }),
      expect.any(Object),
    );
  });

  it("throws error when API call fails", async () => {
    mockClientPost.mockRejectedValue(new Error("Network error"));

    await expect(
      exportOrders({
        date_from: "2024-01-01",
        date_to: "2024-01-31",
        format: "csv",
      }),
    ).rejects.toThrow("Network error");
  });
});
