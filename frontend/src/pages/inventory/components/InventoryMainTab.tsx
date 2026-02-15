import { useMemo } from "react";
import type { Key } from "react";
import { Spin, Empty, Alert, Button, Space, Tag } from "antd";
import type { ColumnsType } from "antd/es/table";
import { VirtualTable } from "@/components/common/VirtualTable";
import { useSelectedColumns } from "@/hooks/useInventory";
import type { InventoryRecord } from "@/types/inventory";
import { StockCell } from "./StockCell";
import { PriceCell } from "./PriceCell";

interface Props {
  records: InventoryRecord[];
  loading: boolean;
  error: Error | null;
  onRetry: () => void;
  selectedRowKeys: Key[];
  onSelectionChange: (keys: Key[], rows: InventoryRecord[]) => void;
}

export function InventoryMainTab({
  records,
  loading,
  error,
  onRetry,
  selectedRowKeys,
  onSelectionChange,
}: Props) {
  const { data: selectedCols = [] } = useSelectedColumns();

  const dynamicColumns: ColumnsType<InventoryRecord> = useMemo(() => {
    const cols: ColumnsType<InventoryRecord> = [
      {
        title: "Key",
        dataIndex: "key_value",
        key: "key_value",
        width: 150,
        fixed: "left" as const,
      },
      {
        title: "Platforms",
        key: "platforms",
        width: 140,
        render: (_, record) => {
          const platforms = record.platform_status || [];
          if (platforms.length === 0) {
            return (
              <span style={{ color: "#999", fontSize: 11 }}>Not Listed</span>
            );
          }
          return (
            <Space size={4} wrap>
              {platforms.map((ps) => (
                <Tag
                  key={ps.platform}
                  color={ps.status === "active" ? "green" : "red"}
                  style={{ fontSize: 10, margin: 0, padding: "0 4px" }}
                >
                  {ps.platform.charAt(0).toUpperCase() + ps.platform.slice(1)}
                </Tag>
              ))}
            </Space>
          );
        },
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

    for (const colName of selectedCols) {
      const lowerName = colName.toLowerCase();
      const isStock = lowerName === "stock" || lowerName === "stok";
      const isPrice = lowerName === "price" || lowerName === "harga";

      cols.push({
        title: colName,
        key: colName,
        width: 150,
        ellipsis: true,
        render: (_, record) => {
          const val = record.data?.[colName];

          if (isStock) {
            return (
              <StockCell record={record} dataIndex={colName} value={val} />
            );
          }
          if (isPrice) {
            return (
              <PriceCell record={record} dataIndex={colName} value={val} />
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
  }, [selectedCols]);

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
