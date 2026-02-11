import { useMemo } from "react";
import { Table, Spin, Empty, Alert, Button } from "antd";
import type { ColumnsType } from "antd/es/table";
import { useSelectedColumns } from "@/hooks/useInventory";
import { InventoryRecord } from "@/types/inventory";

interface Props {
  records: InventoryRecord[];
  loading: boolean;
  error: Error | null;
  onRetry: () => void;
}

export function InventoryMainTab({ records, loading, error, onRetry }: Props) {
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
    ];

    for (const colName of selectedCols) {
      cols.push({
        title: colName,
        key: colName,
        width: 150,
        ellipsis: true,
        render: (_, record) => {
          const val = record.data?.[colName];
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
    <Table
      virtual
      columns={dynamicColumns}
      dataSource={records}
      rowKey="id"
      pagination={false}
      scroll={{ y: 600, x: 1000 }}
      size="small"
      bordered
    />
  );
}
