/**
 * Type shape tests for orderTypes.ts
 * These tests verify that exported types and the union type have the expected runtime structure.
 * TypeScript compilation of these tests ensures type correctness at build time.
 */
import { describe, it, expect } from "vitest";
import type {
  OrderTab,
  GetOrdersParams,
  BulkPrintLabelsResponse,
  BulkPrintLabelsOptions,
  CancelOrderParams,
  ShipOrderParams,
  LazadaDocumentResponse,
} from "./orderTypes";

describe("OrderTab union type", () => {
  it("accepts all valid tab values", () => {
    const tabs: OrderTab[] = [
      "unpaid",
      "unprocess",
      "processed",
      "shipped",
      "completed",
      "cancelled",
      "locked",
      "today",
    ];
    expect(tabs).toHaveLength(8);
    expect(tabs).toContain("unpaid");
    expect(tabs).toContain("today");
  });
});

describe("GetOrdersParams type shape", () => {
  it("accepts empty object (all fields optional)", () => {
    const params: GetOrdersParams = {};
    expect(params).toEqual({});
  });

  it("accepts all optional fields", () => {
    const params: GetOrdersParams = {
      page: 1,
      pageSize: 20,
      status: "shipped",
      platform: "shopee",
      search: "test-order",
      startDate: "2024-01-01",
      endDate: "2024-01-31",
    };
    expect(params.page).toBe(1);
    expect(params.pageSize).toBe(20);
    expect(params.status).toBe("shipped");
    expect(params.platform).toBe("shopee");
    expect(params.search).toBe("test-order");
    expect(params.startDate).toBe("2024-01-01");
    expect(params.endDate).toBe("2024-01-31");
  });
});

describe("BulkPrintLabelsResponse type shape", () => {
  it("accepts required labels and failed arrays plus count", () => {
    const response: BulkPrintLabelsResponse = {
      labels: [
        { order_sn: "SN-001", file_data: "base64data", status: "success" },
      ],
      failed: [{ order_sn: "SN-002", error: "Print failed" }],
      count: 2,
    };
    expect(response.labels).toHaveLength(1);
    expect(response.labels[0].order_sn).toBe("SN-001");
    expect(response.labels[0].file_data).toBe("base64data");
    expect(response.labels[0].status).toBe("success");
    expect(response.failed).toHaveLength(1);
    expect(response.failed[0].error).toBe("Print failed");
    expect(response.count).toBe(2);
  });

  it("accepts empty arrays", () => {
    const response: BulkPrintLabelsResponse = {
      labels: [],
      failed: [],
      count: 0,
    };
    expect(response.labels).toHaveLength(0);
    expect(response.failed).toHaveLength(0);
    expect(response.count).toBe(0);
  });
});

describe("BulkPrintLabelsOptions type shape", () => {
  it("accepts empty object (all fields optional)", () => {
    const opts: BulkPrintLabelsOptions = {};
    expect(opts).toEqual({});
  });

  it("accepts all optional fields including tiktok_document_type", () => {
    const opts: BulkPrintLabelsOptions = {
      platform: "tiktok",
      include_products: true,
      tiktok_document_type: "SHIPPING_LABEL",
    };
    expect(opts.platform).toBe("tiktok");
    expect(opts.include_products).toBe(true);
    expect(opts.tiktok_document_type).toBe("SHIPPING_LABEL");
  });

  it("accepts all tiktok_document_type variants", () => {
    const v1: BulkPrintLabelsOptions = {
      tiktok_document_type: "SHIPPING_LABEL",
    };
    const v2: BulkPrintLabelsOptions = {
      tiktok_document_type: "PACKING_SLIP",
    };
    const v3: BulkPrintLabelsOptions = {
      tiktok_document_type: "SHIPPING_LABEL_AND_PACKING_SLIP",
    };
    expect(v1.tiktok_document_type).toBe("SHIPPING_LABEL");
    expect(v2.tiktok_document_type).toBe("PACKING_SLIP");
    expect(v3.tiktok_document_type).toBe("SHIPPING_LABEL_AND_PACKING_SLIP");
  });
});

describe("CancelOrderParams type shape", () => {
  it("accepts required fields", () => {
    const params: CancelOrderParams = {
      order_no: "ORD-001",
      platform: "shopee",
      cancel_reason: "Out of stock",
    };
    expect(params.order_no).toBe("ORD-001");
    expect(params.platform).toBe("shopee");
    expect(params.cancel_reason).toBe("Out of stock");
  });

  it("accepts optional fields", () => {
    const params: CancelOrderParams = {
      order_no: "ORD-001",
      platform: "shopee",
      cancel_reason: "Out of stock",
      reason_detail: "Item not available",
      order_item_id: "ITEM-123",
    };
    expect(params.reason_detail).toBe("Item not available");
    expect(params.order_item_id).toBe("ITEM-123");
  });
});

describe("ShipOrderParams type shape", () => {
  it("accepts required fields", () => {
    const params: ShipOrderParams = {
      order_no: "ORD-001",
      platform: "lazada",
      shipping_provider: "JNE",
    };
    expect(params.order_no).toBe("ORD-001");
    expect(params.platform).toBe("lazada");
    expect(params.shipping_provider).toBe("JNE");
  });

  it("accepts all optional fields", () => {
    const params: ShipOrderParams = {
      order_no: "ORD-001",
      platform: "shopee",
      shipping_provider: "SPX",
      tracking_number: "TRK-12345",
      address_id: 42,
      pickup_time_id: "PT-99",
      branch_id: 7,
      package_id: "PKG-001",
      order_item_ids: ["ITEM-A", "ITEM-B"],
    };
    expect(params.tracking_number).toBe("TRK-12345");
    expect(params.address_id).toBe(42);
    expect(params.pickup_time_id).toBe("PT-99");
    expect(params.branch_id).toBe(7);
    expect(params.package_id).toBe("PKG-001");
    expect(params.order_item_ids).toEqual(["ITEM-A", "ITEM-B"]);
  });
});

describe("LazadaDocumentResponse type shape", () => {
  it("accepts empty object (all fields optional)", () => {
    const response: LazadaDocumentResponse = {};
    expect(response.document).toBeUndefined();
  });

  it("accepts nested document with file, url, mime_type", () => {
    const response: LazadaDocumentResponse = {
      document: {
        file: "base64content",
        url: "https://example.com/doc.pdf",
        mime_type: "application/pdf",
      },
    };
    expect(response.document?.file).toBe("base64content");
    expect(response.document?.url).toBe("https://example.com/doc.pdf");
    expect(response.document?.mime_type).toBe("application/pdf");
  });

  it("accepts empty document object", () => {
    const response: LazadaDocumentResponse = { document: {} };
    expect(response.document).toEqual({});
  });
});
