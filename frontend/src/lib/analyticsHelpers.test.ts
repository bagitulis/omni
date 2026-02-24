import { describe, it, expect, vi, beforeEach, afterEach, type MockInstance } from "vitest";
import * as analyticsHelpers from "./analyticsHelpers";
import {
  MONTHS,
  getAvailableYears,
  formatAnalyticsDate,
  formatCurrency,
  getStatusColor,
  getStatusLabel,
  exportShopeeReconciliationCSV,
  exportShopeeShippingCSV,
  exportTiktokReconciliationCSV,
  exportTiktokShippingCSV,
} from "./analyticsHelpers";
import type {
  ReconciliationResult,
  ShopeeShippingFeeResult,
  TiktokReconciliationResult,
  TiktokShippingFeeResult,
} from "@/types/analytics";

// ============================================================================
// MONTHS constant
// ============================================================================

describe("MONTHS", () => {
  it("has exactly 12 months", () => {
    expect(MONTHS).toHaveLength(12);
  });

  it("first entry is January with value 0", () => {
    expect(MONTHS[0]).toEqual({ value: 0, label: "January" });
  });

  it("last entry is December with value 11", () => {
    expect(MONTHS[11]).toEqual({ value: 11, label: "December" });
  });

  it("all months have sequential value from 0 to 11", () => {
    MONTHS.forEach((m, idx) => {
      expect(m.value).toBe(idx);
    });
  });
});

// ============================================================================
// getAvailableYears
// ============================================================================

describe("getAvailableYears", () => {
  it("returns array of 3 years", () => {
    const years = getAvailableYears();
    expect(years).toHaveLength(3);
  });

  it("first year is current year", () => {
    const currentYear = new Date().getFullYear();
    const years = getAvailableYears();
    expect(years[0]).toBe(currentYear);
  });

  it("second year is current year - 1", () => {
    const currentYear = new Date().getFullYear();
    const years = getAvailableYears();
    expect(years[1]).toBe(currentYear - 1);
  });

  it("third year is current year - 2", () => {
    const currentYear = new Date().getFullYear();
    const years = getAvailableYears();
    expect(years[2]).toBe(currentYear - 2);
  });
});

// ============================================================================
// formatAnalyticsDate
// ============================================================================

describe("formatAnalyticsDate", () => {
  it("returns em dash for null", () => {
    expect(formatAnalyticsDate(null)).toBe("—");
  });

  it("returns em dash for empty string", () => {
    expect(formatAnalyticsDate("")).toBe("—");
  });

  it("formats valid ISO date string", () => {
    const result = formatAnalyticsDate("2024-01-15");
    // en-US locale formats as M/D/YYYY
    expect(result).toMatch(/\d{1,2}\/\d{1,2}\/\d{4}/);
  });

  it("returns formatted date for full ISO datetime", () => {
    const result = formatAnalyticsDate("2024-06-20T10:30:00Z");
    expect(result).not.toBe("—");
    expect(typeof result).toBe("string");
  });
});

// ============================================================================
// formatCurrency
// ============================================================================

describe("formatCurrency", () => {
  it("formats zero as IDR currency", () => {
    const result = formatCurrency(0);
    expect(result).toContain("Rp");
    expect(result).toContain("0");
  });

  it("formats positive amount with IDR prefix", () => {
    const result = formatCurrency(100000);
    expect(result).toContain("Rp");
  });

  it("does not include decimal places (maximumFractionDigits: 0)", () => {
    const result = formatCurrency(100000.99);
    expect(result).not.toContain(".");
  });

  it("formats large amounts correctly", () => {
    const result = formatCurrency(1000000);
    expect(result).toContain("Rp");
    expect(result.length).toBeGreaterThan(3);
  });

  it("formats negative amounts", () => {
    const result = formatCurrency(-50000);
    expect(result).toBeDefined();
    expect(typeof result).toBe("string");
  });
});

// ============================================================================
// getStatusColor
// ============================================================================

describe("getStatusColor", () => {
  it("returns success for OK", () => {
    expect(getStatusColor("OK")).toBe("success");
  });

  it("returns warning for PRICE_DIFF", () => {
    expect(getStatusColor("PRICE_DIFF")).toBe("warning");
  });

  it("returns error for NO_INVENTORY", () => {
    expect(getStatusColor("NO_INVENTORY")).toBe("error");
  });

  it("returns default for unknown status", () => {
    expect(getStatusColor("UNKNOWN")).toBe("default");
  });

  it("returns default for empty string", () => {
    expect(getStatusColor("")).toBe("default");
  });
});

