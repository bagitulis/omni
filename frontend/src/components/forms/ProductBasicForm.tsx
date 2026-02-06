import { Form, Input, Button } from "antd";
import { useEffect } from "react";

export interface BasicFormValues {
  item_name: string;
  description: string;
  brand?: string;
}

interface Props {
  initialValues: BasicFormValues;
  onFinish: (values: BasicFormValues) => void;
  submitLabel?: string;
  hideSubmit?: boolean;
}

export function ProductBasicForm({
  initialValues,
  onFinish,
  submitLabel = "Next Step",
  hideSubmit = false,
}: Props) {
  const [form] = Form.useForm<BasicFormValues>();

  useEffect(() => {
    form.setFieldsValue(initialValues);
  }, [initialValues, form]);

  return (
    <Form
      form={form}
      layout="vertical"
      onFinish={onFinish}
      initialValues={initialValues}
      autoComplete="off"
      style={{ maxWidth: 600, margin: "0 auto" }}
    >
      <Form.Item
        label={<span className="font-medium">Product Name</span>}
        name="item_name"
        rules={[
          { required: true, message: "Please enter product name" },
          { max: 120, message: "Max 120 characters" },
        ]}
      >
        <Input
          showCount
          maxLength={120}
          placeholder="Ex: Samsung Galaxy S24 Ultra"
        />
      </Form.Item>

      <Form.Item
        label={<span className="font-medium">Description</span>}
        name="description"
        rules={[
          { required: true, message: "Please enter description" },
          { max: 5000, message: "Max 5000 characters" },
        ]}
      >
        <Input.TextArea
          showCount
          maxLength={5000}
          rows={6}
          placeholder="Product details, specifications, etc."
        />
      </Form.Item>

      <Form.Item
        label={<span className="font-medium">Brand</span>}
        name="brand"
      >
        <Input placeholder="Ex: Samsung" />
      </Form.Item>

      {!hideSubmit && (
        <div
          style={{ display: "flex", justifyContent: "flex-end", marginTop: 24 }}
        >
          <Button type="primary" htmlType="submit">
            {submitLabel}
          </Button>
        </div>
      )}
    </Form>
  );
}
