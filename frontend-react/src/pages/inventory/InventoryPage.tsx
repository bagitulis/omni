import { useState, useMemo, useEffect, useRef } from "react";
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
  InputNumber,
  message,
} from "antd";
import {
  SearchOutlined,
  SettingOutlined,
  WarningOutlined,
  ExclamationCircleOutlined,
  CheckCircleOutlined,
} from "@ant-design/icons";
import type { ColumnsType } from "antd/es/table";
import { InventoryItem } from "../../types/product";

// Inline editable stock cell component
interface StockCellProps {
  value: number;
  itemId: string;
  isEditing: boolean;
  onStartEdit: () => void;
  onSave: (id: string, val: number) => void;
  onCancel: () => void;
}

function StockCell({
  value,
  itemId,
  isEditing,
  onStartEdit,
  onSave,
  onCancel,
}: StockCellProps) {
  const [editValue, setEditValue] = useState(value);
  const wrapperRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (isEditing) {
      setEditValue(value);
      // Focus the input inside wrapper after render
      setTimeout(() => {
        const input = wrapperRef.current?.querySelector("input");
        input?.focus();
        input?.select();
      }, 0);
    }
  }, [isEditing, value]);

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Escape") {
      e.preventDefault();
      onCancel();
    } else if (e.key === "Enter") {
      e.preventDefault();
      onSave(itemId, editValue);
    }
  };

  if (!isEditing) {
    return (
      <div
        onDoubleClick={onStartEdit}
        style={{
          cursor: "pointer",
          padding: "4px 8px",
          borderRadius: 3,
          minHeight: 24,
          display: "flex",
          alignItems: "center",
        }}
        title="Double-click to edit"
      >
        {value}
      </div>
    );
  }

  return (
    <div ref={wrapperRef}>
      <InputNumber
        value={editValue}
        min={0}
        onChange={(val) => setEditValue(val ?? 0)}
        onKeyDown={handleKeyDown}
        onBlur={() => onSave(itemId, editValue)}
        style={{
          width: "100%",
          borderColor: "#0369a1",
          borderRadius: 3,
        }}
      />
    </div>
  );
}

const generateMockData = (count: number): InventoryItem[] =>
  Array.from({ length: count }, (_, i) => ({
    item_id: `inv_${i}`,
    item_sku: `SKU-${1000 + i}`,
    item_name: `Product ${i} - ${Math.random().toString(36).substring(7)}`,
    current_stock: Math.floor(Math.random() * 100),
    reserved_stock: Math.floor(Math.random() * 10),
    available_stock: Math.floor(Math.random() * 90),
    warehouse: ["Main", "North"][i % 2],
    last_updated: new Date().toISOString(),
  }));

export default function InventoryPage() {
  const [data, setData] = useState<InventoryItem[]>([]);
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

  useEffect(() => {
    setData(generateMockData(1000));
  }, []);

  const handleStockUpdate = (id: string, val: number) => {
    setData((p) =>
      p.map((i) =>
        i.item_id === id
          ? {
              ...i,
              available_stock: val,
              current_stock: val + i.reserved_stock,
            }
          : i,
      ),
    );
    setEditingId(null);
    message.success("Updated");
  };

  const handleCancelEdit = () => {
    setEditingId(null);
  };

  const filteredData = useMemo(
    () =>
      data.filter((item) => {
        const matchesSearch =
          item.item_sku.toLowerCase().includes(searchText.toLowerCase()) ||
          item.item_name.toLowerCase().includes(searchText.toLowerCase());
        const s = item.available_stock;
        const matchesFilter =
          statusFilter === "all" ||
          (statusFilter === "low" && s <= 10 && s > 0) ||
          (statusFilter === "out" && s === 0) ||
          (statusFilter === "ok" && s > 10);
        return matchesSearch && matchesFilter;
      }),
    [data, searchText, statusFilter],
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
          itemId={r.item_id}
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
      render: (v) => new Date(v).toLocaleString(),
    },
  ];

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
    </div>
  );
}
