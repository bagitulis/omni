import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  getOrders,
  bulkShipOrders,
  bulkPrintLabels,
  cancelOrder,
  shipOrder,
} from "@/api/orders";
import type { GetOrdersParams } from "@/api/orders";
import apiClient from "@/api/client";
import { message } from "antd";

interface UseOrdersOptions {
  autoRefresh?: boolean;
  refetchInterval?: number; // in milliseconds
}

const DEFAULT_REFETCH_INTERVAL = 30000; // 30 seconds

export interface ShopeeTimeSlot {
  pickup_time_id?: string;
  pickup_time?: string;
  date?: string;
  time_text?: string;
  time_slot?: string;
  time?: string;
}

export interface ShopeePickupAddress {
  address_id: number;
  address: string;
  time_slots?: ShopeeTimeSlot[];
  time_slot_list?: ShopeeTimeSlot[];
}

export interface ShopeeDropoffBranch {
  branch_id: number;
  address: string;
}

export interface ShopeeShippingOptions {
  pickup?: ShopeePickupAddress[];
  dropoff?: ShopeeDropoffBranch[];
}

export interface ShopeeArrangeShipmentPayload {
  order_sn: string;
  pickup?: {
    address_id: number;
    pickup_time_id?: string;
  };
  dropoff?: {
    branch_id: number;
  };
  tracking_number?: string;
}

export interface TikTokTimeSlot {
  start_time: number;
  end_time: number;
  type?: string;
}

export interface TikTokHandoverSlots {
  time_slots: TikTokTimeSlot[];
}

export interface TikTokArrangeShipmentPayload {
  package_id?: string;
  order_id?: string;
  handover_method: "PICKUP" | "DROP_OFF" | "SELF_SHIPMENT";
  pickup_slot?: {
    start_time: number;
    end_time: number;
  };
  self_shipment?: {
    tracking_number: string;
    shipping_provider_id: string;
  };
}

export interface LazadaArrangeShipmentPayload {
  order_item_ids: string[];
  shipping_provider: string;
  tracking_number?: string;
}

export async function getShippingOptions(
  orderSn: string,
): Promise<ShopeeShippingOptions> {
  const response = await apiClient.get<ShopeeShippingOptions>(
    "/shopee/shipping/options",
    {
      params: { orderSn },
    },
  );

  if (!response.success || !response.data) {
    throw new Error(response.error || "Failed to load Shopee shipping options");
  }

  return response.data;
}

export async function getHandoverTimeSlots(
  orderOrPackageId: string,
): Promise<TikTokHandoverSlots> {
  const response = await apiClient.get<TikTokHandoverSlots>(
    `/tiktok/shipping/timeslots/${encodeURIComponent(orderOrPackageId)}`,
  );

  if (!response.success || !response.data) {
    throw new Error(response.error || "Failed to load TikTok handover slots");
  }

  return response.data;
}

export async function arrangeShopeeShipment(
  data: ShopeeArrangeShipmentPayload,
): Promise<void> {
  const response = await apiClient.post("/shopee/shipping/arrange", data);
  if (!response.success) {
    throw new Error(response.error || "Failed to arrange Shopee shipment");
  }
}

export async function arrangeTikTokShipment(
  data: TikTokArrangeShipmentPayload,
): Promise<void> {
  const response = await apiClient.post("/tiktok/shipping/arrange", data);
  if (!response.success) {
    throw new Error(response.error || "Failed to arrange TikTok shipment");
  }
}

export async function arrangeLazadaShipment(
  data: LazadaArrangeShipmentPayload,
): Promise<void> {
  const response = await apiClient.post("/lazada/orders/ship", data);
  if (!response.success) {
    throw new Error(response.error || "Failed to arrange Lazada shipment");
  }
}

export function useOrders(
  params: GetOrdersParams,
  options: UseOrdersOptions = {},
) {
  const { autoRefresh = true, refetchInterval = DEFAULT_REFETCH_INTERVAL } =
    options;

  const query = useQuery({
    queryKey: ["orders", params],
    queryFn: () => getOrders(params),
    staleTime: 0, // Always refetch on tab change - ensures fresh data
    gcTime: 0, // Don't cache data between tab switches - prevents stale data display
    // NOTE: Removed placeholderData to show loading state immediately on tab switch
    // This prevents showing stale data from previous tab while new data loads
    // Auto-refresh every 30 seconds when enabled
    refetchInterval: autoRefresh ? refetchInterval : false,
    refetchIntervalInBackground: false, // Don't refetch when tab is not focused
  });

  return {
    ...query,
    // Combine isLoading (first load) and isFetching (any fetch) for comprehensive loading state
    // This ensures loading indicator shows on tab switch, not just initial load
    isLoading: query.isLoading || query.isFetching,
  };
}

export function useOrderActions() {
  const queryClient = useQueryClient();

  const shipMutation = useMutation({
    mutationFn: ({
      orderSns,
      platform,
    }: {
      orderSns: string[];
      platform?: string;
    }) => bulkShipOrders(orderSns, platform),
    onSuccess: () => {
      message.success("Orders shipped successfully");
      queryClient.invalidateQueries({ queryKey: ["orders"] });
    },
    onError: (error: Error) => {
      message.error(`Failed to ship orders: ${error.message}`);
    },
  });

  const printMutation = useMutation({
    mutationFn: (orderSns: string[]) => bulkPrintLabels(orderSns),
    onSuccess: () => {
      message.success("Labels generated successfully");
    },
    onError: (error: Error) => {
      message.error(`Failed to print labels: ${error.message}`);
    },
  });

  const cancelMutation = useMutation({
    mutationFn: cancelOrder,
    onSuccess: () => {
      message.success("Order cancelled successfully");
      queryClient.invalidateQueries({ queryKey: ["orders"] });
    },
    onError: (error: Error) => {
      message.error(`Failed to cancel order: ${error.message}`);
    },
  });

  const singleShipMutation = useMutation({
    mutationFn: shipOrder,
    onSuccess: () => {
      message.success("Order shipped successfully");
      queryClient.invalidateQueries({ queryKey: ["orders"] });
    },
    onError: (error: Error) => {
      message.error(`Failed to ship order: ${error.message}`);
    },
  });

  return {
    shipOrders: (orderSns: string[], platform?: string) =>
      shipMutation.mutateAsync({ orderSns, platform }),
    isShipping: shipMutation.isPending,
    printLabels: printMutation.mutateAsync,
    isPrinting: printMutation.isPending,
    cancelOrder: cancelMutation.mutateAsync,
    isCancelling: cancelMutation.isPending,
    shipOrder: singleShipMutation.mutateAsync,
    isSingleShipping: singleShipMutation.isPending,
  };
}
