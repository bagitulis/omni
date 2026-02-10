import { Modal, Form, Input, Button } from "antd";

interface SkuMappingModalProps {
  open: boolean;
  onClose: () => void;
  platform: string;
  isSubmitting?: boolean;
  onSubmit: (payload: {
    platformItemId: string;
    platformSkuId: string;
  }) => void;
}

export function SkuMappingModal({
  open,
  onClose,
  platform,
  isSubmitting = false,
  onSubmit,
}: SkuMappingModalProps) {
  const [form] = Form.useForm();

  const handleOk = async () => {
    try {
      const values = await form.validateFields();
      onSubmit(values);
    } catch (error) {
      // Validation failed
    }
  };

  return (
    <Modal
      title={`Link SKU to ${platform}`}
      open={open}
      onCancel={onClose}
      footer={[
        <Button key="cancel" onClick={onClose}>
          Cancel
        </Button>,
        <Button
          key="submit"
          type="primary"
          onClick={handleOk}
          loading={isSubmitting}
        >
          {isSubmitting ? "Linking..." : "Link"}
        </Button>,
      ]}
      width={400}
      destroyOnClose
    >
      <Form form={form} layout="vertical" preserve={false}>
        <Form.Item
          name="platformItemId"
          label="Platform Item ID"
          rules={[{ required: true, message: "Please enter Platform Item ID" }]}
        >
          <Input placeholder="e.g., 12345678" />
        </Form.Item>
        <Form.Item name="platformSkuId" label="Platform SKU ID (Optional)">
          <Input placeholder="e.g., sku_12345" />
        </Form.Item>
      </Form>
    </Modal>
  );
}
