import { Image, Button, Space, Typography, Tag, Tooltip, Badge } from "antd";
import { Link } from "react-router-dom";
import {
  EditOutlined,
  DeleteOutlined,
  WarningOutlined,
  CloseCircleOutlined,
  CopyOutlined,
  SyncOutlined,
} from "@ant-design/icons";
import { Product } from "@/types/product";
// PlatformBadge removed as it's not used in MasterProduct table (which shows aggregated status)
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
  onClone?: (product: Product) => void;
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
  onClone,
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
      render: (text, record) => (
        <Space direction="vertical" size={0}>
          <Typography.Text strong>{text}</Typography.Text>
          <Typography.Text type="secondary" style={{ fontSize: 11 }}>
            ID: {record.item_id}
          </Typography.Text>
        </Space>
      ),
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
      title: "Sync Status",
      key: "sync_status",
      render: (_, record) => {
        // Aggregate sync status from all SKUs and their platform links
        const links = record.skus?.flatMap((s) => s.platform_links || []) || [];
        if (links.length === 0) {
          return <Tag color="default">Not Synced</Tag>;
        }

        const platforms = Array.from(new Set(links.map((l) => l.platform)));
        const allSynced = links.every((l) => l.sync_status === "synced");
        const hasError = links.some((l) => l.sync_status === "failed");

        if (hasError)
          return (
            <Tag color="error" icon={<WarningOutlined />}>
              Sync Error
            </Tag>
          );
        if (allSynced)
          return <Tag color="success">Synced ({platforms.length})</Tag>;

        return (
          <Tag color="processing" icon={<SyncOutlined spin />}>
            Syncing...
          </Tag>
        );
      },
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
        return <Badge status={color} text={status.toUpperCase()} />;
      },
    },
    {
      title: "Actions",
      key: "actions",
      width: 140,
      render: (_, record) => (
        <Space>
          <Button
            size="small"
            icon={<CopyOutlined />}
            onClick={() => onClone?.(record)}
            title="Clone"
          />
          <Link to={`/master-products/edit/${record.item_id}`}>
            <Button size="small" icon={<EditOutlined />} title="Edit" />
          </Link>
          <Button
            size="small"
            danger
            icon={<DeleteOutlined />}
            onClick={() => onDelete(record.item_id)}
            title="Delete"
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
      scroll={{ x: 1000 }}
      enableVirtual={products.length > 20}
      offsetBottom={280}
    />
  );
}
