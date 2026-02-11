import { useMemo, useState } from "react";
import { Table, Tag, InputNumber, Button, Space, message } from "antd";
import { CheckOutlined, CloseOutlined } from "@ant-design/icons";
import type { ColumnsType } from "antd/es/table";
import type {
  DbProductRow,
  ProductManagerPlatform,
} from "@/types/product_manager";
import type { SkuCheckResult } from "@/types/inventory";
import { getProductImage, getRowKey, extractSku } from "../utils";

interface ProductManagerTableProps {
  activeTab: ProductManagerPlatform;
  isLoading: boolean;
  products: DbProductRow[];
  total: number;
  page: number;
  pageSize: number;
  onPageChange: (page: number, pageSize: number) => void;
  onUpdatePrice: (sku: string, price: number) => Promise<void>;
  selectedRowKeys: React.Key[];
  onSelectionChange: (selectedRowKeys: React.Key[]) => void;
  skuCheckResults: SkuCheckResult[];
  getSkuResult: (sku: string) => SkuCheckResult | null;
}

function PriceCell({
  value,
  row,
  onSave,
}: {
  value: number | null;
  row: DbProductRow;
  onSave: (sku: string, price: number) => Promise<void>;
}) {
  const [editing, setEditing] = useState(false);
  const [price, setPrice] = useState<number | null>(value);

  const handleSave = async () => {
    if (price === null || price <= 0) {
      message.error("Price must be greater than 0");
      return;
    }
    const sku = extractSku(row);
    if (!sku) {
      message.error("Product has no SKU");
      return;
    }
    await onSave(sku, price);
    setEditing(false);
  };

  if (editing) {
    return (
      <Space>
        <InputNumber
          min={1}
          value={price}
          onChange={setPrice}
          style={{ width: 100 }}
          onPressEnter={() => void handleSave()}
        />
        <Button
          size="small"
          type="link"
          icon={<CheckOutlined />}
          onClick={() => void handleSave()}
        />
        <Button
          size="small"
          type="link"
          icon={<CloseOutlined />}
          onClick={() => {
            setEditing(false);
            setPrice(value);
          }}
        />
      </Space>
    );
  }

  return (
    <div
      onClick={() => {
        setPrice(value);
        setEditing(true);
      }}
      style={{
        cursor: "pointer",
        padding: "4px 8px",
        borderRadius: 4,
        border: "1px solid transparent",
        transition: "border-color 0.2s",
      }}
      onMouseEnter={(e) => {
        e.currentTarget.style.borderColor = "#d9d9d9";
      }}
      onMouseLeave={(e) => {
        e.currentTarget.style.borderColor = "transparent";
      }}
    >
      {value != null ? `Rp ${Math.round(value).toLocaleString("id-ID")}` : "—"}
    </div>
  );
}

