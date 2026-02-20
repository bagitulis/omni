import {
  ApiOutlined,
  DisconnectOutlined,
} from "@ant-design/icons";
import {
  Badge,
  Button,
  Card,
  message,
  Popconfirm,
  Table,
  Tag,
  theme,
} from "antd";
import type { ColumnsType } from "antd/es/table";
import { useState } from "react";
import { autoMapSkus, unlinkSkuFromPlatform } from "@/api/products";
import type { MasterProduct, MasterProductSku } from "@/types/product";
import type { Platform } from "@/types/shared";
import {
  SkuLinkModal,
  PlatformLinkButton,
  type LinkModalState,
} from "./SkuLinkModal";

interface SkuMappingPanelProps {
  masterProduct: MasterProduct;
  onUpdate?: () => void;
}

const SYNC_STATUS_MAP: Record<
  string,
  { color: "success" | "warning" | "processing" | "error" | "default"; label: string }
> = {
  synced: { color: "success", label: "Synced" },
  outdated: { color: "warning", label: "Outdated" },
  pending: { color: "processing", label: "Pending" },
  not_synced: { color: "processing", label: "Pending" },
  error: { color: "error", label: "Error" },
  failed: { color: "error", label: "Error" },
};

const DEFAULT_STATUS = { color: "default" as const, label: "Unknown" };

export function SkuMappingPanel({
  masterProduct,
  onUpdate,
}: SkuMappingPanelProps) {
  const { token } = theme.useToken();
  const [loading, setLoading] = useState(false);
  const [linkModal, setLinkModal] = useState<LinkModalState>({
    visible: false,
    masterSkuId: 0,
    platform: "shopee",
  });

  const handleAutoMap = async () => {
    setLoading(true);
    try {
      const skus =
        masterProduct.skus?.map((s) => s.seller_sku).filter(Boolean) || [];
      if (skus.length === 0) {
        message.warning("No SKUs to map");
        return;
      }
      const res = await autoMapSkus(skus);
      if (res.mapped_count > 0) {
        message.success(
          `Auto-mapping completed: ${res.mapped_count} mapped, ${res.skipped_count} skipped`,
        );
        onUpdate?.();
      } else {
        message.warning("No platform matches were linked from auto-map");
      }
    } catch (error: unknown) {
      message.error((error as { message?: string }).message || "Auto-mapping failed");
    } finally {
      setLoading(false);
    }
  };

  const handleUnlink = async (skuId: number, platform: Platform) => {
    setLoading(true);
    try {
      await unlinkSkuFromPlatform({ master_sku_id: skuId, platform });
      message.success(`Unlinked from ${platform}`);
      onUpdate?.();
    } catch (error: unknown) {
      message.error((error as { message?: string }).message || "Unlink failed");
    } finally {
      setLoading(false);
    }
  };

  const renderPlatformCell = (sku: MasterProductSku, platform: Platform) => {
    const link = sku.platform_links?.find((l) => l.platform === platform);

    if (link && link.sync_status !== "not_synced") {
      const badge = SYNC_STATUS_MAP[link.sync_status] || DEFAULT_STATUS;
      const pid = link.platform_product_id || link.platform_item_id || "Linked";
      return (
        <div className="flex items-center gap-2">
          <Tag color={badge.color} className="mr-0">
            {pid}
            {link.platform_sku_id && link.platform_sku_id !== pid
              ? ` / ${link.platform_sku_id}`
              : null}
          </Tag>
          <Tag color={badge.color} className="mr-0">{badge.label}</Tag>
          <Popconfirm
            title={`Unlink from ${platform}?`}
            onConfirm={() => handleUnlink(sku.id, platform)}
            okText="Yes"
            cancelText="No"
          >
            <Button
              type="text"
              size="small"
              danger
              icon={<DisconnectOutlined />}
              loading={loading}
              className="flex-shrink-0"
            />
          </Popconfirm>
        </div>
      );
    }

    return (
      <PlatformLinkButton
        platform={platform}
        onClick={() => {
          setLinkModal({ visible: true, masterSkuId: sku.id, platform });
        }}
      />
    );
  };

  const columns: ColumnsType<MasterProductSku> = [
    { title: "SKU", dataIndex: "seller_sku", key: "seller_sku", width: 150, fixed: "left" },
    { title: "Variant", dataIndex: "variant_name", key: "variant_name", width: 150 },
    { title: "Shopee", key: "shopee", width: 200, render: (_, r) => renderPlatformCell(r, "shopee") },
    { title: "TikTok", key: "tiktok", width: 200, render: (_, r) => renderPlatformCell(r, "tiktok") },
    { title: "Lazada", key: "lazada", width: 200, render: (_, r) => renderPlatformCell(r, "lazada") },
    {
      title: "Actions", key: "actions", width: 100, fixed: "right",
      render: (_, record) => {
        const linked = record.platform_links?.length || 0;
        return (
          <Badge
            count={`${linked}/3`}
            style={{ backgroundColor: linked === 3 ? token.colorSuccess : token.colorPrimary }}
          />
        );
      },
    },
  ];

  return (
    <Card
      title="SKU Mapping"
      extra={
        <Button type="primary" icon={<ApiOutlined />} onClick={handleAutoMap} loading={loading}>
          Auto Map
        </Button>
      }
      className="shadow-sm"
      styles={{ body: { padding: 0 } }}
    >
      <Table
        dataSource={masterProduct.skus || []}
        columns={columns}
        rowKey="id"
        pagination={false}
        scroll={{ x: 1000 }}
      />
      <SkuLinkModal
        state={linkModal}
        onClose={() => setLinkModal((prev) => ({ ...prev, visible: false }))}
        onSuccess={() => onUpdate?.()}
      />
    </Card>
  );
}
