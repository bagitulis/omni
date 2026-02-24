/**
 * Tests for tiktokCsvExport.ts
 *
 * NOTE on ESM module structure:
 * exportPriceToCSV and exportShippingToCSV call downloadCSV from within the same
 * ES module. Vitest cannot intercept same-module internal calls via spyOn or vi.mock.
 * Strategy: test everything end-to-end by mocking URL.createObjectURL and reading
 * the Blob content to verify what CSV was produced.
 */
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import {
  downloadCSV,
  exportPriceToCSV,
  exportShippingToCSV,
} from "./tiktokCsvExport";
// Helper: read Blob text content passed to URL.createObjectURL
function capturedBlobText(
  createObjectURLMock: ReturnType<typeof vi.fn>,
): Promise<string> {
  return new Promise((resolve) => {
    const blob = createObjectURLMock.mock.calls[0][0] as Blob;
    const reader = new FileReader();
    reader.onload = () => resolve(reader.result as string);
    reader.readAsText(blob);
  });
}

// Shared DOM setup
function setupDomMocks(clickMock: ReturnType<typeof vi.fn>) {
  vi.spyOn(document.body, "appendChild").mockImplementation((node) => node);
  vi.spyOn(document, "createElement").mockImplementation((tag: string) => {
    if (tag === "a") {
      return {
        href: "",
        download: "",
        click: clickMock,
      } as unknown as HTMLAnchorElement;
    }
    return document.createElement(tag);
  });
}

// ---------------------------------------------------------------
// downloadCSV
// ---------------------------------------------------------------
describe("downloadCSV", () => {
  let createObjectURLMock: ReturnType<typeof vi.fn>;
  let revokeObjectURLMock: ReturnType<typeof vi.fn>;
  let clickMock: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    createObjectURLMock = vi.fn().mockReturnValue("blob:mock-url");
    revokeObjectURLMock = vi.fn();
    clickMock = vi.fn();
    globalThis.URL.createObjectURL = createObjectURLMock;
    globalThis.URL.revokeObjectURL = revokeObjectURLMock;
    setupDomMocks(clickMock);
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("calls URL.createObjectURL with a Blob instance", () => {
    downloadCSV([["Name", "Age"], ["Alice", "30"]], "test.csv");
    expect(createObjectURLMock).toHaveBeenCalledOnce();
    const arg = createObjectURLMock.mock.calls[0][0];
    expect(arg).toBeInstanceOf(Blob);
  });
  it("triggers click on the anchor element", () => {
    downloadCSV([["a"]], "my-export.csv");
    expect(clickMock).toHaveBeenCalled();
  });
  it("revokes the object URL after click", () => {
    downloadCSV([["x"]], "cleanup.csv");
    expect(revokeObjectURLMock).toHaveBeenCalledWith("blob:mock-url");
  });
  it("handles empty rows array without throwing", () => {
    expect(() => downloadCSV([], "empty.csv")).not.toThrow();
    expect(createObjectURLMock).toHaveBeenCalled();
  });

  it("passes CSV-formatted content to the Blob", async () => {
    downloadCSV(
      [["r1c1", "r1c2"], ["r2c1", "r2c2"]],
      "multi.csv",
    );
    const text = await capturedBlobText(createObjectURLMock);
    expect(text).toBe("r1c1,r1c2\nr2c1,r2c2");
  });

  it("produces single row CSV correctly", async () => {
    downloadCSV([["col1", "col2", "col3"]], "single.csv");
    const text = await capturedBlobText(createObjectURLMock);
    expect(text).toBe("col1,col2,col3");
  });
});

