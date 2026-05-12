import { useState } from "react";
import { App, Form, Input, Modal, Typography } from "antd";
import { developerApi } from "@/api/developer";

const { Text } = Typography;

interface CreateTenantModalProps {
  open: boolean;
  onClose: () => void;
  onSuccess: () => void;
}

const TENANT_NAME_REGEX = /^[a-z][a-z0-9_]{2,49}$/;

export function CreateTenantModal({ open, onClose, onSuccess }: CreateTenantModalProps) {
  const { message } = App.useApp();
  const [form] = Form.useForm<{ name: string }>();
  const [loading, setLoading] = useState(false);
  const [apiError, setApiError] = useState<string | null>(null);

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      setLoading(true);
      setApiError(null);

      const response = await developerApi.createTenant(values.name);
      if (response.success) {
        message.success(`Tenant "${values.name}" created successfully`);
        form.resetFields();
        onSuccess();
      } else {
        setApiError(response.error || "Failed to create tenant");
      }
    } catch (err) {
      if (err instanceof Error) {
        setApiError(err.message);
      }
      // Form validation errors are handled by Ant Design
    } finally {
      setLoading(false);
    }
  };

  const handleCancel = () => {
    form.resetFields();
    setApiError(null);
    onClose();
  };

  return (
    <Modal
      title="Create Tenant"
      open={open}
      onCancel={handleCancel}
      onOk={handleSubmit}
      okText="Create"
      confirmLoading={loading}
      destroyOnClose
      width={480}
    >
      <Form form={form} layout="vertical" style={{ marginTop: 16 }}>
        <Form.Item
          name="name"
          label="Tenant Name"
          rules={[
            { required: true, message: "Tenant name is required" },
            {
              pattern: TENANT_NAME_REGEX,
              message: "Must start with lowercase letter, 3-50 chars, only a-z, 0-9, underscore",
            },
          ]}
          help={apiError ? undefined : undefined}
          extra={
            <Text type="secondary" style={{ fontSize: 12 }}>
              Lowercase letters, numbers, and underscores only. Must start with a letter.
            </Text>
          }
        >
          <Input placeholder="e.g. my_store_01" autoFocus />
        </Form.Item>

        {apiError && (
          <div style={{ color: "#dc2626", marginTop: -8, marginBottom: 8, fontSize: 12 }}>
            {apiError}
          </div>
        )}
      </Form>
    </Modal>
  );
}
