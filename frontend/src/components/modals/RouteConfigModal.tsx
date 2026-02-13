import { Modal, Form, Input, Select, Button } from "antd";
import { useEffect } from "react";
import { RouteExecutionConfig } from "@/types/routeExecutionConfig";

const { Option } = Select;

interface RouteConfigModalProps {
  open: boolean;
  onClose: () => void;
  config: RouteExecutionConfig | null;
  onSave: (config: RouteExecutionConfig) => void;
}

export function RouteConfigModal({
  open,
  onClose,
  config,
  onSave,
}: RouteConfigModalProps) {
  const [form] = Form.useForm();
  const isEdit = !!config;

  useEffect(() => {
    if (open) {
      if (config) {
        form.setFieldsValue(config);
      } else {
        form.resetFields();
        form.setFieldsValue({
          execution_mode: "queue",
          priority: "normal",
          category: "general",
        });
      }
    }
  }, [open, config, form]);

  const handleFinish = (values: unknown) => {
    onSave({
      ...(config as Record<string, unknown>),
      ...values,
    } as RouteExecutionConfig);
    onClose();
  };

  return (
    <Modal
      title={isEdit ? "Edit Route" : "Add New Route"}
      open={open}
      onCancel={onClose}
      footer={null}
      width={500}
    >
      <Form form={form} layout="vertical" onFinish={handleFinish}>
        <Form.Item
          name="route_key"
          label="Route Key"
          rules={[{ required: true, message: "Please enter route key" }]}
          extra="Unique identifier (no spaces)"
        >
          <Input placeholder="e.g., update_stock" disabled={isEdit} />
        </Form.Item>

        <Form.Item
          name="route_name"
          label="Route Name"
          rules={[{ required: true, message: "Please enter route name" }]}
        >
          <Input placeholder="e.g., Update Stock" />
        </Form.Item>

        <Form.Item name="description" label="Description">
          <Input placeholder="e.g., Update stock to all platforms" />
        </Form.Item>

        <div style={{ display: "flex", gap: 16 }}>
          <Form.Item
            name="execution_mode"
            label="Execution Mode"
            style={{ flex: 1 }}
          >
            <Select>
              <Option value="queue">Queue</Option>
              <Option value="direct">Direct</Option>
            </Select>
          </Form.Item>

          <Form.Item
            name="priority"
            label="Priority"
            style={{ flex: 1 }}
            shouldUpdate={(
              prev: Record<string, unknown>,
              curr: Record<string, unknown>,
            ) => prev.execution_mode !== curr.execution_mode}
          >
            {({ getFieldValue }) => (
              <Select disabled={getFieldValue("execution_mode") !== "queue"}>
                <Option value="high">High</Option>
                <Option value="normal">Normal</Option>
                <Option value="low">Low</Option>
              </Select>
            )}
          </Form.Item>
        </div>

        <div style={{ display: "flex", gap: 16 }}>
          <Form.Item name="icon" label="Icon" style={{ flex: 1 }}>
            <Input placeholder="Icon name" />
          </Form.Item>

          <Form.Item name="category" label="Category" style={{ flex: 1 }}>
            <Select>
              <Option value="inventory">Inventory</Option>
              <Option value="orders">Orders</Option>
              <Option value="auth">Auth</Option>
              <Option value="sync">Sync</Option>
              <Option value="general">General</Option>
            </Select>
          </Form.Item>
        </div>

        <div style={{ display: "flex", justifyContent: "flex-end", gap: 8 }}>
          <Button onClick={onClose}>Cancel</Button>
          <Button type="primary" htmlType="submit">
            {isEdit ? "Update" : "Create"}
          </Button>
        </div>
      </Form>
    </Modal>
  );
}
