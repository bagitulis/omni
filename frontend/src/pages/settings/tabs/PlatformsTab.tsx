import {
  getCredentialAudit,
  getCredentialPlatforms,
  saveManualToken,
} from "@/api/credentials";
import type {
  CredentialAuditEvent,
  CredentialPlatformSummary,
} from "@/api/credentials";
import { message } from "@/components/AntStaticApi";
import {
  LoadingOutlined,
  ShopOutlined,
  ShoppingOutlined,
  VideoCameraOutlined,
} from "@ant-design/icons";
import { Card, Col, Form, Row, Spin, Typography, theme } from "antd";
import { useCallback, useEffect, useMemo, useState } from "react";
import { CredentialAppCredentialsSection } from "../components/CredentialAppCredentialsSection";
import { CredentialHistoryDrawer } from "../components/CredentialHistoryDrawer";
import {
  ManualTokenDrawer,
} from "../components/ManualTokenDrawer";
import type { ManualTokenFormValues } from "../components/ManualTokenDrawer";
import { PlatformCard } from "../components/PlatformCard";
import type { PlatformConnectionSummary } from "../components/PlatformCard";

const { Text, Title } = Typography;

const PLATFORM_NAMES: Record<string, string> = {
  shopee: "Shopee",
  tiktok: "TikTok Shop",
  lazada: "Lazada",
};

