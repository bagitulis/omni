import { useMemo } from "react";
import { Table, Tag } from "antd";
import type { ColumnsType } from "antd/es/table";
import type {
  DbProductRow,
  ProductManagerPlatform,
} from "@/types/product_manager";
import { getProductImage, getRowKey } from "../utils";

interface ProductManagerTableProps {
  activeTab: ProductManagerPlatform;
  isLoading: boolean;
  products: DbProductRow[];
  total: number;
  page: number;
  pageSize: number;
  onPageChange: (page: number, pageSize: number) => void;
}

function buildColumns(
  platform: ProductManagerPlatform,
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
        const sku =
          typeof row.sku === "string"
            ? row.sku
            : typeof row.sku_name === "string"
              ? row.sku_name
              : typeof row.seller_sku === "string"
                ? row.seller_sku
                : typeof row.shop_sku === "string"
                  ? row.shop_sku
                  : "";

        return (
          <div>
            <div style={{ fontWeight: 600, color: "#0f172a" }}>
              {name || "(Unnamed)"}
            </div>
            <div style={{ fontSize: 12, color: "#64748b" }}>{sku}</div>
          </div>
        );
      },
    },
    {
      title: "Price",
      key: "price",
      width: 120,
      align: "right",
      render: (_, row) => {
        const price =
          typeof row.price === "number"
            ? row.price
            : typeof row.special_price === "number"
              ? row.special_price
              : null;
        if (price == null) return "—";
        return `Rp ${Math.round(price).toLocaleString("id-ID")}`;
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
}: ProductManagerTableProps) {
  const columns = useMemo(() => buildColumns(activeTab), [activeTab]);

  return (
    <Table<DbProductRow>
      data-testid="product-manager-table"
      style={{ marginTop: 12 }}
      loading={isLoading}
      dataSource={products}
      columns={columns}
      rowKey={(row) => getRowKey(activeTab, row)}
      scroll={{ x: 1200 }}
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