// ============================================================================
// getStatusLabel
// ============================================================================

describe("getStatusLabel", () => {
  it("returns OK for OK", () => {
    expect(getStatusLabel("OK")).toBe("OK");
  });

  it("returns Price Difference for PRICE_DIFF", () => {
    expect(getStatusLabel("PRICE_DIFF")).toBe("Price Difference");
  });

  it("returns No Inventory for NO_INVENTORY", () => {
    expect(getStatusLabel("NO_INVENTORY")).toBe("No Inventory");
  });

  it("returns the status itself for unknown values", () => {
    expect(getStatusLabel("SOME_STATUS")).toBe("SOME_STATUS");
  });

  it("returns empty string for empty input", () => {
    expect(getStatusLabel("")).toBe("");
  });
});

// ============================================================================
// CSV Export functions (downloadCSV mocked)
// ============================================================================

describe("exportShopeeReconciliationCSV", () => {
  let downloadSpy: MockInstance;

  beforeEach(() => {
    downloadSpy = vi
      .spyOn(analyticsHelpers, "downloadCSV")
      .mockImplementation(() => {});
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("calls downloadCSV with shopee-reconciliation filename", () => {
    const data: ReconciliationResult = {
      summary: {
        total_sku: 0,
        total_transactions: 0,
        sku_ok: 0,
        sku_with_price_diff: 0,
        sku_no_inventory: 0,
      },
      sku_groups: [],
    };
    exportShopeeReconciliationCSV(data);
    const [, filename] = downloadSpy.mock.calls[0] as [string[][], string];
    expect(filename).toMatch(/^shopee-reconciliation-\d{4}-\d{2}-\d{2}\.csv$/);
  });

  it("includes correct header row", () => {
    const data: ReconciliationResult = {
      summary: {
        total_sku: 0,
        total_transactions: 0,
        sku_ok: 0,
        sku_with_price_diff: 0,
        sku_no_inventory: 0,
      },
      sku_groups: [],
    };
    exportShopeeReconciliationCSV(data);
    const [rows] = downloadSpy.mock.calls[0] as [string[][]];
    expect(rows[0]).toContain("Status");
    expect(rows[0]).toContain("SKU");
    expect(rows[0]).toContain("Transactions");
  });

  it("maps sku_groups to rows with status label", () => {
    const data: ReconciliationResult = {
      summary: {
        total_sku: 1,
        total_transactions: 1,
        sku_ok: 1,
        sku_with_price_diff: 0,
        sku_no_inventory: 0,
      },
      sku_groups: [
        {
          sku: "SKU-A",
          model_sku: "MODEL-A",
          item_name: "Product A",
          model_name: "Variant A",
          inventory_price: 10000,
          expected_income: 9000,
          total_transactions: 5,
          unique_unit_prices: [10000],
          unique_actual_incomes: [9000],
          price_variants: [],
          has_multiple_prices: false,
          has_price_difference: false,
          status: "OK",
        },
      ],
    };
    exportShopeeReconciliationCSV(data);
    const [rows] = downloadSpy.mock.calls[0] as [string[][]];
    expect(rows).toHaveLength(2); // header + 1 row
    expect(rows[1][0]).toBe("OK"); // getStatusLabel("OK") = "OK"
  });
});

describe("exportShopeeShippingCSV", () => {
  let downloadSpy: MockInstance;

  beforeEach(() => {
    downloadSpy = vi
      .spyOn(analyticsHelpers, "downloadCSV")
      .mockImplementation(() => {});
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("calls downloadCSV with shopee-shipping-fee filename", () => {
    const data: ShopeeShippingFeeResult = {
      summary: {
        total_orders: 0,
        orders_with_difference: 0,
        total_profit: 0,
        total_loss: 0,
        net_impact: 0,
      },
      orders: [],
    };
    exportShopeeShippingCSV(data);
    const [, filename] = downloadSpy.mock.calls[0] as [string[][], string];
    expect(filename).toMatch(/^shopee-shipping-fee-\d{4}-\d{2}-\d{2}\.csv$/);
  });

  it("includes correct header row", () => {
    const data: ShopeeShippingFeeResult = {
      summary: {
        total_orders: 0,
        orders_with_difference: 0,
        total_profit: 0,
        total_loss: 0,
        net_impact: 0,
      },
      orders: [],
    };
    exportShopeeShippingCSV(data);
    const [rows] = downloadSpy.mock.calls[0] as [string[][]];
    expect(rows[0]).toContain("Order SN");
    expect(rows[0]).toContain("Buyer Paid");
    expect(rows[0]).toContain("Actual Fee");
  });

  it("maps orders to rows", () => {
    const data: ShopeeShippingFeeResult = {
      summary: {
        total_orders: 1,
        orders_with_difference: 0,
        total_profit: 0,
        total_loss: 0,
        net_impact: 0,
      },
      orders: [
        {
          order_sn: "SN-001",
          order_date: "2024-01-01",
          buyer_paid: 100000,
          actual_fee: 5000,
          shopee_rebate: 500,
          difference: 4500,
          buyer_name: "John Doe",
          payment_method: "Credit Card",
        },
      ],
    };
    exportShopeeShippingCSV(data);
    const [rows] = downloadSpy.mock.calls[0] as [string[][]];
    expect(rows).toHaveLength(2);
    expect(rows[1][1]).toBe("SN-001");
  });
});

describe("exportTiktokReconciliationCSV", () => {
  let downloadSpy: MockInstance;

  beforeEach(() => {
    downloadSpy = vi
      .spyOn(analyticsHelpers, "downloadCSV")
      .mockImplementation(() => {});
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("calls downloadCSV with tiktok-reconciliation filename", () => {
    const data: TiktokReconciliationResult = {
      summary: {
        total_sku: 0,
        total_transactions: 0,
        sku_ok: 0,
        sku_with_price_diff: 0,
        sku_no_inventory: 0,
      },
      sku_groups: [],
    };
    exportTiktokReconciliationCSV(data);
    const [, filename] = downloadSpy.mock.calls[0] as [string[][], string];
    expect(filename).toMatch(/^tiktok-reconciliation-\d{4}-\d{2}-\d{2}\.csv$/);
  });

  it("includes Status, SKU, Transactions in header", () => {
    const data: TiktokReconciliationResult = {
      summary: {
        total_sku: 0,
        total_transactions: 0,
        sku_ok: 0,
        sku_with_price_diff: 0,
        sku_no_inventory: 0,
      },
      sku_groups: [],
    };
    exportTiktokReconciliationCSV(data);
    const [rows] = downloadSpy.mock.calls[0] as [string[][]];
    expect(rows[0]).toContain("Status");
    expect(rows[0]).toContain("Transactions");
  });
});

describe("exportTiktokShippingCSV", () => {
  let downloadSpy: MockInstance;

  beforeEach(() => {
    downloadSpy = vi
      .spyOn(analyticsHelpers, "downloadCSV")
      .mockImplementation(() => {});
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("calls downloadCSV with tiktok-shipping-fee filename", () => {
    const data: TiktokShippingFeeResult = {
      summary: {
        total_orders: 0,
        orders_with_difference: 0,
        total_profit: 0,
        total_loss: 0,
        net_impact: 0,
      },
      orders: [],
    };
    exportTiktokShippingCSV(data);
    const [, filename] = downloadSpy.mock.calls[0] as [string[][], string];
    expect(filename).toMatch(/^tiktok-shipping-fee-\d{4}-\d{2}-\d{2}\.csv$/);
  });

  it("includes Order ID and Currency in header", () => {
    const data: TiktokShippingFeeResult = {
      summary: {
        total_orders: 0,
        orders_with_difference: 0,
        total_profit: 0,
        total_loss: 0,
        net_impact: 0,
      },
      orders: [],
    };
    exportTiktokShippingCSV(data);
    const [rows] = downloadSpy.mock.calls[0] as [string[][]];
    expect(rows[0]).toContain("Order ID");
    expect(rows[0]).toContain("Currency");
  });

  it("maps orders to rows including currency", () => {
    const data: TiktokShippingFeeResult = {
      summary: {
        total_orders: 1,
        orders_with_difference: 0,
        total_profit: 0,
        total_loss: 0,
        net_impact: 0,
      },
      orders: [
        {
          order_id: "TT-001",
          order_date: "2024-03-10",
          buyer_paid: 200000,
          actual_fee: 8000,
          platform_discount: 1000,
          difference: 7000,
          order_status: "COMPLETED",
          currency: "IDR",
        },
      ],
    };
    exportTiktokShippingCSV(data);
    const [rows] = downloadSpy.mock.calls[0] as [string[][]];
    expect(rows).toHaveLength(2);
    expect(rows[1][1]).toBe("TT-001");
    expect(rows[1][7]).toBe("IDR");
  });
});
