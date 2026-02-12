import { Alert, Button, Form, Input, Radio, Select, Space, Spin, theme } from "antd";
import { useEffect, useMemo, useState } from "react";
import { Order } from "@/types/order";
import type { ShipConfirmPayload } from "@/components/modals/OrderShipModal";
import { getShippingOptions, ShopeePickupAddress, ShopeeTimeSlot } from "@/hooks/useOrders";

interface ShopeeShipFormValues {
  mode: "pickup" | "dropoff";
  address_id?: number;
  pickup_time_id?: string;
  branch_id?: number;
  tracking_number?: string;
}

interface ShopeeShipFormProps {
  order: Order;
  loading?: boolean;
  onClose: () => void;
  onConfirm: (payload: ShipConfirmPayload) => Promise<void>;
}

interface ShopeeOptionsState {
  pickup: ShopeePickupAddress[];
  dropoff: Array<{ branch_id: number; address: string }>;
}

const getTimeSlots = (address?: ShopeePickupAddress): ShopeeTimeSlot[] => address?.time_slots || address?.time_slot_list || [];
const getTimeSlotLabel = (slot: ShopeeTimeSlot): string => {
  const text = slot.pickup_time || slot.time_text || slot.time || slot.time_slot || "No label";
  return slot.date ? `${slot.date} - ${text}` : text;
};

export function ShopeeShipForm({ order, loading = false, onClose, onConfirm }: ShopeeShipFormProps) {
  const { token } = theme.useToken();
  const [form] = Form.useForm<ShopeeShipFormValues>();
  const [options, setOptions] = useState<ShopeeOptionsState>({ pickup: [], dropoff: [] });
  const [fetching, setFetching] = useState(false);
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    const load = async () => {
      setFetching(true);
      try {
        const response = await getShippingOptions(order.order_sn);
        setOptions({ pickup: response.pickup || [], dropoff: response.dropoff || [] });
      } finally {
        setFetching(false);
      }
    };
    void load();
  }, [order.order_sn]);

  const hasPickup = options.pickup.length > 0;
  const hasDropoff = options.dropoff.length > 0;
  const showTrackingFallback = !hasPickup && !hasDropoff;
  const mode = Form.useWatch("mode", form);
  const addressId = Form.useWatch("address_id", form);

  const selectedAddress = useMemo(() => options.pickup.find((item) => item.address_id === addressId), [options.pickup, addressId]);
  const timeSlotOptions = useMemo(
    () => getTimeSlots(selectedAddress).map((slot) => ({ value: slot.pickup_time_id || slot.time_slot || "", label: getTimeSlotLabel(slot) })),
    [selectedAddress],
  );

  const handleSubmit = async () => {
    const values = await form.validateFields();
    const payload: ShipConfirmPayload = { platform: "shopee", data: { order_sn: order.order_sn } };
    if (showTrackingFallback) {
      payload.data.tracking_number = values.tracking_number?.trim();
    } else if ((values.mode || "pickup") === "pickup") {
      payload.data.pickup = { address_id: values.address_id!, pickup_time_id: values.pickup_time_id };
    } else {
      payload.data.dropoff = { branch_id: values.branch_id! };
    }

    setSubmitting(true);
    try {
      await onConfirm(payload);
      onClose();
    } finally {
      setSubmitting(false);
    }
  };

  if (fetching) return <Spin />;

  return (
    <Form form={form} layout="vertical" initialValues={{ mode: hasPickup ? "pickup" : "dropoff" }}>
      <Alert
        type="info"
        showIcon
        message="Shopee shipment parameters"
        description="Select pickup/dropoff from Shopee options."
        style={{ marginBottom: 16, borderColor: token.colorInfoBorder }}
      />

      {hasPickup && hasDropoff && (
        <Form.Item name="mode" label="Handover Method" rules={[{ required: true }]}>
          <Radio.Group options={[{ label: "Pickup", value: "pickup" }, { label: "Dropoff", value: "dropoff" }]} />
        </Form.Item>
      )}

      {(mode === "pickup" || (hasPickup && !hasDropoff)) && hasPickup && (
        <>
          <Form.Item name="address_id" label="Pickup Address" rules={[{ required: true, message: "Select pickup address" }]}>
            <Select options={options.pickup.map((address) => ({ value: address.address_id, label: address.address }))} placeholder="Select address" />
          </Form.Item>
          {timeSlotOptions.length > 0 && (
            <Form.Item name="pickup_time_id" label="Pickup Timeslot" rules={[{ required: true, message: "Select pickup timeslot" }]}>
              <Select options={timeSlotOptions} placeholder="Select pickup timeslot" />
            </Form.Item>
          )}
        </>
      )}

      {(mode === "dropoff" || (!hasPickup && hasDropoff)) && hasDropoff && (
        <Form.Item name="branch_id" label="Dropoff Branch" rules={[{ required: true, message: "Select dropoff branch" }]}>
          <Select options={options.dropoff.map((branch) => ({ value: branch.branch_id, label: branch.address }))} placeholder="Select dropoff branch" />
        </Form.Item>
      )}

      {showTrackingFallback && (
        <Form.Item
          name="tracking_number"
          label="Tracking Number"
          rules={[{ required: true, message: "Enter tracking number" }, { min: 5, message: "Tracking number is too short" }]}
        >
          <Input placeholder="Enter tracking number" />
        </Form.Item>
      )}

      <Space style={{ justifyContent: "flex-end", width: "100%" }}>
        <Button onClick={onClose} disabled={loading || submitting}>Cancel</Button>
        <Button type="primary" loading={loading || submitting} onClick={() => void handleSubmit()}>Confirm Shipment</Button>
      </Space>
    </Form>
  );
}