export default function PlatformsTab() {
  const { token } = theme.useToken();
  const [platforms, setPlatforms] = useState<CredentialPlatformSummary[]>([]);
  const [loading, setLoading] = useState(true);
  const [historyOpen, setHistoryOpen] = useState(false);
  const [historyPlatform, setHistoryPlatform] =
    useState<CredentialPlatformSummary | null>(null);
  const [historyEvents, setHistoryEvents] = useState<CredentialAuditEvent[]>([]);
  const [historyLoading, setHistoryLoading] = useState(false);
  const [manualOpen, setManualOpen] = useState(false);
  const [manualPlatform, setManualPlatform] =
    useState<CredentialPlatformSummary | null>(null);
  const [manualSaving, setManualSaving] = useState(false);
  const [manualForm] = Form.useForm<ManualTokenFormValues>();

  const fetchPlatformStatus = useCallback(async () => {
    try {
      setLoading(true);
      setPlatforms(await getCredentialPlatforms());
    } catch (err) {
      const msg =
        err instanceof Error ? err.message : "Failed to load platform status";
      message.error(msg);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void fetchPlatformStatus();
  }, [fetchPlatformStatus]);

  const platformCards = useMemo(
    () =>
      platforms.map((platform) => ({
        platform,
        color:
          platform.platform === "shopee"
            ? token.colorPrimary
            : platform.platform === "tiktok"
              ? token.colorText
              : token.colorInfo,
        icon:
          platform.platform === "shopee" ? (
            <ShopOutlined
              style={{ fontSize: 24, color: token.colorPrimary }}
            />
          ) : platform.platform === "tiktok" ? (
            <VideoCameraOutlined
              style={{ fontSize: 24, color: token.colorText }}
            />
          ) : (
            <ShoppingOutlined
              style={{ fontSize: 24, color: token.colorInfo }}
            />
          ),
        name: PLATFORM_NAMES[platform.platform] || platform.platform,
      })),
    [platforms, token.colorPrimary, token.colorInfo, token.colorText],
  );

  const handleConnect = (platform: PlatformConnectionSummary) => {
    const platformName = PLATFORM_NAMES[platform.platform] || platform.platform;
    message.info(`Opening ${platformName} authorization...`);
    window.open(
      `/api/credentials/platforms/${platform.platform}/connections/oauth/initiate`,
      "_blank",
    );
  };

  const handleDisconnect = async (platform: PlatformConnectionSummary) => {
    const platformName = PLATFORM_NAMES[platform.platform] || platform.platform;
    message.info(
      `Disconnect for ${platformName} is handled by backend contract`,
    );
  };

  const formatExpiry = (expiresAt?: string) => {
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

  const openHistory = useCallback(
    async (platform: CredentialPlatformSummary) => {
      try {
        setHistoryPlatform(platform);
        setHistoryOpen(true);
        setHistoryLoading(true);
        setHistoryEvents(await getCredentialAudit(platform.platform));
      } catch (err) {
        const msg =
          err instanceof Error
            ? err.message
            : "Failed to load credential history";
        message.error(msg);
      } finally {
        setHistoryLoading(false);
      }
    },
    [],
  );

  const handleManualTokenOpen = (platform: CredentialPlatformSummary) => {
    setManualPlatform(platform);
    manualForm.setFieldsValue({
      store_identifier: "",
      region: platform.region || "id",
      reason: "emergency_recovery",
    });
    setManualOpen(true);
  };

  const handleManualTokenSubmit = async () => {
    if (!manualPlatform) return;
    try {
      const values = await manualForm.validateFields();
      setManualSaving(true);
      await saveManualToken(manualPlatform.platform, values);
      message.success("Manual token saved");
      setManualOpen(false);
      await fetchPlatformStatus();
    } catch (err) {
      if (err instanceof Error && err.message.includes("required")) {
        return;
      }
      const msg =
        err instanceof Error ? err.message : "Failed to save manual token";
      message.error(msg);
    } finally {
      setManualSaving(false);
    }
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
          Credential Management
        </Title>
        <Text type="secondary" style={{ fontSize: 12 }}>
          Separate store connections from app credentials and review masked credential history
        </Text>
      </div>

      <Card title="Store Connections" size="small" style={{ marginBottom: 24 }}>
        <Row gutter={[16, 16]}>
          {platformCards.map(({ platform, color, icon, name }) => (
            <Col xs={24} md={8} key={platform.platform}>
              <PlatformCard
                platform={{
                  platform: platform.platform,
                  connected: platform.status === "connected",
                  store_identifier: platform.stores[0]?.store_identifier,
                  store_name: platform.stores[0]?.store_name,
                  expires_at: platform.stores[0]?.expires_at,
                  expires_soon: platform.status === "expired",
                  status: platform.status,
                  region: platform.region,
                  last_refresh_at: platform.stores[0]?.last_refresh_at,
                  refresh_status: platform.stores[0]?.refresh_status,
                }}
                color={color}
                icon={icon}
                name={name}
                onConnect={handleConnect}
                onDisconnect={handleDisconnect}
                formatExpiry={formatExpiry}
                onViewHistory={(selectedPlatform) =>
                  void openHistory({
                    platform: selectedPlatform.platform,
                    region: platform.region,
                    status: platform.status,
                    app_configured: platform.app_configured,
                    secret_mask: platform.secret_mask,
                    app_secret_mask: platform.app_secret_mask,
                    stores: platform.stores,
                    app_config: platform.app_config,
                    audit_summary: platform.audit_summary,
                  })
                }
              />
            </Col>
          ))}
        </Row>
      </Card>

      <CredentialAppCredentialsSection
        platforms={platforms}
        platformNames={PLATFORM_NAMES}
        onManualToken={handleManualTokenOpen}
        onViewHistory={(platform) => void openHistory(platform)}
      />

      <CredentialHistoryDrawer
        open={historyOpen}
        platform={historyPlatform}
        events={historyEvents}
        loading={historyLoading}
        platformNames={PLATFORM_NAMES}
        onClose={() => setHistoryOpen(false)}
      />

      <ManualTokenDrawer
        open={manualOpen}
        platform={manualPlatform}
        platformNames={PLATFORM_NAMES}
        form={manualForm}
        saving={manualSaving}
        onClose={() => setManualOpen(false)}
        onSubmit={() => void handleManualTokenSubmit()}
      />
    </div>
  );
}
