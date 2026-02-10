import {
  Modal,
  Tabs,
  Button,
  message,
  Table,
  Form,
  InputNumber,
  Card,
  Alert,
} from "antd";
import { useState } from "react";

export interface UpdateItem {
  sku: string;
  price: number;
  platform: "shopee" | "tiktok" | "lazada";
}

interface WholesaleUpdateModalProps {
  open: boolean;
  onClose: () => void;
  items: UpdateItem[];
}

interface WholesaleSettings {
  min_order_1: number;
  max_order_1: number;
  max_order_tier_3: number;
  // Simplified for this implementation
}

export function WholesaleUpdateModal({
  open,
  onClose,
  items,
}: WholesaleUpdateModalProps) {
  const [activeTab, setActiveTab] = useState("preview");
  const [processing, setProcessing] = useState(false);
  const [settings, setSettings] = useState<WholesaleSettings>({
    min_order_1: 2,
    max_order_1: 3,
    max_order_tier_3: 1000,
  });

  const handleUpdate = async () => {
    setProcessing(true);
    try {
      // Mock update
      await new Promise((resolve) => setTimeout(resolve, 1000));
      message.success("Wholesale updated successfully");
      onClose();
    } catch (error) {
      message.error("Failed to update wholesale");
    } finally {
      setProcessing(false);
    }
  };

  const PreviewTab = () => {
    const columns = [
      { title: "SKU", dataIndex: "sku", key: "sku" },
      {
        title: "Current Price",
        dataIndex: "price",
        key: "price",
        render: (val: number) => `Rp ${val.toLocaleString()}`,
      },
      {
        title: "Platform",
        dataIndex: "platform",
        key: "platform",
        render: (val: string) => val.toUpperCase(),
      },
    ];

    return (
      <div style={{ display: "flex", flexDirection: "column", gap: 16 }}>
        <Alert
          message={`Selected ${items.length} items for wholesale update`}
          type="info"
          showIcon
        />
        <Table
          dataSource={items}
          columns={columns}
          rowKey="sku"
          pagination={{ pageSize: 5 }}
          size="small"
        />
      </div>
    );
  };

  const SettingsTab = () => {
    return (
      <Form layout="vertical">
        <Card size="small" title="Tier Configuration">
          <div style={{ display: "flex", gap: 16 }}>
            <Form.Item label="Min Order Tier 1">
              <InputNumber
                value={settings.min_order_1}
                onChange={(val) =>
                  setSettings({ ...settings, min_order_1: val || 2 })
                }
              />
            </Form.Item>
            <Form.Item label="Max Order Tier 1">
              <InputNumber
                value={settings.max_order_1}
                onChange={(val) =>
                  setSettings({ ...settings, max_order_1: val || 3 })
                }
              />
            </Form.Item>
            <Form.Item label="Max Order Tier 3">
              <InputNumber
                value={settings.max_order_tier_3}
                onChange={(val) =>
                  setSettings({ ...settings, max_order_tier_3: val || 1000 })
                }
              />
            </Form.Item>
          </div>
        </Card>
      </Form>
    );
  };

  const itemsTab = [
    {
      key: "preview",
      label: "Preview",
      children: <PreviewTab />,
    },
    {
      key: "settings",
      label: "Settings",
      children: <SettingsTab />,
    },
  ];

  return (
    <Modal
      title="🛒 Update Wholesale - Shopee"
      open={open}
      onCancel={onClose}
      width={700}
      footer={[
        <Button key="cancel" onClick={onClose} disabled={processing}>
          Cancel
        </Button>,
        <Button
          key="update"
          type="primary"
          onClick={handleUpdate}
          loading={processing}
        >
          Update Wholesale
        </Button>,
      ]}
    >
      <Tabs
        activeKey={activeTab}
        onChange={setActiveTab}
        items={itemsTab}
        style={{ marginBottom: 16 }}
      />
    </Modal>
  );
}
