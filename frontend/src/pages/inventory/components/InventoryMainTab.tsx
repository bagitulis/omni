import { useMemo } from "react";
import type { Key } from "react";
import { Spin, Empty, Alert, Button, Tag } from "antd";
import type { ColumnsType } from "antd/es/table";
import { VirtualTable } from "@/components/common/VirtualTable";
import type { InventoryRecord } from "@/types/inventory";
import {
  resolveMarketplaceAllocationForRecord,
  type MarketplaceAllocationSettings,
} from "../utils/marketplaceAllocation";
import {
  estimateInventoryColumnWidth,
  renderMarketplaceCell,
  resolveSyncStatus,
} from "./inventoryMainTabHelpers";
import { StockCell } from "./StockCell";
import { PriceCell } from "./PriceCell";

interface Props {
  records: InventoryRecord[];
  loading: boolean;
  error: Error | null;
  onRetry: () => void;
  visibleColumns: string[];
  lockedColumns: string[];
  marketplaceSettings: MarketplaceAllocationSettings;
  readOnly?: boolean;
  selectedRowKeys?: Key[];
  onSelectionChange?: (keys: Key[], rows: InventoryRecord[]) => void;
}

export function InventoryMainTab({
  records,
  loading,
  error,
  onRetry,
  visibleColumns,
  lockedColumns,
  marketplaceSettings,
  readOnly = false,
  selectedRowKeys = [],
  onSelectionChange,
}: Props) {
  const lockedColumnsSet = useMemo(
    () => new Set(lockedColumns),
    [lockedColumns],
  );

  const dynamicColumns: ColumnsType<InventoryRecord> = useMemo(() => {
    const cols: ColumnsType<InventoryRecord> = [
      {
        title: "Key",
        dataIndex: "key_value",
        key: "key_value",
        width: 180,
        fixed: "left" as const,
      },
      {
        title: "Shopee",
        key: "shopee_status",
        width: 96,
        align: "center",
        render: (_, record) =>
          renderMarketplaceCell(record, "shopee", marketplaceSettings),
      },
      {
        title: "TikTok",
        key: "tiktok_status",
        width: 96,
        align: "center",
        render: (_, record) =>
          renderMarketplaceCell(record, "tiktok", marketplaceSettings),
      },
      {
        title: "Lazada",
        key: "lazada_status",
        width: 96,
        align: "center",
        render: (_, record) =>
          renderMarketplaceCell(record, "lazada", marketplaceSettings),
      },
      {
        title: "Sync Status",
        key: "sync_status",
        width: 120,
        align: "center",
        render: (_, record) => {
          const status = resolveSyncStatus(record.sync_status);
          return (
            <Tag color={status.color} style={{ fontSize: 11, margin: 0 }}>
              {status.label}
            </Tag>
          );
        },
      },
    ];

    for (const colName of visibleColumns) {
      const normalizedColumnName = colName.trim().toLowerCase();
      if (
        normalizedColumnName === "shopee" ||
        normalizedColumnName === "tiktok" ||
        normalizedColumnName === "lazada"
      ) {
        continue;
      }
      const isStock =
        normalizedColumnName === "stock" || normalizedColumnName === "stok";
      const isPrice =
        normalizedColumnName === "price" || normalizedColumnName === "harga";
      const isLongTextColumn =
        normalizedColumnName.includes("nama barang") ||
        normalizedColumnName.includes("nama variasi") ||
        normalizedColumnName.includes("name") ||
        normalizedColumnName.includes("product");
      const isLocked = lockedColumnsSet.has(colName);
      const dynamicWidth = estimateInventoryColumnWidth(colName, records);

      cols.push({
        title: colName,
        key: colName,
        width: dynamicWidth,
        ellipsis: !isLongTextColumn,
        render: (_, record) => {
          const val = record.data?.[colName];
          const allocation = resolveMarketplaceAllocationForRecord(
            record.data || {},
            marketplaceSettings,
          );

          if (isStock) {
            if (readOnly) {
              return String(Number(val) || 0);
            }
            return (
              <StockCell
                record={record}
                dataIndex={colName}
                value={val}
                locked={isLocked}
              />
            );
          }
          if (isPrice) {
            if (readOnly) {
              return new Intl.NumberFormat("id-ID", {
                style: "currency",
                currency: "IDR",
                minimumFractionDigits: 0,
                maximumFractionDigits: 0,
              }).format(Number(val) || 0);
            }
            return (
              <PriceCell
                record={record}
                dataIndex={colName}
                value={val}
                locked={isLocked}
              />
            );
          }

          if (
            normalizedColumnName === "total" &&
            (val === null || val === undefined || String(val).trim() === "")
          ) {
            return String(allocation.total);
          }

          if (val == null || String(val).trim() === "") {
            return "-";
          }

          const text = String(val);
          if (!isLongTextColumn) {
            return text;
          }

          return (
            <span
              style={{
                display: "block",
                whiteSpace: "normal",
                wordBreak: "break-word",
                lineHeight: 1.45,
              }}
            >
              {text}
            </span>
          );
        },
      });
    }

    cols.push({
      title: "Updated",
      dataIndex: "updated_at",
      key: "updated_at",
      width: 160,
      render: (v: string) => (v ? new Date(v).toLocaleString() : "-"),
    });

    return cols;
  }, [
    lockedColumnsSet,
    marketplaceSettings,
    readOnly,
    records,
    visibleColumns,
  ]);

  if (error) {
    return (
      <div style={{ padding: 24 }}>
        <Alert
          type="error"
          message="Failed to load inventory"
          description={error.message || "Unknown error"}
          action={
            <Button size="small" onClick={onRetry}>
              Retry
            </Button>
          }
        />
      </div>
    );
  }

  if (loading) {
    return (
      <div style={{ textAlign: "center", padding: 48 }}>
        <Spin size="large" />
        <div style={{ marginTop: 16 }}>Loading inventory...</div>
      </div>
    );
  }

  if (records.length === 0) {
    return <Empty description="No inventory items found" />;
  }

  const rowSelection =
    readOnly || !onSelectionChange
      ? undefined
      : {
          columnWidth: 56,
          fixed: true as const,
          selectedRowKeys,
          onChange: onSelectionChange,
        };

  return (
    <VirtualTable<InventoryRecord>
      columns={dynamicColumns}
      dataSource={records}
      rowKey="id"
      rowSelection={rowSelection}
      pagination={false}
      scroll={{ x: "max-content" }}
      size="small"
      bordered
      enableVirtual={records.length > 50}
      offsetBottom={420}
    />
  );
}
