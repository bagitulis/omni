import { Form, InputNumber, Button, Card, Row, Col } from "antd";
import { useEffect } from "react";

export interface PricingFormValues {
  weight: number;
  length: number;
  width: number;
  height: number;
}

interface Props {
  initialValues: PricingFormValues;
  onFinish: (values: PricingFormValues) => void;
  onBack: () => void;
  submitting?: boolean;
}

export function ProductPricingForm({
  initialValues,
  onFinish,
  onBack,
  submitting = false,
}: Props) {
  const [form] = Form.useForm<PricingFormValues>();

  useEffect(() => {
    form.setFieldsValue(initialValues);
  }, [initialValues, form]);

  return (
    <Form
      form={form}
      layout="vertical"
      onFinish={onFinish}
      initialValues={initialValues}
      style={{ maxWidth: 800, margin: "0 auto" }}
    >
      <Card
        title="Shipping Information"
        variant="borderless"
        style={{ marginBottom: 24 }}
      >
        <Row gutter={16}>
          <Col span={24}>
            <Form.Item
              label={<span className="font-medium">Weight (grams)</span>}
              name="weight"
              rules={[{ required: true, message: "Weight is required" }]}
            >
              <InputNumber<number>
                style={{ width: "100%" }}
                placeholder="e.g. 1000"
                min={1}
                addonAfter="g"
              />
            </Form.Item>
          </Col>
        </Row>

        <div style={{ marginBottom: 8, fontWeight: 500 }}>Dimensions</div>
        <Row gutter={16}>
          <Col span={8}>
            <Form.Item
              label="Length"
              name="length"
              rules={[{ required: true, message: "Required" }]}
            >
              <InputNumber<number>
                style={{ width: "100%" }}
                placeholder="Length"
                min={1}
                addonAfter="cm"
              />
            </Form.Item>
          </Col>
          <Col span={8}>
            <Form.Item
              label="Width"
              name="width"
              rules={[{ required: true, message: "Required" }]}
            >
              <InputNumber<number>
                style={{ width: "100%" }}
                placeholder="Width"
                min={1}
                addonAfter="cm"
              />
            </Form.Item>
          </Col>
          <Col span={8}>
            <Form.Item
              label="Height"
              name="height"
              rules={[{ required: true, message: "Required" }]}
            >
              <InputNumber<number>
                style={{ width: "100%" }}
                placeholder="Height"
                min={1}
                addonAfter="cm"
              />
            </Form.Item>
          </Col>
        </Row>
      </Card>

      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          marginTop: 24,
        }}
      >
        <Button onClick={onBack}>Back</Button>
        <Button type="primary" htmlType="submit" loading={submitting}>
          Submit Product
        </Button>
      </div>
    </Form>
  );
}
