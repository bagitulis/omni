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
  selectedRowKeys: Key[];
  onSelectionChange: (keys: Key[], rows: InventoryRecord[]) => void;
}

export function InventoryMainTab({
  records,
  loading,
  error,
  onRetry,
  visibleColumns,
  lockedColumns,
  selectedRowKeys,
  onSelectionChange,
}: Props) {
  const lockedColumnsSet = useMemo(
    () => new Set(lockedColumns),
    [lockedColumns],
  );

  const dynamicColumns: ColumnsType<InventoryRecord> = useMemo(() => {
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
        <Tag color="blue" style={{ fontSize: 11 }}>
          {String(value)}
        </Tag>
      );
    };

    const cols: ColumnsType<InventoryRecord> = [
      {
        title: "Key",
        dataIndex: "key_value",
        key: "key_value",
        width: 150,
        fixed: "left" as const,
      },
      {
        title: "Shopee",
        key: "shopee_status",
        width: 130,
        render: (_, record) => renderMarketplaceCell(record, "shopee"),
      },
      {
        title: "TikTok",
        key: "tiktok_status",
        width: 130,
        render: (_, record) => renderMarketplaceCell(record, "tiktok"),
      },
      {
        title: "Lazada",
        key: "lazada_status",
        width: 130,
        render: (_, record) => renderMarketplaceCell(record, "lazada"),
      },
      {
        title: "Sync Status",
        key: "sync_status",
        width: 100,
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
            <Tag color={status.color} style={{ fontSize: 11 }}>
              {status.label}
            </Tag>
          );
        },
      },
    ];

    for (const colName of visibleColumns) {
      const lowerName = colName.toLowerCase();
      if (
        lowerName === "shopee" ||
        lowerName === "tiktok" ||
        lowerName === "lazada"
      ) {
        continue;
      }
      const isStock = lowerName === "stock" || lowerName === "stok";
      const isPrice = lowerName === "price" || lowerName === "harga";
      const isLocked = lockedColumnsSet.has(colName);

      cols.push({
        title: colName,
        key: colName,
        width: 150,
        ellipsis: true,
        render: (_, record) => {
          const val = record.data?.[colName];

          if (isStock) {
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
            return (
              <PriceCell
                record={record}
                dataIndex={colName}
                value={val}
                locked={isLocked}
              />
            );
          }

          return val != null ? String(val) : "-";
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
  }, [visibleColumns, lockedColumnsSet]);

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

  return (
    <VirtualTable<InventoryRecord>
      columns={dynamicColumns}
      dataSource={records}
      rowKey="id"
      rowSelection={{
        selectedRowKeys,
        onChange: onSelectionChange,
      }}
      pagination={false}
      scroll={{ x: 1000 }}
      size="small"
      bordered
      enableVirtual={records.length > 50}
      offsetBottom={420}
    />
  );
}
