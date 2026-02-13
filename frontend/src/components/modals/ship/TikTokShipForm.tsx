import { Alert, Button, Space, Spin, Typography, theme } from "antd";
import { useEffect, useMemo, useState } from "react";
import type { Order } from "@/types/order";
import type { ShipConfirmPayload } from "@/components/modals/OrderShipModal";
import { getHandoverTimeSlots } from "@/hooks/useOrders";
import type {
  TikTokArrangeShipmentPayload,
  TikTokTimeSlot,
} from "@/hooks/useOrders";

const { Text } = Typography;

interface TikTokShipFormProps {
  order: Order;
  loading?: boolean;
  onClose: () => void;
  onConfirm: (payload: ShipConfirmPayload) => Promise<void>;
}

export function TikTokShipForm({
  order,
  loading = false,
  onClose,
  onConfirm,
}: TikTokShipFormProps) {
  const { token } = theme.useToken();
  const [timeSlots, setTimeSlots] = useState<TikTokTimeSlot[]>([]);
  const [fetching, setFetching] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [slotsError, setSlotsError] = useState<string | null>(null);

  useEffect(() => {
    const load = async () => {
      setFetching(true);
      setSlotsError(null);
      try {
        const response = await getHandoverTimeSlots(order.order_sn);
        setTimeSlots(response.time_slots || []);
      } catch (error) {
        setSlotsError(
          error instanceof Error
            ? error.message
            : "Failed to load TikTok handover slots",
        );
      } finally {
        setFetching(false);
      }
    };

    void load();
  }, [order.order_sn]);

  const selectedPickupSlot = useMemo(() => {
    if (timeSlots.length === 0) return null;
    return timeSlots.reduce((earliest, current) => {
      if (current.start_time < earliest.start_time) {
        return current;
      }
      return earliest;
    });
  }, [timeSlots]);

  const selectedMode: "PICKUP" | "DROP_OFF" =
    selectedPickupSlot !== null ? "PICKUP" : "DROP_OFF";

  const handleSubmit = async () => {
    const data: TikTokArrangeShipmentPayload = {
      order_id: order.order_sn,
      handover_method: selectedMode,
    };

    if (selectedPickupSlot) {
      data.pickup_slot = {
        start_time: selectedPickupSlot.start_time,
        end_time: selectedPickupSlot.end_time,
      };
    }

    const payload: ShipConfirmPayload = {
      platform: "tiktok",
      data,
    };

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
    <Space direction="vertical" size={12} style={{ width: "100%" }}>
      <Alert
        type="info"
        showIcon
        message="TikTok direct shipment flow"
        description="No extra options are required. Click Arrange Shipment and the system will continue to label printing."
        style={{ marginBottom: 16, borderColor: token.colorInfoBorder }}
      />

      {slotsError && (
        <Alert
          type="warning"
          showIcon
          message="Proceeding without manual shipping options"
          description={`tiktok API error: ${slotsError}`}
        />
      )}

      <Text type="secondary">Order: {order.order_sn}</Text>

      <Space style={{ justifyContent: "flex-end", width: "100%" }}>
        <Button onClick={onClose} disabled={loading || submitting}>
          Cancel
        </Button>
        <Button
          type="primary"
          loading={loading || submitting}
          onClick={() => void handleSubmit()}
        >
          Arrange Shipment
        </Button>
      </Space>
    </Space>
  );
}