// ---------------------------------------------------------------
// exportPriceToCSV - tested end-to-end via Blob content
// ---------------------------------------------------------------
describe("exportPriceToCSV", () => {
  let createObjectURLMock: ReturnType<typeof vi.fn>;
  let clickMock: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    createObjectURLMock = vi.fn().mockReturnValue("blob:mock-url");
    clickMock = vi.fn();
    globalThis.URL.createObjectURL = createObjectURLMock;
    globalThis.URL.revokeObjectURL = vi.fn();
    setupDomMocks(clickMock);
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("triggers downloadCSV (URL.createObjectURL called once)", () => {
    exportPriceToCSV({ sku_groups: [] }, 2024, 0);
    expect(createObjectURLMock).toHaveBeenCalledOnce();
  });

  it("CSV contains correct header row", async () => {
    exportPriceToCSV({ sku_groups: [] }, 2024, 5);
    const text = await capturedBlobText(createObjectURLMock);
    expect(text).toContain("Status,SKU,Product Name,Inventory Price,Expected Income,Qty");
  });

  it("maps each sku_group to a CSV data row", async () => {
    const result = {
      sku_groups: [{
        status: "OK",
        seller_sku: "SKU-001",
        sku: "FALLBACK",
        product_name: "Test Product",
        inventory_price: 50000,
        expected_income: 45000,
        total_transactions: 10,
      }],
    };
    exportPriceToCSV(result, 2024, 2);
    const text = await capturedBlobText(createObjectURLMock);
    const lines = text.split("\n");
    expect(lines).toHaveLength(2);
    const cols = lines[1].split(",");
    expect(cols[0]).toBe("OK");
    expect(cols[1]).toBe("SKU-001");
    expect(cols[5]).toBe("10");
  });

  it("falls back to sku when seller_sku is empty", async () => {
    const result = {
      sku_groups: [{
        status: "NO_INVENTORY",
        seller_sku: "",
        sku: "FALLBACK-SKU",
        product_name: "Another Product",
        inventory_price: 0,
        expected_income: 0,
        total_transactions: 0,
      }],
    };
    exportPriceToCSV(result, 2024, 11);
    const text = await capturedBlobText(createObjectURLMock);
    expect(text).toContain("FALLBACK-SKU");
  });

  it("wraps product_name with quotes and escapes internal quotes", async () => {
    const result = {
      sku_groups: [{
        status: "OK",
        seller_sku: "S1",
        sku: "S1",
        product_name: 'Product "With Quotes"',
        inventory_price: 1000,
        expected_income: 900,
        total_transactions: 5,
      }],
    };
    exportPriceToCSV(result, 2025, 0);
    const text = await capturedBlobText(createObjectURLMock);
    expect(text).toContain('"Product ""With Quotes"""');
  });

  it("uses 0 for zero inventory_price and expected_income", async () => {
    const result = {
      sku_groups: [{
        status: "NO_INVENTORY",
        seller_sku: "S2",
        sku: "S2",
        product_name: "No Price Product",
        inventory_price: 0,
        expected_income: 0,
        total_transactions: 3,
      }],
    };
    exportPriceToCSV(result, 2024, 3);
    const text = await capturedBlobText(createObjectURLMock);
    const lines = text.split("\n");
    const cols = lines[1].split(",");
    expect(cols[3]).toBe("0");
    expect(cols[4]).toBe("0");
  });
});

// ---------------------------------------------------------------
// exportShippingToCSV - tested end-to-end via Blob content
// ---------------------------------------------------------------
describe("exportShippingToCSV", () => {
  let createObjectURLMock: ReturnType<typeof vi.fn>;
  let clickMock: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    createObjectURLMock = vi.fn().mockReturnValue("blob:mock-url");
    clickMock = vi.fn();
    globalThis.URL.createObjectURL = createObjectURLMock;
    globalThis.URL.revokeObjectURL = vi.fn();
    setupDomMocks(clickMock);
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("triggers downloadCSV (URL.createObjectURL called once)", () => {
    exportShippingToCSV({ orders: [] }, 2024, 5);
    expect(createObjectURLMock).toHaveBeenCalledOnce();
  });

  it("CSV contains correct header row with 7 columns", async () => {
    exportShippingToCSV({ orders: [] }, 2024, 0);
    const text = await capturedBlobText(createObjectURLMock);
    expect(text).toContain(
      "Order Date,Order ID,Buyer Paid,Actual Fee,Platform Discount,Difference,Status",
    );
  });

  it("maps each order to a CSV row with all fields", async () => {
    const result = {
      orders: [{
        order_date: "2024-01-15",
        order_id: "ORD-001",
        buyer_paid: 100000,
        actual_fee: 5000,
        platform_discount: 2000,
        difference: 3000,
        order_status: "COMPLETED",
      }],
    };
    exportShippingToCSV(result, 2024, 0);
    const text = await capturedBlobText(createObjectURLMock);
    expect(text).toContain("2024-01-15,ORD-001,100000,5000,2000,3000,COMPLETED");
  });

  it("uses empty string for missing order_date and order_status", async () => {
    const result = {
      orders: [{
        order_date: "",
        order_id: "ORD-002",
        buyer_paid: 50000,
        actual_fee: 2500,
        platform_discount: 0,
        difference: 2500,
        order_status: "",
      }],
    };
    exportShippingToCSV(result, 2024, 0);
    const text = await capturedBlobText(createObjectURLMock);
    const lines = text.split("\n");
    const cols = lines[1].split(",");
    expect(cols[0]).toBe("");
    expect(cols[6]).toBe("");
  });

  it("handles multiple orders - correct row count", async () => {
    const result = {
      orders: [
        {
          order_date: "2024-01-01", order_id: "A",
          buyer_paid: 1, actual_fee: 1,
          platform_discount: 0, difference: 0,
          order_status: "DONE",
        },
        {
          order_date: "2024-01-02", order_id: "B",
          buyer_paid: 2, actual_fee: 2,
          platform_discount: 0, difference: 0,
          order_status: "DONE",
        },
      ],
    };
    exportShippingToCSV(result, 2024, 0);
    const text = await capturedBlobText(createObjectURLMock);
    const lines = text.split("\n");
    expect(lines).toHaveLength(3); // header + 2 orders
  });
});
