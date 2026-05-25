import { Table, theme } from "antd";
import type { ColumnsType } from "antd/es/table";
import { formatCurrency } from "@/lib/analyticsHelpers";
import type { SkuGroup } from "@/types/analytics";

interface ReconciliationTableProps {
  data: SkuGroup[] | null | undefined;
  loading: boolean;
}

/**
 * Ant Design Table for reconciliation SKU details.
 * Columns show SKU, item name, quantities, amounts, differences, and order count.
 * Amounts are color-coded (green for positive diff, red for negative).
 */
export function ReconciliationTable({
  data,
  loading,
}: ReconciliationTableProps) {
  const { token } = theme.useToken();

  const columns: ColumnsType<SkuGroup> = [
    {
      title: "SKU",
      dataIndex: "sku",
      key: "sku",
      width: 160,
      ellipsis: true,
    },
    {
      title: "Item Name",
      dataIndex: "item_name",
      key: "item_name",
      ellipsis: true,
    },
    {
      title: "Qty",
      dataIndex: "total_quantity",
      key: "total_quantity",
      width: 80,
      align: "right",
    },
    {
      title: "Total Amount",
      dataIndex: "total_amount",
      key: "total_amount",
      width: 140,
      align: "right",
      render: (val: number) => formatCurrency(val),
    },
    {
      title: "System Amount",
      dataIndex: "system_amount",
      key: "system_amount",
      width: 140,
      align: "right",
      render: (val: number) => formatCurrency(val),
    },
    {
      title: "Price Diff",
      dataIndex: "price_diff",
      key: "price_diff",
      width: 140,
      align: "right",
      render: (val: number) => (
        <span
          style={{
            color: val > 0 ? token.colorSuccess : val < 0 ? token.colorError : undefined,
            fontWeight: val !== 0 ? 600 : undefined,
          }}
        >
          {val > 0 ? "+" : ""}
          {formatCurrency(val)}
        </span>
      ),
    },
    {
      title: "Diff %",
      dataIndex: "price_diff_percent",
      key: "price_diff_percent",
      width: 100,
      align: "right",
      render: (val: number) => {
        const color =
          val > 0
            ? token.colorSuccess
            : val < 0
              ? token.colorError
              : undefined;
        return (
          <span style={{ color }}>
            {val.toFixed(2)}%
          </span>
        );
      },
    },
    {
      title: "Orders",
      dataIndex: "order_count",
      key: "order_count",
      width: 80,
      align: "right",
    },
  ];

  return (
    <Table<SkuGroup>
      columns={columns}
      dataSource={data ?? []}
      rowKey="sku"
      loading={loading}
      size="small"
      pagination={{
        pageSize: 20,
        showSizeChanger: true,
        pageSizeOptions: ["10", "20", "50", "100"],
        showTotal: (total) => `Total ${total} SKUs`,
      }}
      scroll={{ x: 900 }}
    />
  );
}
