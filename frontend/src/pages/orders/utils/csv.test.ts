import { describe, expect, it, vi, beforeEach, afterEach } from "vitest";
import type { MockInstance } from "vitest";
import type { Order } from "@/types/order";
import { generateOrdersCSV, downloadCSV } from "./csv";

function makeOrder(overrides: Partial<Order> = {}): Order {
  return {
    id: "1",
    order_sn: "ORD-001",
    order_no: "ORD-001",
    order_status: "unprocess",
    status: "unprocess",
    platform: "shopee",
    category: "default",
    buyer_username: "buyer1",
    total_amount: 15000,
    currency: "IDR",
    payment_method: "cod",
    shipping_carrier: "JNE",
    ship_by_date: 0,
    sku: "SKU-001",
    product_name: "Test Product",
    variation_name: "Red",
    qty: 1,
    price: 15000,
    product_image: "",
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
    ...overrides,
  };
}

describe("generateOrdersCSV", () => {
  it("returns empty string when orders is empty", () => {
    expect(generateOrdersCSV([])).toBe("");
  });

  it("returns empty string when orders is null/undefined", () => {
    expect(generateOrdersCSV(null)).toBe("");
    expect(generateOrdersCSV(undefined)).toBe("");
  });

  it("generates CSV with correct headers", () => {
    const csv = generateOrdersCSV([makeOrder()]);
    const firstLine = csv.split("\n")[0];
    expect(firstLine).toBe("Order No,Platform,Status,Customer,Total,Date");
  });

  it("generates correct data row", () => {
    const order = makeOrder({
      order_sn: "SHOP-123",
      platform: "shopee",
      status: "shipped",
      buyer_username: "testbuyer",
      total_amount: 25000,
      created_at: "2026-01-15T10:00:00Z",
    });

    const csv = generateOrdersCSV([order]);
    const lines = csv.split("\n");
    expect(lines).toHaveLength(2);
    const dataRow = lines[1];
    expect(dataRow).toContain('"SHOP-123"');
    expect(dataRow).toContain('"shopee"');
    expect(dataRow).toContain('"shipped"');
    expect(dataRow).toContain('"testbuyer"');
    expect(dataRow).toContain("25000.00");
    expect(dataRow).toContain('"2026-01-15T10:00:00Z"');
  });

  it("generates multiple rows for multiple orders", () => {
    const orders = [
      makeOrder({ order_sn: "ORD-001" }),
      makeOrder({ order_sn: "ORD-002" }),
    ];
    const csv = generateOrdersCSV(orders);
    const lines = csv.split("\n");
    expect(lines).toHaveLength(3); // header + 2 rows
  });

  it("handles missing optional fields gracefully", () => {
    const order = makeOrder({
      order_sn: "",
      platform: "",
      status: "",
      buyer_username: "",
      total_amount: 0,
      created_at: "",
    });
    const csv = generateOrdersCSV([order]);
    const lines = csv.split("\n");
    const dataRow = lines[1];
    expect(dataRow).toContain('""');
    expect(dataRow).toContain("0.00");
  });

  it("formats total_amount to 2 decimal places", () => {
    const order = makeOrder({ total_amount: 12345.6789 });
    const csv = generateOrdersCSV([order]);
    expect(csv).toContain("12345.68");
  });

  it("quotes order_sn to preserve numeric-looking values", () => {
    const order = makeOrder({ order_sn: "1234567890" });
    const csv = generateOrdersCSV([order]);
    expect(csv).toContain('"1234567890"');
  });
});

describe("downloadCSV", () => {
  let appendChildSpy: MockInstance;
  let removeChildSpy: MockInstance;
  let createElementSpy: MockInstance;
  let createObjectURLSpy: MockInstance;
  let revokeObjectURLSpy: MockInstance;
  let clickMock: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    clickMock = vi.fn();

    if (!Object.prototype.hasOwnProperty.call(URL, "createObjectURL")) {
      Object.defineProperty(URL, "createObjectURL", {
        configurable: true,
        writable: true,
        value: () => "",
      });
    }
    if (!Object.prototype.hasOwnProperty.call(URL, "revokeObjectURL")) {
      Object.defineProperty(URL, "revokeObjectURL", {
        configurable: true,
        writable: true,
        value: () => undefined,
      });
    }

    const fakeLink = {
      setAttribute: vi.fn(),
      click: clickMock,
      style: { visibility: "visible" },
    };
    createElementSpy = vi
      .spyOn(document, "createElement")
      .mockReturnValue(fakeLink as unknown as HTMLElement);
    appendChildSpy = vi
      .spyOn(document.body, "appendChild")
      .mockReturnValue(fakeLink as unknown as Node);
    removeChildSpy = vi
      .spyOn(document.body, "removeChild")
      .mockReturnValue(fakeLink as unknown as Node);
    createObjectURLSpy = vi
      .spyOn(URL, "createObjectURL")
      .mockReturnValue("blob:fake-url");
    revokeObjectURLSpy = vi
      .spyOn(URL, "revokeObjectURL")
      .mockImplementation(() => undefined);
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("creates a blob and triggers a download", () => {
    downloadCSV("col1,col2\nval1,val2", "test.csv");

    expect(createElementSpy).toHaveBeenCalledWith("a");
    expect(createObjectURLSpy).toHaveBeenCalledOnce();
    expect(clickMock).toHaveBeenCalledOnce();
    expect(revokeObjectURLSpy).toHaveBeenCalledWith("blob:fake-url");
  });

  it("sets the correct filename on the link", () => {
    const fakeLink = {
      setAttribute: vi.fn(),
      click: vi.fn(),
      style: { visibility: "visible" },
    };
    createElementSpy.mockReturnValue(fakeLink as unknown as HTMLElement);
    appendChildSpy.mockReturnValue(fakeLink as unknown as Node);
    removeChildSpy.mockReturnValue(fakeLink as unknown as Node);

    downloadCSV("data", "orders-export.csv");

    expect(fakeLink.setAttribute).toHaveBeenCalledWith(
      "download",
      "orders-export.csv",
    );
  });

  it("appends and removes link from document body", () => {
    downloadCSV("data", "test.csv");
    expect(appendChildSpy).toHaveBeenCalledOnce();
    expect(removeChildSpy).toHaveBeenCalledOnce();
  });
});
