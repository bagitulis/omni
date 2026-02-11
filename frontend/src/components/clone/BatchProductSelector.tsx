import React from "react";
import { Table, Alert, Avatar, Tag, Typography, Space } from "antd";
import type { Product } from "@/types/product";

const { Text } = Typography;

interface BatchProductSelectorProps {
  products: Product[];
  selectedRowKeys: React.Key[];
  onSelectionChange: (keys: React.Key[]) => void;
  sourcePlatform: string;
}

export const BatchProductSelector: React.FC<BatchProductSelectorProps> = ({
  products,
  selectedRowKeys,
  onSelectionChange,
  sourcePlatform,
}) => {
  const columns = [
    {
      title: "Product",
      dataIndex: "item_name",
      key: "item_name",
      render: (text: string, record: Product) => (
        <Space>
          <Avatar shape="square" src={record.image_url} />
          <Text ellipsis style={{ maxWidth: 300 }}>
            {text}
          </Text>
        </Space>
      ),
    },
    {
      title: "SKU",
      dataIndex: "item_sku",
      key: "item_sku",
    },
    {
      title: "Price",
      dataIndex: "price",
      key: "price",
      render: (val: number) =>
        val
          ? new Intl.NumberFormat("id-ID", {
              style: "currency",
              currency: "IDR",
            }).format(val)
          : "-",
    },
    {
      title: "Platform",
      dataIndex: "platform",
      key: "platform",
      render: (val: string) => <Tag>{val.toUpperCase()}</Tag>,
    },
  ];

  return (
    <div style={{ padding: "20px 0" }}>
      <Alert
        message={`Selected Source: ${sourcePlatform.toUpperCase()}`}
        type="info"
        showIcon
        style={{ marginBottom: 16 }}
      />
      <Table
        rowSelection={{
          selectedRowKeys,
          onChange: onSelectionChange,
        }}
        columns={columns}
        dataSource={products}
        rowKey="item_id"
        pagination={{ pageSize: 5 }}
        size="small"
        scroll={{ y: 300 }}
      />
      <div style={{ marginTop: 8, textAlign: "right" }}>
        <Text type="secondary">
          {selectedRowKeys.length} products selected (Max 100)
        </Text>
      </div>
    </div>
  );
};
