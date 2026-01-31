/**
 * useOrderActions Composable
 * Handles order shipping and cancellation operations
 */

import { ref } from "vue";
import { useToast } from "@/composables/useToast";
import { getAuthHeaders, getApiBaseUrl } from "@/utils/apiHeaders";

// API base URLs for each platform
const SHOPEE_API = getApiBaseUrl("/shopee");
const LAZADA_API = getApiBaseUrl("/lazada");
const TIKTOK_API = getApiBaseUrl("/tiktok");

// Types
export interface ShipOrderParams {
  order_no: string;
  platform: string;
  shipping_provider?: string;
  tracking_number?: string;
  address_id?: number;
  pickup_time_id?: string;
  // Lazada specific
  order_item_ids?: string[];
  // TikTok specific
  package_id?: string;
}

export interface CancelOrderParams {
  order_no: string;
  platform: string;
  cancel_reason: string;
  reason_detail?: string;
  // Lazada specific
  order_item_id?: string;
  reason_id?: string;
}

export interface ShipOrderResponse {
  success: boolean;
  order_no?: string;
  tracking_number?: string;
  message?: string;
}

export interface CancelOrderResponse {
  success: boolean;
  order_no?: string;
  message?: string;
}

export function useOrderActions() {
  const toast = useToast();

  // State
  const loading = ref(false);
  const error = ref<string | null>(null);

  /**
   * Get platform-specific API URL
   */
  const getPlatformApiUrl = (platform: string): string => {
    const platformMap: Record<string, string> = {
      shopee: SHOPEE_API,
      lazada: LAZADA_API,
      tiktok: TIKTOK_API,
    };
    return platformMap[platform.toLowerCase()] || SHOPEE_API;
  };

  /**
   * Clean undefined values from request body
   */
  const cleanRequestBody = (body: Record<string, any>): Record<string, any> => {
    const cleaned: Record<string, any> = {};
    for (const key of Object.keys(body)) {
      if (body[key] !== undefined) {
        cleaned[key] = body[key];
      }
    }
    return cleaned;
  };

  /**
   * Build platform-specific ship request body
   */
  const buildShipRequestBody = (
    platform: string,
    params: ShipOrderParams,
  ): Record<string, any> => {
    if (platform === "shopee") {
      return {
        order_sn: params.order_no,
        shipping_provider: params.shipping_provider,
        tracking_number: params.tracking_number,
        address_id: params.address_id,
        pickup_time_id: params.pickup_time_id,
      };
    }

    if (platform === "lazada") {
      return {
        order_item_ids: params.order_item_ids || [],
        shipping_provider: params.shipping_provider || "",
        tracking_number: params.tracking_number,
      };
    }

    if (platform === "tiktok") {
      return {
        order_id: params.order_no,
        package_id: params.package_id || "",
        shipping_provider: params.shipping_provider || "",
        tracking_number: params.tracking_number || "",
      };
    }

    throw new Error(`Unsupported platform: ${platform}`);
  };

  /**
   * Build platform-specific cancel request body
   */
  const buildCancelRequestBody = (
    platform: string,
    params: CancelOrderParams,
  ): Record<string, any> => {
    if (platform === "shopee") {
      return {
        order_sn: params.order_no,
        cancel_reason: params.cancel_reason,
      };
    }

    if (platform === "lazada") {
      return {
        order_item_id: params.order_item_id || params.order_no,
        reason_id: params.reason_id || params.cancel_reason,
        reason_detail: params.reason_detail,
      };
    }

    if (platform === "tiktok") {
      return {
        order_id: params.order_no,
        cancel_reason: params.cancel_reason,
      };
    }

    throw new Error(`Unsupported platform: ${platform}`);
  };

  /**
   * Ship an order
   */
  const shipOrder = async (
    params: ShipOrderParams,
  ): Promise<ShipOrderResponse> => {
    loading.value = true;
    error.value = null;

    try {
      const baseUrl = getPlatformApiUrl(params.platform);
      const platform = params.platform.toLowerCase();

      const requestBody = cleanRequestBody(
        buildShipRequestBody(platform, params),
      );

      const response = await fetch(`${baseUrl}/orders/ship`, {
        method: "POST",
        headers: getAuthHeaders(),
        body: JSON.stringify(requestBody),
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || "Failed to ship order");
      }

      toast.add({
        severity: "success",
        summary: "Success",
        detail: `Order ${params.order_no} has been shipped`,
        life: 3000,
      });

      return {
        success: true,
        order_no: params.order_no,
        tracking_number: data.data?.tracking_number,
      };
    } catch (err: any) {
      const errorMessage = err.message || "Failed to ship order";
      error.value = errorMessage;

      toast.add({
        severity: "error",
        summary: "Error",
        detail: errorMessage,
        life: 5000,
      });

      return {
        success: false,
        message: errorMessage,
      };
    } finally {
      loading.value = false;
    }
  };

  /**
   * Cancel an order
   */
  const cancelOrder = async (
    params: CancelOrderParams,
  ): Promise<CancelOrderResponse> => {
    loading.value = true;
    error.value = null;

    try {
      const baseUrl = getPlatformApiUrl(params.platform);
      const platform = params.platform.toLowerCase();

      const requestBody = cleanRequestBody(
        buildCancelRequestBody(platform, params),
      );

      const response = await fetch(`${baseUrl}/orders/cancel`, {
        method: "POST",
        headers: getAuthHeaders(),
        body: JSON.stringify(requestBody),
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || "Failed to cancel order");
      }

      toast.add({
        severity: "success",
        summary: "Success",
        detail: `Order ${params.order_no} has been cancelled`,
        life: 3000,
      });

      return {
        success: true,
        order_no: params.order_no,
      };
    } catch (err: any) {
      const errorMessage = err.message || "Failed to cancel order";
      error.value = errorMessage;

      toast.add({
        severity: "error",
        summary: "Error",
        detail: errorMessage,
        life: 5000,
      });

      return {
        success: false,
        message: errorMessage,
      };
    } finally {
      loading.value = false;
    }
  };

  /**
   * Get shipping parameters (for Shopee pickup addresses)
   */
  const getShippingParameters = async (
    orderNo: string,
    platform: string,
  ): Promise<any> => {
    if (platform.toLowerCase() !== "shopee") {
      return null;
    }

    try {
      const response = await fetch(
        `${SHOPEE_API}/orders/${orderNo}/shipping-params`,
        {
          method: "GET",
          headers: getAuthHeaders(),
        },
      );

      if (!response.ok) {
        return null;
      }

      const data = await response.json();
      return data.data;
    } catch {
      return null;
    }
  };

  /**
   * Reset error state
   */
  const clearError = () => {
    error.value = null;
  };

  return {
    // State
    loading,
    error,

    // Methods
    shipOrder,
    cancelOrder,
    getShippingParameters,
    clearError,
  };
}
