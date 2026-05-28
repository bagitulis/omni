import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import {
  formatMonthYear,
  formatDate,
  getPriceDiffStatus,
  formatPriceDiff,
  getReconciliationHealth,
  exportToCSV,
  createCSV,
  getAnalyticsExportHeaders,
} from "./analyticsHelpers";
import type { ReconciliationSummary } from "@/types/analytics";

// ──────────────────────────────────────────────
//  formatMonthYear
// ──────────────────────────────────────────────
describe("formatMonthYear", () => {
  it("formats month 3, year 2026 as 'March 2026'", () => {
    expect(formatMonthYear(3, 2026)).toBe("March 2026");
  });

  it("formats January of a given year", () => {
    expect(formatMonthYear(1, 2025)).toBe("January 2025");
  });

  it("formats December of a given year", () => {
    expect(formatMonthYear(12, 2025)).toBe("December 2025");
  });
});

// ──────────────────────────────────────────────
//  formatDate
// ──────────────────────────────────────────────
describe("formatDate", () => {
  it("formats ISO date string to short readable format", () => {
    const result = formatDate("2026-05-01T00:00:00Z");
    expect(result).toBe("May 1, 2026");
  });

  it("returns '-' for null", () => {
    expect(formatDate(null)).toBe("-");
  });

  it("returns '-' for undefined", () => {
    expect(formatDate(undefined)).toBe("-");
  });

  it("returns '-' for empty string", () => {
    expect(formatDate("")).toBe("-");
  });

  it("returns '-' for invalid date string", () => {
    expect(formatDate("not-a-date")).toBe("-");
  });
});

// ──────────────────────────────────────────────
//  getPriceDiffStatus
// ──────────────────────────────────────────────
describe("getPriceDiffStatus", () => {
  it("returns 'success' for positive diff (profit)", () => {
    expect(getPriceDiffStatus(100)).toBe("success");
  });

  it("returns 'danger' for negative diff (loss)", () => {
    expect(getPriceDiffStatus(-50)).toBe("danger");
  });

  it("returns 'warning' for zero diff", () => {
    expect(getPriceDiffStatus(0)).toBe("warning");
  });
});

// ──────────────────────────────────────────────
//  formatPriceDiff
// ──────────────────────────────────────────────
describe("formatPriceDiff", () => {
  it("formats positive diff with +Rp prefix", () => {
    const result = formatPriceDiff(5000);
    expect(result).toMatch(/^\+/);
    expect(result).toContain("5.000");
  });

  it("formats negative diff with -Rp prefix", () => {
    const result = formatPriceDiff(-2000);
    expect(result).toMatch(/^-/);
    expect(result).toContain("2.000");
  });

  it("formats zero diff without sign prefix (just currency)", () => {
    const result = formatPriceDiff(0);
    expect(result.startsWith("+")).toBe(false);
    expect(result.startsWith("-")).toBe(false);
    expect(result).toContain("0");
  });
});

// ──────────────────────────────────────────────
//  getReconciliationHealth
// ──────────────────────────────────────────────
describe("getReconciliationHealth", () => {
  const baseSummary: ReconciliationSummary = {
    total_sku: 10,
    total_transactions: 50,
    sku_ok: 10,
    sku_with_price_diff: 0,
    sku_no_inventory: 0,
  };

  it("returns 'good' when all SKUs are OK", () => {
    expect(getReconciliationHealth(baseSummary)).toBe("good");
  });

  it("returns 'warning' when some SKUs have price differences", () => {
    const summary: ReconciliationSummary = {
      ...baseSummary,
      sku_ok: 8,
      sku_with_price_diff: 2,
    };
    expect(getReconciliationHealth(summary)).toBe("warning");
  });

  it("returns 'error' when some SKUs have no inventory", () => {
    const summary: ReconciliationSummary = {
      ...baseSummary,
      sku_ok: 7,
      sku_no_inventory: 3,
    };
    expect(getReconciliationHealth(summary)).toBe("error");
  });

  it("returns 'error' when both price diff and no-inventory SKUs exist", () => {
    const summary: ReconciliationSummary = {
      ...baseSummary,
      sku_ok: 5,
      sku_with_price_diff: 2,
      sku_no_inventory: 3,
    };
    expect(getReconciliationHealth(summary)).toBe("error");
  });

  it("returns 'warning' when not all SKUs OK but none have no inventory", () => {
    const summary: ReconciliationSummary = {
      ...baseSummary,
      sku_ok: 9,
      sku_with_price_diff: 1,
    };
    expect(getReconciliationHealth(summary)).toBe("warning");
  });

  it("returns 'warning' for an empty SKU summary", () => {
    expect(
      getReconciliationHealth({
        total_sku: 0,
        total_transactions: 0,
        sku_ok: 0,
        sku_with_price_diff: 0,
        sku_no_inventory: 0,
      })
    ).toBe("warning");
  });
});

