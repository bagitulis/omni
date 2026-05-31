import {
  CheckCircleOutlined,
  CloseCircleOutlined,
  DisconnectOutlined,
  LinkOutlined,
} from "@ant-design/icons";
import { Button, Card, Space, Tag, Typography, theme } from "antd";
import type { CredentialStatus } from "@/api/credentials";
import {
  CREDENTIAL_STATUS_COLOR,
  CREDENTIAL_STATUS_LABEL,
} from "./credentialStatus";

const { Text } = Typography;
const { useToken } = theme;

export interface PlatformConnectionSummary {
  platform: string;
  connected: boolean;
  store_identifier?: string;
  store_name?: string;
  expires_at?: string;
  expires_soon?: boolean;
  status: CredentialStatus;
  region?: string;
  last_refresh_at?: string;
  refresh_status?: string;
  app_configured?: boolean;
}

interface PlatformCardProps {
  platform: PlatformConnectionSummary;
  color: string;
  icon: React.ReactNode;
  name: string;
  onConnect: (platform: PlatformConnectionSummary) => void;
  onDisconnect: (platform: PlatformConnectionSummary) => void;
  actionDisabledReason?: string | null;
  destructiveActionKey?: string | null;
  formatExpiry: (expiresAt?: string) => string | null;
  onViewHistory: (platform: PlatformConnectionSummary) => void;
}

export function PlatformCard({
  platform,
  color,
  icon,
  name,
  onConnect,
  onDisconnect,
  actionDisabledReason,
  destructiveActionKey,
  formatExpiry,
  onViewHistory,
}: PlatformCardProps) {
  const { token } = useToken();

  return (
    <Card
      style={{
        borderRadius: token.borderRadius,
        borderTop: `3px solid ${color}`,
        height: "100%",
        overflow: "hidden",
      }}
      styles={{ body: { padding: 16 } }}
    >
      <Space direction="vertical" size={12} style={{ width: "100%" }}>
        <div
          style={{ display: "flex", justifyContent: "space-between", gap: 12 }}
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
            <div>
              <Text strong style={{ fontSize: 14, display: "block", wordBreak: "break-word" }}>
                {name}
              </Text>
              <Text type="secondary" style={{ fontSize: 12 }}>
                {platform.store_name ||
                  platform.store_identifier ||
                  "Store connection"}
              </Text>
            </div>
          </div>
          <Tag
            icon={
              platform.connected ? <CheckCircleOutlined /> : <CloseCircleOutlined />
            }
            color={CREDENTIAL_STATUS_COLOR[platform.status]}
            style={{ margin: 0 }}
          >
            {CREDENTIAL_STATUS_LABEL[platform.status]}
          </Tag>
        </div>

        <Card size="small" style={{ background: token.colorFillTertiary }}>
          <Space direction="vertical" size={4} style={{ width: "100%" }}>
            <Text type="secondary" style={{ fontSize: 11, letterSpacing: 0.4 }}>
              STORE CONNECTION
            </Text>
            <Text style={{ fontSize: 12 }}>
              {platform.connected ? "Connected and syncing" : "Not connected"}
            </Text>
            {platform.store_identifier && (
              <Text type="secondary" style={{ fontSize: 12 }}>
                Store ID: {platform.store_identifier}
              </Text>
            )}
            {platform.expires_at && (
              <Text
                type={platform.expires_soon ? "danger" : "secondary"}
                style={{ fontSize: 12 }}
              >
                {formatExpiry(platform.expires_at)}
              </Text>
            )}
            {platform.last_refresh_at && (
              <Text type="secondary" style={{ fontSize: 12 }}>
                Last refresh:{" "}
                {new Date(platform.last_refresh_at).toLocaleString()}
              </Text>
            )}
          </Space>
        </Card>

        <div style={{ overflow: "hidden" }}>
          {!platform.connected && !platform.app_configured && (
            <div
              style={{
                padding: "6px 10px",
                marginBottom: 8,
                borderRadius: token.borderRadius,
                background: token.colorWarningBg,
                border: `1px solid ${token.colorWarningBorder}`,
                fontSize: 12,
              }}
            >
              ⚠️ Configure app credentials before connecting
            </div>
          )}
          {platform.connected ? (
            <div style={{ display: "flex", gap: 6, flexWrap: "wrap" }}>
              <Button size="small" icon={<LinkOutlined />} disabled={!!actionDisabledReason} title={actionDisabledReason || undefined} onClick={() => onConnect(platform)}>Re-authorize</Button>
              <Button size="small" onClick={() => onViewHistory(platform)}>History</Button>
              <Button size="small" danger icon={<DisconnectOutlined />} disabled={!!actionDisabledReason || destructiveActionKey === platform.platform} loading={destructiveActionKey === platform.platform} title={actionDisabledReason || undefined} onClick={() => onDisconnect(platform)}>Disconnect</Button>
            </div>
          ) : (
            <Button
              type="primary"
              icon={<LinkOutlined />}
              disabled={!!actionDisabledReason || !platform.app_configured}
              title={actionDisabledReason || (!platform.app_configured ? "Configure app credentials first" : undefined)}
              onClick={() => onConnect(platform)}
              style={{ width: "100%", background: platform.app_configured ? color : undefined, borderColor: platform.app_configured ? color : undefined }}
            >
              Connect {name}
            </Button>
          )}
          {actionDisabledReason && (
            <Text type="secondary" style={{ fontSize: 12 }}>
              {actionDisabledReason}
            </Text>
          )}
        </div>
      </Space>
    </Card>
  );
}
