import { useEffect, useState } from "react";
import {
  Card,
  Row,
  Col,
  Button,
  Tag,
  Typography,
  message,
  theme,
  Spin,
} from "antd";
import {
  ShopOutlined,
  VideoCameraOutlined,
  ShoppingOutlined,
  LinkOutlined,
  DisconnectOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  LoadingOutlined,
} from "@ant-design/icons";
import apiClient from "@/api/client";

const { Text, Title } = Typography;
const { useToken } = theme;

// Platform brand colors
const PLATFORM_COLORS: Record<string, string> = {
  shopee: "#ee4d2d",
  tiktok: "#000000",
  lazada: "#0f146d",
};

// Platform icons
const PLATFORM_ICONS: Record<string, React.ReactNode> = {
  shopee: (
    <ShopOutlined style={{ fontSize: 24, color: PLATFORM_COLORS.shopee }} />
  ),
  tiktok: (
    <VideoCameraOutlined
      style={{ fontSize: 24, color: PLATFORM_COLORS.tiktok }}
    />
  ),
  lazada: (
    <ShoppingOutlined style={{ fontSize: 24, color: PLATFORM_COLORS.lazada }} />
  ),
};

// Platform display names
const PLATFORM_NAMES: Record<string, string> = {
  shopee: "Shopee",
  tiktok: "TikTok Shop",
  lazada: "Lazada",
};

interface PlatformStatus {
  platform: string;
  connected: boolean;
  shop_id?: string;
  shop_name?: string;
  expires_at?: number;
  expires_soon?: boolean;
  last_checked?: string;
}

interface PlatformAuthResponse {
  shopee?: PlatformStatus;
  tiktok?: PlatformStatus;
  lazada?: PlatformStatus;
}

export default function PlatformsTab() {
  const { token } = useToken();
  const [platforms, setPlatforms] = useState<PlatformStatus[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchPlatformStatus();
  }, []);

  const fetchPlatformStatus = async () => {
    try {
      setLoading(true);
      const response = await apiClient.get<PlatformAuthResponse>(
        "/platform-auth/status",
      );

      if (response.success && response.data) {
        const platformList: PlatformStatus[] = [];
        const data = response.data;

        // Add platforms in order
        if (data.shopee) platformList.push(data.shopee);
        if (data.tiktok) platformList.push(data.tiktok);
        if (data.lazada) platformList.push(data.lazada);

        setPlatforms(platformList);
      }
    } catch (error) {
      console.error("Failed to fetch platform status:", error);
      message.error("Failed to load platform status");
    } finally {
      setLoading(false);
    }
  };

  const handleConnect = (platform: PlatformStatus) => {
    const platformName = PLATFORM_NAMES[platform.platform] || platform.platform;
    message.info(`Opening ${platformName} authorization...`);
    // TODO: Implement actual OAuth flow
    window.open(`/api/platform-auth/${platform.platform}/authorize`, "_blank");
  };

  const handleDisconnect = async (platform: PlatformStatus) => {
    const platformName = PLATFORM_NAMES[platform.platform] || platform.platform;
    try {
      await apiClient.get(`/platform-auth/${platform.platform}/disconnect`);
      message.success(`Disconnected from ${platformName}`);
      fetchPlatformStatus();
    } catch (error) {
      message.error(`Failed to disconnect from ${platformName}`);
    }
  };

  const formatExpiry = (expiresAt?: number) => {
    if (!expiresAt) return null;
    const date = new Date(expiresAt);
    const now = new Date();
    const days = Math.ceil(
      (date.getTime() - now.getTime()) / (1000 * 60 * 60 * 24),
    );
    if (days < 0) return "Expired";
    if (days === 0) return "Expires today";
    if (days <= 7) return `Expires in ${days} days`;
    return `Expires ${date.toLocaleDateString()}`;
  };

  if (loading) {
    return (
      <div
        style={{
          display: "flex",
          justifyContent: "center",
          alignItems: "center",
          minHeight: 200,
        }}
      >
        <Spin indicator={<LoadingOutlined spin />} size="large" />
      </div>
    );
  }

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
        {platforms.map((platform) => {
          const platformId = platform.platform;
          const color = PLATFORM_COLORS[platformId] || "#666";
          const icon = PLATFORM_ICONS[platformId];
          const name = PLATFORM_NAMES[platformId] || platformId;

          return (
            <Col xs={24} md={8} key={platformId}>
              <Card
                style={{
                  borderRadius: token.borderRadius,
                  borderTop: `3px solid ${color}`,
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
                          background: `${color}15`,
                          display: "flex",
                          alignItems: "center",
                          justifyContent: "center",
                        }}
                      >
                        {icon}
                      </div>
                      <Text strong style={{ fontSize: 14 }}>
                        {name}
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
                      {platform.shop_id && (
                        <>
                          <div style={{ marginBottom: 4 }}>
                            <Text type="secondary" style={{ fontSize: 10 }}>
                              SHOP ID
                            </Text>
                          </div>
                          <Text style={{ fontSize: 12 }}>
                            {platform.shop_id}
                          </Text>
                        </>
                      )}
                      {platform.expires_at && (
                        <div style={{ marginTop: 8 }}>
                          <Text
                            type={
                              platform.expires_soon ? "danger" : "secondary"
                            }
                            style={{ fontSize: 10 }}
                          >
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
                          background: color,
                          borderColor: color,
                        }}
                      >
                        Connect {name}
                      </Button>
                    )}
                  </div>
                </div>
              </Card>
            </Col>
          );
        })}
      </Row>
    </div>
  );
}
