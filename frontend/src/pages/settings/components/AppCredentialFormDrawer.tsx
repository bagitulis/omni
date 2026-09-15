import { useEffect, useState } from "react";
import {
  Alert,
  Button,
  DatePicker,
  Divider,
  Drawer,
  Form,
  Input,
  InputNumber,
  Segmented,
  Select,
  Space,
  Typography,
} from "antd";
import dayjs, { type Dayjs } from "dayjs";
import { ExclamationCircleOutlined, SafetyOutlined } from "@ant-design/icons";
import type { AppCredentialUpsertPayload } from "@/api/credentials";
import {
  LAZADA_REGIONS,
  PLATFORM_DISPLAY_NAMES,
  SHOPEE_REGIONS,
} from "../constants/platformRegions";

const { Text } = Typography;

export type AppCredentialDrawerMode = "setup" | "rotate";

interface AppCredentialFormDrawerProps {
  open: boolean;
  platform: string | null;
  mode: AppCredentialDrawerMode;
  appConfigured: boolean;
  saving: boolean;
  onSubmit: (values: AppCredentialUpsertPayload) => void;
  onClose: () => void;
  // Phase 8 — pre-fill from server-side masked response so the drawer opens
  // with the current toggles/expiries. Secret fields (partner_key,
  // test_partner_key) are NEVER pre-filled — user must re-enter to change
  // them, and empty means "leave existing value untouched".
  initialShopee?: {
    partner_id?: number;
    test_partner_id?: number;
    partner_key_expires_at?: string;
    test_partner_key_expires_at?: string;
    app_status?: "online" | "offline";
    active_partner_env?: "live" | "test";
  };
}

interface FormValues {
  partner_id?: number;
  partner_key?: string;
  app_key?: string;
  app_secret?: string;
  region?: string;
  reason?: string;
  // Phase 8 — Shopee-only extras (typed loosely because DatePicker uses Dayjs).
  test_partner_id?: number;
  test_partner_key?: string;
  partner_key_expires_at?: Dayjs;
  test_partner_key_expires_at?: Dayjs;
  app_status?: "online" | "offline";
  active_partner_env?: "live" | "test";
}

const secretFieldProps = {
  autoComplete: "new-password",
};

