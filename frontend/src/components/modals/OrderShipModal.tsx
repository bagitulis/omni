import { Alert, Drawer, Space, Tag, Typography, theme } from "antd";
import { Order } from "@/types/order";
import { ShopeeShipForm } from "@/components/modals/ship/ShopeeShipForm";
import { TikTokShipForm } from "@/components/modals/ship/TikTokShipForm";
import { LazadaShipForm } from "@/components/modals/ship/LazadaShipForm";
import {
  LazadaArrangeShipmentPayload,
  ShopeeArrangeShipmentPayload,
  TikTokArrangeShipmentPayload,
  arrangeLazadaShipment,
  arrangeShopeeShipment,
  arrangeTikTokShipment,
} from "@/hooks/useOrders";

const { Text } = Typography;

export type ShipConfirmPayload =
  | { platform: "shopee"; data: ShopeeArrangeShipmentPayload }
  | { platform: "tiktok"; data: TikTokArrangeShipmentPayload }
  | { platform: "lazada"; data: LazadaArrangeShipmentPayload };

interface OrderShipModalProps {
  open: boolean;
  onClose: () => void;
  order: Order | null;
  onConfirm?: (payload: ShipConfirmPayload) => Promise<void>;
  loading?: boolean;
}

const resolvePlatform = (
  value?: string,
): "shopee" | "tiktok" | "lazada" | "" => {
  const normalized = value?.toLowerCase();
  if (
    normalized === "shopee" ||
    normalized === "tiktok" ||
    normalized === "lazada"
  ) {
    return normalized;
  }
  return "";
};

async function submitByPlatform(payload: ShipConfirmPayload): Promise<void> {
  if (payload.platform === "shopee") {
    await arrangeShopeeShipment(payload.data);
    return;
  }
  if (payload.platform === "tiktok") {
    await arrangeTikTokShipment(payload.data);
    return;
  }
  await arrangeLazadaShipment(payload.data);
}

export function OrderShipModal({
  open,
  onClose,
  order,
  onConfirm,
  loading = false,
}: OrderShipModalProps) {
  const { token } = theme.useToken();

  if (!order) {
    return null;
  }

  const platform = resolvePlatform(order.platform);
  const handleConfirm = async (payload: ShipConfirmPayload) => {
    if (onConfirm) {
      await onConfirm(payload);
      return;
    }
    await submitByPlatform(payload);
  };

  return (
    <Drawer
      title="Arrange Shipment"
      placement="right"
      width={420}
      open={open}
      onClose={onClose}
      destroyOnClose
    >
      <Space direction="vertical" size={16} style={{ width: "100%" }}>
        <Space
          align="center"
          style={{ justifyContent: "space-between", width: "100%" }}
        >
          <Text type="secondary">Order SN</Text>
          <Text strong copyable>
            {order.order_sn}
          </Text>
        </Space>
        <Space align="center">
          <Text type="secondary">Platform</Text>
          <Tag color="blue">{platform || "unknown"}</Tag>
        </Space>

        {platform === "shopee" && (
          <ShopeeShipForm
            order={order}
            loading={loading}
            onConfirm={handleConfirm}
            onClose={onClose}
          />
        )}
        {platform === "tiktok" && (
          <TikTokShipForm
            order={order}
            loading={loading}
            onConfirm={handleConfirm}
            onClose={onClose}
          />
        )}
        {platform === "lazada" && (
          <LazadaShipForm
            order={order}
            loading={loading}
            onConfirm={handleConfirm}
            onClose={onClose}
          />
        )}

        {!platform && (
          <Alert
            type="error"
            showIcon
            message="Unsupported platform"
            description={`Shipping flow for platform '${order.platform || "unknown"}' is not implemented.`}
            style={{ borderColor: token.colorErrorBorder }}
          />
        )}
      </Space>
    </Drawer>
  );
}
