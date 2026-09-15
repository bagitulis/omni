import {
  CheckCircleOutlined,
  CloseCircleOutlined,
  DisconnectOutlined,
  LinkOutlined,
} from "@ant-design/icons";
import { Button, Card, Space, Tag, Typography, theme } from "antd";
import type { GlobalToken } from "antd/es/theme/interface";
import type { CredentialStatus } from "@/api/credentials";
import {
  CREDENTIAL_STATUS_COLOR,
  CREDENTIAL_STATUS_LABEL,
} from "./credentialStatus";
import { isPartnerKeyExpiring } from "./partnerKeyExpiry";

const { Text } = Typography;
const { useToken } = theme;

function renderPartnerKeyExpiryBanner(
  platform: PlatformConnectionSummary,
  token: GlobalToken,
) {
  const { expired, expiring, days_left } = isPartnerKeyExpiring(
    platform.partner_key_expires_at,
  );
  if (!expiring) return null;
  const isRed = expired;
  return (
    <div
      style={{
        padding: "6px 10px",
        marginBottom: 8,
        borderRadius: token.borderRadius,
        background: isRed ? token.colorErrorBg : token.colorWarningBg,
        border: `1px solid ${
          isRed ? token.colorErrorBorder : token.colorWarningBorder
        }`,
        fontSize: 12,
      }}
    >
      {isRed ? "🔴" : "⚠️"}{" "}
      <strong>Partner Key {isRed ? "expired" : `expires in ${days_left}d`}</strong>{" "}
      — rotate soon via the Configure drawer.
    </div>
  );
}

// Phase 10.2 — Surface backend `EffectiveStatus` results ("expired" hard-dead,
// "refresh_required" auto-recoverable) as a prominent banner so sellers know
// they need to re-authorize instead of just seeing a small status tag.
function renderTokenExpiredBanner(
  platform: PlatformConnectionSummary,
  token: GlobalToken,
) {
  const isHardExpired = platform.status === "expired";
  const isRefreshRequired = platform.status === "refresh_required";
  if (!isHardExpired && !isRefreshRequired) return null;
  return (
    <div
      style={{
        padding: "6px 10px",
        marginBottom: 8,
        borderRadius: token.borderRadius,
        background: isHardExpired ? token.colorErrorBg : token.colorWarningBg,
        border: `1px solid ${
          isHardExpired ? token.colorErrorBorder : token.colorWarningBorder
        }`,
        fontSize: 12,
      }}
    >
      {isHardExpired ? "🔴" : "⚠️"}{" "}
      <strong>Access token expired</strong> —{" "}
      {isHardExpired
        ? "click Re-authorize below to reconnect."
        : "auto-refresh will run on next sync, or click Re-authorize to refresh now."}
    </div>
  );
}

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
  // Phase 8 — Shopee partner-key expiry + operational toggles surfaced to the card.
  partner_key_expires_at?: string;
  app_status?: "online" | "offline";
  active_partner_env?: "live" | "test";
  test_configured?: boolean;
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
          {renderTokenExpiredBanner(platform, token)}
          {renderPartnerKeyExpiryBanner(platform, token)}
          {platform.app_status === "offline" && (
            <div
              style={{
                padding: "6px 10px",
                marginBottom: 8,
                borderRadius: token.borderRadius,
                background: token.colorFillTertiary,
                border: `1px solid ${token.colorBorder}`,
                fontSize: 12,
              }}
            >
              ⏸️ App marked <strong>Offline</strong> — sync workers will skip
              this platform.
            </div>
          )}
          {platform.active_partner_env === "test" && (
            <div
              style={{
                padding: "6px 10px",
                marginBottom: 8,
                borderRadius: token.borderRadius,
                background: token.colorInfoBg,
                border: `1px solid ${token.colorInfoBorder}`,
                fontSize: 12,
              }}
            >
              🧪 Runtime pinned to <strong>Test</strong> partner pair.
            </div>
          )}
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
              {/* Phase 10.2 — "expired"/"refresh_required" carry an existing token
                  pair; the action is a re-authorization, not a first-time connect. */}
              {platform.status === "expired" || platform.status === "refresh_required"
                ? `Re-authorize ${name}`
                : `Connect ${name}`}
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
