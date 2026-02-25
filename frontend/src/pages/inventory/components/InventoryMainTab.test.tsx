import { render } from "@testing-library/react";
import type { ReactNode } from "react";
import { describe, expect, it, vi } from "vitest";
import { InventoryMainTab } from "./InventoryMainTab";
import type { InventoryRecord } from "@/types/inventory";
import type { MarketplaceAllocationSettings } from "../utils/marketplaceAllocation";

const capturedProps: Array<Record<string, unknown>> = [];

const marketplaceSettings: MarketplaceAllocationSettings = {
  keyColumn: "SKU",
  totalColumn: "Stock",
  rawTotalColumn: "Stock",
  autoColumn: "AUTO",
  shopeeRatio: 0.6,
  tiktokRatio: 0.3,
};

vi.mock("@/components/common/VirtualTable", () => ({
  VirtualTable: (props: Record<string, unknown>) => {
    capturedProps.push(props);
    return <div data-testid="virtual-table" />;
  },
}));

function createRecord(overrides?: Partial<InventoryRecord>): InventoryRecord {
  return {
    id: "1",
    key_value: "SKU-1",
    key_column_name: "SKU",
    data: {
      SKU: "SKU-1",
      SHOPEE: 10,
      TIKTOK: 6,
      LAZADA: 4,
      Stock: 20,
    },
    sync_status: "synced",
    platform_status: [],
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
    ...overrides,
  };
}

function getNodeText(node: ReactNode): string {
  const { container, unmount } = render(<span>{node}</span>);
  const text = container.textContent ?? "";
  unmount();
  return text;
}

describe("InventoryMainTab", () => {
  it("renders fixed marketplace columns and removes legacy Platforms column", () => {
    capturedProps.length = 0;

    render(
      <InventoryMainTab
        records={[createRecord()]}
        loading={false}
        error={null}
        onRetry={vi.fn()}
        visibleColumns={["SKU", "SHOPEE", "Stock"]}
        lockedColumns={[]}
        marketplaceSettings={marketplaceSettings}
        selectedRowKeys={[]}
        onSelectionChange={vi.fn()}
      />,
    );

    const latestProps = capturedProps[capturedProps.length - 1];
    const columns = (latestProps.columns as Array<{ title: string }>).map(
      (column) => column.title,
    );

    expect(columns).toContain("Shopee");
    expect(columns).toContain("TikTok");
    expect(columns).toContain("Lazada");
    expect(columns).not.toContain("Platforms");
    expect(columns.filter((title) => title === "Shopee")).toHaveLength(1);

    const rowSelection = latestProps.rowSelection as {
      columnWidth?: number;
      fixed?: boolean;
    };
    expect(rowSelection.columnWidth).toBe(56);
    expect(rowSelection.fixed).toBe(true);
    expect(latestProps.scroll).toEqual({ x: "max-content" });
  });

  it("keeps dynamic column order and binds values by column key", () => {
    capturedProps.length = 0;
    const record = createRecord({
      data: {
        Masuk: 12,
        "Nama Variasi": "Floral Fantasy",
        TOTAL: 0,
      },
    });

    render(
      <InventoryMainTab
        records={[record]}
        loading={false}
        error={null}
        onRetry={vi.fn()}
        visibleColumns={["Masuk", "Nama Variasi", "TOTAL"]}
        lockedColumns={[]}
        marketplaceSettings={marketplaceSettings}
        selectedRowKeys={[]}
        onSelectionChange={vi.fn()}
      />,
    );

    const latestProps = capturedProps[capturedProps.length - 1];
    const columns = latestProps.columns as Array<{
      title?: ReactNode;
      dataIndex?: string;
      render?: (value: unknown, row: InventoryRecord) => ReactNode;
    }>;

    const getColumnTitle = (column: { title?: ReactNode }) =>
      typeof column.title === "string" ? column.title : "";

    const dynamicColumns = columns.filter((column) =>
      ["Masuk", "Nama Variasi", "TOTAL"].includes(getColumnTitle(column)),
    );

    expect(dynamicColumns.map((column) => getColumnTitle(column))).toEqual([
      "Masuk",
      "Nama Variasi",
      "TOTAL",
    ]);

    const masukCell = dynamicColumns
      .find((column) => getColumnTitle(column) === "Masuk")
      ?.render?.(record.data.Masuk, record);
    const namaVariasiCell = dynamicColumns
      .find((column) => getColumnTitle(column) === "Nama Variasi")
      ?.render?.(record.data["Nama Variasi"], record);
    const totalCell = dynamicColumns
      .find((column) => getColumnTitle(column) === "TOTAL")
      ?.render?.(record.data.TOTAL, record);

    expect(getNodeText(masukCell)).toContain("12");
    expect(getNodeText(namaVariasiCell)).toContain("Floral Fantasy");
    expect(getNodeText(totalCell)).toContain("0");
  });

  it("keeps long text columns readable without ellipsis", () => {
    capturedProps.length = 0;
    const longName =
      "Baygon Semprot Aerosol Obat Anti Nyamuk Kecoa Serangga 200 ML";

    render(
      <InventoryMainTab
        records={[
          createRecord({
            data: {
              "Nama Barang": longName,
              "Nama Variasi": "Flower Garden 200ml",
            },
          }),
        ]}
        loading={false}
        error={null}
        onRetry={vi.fn()}
        visibleColumns={["Nama Barang", "Nama Variasi"]}
        lockedColumns={[]}
        marketplaceSettings={marketplaceSettings}
        selectedRowKeys={[]}
        onSelectionChange={vi.fn()}
      />,
    );

    const latestProps = capturedProps[capturedProps.length - 1];
    const columns = latestProps.columns as Array<{
      title?: ReactNode;
      ellipsis?: boolean;
      width?: number;
      render?: (value: unknown, row: InventoryRecord) => ReactNode;
    }>;

    const nameColumn = columns.find((column) => column.title === "Nama Barang");
    const variationColumn = columns.find(
      (column) => column.title === "Nama Variasi",
    );

    expect(nameColumn?.ellipsis).toBe(false);
    expect(variationColumn?.ellipsis).toBe(false);
    expect((nameColumn?.width ?? 0) >= 320).toBe(true);

    const renderedName = nameColumn?.render?.(
      longName,
      createRecord({
        data: { "Nama Barang": longName },
      }),
    );
    expect(getNodeText(renderedName)).toContain(longName);
  });
});