export function AppCredentialFormDrawer({
  open,
  platform,
  mode,
  appConfigured,
  saving,
  onSubmit,
  onClose,
  initialShopee,
}: AppCredentialFormDrawerProps) {
  const [form] = Form.useForm<FormValues>();
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (open) {
      form.resetFields();
      setSubmitting(false);
      // Default region to "id" for Shopee/Lazada
      if (platform === "shopee" || platform === "lazada") {
        form.setFieldsValue({ region: "id" });
      }
      // Phase 8 — pre-fill Shopee non-secret fields from initialShopee.
      // Secrets stay blank on purpose (see prop docstring).
      if (platform === "shopee" && initialShopee) {
        form.setFieldsValue({
          partner_id: initialShopee.partner_id,
          test_partner_id: initialShopee.test_partner_id,
          partner_key_expires_at: initialShopee.partner_key_expires_at
            ? dayjs(initialShopee.partner_key_expires_at)
            : undefined,
          test_partner_key_expires_at: initialShopee.test_partner_key_expires_at
            ? dayjs(initialShopee.test_partner_key_expires_at)
            : undefined,
          app_status: initialShopee.app_status || "online",
          active_partner_env: initialShopee.active_partner_env || "live",
        });
      }
    }
  }, [open, platform, form, initialShopee]);

  const platformName = platform ? PLATFORM_DISPLAY_NAMES[platform] || platform : "";
  const isRotate = mode === "rotate";

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      setSubmitting(true);

      let payload: AppCredentialUpsertPayload;
      if (platform === "shopee") {
        // Phase 8 — include sandbox pair + expiry + operational toggles when
        // the user has filled them in. Empty fields are omitted so the
        // backend's "leave existing value" behaviour applies.
        const shopeePayload: AppCredentialUpsertPayload = {
          partner_id: values.partner_id!,
          partner_key: values.partner_key!,
          region: values.region || "id",
          reason: values.reason,
          app_status: values.app_status || "online",
          active_partner_env: values.active_partner_env || "live",
        };
        if (values.test_partner_id && values.test_partner_key) {
          shopeePayload.test_partner_id = values.test_partner_id;
          shopeePayload.test_partner_key = values.test_partner_key;
        }
        if (values.partner_key_expires_at) {
          shopeePayload.partner_key_expires_at =
            values.partner_key_expires_at.toISOString();
        }
        if (values.test_partner_key_expires_at) {
          shopeePayload.test_partner_key_expires_at =
            values.test_partner_key_expires_at.toISOString();
        }
        payload = shopeePayload;
      } else if (platform === "lazada") {
        payload = {
          app_key: values.app_key!,
          app_secret: values.app_secret!,
          region: values.region || "id",
          reason: values.reason,
        };
      } else {
        // TikTok
        payload = {
          app_key: values.app_key!,
          app_secret: values.app_secret!,
          reason: values.reason,
        };
      }
      onSubmit(payload);
    } catch {
      // Validation failed — Ant Design shows errors inline
    }
  };

  const handleClose = () => {
    if (!saving && !submitting) {
      form.resetFields();
      onClose();
    }
  };

  const isLoading = saving || submitting;

  return (
    <Drawer
      title={
        <Space>
          <SafetyOutlined />
          {isRotate
            ? `Rotate ${platformName} Credentials`
            : `Configure ${platformName} App Credentials`}
        </Space>
      }
      open={open}
      onClose={handleClose}
      width={480}
      destroyOnClose
      footer={
        <div style={{ display: "flex", justifyContent: "flex-end", gap: 8 }}>
          <Button onClick={handleClose} disabled={isLoading}>
            Cancel
          </Button>
          <Button
            type="primary"
            danger={isRotate}
            loading={isLoading}
            onClick={handleSubmit}
          >
            {isRotate ? "Rotate Credentials" : "Save Credentials"}
          </Button>
        </div>
      }
    >
      {/* Rotation warning banner (FIX-2) */}
      {isRotate && (
        <Alert
          type="warning"
          showIcon
          icon={<ExclamationCircleOutlined />}
          message="Credential Rotation"
          description={
            <span>
              Rotating app credentials will <strong>invalidate all existing OAuth tokens</strong>{" "}
              for {platformName}. Connected stores will need to re-authorize via OAuth.
            </span>
          }
          style={{ marginBottom: 24 }}
        />
      )}

      {/* Setup info when not yet configured */}
      {!isRotate && !appConfigured && (
        <Alert
          type="info"
          showIcon
          message="Initial Setup"
          description={`Enter the app credentials from your ${platformName} developer console. These credentials are used to initiate OAuth flows for store connections.`}
          style={{ marginBottom: 24 }}
        />
      )}

      <Form form={form} layout="vertical" disabled={isLoading}>
        {/* Shopee-specific fields */}
        {platform === "shopee" && (
          <>
            <Form.Item
              name="partner_id"
              label="Partner ID"
              rules={[
                { required: true, message: "Partner ID is required" },
                { type: "number", min: 1, message: "Must be a positive number" },
              ]}
            >
              <InputNumber
                style={{ width: "100%" }}
                placeholder="e.g. 123456"
                min={1}
              />
            </Form.Item>
            <Form.Item
              name="partner_key"
              label="Partner Key"
              rules={[
                { required: true, message: "Partner Key is required" },
                { min: 8, message: "Must be at least 8 characters" },
                {
                  pattern: /^\S+$/,
                  message: "Must not contain whitespace",
                },
              ]}
            >
              <Input.Password
                placeholder="Enter partner key"
                {...secretFieldProps}
              />
            </Form.Item>
            <Form.Item
              name="region"
              label="Region"
              rules={[{ required: true, message: "Region is required" }]}
            >
              <Select
                placeholder="Select region"
                options={SHOPEE_REGIONS.map((r) => ({
                  value: r.value,
                  label: r.label,
                }))}
              />
            </Form.Item>
            <Form.Item
              name="partner_key_expires_at"
              label="Live Partner Key Expire Time"
              tooltip="From the Shopee developer console — used to render a rotation warning banner 7 days before expiry."
              rules={[
                { required: true, message: "Live partner-key expiry is required" },
              ]}
            >
              <DatePicker
                showTime
                style={{ width: "100%" }}
                placeholder="Select expire date & time"
              />
            </Form.Item>

            <Divider orientation="left" plain>
              Sandbox pair (optional — for dry-run probing)
            </Divider>
            <Form.Item
              name="test_partner_id"
              label="Test Partner ID"
              rules={[{ type: "number", min: 1, message: "Must be a positive number" }]}
            >
              <InputNumber style={{ width: "100%" }} placeholder="e.g. 1187586" min={1} />
            </Form.Item>
            <Form.Item
              name="test_partner_key"
              label="Test Partner Key"
              rules={[
                { min: 8, message: "Must be at least 8 characters" },
                { pattern: /^\S+$/, message: "Must not contain whitespace" },
              ]}
            >
              <Input.Password placeholder="Enter test partner key (leave blank to keep)" {...secretFieldProps} />
            </Form.Item>
            <Form.Item name="test_partner_key_expires_at" label="Test Partner Key Expire Time">
              <DatePicker
                showTime
                style={{ width: "100%" }}
                placeholder="Optional — expiry for sandbox pair"
              />
            </Form.Item>

            <Divider orientation="left" plain>
              Runtime & status
            </Divider>
            <Form.Item
              name="active_partner_env"
              label="Active Environment"
              tooltip="Determines which pair signs outgoing Shopee API calls. Falls back to Live if Test pair is empty."
              initialValue="live"
            >
              <Segmented
                options={[
                  { label: "Live", value: "live" },
                  { label: "Test (sandbox)", value: "test" },
                ]}
              />
            </Form.Item>
            <Form.Item
              name="app_status"
              label="App Status"
              tooltip="When Offline, sync workers will skip this platform until re-enabled."
              initialValue="online"
            >
              <Segmented
                options={[
                  { label: "Online", value: "online" },
                  { label: "Offline", value: "offline" },
                ]}
              />
            </Form.Item>
          </>
        )}

        {/* Lazada-specific fields */}
        {platform === "lazada" && (
          <>
            <Form.Item
              name="app_key"
              label="App Key (Client ID)"
              rules={[
                { required: true, message: "App Key is required" },
                { min: 8, message: "Must be at least 8 characters" },
                {
                  pattern: /^\S+$/,
                  message: "Must not contain whitespace",
                },
              ]}
            >
              <Input placeholder="Enter app key" {...secretFieldProps} />
            </Form.Item>
            <Form.Item
              name="app_secret"
              label="App Secret"
              rules={[
                { required: true, message: "App Secret is required" },
                { min: 8, message: "Must be at least 8 characters" },
                {
                  pattern: /^\S+$/,
                  message: "Must not contain whitespace",
                },
              ]}
            >
              <Input.Password
                placeholder="Enter app secret"
                {...secretFieldProps}
              />
            </Form.Item>
            <Form.Item
              name="region"
              label="Region"
              rules={[{ required: true, message: "Region is required" }]}
            >
              <Select
                placeholder="Select region"
                options={LAZADA_REGIONS.map((r) => ({
                  value: r.value,
                  label: r.label,
                }))}
              />
            </Form.Item>
          </>
        )}

        {/* TikTok-specific fields */}
        {platform === "tiktok" && (
          <>
            <Form.Item
              name="app_key"
              label="App Key"
              rules={[
                { required: true, message: "App Key is required" },
                { min: 8, message: "Must be at least 8 characters" },
                {
                  pattern: /^\S+$/,
                  message: "Must not contain whitespace",
                },
              ]}
            >
              <Input placeholder="Enter app key" {...secretFieldProps} />
            </Form.Item>
            <Form.Item
              name="app_secret"
              label="App Secret"
              rules={[
                { required: true, message: "App Secret is required" },
                { min: 8, message: "Must be at least 8 characters" },
                {
                  pattern: /^\S+$/,
                  message: "Must not contain whitespace",
                },
              ]}
            >
              <Input.Password
                placeholder="Enter app secret"
                {...secretFieldProps}
              />
            </Form.Item>
          </>
        )}

        {/* Reason field for rotate mode */}
        {isRotate && (
          <Form.Item
            name="reason"
            label="Reason for Rotation"
            rules={[
              { required: true, message: "Reason is required for credential rotation" },
              {
                min: 10,
                message: "Please provide a meaningful reason (at least 10 characters)",
              },
            ]}
          >
            <Input.TextArea
              placeholder="e.g. Compromised credentials, scheduled quarterly rotation..."
              rows={3}
              maxLength={200}
              showCount
            />
          </Form.Item>
        )}

        {/* Security note */}
        <Text type="secondary" style={{ fontSize: 11, display: "block", marginTop: 8 }}>
          🔒 Credentials are encrypted at rest with AES-256 and never exposed in API
          responses. Only masked values are shown after saving.
        </Text>
      </Form>
    </Drawer>
  );
}
