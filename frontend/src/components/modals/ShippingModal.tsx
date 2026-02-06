import {
  Drawer,
  Form,
  Select,
  Input,
  Button,
  Space,
  Typography,
  Alert,
  Divider,
  Descriptions,
} from "antd";
import { Order } from "@/types/order";
import { useState, useEffect } from "react";

const { Text } = Typography;

export interface ShippingFormValues {
  shipping_provider: string;
  tracking_number: string;
  label_format?: string;
}

interface ShippingModalProps {
  open: boolean;
  onClose: () => void;
  onConfirm: (orderSn: string, values: ShippingFormValues) => Promise<void>;
  order: Order | null;
  loading?: boolean;
}

const SHIPPING_PROVIDERS = [
  { label: "JNE", value: "jne" },
  { label: "J&T", value: "jnt" },
  { label: "Sicepat", value: "sicepat" },
  { label: "GoSend", value: "gosend" },
  { label: "GrabExpress", value: "grabexpress" },
  { label: "AnterAja", value: "anteraja" },
  { label: "Shopee Xpress", value: "shopee_xpress" },
];

const LABEL_FORMATS = [
  { label: "Thermal 4x6", value: "thermal_4x6" },
  { label: "A4", value: "a4" },
  { label: "A5", value: "a5" },
];

export function ShippingModal({
  open,
  onClose,
  onConfirm,
  order,
  loading = false,
}: ShippingModalProps) {
  const [form] = Form.useForm<ShippingFormValues>();
  const [submitting, setSubmitting] = useState(false);
  const [selectedProvider, setSelectedProvider] = useState<string | null>(null);

  useEffect(() => {
    if (open && order) {
      form.resetFields();
      setSelectedProvider(null);
    }
  }, [open, order, form]);

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      if (!order) return;

      setSubmitting(true);
      await onConfirm(order.order_sn, values);
      setSubmitting(false);
      onClose();
    } catch (error) {
      setSubmitting(false);
    }
  };

  if (!order) return null;

  return (
    <Drawer
      title="Generate Shipping Label"
      placement="right"
      width={450}
      onClose={onClose}
      open={open}
      footer={
        <Space style={{ display: "flex", justifyContent: "flex-end" }}>
          <Button disabled={submitting || loading} onClick={onClose}>
            Cancel
          </Button>
          <Button
            type="primary"
            onClick={handleSubmit}
            loading={submitting || loading}
          >
            Generate Label
          </Button>
        </Space>
      }
    >
      <Descriptions title="Order Information" size="small" column={1} bordered>
        <Descriptions.Item label="Order SN">
          <Text strong copyable>
            {order.order_sn}
          </Text>
        </Descriptions.Item>
        <Descriptions.Item label="Amount">
          {order.total_amount.toLocaleString()}
        </Descriptions.Item>
        <Descriptions.Item label="Buyer">
          {order.buyer_username}
        </Descriptions.Item>
      </Descriptions>

      <Divider />

      <Alert
        message="Ensure tracking number is correct before generating label."
        type="info"
        showIcon
        style={{ marginBottom: 24 }}
      />

      <Form
        form={form}
        layout="vertical"
        initialValues={{
          shipping_provider: "",
          tracking_number: "",
          label_format: "thermal_4x6",
        }}
      >
        <Form.Item
          name="shipping_provider"
          label="Shipping Provider"
          rules={[{ required: true, message: "Please select a provider" }]}
        >
          <Select
            placeholder="Select provider"
            options={SHIPPING_PROVIDERS}
            onChange={setSelectedProvider}
            showSearch
          />
        </Form.Item>

        <Form.Item
          name="tracking_number"
          label="Tracking Number"
          rules={[
            { required: true, message: "Please enter tracking number" },
            { min: 5, message: "Tracking number too short" },
          ]}
        >
          <Input placeholder="Enter tracking number" />
        </Form.Item>

        <Form.Item
          name="label_format"
          label="Label Format"
          rules={[{ required: true }]}
        >
          <Select placeholder="Select format" options={LABEL_FORMATS} />
        </Form.Item>
      </Form>

      {selectedProvider && (
        <>
          <Divider />
          <div
            style={{ padding: "16px", background: "#f5f5f5", borderRadius: 3 }}
          >
            <Text strong>Label Preview</Text>
            <div style={{ marginTop: 12, fontSize: 12, color: "#666" }}>
              <p>Provider: {selectedProvider.toUpperCase()}</p>
              <p>Tracking: {form.getFieldValue("tracking_number") || "—"}</p>
              <p>Format: {form.getFieldValue("label_format")}</p>
            </div>
          </div>
        </>
      )}
    </Drawer>
  );
}
