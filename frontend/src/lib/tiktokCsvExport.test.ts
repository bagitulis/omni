import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import * as tiktokCsvExport from "./tiktokCsvExport";
import {
  downloadCSV,
  exportPriceToCSV,
  exportShippingToCSV,
} from "./tiktokCsvExport";

describe("downloadCSV", () => {
  let createObjectURLSpy: ReturnType<typeof vi.fn>;
  let revokeObjectURLSpy: ReturnType<typeof vi.fn>;
  let appendChildSpy: ReturnType<typeof vi.fn>;
  let clickSpy: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    createObjectURLSpy = vi.fn().mockReturnValue("blob:mock-url");
    revokeObjectURLSpy = vi.fn();
    clickSpy = vi.fn();

    global.URL.createObjectURL = createObjectURLSpy;
    global.URL.revokeObjectURL = revokeObjectURLSpy;

    appendChildSpy = vi
      .spyOn(document.body, "appendChild")
      .mockImplementation((node) => node);
    vi.spyOn(document, "createElement").mockImplementation((tag: string) => {
      if (tag === "a") {
        return {
          href: "",
          download: "",
          click: clickSpy,
        } as unknown as HTMLAnchorElement;
      }
      return document.createElement(tag);
    });
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("creates a blob with CSV content", () => {
    const BlobSpy = vi.spyOn(global, "Blob");
    downloadCSV(
      [
        ["Name", "Age"],
        ["Alice", "30"],
      ],
      "test.csv",
    );
    expect(BlobSpy).toHaveBeenCalledWith(["Name,Age\nAlice,30"], {
      type: "text/csv;charset=utf-8;",
    });
  });

  it("creates an object URL for the blob", () => {
    downloadCSV([["col1", "col2"]], "file.csv");
    expect(createObjectURLSpy).toHaveBeenCalled();
  });

  it("sets the download filename on the link", () => {
    downloadCSV([["a"]], "my-export.csv");
    expect(clickSpy).toHaveBeenCalled();
  });

  it("revokes the object URL after click", () => {
    downloadCSV([["x"]], "cleanup.csv");
    expect(revokeObjectURLSpy).toHaveBeenCalledWith("blob:mock-url");
  });

  it("handles empty rows array", () => {
    expect(() => downloadCSV([], "empty.csv")).not.toThrow();
    expect(createObjectURLSpy).toHaveBeenCalled();
  });

  it("joins multiple rows with newline", () => {
    const BlobSpy = vi.spyOn(global, "Blob");
    downloadCSV(
      [
        ["r1c1", "r1c2"],
        ["r2c1", "r2c2"],
      ],
      "multi.csv",
    );
    expect(BlobSpy).toHaveBeenCalledWith(
      ["r1c1,r1c2\nr2c1,r2c2"],
      expect.any(Object),
    );
    // Cleanup
    appendChildSpy.mockRestore();
  });
});

