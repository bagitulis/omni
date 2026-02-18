import type { ShipConfirmPayload } from "@/components/modals/OrderShipModal";
import type {
  ShopeeInfoNeeded,
  ShopeePickupAddress,
  ShopeeTimeSlot,
} from "@/hooks/useOrders";
import type { Order } from "@/types/order";

export interface ShopeeShipFormValues {
  mode: "pickup" | "dropoff";
  address_id?: number;
  pickup_time_id?: string;
  branch_id?: number;
  tracking_number?: string;
}

export interface ShopeeShipFormProps {
  order: Order;
  loading?: boolean;
  onClose: () => void;
  onConfirm: (payload: ShipConfirmPayload) => Promise<void>;
}

export interface ShopeeOptionsState {
  pickup: ShopeePickupAddress[];
  dropoff: Array<{ branch_id: number; address: string }>;
  info_needed?: ShopeeInfoNeeded;
}

export const getTimeSlots = (address?: ShopeePickupAddress): ShopeeTimeSlot[] =>
  address?.time_slots || address?.time_slot_list || [];

export const hasRequiredField = (
  infoNeeded: ShopeeInfoNeeded | undefined,
  mode: "pickup" | "dropoff",
  field: string,
): boolean => {
  const fields = mode === "pickup" ? infoNeeded?.pickup : infoNeeded?.dropoff;
  return Boolean(fields?.includes(field));
};

export const formatShopeeDate = (value?: string | number): string => {
  if (!value) return "";

  if (typeof value === "number") {
    const timestamp = value > 1_000_000_000_000 ? value : value * 1000;
    return new Date(timestamp).toLocaleDateString("en-US", {
      weekday: "short",
      month: "short",
      day: "numeric",
    });
  }

  const parsed = Number(value);
  if (!Number.isNaN(parsed)) {
    const timestamp = parsed > 1_000_000_000_000 ? parsed : parsed * 1000;
    return new Date(timestamp).toLocaleDateString("en-US", {
      weekday: "short",
      month: "short",
      day: "numeric",
    });
  }

  return value;
};

export const getTimeSlotLabel = (slot: ShopeeTimeSlot): string => {
  const text =
    slot.pickup_time ||
    slot.time_text ||
    slot.time ||
    slot.time_slot ||
    "No label";
  const dateLabel = formatShopeeDate(slot.date);
  return dateLabel ? `${dateLabel} - ${text}` : text;
};
