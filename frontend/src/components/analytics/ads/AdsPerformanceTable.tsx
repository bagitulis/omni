import React from "react";
import { Table, Tag, Typography } from "antd";
import type { TableColumnsType } from "antd";
import {
  ShopeeAdsDashboardTopProduct,
  TiktokAdsDashboardTopProduct,
} from "@/types/ads";

const { Text } = Typography;

export type PerformanceData =
  | ShopeeAdsDashboardTopProduct
  | TiktokAdsDashboardTopProduct;

interface AdsPerformanceTableProps {
  data: PerformanceData[];
  loading?: boolean;
}

export const AdsPerformanceTable: React.FC<AdsPerformanceTableProps> = ({
  data,
  loading,
}) => {
  const formatCurrency = (val: number) =>
    new Intl.NumberFormat("id-ID", {
      style: "currency",
      currency: "IDR",
      maximumFractionDigits: 0,
    }).format(val);

  const getRoasColor = (roas: number) => {
    if (roas >= 5) return "success";
    if (roas >= 2) return "warning";
    return "error";
  };

  const columns: TableColumnsType<PerformanceData> = [
    {
      title: "Product",
      dataIndex: "product_name",
      key: "product_name",
      width: 250,
      ellipsis: true,
      render: (text, record) => (
        <div>
          <Text strong>{text}</Text>
          <br />
          <Text type="secondary" style={{ fontSize: 11 }}>
            {record.product_id}
          </Text>
        </div>
      ),
      sorter: (a, b) => a.product_name.localeCompare(b.product_name),
    },
    {
      title: "Spend",
      dataIndex: "cost",
      key: "cost",
      align: "right",
      width: 120,
      render: (val) => formatCurrency(val),
      sorter: (a, b) => a.cost - b.cost,
    },
    {
      title: "GMV",
      dataIndex: "revenue",
      key: "revenue",
      align: "right",
      width: 120,
      render: (val) => formatCurrency(val),
      sorter: (a, b) => a.revenue - b.revenue,
    },
    {
      title: "Orders",
      dataIndex: "orders",
      key: "orders",
      align: "right",
      width: 100,
      sorter: (a, b) => a.orders - b.orders,
    },
    {
      title: "ROAS",
      dataIndex: "roas",
      key: "roas",
      align: "center",
      width: 100,
      render: (val) => <Tag color={getRoasColor(val)}>{val.toFixed(2)}x</Tag>,
      sorter: (a, b) => a.roas - b.roas,
      defaultSortOrder: "descend",
    },
  ];

  return (
    <Table
      columns={columns}
      dataSource={data}
      rowKey="product_id"
      size="small"
      loading={loading}
      pagination={{ pageSize: 5, hideOnSinglePage: true }}
    />
  );
};

export default AdsPerformanceTable;
