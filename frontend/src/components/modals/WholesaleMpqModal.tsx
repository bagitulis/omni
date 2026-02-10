import {
  Modal,
  Tabs,
  Button,
  message,
  Typography,
  Table,
  Form,
  InputNumber,
  Card,
  Alert,
  Tag,
  Space,
} from "antd";
import { useState } from "react";
import { UpdateItem } from "./WholesaleUpdateModal";

const { Text } = Typography;

interface WholesaleMpqModalProps {
  open: boolean;
  onClose: () => void;
  items: UpdateItem[];
}

export function WholesaleMpqModal({
  open,
  onClose,
  items,
}: WholesaleMpqModalProps) {
  const [activeTab, setActiveTab] = useState("wholesale");
  const [processing, setProcessing] = useState(false);

  const shopeeItems = items.filter((i) => i.platform === "shopee");
  const tiktokItems = items.filter((i) => i.platform === "tiktok");
  const lazadaItems = items.filter((i) => i.platform === "lazada");

  const handleUpdate = async () => {
    setProcessing(true);
    try {
      await new Promise((resolve) => setTimeout(resolve, 1000));
      message.success(`${activeTab.toUpperCase()} updated successfully`);
      onClose();
    } catch (error) {
      message.error("Failed to update");
    } finally {
      setProcessing(false);
    }
  };

  const WholesaleTab = () => (
    <div style={{ display: "flex", flexDirection: "column", gap: 16 }}>
      <Alert
        message={`Shopee Wholesale: ${shopeeItems.length} items`}
        type="info"
        showIcon
      />
      <Table
        dataSource={shopeeItems}
        columns={[
          { title: "SKU", dataIndex: "sku" },
          { title: "Price", dataIndex: "price" },
        ]}
        size="small"
        pagination={{ pageSize: 5 }}
        rowKey="sku"
      />
    </div>
  );

  const MpqTab = () => (
    <div style={{ display: "flex", flexDirection: "column", gap: 16 }}>
      <Alert
        message={`MPQ Update: ${shopeeItems.length + tiktokItems.length} items`}
        type="warning"
        showIcon
      />
      <Card title="MPQ Settings" size="small">
        <Form layout="inline">
          <Form.Item label="Minimum Order">
            <InputNumber min={1} defaultValue={1} />
          </Form.Item>
        </Form>
      </Card>
      <Table
        dataSource={[...shopeeItems, ...tiktokItems]}
        columns={[
          { title: "SKU", dataIndex: "sku" },
          { title: "Platform", dataIndex: "platform" },
        ]}
        size="small"
        pagination={{ pageSize: 5 }}
        rowKey={(record) => `${record.platform}-${record.sku}`}
      />
    </div>
  );

  const DeleteTab = () => (
    <div style={{ display: "flex", flexDirection: "column", gap: 16 }}>
      <Alert
        message="This will remove wholesale pricing from selected items"
        type="error"
        showIcon
      />
      <Table
        dataSource={items}
        columns={[
          { title: "SKU", dataIndex: "sku" },
          { title: "Platform", dataIndex: "platform" },
        ]}
        size="small"
        pagination={{ pageSize: 5 }}
        rowKey={(record) => `${record.platform}-${record.sku}`}
      />
    </div>
  );

  const SettingsTab = () => (
    <Card size="small" title="General Settings">
      <Form layout="vertical">
        <Form.Item label="Default Min Order">
          <InputNumber defaultValue={2} />
        </Form.Item>
      </Form>
    </Card>
  );

  const itemsTab = [
    { key: "wholesale", label: "Wholesale", children: <WholesaleTab /> },
    { key: "mpq", label: "MPQ", children: <MpqTab /> },
    { key: "delete", label: "Delete", children: <DeleteTab /> },
    { key: "settings", label: "Settings", children: <SettingsTab /> },
  ];

  return (
    <Modal
      title="🛒 Update Wholesale & MPQ"
      open={open}
      onCancel={onClose}
      width={700}
      footer={[
        <Button key="cancel" onClick={onClose} disabled={processing}>
          Cancel
        </Button>,
        <Button
          key="submit"
          type="primary"
          onClick={handleUpdate}
          loading={processing}
          danger={activeTab === "delete"}
        >
          {activeTab === "delete" ? "Delete Wholesale" : "Update"}
        </Button>,
      ]}
    >
      <Tabs
        activeKey={activeTab}
        onChange={setActiveTab}
        items={itemsTab}
        style={{ marginBottom: 16 }}
      />
      <div
        style={{
          marginTop: 16,
          padding: 12,
          background: "#f5f5f5",
          borderRadius: 6,
        }}
      >
        <Space size="large">
          <Text>
            Shopee: {shopeeItems.length}{" "}
            {["wholesale", "mpq"].includes(activeTab) && (
              <Tag color="success">Supported</Tag>
            )}
          </Text>
          <Text>
            TikTok: {tiktokItems.length}{" "}
            {activeTab === "mpq" ? (
              <Tag color="success">Supported</Tag>
            ) : (
              <Tag color="warning">MPQ Only</Tag>
            )}
          </Text>
          <Text>
            Lazada: {lazadaItems.length} <Tag color="error">Unsupported</Tag>
          </Text>
        </Space>
      </div>
    </Modal>
  );
}
