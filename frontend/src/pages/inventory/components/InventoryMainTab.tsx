import { useMemo } from "react";
import type { Key } from "react";
import { Spin, Empty, Alert, Button, Tag } from "antd";
import type { ColumnsType } from "antd/es/table";
import { VirtualTable } from "@/components/common/VirtualTable";
import type { InventoryRecord } from "@/types/inventory";
import { StockCell } from "./StockCell";
import { PriceCell } from "./PriceCell";

interface Props {
  records: InventoryRecord[];
  loading: boolean;
  error: Error | null;
  onRetry: () => void;
  visibleColumns: string[];
  lockedColumns: string[];
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
  readOnly = false,
  selectedRowKeys = [],
  onSelectionChange,
}: Props) {
  const lockedColumnsSet = useMemo(
    () => new Set(lockedColumns),
    [lockedColumns],
  );

  const dynamicColumns: ColumnsType<InventoryRecord> = useMemo(() => {
    const estimateColumnWidth = (columnName: string): number => {
      const normalized = columnName.trim().toLowerCase();

      if (normalized === "nama barang") {
        return 360;
      }

      if (normalized === "nama variasi") {
        return 260;
      }

      if (normalized === "harga" || normalized === "price") {
        return 130;
      }

      if (normalized === "total") {
        return 90;
      }

      if (normalized === "stock" || normalized === "stok") {
        return 110;
      }

      const sampleRecords = records.slice(0, 120);
      const longestValueLength = sampleRecords.reduce((max, record) => {
        const value = record.data?.[columnName];
        const currentLength = value == null ? 0 : String(value).trim().length;
        return Math.max(max, currentLength);
      }, columnName.length);

      const estimated = longestValueLength * 8 + 40;
      return Math.min(280, Math.max(120, estimated));
    };

    const getRecordDataValue = (
      record: InventoryRecord,
      candidates: string[],
    ): unknown => {
      const rowData = record.data || {};
      for (const candidate of candidates) {
        if (candidate in rowData) {
          return rowData[candidate];
        }
      }

      const entry = Object.entries(rowData).find(
        ([columnName]) =>
          candidates.includes(columnName) ||
          candidates.includes(columnName.toLowerCase()) ||
          candidates.includes(columnName.toUpperCase()),
      );

      return entry?.[1];
    };

    const renderMarketplaceCell = (
      record: InventoryRecord,
      platformName: "shopee" | "tiktok" | "lazada",
    ) => {
      const value = getRecordDataValue(record, [
        platformName,
        platformName.toUpperCase(),
        platformName.charAt(0).toUpperCase() + platformName.slice(1),
      ]);

      if (
        value === null ||
        value === undefined ||
        String(value).trim() === ""
      ) {
        return <span style={{ color: "#999", fontSize: 11 }}>-</span>;
      }

      return (
        <Tag
          color="blue"
          style={{
            fontSize: 11,
            margin: 0,
            minWidth: 30,
            display: "inline-flex",
            justifyContent: "center",
            textAlign: "center",
          }}
        >
          {String(value)}
        </Tag>
      );
    };

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
        render: (_, record) => renderMarketplaceCell(record, "shopee"),
      },
      {
        title: "TikTok",
        key: "tiktok_status",
        width: 96,
        align: "center",
        render: (_, record) => renderMarketplaceCell(record, "tiktok"),
      },
      {
        title: "Lazada",
        key: "lazada_status",
        width: 96,
        align: "center",
        render: (_, record) => renderMarketplaceCell(record, "lazada"),
      },
      {
        title: "Sync Status",
        key: "sync_status",
        width: 120,
        align: "center",
        render: (_, record) => {
          const statusMap: Record<string, { color: string; label: string }> = {
            synced: { color: "green", label: "Synced" },
            not_synced: { color: "default", label: "Not Synced" },
            error: { color: "red", label: "Error" },
          };
          const status =
            statusMap[record.sync_status || "not_synced"] ||
            statusMap.not_synced;
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
      const dynamicWidth = estimateColumnWidth(colName);

      cols.push({
        title: colName,
        key: colName,
        width: dynamicWidth,
        ellipsis: !isLongTextColumn,
        render: (_, record) => {
          const val = record.data?.[colName];

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
  }, [records, visibleColumns, lockedColumnsSet, readOnly]);

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
