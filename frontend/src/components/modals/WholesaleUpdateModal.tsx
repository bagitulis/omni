import {
  Modal,
  Form,
  InputNumber,
  Button,
  Space,
  Typography,
  message,
} from "antd";
import { MinusCircleOutlined, PlusOutlined } from "@ant-design/icons";
import { useEffect } from "react";
import { useBatchUpdateInventoryWholesale } from "@/hooks/useWholesale";
import type { InventoryWholesaleTier } from "@/types/wholesale";

interface WholesaleUpdateModalProps {
  open: boolean;
  onClose: () => void;
  selectedSkus: string[];
}

export function WholesaleUpdateModal({
  open,
  onClose,
  selectedSkus,
}: WholesaleUpdateModalProps) {
  const [form] = Form.useForm();
  const { mutate: batchUpdate, isPending } = useBatchUpdateInventoryWholesale();

  useEffect(() => {
    if (open) {
      form.resetFields();
    }
  }, [open, form]);

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      const tiers = values.tiers as InventoryWholesaleTier[];

      if (!tiers || tiers.length === 0) {
        message.error("Please add at least one tier");
        return;
      }

      const items = selectedSkus.map((sku) => ({
        sku,
        tiers: tiers.map((t) => ({
          ...t,
          sku,
        })),
      }));

      batchUpdate(items, {
        onSuccess: (data) => {
          message.success(
            `Update successful: ${data.successful} updated, ${data.failed} failed`,
          );
          onClose();
        },
        onError: (error) => {
          message.error(`Failed to update: ${error.message}`);
        },
      });
    } catch {
      // Form validation error
    }
  };

  return (
    <Modal
      title={`Batch Update Wholesale (${selectedSkus.length} items)`}
      open={open}
      onCancel={onClose}
      destroyOnClose
      width={600}
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
          This will apply the following wholesale tiers to all selected SKUs.
          Existing tiers will be replaced.
        </Typography.Text>
      </div>

      <Form form={form} layout="vertical">
        <Form.List name="tiers" initialValue={[{ min_qty: 1, price: 0 }]}>
          {(fields, { add, remove }) => (
            <>
              {fields.map(({ key, name, ...restField }) => (
                <Space
                  key={key}
                  style={{ display: "flex", marginBottom: 8 }}
                  align="baseline"
                >
                  <Form.Item
                    {...restField}
                    name={[name, "min_qty"]}
                    label={key === 0 ? "Min Qty" : undefined}
                    rules={[{ required: true, message: "Missing min qty" }]}
                  >
                    <InputNumber min={1} placeholder="Min Qty" />
                  </Form.Item>
                  <Form.Item
                    {...restField}
                    name={[name, "price"]}
                    label={key === 0 ? "Unit Price" : undefined}
                    rules={[{ required: true, message: "Missing price" }]}
                  >
                    <InputNumber
                      min={0}
                      precision={2}
                      prefix="$"
                      placeholder="Price"
                      style={{ width: "100%" }}
                    />
                  </Form.Item>
                  <MinusCircleOutlined onClick={() => remove(name)} />
                </Space>
              ))}
              <Form.Item>
                <Button
                  type="dashed"
                  onClick={() => add()}
                  block
                  icon={<PlusOutlined />}
                >
                  Add Tier
                </Button>
              </Form.Item>
            </>
          )}
        </Form.List>
      </Form>
    </Modal>
  );
}
