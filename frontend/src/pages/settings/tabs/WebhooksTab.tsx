import { useState } from "react";
import {
  Button,
  Card,
  Form,
  Input,
  Space,
  Tabs,
  Typography,
} from "antd";
import { SendOutlined } from "@ant-design/icons";
import { saveWebhookConfig } from "@/api/settings";
import IntegrationUrls from "../components/IntegrationUrls";
import WebhookLogsViewer from "../components/WebhookLogsViewer";
import OAuthLogsViewer from "../components/OAuthLogsViewer";
import { message } from "@/components/AntStaticHolder";

const { Text, Title } = Typography;

export default function WebhooksTab() {
  const [form] = Form.useForm();
  const [saving, setSaving] = useState(false);

  const handleSaveConfig = async () => {
    const values = form.getFieldsValue();
    if (!values.custom_url) {
      message.warning("Please enter a webhook URL");
      return;
    }

    setSaving(true);
    try {
      await saveWebhookConfig({
        custom_url: values.custom_url,
        secret_key: values.secret_key,
      });
      message.success("Webhook configuration saved");
    } catch {
      message.error("Failed to save webhook configuration");
    } finally {
      setSaving(false);
    }
  };

  const tabItems = [
    {
      key: "configuration",
      label: "Configuration",
      children: (
        <div>
          <IntegrationUrls />

          <Card title="Custom Webhook" size="small">
            <Form form={form} layout="vertical">
              <Form.Item
                label="Webhook URL"
                name="custom_url"
                style={{ marginBottom: 12 }}
              >
                <Input placeholder="https://your-server.com/webhook" />
              </Form.Item>
              <Form.Item
                label="Secret Key"
                name="secret_key"
                style={{ marginBottom: 12 }}
              >
                <Input.Password
                  placeholder="Optional: webhook signature key"
                  autoComplete="off"
                />
              </Form.Item>
              <Space>
                <Button
                  type="primary"
                  icon={<SendOutlined />}
                  onClick={handleSaveConfig}
                  loading={saving}
                >
                  Save Configuration
                </Button>
              </Space>
            </Form>
          </Card>
        </div>
      ),
    },
    {
      key: "webhook_logs",
      label: "Webhook Logs",
      children: <WebhookLogsViewer />,
    },
    {
      key: "oauth_logs",
      label: "OAuth Logs",
      children: <OAuthLogsViewer />,
    },
  ];

  return (
    <div>
      <div style={{ marginBottom: 16 }}>
        <Title level={5} style={{ margin: 0, fontSize: 14 }}>
          Webhook Configuration
        </Title>
        <Text type="secondary" style={{ fontSize: 12 }}>
          Configure webhook endpoints and review integration activity logs
        </Text>
      </div>

      <Tabs defaultActiveKey="configuration" items={tabItems} />
    </div>
  );
}
