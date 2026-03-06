import { CopyOutlined } from "@ant-design/icons";
import { Button, Card, Space, Typography } from "antd";
import { message } from "@/components/AntStaticHolder";

const { Text } = Typography;

interface IntegrationUrlItem {
  key: string;
  label: string;
  url: string;
}

function buildIntegrationUrls(baseUrl: string): IntegrationUrlItem[] {
  return [
    {
      key: "shopee",
      label: "Shopee",
      url: `${baseUrl}/api/webhooks/shopee`,
    },
    {
      key: "tiktok",
      label: "TikTok",
      url: `${baseUrl}/api/webhooks/tiktok`,
    },
    {
      key: "lazada",
      label: "Lazada",
      url: `${baseUrl}/api/webhooks/lazada`,
    },
  ];
}

export default function IntegrationUrls() {
  const baseUrl = typeof window !== "undefined" ? window.location.origin : "";
  const integrationUrls = buildIntegrationUrls(baseUrl);

  const handleCopy = async (url: string) => {
    await navigator.clipboard.writeText(url);
    message.success("Copied to clipboard");
  };

  return (
    <Card title="Integration URLs" size="small" style={{ marginBottom: 16 }}>
      <Space direction="vertical" size={12} style={{ width: "100%" }}>
        {integrationUrls.map((item) => (
          <div
            key={item.key}
            style={{
              display: "flex",
              justifyContent: "space-between",
              alignItems: "center",
              gap: 12,
            }}
          >
            <div style={{ minWidth: 0 }}>
              <Text strong style={{ display: "block", fontSize: 12 }}>
                {item.label} Webhook URL
              </Text>
              <Text
                code
                style={{
                  fontSize: 12,
                  fontFamily:
                    "ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, 'Liberation Mono', 'Courier New', monospace",
                  whiteSpace: "normal",
                  overflowWrap: "anywhere",
                }}
              >
                {item.url}
              </Text>
            </div>
            <Button
              icon={<CopyOutlined />}
              size="small"
              onClick={() => void handleCopy(item.url)}
            >
              Copy
            </Button>
          </div>
        ))}
      </Space>
    </Card>
  );
}
