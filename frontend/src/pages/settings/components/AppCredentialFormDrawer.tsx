import { useEffect, useState } from "react";
import {
  Alert,
  Button,
  Drawer,
  Form,
  Input,
  InputNumber,
  Select,
  Space,
  Typography,
} from "antd";
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
}

interface FormValues {
  partner_id?: number;
  partner_key?: string;
  app_key?: string;
  app_secret?: string;
  region?: string;
  reason?: string;
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
    }
  }, [open, platform, form]);

  const platformName = platform ? PLATFORM_DISPLAY_NAMES[platform] || platform : "";
  const isRotate = mode === "rotate";

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      setSubmitting(true);

      let payload: AppCredentialUpsertPayload;
      if (platform === "shopee") {
        payload = {
          partner_id: values.partner_id!,
          partner_key: values.partner_key!,
          region: values.region || "id",
          reason: values.reason,
        };
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
