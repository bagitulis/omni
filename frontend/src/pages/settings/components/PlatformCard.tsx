import { Button, Card, Tag, Typography, theme } from "antd";
import {
  LinkOutlined,
  DisconnectOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
} from "@ant-design/icons";

const { Text } = Typography;
const { useToken } = theme;

interface PlatformStatus {
  platform: string;
  connected: boolean;
  shop_id?: string;
  shop_name?: string;
  expires_at?: number;
  expires_soon?: boolean;
  last_checked?: string;
}

interface PlatformCardProps {
  platform: PlatformStatus;
  color: string;
  icon: React.ReactNode;
  name: string;
  onConnect: (platform: PlatformStatus) => void;
  onDisconnect: (platform: PlatformStatus) => void;
  formatExpiry: (expiresAt?: number) => string | null;
}

export type { PlatformStatus };

export function PlatformCard({
  platform,
  color,
  icon,
  name,
  onConnect,
  onDisconnect,
  formatExpiry,
}: PlatformCardProps) {
  const { token } = useToken();

  return (
    <Card
      style={{
        borderRadius: token.borderRadius,
        borderTop: `3px solid ${color}`,
        height: "100%",
      }}
      styles={{ body: { padding: 16 } }}
    >
      <div style={{ display: "flex", flexDirection: "column", gap: 12 }}>
        {/* Header */}
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
          }}
        >
          <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
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
                <Text style={{ fontSize: 12 }}>{platform.shop_id}</Text>
              </>
            )}
            {platform.expires_at && (
              <div style={{ marginTop: 8 }}>
                <Text
                  type={platform.expires_soon ? "danger" : "secondary"}
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
                onClick={() => onConnect(platform)}
                style={{ flex: 1 }}
              >
                Re-authorize
              </Button>
              <Button
                icon={<DisconnectOutlined />}
                danger
                onClick={() => onDisconnect(platform)}
              >
                Disconnect
              </Button>
            </div>
          ) : (
            <Button
              type="primary"
              icon={<LinkOutlined />}
              onClick={() => onConnect(platform)}
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
  );
}