// ──────────────────────────────────────────────
//  exportToCSV
// ──────────────────────────────────────────────
describe("exportToCSV", () => {
  let createElementSpy: ReturnType<typeof vi.spyOn>;
  let clickFn: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    clickFn = vi.fn();
    // URL.createObjectURL/revokeObjectURL are not available in jsdom,
    // so we define them manually
    Object.defineProperty(URL, "createObjectURL", {
      value: vi.fn(() => "blob:mock-url"),
      writable: true,
      configurable: true,
    });
    Object.defineProperty(URL, "revokeObjectURL", {
      value: vi.fn(),
      writable: true,
      configurable: true,
    });
    createElementSpy = vi.spyOn(document, "createElement").mockReturnValue({
      href: "",
      download: "",
      click: clickFn,
    } as unknown as HTMLAnchorElement);
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("creates a CSV download link and clicks it", () => {
    const data = [
      { name: "Alice", age: 30 },
      { name: "Bob", age: 25 },
    ];

    exportToCSV(data, "test-export");

    expect(createElementSpy).toHaveBeenCalledWith("a");
    expect(clickFn).toHaveBeenCalledTimes(1);
    expect(URL.createObjectURL).toHaveBeenCalledWith(expect.any(Blob));
    expect(URL.revokeObjectURL).toHaveBeenCalledWith("blob:mock-url");
  });

  it("uses custom headers when provided", () => {
    const data = [{ name: "Alice", age: 30 }];

    exportToCSV(data, "headers-test", {
      name: "Full Name",
      age: "Age (years)",
    });

    expect(createElementSpy).toHaveBeenCalledWith("a");
    expect(clickFn).toHaveBeenCalledTimes(1);
  });

  it("returns early for empty data array", () => {
    exportToCSV([], "empty");

    expect(createElementSpy).not.toHaveBeenCalled();
    expect(clickFn).not.toHaveBeenCalled();
  });

  it("returns an empty CSV for null rows instead of throwing", () => {
    expect(createCSV([null] as unknown as Record<string, unknown>[])).toBe("");
  });

  it("escapes CSV values containing commas or quotes", () => {
    const data = [
      { name: 'Smith, John', note: 'Said "hello"' },
    ];

    exportToCSV(data, "escape-test");

    expect(createElementSpy).toHaveBeenCalledWith("a");
    expect(clickFn).toHaveBeenCalledTimes(1);
  });
});

