import { Image, Button, Space, Typography, Tag, Tooltip } from "antd";
import {
  EditOutlined,
  DeleteOutlined,
  WarningOutlined,
  CloseCircleOutlined,
} from "@ant-design/icons";
import { Product } from "@/types/product";
import { PlatformBadge } from "@/components/ui/PlatformBadge";
import type { TableProps } from "antd";
import { VirtualTable } from "@/components/common/VirtualTable";

interface ProductTableProps {
  loading: boolean;
  products: Product[];
  total: number;
  page: number;
  pageSize: number;
  selectedRowKeys: React.Key[];
  onPageChange: (page: number, pageSize: number) => void;
  onSelectionChange: (selectedRowKeys: React.Key[]) => void;
  onDelete: (id: string) => void;
}

export function ProductTable({
  loading,
  products,
  total,
  page,
  pageSize,
  selectedRowKeys,
  onPageChange,
  onSelectionChange,
  onDelete,
}: ProductTableProps) {
  const columns: TableProps<Product>["columns"] = [
    {
      title: "Image",
      dataIndex: "image_url",
      key: "image",
      width: 60,
      render: (url) => (
        <Image
          src={url}
          alt="Product"
          width={40}
          height={40}
          style={{ objectFit: "cover", borderRadius: 3 }}
          fallback="https://via.placeholder.com/40"
        />
      ),
    },
    {
      title: "Product Name",
      dataIndex: "item_name",
      key: "name",
      render: (text) => <Typography.Text strong>{text}</Typography.Text>,
    },
    {
      title: "SKU",
      dataIndex: "item_sku",
      key: "sku",
    },
    {
      title: "Stock",
      dataIndex: "stock",
      key: "stock",
      render: (stock) => {
        if (stock === null || stock === undefined) {
          return <Typography.Text type="secondary">—</Typography.Text>;
        }
        if (stock === 0) {
          return (
            <Tooltip title="Out of Stock">
              <Typography.Text type="danger">
                <CloseCircleOutlined /> {stock}
              </Typography.Text>
            </Tooltip>
          );
        }
        if (stock <= 10) {
          return (
            <Tooltip title="Low Stock">
              <Typography.Text type="warning">
                <WarningOutlined /> {stock}
              </Typography.Text>
            </Tooltip>
          );
        }
        return stock;
      },
    },
    {
      title: "Price",
      dataIndex: "price",
      key: "price",
      render: (price) => {
        if (price === null || price === undefined) {
          return <Typography.Text type="secondary">—</Typography.Text>;
        }
        return new Intl.NumberFormat("id-ID", {
          style: "currency",
          currency: "IDR",
          maximumFractionDigits: 0,
        }).format(price);
      },
    },
    {
      title: "Platform",
      dataIndex: "platform",
      key: "platform",
      render: (platform) => <PlatformBadge platform={platform} />,
    },
    {
      title: "Status",
      dataIndex: "status",
      key: "status",
      render: (status) => {
        const color =
          status === "active"
            ? "success"
            : status === "inactive"
              ? "error"
              : "default";
        return <Tag color={color}>{status.toUpperCase()}</Tag>;
      },
    },
    {
      title: "Actions",
      key: "actions",
      width: 100,
      render: (_, record) => (
        <Space>
          <Button size="small" icon={<EditOutlined />} />
          <Button
            size="small"
            danger
            icon={<DeleteOutlined />}
            onClick={() => onDelete(record.item_id)}
          />
        </Space>
      ),
    },
  ];

  return (
    <VirtualTable
      rowKey="item_id"
      loading={loading}
      dataSource={products}
      columns={columns}
      pagination={{
        current: page,
        pageSize: pageSize,
        total: total,
        onChange: onPageChange,
        showSizeChanger: true,
      }}
      rowSelection={{
        selectedRowKeys,
        onChange: onSelectionChange,
      }}
      size="middle"
      scroll={{ x: 800 }}
      enableVirtual={products.length > 20}
      offsetBottom={280}
    />
  );
}
