import { useEffect } from "react";
import { Modal, Form, Input, InputNumber, Button, Space } from "antd";
import type { ReportSettings } from "@/types/analytics";

interface ReportSettingsModalProps {
  open: boolean;
  onClose: () => void;
  settings: ReportSettings | null | undefined;
  onSave: (settings: Partial<ReportSettings>) => void;
  saving?: boolean;
}

/**
 * Modal dialog for editing report settings.
 * Fields: price_column, formula_deduction, formula_multiplier.
 */
export function ReportSettingsModal({
  open,
  onClose,
  settings,
  onSave,
  saving,
}: ReportSettingsModalProps) {
  const [form] = Form.useForm();

  useEffect(() => {
    if (open && settings) {
      form.setFieldsValue({
        price_column: settings.price_column,
        formula_deduction: settings.formula_deduction,
        formula_multiplier: settings.formula_multiplier,
      });
    }
  }, [open, settings, form]);

  const handleFinish = (values: {
    price_column: string;
    formula_deduction: number;
    formula_multiplier: number;
  }) => {
    onSave(values);
  };

  return (
    <Modal
      title="Report Settings"
      open={open}
      onCancel={onClose}
      footer={null}
      destroyOnHidden
      width={480}
    >
      <Form
        form={form}
        layout="vertical"
        onFinish={handleFinish}
        initialValues={{
          price_column: settings?.price_column || "price",
          formula_deduction: settings?.formula_deduction || 0,
          formula_multiplier: settings?.formula_multiplier || 1,
        }}
      >
        <Form.Item
          name="price_column"
          label="Price Column"
          rules={[{ required: true, message: "Please select a price column" }]}
        >
          <Input placeholder="e.g. price, cost, etc." />
        </Form.Item>

        <Form.Item
          name="formula_deduction"
          label="Formula Deduction"
          rules={[
            { required: true, message: "Please enter a deduction value" },
          ]}
        >
          <InputNumber
            style={{ width: "100%" }}
            placeholder="0"
            precision={2}
            step={100}
          />
        </Form.Item>

        <Form.Item
          name="formula_multiplier"
          label="Formula Multiplier"
          rules={[
            { required: true, message: "Please enter a multiplier value" },
          ]}
        >
          <InputNumber
            style={{ width: "100%" }}
            placeholder="1"
            precision={2}
            step={0.1}
            min={0}
          />
        </Form.Item>

        <Form.Item style={{ marginBottom: 0, textAlign: "right" }}>
          <Space>
            <Button onClick={onClose} disabled={saving}>
              Cancel
            </Button>
            <Button type="primary" htmlType="submit" loading={saving}>
              Save
            </Button>
          </Space>
        </Form.Item>
      </Form>
    </Modal>
  );
}