describe("analytics export parity", () => {
  it("includes restored Shopee reconciliation headers and formula value", () => {
    const headers = getAnalyticsExportHeaders("shopee", "reconciliation");
    const csv = createCSV(
      [
        {
          sku: "SKU-001",
          model_sku: "MODEL-001",
          item_name: "Test Item",
          model_name: "Blue",
          variant_name: "Large",
          inventory_price: 10000,
          expected_income: 4500,
          total_transactions: 1,
          unique_unit_prices: [10000],
          unique_actual_incomes: [4500],
          has_multiple_prices: false,
          has_price_difference: true,
          status: "PRICE_DIFF",
        },
      ],
      headers,
    );

    expect(csv.split("\r\n")[0]).toContain("Model SKU");
    expect(csv.split("\r\n")[0]).toContain("Expected Income");
    expect(csv).toContain("4500");
  });

  it("includes restored TikTok reconciliation headers and formula value", () => {
    const headers = getAnalyticsExportHeaders("tiktok", "reconciliation");
    const csv = createCSV(
      [
        {
          sku: "SKU-ID-001",
          seller_sku: "SELLER-SKU-001",
          product_name: "TikTok Item",
          variant_name: "Default",
          inventory_price: 12000,
          expected_income: 5000,
          total_transactions: 1,
          unique_unit_prices: [12000],
          unique_actual_incomes: [5000],
          has_multiple_prices: false,
          has_price_difference: true,
          status: "PRICE_DIFF",
        },
      ],
      headers,
    );

    expect(csv.split("\r\n")[0]).toContain("Seller SKU");
    expect(csv.split("\r\n")[0]).toContain("Unique Actual Incomes");
    expect(csv).toContain("5000");
  });

  it("includes platform-specific shipping fee headers and values", () => {
    const shopeeCsv = createCSV(
      [
        {
          order_sn: "TEST-SHOPEE-001",
          buyer_paid: 10000,
          actual_fee: 7000,
          shopee_rebate: 1500,
          difference: 4500,
          status: "profit",
          order_date: "2026-05-01",
          buyer_name: "Test User",
          payment_method: "COD",
        },
      ],
      getAnalyticsExportHeaders("shopee", "shipping_fee"),
    );
    const tiktokCsv = createCSV(
      [
        {
          order_sn: "TEST-TIKTOK-001",
          customer_paid: 12000,
          actual_fee: 8000,
          platform_discount: 1000,
          difference: 5000,
          status: "profit",
          order_date: "2026-05-01",
          order_status: "COMPLETED",
          currency: "IDR",
        },
      ],
      getAnalyticsExportHeaders("tiktok", "shipping_fee"),
    );

    expect(shopeeCsv.split("\r\n")[0]).toContain("Shopee Shipping Rebate");
    expect(shopeeCsv).toContain("4500");
    expect(tiktokCsv.split("\r\n")[0]).toContain("Shipping Fee Platform Discount");
    expect(tiktokCsv).toContain("5000");
  });
});


