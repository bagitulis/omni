import { Button, Drawer, InputNumber, message, Select, Space, Typography } from "antd";
import { useEffect, useState } from "react";
import { notificationApi, type NotificationSettings } from "@/api/notifications";
import { logger } from "@/lib/logger";

const { Text, Title } = Typography;

interface NotificationSettingsDrawerProps {
  open: boolean;
  onClose: () => void;
}

const RETENTION_OPTIONS = [
  { label: "7 days", value: 7 },
  { label: "30 days", value: 30 },
  { label: "90 days", value: 90 },
  { label: "Manual", value: 0 },
];

const SEVERITY_OPTIONS = [
  { label: "Info (all)", value: 10 },
  { label: "Low", value: 20 },
  { label: "Medium", value: 30 },
  { label: "High", value: 40 },
  { label: "Critical only", value: 50 },
];

/**
 * Right-side settings drawer for notifications. Two knobs today:
 *  - retention_days: cron auto-purge horizon (0 = manual).
 *  - min_severity_toast: FE suppresses toast for lower severities.
 * More knobs (channel prefs) land in a follow-up.
 */
export function NotificationSettingsDrawer(props: NotificationSettingsDrawerProps) {
  const [settings, setSettings] = useState<NotificationSettings>({ retention_days: 30, min_severity_toast: 20 });
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    if (!props.open) return;
    setLoading(true);
    notificationApi
      .getSettings()
      .then((res) => {
        if (res.success && res.data) setSettings({ ...settings, ...res.data });
      })
      .catch((err) => logger.warn("Failed to load notification settings", { err }))
      .finally(() => setLoading(false));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [props.open]);

  const save = async () => {
    setSaving(true);
    try {
      const res = await notificationApi.updateSettings(settings);
      if (res.success) {
        message.success("Notification settings saved");
        props.onClose();
      }
    } catch (err) {
      logger.warn("Failed to save notification settings", { err });
      message.error("Failed to save settings");
    } finally {
      setSaving(false);
    }
  };

  return (
    <Drawer
      title="Notification Settings"
      placement="right"
      onClose={props.onClose}
      open={props.open}
      width={360}
      loading={loading}
      extra={
        <Button type="primary" onClick={save} loading={saving}>
          Save
        </Button>
      }
    >
      <Space direction="vertical" size="large" style={{ width: "100%" }}>
        <div>
          <Title level={5}>Retention</Title>
          <Text type="secondary">Cron auto-deletes notifications older than this.</Text>
          <div style={{ marginTop: 8 }}>
            <Select
              value={settings.retention_days}
              onChange={(v) => setSettings({ ...settings, retention_days: v })}
              options={RETENTION_OPTIONS}
              style={{ width: "100%" }}
            />
          </div>
        </div>
        <div>
          <Title level={5}>Toast threshold</Title>
          <Text type="secondary">Only notifications at or above this severity trigger a toast.</Text>
          <div style={{ marginTop: 8 }}>
            <Select
              value={settings.min_severity_toast ?? 20}
              onChange={(v) => setSettings({ ...settings, min_severity_toast: v })}
              options={SEVERITY_OPTIONS}
              style={{ width: "100%" }}
            />
          </div>
        </div>
        <div>
          <Title level={5}>Manual retention override</Title>
          <Text type="secondary">Days (0 = disabled).</Text>
          <div style={{ marginTop: 8 }}>
            <InputNumber
              min={0}
              max={365}
              value={settings.retention_days}
              onChange={(v) => setSettings({ ...settings, retention_days: v ?? 0 })}
              style={{ width: "100%" }}
            />
          </div>
        </div>
      </Space>
    </Drawer>
  );
}
