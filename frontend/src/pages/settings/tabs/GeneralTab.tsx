import {
  Card,
  Form,
  Select,
  Switch,
  Button,
  Typography,
  Divider,
  theme,
  Spin,
} from "antd";
import { useEffect, useState, useMemo } from "react";
import { saveGeneralSettings } from "@/api/settings";
import apiClient from "@/api/client";
import { message } from "@/components/AntStaticApi";

const { Text } = Typography;
const { useToken } = theme;

interface GeneralSettings {
  language: string;
  timezone: string;
  notifications_email: boolean;
  notifications_browser: boolean;
  auto_sync: boolean;
  sync_interval: string;
}

export default function GeneralTab() {
  const { token } = useToken();
  const [form] = Form.useForm<GeneralSettings>();
  const [loading, setLoading] = useState(true);

  const defaultSettings = useMemo<GeneralSettings>(
    () => ({
      language: "en",
      timezone: "Asia/Jakarta",
      notifications_email: true,
      notifications_browser: true,
      auto_sync: true,
      sync_interval: "30",
    }),
    [],
  );

  // Load settings on mount: try API first, then localStorage, then defaults
  useEffect(() => {
    const loadSettings = async () => {
      setLoading(true);
      try {
        // Try API first
        const response =
          await apiClient.get<GeneralSettings>("/settings/general");
        if (response.success && response.data) {
          form.setFieldsValue(response.data);
          setLoading(false);
          return;
        }
      } catch (err) { console.warn("Operation failed:", err);
        // API failed, try localStorage
      }

      // Try localStorage fallback
      try {
        const localData = localStorage.getItem("omni_general_settings");
        if (localData) {
          const parsed = JSON.parse(localData);
          form.setFieldsValue(parsed);
          setLoading(false);
          return;
        }
      } catch (err) { console.warn("Operation failed:", err);
        // localStorage parse failed, use defaults
      }

      // Use default values
      form.setFieldsValue(defaultSettings);
      setLoading(false);
    };

    loadSettings();
  }, [form, defaultSettings]);

  const handleSave = async (values: GeneralSettings) => {
    try {
      await saveGeneralSettings(values);
      message.success("Settings saved successfully");
    } catch (err) { console.warn("Operation failed:", err);
      message.warning("Settings saved locally (server unavailable)");
    }

    // Always save to localStorage as backup
    try {
      localStorage.setItem("omni_general_settings", JSON.stringify(values));
    } catch (err) { console.warn("Operation failed:", err);
      message.error("Failed to save settings locally");
    }
  };

  return (
    <div>
      <Spin spinning={loading}>
        <Form
          form={form}
          layout="vertical"
          initialValues={defaultSettings}
          onFinish={handleSave}
          style={{ maxWidth: 600 }}
        >
          <Card
            title="Language & Region"
            size="small"
            style={{ marginBottom: 16, borderRadius: token.borderRadius }}
          >
            <Form.Item
              label="Language"
              name="language"
              style={{ marginBottom: 16 }}
            >
              <Select
                options={[
                  { value: "en", label: "English" },
                  { value: "id", label: "Bahasa Indonesia" },
                  { value: "zh", label: "中文" },
                ]}
              />
            </Form.Item>
            <Form.Item
              label="Timezone"
              name="timezone"
              style={{ marginBottom: 0 }}
            >
              <Select
                options={[
                  { value: "Asia/Jakarta", label: "Asia/Jakarta (WIB)" },
                  { value: "Asia/Singapore", label: "Asia/Singapore (SGT)" },
                  {
                    value: "Asia/Kuala_Lumpur",
                    label: "Asia/Kuala Lumpur (MYT)",
                  },
                  { value: "UTC", label: "UTC" },
                ]}
              />
            </Form.Item>
          </Card>

          <Card
            title="Notifications"
            size="small"
            style={{ marginBottom: 16, borderRadius: token.borderRadius }}
          >
            <Form.Item
              name="notifications_email"
              valuePropName="checked"
              style={{ marginBottom: 12 }}
            >
              <div
                style={{
                  display: "flex",
                  justifyContent: "space-between",
                  alignItems: "center",
                }}
              >
                <div>
                  <Text strong style={{ fontSize: 12 }}>
                    Email Notifications
                  </Text>
                  <br />
                  <Text type="secondary" style={{ fontSize: 12 }}>
                    Receive email notifications (not yet available in this remediation)
                  </Text>
                </div>
                <Switch defaultChecked disabled />
              </div>
            </Form.Item>
            <Divider style={{ margin: "12px 0" }} />
            <Form.Item
              name="notifications_browser"
              valuePropName="checked"
              style={{ marginBottom: 0 }}
            >
              <div
                style={{
                  display: "flex",
                  justifyContent: "space-between",
                  alignItems: "center",
                }}
              >
                <div>
                  <Text strong style={{ fontSize: 12 }}>
                    Browser Notifications
                  </Text>
                  <br />
                  <Text type="secondary" style={{ fontSize: 12 }}>
                    Receive browser notifications (not yet available in this remediation)
                  </Text>
                </div>
                <Switch defaultChecked disabled />
              </div>
            </Form.Item>
          </Card>

          <Card
            title="Data Sync"
            size="small"
            style={{ marginBottom: 16, borderRadius: token.borderRadius }}
          >
            <Form.Item
              name="auto_sync"
              valuePropName="checked"
              style={{ marginBottom: 12 }}
            >
              <div
                style={{
                  display: "flex",
                  justifyContent: "space-between",
                  alignItems: "center",
                }}
              >
                <div>
                  <Text strong style={{ fontSize: 12 }}>
                    Auto Sync
                  </Text>
                  <br />
                  <Text type="secondary" style={{ fontSize: 12 }}>
                    Automatically sync data with platforms
                  </Text>
                </div>
                <Switch defaultChecked />
              </div>
            </Form.Item>
            <Divider style={{ margin: "12px 0" }} />
            <Form.Item
              label="Sync Interval"
              name="sync_interval"
              style={{ marginBottom: 0 }}
            >
              <Select
                options={[
                  { value: "15", label: "Every 15 minutes" },
                  { value: "30", label: "Every 30 minutes" },
                  { value: "60", label: "Every hour" },
                  { value: "manual", label: "Manual only" },
                ]}
              />
            </Form.Item>
          </Card>

          <Form.Item style={{ marginBottom: 0 }}>
            <Button type="primary" htmlType="submit">
              Save Settings
            </Button>
          </Form.Item>
        </Form>
      </Spin>
    </div>
  );
}
