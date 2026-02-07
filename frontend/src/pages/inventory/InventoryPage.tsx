import { useState, useMemo } from "react";
import {
  Table,
  Input,
  Button,
  Popover,
  Checkbox,
  Space,
  Typography,
  Spin,
  Empty,
  Alert,
} from "antd";
import {
  SearchOutlined,
  SettingOutlined,
  ReloadOutlined,
} from "@ant-design/icons";
import type { ColumnsType } from "antd/es/table";
import { useQueryClient, useMutation } from "@tanstack/react-query";
import { InventoryRecord } from "../../types/product";
import {
  useInventory,
  useSelectedColumns,
  useAvailableColumns,
} from "../../hooks/useInventory";
import { saveSelectedColumns } from "../../api/inventory";

export default function InventoryPage() {
  const [searchText, setSearchText] = useState("");
  const queryClient = useQueryClient();

  const { data, isLoading, isError, error, refetch } = useInventory({
    search: searchText || undefined,
  });

  const { data: selectedCols = [] } = useSelectedColumns();
  const { data: availableCols = [] } = useAvailableColumns();

  const saveColumnsMutation = useMutation({
    mutationFn: saveSelectedColumns,
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ["inventory-columns-selected"],
      });
    },
  });

  const handleColumnToggle = (col: string, checked: boolean) => {
    const newCols = checked
      ? [...selectedCols, col]
      : selectedCols.filter((c) => c !== col);
    saveColumnsMutation.mutate(newCols);
  };

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

  if (isError) {
    return (
      <div style={{ padding: 24 }}>
        <Alert
          type="error"
          message="Failed to load inventory"
          description={(error as Error)?.message || "Unknown error"}
          action={
            <Button size="small" onClick={() => refetch()}>
              Retry
            </Button>
          }
        />
      </div>
    );
  }

  const records = data?.records || [];

  return (
    <div
      style={{
        padding: 24,
        height: "100%",
        display: "flex",
        flexDirection: "column",
      }}
    >
      <div
        style={{
          marginBottom: 16,
          display: "flex",
          justifyContent: "space-between",
        }}
      >
        <Typography.Title level={2} style={{ margin: 0 }}>
          Inventory
        </Typography.Title>
        <Space>
          <Input
            placeholder="Search"
            prefix={<SearchOutlined />}
            onChange={(e) => setSearchText(e.target.value)}
            style={{ width: 200 }}
          />
          <Button
            icon={<ReloadOutlined />}
            onClick={() => refetch()}
            loading={isLoading}
          >
            Refresh
          </Button>
          <Popover
            trigger="click"
            placement="bottomRight"
            title="Columns"
            content={
              <div
                style={{
                  display: "flex",
                  flexDirection: "column",
                  maxHeight: 300,
                  overflowY: "auto",
                }}
              >
                {availableCols.map((col) => (
                  <Checkbox
                    key={col}
                    checked={selectedCols.includes(col)}
                    onChange={(e) => handleColumnToggle(col, e.target.checked)}
                  >
                    {col}
                  </Checkbox>
                ))}
              </div>
            }
          >
            <Button icon={<SettingOutlined />}>Cols</Button>
          </Popover>
        </Space>
      </div>

      {isLoading ? (
        <div style={{ textAlign: "center", padding: 48 }}>
          <Spin size="large" />
          <div style={{ marginTop: 16 }}>Loading inventory...</div>
        </div>
      ) : records.length === 0 ? (
        <Empty description="No inventory items found" />
      ) : (
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
      )}
    </div>
  );
}
