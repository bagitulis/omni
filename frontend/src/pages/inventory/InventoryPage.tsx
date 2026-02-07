import { useState, useMemo } from "react";
import {
  Table,
  Input,
  Select,
  Button,
  Popover,
  Checkbox,
  Tag,
  Space,
  Typography,
  Spin,
  Empty,
  Alert,
} from "antd";
import {
  SearchOutlined,
  SettingOutlined,
  WarningOutlined,
  ExclamationCircleOutlined,
  CheckCircleOutlined,
  ReloadOutlined,
} from "@ant-design/icons";
import type { ColumnsType } from "antd/es/table";
import { InventoryItem } from "../../types/product";
import { useInventory, useUpdateStock } from "../../hooks/useInventory";
import { StockCell } from "./components/StockCell";

export default function InventoryPage() {
  const [searchText, setSearchText] = useState("");
  const [statusFilter, setStatusFilter] = useState("all");
  const [editingId, setEditingId] = useState<string | null>(null);
  const [visibleColumns, setVisibleColumns] = useState<string[]>([
    "item_sku",
    "item_name",
    "available_stock",
    "status",
    "warehouse",
  ]);

  const { data, isLoading, isError, error, refetch } = useInventory({
    search: searchText || undefined,
  });

  const updateStockMutation = useUpdateStock();

  const handleStockUpdate = (sku: string, val: number) => {
    updateStockMutation.mutate({ sku, stock: val });
    setEditingId(null);
  };

  const handleCancelEdit = () => {
    setEditingId(null);
  };

  const items = data?.items || [];

  const filteredData = useMemo(
    () =>
      items.filter((item) => {
        const s = item.available_stock;
        const matchesFilter =
          statusFilter === "all" ||
          (statusFilter === "low" && s <= 10 && s > 0) ||
          (statusFilter === "out" && s === 0) ||
          (statusFilter === "ok" && s > 10);
        return matchesFilter;
      }),
    [items, statusFilter],
  );

  const allColumns: ColumnsType<InventoryItem> = [
    {
      title: "SKU",
      dataIndex: "item_sku",
      key: "item_sku",
      width: 120,
      fixed: "left",
    },
    {
      title: "Name",
      dataIndex: "item_name",
      key: "item_name",
      width: 250,
      ellipsis: true,
    },
    {
      title: "Available",
      dataIndex: "available_stock",
      key: "available_stock",
      width: 120,
      render: (v, r) => (
        <StockCell
          value={v}
          itemSku={r.item_sku}
          isEditing={editingId === r.item_id}
          onStartEdit={() => setEditingId(r.item_id)}
          onSave={handleStockUpdate}
          onCancel={handleCancelEdit}
        />
      ),
    },
    {
      title: "Status",
      key: "status",
      width: 120,
      render: (_, r) => {
        const s = r.available_stock;
        const cfg =
          s === 0
            ? {
                color: "error" as const,
                icon: <ExclamationCircleOutlined />,
                text: "Out",
              }
            : s <= 10
              ? {
                  color: "warning" as const,
                  icon: <WarningOutlined />,
                  text: "Low",
                }
              : {
                  color: "success" as const,
                  icon: <CheckCircleOutlined />,
                  text: "OK",
                };
        return (
          <Tag color={cfg.color} icon={cfg.icon}>
            {cfg.text}
          </Tag>
        );
      },
    },
    {
      title: "Warehouse",
      dataIndex: "warehouse",
      key: "warehouse",
      width: 100,
    },
    {
      title: "Updated",
      dataIndex: "last_updated",
      key: "last_updated",
      width: 160,
      render: (v) => (v ? new Date(v).toLocaleString() : "-"),
    },
  ];

  if (isError) {
    return (
      <div style={{ padding: 24 }}>
        <Alert
          type="error"
          message="Failed to load inventory"
          description={error?.message || "Unknown error"}
          action={
            <Button size="small" onClick={() => refetch()}>
              Retry
            </Button>
          }
        />
      </div>
    );
  }

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
          <Select
            defaultValue="all"
            onChange={setStatusFilter}
            style={{ width: 120 }}
            options={[
              { value: "all", label: "All" },
              { value: "ok", label: "In Stock" },
              { value: "low", label: "Low" },
              { value: "out", label: "Out" },
            ]}
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
              <div style={{ display: "flex", flexDirection: "column" }}>
                {allColumns.map((col) => (
                  <Checkbox
                    key={col.key}
                    checked={visibleColumns.includes(col.key as string)}
                    onChange={(e) =>
                      setVisibleColumns(
                        e.target.checked
                          ? [...visibleColumns, col.key as string]
                          : visibleColumns.filter((k) => k !== col.key),
                      )
                    }
                  >
                    {col.title as string}
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
      ) : filteredData.length === 0 ? (
        <Empty description="No inventory items found" />
      ) : (
        <Table
          virtual
          columns={allColumns.filter((c) =>
            visibleColumns.includes(c.key as string),
          )}
          dataSource={filteredData}
          rowKey="item_id"
          pagination={false}
          scroll={{ y: 600, x: 1000 }}
          size="small"
          bordered
        />
      )}
    </div>
  );
}
