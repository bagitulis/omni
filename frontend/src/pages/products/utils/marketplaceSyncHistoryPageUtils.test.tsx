import "@testing-library/jest-dom";
import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import type { MarketplaceSyncHistoryEntry } from "@/types/shared";
import {
  toPrettyPayload,
  createMarketplaceSyncHistoryColumns,
  isSyncHistoryRowExpandable,
  renderSyncHistoryExpandedRow,
} from "./marketplaceSyncHistoryPageUtils";

// Mock PlatformIndicator to avoid extra hook deps
vi.mock("@/components/shared/PlatformIndicator", () => ({
  PlatformIndicator: ({ data }: { data: { platform: string } }) => (
    <span data-testid="platform-indicator">{data.platform}</span>
  ),
}));

Object.defineProperty(window, "matchMedia", {
  writable: true,
  value: vi.fn().mockImplementation((query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
});

function makeEntry(
  overrides: Partial<MarketplaceSyncHistoryEntry> = {},
): MarketplaceSyncHistoryEntry {
  return {
    id: "entry-1",
    tenant_id: "tenant-1",
    sku: "SKU-001",
    platform: "shopee",
    operation: "stock_update",
    status: "success",
    created_at: "2026-01-15T10:30:00Z",
    ...overrides,
  };
}

describe("toPrettyPayload", () => {
  it("returns '-' for null/undefined/empty", () => {
    expect(toPrettyPayload(null)).toBe("-");
    expect(toPrettyPayload(undefined)).toBe("-");
    expect(toPrettyPayload("")).toBe("-");
  });

  it("pretty-prints a JSON string", () => {
    const input = '{"key":"value"}';
    const result = toPrettyPayload(input);
    expect(result).toBe(JSON.stringify({ key: "value" }, null, 2));
  });

  it("returns raw string when JSON parse fails", () => {
    expect(toPrettyPayload("not-json")).toBe("not-json");
  });

  it("serializes an object", () => {
    const obj = { a: 1, b: "hello" };
    expect(toPrettyPayload(obj)).toBe(JSON.stringify(obj, null, 2));
  });
});

describe("isSyncHistoryRowExpandable", () => {
  it("returns false when no request/response/error", () => {
    const entry = makeEntry();
    expect(isSyncHistoryRowExpandable(entry)).toBe(false);
  });

  it("returns true when request_data present", () => {
    const entry = makeEntry({ request_data: '{"sku":"A"}' });
    expect(isSyncHistoryRowExpandable(entry)).toBe(true);
  });

  it("returns true when response_data present", () => {
    const entry = makeEntry({ response_data: '{"status":"ok"}' });
    expect(isSyncHistoryRowExpandable(entry)).toBe(true);
  });

  it("returns true when error_message present", () => {
    const entry = makeEntry({ error_message: "rate limit exceeded" });
    expect(isSyncHistoryRowExpandable(entry)).toBe(true);
  });
});

describe("renderSyncHistoryExpandedRow", () => {
  it("renders Request Data and Response Data labels", () => {
    const entry = makeEntry({
      request_data: '{"sku":"A"}',
      response_data: '{"result":"ok"}',
    });
    render(<>{renderSyncHistoryExpandedRow(entry)}</>);
    expect(screen.getByText("Request Data")).toBeInTheDocument();
    expect(screen.getByText("Response Data")).toBeInTheDocument();
  });

  it("shows error alert when error_message is set", () => {
    const entry = makeEntry({ error_message: "platform error occurred" });
    render(<>{renderSyncHistoryExpandedRow(entry)}</>);
    expect(screen.getByText("Platform Error")).toBeInTheDocument();
    expect(screen.getByText("platform error occurred")).toBeInTheDocument();
  });

  it("shows Empty when no data at all", () => {
    const entry = makeEntry({
      request_data: undefined,
      response_data: undefined,
      error_message: undefined,
    });
    render(<>{renderSyncHistoryExpandedRow(entry)}</>);
    expect(screen.getByText("No details available")).toBeInTheDocument();
  });
});

describe("createMarketplaceSyncHistoryColumns", () => {
  const columns = createMarketplaceSyncHistoryColumns();

  it("returns 7 columns", () => {
    expect(columns).toHaveLength(7);
  });

  it("has expected column keys", () => {
    const keys = columns.map((c) => c.key);
    expect(keys).toContain("created_at");
    expect(keys).toContain("operation");
    expect(keys).toContain("platform");
    expect(keys).toContain("sku");
    expect(keys).toContain("status");
    expect(keys).toContain("items_count");
    expect(keys).toContain("details");
  });

  it("renders operation tag", () => {
    const operationCol = columns.find((c) => c.key === "operation");
    const renderFn = operationCol?.render as (value: string) => React.ReactNode;
    render(<>{renderFn("stock_update")}</>);
    expect(screen.getByText("STOCK UPDATE")).toBeInTheDocument();
  });

  it("renders sku as code", () => {
    const skuCol = columns.find((c) => c.key === "sku");
    const renderFn = skuCol?.render as (value: string) => React.ReactNode;
    const { container } = render(<>{renderFn("TEST-SKU")}</>);
    const codeEl = container.querySelector("code");
    expect(codeEl).toBeTruthy();
    expect(codeEl?.textContent).toBe("TEST-SKU");
  });

  it("renders status badge", () => {
    const statusCol = columns.find((c) => c.key === "status");
    const renderFn = statusCol?.render as (value: string) => React.ReactNode;
    render(<>{renderFn("success")}</>);
    expect(screen.getByText("SUCCESS")).toBeInTheDocument();
  });

  it("renders items_count — defaults to 1 when no count fields", () => {
    const itemsCountCol = columns.find((c) => c.key === "items_count");
    const renderFn = itemsCountCol?.render as (
      _: unknown,
      record: MarketplaceSyncHistoryEntry,
    ) => React.ReactNode;
    const entry = makeEntry();
    render(<>{renderFn(undefined, entry)}</>);
    expect(screen.getByText("1")).toBeInTheDocument();
  });

  it("renders items_count from response_data.items_count", () => {
    const itemsCountCol = columns.find((c) => c.key === "items_count");
    const renderFn = itemsCountCol?.render as (
      _: unknown,
      record: MarketplaceSyncHistoryEntry,
    ) => React.ReactNode;
    const entry = makeEntry({ response_data: '{"items_count":5}' });
    render(<>{renderFn(undefined, entry)}</>);
    expect(screen.getByText("5")).toBeInTheDocument();
  });

  it("renders details with error text when error_message set", () => {
    const detailsCol = columns.find((c) => c.key === "details");
    const renderFn = detailsCol?.render as (
      _: unknown,
      record: MarketplaceSyncHistoryEntry,
    ) => React.ReactNode;
    const entry = makeEntry({ error_message: "bad request" });
    render(<>{renderFn(undefined, entry)}</>);
    expect(screen.getByText("bad request")).toBeInTheDocument();
  });
});
