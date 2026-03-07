import { useCallback, useEffect, useState } from "react";
import { Row, Col, Typography, Spin, theme } from "antd";
import {
  ShopOutlined,
  VideoCameraOutlined,
  ShoppingOutlined,
  LoadingOutlined,
} from "@ant-design/icons";
import apiClient from "@/api/client";
import { PlatformCard, type PlatformStatus } from "../components/PlatformCard";
import { message } from "@/components/AntStaticHolder";

const { Text, Title } = Typography;

// Platform display names
const PLATFORM_NAMES: Record<string, string> = {
  shopee: "Shopee",
  tiktok: "TikTok Shop",
  lazada: "Lazada",
};

interface PlatformAuthResponse {
  shopee?: PlatformStatus;
  tiktok?: PlatformStatus;
  lazada?: PlatformStatus;
}

export default function PlatformsTab() {
  const { token } = theme.useToken();
  const [platforms, setPlatforms] = useState<PlatformStatus[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchPlatformStatus = useCallback(async () => {
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
    } catch (err) { console.warn("Operation failed:", err);
      message.error("Failed to load platform status");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchPlatformStatus();
  }, [fetchPlatformStatus]);

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
    } catch (err) { console.warn("Operation failed:", err);
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
          const colorMap: Record<string, string> = {
            shopee: token.colorPrimary,
            tiktok: token.colorText,
            lazada: token.colorInfo,
          };
          const iconMap: Record<string, React.ReactNode> = {
            shopee: (
              <ShopOutlined
                style={{ fontSize: 24, color: token.colorPrimary }}
              />
            ),
            tiktok: (
              <VideoCameraOutlined
                style={{ fontSize: 24, color: token.colorText }}
              />
            ),
            lazada: (
              <ShoppingOutlined
                style={{ fontSize: 24, color: token.colorInfo }}
              />
            ),
          };
          const color = colorMap[platformId] || token.colorTextSecondary;
          const icon = iconMap[platformId];
          const name = PLATFORM_NAMES[platformId] || platformId;

          return (
            <Col xs={24} md={8} key={platformId}>
              <PlatformCard
                platform={platform}
                color={color}
                icon={icon}
                name={name}
                onConnect={handleConnect}
                onDisconnect={handleDisconnect}
                formatExpiry={formatExpiry}
              />
            </Col>
          );
        })}
      </Row>
    </div>
  );
}
