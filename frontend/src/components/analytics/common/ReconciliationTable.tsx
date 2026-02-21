import { Table, Tag, Typography } from "antd";
import type { ColumnsType } from "antd/es/table";
import type { SkuGroup, TiktokSkuGroup, Platform } from "@/types/analytics";
import {
  formatCurrency,
  getStatusColor,
  getStatusLabel,
} from "@/lib/analyticsHelpers";

const { Text } = Typography;

interface Props {
  platform: Platform;
  data: SkuGroup[] | TiktokSkuGroup[];
  loading?: boolean;
}

export function ReconciliationTable({ platform, data, loading }: Props) {
  // Define columns based on platform
  const getColumns = (): ColumnsType<SkuGroup | TiktokSkuGroup> => {
    const commonColumns: ColumnsType<SkuGroup | TiktokSkuGroup> = [
      {
        title: "Status",
        dataIndex: "status",
        key: "status",
        width: 120,
        render: (status: string) => (
          <Tag color={getStatusColor(status)}>{getStatusLabel(status)}</Tag>
        ),
      },
      {
        title: "SKU",
        dataIndex: "sku",
        key: "sku",
        width: 150,
        render: (sku: string) => <Text strong>{sku}</Text>,
      },
    ];

    if (platform === "shopee") {
      return [
        ...commonColumns,
        {
          title: "Model SKU",
          dataIndex: "model_sku",
          key: "model_sku",
          width: 150,
        },
        {
          title: "Item Name",
          dataIndex: "item_name",
          key: "item_name",
          ellipsis: true,
        },
        {
          title: "Model Name",
          dataIndex: "model_name",
          key: "model_name",
          ellipsis: true,
        },
        {
          title: "Inventory Price",
          dataIndex: "inventory_price",
          key: "inventory_price",
          width: 130,
          align: "right",
          render: (val: number | null) =>
            val !== null ? formatCurrency(val) : "—",
        },
        {
          title: "Expected Income",
          dataIndex: "expected_income",
          key: "expected_income",
          width: 130,
          align: "right",
          render: (val: number | null) =>
            val !== null ? formatCurrency(val) : "—",
        },
        {
          title: "Transactions",
          dataIndex: "total_transactions",
          key: "total_transactions",
          width: 100,
          align: "center",
        },
        {
          title: "Unit Prices",
          dataIndex: "unique_unit_prices",
          key: "unique_unit_prices",
          width: 150,
          render: (prices: number[]) =>
            prices.map((p) => formatCurrency(p)).join(", "),
        },
      ];
    }

    // TikTok Columns
    return [
      ...commonColumns,
      {
        title: "Seller SKU",
        dataIndex: "seller_sku",
        key: "seller_sku",
        width: 150,
      },
      {
        title: "Product Name",
        dataIndex: "product_name",
        key: "product_name",
        ellipsis: true,
      },
      {
        title: "Inventory Price",
        dataIndex: "inventory_price",
        key: "inventory_price",
        width: 130,
        align: "right",
        render: (val: number | null) =>
          val !== null ? formatCurrency(val) : "—",
      },
      {
        title: "Expected Income",
        dataIndex: "expected_income",
        key: "expected_income",
        width: 130,
        align: "right",
        render: (val: number | null) =>
          val !== null ? formatCurrency(val) : "—",
      },
      {
        title: "Transactions",
        dataIndex: "total_transactions",
        key: "total_transactions",
        width: 100,
        align: "center",
      },
      {
        title: "Unit Prices",
        dataIndex: "unique_unit_prices",
        key: "unique_unit_prices",
        width: 150,
        render: (prices: number[]) =>
          prices.map((p) => formatCurrency(p)).join(", "),
      },
    ];
  };

  return (
    <Table
      columns={getColumns()}
      dataSource={data}
      loading={loading}
      rowKey="sku"
      bordered
      size="small"
      pagination={{
        defaultPageSize: 10,
        showSizeChanger: true,
        showTotal: (total) => `Total ${total} items`,
      }}
      style={{ fontSize: 12 }}
    />
  );
}
