import {
  disconnectCredentialStore,
  getCredentialAudit,
  getCredentialPlatforms,
  initiateOAuth,
  saveManualToken,
} from "@/api/credentials";
import type {
  CredentialAuditEvent,
  CredentialPlatformSummary,
} from "@/api/credentials";
import { message } from "@/components/AntStaticApi";
import { ShopOutlined, ShoppingOutlined, VideoCameraOutlined } from "@ant-design/icons";
import { Form, Spin, theme } from "antd";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useAuthStore } from "@/stores/authStore";
import { CredentialAppCredentialsSection } from "../components/CredentialAppCredentialsSection";
import { CredentialHistoryDrawer } from "../components/CredentialHistoryDrawer";
import { PlatformsTabHeader } from "../components/PlatformsTabHeader";
import { PlatformStoreConnectionsSection } from "../components/PlatformStoreConnectionsSection";
import { ManualTokenDrawer } from "../components/ManualTokenDrawer";
import type { ManualTokenFormValues } from "../components/ManualTokenDrawer";
import type { PlatformConnectionSummary } from "../components/PlatformCard";
import {
  getPrivilegedActionReason,
  getStoreActionReason,
  PLATFORM_NAMES,
  sanitizeStatusText,
} from "./platformsTabUtils";

export default function PlatformsTab() {
  const { token } = theme.useToken();
  const tenantId = useAuthStore((state) => state.tenantId);
  const role = useAuthStore((state) => state.user?.role);
  const [loadedTenantId, setLoadedTenantId] = useState<string | null>(null);
  const [platforms, setPlatforms] = useState<CredentialPlatformSummary[]>([]);
  const [loading, setLoading] = useState(true);
  const [historyOpen, setHistoryOpen] = useState(false);
  const [historyPlatform, setHistoryPlatform] = useState<CredentialPlatformSummary | null>(null);
  const [historyEvents, setHistoryEvents] = useState<CredentialAuditEvent[]>([]);
  const [historyLoading, setHistoryLoading] = useState(false);
  const [manualOpen, setManualOpen] = useState(false);
  const [manualPlatform, setManualPlatform] = useState<CredentialPlatformSummary | null>(null);
  const [manualSaving, setManualSaving] = useState(false);
  const [destructiveActionKey, setDestructiveActionKey] = useState<string | null>(null);
  const [actionStatus, setActionStatus] = useState<string | null>(null);
  const requestSequenceRef = useRef(0);
  const [manualForm] = Form.useForm<ManualTokenFormValues>();

  const privilegedActionReason = getPrivilegedActionReason(role);
  const storeActionReason = getStoreActionReason(tenantId, role);

  const fetchPlatformStatus = useCallback(async (requestedTenantId: string) => {
    const requestId = requestSequenceRef.current + 1;
    requestSequenceRef.current = requestId;
    setPlatforms([]);
    setLoadedTenantId(null);
    setActionStatus(null);
    try {
      setLoading(true);
      const nextPlatforms = await getCredentialPlatforms({ tenant_id: requestedTenantId });
      if (requestSequenceRef.current !== requestId) return;
      setPlatforms(nextPlatforms);
      setLoadedTenantId(requestedTenantId);
    } catch (err) {
      if (requestSequenceRef.current !== requestId) return;
      const msg = sanitizeStatusText(
        err instanceof Error ? err.message : null,
        "credential_status_load_failed",
      );
      message.error(msg);
    } finally {
      if (requestSequenceRef.current === requestId) {
        setLoading(false);
      }
    }
  }, []);

  useEffect(() => {
    if (!tenantId) {
      setPlatforms([]);
      setLoadedTenantId(null);
      setLoading(false);
      return;
    }
    void fetchPlatformStatus(tenantId);
  }, [fetchPlatformStatus, tenantId]);

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
<ShopOutlined style={{ fontSize: 24, color: token.colorPrimary }} />
          ) : platform.platform === "tiktok" ? (
<VideoCameraOutlined style={{ fontSize: 24, color: token.colorText }} />
          ) : (
<ShoppingOutlined style={{ fontSize: 24, color: token.colorInfo }} />
          ),
        name: PLATFORM_NAMES[platform.platform] || platform.platform,
      })),
    [platforms, token.colorPrimary, token.colorInfo, token.colorText],
  );

  const handleConnect = async (platform: PlatformConnectionSummary) => {
    const currentTenantId = useAuthStore.getState().tenantId;
    const currentRole = useAuthStore.getState().user?.role;
    const reason = getStoreActionReason(currentTenantId, currentRole);
    if (reason || !currentTenantId) {
      message.error(reason || 'Tenant context required before managing store connections.');
      return;
    }
    try {
      const result = await initiateOAuth(
        platform.platform,
        { intent: 'connect', redirect_path: window.location.pathname },
        { tenant_id: currentTenantId },
      );
      window.open(result.auth_url, '_blank');
    } catch (err: any) {
      message.error(err?.message || 'Failed to initiate OAuth');
    }
  };

  const handleDisconnect = async (platform: PlatformConnectionSummary) => {
    const currentState = useAuthStore.getState();
    const currentTenantId = currentState.tenantId;
    const currentRole = currentState.user?.role;
    const reason = getStoreActionReason(currentTenantId, currentRole);
    if (reason || !currentTenantId) {
      message.error(reason || "Tenant context required before managing store connections.");
      return;
    }
    if (!platform.store_identifier || destructiveActionKey) return;
    setDestructiveActionKey(platform.platform);
    try {
      const result = await disconnectCredentialStore(
        platform.platform,
        platform.store_identifier,
        { tenant_id: currentTenantId },
      );
      const statusText = sanitizeStatusText(
        result.code || result.remote_revoke_status || result.status,
        "disconnect_accepted",
      );
      setActionStatus(`Disconnect result: ${statusText}`);
      await fetchPlatformStatus(currentTenantId);
    } catch (err) {
      const msg = sanitizeStatusText(
        err instanceof Error ? err.message : null,
        "disconnect_failed",
      );
      setActionStatus(`Disconnect result: ${msg}`);
      message.error(msg);
    } finally {
      setDestructiveActionKey(null);
    }
  };

  const openHistory = useCallback(
    async (platform: CredentialPlatformSummary) => {
      try {
        setHistoryPlatform(platform);
        setHistoryOpen(true);
        setHistoryLoading(true);
        const currentTenantId = useAuthStore.getState().tenantId;
        if (!currentTenantId) {
          message.error("Tenant context required before loading credential history.");
          return;
        }
        setHistoryEvents(await getCredentialAudit(platform.platform, { tenant_id: currentTenantId }));
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
    const currentRole = useAuthStore.getState().user?.role;
    const reason = getPrivilegedActionReason(currentRole);
    if (reason) {
      message.error(reason);
      return;
    }
    setManualPlatform(platform);
    manualForm.setFieldsValue({
      store_identifier: "",
      region: platform.region || "id",
      reason: "emergency_recovery",
    });
    setManualOpen(true);
  };

  const handleManualTokenSubmit = async () => {
    if (!manualPlatform || manualSaving) return;
    try {
      const currentState = useAuthStore.getState();
      const currentTenantId = currentState.tenantId;
      const currentRole = currentState.user?.role;
      const reason = getPrivilegedActionReason(currentRole) || getStoreActionReason(currentTenantId, currentRole);
      if (reason || !currentTenantId) {
        message.error(reason || "Tenant context required before managing store connections.");
        return;
      }
      const values = await manualForm.validateFields();
      setManualSaving(true);
      await saveManualToken(manualPlatform.platform, values, { tenant_id: currentTenantId });
      setActionStatus("Manual token result: saved");
      message.success("Manual token saved");
      setManualOpen(false);
      await fetchPlatformStatus(currentTenantId);
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

  const tenantBanner = tenantId
    ? `Credential status is scoped to tenant ${tenantId}.`
    : "Credential status is unavailable until a tenant is selected.";
  const isTenantDataReady = !!tenantId && loadedTenantId === tenantId;

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
        <Spin size="large" />
      </div>
    );
  }

  return (
    <div>
      <PlatformsTabHeader
        tenantBanner={tenantBanner}
        isTenantDataReady={isTenantDataReady}
        actionStatus={actionStatus}
      />

      <PlatformStoreConnectionsSection
        platformCards={platformCards}
        storeActionReason={storeActionReason}
        destructiveActionKey={destructiveActionKey}
        onConnect={handleConnect}
        onDisconnect={handleDisconnect}
        onViewHistory={(platform) => void openHistory(platform)}
      />

      <CredentialAppCredentialsSection
        platforms={platforms}
        platformNames={PLATFORM_NAMES}
        privilegedActionReason={privilegedActionReason}
        manualSavingPlatform={manualSaving && manualPlatform ? manualPlatform.platform : null}
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
