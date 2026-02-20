import { useState } from "react";
import { Button, Form, Input, Modal, message } from "antd";
import { LinkOutlined } from "@ant-design/icons";
import { linkSkuToPlatform } from "@/api/products";
import type { Platform } from "@/types/shared";
import { PLATFORM_BRAND_COLORS } from "@/lib/platformColors";

interface LinkModalState {
  visible: boolean;
  masterSkuId: number;
  platform: Platform;
}

interface SkuLinkModalProps {
  state: LinkModalState;
  onClose: () => void;
  onSuccess: () => void;
}

/**
 * Modal for linking a master SKU to a platform product/SKU.
 * Extracted from SkuMappingPanel for SRP.
 */
export function SkuLinkModal({ state, onClose, onSuccess }: SkuLinkModalProps) {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      setLoading(true);

      await linkSkuToPlatform({
        master_sku_id: state.masterSkuId,
        platform: state.platform,
        platform_item_id: values.platform_item_id,
        platform_sku_id: values.platform_sku_id,
      });

      message.success(`Linked to ${state.platform}`);
      onClose();
      onSuccess();
    } catch (error: unknown) {
      const err = error as { errorFields?: unknown[]; message?: string };
      if (err.errorFields) return;
      message.error(err.message || "Link failed");
    } finally {
      setLoading(false);
    }
  };

  return (
    <Modal
      title={`Link to ${state.platform}`}
      open={state.visible}
      onOk={handleSubmit}
      onCancel={onClose}
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
  );
}

// --- Helper: Open Link Button ---

interface LinkButtonProps {
  platform: Platform;
  onClick: () => void;
}

export function PlatformLinkButton({ platform, onClick }: LinkButtonProps) {
  return (
    <Button
      size="small"
      type="default"
      icon={<LinkOutlined />}
      onClick={onClick}
      style={{
        color: PLATFORM_BRAND_COLORS[platform],
        borderColor: PLATFORM_BRAND_COLORS[platform],
      }}
    >
      Link
    </Button>
  );
}

export type { LinkModalState };
