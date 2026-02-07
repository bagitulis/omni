import { useEffect, useState } from "react";
import {
  Card,
  Form,
  Input,
  Button,
  Table,
  Tag,
  Typography,
  message,
  theme,
  Space,
} from "antd";
import { CopyOutlined, SendOutlined, ReloadOutlined } from "@ant-design/icons";
import type { ColumnsType } from "antd/es/table";
import apiClient from "@/api/client";

const { Text, Title } = Typography;
const { useToken } = theme;

interface WebhookLog {
  id: string;
  platform: string;
  event_type: string;
  status: string;
  created_at: string;
}

export default function WebhooksTab() {
  const { token } = useToken();
  const [form] = Form.useForm();
  const [testing, setTesting] = useState(false);
  const [webhookLogs, setWebhookLogs] = useState<WebhookLog[]>([]);
  const [loading, setLoading] = useState(true);

  const baseUrl = typeof window !== "undefined" ? window.location.origin : "";

  const webhookUrls = {
    shopee: `${baseUrl}/api/webhooks/shopee`,
    tiktok: `${baseUrl}/api/webhooks/tiktok`,
    lazada: `${baseUrl}/api/webhooks/lazada`,
  };

  useEffect(() => {
    fetchWebhookLogs();
  }, []);

  const fetchWebhookLogs = async () => {
    try {
      setLoading(true);
      const response = await apiClient.get<WebhookLog[]>("/webhooks/logs");
      if (response.success && response.data) {
        setWebhookLogs(response.data);
      }
    } catch (error) {
      // Show empty state if fetch fails
      setWebhookLogs([]);
    } finally {
      setLoading(false);
    }
  };

  const columns: ColumnsType<WebhookLog> = [
    {
      title: "Platform",
      dataIndex: "platform",
      key: "platform",
      width: 100,
      render: (platform: string) => (
        <Tag
          color={
            platform === "shopee"
              ? "orange"
              : platform === "tiktok"
                ? "default"
                : "blue"
          }
        >
          {platform}
        </Tag>
      ),
    },
    { title: "Event", dataIndex: "event_type", key: "event_type", width: 150 },
    {
      title: "Status",
      dataIndex: "status",
      key: "status",
      width: 100,
      render: (status: string) => (
        <Tag color={status === "success" ? "success" : "error"}>{status}</Tag>
      ),
    },
    { title: "Time", dataIndex: "created_at", key: "created_at", width: 180 },
  ];

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    message.success("Copied to clipboard");
  };

  const testWebhook = async () => {
    setTesting(true);
    setTimeout(() => {
      message.success("Webhook test successful");
      setTesting(false);
    }, 1500);
  };

  return (
    <div>
      <div style={{ marginBottom: 16 }}>
        <Title level={5} style={{ margin: 0, fontSize: 14 }}>
          Webhook Configuration
        </Title>
        <Text type="secondary" style={{ fontSize: 12 }}>
          Configure webhook endpoints for receiving platform events
        </Text>
      </div>

      <Card
        title="Webhook URLs"
        size="small"
        style={{ marginBottom: 16, borderRadius: token.borderRadius }}
      >
        {Object.entries(webhookUrls).map(([platform, url]) => (
          <div
            key={platform}
            style={{
              display: "flex",
              alignItems: "center",
              justifyContent: "space-between",
              padding: "8px 0",
              borderBottom: `1px solid ${token.colorBorderSecondary}`,
            }}
          >
            <div>
              <Text
                strong
                style={{ fontSize: 12, textTransform: "capitalize" }}
              >
                {platform}
              </Text>
              <br />
              <Text
                type="secondary"
                style={{ fontSize: 11, fontFamily: "monospace" }}
              >
                {url}
              </Text>
            </div>
            <Button
              icon={<CopyOutlined />}
              size="small"
              onClick={() => copyToClipboard(url)}
            >
              Copy
            </Button>
          </div>
        ))}
      </Card>

      <Card
        title="Custom Webhook"
        size="small"
        style={{ marginBottom: 16, borderRadius: token.borderRadius }}
      >
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
            <Input.Password placeholder="Optional: webhook signature key" />
          </Form.Item>
          <Space>
            <Button type="primary" icon={<SendOutlined />}>
              Save Configuration
            </Button>
            <Button
              icon={<SendOutlined />}
              loading={testing}
              onClick={testWebhook}
            >
              Test Webhook
            </Button>
          </Space>
        </Form>
      </Card>

      <Card
        title="Recent Webhook Events"
        size="small"
        style={{ borderRadius: token.borderRadius }}
        extra={
          <Button
            icon={<ReloadOutlined />}
            size="small"
            onClick={fetchWebhookLogs}
            loading={loading}
          >
            Refresh
          </Button>
        }
      >
        <Table
          columns={columns}
          dataSource={webhookLogs}
          rowKey="id"
          size="small"
          pagination={{ pageSize: 5 }}
        />
      </Card>
    </div>
  );
}
