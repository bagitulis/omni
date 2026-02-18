import {
  ApiOutlined,
  DisconnectOutlined,
  LinkOutlined,
} from "@ant-design/icons";
import {
  Badge,
  Button,
  Card,
  Form,
  Input,
  Modal,
  message,
  Popconfirm,
  Table,
  Tag,
  theme,
} from "antd";
import type { ColumnsType } from "antd/es/table";
import { useState } from "react";
import {
  autoMapSkus,
  linkSkuToPlatform,
  unlinkSkuFromPlatform,
} from "@/api/products";
import { PLATFORM_BRAND_COLORS } from "@/lib/platformColors";
import type { MasterProduct, MasterProductSku } from "@/types/product";
import type { Platform } from "@/types/shared";

interface SkuMappingPanelProps {
  masterProduct: MasterProduct;
  onUpdate?: () => void;
}

interface LinkModalState {
  visible: boolean;
  masterSkuId: number;
  platform: Platform;
  initialValues?: {
    platform_item_id?: string;
    platform_sku_id?: string;
  };
}

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
  const [form] = Form.useForm();

  const handleAutoMap = async () => {
    setLoading(true);
    try {
      // Collect all seller_skus from the product
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
      const err = error as { message?: string };
      message.error(err.message || "Auto-mapping failed");
    } finally {
      setLoading(false);
    }
  };

  const openLinkModal = (skuId: number, platform: Platform) => {
    setLinkModal({
      visible: true,
      masterSkuId: skuId,
      platform,
    });
    form.resetFields();
  };

  const handleLinkSubmit = async () => {
    try {
      const values = await form.validateFields();
      setLoading(true);

      const payload = {
        master_sku_id: linkModal.masterSkuId,
        platform: linkModal.platform,
        platform_item_id: values.platform_item_id,
        platform_sku_id: values.platform_sku_id,
      };

      await linkSkuToPlatform(payload);
      message.success(`Linked to ${linkModal.platform}`);
      setLinkModal((prev) => ({ ...prev, visible: false }));
      onUpdate?.();
    } catch (error: unknown) {
      // Form validation error or API error
      const err = error as {
        errorFields?: unknown[];
        message?: string;
      };
      if (err.errorFields) return;
      message.error(err.message || "Link failed");
    } finally {
      setLoading(false);
    }
  };

  const handleUnlink = async (skuId: number, platform: Platform) => {
    setLoading(true);
    try {
      const payload = {
        master_sku_id: skuId,
        platform,
      };

      await unlinkSkuFromPlatform(payload);
      message.success(`Unlinked from ${platform}`);
      onUpdate?.();
    } catch (error: unknown) {
      const err = error as { message?: string };
      message.error(err.message || "Unlink failed");
    } finally {
      setLoading(false);
    }
  };

  const getPlatformLink = (sku: MasterProductSku, platform: Platform) => {
    return sku.platform_links?.find((link) => link.platform === platform);
  };

  const getSyncStatusBadge = (syncStatus: string) => {
    switch (syncStatus) {
      case "synced":
        return { color: "success" as const, label: "Synced" };
      case "outdated":
        return { color: "warning" as const, label: "Outdated" };
      case "pending":
      case "not_synced":
        return { color: "processing" as const, label: "Pending" };
      case "error":
      case "failed":
        return { color: "error" as const, label: "Error" };
      default:
        return { color: "default" as const, label: "Unknown" };
    }
  };

  const renderPlatformCell = (sku: MasterProductSku, platform: Platform) => {
    const link = getPlatformLink(sku, platform);

    if (link && link.sync_status !== "not_synced") {
      const badge = getSyncStatusBadge(link.sync_status);
      const platformProductID =
        link.platform_product_id || link.platform_item_id || "Linked";

      return (
        <div className="flex items-center gap-2">
          <Tag color={badge.color} className="mr-0">
            {platformProductID}
            {link.platform_sku_id && link.platform_sku_id !== platformProductID
              ? ` / ${link.platform_sku_id}`
              : null}
          </Tag>
          <Tag color={badge.color} className="mr-0">
            {badge.label}
          </Tag>
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
      <Button
        size="small"
        type="default"
        icon={<LinkOutlined />}
        onClick={() => openLinkModal(sku.id, platform)}
        style={{
          color: PLATFORM_BRAND_COLORS[platform],
          borderColor: PLATFORM_BRAND_COLORS[platform],
        }}
      >
        Link
      </Button>
    );
  };

  const columns: ColumnsType<MasterProductSku> = [
    {
      title: "SKU",
      dataIndex: "seller_sku",
      key: "seller_sku",
      width: 150,
      fixed: "left",
    },
    {
      title: "Variant",
      dataIndex: "variant_name",
      key: "variant_name",
      width: 150,
    },
    {
      title: "Shopee",
      key: "shopee",
      width: 200,
      render: (_, record) => renderPlatformCell(record, "shopee"),
    },
    {
      title: "TikTok",
      key: "tiktok",
      width: 200,
      render: (_, record) => renderPlatformCell(record, "tiktok"),
    },
    {
      title: "Lazada",
      key: "lazada",
      width: 200,
      render: (_, record) => renderPlatformCell(record, "lazada"),
    },
    {
      title: "Actions",
      key: "actions",
      width: 100,
      fixed: "right",
      render: (_, record) => {
        const linkedCount = record.platform_links?.length || 0;
        const totalPlatforms = 3; // Shopee, TikTok, Lazada
        return (
          <Badge
            count={`${linkedCount}/${totalPlatforms}`}
            style={{
              backgroundColor:
                linkedCount === totalPlatforms
                  ? token.colorSuccess
                  : token.colorPrimary,
            }}
          />
        );
      },
    },
  ];

  return (
    <Card
      title="SKU Mapping"
      extra={
        <Button
          type="primary"
          icon={<ApiOutlined />}
          onClick={handleAutoMap}
          loading={loading}
        >
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

      <Modal
        title={`Link to ${linkModal.platform}`}
        open={linkModal.visible}
        onOk={handleLinkSubmit}
        onCancel={() => setLinkModal((prev) => ({ ...prev, visible: false }))}
        confirmLoading={loading}
        destroyOnHidden
      >
        <Form form={form} layout="vertical">
          <Form.Item
            name="platform_item_id"
            label="Platform Product ID"
            rules={[
              { required: true, message: "Please enter Platform Product ID" },
            ]}
          >
            <Input placeholder="e.g. 123456789" />
          </Form.Item>
          <Form.Item name="platform_sku_id" label="Platform SKU ID (Optional)">
            <Input placeholder="e.g. 987654321" />
          </Form.Item>
        </Form>
      </Modal>
    </Card>
  );
}
