import { Table, Tag, Typography, theme } from "antd";
import type { ColumnsType } from "antd/es/table";
import { formatCurrency, formatDate } from "@/lib/analyticsHelpers";
import type {
  ShopeeShippingOrder,
  TiktokShippingOrder,
  ReportPlatform,
} from "@/types/analytics";

interface ShippingFeeTableProps {
  data: ShopeeShippingOrder[] | TiktokShippingOrder[] | null | undefined;
  loading: boolean;
  platform: ReportPlatform;
  onOrderClick?: (row: ShopeeShippingOrder | TiktokShippingOrder) => void;
}

function selectorId(value: string) {
  return value.replace(/[^a-zA-Z0-9_-]/g, "-");
}

/**
 * Ant Design Table for shipping fee order details.
 * Columns adapt based on platform (platform_fee vs shipping_fee).
 * Differences are color-coded (green positive, red negative).
 */
export function ShippingFeeTable({
  data,
  loading,
  platform,
  onOrderClick,
}: ShippingFeeTableProps) {
  const { token } = theme.useToken();

  const columns: ColumnsType<ShopeeShippingOrder | TiktokShippingOrder> = [
    {
      title: "Order SN",
      dataIndex: "order_sn",
      key: "order_sn",
      width: 180,
      ellipsis: true,
      render: (value: string, row) => (
        <Typography.Link
          data-testid={`shipping-order-${selectorId(value)}`}
          onClick={(event) => {
            event.stopPropagation();
            onOrderClick?.(row);
          }}
        >
          {value}
        </Typography.Link>
      ),
    },
    {
      title: "Status",
      dataIndex: "status",
      key: "status",
      width: 120,
      render: (status: string) => {
        const colorMap: Record<string, string> = {
          completed: "green",
          delivered: "blue",
          pending: "orange",
          cancelled: "red",
          returned: "purple",
        };
        return (
          <Tag color={colorMap[status.toLowerCase()] || "default"}>
            {status}
          </Tag>
        );
      },
    },
    {
      title: "Order Date",
      dataIndex: "order_date",
      key: "order_date",
      width: 140,
      render: (val: string) => formatDate(val),
    },
    ...(platform === "shopee"
      ? [
          {
            title: "Buyer Paid",
            dataIndex: "buyer_paid",
            key: "buyer_paid",
            width: 130,
            align: "right" as const,
            render: (val: number) => formatCurrency(val),
          },
          {
            title: "Shopee Rebate",
            dataIndex: "shopee_rebate",
            key: "shopee_rebate",
            width: 130,
            align: "right" as const,
            render: (val: number) => formatCurrency(val),
          },
          {
            title: "Actual Fee",
            dataIndex: "actual_fee",
            key: "actual_fee",
            width: 130,
            align: "right" as const,
            render: (val: number) => formatCurrency(val),
          },
        ]
      : [
          {
            title: "Customer Paid",
            dataIndex: "customer_paid",
            key: "customer_paid",
            width: 130,
            align: "right" as const,
            render: (val: number) => formatCurrency(val),
          },
          {
            title: "Platform Discount",
            dataIndex: "platform_discount",
            key: "platform_discount",
            width: 150,
            align: "right" as const,
            render: (val: number) => formatCurrency(val),
          },
          {
            title: "Actual Fee",
            dataIndex: "actual_fee",
            key: "actual_fee",
            width: 130,
            align: "right" as const,
            render: (val: number) => formatCurrency(val),
          },
        ]),
    {
      title: "Difference",
      dataIndex: "difference",
      key: "difference",
      width: 130,
      align: "right",
      render: (val: number) => (
        <span
          style={{
            color:
              val > 0
                ? token.colorSuccess
                : val < 0
                  ? token.colorError
                  : undefined,
            fontWeight: val !== 0 ? 600 : undefined,
          }}
        >
          {val > 0 ? "+" : ""}
          {formatCurrency(val)}
        </span>
      ),
    },
  ];

  return (
    <Table<ShopeeShippingOrder | TiktokShippingOrder>
      columns={columns}
      dataSource={data ?? []}
      rowKey="order_sn"
      loading={loading}
      size="small"
      pagination={{
        pageSize: 20,
        showSizeChanger: true,
        pageSizeOptions: ["10", "20", "50", "100"],
        showTotal: (total) => `Total ${total} orders`,
      }}
      scroll={{ x: 1000 }}
      onRow={(row) => ({
        onClick: () => onOrderClick?.(row),
        style: onOrderClick ? { cursor: "pointer" } : undefined,
      })}
    />
  );
}
