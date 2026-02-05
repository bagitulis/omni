import { Table, Button, Space, Tag, Modal, Typography } from "antd";
import { LinkOutlined, DeleteOutlined } from "@ant-design/icons";
import { PlatformBadge } from "@/components/ui/PlatformBadge";
import type { TableProps } from "antd";
import { useState } from "react";

/**
 * SKU Mapping represents the relationship between a master SKU and platform SKUs
 */
export interface SkuMapping {
  mapping_id: string;
  master_sku: string;
  platform: string;
  platform_sku: string;
  status: "active" | "inactive" | "pending";
  created_at?: string;
}

interface SkuMappingTableProps {
  mappings: SkuMapping[];
  loading?: boolean;
  onLink?: (masterSku: string) => void;
  onUnlink?: (mappingId: string) => void;
}

export function SkuMappingTable({
  mappings,
  loading = false,
  onLink,
  onUnlink,
}: SkuMappingTableProps) {
  const [unlinkingId, setUnlinkingId] = useState<string | null>(null);

  const handleUnlink = (mappingId: string) => {
    Modal.confirm({
      title: "Unlink SKU Mapping",
      content: "Are you sure you want to unlink this SKU mapping?",
      okText: "Unlink",
      okType: "danger",
      onOk() {
        setUnlinkingId(null);
        onUnlink?.(mappingId);
      },
      onCancel() {
        setUnlinkingId(null);
      },
    });
  };

  const handleLink = (masterSku: string) => {
    // Placeholder - opens modal for linking
    Modal.info({
      title: "Link SKU Mapping",
      content: `Link new platform SKU for master SKU: ${masterSku}`,
      okText: "Close",
    });
    onLink?.(masterSku);
  };

  const columns: TableProps<SkuMapping>["columns"] = [
    {
      title: "Master SKU",
      dataIndex: "master_sku",
      key: "master_sku",
      width: 120,
      render: (text) => <Typography.Text strong>{text}</Typography.Text>,
    },
    {
      title: "Platform",
      dataIndex: "platform",
      key: "platform",
      width: 100,
      render: (platform) => <PlatformBadge platform={platform} />,
    },
    {
      title: "Platform SKU",
      dataIndex: "platform_sku",
      key: "platform_sku",
      render: (text) => <Typography.Text code>{text}</Typography.Text>,
    },
    {
      title: "Status",
      dataIndex: "status",
      key: "status",
      width: 100,
      render: (status: string) => {
        const colorMap: Record<string, string> = {
          active: "success",
          inactive: "default",
          pending: "warning",
        };
        return <Tag color={colorMap[status]}>{status.toUpperCase()}</Tag>;
      },
    },
    {
      title: "Actions",
      key: "actions",
      width: 140,
      render: (_, record) => (
        <Space size="small">
          <Button
            type="primary"
            size="small"
            icon={<LinkOutlined />}
            onClick={() => handleLink(record.master_sku)}
          >
            Link
          </Button>
          <Button
            danger
            size="small"
            icon={<DeleteOutlined />}
            loading={unlinkingId === record.mapping_id}
            onClick={() => handleUnlink(record.mapping_id)}
          >
            Unlink
          </Button>
        </Space>
      ),
    },
  ];

  return (
    <Table
      rowKey="mapping_id"
      loading={loading}
      dataSource={mappings}
      columns={columns}
      pagination={{ pageSize: 10, showSizeChanger: true }}
      size="middle"
      scroll={{ x: 800 }}
    />
  );
}
