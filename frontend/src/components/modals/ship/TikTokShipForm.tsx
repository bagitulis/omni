import { Alert, Button, Form, Input, Select, Space, Spin, theme } from "antd";
import { useEffect, useMemo, useState } from "react";
import { Order } from "@/types/order";
import type { ShipConfirmPayload } from "@/components/modals/OrderShipModal";
import { TikTokTimeSlot, getHandoverTimeSlots } from "@/hooks/useOrders";

interface TikTokShipFormValues {
  handover_method: "PICKUP" | "DROP_OFF" | "SELF_SHIPMENT";
  pickup_slot_key?: string;
  tracking_number?: string;
  shipping_provider_id?: string;
}

interface TikTokShipFormProps {
  order: Order;
  loading?: boolean;
  onClose: () => void;
  onConfirm: (payload: ShipConfirmPayload) => Promise<void>;
}

const getSlotKey = (slot: TikTokTimeSlot): string =>
  `${slot.start_time}:${slot.end_time}`;

export function TikTokShipForm({
  order,
  loading = false,
  onClose,
  onConfirm,
}: TikTokShipFormProps) {
  const { token } = theme.useToken();
  const [form] = Form.useForm<TikTokShipFormValues>();
  const [timeSlots, setTimeSlots] = useState<TikTokTimeSlot[]>([]);
  const [fetching, setFetching] = useState(false);
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    const load = async () => {
      setFetching(true);
      try {
        const response = await getHandoverTimeSlots(order.order_sn);
        setTimeSlots(response.time_slots || []);
      } finally {
        setFetching(false);
      }
    };

    void load();
  }, [order.order_sn]);

  const handoverMethod = Form.useWatch("handover_method", form);

  const slotMap = useMemo(() => {
    const map = new Map<string, TikTokTimeSlot>();
    for (const slot of timeSlots) {
      map.set(getSlotKey(slot), slot);
    }
    return map;
  }, [timeSlots]);

  const slotOptions = useMemo(
    () =>
      timeSlots.map((slot) => ({
        value: getSlotKey(slot),
        label: `${new Date(slot.start_time * 1000).toLocaleString()} - ${new Date(slot.end_time * 1000).toLocaleString()}`,
      })),
    [timeSlots],
  );

  const handleSubmit = async () => {
    const values = await form.validateFields();
    const payload: ShipConfirmPayload = {
      platform: "tiktok",
      data: {
        order_id: order.order_sn,
        handover_method: values.handover_method,
      },
    };

    if (values.handover_method === "PICKUP") {
      const selectedSlot = slotMap.get(values.pickup_slot_key || "");
      if (selectedSlot) {
        payload.data.pickup_slot = {
          start_time: selectedSlot.start_time,
          end_time: selectedSlot.end_time,
        };
      }
    }

    if (values.handover_method === "SELF_SHIPMENT") {
      payload.data.self_shipment = {
        tracking_number: values.tracking_number!.trim(),
        shipping_provider_id: values.shipping_provider_id!.trim(),
      };
    }

    setSubmitting(true);
    try {
      await onConfirm(payload);
      onClose();
    } finally {
      setSubmitting(false);
    }
  };

  if (fetching) {
    return <Spin />;
  }

  return (
    <Form
      form={form}
      layout="vertical"
      initialValues={{ handover_method: "PICKUP" }}
    >
      <Alert
        type="info"
        showIcon
        message="TikTok shipping arrangement"
        description="Pickup requires timeslot. Self shipment requires provider ID and tracking number."
        style={{ marginBottom: 16, borderColor: token.colorInfoBorder }}
      />

      <Form.Item
        name="handover_method"
        label="Handover Method"
        rules={[{ required: true, message: "Select handover method" }]}
      >
        <Select
          options={[
            { value: "PICKUP", label: "Pickup" },
            { value: "DROP_OFF", label: "Dropoff" },
            { value: "SELF_SHIPMENT", label: "Self Shipment" },
          ]}
        />
      </Form.Item>

      {handoverMethod === "PICKUP" && (
        <Form.Item
          name="pickup_slot_key"
          label="Pickup Timeslot"
          rules={[{ required: true, message: "Select pickup timeslot" }]}
        >
          <Select options={slotOptions} placeholder="Select pickup timeslot" />
        </Form.Item>
      )}

      {handoverMethod === "SELF_SHIPMENT" && (
        <>
          <Form.Item
            name="shipping_provider_id"
            label="Shipping Provider ID"
            rules={[{ required: true, message: "Enter shipping provider ID" }]}
          >
            <Input placeholder="e.g. 100034" />
          </Form.Item>

          <Form.Item
            name="tracking_number"
            label="Tracking Number"
            rules={[{ required: true, message: "Enter tracking number" }]}
          >
            <Input placeholder="Enter tracking number" />
          </Form.Item>
        </>
      )}

      <Space style={{ justifyContent: "flex-end", width: "100%" }}>
        <Button onClick={onClose} disabled={loading || submitting}>
          Cancel
        </Button>
        <Button
          type="primary"
          loading={loading || submitting}
          onClick={() => void handleSubmit()}
        >
          Confirm Shipment
        </Button>
      </Space>
    </Form>
  );
}
