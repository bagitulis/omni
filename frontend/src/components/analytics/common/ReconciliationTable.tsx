import { Table, theme } from "antd";
import type { ColumnsType } from "antd/es/table";
import { formatCurrency } from "@/lib/analyticsHelpers";
import type { SkuGroup, TiktokSkuGroup } from "@/types/analytics";

type ReconciliationRow = SkuGroup | TiktokSkuGroup;

interface ReconciliationTableProps {
  data: ReconciliationRow[] | null | undefined;
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

  const columns: ColumnsType<ReconciliationRow> = [
    {
      title: "SKU",
      dataIndex: "sku",
      key: "sku",
      width: 160,
      ellipsis: true,
    },
    {
      title: "Product",
      key: "name",
      ellipsis: true,
      render: (_, row) => ("item_name" in row ? row.item_name : row.product_name),
    },
    {
      title: "Expected Income",
      dataIndex: "expected_income",
      key: "expected_income",
      width: 140,
      align: "right",
      render: (val: number | null) => (val == null ? "-" : formatCurrency(val)),
    },
    {
      title: "Inventory Price",
      dataIndex: "inventory_price",
      key: "inventory_price",
      width: 140,
      align: "right",
      render: (val: number | null) => (val == null ? "-" : formatCurrency(val)),
    },
    {
      title: "Unit Prices",
      dataIndex: "unique_unit_prices",
      key: "unique_unit_prices",
      width: 140,
      align: "right",
      render: (values: number[]) => values.map(formatCurrency).join(" / "),
    },
    {
      title: "Actual Incomes",
      dataIndex: "unique_actual_incomes",
      key: "unique_actual_incomes",
      width: 160,
      align: "right",
      render: (values: number[]) => values.map(formatCurrency).join(" / "),
    },
    {
      title: "Transactions",
      dataIndex: "total_transactions",
      key: "total_transactions",
      width: 110,
      align: "right",
    },
    {
      title: "Status",
      dataIndex: "status",
      key: "status",
      width: 120,
      render: (status: string, row) => (
        <span
          style={{
            color: row.has_price_difference ? token.colorError : token.colorSuccess,
            fontWeight: 600,
          }}
        >
          {status}
        </span>
      ),
    },
  ];

  return (
    <Table<ReconciliationRow>
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
