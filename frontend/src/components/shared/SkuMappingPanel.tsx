import { useState } from "react";
import {
  Card,
  Table,
  Button,
  Tag,
  Modal,
  Form,
  Input,
  Popconfirm,
  message,
  Badge,
} from "antd";
import {
  LinkOutlined,
  DisconnectOutlined,
  ApiOutlined,
} from "@ant-design/icons";
import type { ColumnsType } from "antd/es/table";
import { apiClient } from "@/api/client";
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
    platform_product_id?: string;
    platform_sku_id?: string;
  };
}

export function SkuMappingPanel({
  masterProduct,
  onUpdate,
}: SkuMappingPanelProps) {
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
      const res = await apiClient.post("/products/auto-map", {
        master_product_id: masterProduct.id,
      });
      if (res.success) {
        message.success("Auto-mapping completed");
        onUpdate?.();
      } else {
        message.error(res.error || "Auto-mapping failed");
      }
    } catch (error: unknown) {
      const err = error as {
        response?: { data?: { error?: string } };
        message?: string;
      };
      message.error(
        err.response?.data?.error || err.message || "Auto-mapping failed",
      );
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
        platform_product_id: values.platform_product_id,
        platform_sku_id: values.platform_sku_id,
      };

      const res = await apiClient.post("/products/link-sku", payload);

      if (res.success) {
        message.success(`Linked to ${linkModal.platform}`);
        setLinkModal((prev) => ({ ...prev, visible: false }));
        onUpdate?.();
      } else {
        message.error(res.error || "Link failed");
      }
    } catch (error: unknown) {
      // Form validation error or API error
      const err = error as {
        errorFields?: unknown[];
        response?: { data?: { error?: string } };
        message?: string;
      };
      if (err.errorFields) return;
      message.error(err.response?.data?.error || err.message || "Link failed");
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

      const res = await apiClient.post("/products/unlink-sku", payload);

      if (res.success) {
        message.success(`Unlinked from ${platform}`);
        onUpdate?.();
      } else {
        message.error(res.error || "Unlink failed");
      }
    } catch (error: unknown) {
      const err = error as {
        response?: { data?: { error?: string } };
        message?: string;
      };
      message.error(
        err.response?.data?.error || err.message || "Unlink failed",
      );
    } finally {
      setLoading(false);
    }
  };

  const getPlatformLink = (sku: MasterProductSku, platform: Platform) => {
    return sku.platform_links?.find((link) => link.platform === platform);
  };

  const renderPlatformCell = (sku: MasterProductSku, platform: Platform) => {
    const link = getPlatformLink(sku, platform);
    const platformColors: Record<Platform, string> = {
      shopee: "#ee4d2d",
      lazada: "#0f146d",
      tiktok: "#000000",
    };

    if (link && link.sync_status !== "not_synced") {
      return (
        <div className="flex items-center gap-2">
          <Tag color="success" className="mr-0">
            {link.platform_product_id}
            {link.platform_sku_id && ` / ${link.platform_sku_id}`}
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
          color: platformColors[platform],
          borderColor: platformColors[platform],
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
                linkedCount === totalPlatforms ? "#52c41a" : "#1890ff",
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
      bodyStyle={{ padding: 0 }}
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
        destroyOnClose
      >
        <Form form={form} layout="vertical">
          <Form.Item
            name="platform_product_id"
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
