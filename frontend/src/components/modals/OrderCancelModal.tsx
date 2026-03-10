import {
  Modal,
  Form,
  Radio,
  Input,
  Checkbox,
  Button,
  Alert,
  Typography,
  theme,
} from "antd";
import { Order } from "@/types/order";
import { useState, useEffect } from "react";
import { message } from "@/components/AntStaticHolder";

const { Text } = Typography;
const { TextArea } = Input;

export interface CancelFormValues {
  cancel_reason: string;
  reason_detail?: string;
}

interface OrderCancelModalProps {
  open: boolean;
  onClose: () => void;
  onConfirm: (orderSn: string, values: CancelFormValues) => Promise<void>;
  order: Order | null;
  loading?: boolean;
}

const CANCEL_REASONS = [
  { label: "Out of Stock", value: "out_of_stock" },
  { label: "Customer Request", value: "customer_request" },
  { label: "Undeliverable Area", value: "undeliverable_area" },
  { label: "Pricing Error", value: "pricing_error" },
  { label: "Other", value: "other" },
];

export function OrderCancelModal({
  open,
  onClose,
  onConfirm,
  order,
  loading = false,
}: OrderCancelModalProps) {
  const { token } = theme.useToken();
  const [form] = Form.useForm();
  const [submitting, setSubmitting] = useState(false);
  const [confirmed, setConfirmed] = useState(false);

  useEffect(() => {
    if (open) {
      form.resetFields();
      setConfirmed(false);
    }
  }, [open, form]);

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      if (!order) return;

      setSubmitting(true);
      await onConfirm(order.order_sn, values);
      setSubmitting(false);
      onClose();
    } catch (err: unknown) {
      const errorMsg =
        err instanceof Error ? err.message : "Failed to cancel order";
      message.error(errorMsg);
      setSubmitting(false);
    }
  };

  if (!order) return null;

  return (
    <Modal
      title="Cancel Order"
      open={open}
      onCancel={onClose}
      footer={[
        <Button key="back" onClick={onClose} disabled={submitting || loading}>
          Go Back
        </Button>,
        <Button
          key="submit"
          type="primary"
          danger
          onClick={handleSubmit}
          loading={submitting || loading}
          disabled={!confirmed}
        >
          Cancel Order
        </Button>,
      ]}
    >
      <Alert
        message="Warning: Cancellation cannot be undone"
        description="This action will permanently cancel the order and notify the buyer. Penalties may apply depending on the platform policy."
        type="warning"
        showIcon
        style={{ marginBottom: 20 }}
      />

      <div style={{ marginBottom: 16 }}>
        <Text type="secondary">Cancelling Order: </Text>
        <Text strong>{order.order_sn}</Text>
      </div>

      <Form
        form={form}
        layout="vertical"
        initialValues={{ cancel_reason: "out_of_stock" }}
      >
        <Form.Item
          name="cancel_reason"
          label="Reason for Cancellation"
          rules={[{ required: true, message: "Please select a reason" }]}
        >
          <Radio.Group
            style={{ display: "flex", flexDirection: "column", gap: 8 }}
          >
            {CANCEL_REASONS.map((reason) => (
              <Radio key={reason.value} value={reason.value}>
                {reason.label}
              </Radio>
            ))}
          </Radio.Group>
        </Form.Item>

        <Form.Item name="reason_detail" label="Additional Details (Optional)">
          <TextArea
            rows={3}
            placeholder="Provide more details about why this order is being cancelled..."
          />
        </Form.Item>
      </Form>

      <div
        style={{
          marginTop: 16,
          padding: "12px",
          background: token.colorErrorBg,
          borderRadius: 4,
          border: `1px solid ${token.colorBorder}`,
        }}
      >
        <Checkbox
          checked={confirmed}
          onChange={(e) => setConfirmed(e.target.checked)}
        >
          I confirm that I want to cancel order{" "}
          <Text strong>{order.order_sn}</Text>
        </Checkbox>
      </div>
    </Modal>
  );
}
