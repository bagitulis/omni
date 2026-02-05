import { Card, Row, Col, Button, Tag, Typography, message, theme } from "antd";
import {
  ShopOutlined,
  VideoCameraOutlined,
  ShoppingOutlined,
  LinkOutlined,
  DisconnectOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
} from "@ant-design/icons";

const { Text, Title } = Typography;
const { useToken } = theme;

// Platform brand colors
const PLATFORM_COLORS = {
  shopee: "#ee4d2d",
  tiktok: "#000000",
  lazada: "#0f146d",
};

interface Platform {
  id: string;
  name: string;
  icon: React.ReactNode;
  color: string;
  connected: boolean;
  shop_name?: string;
  expires_at?: string;
}

export default function PlatformsTab() {
  const { token } = useToken();

  const platforms: Platform[] = [
    {
      id: "shopee",
      name: "Shopee",
      icon: (
        <ShopOutlined style={{ fontSize: 24, color: PLATFORM_COLORS.shopee }} />
      ),
      color: PLATFORM_COLORS.shopee,
      connected: true,
      shop_name: "My Shopee Store",
      expires_at: "2026-03-15",
    },
    {
      id: "tiktok",
      name: "TikTok Shop",
      icon: (
        <VideoCameraOutlined
          style={{ fontSize: 24, color: PLATFORM_COLORS.tiktok }}
        />
      ),
      color: PLATFORM_COLORS.tiktok,
      connected: false,
    },
    {
      id: "lazada",
      name: "Lazada",
      icon: (
        <ShoppingOutlined
          style={{ fontSize: 24, color: PLATFORM_COLORS.lazada }}
        />
      ),
      color: PLATFORM_COLORS.lazada,
      connected: true,
      shop_name: "Lazada Official",
      expires_at: "2026-04-01",
    },
  ];

  const handleConnect = (platform: Platform) => {
    console.log("Connecting to:", platform.name);
    message.info(`Opening ${platform.name} authorization...`);
  };

  const handleDisconnect = (platform: Platform) => {
    console.log("Disconnecting:", platform.name);
    message.warning(`Disconnected from ${platform.name}`);
  };

  const formatExpiry = (dateStr?: string) => {
    if (!dateStr) return null;
    const date = new Date(dateStr);
    const now = new Date();
    const days = Math.ceil(
      (date.getTime() - now.getTime()) / (1000 * 60 * 60 * 24),
    );
    if (days < 0) return "Expired";
    if (days === 0) return "Expires today";
    if (days <= 7) return `Expires in ${days} days`;
    return `Expires ${date.toLocaleDateString()}`;
  };

  return (
    <div>
      <div style={{ marginBottom: 16 }}>
        <Title level={5} style={{ margin: 0, fontSize: 14 }}>
          Platform Connections
        </Title>
        <Text type="secondary" style={{ fontSize: 12 }}>
          Connect your marketplace accounts to sync orders and products
        </Text>
      </div>

      <Row gutter={[16, 16]}>
        {platforms.map((platform) => (
          <Col xs={24} md={8} key={platform.id}>
            <Card
              style={{
                borderRadius: token.borderRadius,
                borderTop: `3px solid ${platform.color}`,
                height: "100%",
              }}
              styles={{ body: { padding: 16 } }}
            >
              <div
                style={{ display: "flex", flexDirection: "column", gap: 12 }}
              >
                {/* Header */}
                <div
                  style={{
                    display: "flex",
                    justifyContent: "space-between",
                    alignItems: "center",
                  }}
                >
                  <div
                    style={{ display: "flex", alignItems: "center", gap: 8 }}
                  >
                    <div
                      style={{
                        width: 40,
                        height: 40,
                        borderRadius: token.borderRadius,
                        background: `${platform.color}15`,
                        display: "flex",
                        alignItems: "center",
                        justifyContent: "center",
                      }}
                    >
                      {platform.icon}
                    </div>
                    <Text strong style={{ fontSize: 14 }}>
                      {platform.name}
                    </Text>
                  </div>
                  <Tag
                    icon={
                      platform.connected ? (
                        <CheckCircleOutlined />
                      ) : (
                        <CloseCircleOutlined />
                      )
                    }
                    color={platform.connected ? "success" : "default"}
                    style={{ margin: 0 }}
                  >
                    {platform.connected ? "Connected" : "Disconnected"}
                  </Tag>
                </div>

                {/* Shop Info */}
                {platform.connected && (
                  <div
                    style={{
                      padding: 12,
                      background: token.colorFillTertiary,
                      borderRadius: token.borderRadius,
                    }}
                  >
                    <div style={{ marginBottom: 4 }}>
                      <Text type="secondary" style={{ fontSize: 10 }}>
                        SHOP NAME
                      </Text>
                    </div>
                    <Text style={{ fontSize: 12 }}>{platform.shop_name}</Text>
                    {platform.expires_at && (
                      <div style={{ marginTop: 8 }}>
                        <Text type="secondary" style={{ fontSize: 10 }}>
                          {formatExpiry(platform.expires_at)}
                        </Text>
                      </div>
                    )}
                  </div>
                )}

                {/* Actions */}
                <div style={{ marginTop: "auto" }}>
                  {platform.connected ? (
                    <div style={{ display: "flex", gap: 8 }}>
                      <Button
                        icon={<LinkOutlined />}
                        onClick={() => handleConnect(platform)}
                        style={{ flex: 1 }}
                      >
                        Re-authorize
                      </Button>
                      <Button
                        icon={<DisconnectOutlined />}
                        danger
                        onClick={() => handleDisconnect(platform)}
                      >
                        Disconnect
                      </Button>
                    </div>
                  ) : (
                    <Button
                      type="primary"
                      icon={<LinkOutlined />}
                      onClick={() => handleConnect(platform)}
                      style={{
                        width: "100%",
                        background: platform.color,
                        borderColor: platform.color,
                      }}
                    >
                      Connect {platform.name}
                    </Button>
                  )}
                </div>
              </div>
            </Card>
          </Col>
        ))}
      </Row>
    </div>
  );
}
