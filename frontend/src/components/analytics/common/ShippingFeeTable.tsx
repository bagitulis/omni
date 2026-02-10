import { Table, Typography } from "antd";
import type { ColumnsType } from "antd/es/table";
import type {
  ShopeeShippingFeeOrder,
  TiktokShippingFeeOrder,
  Platform,
} from "@/types/analytics";
import { formatCurrency, formatAnalyticsDate } from "@/lib/analyticsHelpers";

const { Text } = Typography;

interface Props {
  platform: Platform;
  data: ShopeeShippingFeeOrder[] | TiktokShippingFeeOrder[];
  loading?: boolean;
}

export function ShippingFeeTable({ platform, data, loading }: Props) {
  const getColumns = (): ColumnsType<
    ShopeeShippingFeeOrder | TiktokShippingFeeOrder
  > => {
    if (platform === "shopee") {
      return [
        {
          title: "Order Date",
          dataIndex: "order_date",
          key: "order_date",
          width: 120,
          render: (date: string | null) => formatAnalyticsDate(date),
        },
        {
          title: "Order SN",
          dataIndex: "order_sn",
          key: "order_sn",
          width: 180,
          render: (sn: string) => <Text copyable>{sn}</Text>,
        },
        {
          title: "Buyer Paid",
          dataIndex: "buyer_paid",
          key: "buyer_paid",
          align: "right",
          width: 130,
          render: (val: number) => formatCurrency(val),
        },
        {
          title: "Actual Fee",
          dataIndex: "actual_fee",
          key: "actual_fee",
          align: "right",
          width: 130,
          render: (val: number) => formatCurrency(val),
        },
        {
          title: "Shopee Rebate",
          dataIndex: "shopee_rebate",
          key: "shopee_rebate",
          align: "right",
          width: 130,
          render: (val: number) => formatCurrency(val),
        },
        {
          title: "Difference",
          dataIndex: "difference",
          key: "difference",
          align: "right",
          width: 130,
          render: (val: number) => (
            <Text type={val >= 0 ? "success" : "danger"}>
              {formatCurrency(val)}
            </Text>
          ),
        },
        {
          title: "Buyer Name",
          dataIndex: "buyer_name",
          key: "buyer_name",
          ellipsis: true,
        },
        {
          title: "Payment Method",
          dataIndex: "payment_method",
          key: "payment_method",
          width: 150,
        },
      ];
    }

    // TikTok Columns
    return [
      {
        title: "Order Date",
        dataIndex: "order_date",
        key: "order_date",
        width: 120,
        render: (date: string | null) => formatAnalyticsDate(date),
      },
      {
        title: "Order ID",
        dataIndex: "order_id",
        key: "order_id",
        width: 180,
        render: (id: string) => <Text copyable>{id}</Text>,
      },
      {
        title: "Buyer Paid",
        dataIndex: "buyer_paid",
        key: "buyer_paid",
        align: "right",
        width: 130,
        render: (val: number) => formatCurrency(val),
      },
      {
        title: "Actual Fee",
        dataIndex: "actual_fee",
        key: "actual_fee",
        align: "right",
        width: 130,
        render: (val: number) => formatCurrency(val),
      },
      {
        title: "Platform Discount",
        dataIndex: "platform_discount",
        key: "platform_discount",
        align: "right",
        width: 130,
        render: (val: number) => formatCurrency(val),
      },
      {
        title: "Difference",
        dataIndex: "difference",
        key: "difference",
        align: "right",
        width: 130,
        render: (val: number) => (
          <Text type={val >= 0 ? "success" : "danger"}>
            {formatCurrency(val)}
          </Text>
        ),
      },
      {
        title: "Order Status",
        dataIndex: "order_status",
        key: "order_status",
        width: 120,
      },
      {
        title: "Currency",
        dataIndex: "currency",
        key: "currency",
        width: 80,
      },
    ];
  };

  return (
    <Table
      columns={getColumns()}
      dataSource={data}
      loading={loading}
      rowKey={platform === "shopee" ? "order_sn" : "order_id"}
      bordered
      size="small"
      pagination={{
        pageSize: 10,
        showSizeChanger: true,
        showTotal: (total) => `Total ${total} items`,
      }}
      style={{ fontSize: 12 }}
    />
  );
}
