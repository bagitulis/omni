import {
  Drawer,
  Form,
  Select,
  Input,
  Button,
  Space,
  Typography,
  Alert,
} from "antd";
import { Order } from "@/types/order";
import { useState, useEffect } from "react";

const { Text } = Typography;

export interface ShipFormValues {
  shipping_provider: string;
  tracking_number: string;
}

interface OrderShipModalProps {
  open: boolean;
  onClose: () => void;
  onConfirm: (orderSn: string, values: ShipFormValues) => Promise<void>;
  order: Order | null;
  loading?: boolean;
}

const MOCK_PROVIDERS = [
  { label: "JNE", value: "jne" },
  { label: "J&T", value: "jnt" },
  { label: "Sicepat", value: "sicepat" },
  { label: "GoSend", value: "gosend" },
  { label: "GrabExpress", value: "grabexpress" },
  { label: "AnterAja", value: "anteraja" },
  { label: "Shopee Xpress", value: "shopee_xpress" },
];

export function OrderShipModal({
  open,
  onClose,
  onConfirm,
  order,
  loading = false,
}: OrderShipModalProps) {
  const [form] = Form.useForm<ShipFormValues>();
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (open && order) {
      form.resetFields();
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
      // Validation failed or onConfirm failed
    }
  };

  if (!order) return null;

  return (
    <Drawer
      title="Arrange Shipment"
      placement="right"
      width={400}
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
            Confirm Shipment
          </Button>
        </Space>
      }
    >
      <div style={{ marginBottom: 24 }}>
        <Text type="secondary">Order SN: </Text>
        <Text strong copyable>
          {order.order_sn}
        </Text>
      </div>

      <Alert
        message="Please ensure the tracking number is correct."
        type="info"
        showIcon
        style={{ marginBottom: 24 }}
      />

      <Form
        form={form}
        layout="vertical"
        initialValues={{ shipping_provider: "", tracking_number: "" }}
      >
        <Form.Item
          name="shipping_provider"
          label="Shipping Provider"
          rules={[
            { required: true, message: "Please select a shipping provider" },
          ]}
        >
          <Select
            placeholder="Select provider"
            options={MOCK_PROVIDERS}
            showSearch
          />
        </Form.Item>

        <Form.Item
          name="tracking_number"
          label="Tracking Number"
          rules={[
            { required: true, message: "Please enter tracking number" },
            { min: 5, message: "Tracking number seems too short" },
          ]}
        >
          <Input placeholder="Enter tracking number" />
        </Form.Item>
      </Form>
    </Drawer>
  );
}
