import React from "react";
import { Table, Tag, Typography, Tooltip, theme } from "antd";
import type { ColumnsType } from "antd/es/table";
import {
  CheckCircleOutlined,
  WarningOutlined,
  QuestionCircleOutlined,
} from "@ant-design/icons";
import type { TiktokSkuGroup } from "@/types/analytics";
import { formatCurrency } from "@/lib/analyticsHelpers";

const { Text } = Typography;

interface Props {
  skuGroups: TiktokSkuGroup[];
  loading?: boolean;
}

function TiktokPriceResultsTable({ skuGroups, loading }: Props) {
  const { token } = theme.useToken();

  const getStatusColor = (status: string) => {
    switch (status) {
      case "OK":
        return "success";
      case "PRICE_DIFF":
        return "warning";
      case "NO_INVENTORY":
        return "default";
      default:
        return "default";
    }
  };

  const getStatusLabel = (status: string) => {
    switch (status) {
      case "OK":
        return (
          <span>
            <CheckCircleOutlined /> OK
          </span>
        );
      case "PRICE_DIFF":
        return (
          <span>
            <WarningOutlined /> Diff
          </span>
        );
      case "NO_INVENTORY":
        return (
          <span>
            <QuestionCircleOutlined /> N/A
          </span>
        );
      default:
        return status;
    }
  };

  const getTopTwoSmallestIncomes = (incomes: number[]): number[] => {
    if (!incomes || incomes.length === 0) return [];
    const sorted = [...incomes].sort((a, b) => a - b);
    return sorted.slice(0, 2);
  };

  const columns: ColumnsType<TiktokSkuGroup> = [
    {
      title: "Status",
      dataIndex: "status",
      key: "status",
      width: 100,
      fixed: "left",
      render: (status: string) => (
        <Tag
          color={getStatusColor(status)}
          style={{ minWidth: 60, textAlign: "center" }}
        >
          {getStatusLabel(status)}
        </Tag>
      ),
      filters: [
        { text: "OK", value: "OK" },
        { text: "Diff", value: "PRICE_DIFF" },
        { text: "N/A", value: "NO_INVENTORY" },
      ],
      onFilter: (value, record) => record.status === value,
    },
    {
      title: "SKU",
      dataIndex: "sku", // Using sku as primary identifier, though Vue uses model_sku or sku
      key: "sku",
      width: 150,
      fixed: "left",
      render: (sku: string, record) => (
        <Tooltip
          title={
            record.seller_sku !== sku
              ? `Seller SKU: ${record.seller_sku}`
              : null
          }
        >
          <Text
            style={{
              fontFamily: "monospace",
              color: token.colorPrimary,
              fontWeight: 600,
            }}
          >
            {sku}
          </Text>
        </Tooltip>
      ),
      sorter: (a, b) => a.sku.localeCompare(b.sku),
    },
    {
      title: "Item Name",
      dataIndex: "product_name",
      key: "item_name",
      width: 300,
      render: (name: string) => (
        <Text
          style={{
            whiteSpace: "normal",
            wordWrap: "break-word",
            lineHeight: 1.4,
          }}
        >
          {name}
        </Text>
      ),
      sorter: (a, b) => a.product_name.localeCompare(b.product_name),
    },
    {
      title: "Marketplace Price",
      dataIndex: "unique_unit_prices",
      key: "marketplace_price",
      width: 150,
      render: (prices: number[]) => (
        <Text code>{prices.map((p) => formatCurrency(p)).join(", ")}</Text>
      ),
    },
    {
      title: "Inventory Price",
      dataIndex: "inventory_price",
      key: "inventory_price",
      width: 140,
      align: "right",
      render: (price: number | null) => (
        <Text code>{price !== null ? formatCurrency(price) : "-"}</Text>
      ),
      sorter: (a, b) => (a.inventory_price || 0) - (b.inventory_price || 0),
    },
    {
      title: "Actual Income",
      dataIndex: "unique_actual_incomes",
      key: "actual_income",
      width: 160,
      align: "right",
      render: (incomes: number[]) => {
        const topTwo = getTopTwoSmallestIncomes(incomes);
        return (
          <Text code>
            {topTwo.map((p) => formatCurrency(p)).join(", ")}
            {incomes.length > 2 && "..."}
          </Text>
        );
      },
    },
    {
      title: "Expected Income",
      dataIndex: "expected_income",
      key: "expected_income",
      width: 140,
      align: "right",
      render: (price: number | null) => (
        <Text
          style={{
            color: token.colorSuccess,
            fontWeight: 600,
            fontFamily: "monospace",
          }}
        >
          {price !== null ? formatCurrency(price) : "-"}
        </Text>
      ),
      sorter: (a, b) => (a.expected_income || 0) - (b.expected_income || 0),
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
          📋 TikTok SKU Price Analysis
        </Typography.Title>
        <Text type="secondary" style={{ fontSize: "0.9rem" }}>
          {skuGroups.length} items
        </Text>
      </div>
      <Table
        columns={columns}
        dataSource={skuGroups}
        rowKey="sku"
        loading={loading}
        pagination={{
          defaultPageSize: 10,
          showSizeChanger: true,
          showTotal: (total) => `Total ${total} items`,
          position: ["bottomRight"],
        }}
        scroll={{ x: 1200 }}
        size="small"
        rowClassName={(record) => {
          switch (record.status) {
            case "OK":
              return "table-row-ok";
            case "PRICE_DIFF":
              return "table-row-warning";
            case "NO_INVENTORY":
              return "table-row-info";
            default:
              return "";
          }
        }}
        style={{ fontSize: "0.9rem" }}
      />
      <style>
        {`
          .table-row-ok { background-color: ${token.colorSuccessBg}; }
          .table-row-warning { background-color: ${token.colorWarningBg}; }
          .table-row-info { background-color: ${token.colorFillSecondary}; }
        `}
      </style>
    </div>
  );
}

export default React.memo(TiktokPriceResultsTable);
export { TiktokPriceResultsTable };
