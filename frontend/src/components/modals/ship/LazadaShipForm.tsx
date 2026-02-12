import { Alert, Button, Form, Input, Select, Space, Steps, theme } from "antd";
import { Order } from "@/types/order";
import type { ShipConfirmPayload } from "@/components/modals/OrderShipModal";

interface LazadaShipFormValues {
  shipping_provider: string;
  tracking_number?: string;
}

interface LazadaShipFormProps {
  order: Order;
  loading?: boolean;
  onClose: () => void;
  onConfirm: (payload: ShipConfirmPayload) => Promise<void>;
}

const SHIPPING_PROVIDER_OPTIONS = [
  { value: "Lazada Logistics", label: "Lazada Logistics" },
  { value: "JNE", label: "JNE" },
  { value: "J&T", label: "J&T" },
  { value: "SiCepat", label: "SiCepat" },
  { value: "Ninja Van", label: "Ninja Van" },
];

function extractOrderItemIds(order: Order): string[] {
  const record = order as unknown as Record<string, unknown>;
  const directValue = record.order_item_ids;
  if (Array.isArray(directValue)) {
    const ids = directValue.filter(
      (item): item is string => typeof item === "string",
    );
    if (ids.length > 0) {
      return ids;
    }
  }

  const single = record.order_item_id;
  if (typeof single === "string" && single.trim() !== "") {
    return [single];
  }

  return [order.order_sn];
}

export function LazadaShipForm({
  order,
  loading = false,
  onClose,
  onConfirm,
}: LazadaShipFormProps) {
  const { token } = theme.useToken();
  const [form] = Form.useForm<LazadaShipFormValues>();

  const handleSubmit = async () => {
    const values = await form.validateFields();
    const payload: ShipConfirmPayload = {
      platform: "lazada",
      data: {
        order_item_ids: extractOrderItemIds(order),
        shipping_provider: values.shipping_provider,
        tracking_number: values.tracking_number?.trim() || undefined,
      },
    };

    await onConfirm(payload);
    onClose();
  };

  return (
    <Form
      form={form}
      layout="vertical"
      initialValues={{
        shipping_provider: order.shipping_carrier || "",
        tracking_number: order.tracking_number || "",
      }}
    >
      <Alert
        type="info"
        showIcon
        message="Lazada shipment flow"
        description="Backend performs Pack and Ready To Ship automatically."
        style={{ marginBottom: 16, borderColor: token.colorInfoBorder }}
      />

      <Steps
        size="small"
        current={0}
        items={[{ title: "Pack" }, { title: "Ready To Ship" }]}
        style={{ marginBottom: 16 }}
      />

      <Form.Item
        name="shipping_provider"
        label="Shipping Provider"
        rules={[{ required: true, message: "Select shipping provider" }]}
      >
        <Select
          options={SHIPPING_PROVIDER_OPTIONS}
          placeholder="Select provider"
          showSearch
        />
      </Form.Item>

      <Form.Item name="tracking_number" label="Tracking Number (Optional)">
        <Input placeholder="Enter tracking number" />
      </Form.Item>

      <Space style={{ justifyContent: "flex-end", width: "100%" }}>
        <Button onClick={onClose} disabled={loading}>
          Cancel
        </Button>
        <Button type="primary" loading={loading} onClick={() => void handleSubmit()}>
          Confirm Shipment
        </Button>
      </Space>
    </Form>
  );
}
