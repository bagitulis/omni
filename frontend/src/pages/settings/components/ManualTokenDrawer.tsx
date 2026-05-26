import { Button, Drawer, Form, Input, Select } from "antd";
import type { FormInstance } from "antd";
import type { CredentialPlatformSummary } from "@/api/credentials";

export interface ManualTokenFormValues {
  store_identifier: string;
  region?: string;
  access_token: string;
  refresh_token?: string;
  expires_at?: string;
  shop_cipher?: string;
  reason: string;
}

interface ManualTokenDrawerProps {
  open: boolean;
  platform: CredentialPlatformSummary | null;
  platformNames: Record<string, string>;
  form: FormInstance<ManualTokenFormValues>;
  saving: boolean;
  onClose: () => void;
  onSubmit: () => void;
}

export function ManualTokenDrawer({
  open,
  platform,
  platformNames,
  form,
  saving,
  onClose,
  onSubmit,
}: ManualTokenDrawerProps) {
  const title = platform
    ? `Manual Token - ${platformNames[platform.platform] || platform.platform}`
    : "Manual Token";

  return (
    <Drawer open={open} title={title} width={560} onClose={onClose}>
      <Form form={form} layout="vertical">
        <Form.Item
          label="Store Identifier"
          name="store_identifier"
          rules={[{ required: true, message: "Store identifier is required" }]}
        >
          <Input placeholder="123456789" />
        </Form.Item>
        <Form.Item label="Region" name="region">
          <Select options={[{ label: "Indonesia", value: "id" }]} />
        </Form.Item>
        <Form.Item
          label="Access Token"
          name="access_token"
          rules={[{ required: true, message: "Access token is required" }]}
        >
          <Input.Password autoComplete="off" />
        </Form.Item>
        <Form.Item label="Refresh Token" name="refresh_token">
          <Input.Password autoComplete="off" />
        </Form.Item>
        <Form.Item label="Expires At" name="expires_at">
          <Input placeholder="2026-05-26T12:00:00Z" />
        </Form.Item>
        <Form.Item label="Shop Cipher" name="shop_cipher">
          <Input.Password autoComplete="off" />
        </Form.Item>
        <Form.Item
          label="Reason"
          name="reason"
          rules={[{ required: true, message: "Reason is required" }]}
        >
          <Input placeholder="emergency_recovery" />
        </Form.Item>
        <Button type="primary" onClick={onSubmit} loading={saving}>
          Save Manual Token
        </Button>
      </Form>
    </Drawer>
  );
}
