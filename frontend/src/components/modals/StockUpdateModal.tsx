import {
  Modal,
  Form,
  InputNumber,
  Select,
  Checkbox,
  Button,
  Alert,
} from "antd";
import { useState, useEffect } from "react";

export interface StockUpdateFormValues {
  sku: string;
  quantity: number;
  operation: "set" | "add" | "subtract";
  sync_to_platforms: boolean;
}
interface StockUpdateModalProps {
  open: boolean;
  onClose: () => void;
  onConfirm: (values: StockUpdateFormValues) => Promise<void>;
  skuOptions: Array<{ label: string; value: string }>;
  loading?: boolean;
}

export function StockUpdateModal({
  open,
  onClose,
  onConfirm,
  skuOptions,
  loading = false,
}: StockUpdateModalProps) {
  const [form] = Form.useForm<StockUpdateFormValues>();
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (open) {
      form.resetFields();
    }
  }, [open, form]);

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      setSubmitting(true);
      await onConfirm(values);
      setSubmitting(false);
      onClose();
    } catch (err) { console.warn("Operation failed:", err);
      setSubmitting(false);
    }
  };

  const operation = Form.useWatch("operation", form);

  return (
    <Modal
      title="Update Stock Quantity"
      open={open}
      onCancel={onClose}
      footer={[
        <Button key="cancel" disabled={submitting || loading} onClick={onClose}>
          Cancel
        </Button>,
        <Button
          key="submit"
          type="primary"
          onClick={handleSubmit}
          loading={submitting || loading}
        >
          Update Stock
        </Button>,
      ]}
    >
      <Alert
        message="Changes will be applied immediately to inventory"
        type="info"
        showIcon
        style={{ marginBottom: 24 }}
      />

      <Form
        form={form}
        layout="vertical"
        initialValues={{
          operation: "set",
          sync_to_platforms: true,
        }}
      >
        <Form.Item
          name="sku"
          label="Product SKU"
          rules={[{ required: true, message: "Please select a product" }]}
        >
          <Select
            placeholder="Select product SKU"
            options={skuOptions}
            showSearch
            filterOption={(input, option) =>
              (option?.label ?? "").toLowerCase().includes(input.toLowerCase())
            }
          />
        </Form.Item>

        <Form.Item
          name="operation"
          label="Operation Type"
          rules={[{ required: true, message: "Please select an operation" }]}
        >
          <Select
            options={[
              { label: "Set to exact value", value: "set" },
              { label: "Add to current stock", value: "add" },
              { label: "Subtract from stock", value: "subtract" },
            ]}
          />
        </Form.Item>

        <Form.Item
          name="quantity"
          label={
            operation === "set"
              ? "New Stock Quantity"
              : operation === "add"
                ? "Quantity to Add"
                : "Quantity to Subtract"
          }
          rules={[
            { required: true, message: "Please enter quantity" },
            {
              pattern: /^[0-9]+$/,
              message: "Quantity must be a positive number",
            },
          ]}
        >
          <InputNumber
            placeholder="0"
            min={0}
            precision={0}
            style={{ width: "100%" }}
          />
        </Form.Item>

        <Form.Item
          name="sync_to_platforms"
          valuePropName="checked"
          initialValue={true}
        >
          <Checkbox>Sync changes to all connected platforms</Checkbox>
        </Form.Item>
      </Form>
    </Modal>
  );
}