describe("explicit export header parity", () => {
  it("includes ALL Shopee reconciliation export headers explicitly", () => {
    const headers = getAnalyticsExportHeaders("shopee", "reconciliation");
    const csv = createCSV(
      [{
        sku: "S1", model_sku: "M1", item_name: "I", model_name: "M",
        variant_name: "V", inventory_price: 100, expected_income: 90,
        total_transactions: 1, unique_unit_prices: [100], unique_actual_incomes: [90],
        has_multiple_prices: false, has_price_difference: false, status: "OK",
      }],
      headers,
    );
    const headerRow = csv.split("\r\n")[0];
    expect(headerRow).toContain("SKU");
    expect(headerRow).toContain("Model SKU");
    expect(headerRow).toContain("Item Name");
    expect(headerRow).toContain("Model Name");
    expect(headerRow).toContain("Variant Name");
    expect(headerRow).toContain("Inventory Price");
    expect(headerRow).toContain("Expected Income");
    expect(headerRow).toContain("Total Transactions");
    expect(headerRow).toContain("Unique Unit Prices");
    expect(headerRow).toContain("Unique Actual Incomes");
    expect(headerRow).toContain("Has Multiple Prices");
    expect(headerRow).toContain("Has Price Difference");
    expect(headerRow).toContain("Status");
    expect(headerRow.split(",").length).toBe(13);
  });

  it("includes ALL TikTok reconciliation export headers explicitly", () => {
    const headers = getAnalyticsExportHeaders("tiktok", "reconciliation");
    const csv = createCSV(
      [{
        sku: "S1", seller_sku: "SS1", product_name: "P",
        variant_name: "V", inventory_price: 100, expected_income: 90,
        total_transactions: 1, unique_unit_prices: [100], unique_actual_incomes: [90],
        has_multiple_prices: false, has_price_difference: false, status: "OK",
      }],
      headers,
    );
    const headerRow = csv.split("\r\n")[0];
    expect(headerRow).toContain("SKU ID");
    expect(headerRow).toContain("Seller SKU");
    expect(headerRow).toContain("Product Name");
    expect(headerRow).toContain("Variant Name");
    expect(headerRow).toContain("Inventory Price");
    expect(headerRow).toContain("Expected Income");
    expect(headerRow).toContain("Total Transactions");
    expect(headerRow).toContain("Unique Unit Prices");
    expect(headerRow).toContain("Unique Actual Incomes");
    expect(headerRow).toContain("Has Multiple Prices");
    expect(headerRow).toContain("Has Price Difference");
    expect(headerRow).toContain("Status");
    expect(headerRow.split(",").length).toBe(12);
  });

  it("includes ALL Shopee shipping export headers explicitly", () => {
    const headers = getAnalyticsExportHeaders("shopee", "shipping_fee");
    const csv = createCSV(
      [{
        order_sn: "O1", buyer_paid: 100, actual_fee: 50,
        shopee_rebate: 10, difference: 60, status: "profit",
        order_date: "2026-01-01", buyer_name: "Buyer", payment_method: "COD",
      }],
      headers,
    );
    const headerRow = csv.split("\r\n")[0];
    expect(headerRow).toContain("Order SN");
    expect(headerRow).toContain("Buyer Paid Shipping Fee");
    expect(headerRow).toContain("Actual Shipping Fee");
    expect(headerRow).toContain("Shopee Shipping Rebate");
    expect(headerRow).toContain("Difference");
    expect(headerRow).toContain("Status");
    expect(headerRow).toContain("Order Date");
    expect(headerRow).toContain("Buyer Name");
    expect(headerRow).toContain("Payment Method");
    expect(headerRow.split(",").length).toBe(9);
  });

  it("includes ALL TikTok shipping export headers explicitly", () => {
    const headers = getAnalyticsExportHeaders("tiktok", "shipping_fee");
    const csv = createCSV(
      [{
        order_sn: "O1", customer_paid: 100, actual_fee: 50,
        platform_discount: 10, difference: 60, status: "profit",
        order_date: "2026-01-01", order_status: "COMPLETED", currency: "IDR",
      }],
      headers,
    );
    const headerRow = csv.split("\r\n")[0];
    expect(headerRow).toContain("Order SN");
    expect(headerRow).toContain("Shipping Fee Customer Paid");
    expect(headerRow).toContain("Shipping Fee Actual");
    expect(headerRow).toContain("Shipping Fee Platform Discount");
    expect(headerRow).toContain("Difference");
    expect(headerRow).toContain("Status");
    expect(headerRow).toContain("Order Date");
    expect(headerRow).toContain("Order Status");
    expect(headerRow).toContain("Currency");
    expect(headerRow.split(",").length).toBe(9);
  });

// ──────────────────────────────────────────────
//  Timezone Consistency
// ──────────────────────────────────────────────
describe("Timezone stability (analytics)", () => {
  it("formatDate always returns UTC date regardless of system timezone", () => {
    // "2026-05-01T00:00:00Z" is UTC midnight May 1
    // Without timeZone: 'UTC', browser in UTC-5 shows "Apr 30, 2026"
    const result = formatDate("2026-05-01T00:00:00Z");
    expect(result).toBe("May 1, 2026");
  });

  it("formatDate handles late-night UTC without date shift", () => {
    // "2026-05-01T23:59:00Z" is 11:59pm UTC on May 1
    // In UTC+7 this becomes "May 2, 2026 06:59am" - must NOT shift
    const result = formatDate("2026-05-01T23:59:00Z");
    expect(result).toBe("May 1, 2026");
  });

  it("formatDate handles first-second of month without date shift", () => {
    // "2026-02-01T00:00:00Z" in UTC-12 would be "Jan 31" without timezone fix
    const result = formatDate("2026-02-01T00:00:00Z");
    expect(result).toBe("Feb 1, 2026");
  });
});

// ──────────────────────────────────────────────
//  Locale Consistency
// ──────────────────────────────────────────────
describe("Locale consistency (analytics)", () => {
  it("formatMonthYear uses en-US locale consistently", () => {
    expect(formatMonthYear(1, 2026)).toBe("January 2026");
    expect(formatMonthYear(6, 2026)).toBe("June 2026");
    expect(formatMonthYear(12, 2026)).toBe("December 2026");
  });

  it("formatMonthYear is stable for UTC boundary months", () => {
    // Month/year pair at UTC boundaries should be stable
    expect(formatMonthYear(1, 2026)).toBe("January 2026");
    expect(formatMonthYear(2, 2026)).toBe("February 2026");
    expect(formatMonthYear(12, 2025)).toBe("December 2025");
  });
});
});
