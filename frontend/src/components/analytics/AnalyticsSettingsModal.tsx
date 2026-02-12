import { Modal, Form, InputNumber, Button, message, Tabs, Select } from "antd";
import { useState, useEffect, useCallback } from "react";
import api from "../../api/client";

interface Props {
  open: boolean;
  onClose: () => void;
}

export const AnalyticsSettingsModal = ({ open, onClose }: Props) => {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);

  // Load settings
  const fetchSettings = useCallback(async () => {
    setLoading(true);
    try {
      const response = await api.get("/analytics/settings");
      if (response.data.success) {
        form.setFieldsValue(response.data.data);
      }
    } catch (error) {
      console.error("Failed to load settings:", error);
      message.error("Failed to load settings");
    } finally {
      setLoading(false);
    }
  }, [form]);

  useEffect(() => {
    if (open) {
      fetchSettings();
    }
  }, [open, fetchSettings]);

  const handleSave = async () => {
    try {
      const values = await form.validateFields();
      setSaving(true);
      const response = await api.post("/analytics/settings", values);

      if (response.data.success) {
        message.success("Settings saved successfully");
        onClose();
      } else {
        message.error(response.data.error || "Failed to save settings");
      }
    } catch (error) {
      console.error("Save error:", error);
      message.error("Failed to save settings");
    } finally {
      setSaving(false);
    }
  };

  const GeneralSettings = () => (
    <>
      <Form.Item
        name={["general", "currency"]}
        label="Currency"
        initialValue="IDR"
      >
        <Select disabled>
          <Select.Option value="IDR">IDR (Indonesian Rupiah)</Select.Option>
        </Select>
      </Form.Item>
      <Form.Item
        name={["general", "timezone"]}
        label="Timezone"
        initialValue="Asia/Jakarta"
      >
        <Select disabled>
          <Select.Option value="Asia/Jakarta">
            Asia/Jakarta (GMT+7)
          </Select.Option>
        </Select>
      </Form.Item>
    </>
  );

  const ThresholdSettings = ({ platform }: { platform: string }) => (
    <>
      <Form.Item
        name={[platform, "high_roas_threshold"]}
        label="High ROAS Threshold"
        tooltip="ROAS above this value is considered high performance"
        rules={[{ required: true }]}
      >
        <InputNumber min={0} step={0.1} style={{ width: "100%" }} />
      </Form.Item>
      <Form.Item
        name={[platform, "low_roas_threshold"]}
        label="Low ROAS Threshold"
        tooltip="ROAS below this value is considered low performance"
        rules={[{ required: true }]}
      >
        <InputNumber min={0} step={0.1} style={{ width: "100%" }} />
      </Form.Item>
      <Form.Item
        name={[platform, "min_spend_threshold"]}
        label="Minimum Spend Threshold"
        tooltip="Minimum spend required before analyzing performance"
        rules={[{ required: true }]}
      >
        <InputNumber
          min={0}
          step={1000}
          addonBefore="Rp"
          style={{ width: "100%" }}
        />
      </Form.Item>
    </>
  );

  return (
    <Modal
      title="Analytics Settings"
      open={open}
      onCancel={onClose}
      footer={[
        <Button key="cancel" onClick={onClose}>
          Cancel
        </Button>,
        <Button key="save" type="primary" onClick={handleSave} loading={saving}>
          Save Changes
        </Button>,
      ]}
      width={600}
    >
      <Form form={form} layout="vertical" disabled={loading}>
        <Tabs
          defaultActiveKey="general"
          items={[
            {
              key: "general",
              label: "General",
              children: <GeneralSettings />,
            },
            {
              key: "tiktok",
              label: "TikTok Ads",
              children: <ThresholdSettings platform="tiktok" />,
            },
            {
              key: "shopee",
              label: "Shopee Ads",
              children: <ThresholdSettings platform="shopee" />,
            },
          ]}
        />
      </Form>
    </Modal>
  );
};
