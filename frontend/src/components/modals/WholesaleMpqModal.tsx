import {
  Modal,
  Form,
  InputNumber,
  Button,
  Typography,
  message,
  Switch,
} from "antd";
import { useEffect } from "react";
import { useBatchUpdateInventoryMpq } from "@/hooks/useWholesale";

interface WholesaleMpqModalProps {
  open: boolean;
  onClose: () => void;
  selectedSkus: string[];
}

export function WholesaleMpqModal({
  open,
  onClose,
  selectedSkus,
}: WholesaleMpqModalProps) {
  const [form] = Form.useForm();
  const { mutate: batchUpdate, isPending } = useBatchUpdateInventoryMpq();

  useEffect(() => {
    if (open) {
      form.resetFields();
    }
  }, [open, form]);

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();

      const items = selectedSkus.map((sku) => ({
        sku,
        min_purchase_qty: values.min_purchase_qty,
        enabled: values.enabled,
      }));

      batchUpdate(items, {
        onSuccess: (data) => {
          message.success(`Updated MPQ for ${data.successful} items`);
          onClose();
        },
        onError: (error) => {
          message.error(`Failed to update MPQ: ${error.message}`);
        },
      });
    } catch {
      // Form validation error
    }
  };

  return (
    <Modal
      title={`Batch Update MPQ (${selectedSkus.length} items)`}
      open={open}
      onCancel={onClose}
      destroyOnClose
      width={500}
      footer={[
        <Button key="cancel" onClick={onClose}>
          Cancel
        </Button>,
        <Button
          key="submit"
          type="primary"
          onClick={handleSubmit}
          loading={isPending}
        >
          Update All
        </Button>,
      ]}
    >
      <div style={{ marginBottom: 16 }}>
        <Typography.Text type="secondary">
          Set Minimum Purchase Quantity (MPQ) for selected items.
        </Typography.Text>
      </div>

      <Form
        form={form}
        layout="vertical"
        initialValues={{ min_purchase_qty: 1, enabled: true }}
      >
        <Form.Item
          name="min_purchase_qty"
          label="Minimum Purchase Quantity"
          rules={[{ required: true, message: "Required" }]}
        >
          <InputNumber min={1} style={{ width: "100%" }} />
        </Form.Item>

        <Form.Item name="enabled" label="Enable MPQ" valuePropName="checked">
          <Switch />
        </Form.Item>
      </Form>
    </Modal>
  );
}
