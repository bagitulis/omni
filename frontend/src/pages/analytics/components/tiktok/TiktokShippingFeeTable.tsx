import { Table, Typography, Tag, theme } from "antd";
import type { ColumnsType } from "antd/es/table";
import type { TiktokShippingFeeOrder } from "@/types/analytics";
import { formatCurrency, formatAnalyticsDate } from "@/lib/analyticsHelpers";

const { Text } = Typography;

interface Props {
  orders: TiktokShippingFeeOrder[];
  loading?: boolean;
}

export function TiktokShippingFeeTable({ orders, loading }: Props) {
  const { token } = theme.useToken();

  const columns: ColumnsType<TiktokShippingFeeOrder> = [
    {
      title: "Order Date",
      dataIndex: "order_date",
      key: "order_date",
      width: 120,
      fixed: "left",
      render: (date: string | null) => (
        <Text type="secondary">{formatAnalyticsDate(date)}</Text>
      ),
      sorter: (a, b) => {
        if (!a.order_date) return -1;
        if (!b.order_date) return 1;
        return (
          new Date(a.order_date).getTime() - new Date(b.order_date).getTime()
        );
      },
    },
    {
      title: "Order ID",
      dataIndex: "order_id",
      key: "order_id",
      width: 180,
      fixed: "left",
      render: (id: string) => (
        <Text
          copyable={{ text: id }}
          style={{ fontFamily: "monospace", fontSize: "0.85rem" }}
        >
          {id}
        </Text>
      ),
    },
    {
      title: "Buyer Paid",
      dataIndex: "buyer_paid",
      key: "buyer_paid",
      width: 130,
      align: "right",
      render: (val: number) => <Text>{formatCurrency(val)}</Text>,
      sorter: (a, b) => a.buyer_paid - b.buyer_paid,
    },
    {
      title: "Actual Fee",
      dataIndex: "actual_fee",
      key: "actual_fee",
      width: 130,
      align: "right",
      render: (val: number) => <Text>{formatCurrency(val)}</Text>,
      sorter: (a, b) => a.actual_fee - b.actual_fee,
    },
    {
      title: "Platform Discount",
      dataIndex: "platform_discount",
      key: "platform_discount",
      width: 150,
      align: "right",
      render: (val: number) => <Text>{formatCurrency(val)}</Text>,
    },
    {
      title: "Difference",
      dataIndex: "difference",
      key: "difference",
      width: 130,
      align: "right",
      render: (val: number) => (
        <Text
          style={{
            color: val >= 0 ? token.colorSuccess : token.colorError,
            fontWeight: 600,
          }}
        >
          {val > 0 ? "+" : ""}
          {formatCurrency(val)}
        </Text>
      ),
      sorter: (a, b) => a.difference - b.difference,
    },
    {
      title: "Status",
      dataIndex: "order_status",
      key: "order_status",
      width: 140,
      render: (status: string) => <Tag>{status}</Tag>,
      filters: Array.from(new Set(orders.map((o) => o.order_status))).map(
        (s) => ({ text: s, value: s || "" }),
      ),
      onFilter: (value, record) => record.order_status === value,
    },
  ];

  return (
    <div
      style={{
        background: token.colorBgContainer,
        borderRadius: token.borderRadiusLG,
        padding: 0,
        overflow: "hidden",
        border: `1px solid ${token.colorBorderSecondary}`,
      }}
    >
      <div
        style={{
          padding: "16px 20px",
          borderBottom: `2px solid ${token.colorBorderSecondary}`,
          background: token.colorFillAlter,
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
        }}
      >
        <Typography.Title
          level={4}
          style={{
            margin: 0,
            fontSize: "1.1rem",
            color: token.colorTextHeading,
          }}
        >
          🚚 TikTok Shipping Fee Analysis
        </Typography.Title>
        <Text type="secondary" style={{ fontSize: "0.9rem" }}>
          {orders.length} orders with difference
        </Text>
      </div>
      <Table
        columns={columns}
        dataSource={orders}
        rowKey="order_id"
        loading={loading}
        pagination={{
          pageSize: 10,
          showSizeChanger: true,
          showTotal: (total) => `Total ${total} items`,
          position: ["bottomRight"],
        }}
        scroll={{ x: 1000 }}
        size="small"
        rowClassName={(record) =>
          record.difference < 0
            ? "table-row-loss"
            : record.difference > 0
              ? "table-row-profit"
              : ""
        }
        style={{ fontSize: "0.85rem" }}
      />
      <style>
        {`
          .table-row-profit { background-color: rgba(34, 197, 94, 0.05); }
          .table-row-loss { background-color: rgba(239, 68, 68, 0.05); }
        `}
      </style>
    </div>
  );
}

export default TiktokShippingFeeTable;