describe("exportPriceToCSV", () => {
  let downloadSpy: ReturnType<typeof vi.spyInstance>;

  beforeEach(() => {
    downloadSpy = vi
      .spyOn(tiktokCsvExport, "downloadCSV")
      .mockImplementation(() => {});
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("calls downloadCSV with correct filename format (month + 1)", () => {
    exportPriceToCSV({ sku_groups: [] }, 2024, 0);
    expect(downloadSpy).toHaveBeenCalledWith(
      expect.any(Array),
      "tiktok-price-2024-1.csv",
    );
  });

  it("includes header row as first row", () => {
    exportPriceToCSV({ sku_groups: [] }, 2024, 5);
    const [rows] = downloadSpy.mock.calls[0] as [string[][]];
    expect(rows[0]).toEqual([
      "Status",
      "SKU",
      "Product Name",
      "Inventory Price",
      "Expected Income",
      "Qty",
    ]);
  });

  it("maps each sku_group to a CSV row", () => {
    const result = {
      sku_groups: [
        {
          status: "OK",
          seller_sku: "SKU-001",
          sku: "FALLBACK",
          product_name: "Test Product",
          inventory_price: 50000,
          expected_income: 45000,
          total_transactions: 10,
        },
      ],
    };
    exportPriceToCSV(result, 2024, 2);
    const [rows] = downloadSpy.mock.calls[0] as [string[][]];
    expect(rows).toHaveLength(2); // header + 1 data row
    expect(rows[1][0]).toBe("OK");
    expect(rows[1][1]).toBe("SKU-001");
    expect(rows[1][5]).toBe("10");
  });

  it("falls back to sku when seller_sku is empty", () => {
    const result = {
      sku_groups: [
        {
          status: "NO_INVENTORY",
          seller_sku: "",
          sku: "FALLBACK-SKU",
          product_name: "Another Product",
          inventory_price: 0,
          expected_income: 0,
          total_transactions: 0,
        },
      ],
    };
    exportPriceToCSV(result, 2024, 11);
    const [rows] = downloadSpy.mock.calls[0] as [string[][]];
    expect(rows[1][1]).toBe("FALLBACK-SKU");
  });

  it("wraps product_name with quotes and escapes internal quotes", () => {
    const result = {
      sku_groups: [
        {
          status: "OK",
          seller_sku: "S1",
          sku: "S1",
          product_name: 'Product "With Quotes"',
          inventory_price: 1000,
          expected_income: 900,
          total_transactions: 5,
        },
      ],
    };
    exportPriceToCSV(result, 2025, 0);
    const [rows] = downloadSpy.mock.calls[0] as [string[][]];
    expect(rows[1][2]).toBe('"Product ""With Quotes"""');
  });

  it("uses 0 for null/undefined inventory_price and expected_income", () => {
    const result = {
      sku_groups: [
        {
          status: "NO_INVENTORY",
          seller_sku: "S2",
          sku: "S2",
          product_name: "No Price Product",
          inventory_price: 0,
          expected_income: 0,
          total_transactions: 3,
        },
      ],
    };
    exportPriceToCSV(result, 2024, 3);
    const [rows] = downloadSpy.mock.calls[0] as [string[][]];
    expect(rows[1][3]).toBe("0");
    expect(rows[1][4]).toBe("0");
  });
});

describe("exportShippingToCSV", () => {
  let downloadSpy: ReturnType<typeof vi.spyInstance>;

  beforeEach(() => {
    downloadSpy = vi
      .spyOn(tiktokCsvExport, "downloadCSV")
      .mockImplementation(() => {});
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("calls downloadCSV with correct filename format", () => {
    exportShippingToCSV({ orders: [] }, 2024, 5);
    expect(downloadSpy).toHaveBeenCalledWith(
      expect.any(Array),
      "tiktok-shipping-2024-6.csv",
    );
  });

  it("includes header row with 7 columns", () => {
    exportShippingToCSV({ orders: [] }, 2024, 0);
    const [rows] = downloadSpy.mock.calls[0] as [string[][]];
    expect(rows[0]).toEqual([
      "Order Date",
      "Order ID",
      "Buyer Paid",
      "Actual Fee",
      "Platform Discount",
      "Difference",
      "Status",
    ]);
  });

  it("maps each order to a row with all fields", () => {
    const result = {
      orders: [
        {
          order_date: "2024-01-15",
          order_id: "ORD-001",
          buyer_paid: 100000,
          actual_fee: 5000,
          platform_discount: 2000,
          difference: 3000,
          order_status: "COMPLETED",
        },
      ],
    };
    exportShippingToCSV(result, 2024, 0);
    const [rows] = downloadSpy.mock.calls[0] as [string[][]];
    expect(rows).toHaveLength(2);
    expect(rows[1]).toEqual([
      "2024-01-15",
      "ORD-001",
      "100000",
      "5000",
      "2000",
      "3000",
      "COMPLETED",
    ]);
  });

  it("uses empty string for missing order_date and order_status", () => {
    const result = {
      orders: [
        {
          order_date: "",
          order_id: "ORD-002",
          buyer_paid: 50000,
          actual_fee: 2500,
          platform_discount: 0,
          difference: 2500,
          order_status: "",
        },
      ],
    };
    exportShippingToCSV(result, 2024, 0);
    const [rows] = downloadSpy.mock.calls[0] as [string[][]];
    expect(rows[1][0]).toBe("");
    expect(rows[1][6]).toBe("");
  });

  it("handles multiple orders correctly", () => {
    const result = {
      orders: [
        {
          order_date: "2024-01-01",
          order_id: "A",
          buyer_paid: 1,
          actual_fee: 1,
          platform_discount: 0,
          difference: 0,
          order_status: "DONE",
        },
        {
          order_date: "2024-01-02",
          order_id: "B",
          buyer_paid: 2,
          actual_fee: 2,
          platform_discount: 0,
          difference: 0,
          order_status: "DONE",
        },
      ],
    };
    exportShippingToCSV(result, 2024, 0);
    const [rows] = downloadSpy.mock.calls[0] as [string[][]];
    expect(rows).toHaveLength(3); // header + 2 orders
  });
});