function buildColumns(
  platform: ProductManagerPlatform,
  onUpdatePrice: (sku: string, price: number) => Promise<void>,
  getSkuResult: (sku: string) => SkuCheckResult | null,
): ColumnsType<DbProductRow> {
  const base: ColumnsType<DbProductRow> = [
    {
      title: "Image",
      key: "image",
      width: 64,
      render: (_, row) => {
        const image = getProductImage(row);
        if (!image) return <span style={{ color: "#94a3b8" }}>—</span>;
        return (
          <img
            src={image}
            alt="product"
            width={40}
            height={40}
            style={{
              objectFit: "cover",
              borderRadius: 4,
              border: "1px solid #e2e8f0",
            }}
          />
        );
      },
    },
    {
      title: "Product",
      key: "product",
      render: (_, row) => {
        const name = typeof row.item_name === "string" ? row.item_name : "";
        const sku = extractSku(row);
        const skuResult = sku ? getSkuResult(sku) : null;

        return (
          <div>
            <div style={{ fontWeight: 600, color: "#0f172a" }}>
              {name || "(Unnamed)"}
            </div>
            <div style={{ fontSize: 12, color: "#64748b" }}>{sku}</div>
            {skuResult && (
              <Space size={4} style={{ marginTop: 4 }}>
                <Tag
                  color={skuResult.shopee ? "success" : "error"}
                  style={{ fontSize: 10, padding: "0 4px", margin: 0 }}
                >
                  S
                </Tag>
                <Tag
                  color={skuResult.lazada ? "success" : "error"}
                  style={{ fontSize: 10, padding: "0 4px", margin: 0 }}
                >
                  L
                </Tag>
                <Tag
                  color={skuResult.tiktok ? "success" : "error"}
                  style={{ fontSize: 10, padding: "0 4px", margin: 0 }}
                >
                  T
                </Tag>
              </Space>
            )}
          </div>
        );
      },
    },
    {
      title: "Price",
      key: "price",
      width: 180,
      align: "right",
      render: (_, row) => {
        const price =
          typeof row.price === "number"
            ? row.price
            : typeof row.special_price === "number"
              ? row.special_price
              : null;
        return <PriceCell value={price} row={row} onSave={onUpdatePrice} />;
      },
    },
    {
      title: "Stock",
      key: "stock",
      width: 100,
      align: "right",
      render: (_, row) => {
        const stock =
          typeof row.stock === "number"
            ? row.stock
            : typeof row.quantity === "number"
              ? row.quantity
              : typeof row.available === "number"
                ? row.available
                : null;
        return stock == null ? "—" : stock.toLocaleString("id-ID");
      },
    },
    {
      title: "Status",
      key: "status",
      width: 120,
      render: (_, row) => {
        const status = typeof row.status === "string" ? row.status : "";
        if (!status) return <Tag>unknown</Tag>;
        const normalized = status.toLowerCase();
        const color =
          normalized === "active" || normalized === "normal"
            ? "green"
            : normalized.includes("delete") || normalized.includes("inactive")
              ? "red"
              : "default";
        return <Tag color={color}>{status}</Tag>;
      },
    },
    {
      title: "Updated",
      dataIndex: "updated_at",
      key: "updated_at",
      width: 180,
      render: (value) => {
        if (typeof value !== "string") return "—";
        return value.replace("T", " ").replace("Z", "");
      },
    },
  ];

  if (platform === "shopee") {
    return [
      { title: "Item ID", dataIndex: "item_id", key: "item_id", width: 120 },
      { title: "Model ID", dataIndex: "model_id", key: "model_id", width: 120 },
      ...base,
    ];
  }

  if (platform === "lazada") {
    return [
      { title: "Item ID", dataIndex: "item_id", key: "item_id", width: 120 },
      { title: "SKU ID", dataIndex: "sku_id", key: "sku_id", width: 140 },
      ...base,
    ];
  }

  return [
    {
      title: "Product ID",
      dataIndex: "product_id",
      key: "product_id",
      width: 140,
    },
    { title: "SKU ID", dataIndex: "sku_id", key: "sku_id", width: 140 },
    ...base,
  ];
}

export function ProductManagerTable({
  activeTab,
  isLoading,
  products,
  total,
  page,
  pageSize,
  onPageChange,
  onUpdatePrice,
  selectedRowKeys,
  onSelectionChange,
  getSkuResult,
}: ProductManagerTableProps) {
  const columns = useMemo(
    () => buildColumns(activeTab, onUpdatePrice, getSkuResult),
    [activeTab, onUpdatePrice, getSkuResult],
  );

  return (
    <Table<DbProductRow>
      data-testid="product-manager-table"
      style={{ marginTop: 12 }}
      loading={isLoading}
      dataSource={products}
      columns={columns}
      rowKey={(row) => getRowKey(activeTab, row)}
      scroll={{ x: 1200 }}
      rowSelection={{
        selectedRowKeys,
        onChange: onSelectionChange,
      }}
      pagination={{
        current: page,
        pageSize,
        total,
        onChange: onPageChange,
        showSizeChanger: true,
      }}
    />
  );
}
